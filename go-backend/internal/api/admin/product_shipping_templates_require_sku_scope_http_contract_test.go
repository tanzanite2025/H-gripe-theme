package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestProductMutationsRejectProductLevelShippingTemplatesEvenWhenEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &ProductHandler{}
	for _, method := range []string{http.MethodPost, http.MethodPut} {
		for _, value := range []string{"[]", "null", "[1]"} {
			t.Run(method+value, func(t *testing.T) {
				router := gin.New()
				router.POST("/products/1", handler.CreateProduct)
				router.PUT("/products/:id", handler.UpdateProduct)
				recorder := httptest.NewRecorder()
				request := httptest.NewRequestWithContext(t.Context(), method, "/products/1", strings.NewReader(`{"name":"Rim","slug":"rim","status":"active","shipping_template_ids":`+value+`}`))
				request.Header.Set("Content-Type", "application/json")
				router.ServeHTTP(recorder, request)
				require.Equal(t, http.StatusBadRequest, recorder.Code)
				require.Contains(t, recorder.Body.String(), "must be configured on each SKU")
			})
		}
	}
}
