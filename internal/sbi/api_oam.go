package sbi

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func (s *Server) getOAMRoutes() []Route {
	return []Route{
		{
			Name:    "Index",
			Method:  http.MethodGet,
			Pattern: "/",
			APIFunc: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "Service Available"})
			},
		},
		{
			Name:    "Get UE PDU Session Info",
			Method:  http.MethodGet,
			Pattern: "/ue-pdu-session-info/:smContextRef",
			APIFunc: s.HTTPGetUEPDUSessionInfo,
		},
		{
			Name:    "Get SM Context Inventory",
			Method:  http.MethodGet,
			Pattern: "/sm-contexts",
			APIFunc: s.HTTPGetSMContextInventory,
		},
		{
			Name:    "Get SMF Userplane Information",
			Method:  http.MethodGet,
			Pattern: "/user-plane-info/",
			APIFunc: s.HTTPGetSMFUserPlaneInfo,
		},
	}
}

func (s *Server) HTTPGetUEPDUSessionInfo(c *gin.Context) {
	smContextRef := c.Params.ByName("smContextRef")

	s.Processor().HandleOAMGetUEPDUSessionInfo(c, smContextRef)
}

func (s *Server) HTTPGetSMContextInventory(c *gin.Context) {
	expected := os.Getenv("CITYVERSE_SMF_INVENTORY_TOKEN")
	if expected == "" {
		c.Status(http.StatusNotFound)
		return
	}
	provided := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	if subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
		c.Status(http.StatusUnauthorized)
		return
	}
	s.Processor().HandleOAMGetSMContextInventory(c)
}

func (s *Server) HTTPGetSMFUserPlaneInfo(c *gin.Context) {
	s.Processor().HandleGetSMFUserPlaneInfo(c)
}
