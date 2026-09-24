package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRespondOrderServiceErrorMapsOrderHideNotAllowedToBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	respondOrderServiceError(
		context,
		service.ErrOrderHideNotAllowed,
		"fallback",
		http.StatusInternalServerError,
	)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "order_hide_not_allowed")
	assert.Contains(t, recorder.Body.String(), "must be retained")
}
