package sbi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSMContextInventoryRequiresConfiguredBearer(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		authorize  string
		wantStatus int
	}{
		{name: "disabled", wantStatus: http.StatusNotFound},
		{name: "missing bearer", token: "secret", wantStatus: http.StatusUnauthorized},
		{name: "wrong bearer", token: "secret", authorize: "Bearer wrong", wantStatus: http.StatusUnauthorized},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("CITYVERSE_SMF_INVENTORY_TOKEN", test.token)
			response := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(response)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/nsmf-oam/v1/sm-contexts", nil)
			if test.authorize != "" {
				ctx.Request.Header.Set("Authorization", test.authorize)
			}
			(&Server{}).HTTPGetSMContextInventory(ctx)
			if ctx.Writer.Status() != test.wantStatus {
				t.Fatalf("status = %d, want %d", ctx.Writer.Status(), test.wantStatus)
			}
		})
	}
}
