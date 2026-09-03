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

func TestRespondProductBrandErrorForSpokeRimCatalogReference(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	respondProductBrandError(context, fmt.Errorf("%w: 1 spoke rim brand still references this brand", service.ErrProductBrandInSpokeRimCatalog))

	require.Equal(t, http.StatusConflict, recorder.Code)
	var response map[string]string
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Equal(t, "Product brand is referenced by the spoke rim catalog and cannot be deleted", response["error"])
}
