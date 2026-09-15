package marketing

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLegacyReferralEndpointReturnsGone(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/api/v1/marketing/loyalty/referral", strings.NewReader(`{"referee_id":2}`))

	(&Handler{}).CreateReferral(context)

	assert.Equal(t, http.StatusGone, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "referral_legacy_endpoint_retired")
}
