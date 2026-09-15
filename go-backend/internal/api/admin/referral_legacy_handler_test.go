package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLegacyReferralStatusMutationReturnsGone(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPatch, "/api/admin/marketing/loyalty/referrals/1/status", nil)

	(&MarketingHandler{}).UpdateReferralStatus(context)

	assert.Equal(t, http.StatusGone, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "referral_legacy_endpoint_retired")
}
