package context

import (
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	smContextInventoryMu       sync.RWMutex
	smContextInventoryRevision uint64
	smfBootEpoch               = uuid.NewString()
)

func storeSMContext(smContext *SMContext) {
	smContextInventoryMu.Lock()
	defer smContextInventoryMu.Unlock()
	smContextPool.Store(smContext.Ref, smContext)
	smContextInventoryRevision++
}

func deleteSMContext(ref string) {
	smContextInventoryMu.Lock()
	defer smContextInventoryMu.Unlock()
	smContextPool.Delete(ref)
	smContextInventoryRevision++
}

// VisitSMContextSnapshot visits one authoritative process-wide snapshot. The
// callback sees every context that existed at observedAt and no context that did
// not. All visited contexts remain locked for the complete projection.
//
// Locking is deliberately two-phase. Release paths already enter RemoveSMContext
// while holding SMLock, so holding the inventory lock while first acquiring the
// context locks would invert that established order and deadlock.
func VisitSMContextSnapshot(visitor func(*SMContext)) (bootEpoch string, observedAt time.Time) {
	for {
		smContextInventoryMu.RLock()
		revision := smContextInventoryRevision
		contexts := make([]*SMContext, 0)
		smContextPool.Range(func(_, value interface{}) bool {
			if smContext, ok := value.(*SMContext); ok {
				contexts = append(contexts, smContext)
			}
			return true
		})
		smContextInventoryMu.RUnlock()

		sort.Slice(contexts, func(i, j int) bool { return contexts[i].Ref < contexts[j].Ref })
		for _, smContext := range contexts {
			smContext.SMLock.Lock()
		}

		smContextInventoryMu.RLock()
		if smContextInventoryRevision != revision {
			smContextInventoryMu.RUnlock()
			for i := len(contexts) - 1; i >= 0; i-- {
				contexts[i].SMLock.Unlock()
			}
			continue
		}

		observedAt = time.Now().UTC()
		for _, smContext := range contexts {
			visitor(smContext)
		}
		smContextInventoryMu.RUnlock()
		for i := len(contexts) - 1; i >= 0; i-- {
			contexts[i].SMLock.Unlock()
		}
		return smfBootEpoch, observedAt
	}
}
