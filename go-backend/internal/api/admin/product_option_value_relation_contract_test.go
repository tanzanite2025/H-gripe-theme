package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRetiredOptionValueRelationFieldIsRejected(t *testing.T) {
	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(`{"option_value_relations":[]}`), &raw))
	require.Equal(t, "option_value_relations", retiredProductSpecificationRequestField(raw))
}

func TestProductOptionValueRelationErrorsReturnBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	respondProductServiceError(context, fmt.Errorf("invalid graph: %w", service.ErrProductOptionRelationInvalid), "fallback")

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Contains(t, recorder.Body.String(), service.ErrProductOptionRelationInvalid.Error())
}
