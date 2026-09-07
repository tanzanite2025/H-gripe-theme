package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRespondOrderServiceErrorMapsFulfillmentTrackingConflictToConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	respondOrderServiceError(
		context,
		service.ErrOrderFulfillmentTrackingConflict,
		"fallback",
		http.StatusInternalServerError,
	)

	assert.Equal(t, http.StatusConflict, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "different tracking information")
}
