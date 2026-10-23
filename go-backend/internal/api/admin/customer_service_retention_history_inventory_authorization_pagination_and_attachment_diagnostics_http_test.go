package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func TestCustomerServiceRetentionHistoryInventoryRequiresAdminAndPreservesRecoveryWindowAndAttachmentDiagnostics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&ticket.Ticket{}, &ticket.TicketMessage{}))
	now := time.Now().UTC()
	conversation := ticket.Ticket{ID: 801, TicketNumber: "RECOVERABLE-HISTORY", UserID: 1,
		Subject: "History preserved during recovery", Category: "customer_service", Status: "closed",
		UpdatedAt: now.Add(-time.Hour), DeletedAt: gorm.DeletedAt{Time: now.Add(-time.Hour), Valid: true}}
	require.NoError(t, db.Create(&conversation).Error)
	require.NoError(t, db.Create(&ticket.TicketMessage{TicketID: conversation.ID, Content: "older retained content",
		Attachments: `["uploads/retained-history.png"]`, CreatedAt: now.Add(-time.Hour)}).Error)
	require.NoError(t, db.Create(&ticket.TicketMessage{TicketID: conversation.ID, Content: "latest retained content",
		Attachments: "malformed historical JSON", CreatedAt: now}).Error)
	handler := &TicketHandler{customerServiceRetention: service.NewCustomerServiceRetentionService(
		repository.NewTicketRepository(db), nil, nil, nil, nil, nil)}

	for _, scenario := range []struct {
		name, role, query string
		expectedStatus    int
		expectedRecords   int
		expectedPageSize  int
	}{
		{name: "administrator sees recovery inventory", role: "admin", query: "?status=closed", expectedStatus: 200, expectedRecords: 1, expectedPageSize: 20},
		{name: "support denied", role: "support", expectedStatus: 403},
		{name: "manager denied", role: "manager", expectedStatus: 403},
		{name: "invalid lifecycle rejected", role: "admin", query: "?status=invalid", expectedStatus: 400},
		{name: "out of range page is empty", role: "admin", query: "?page=2&page_size=1", expectedStatus: 200, expectedPageSize: 1},
		{name: "page size capped", role: "admin", query: "?page_size=10000", expectedStatus: 200, expectedRecords: 1, expectedPageSize: 100},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("role", scenario.role)
				c.Set("user_role", scenario.role)
				c.Set("user_id", uint(1))
			})
			router.GET("/api/admin/customer-service/conversations/retention/records", middleware.AdminOnly(), handler.ListCustomerServiceRetentionRecords)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/admin/customer-service/conversations/retention/records"+scenario.query, nil))
			require.Equal(t, scenario.expectedStatus, recorder.Code, recorder.Body.String())
			if recorder.Code != http.StatusOK {
				return
			}
			var body struct {
				Code int `json:"code"`
				Data struct {
					Records []struct {
						ID        uint                                        `json:"id"`
						Retention service.CustomerServiceRetentionEligibility `json:"retention"`
					} `json:"records"`
					Pagination struct {
						Total    int `json:"total"`
						PageSize int `json:"page_size"`
					} `json:"pagination"`
				} `json:"data"`
			}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
			require.Zero(t, body.Code)
			require.NotNil(t, body.Data.Records)
			require.Len(t, body.Data.Records, scenario.expectedRecords)
			require.Equal(t, 1, body.Data.Pagination.Total)
			require.Equal(t, scenario.expectedPageSize, body.Data.Pagination.PageSize)
			if scenario.expectedRecords > 0 {
				item := body.Data.Records[0]
				require.Equal(t, conversation.ID, item.ID)
				require.False(t, item.Retention.Eligible)
				require.NotNil(t, item.Retention.SoftDeletedAt)
				require.NotNil(t, item.Retention.PurgeAfter)
				require.Contains(t, item.Retention.Reason, "recovery window")
				require.Equal(t, 1, item.Retention.AttachmentReferenceCount)
				require.Equal(t, 1, item.Retention.MalformedAttachmentPayloads)
			}
		})
	}
}
