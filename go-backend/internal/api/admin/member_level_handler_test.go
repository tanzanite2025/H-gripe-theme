package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCreateMemberLevelRejectsFloatDiscountRate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	router := gin.New()
	router.POST("/levels", NewMarketingHandler(nil).CreateMemberLevel)

	request := httptest.NewRequest(http.MethodPost, "/levels", strings.NewReader(`{
		"name":"Silver",
		"min_points":0,
		"max_points":999,
		"discount_rate_decimal":5.5
	}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestUpdateMemberLevelRejectsFloatDiscountRate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	router := gin.New()
	router.PUT("/levels/:id", NewMarketingHandler(nil).UpdateMemberLevel)

	request := httptest.NewRequest(http.MethodPut, "/levels/1", strings.NewReader(`{"discount_rate_decimal":5.5}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}
