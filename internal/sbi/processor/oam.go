package processor

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/smf/internal/context"
	"github.com/free5gc/smf/pkg/factory"
	"github.com/free5gc/util/metrics/sbi"
)

type PDUSessionInfo struct {
	SmContextRef string
	Supi         string
	PDUSessionID string
	Dnn          string
	Sst          string
	Sd           string
	AnType       models.AccessType
	PDUAddress   string
	SessionRule  models.SessionRule
	UpCnxState   models.UpCnxState
	Tunnel       context.UPTunnel
}

type SMContextInventory struct {
	SchemaVersion     string
	InventoryComplete bool
	Consistency       string
	BootEpoch         string
	ObservedAt        string
	Count             int
	Contexts          []*PDUSessionInfo
}

func pduSessionInfoFromLockedSMContext(smContext *context.SMContext) *PDUSessionInfo {
	info := &PDUSessionInfo{
		SmContextRef: smContext.Ref,
		Supi:         smContext.Supi, PDUSessionID: strconv.Itoa(int(smContext.PDUSessionID)),
		Dnn: smContext.Dnn, AnType: smContext.AnType,
		PDUAddress: smContext.PDUAddress.String(), UpCnxState: smContext.UpCnxState,
	}
	if smContext.SNssai != nil {
		info.Sst = strconv.Itoa(int(smContext.SNssai.Sst))
		info.Sd = smContext.SNssai.Sd
	}
	return info
}

func (p *Processor) HandleOAMGetUEPDUSessionInfo(c *gin.Context, smContextRef string) {
	smContext := context.GetSMContextByRef(smContextRef)
	if smContext == nil {
		statusCode := http.StatusNotFound
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, http.StatusText(statusCode))
		c.JSON(statusCode, nil)
		return
	}

	smContext.SMLock.Lock()
	pduSessionInfo := pduSessionInfoFromLockedSMContext(smContext)
	smContext.SMLock.Unlock()
	c.JSON(http.StatusOK, pduSessionInfo)
}

func (p *Processor) HandleOAMGetSMContextInventory(c *gin.Context) {
	contexts := make([]*PDUSessionInfo, 0)
	bootEpoch, observedAt := context.VisitSMContextSnapshot(func(smContext *context.SMContext) {
		contexts = append(contexts, pduSessionInfoFromLockedSMContext(smContext))
	})
	c.JSON(http.StatusOK, &SMContextInventory{
		SchemaVersion:     "cityverse.smf-oam/v2",
		InventoryComplete: true,
		Consistency:       "atomic-membership-and-projection",
		BootEpoch:         bootEpoch,
		ObservedAt:        observedAt.Format(time.RFC3339Nano),
		Count:             len(contexts),
		Contexts:          contexts,
	})
}

func (p *Processor) HandleGetSMFUserPlaneInfo(c *gin.Context) {
	c.JSON(http.StatusOK, factory.SmfConfig.Configuration.UserPlaneInformation)
}
