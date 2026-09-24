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

func TestProductOptionValueRelationUpdateContractMapsAllFields(t *testing.T) {
	var request productUpdateRequest
	require.NoError(t, json.Unmarshal([]byte(`{
		"option_value_relations": [{
			"id": 17,
			"source_option_value_id": 101,
			"target_option_value_id": 202,
			"relation_type": "requires"
		}]
	}`), &request))

	require.Len(t, request.OptionValueRelations, 1)
	input := normalizeProductOptionValueRelationRequests(request.OptionValueRelations)
	require.Len(t, input, 1)
	require.NotNil(t, input[0].ID)
	require.Equal(t, uint(17), *input[0].ID)
	require.Equal(t, uint(101), input[0].SourceOptionValueID)
	require.Equal(t, uint(202), input[0].TargetOptionValueID)
	require.Equal(t, "requires", input[0].RelationType)
}

func TestProductOptionValueRelationErrorsReturnBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	respondProductServiceError(context, fmt.Errorf("invalid graph: %w", service.ErrProductOptionRelationInvalid), "fallback")

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Contains(t, recorder.Body.String(), service.ErrProductOptionRelationInvalid.Error())
}
