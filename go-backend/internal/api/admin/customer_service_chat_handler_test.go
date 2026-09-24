package admin

import (
	"commerce-platform/internal/repository"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCustomerServiceRetentionIDsAcceptsAliasesAndDeduplicates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/?conversation_ids=7,8&ticket_ids=8,9", nil)

	ids, ok := parseCustomerServiceRetentionIDs(context)

	require.True(t, ok)
	require.Equal(t, []uint{7, 8, 9}, ids)
	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestParseCustomerServiceRetentionIDsRejectsInvalidAndOversizedInput(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("invalid id", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)
		context.Request = httptest.NewRequest(http.MethodGet, "/?conversation_ids=7,nope", nil)

		ids, ok := parseCustomerServiceRetentionIDs(context)

		require.False(t, ok)
		require.Nil(t, ids)
		require.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("more than one hundred ids", func(t *testing.T) {
		values := make([]string, 101)
		for index := range values {
			values[index] = strconv.Itoa(index + 1)
		}
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)
		context.Request = httptest.NewRequest(http.MethodGet, "/?ticket_ids="+strings.Join(values, ","), nil)

		ids, ok := parseCustomerServiceRetentionIDs(context)

		require.False(t, ok)
		require.Nil(t, ids)
		require.Equal(t, http.StatusBadRequest, recorder.Code)
	})
}

func TestRespondAdminCustomerServiceErrorMapsStatusVersionConflictToHTTP409(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	respondAdminCustomerServiceError(context, repository.ErrTicketStatusVersionConflict)

	assert.Equal(t, http.StatusConflict, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"code":"conflict"`)
}

func TestNormalizeAdminCustomerServiceMessageTypeAllowsVideo(t *testing.T) {
	require.Equal(t, "video", normalizeAdminCustomerServiceMessageType(" video "))
	require.Equal(t, "text", normalizeAdminCustomerServiceMessageType("unsupported"))
}

func TestParseAdminCustomerServiceConversationFiltersDefaultsToInboxAndPreservesAllView(t *testing.T) {
	gin.SetMode(gin.TestMode)

	defaultContext, _ := gin.CreateTestContext(httptest.NewRecorder())
	defaultContext.Request = httptest.NewRequest("GET", "/api/admin/customer-service/conversations", nil)
	defaultFilters, ok := parseAdminCustomerServiceConversationFilters(defaultContext)
	require.True(t, ok)
	assert.Empty(t, defaultFilters.View)
	assert.Equal(t, "inbox", adminCustomerServiceConversationFilterResponse(defaultFilters)["view"])

	allContext, _ := gin.CreateTestContext(httptest.NewRecorder())
	allContext.Request = httptest.NewRequest("GET", "/api/admin/customer-service/conversations?view=all", nil)
	allFilters, ok := parseAdminCustomerServiceConversationFilters(allContext)
	require.True(t, ok)
	assert.Equal(t, "all", allFilters.View)
	assert.Equal(t, "all", adminCustomerServiceConversationFilterResponse(allFilters)["view"])
}

func TestMarshalAdminCustomerServiceMessageMetadataIgnoresNull(t *testing.T) {
	payload, err := marshalAdminCustomerServiceMessageMetadata(map[string]any{
		"url":       "https://example.test/products/demo",
		"source":    "admin",
		"thumbnail": nil,
	})
	require.NoError(t, err)
	assert.Contains(t, payload, `"url":"https://example.test/products/demo"`)
	assert.Contains(t, payload, `"source":"admin"`)

	empty, err := marshalAdminCustomerServiceMessageMetadata(nil)
	require.NoError(t, err)
	assert.Empty(t, empty)
}
