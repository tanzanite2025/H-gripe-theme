package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseReferralAdminFiltersUsesInclusiveCalendarDates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(http.MethodGet, "/?keyword=ALEX&status=vesting&from=2026-09-01&to=2026-09-15", nil)

	filters, err := parseReferralAdminFilters(context)
	require.NoError(t, err)
	require.NotNil(t, filters.From)
	require.NotNil(t, filters.To)
	assert.Equal(t, "ALEX", filters.Keyword)
	assert.Equal(t, "vesting", filters.Status)
	assert.Equal(t, time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), *filters.From)
	assert.Equal(t, time.Date(2026, 9, 16, 0, 0, 0, 0, time.UTC), *filters.To)
}

func TestParseReferralAdminFiltersRejectsInvalidDateRange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name  string
		query string
	}{
		{name: "invalid format", query: "from=09-01-2026"},
		{name: "end before start", query: "from=2026-09-16&to=2026-09-15"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _ := gin.CreateTestContext(httptest.NewRecorder())
			context.Request = httptest.NewRequest(http.MethodGet, "/?"+test.query, nil)

			_, err := parseReferralAdminFilters(context)
			assert.Error(t, err)
		})
	}
}
