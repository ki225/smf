package context

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestVisitSMContextSnapshotIsAuthoritativeAcrossStoreAndDelete(t *testing.T) {
	first := &SMContext{Ref: "inventory-test-first"}
	second := &SMContext{Ref: "inventory-test-second"}
	storeSMContext(first)
	storeSMContext(second)
	t.Cleanup(func() {
		deleteSMContext(first.Ref)
		deleteSMContext(second.Ref)
	})

	projectionStarted := make(chan struct{})
	allowProjection := make(chan struct{})
	snapshotDone := make(chan []string)
	go func() {
		refs := make([]string, 0, 2)
		bootEpoch, observedAt := VisitSMContextSnapshot(func(smContext *SMContext) {
			if len(refs) == 0 {
				close(projectionStarted)
				<-allowProjection
			}
			refs = append(refs, smContext.Ref)
		})
		require.NotEmpty(t, bootEpoch)
		require.False(t, observedAt.IsZero())
		snapshotDone <- refs
	}()

	<-projectionStarted
	mutationDone := make(chan struct{})
	go func() {
		deleteSMContext(first.Ref)
		storeSMContext(&SMContext{Ref: "inventory-test-third"})
		close(mutationDone)
	}()

	select {
	case <-mutationDone:
		t.Fatal("pool mutation completed during an authoritative projection")
	case <-time.After(25 * time.Millisecond):
	}
	close(allowProjection)
	refs := <-snapshotDone
	require.Contains(t, refs, first.Ref)
	require.Contains(t, refs, second.Ref)
	<-mutationDone
	deleteSMContext("inventory-test-third")
}

func TestVisitSMContextSnapshotKeepsBootEpochStable(t *testing.T) {
	first, _ := VisitSMContextSnapshot(func(*SMContext) {})
	second, _ := VisitSMContextSnapshot(func(*SMContext) {})
	require.NotEmpty(t, first)
	require.Equal(t, first, second)
}

func TestVisitSMContextSnapshotSerializesConcurrentProjections(t *testing.T) {
	smContext := &SMContext{Ref: "inventory-test-serialized"}
	storeSMContext(smContext)
	t.Cleanup(func() { deleteSMContext(smContext.Ref) })

	var wg sync.WaitGroup
	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func() {
			defer wg.Done()
			VisitSMContextSnapshot(func(*SMContext) {})
		}()
	}
	wg.Wait()
}
