package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createDashboardBusinessFollowUpReadServiceForTestDatabase(db *gorm.DB) *service.DashboardBusinessFollowUpReadService {
	return service.NewDashboardBusinessFollowUpReadService(repository.NewDashboardOutboxBusinessFollowUpReadRepository(db), service.NewDashboardBusinessFollowUpBusinessReferenceResolver(repository.NewDashboardBusinessFollowUpBusinessIdentityReadRepository(db)))
}

func createDashboardOutboxFailureClosedLoopTestRouter(t *testing.T) (*gin.Engine, *repository.OutboxRepository, *gorm.DB) {
	t.Helper()
	db := newOutboxReconciliationTestDB(t)
	require.NoError(t, db.Exec(`CREATE TABLE orders (id INTEGER PRIMARY KEY, order_number TEXT, deleted_at DATETIME)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE tickets (id INTEGER PRIMARY KEY, ticket_number TEXT, assigned_to INTEGER, deleted_at DATETIME)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE products (id INTEGER PRIMARY KEY, name TEXT, deleted_at DATETIME)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE google_merchant_offers (id INTEGER PRIMARY KEY, title TEXT, deleted_at DATETIME)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO orders (id, order_number) VALUES (1, 'SALES-1001')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO tickets (id, ticket_number, assigned_to) VALUES (12, 'CHAT-12', 12), (13, 'CHAT-13', 13)`).Error)
	repo := repository.NewOutboxRepository(db)
	dashboardHandler := NewDashboardHandler(nil, createDashboardBusinessFollowUpReadServiceForTestDatabase(db))
	mutationHandler := NewOutboxReconciliationHandler(service.NewOutboxService(repo))
	router := gin.New()
	router.Use(func(c *gin.Context) {
		if role := c.GetHeader("X-Test-Role"); role != "" {
			c.Set("user_role", role)
			c.Set("user_id", uint(12))
		}
		c.Next()
	})
	registerAuthenticatedCoreRoutes(router.Group("/api/admin"), nil, dashboardHandler, nil, nil)
	router.POST("/api/admin/ops/outbox/failures/:id/retry", middleware.RequirePermission(auth.PermOpsView), middleware.RequirePermission(auth.PermSystemManage), mutationHandler.RetryFailure)
	router.POST("/api/admin/ops/outbox/failures/:id/ignore", middleware.RequirePermission(auth.PermOpsView), middleware.RequirePermission(auth.PermSystemManage), mutationHandler.IgnoreFailure)
	return router, repo, db
}

func requestDashboardOutboxFailureClosedLoopRoute(router *gin.Engine, method, path, role, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequestWithContext(context.Background(), method, path, strings.NewReader(body))
	request.Header.Set("X-Test-Role", role)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func readDashboardOutboxFailureClosedLoopSnapshot(t *testing.T, router *gin.Engine) service.DashboardCommandCenterResponse {
	t.Helper()
	recorder := requestDashboardOutboxFailureClosedLoopRoute(router, http.MethodGet, "/api/admin/dashboard/command-center", string(auth.RoleAdmin), "")
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	require.Equal(t, "no-store", recorder.Header().Get("Cache-Control"))
	require.NotContains(t, recorder.Body.String(), "recipient_email")
	require.NotContains(t, recorder.Body.String(), "private-query-error")
	require.NotContains(t, recorder.Body.String(), "payload")
	require.NotContains(t, recorder.Body.String(), "last_error")
	var snapshot service.DashboardCommandCenterResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &snapshot))
	return snapshot
}

func TestDashboardOutboxFailureClosedLoopCountsEntireQueueAndReflectsManualTransitions(t *testing.T) {
	router, repo, _ := createDashboardOutboxFailureClosedLoopTestRouter(t)
	now := time.Now().UTC()
	for index, status := range []string{"failed", "failed", "failed", "failed", "failed", "failed", "dead_letter", "dead_letter", "pending", "processing", "processed", "unknown", "ignored"} {
		require.NoError(t, repo.CreateEvent(&outbox.Event{
			EventKey: fmt.Sprintf("dashboard-outbox-test:%d", index), EventType: outbox.EventTypeOrderPaymentSucceeded,
			AggregateType: "order", AggregateID: fmt.Sprint(index), Status: status,
			Payload: []byte(`{"recipient_email":"private@example.test"}`), LastError: "private-query-error",
			CreatedAt: now.Add(-time.Duration(index) * time.Hour), UpdatedAt: now.Add(-time.Duration(index) * time.Hour),
		}))
	}
	snapshot := readDashboardOutboxFailureClosedLoopSnapshot(t, router)
	failures := snapshot.ActionCenter.OutboxFailures
	require.Equal(t, "ready", failures.Status)
	require.EqualValues(t, 8, *failures.TotalCount)
	require.EqualValues(t, 6, *failures.FailedCount)
	require.EqualValues(t, 2, *failures.DeadLetterCount)
	require.Len(t, failures.Items, 5)
	require.Equal(t, "manual_attention", failures.Items[0].ProcessingState)
	require.Equal(t, "manual_attention", failures.Items[1].ProcessingState)
	require.True(t, failures.Items[0].UpdatedAt.Before(failures.Items[1].UpdatedAt))
	require.Equal(t, "automatic_retry_scheduled", failures.Items[2].ProcessingState)
	require.EqualValues(t, 2, *failures.ManualAttentionCount)
	require.EqualValues(t, 6, *failures.AutomaticRetryCount)
	require.Empty(t, snapshot.SystemDegradedNodes)

	for index, operation := range []string{"retry", "ignore"} {
		eventID := failures.Items[index].ID
		path := fmt.Sprintf("/api/admin/ops/outbox/failures/%d/%s", eventID, operation)
		recorder := requestDashboardOutboxFailureClosedLoopRoute(router, http.MethodPost, path, string(auth.RoleAdmin), `{"note":"checked external delivery state"}`)
		require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
		updated := readDashboardOutboxFailureClosedLoopSnapshot(t, router).ActionCenter.OutboxFailures
		require.EqualValues(t, 7-index, *updated.TotalCount)
		require.EqualValues(t, 1-index, *updated.DeadLetterCount)
		event, err := repo.FindEventByID(eventID)
		require.NoError(t, err)
		if operation == "retry" {
			require.Equal(t, outbox.EventStatusPending, event.Status)
		} else {
			require.Equal(t, outbox.EventStatusIgnored, event.Status)
		}
	}
}

func TestDashboardOutboxFailureSnapshotSeparatesEmptyQueueFromQueryFailure(t *testing.T) {
	router, _, _ := createDashboardOutboxFailureClosedLoopTestRouter(t)
	emptySnapshot := readDashboardOutboxFailureClosedLoopSnapshot(t, router)
	require.Equal(t, "ready", emptySnapshot.ActionCenter.OutboxFailures.Status)
	require.NotNil(t, emptySnapshot.ActionCenter.OutboxFailures.TotalCount)
	require.Zero(t, *emptySnapshot.ActionCenter.OutboxFailures.TotalCount)
	require.NotNil(t, emptySnapshot.ActionCenter.OutboxFailures.Items)

	db := newOutboxReconciliationTestDB(t)
	require.NoError(t, db.Migrator().DropTable(&outbox.Event{}))
	dashboardService := createDashboardBusinessFollowUpReadServiceForTestDatabase(db)
	degraded := dashboardService.GetCommandCenterOutboxFailures(context.Background(), auth.RoleAdmin, 1)
	require.Equal(t, "unavailable", degraded.ActionCenter.OutboxFailures.Status)
	require.Nil(t, degraded.ActionCenter.OutboxFailures.TotalCount)
	require.Nil(t, degraded.ActionCenter.OutboxFailures.FailedCount)
	require.Nil(t, degraded.ActionCenter.OutboxFailures.DeadLetterCount)
	require.Equal(t, []string{"outbox_failures"}, degraded.SystemDegradedNodes)

	missingRepository := service.NewDashboardBusinessFollowUpReadService(nil, nil).GetCommandCenterOutboxFailures(context.Background(), auth.RoleAdmin, 1)
	require.Equal(t, "unavailable", missingRepository.ActionCenter.OutboxFailures.Status)
	require.Nil(t, missingRepository.ActionCenter.OutboxFailures.TotalCount)
}

func TestDashboardOutboxFailureCommandCenterEnforcesReadPermission(t *testing.T) {
	router, _, _ := createDashboardOutboxFailureClosedLoopTestRouter(t)
	for _, permissionCase := range []struct {
		role string
		code int
	}{
		{string(auth.RoleAdmin), http.StatusOK},
		{string(auth.RoleManager), http.StatusOK},
		{string(auth.RoleSupport), http.StatusOK},
		{string(auth.RoleEditor), http.StatusOK},
		{string(auth.RoleUser), http.StatusForbidden},
		{"", http.StatusUnauthorized},
	} {
		t.Run("role="+permissionCase.role, func(t *testing.T) {
			recorder := requestDashboardOutboxFailureClosedLoopRoute(router, http.MethodGet, "/api/admin/dashboard/command-center", permissionCase.role, "")
			require.Equal(t, permissionCase.code, recorder.Code, recorder.Body.String())
			if permissionCase.code != http.StatusOK {
				require.NotContains(t, recorder.Body.String(), "total_count")
			}
		})
	}
}

func TestDashboardOutboxFailureBusinessFollowUpsFilterCountsReferencesAndDirectDetailsByDomainAndAssignedConversation(t *testing.T) {
	router, repo, _ := createDashboardOutboxFailureClosedLoopTestRouter(t)
	events := []*outbox.Event{
		{EventKey: "scoped-order", EventType: outbox.EventTypeOrderPaymentSucceeded, AggregateType: outbox.AggregateTypeOrder, AggregateID: "1", Status: "failed", Attempts: 3, LastError: "timeout private-secret-token"},
		{EventKey: "scoped-product", EventType: outbox.EventTypeProductCacheInvalidate, AggregateType: outbox.AggregateTypeProductCache, AggregateID: "products:1", Status: "dead_letter"},
		{EventKey: "scoped-assigned-ticket", EventType: outbox.EventTypeCustomerServiceRealtime, AggregateType: outbox.AggregateTypeCustomerServiceConversation, AggregateID: "12", Status: "failed", Attempts: 10},
		{EventKey: "scoped-other-ticket", EventType: outbox.EventTypeCustomerServiceRealtime, AggregateType: outbox.AggregateTypeCustomerServiceConversation, AggregateID: "13", Status: "failed"},
		{EventKey: "scoped-unknown", EventType: "unclassified.maintenance", AggregateType: "system", AggregateID: "1", Status: "failed"},
	}
	for _, event := range events {
		require.NoError(t, repo.CreateEvent(event))
	}
	for _, testCase := range []struct {
		role          auth.Role
		total, manual int64
		allowedIDs    []uint
	}{
		{auth.RoleSupport, 2, 1, []uint{events[0].ID, events[2].ID}},
		{auth.RoleEditor, 1, 1, []uint{events[1].ID}},
		{auth.RoleManager, 5, 2, []uint{events[0].ID, events[1].ID, events[2].ID, events[3].ID, events[4].ID}},
	} {
		t.Run(string(testCase.role), func(t *testing.T) {
			for _, path := range []string{"/api/admin/dashboard/command-center", "/api/admin/dashboard/business-follow-ups"} {
				recorder := requestDashboardOutboxFailureClosedLoopRoute(router, http.MethodGet, path, string(testCase.role), "")
				require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
				require.NotContains(t, recorder.Body.String(), "private-secret-token")
				require.NotContains(t, recorder.Body.String(), "event_type")
				var response service.DashboardCommandCenterResponse
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
				failures := response.ActionCenter.OutboxFailures
				require.EqualValues(t, testCase.total, *failures.TotalCount)
				require.EqualValues(t, testCase.manual, *failures.ManualAttentionCount)
				ids := []uint{}
				for _, item := range failures.Items {
					ids = append(ids, item.ID)
				}
				require.ElementsMatch(t, testCase.allowedIDs, ids)
			}
			for _, event := range events {
				recorder := requestDashboardOutboxFailureClosedLoopRoute(router, http.MethodGet, fmt.Sprintf("/api/admin/dashboard/business-follow-ups/%d", event.ID), string(testCase.role), "")
				allowed := false
				for _, id := range testCase.allowedIDs {
					allowed = allowed || id == event.ID
				}
				if !allowed {
					require.Equal(t, http.StatusNotFound, recorder.Code)
					continue
				}
				require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
				var item service.DashboardBusinessFollowUpItem
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &item))
				require.NotEmpty(t, item.BusinessImpact)
				require.NotEmpty(t, item.SalesNextSteps)
				if event.ID == events[0].ID {
					require.Contains(t, item.RelatedRecord.Label, "SALES-1001")
					require.Equal(t, "/orders/list?order_id=1", item.RelatedRecord.Path)
					require.Contains(t, item.BusinessImpact, "勿据此重复收款")
					require.Equal(t, "automatic_retry_scheduled", item.ProcessingState)
					require.NotNil(t, item.NextRetryAt)
				}
				if event.ID == events[2].ID {
					require.Equal(t, "manual_attention", item.ProcessingState)
					require.Nil(t, item.NextRetryAt)
					require.Contains(t, item.RelatedRecord.Path, "conversation_id=12")
				}
			}
		})
	}
	for _, operation := range []string{"retry", "ignore"} {
		recorder := requestDashboardOutboxFailureClosedLoopRoute(router, http.MethodPost, fmt.Sprintf("/api/admin/ops/outbox/failures/%d/%s", events[0].ID, operation), string(auth.RoleSupport), `{"note":"sales checked"}`)
		require.Equal(t, http.StatusForbidden, recorder.Code)
	}
}

func TestDashboardOutboxFailureBusinessReferenceQueryFailurePreservesCountsAndExplainsMissingContext(t *testing.T) {
	db := newOutboxReconciliationTestDB(t)
	repo := repository.NewOutboxRepository(db)
	require.NoError(t, repo.CreateEvent(&outbox.Event{EventKey: "context-failure", EventType: outbox.EventTypeOrderRefunded, AggregateType: outbox.AggregateTypeOrder, AggregateID: "2", Status: "dead_letter"}))
	response := createDashboardBusinessFollowUpReadServiceForTestDatabase(db).GetCommandCenterOutboxFailures(context.Background(), auth.RoleAdmin, 1)
	failures := response.ActionCenter.OutboxFailures
	require.Equal(t, "ready", failures.Status)
	require.EqualValues(t, 1, *failures.TotalCount)
	require.Contains(t, response.SystemDegradedNodes, "outbox_business_references")
	require.Contains(t, failures.Items[0].RelatedRecordNotice, "待办数量仍有效")
	require.Nil(t, failures.Items[0].RelatedRecord)
	require.Contains(t, failures.Items[0].BusinessImpact, "退款失败或到账")
}

func TestDashboardOutboxFailureMismatchedEventTypeCannotExposeAnotherAgentsConversation(t *testing.T) {
	router, repo, _ := createDashboardOutboxFailureClosedLoopTestRouter(t)
	event := &outbox.Event{EventKey: "mismatched-event-type-other-agents-conversation", EventType: outbox.EventTypeOrderPaymentSucceeded, AggregateType: outbox.AggregateTypeCustomerServiceConversation, AggregateID: "13", Status: "failed"}
	require.NoError(t, repo.CreateEvent(event))
	for _, path := range []string{"/api/admin/dashboard/command-center", "/api/admin/dashboard/business-follow-ups"} {
		recorder := requestDashboardOutboxFailureClosedLoopRoute(router, http.MethodGet, path, string(auth.RoleSupport), "")
		require.Equal(t, http.StatusOK, recorder.Code)
		var response service.DashboardCommandCenterResponse
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
		require.Zero(t, *response.ActionCenter.OutboxFailures.TotalCount)
		require.Empty(t, response.ActionCenter.OutboxFailures.Items)
		require.NotContains(t, recorder.Body.String(), "CHAT-13")
	}
	detailPath := fmt.Sprintf("/api/admin/dashboard/business-follow-ups/%d", event.ID)
	require.Equal(t, http.StatusNotFound, requestDashboardOutboxFailureClosedLoopRoute(router, http.MethodGet, detailPath, string(auth.RoleSupport), "").Code)
	managerDetail := requestDashboardOutboxFailureClosedLoopRoute(router, http.MethodGet, detailPath, string(auth.RoleManager), "")
	require.Equal(t, http.StatusOK, managerDetail.Code)
	require.Contains(t, managerDetail.Body.String(), "CHAT-13")
}

func TestDashboardOutboxFailureConversationReferenceRechecksAssignmentAfterSnapshot(t *testing.T) {
	db := newOutboxReconciliationTestDB(t)
	require.NoError(t, db.Exec(`CREATE TABLE tickets (id INTEGER PRIMARY KEY, ticket_number TEXT, assigned_to INTEGER, deleted_at DATETIME)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO tickets (id, ticket_number, assigned_to) VALUES (12, 'REASSIGNED-PRIVATE-TICKET', 12)`).Error)
	repo := repository.NewOutboxRepository(db)
	require.NoError(t, repo.CreateEvent(&outbox.Event{EventKey: "conversation-reassigned-after-snapshot", EventType: outbox.EventTypeCustomerServiceRealtime, AggregateType: outbox.AggregateTypeCustomerServiceConversation, AggregateID: "12", Status: "failed"}))
	agentID := uint(12)
	scope := repository.DashboardOutboxBusinessReadScope{AllowedEventTypes: []string{outbox.EventTypeCustomerServiceRealtime}, AssignedAgentUserID: &agentID, AllowedBusinessAggregateTypes: []string{outbox.AggregateTypeCustomerServiceConversation}}
	snapshot, err := repository.NewDashboardOutboxBusinessFollowUpReadRepository(db).GetDashboardOutboxFailureSnapshot(context.Background(), 100, scope)
	require.NoError(t, err)
	require.Len(t, snapshot.Items, 1)
	resolver := service.NewDashboardBusinessFollowUpBusinessReferenceResolver(repository.NewDashboardBusinessFollowUpBusinessIdentityReadRepository(db))
	resolution := resolver.ResolveDashboardBusinessFollowUpReferences(context.Background(), snapshot.Items, scope)
	require.Empty(t, resolution.DegradedBusinessAggregateTypes)
	require.Len(t, resolution.References, 1)
	require.NoError(t, db.Exec(`UPDATE tickets SET assigned_to = 13 WHERE id = 12`).Error)
	resolution = resolver.ResolveDashboardBusinessFollowUpReferences(context.Background(), snapshot.Items, scope)
	require.Empty(t, resolution.DegradedBusinessAggregateTypes)
	require.Empty(t, resolution.References)
	managerResolution := resolver.ResolveDashboardBusinessFollowUpReferences(context.Background(), snapshot.Items, repository.DashboardOutboxBusinessReadScope{CanViewAllEventTypes: true, AllowedBusinessAggregateTypes: []string{outbox.AggregateTypeCustomerServiceConversation}})
	require.Empty(t, managerResolution.DegradedBusinessAggregateTypes)
	require.Len(t, managerResolution.References, 1)
}

func TestDashboardOutboxFailureBusinessReferencesUseOriginalRecordsAndRespectSoftDeletion(t *testing.T) {
	db := newOutboxReconciliationTestDB(t)
	for _, statement := range []string{
		`CREATE TABLE orders (id INTEGER PRIMARY KEY, order_number TEXT, deleted_at DATETIME)`,
		`CREATE TABLE after_sales_cases (id INTEGER PRIMARY KEY, order_id INTEGER, deleted_at DATETIME)`,
		`CREATE TABLE refunds (id INTEGER PRIMARY KEY, order_id INTEGER, deleted_at DATETIME)`,
		`CREATE TABLE products (id INTEGER PRIMARY KEY, name TEXT, deleted_at DATETIME)`,
		`CREATE TABLE google_merchant_offers (id INTEGER PRIMARY KEY, title TEXT, deleted_at DATETIME)`,
		`INSERT INTO orders (id, order_number) VALUES (1, 'ACTUAL-1001')`,
		`INSERT INTO orders (id, order_number, deleted_at) VALUES (2, 'DELETED-1002', CURRENT_TIMESTAMP)`,
		`INSERT INTO after_sales_cases (id, order_id) VALUES (11, 1)`,
		`INSERT INTO refunds (id, order_id) VALUES (21, 1)`,
		`INSERT INTO products (id, name) VALUES (31, 'Verified product')`,
		`INSERT INTO google_merchant_offers (id, title) VALUES (41, 'Verified channel offer')`,
	} {
		require.NoError(t, db.Exec(statement).Error)
	}
	repo := repository.NewOutboxRepository(db)
	events := []*outbox.Event{
		{EventKey: "reference-after-sales", EventType: outbox.EventTypeAfterSalesStatusChanged, AggregateType: outbox.AggregateTypeAfterSalesCase, AggregateID: "11", Status: "failed"},
		{EventKey: "reference-refund", EventType: outbox.EventTypePaymentRefundExecutionRequested, AggregateType: outbox.AggregateTypePayment, AggregateID: "21", Status: "failed"},
		{EventKey: "reference-product", EventType: outbox.EventTypeMerchantProductUpsert, AggregateType: outbox.AggregateTypeProduct, AggregateID: "31", Status: "failed"},
		{EventKey: "reference-offer", EventType: outbox.EventTypeMerchantOfferRevalidate, AggregateType: outbox.AggregateTypeMerchantOffer, AggregateID: "41", Status: "failed"},
		{EventKey: "reference-deleted-order", EventType: outbox.EventTypeOrderRefunded, AggregateType: outbox.AggregateTypeOrder, AggregateID: "2", Status: "failed"},
	}
	for _, event := range events {
		event.Payload = []byte(`{"order_number":"UNTRUSTED-PAYLOAD-ORDER","customer_email":"private@example.test"}`)
		require.NoError(t, repo.CreateEvent(event))
	}
	dashboardService := createDashboardBusinessFollowUpReadServiceForTestDatabase(db)
	response := dashboardService.GetDashboardBusinessFollowUpList(context.Background(), auth.RoleAdmin, 1)
	require.Empty(t, response.SystemDegradedNodes)
	require.EqualValues(t, 5, *response.ActionCenter.OutboxFailures.TotalCount)
	for _, event := range events {
		item, err := dashboardService.GetDashboardBusinessFollowUpDetail(context.Background(), auth.RoleAdmin, 1, event.ID)
		require.NoError(t, err)
		serialized, err := json.Marshal(item)
		require.NoError(t, err)
		require.NotContains(t, string(serialized), "UNTRUSTED-PAYLOAD")
		require.NotContains(t, string(serialized), "private@example.test")
		if event.ID == events[0].ID || event.ID == events[1].ID {
			require.Contains(t, item.RelatedRecord.Label, "ACTUAL-1001")
			require.Equal(t, "/orders/list?order_id=1", item.RelatedRecord.Path)
		}
		if event.ID == events[2].ID || event.ID == events[3].ID {
			require.Contains(t, item.RelatedRecord.Label, "Verified")
			require.Equal(t, "/google-merchant", item.RelatedRecord.Path)
		}
		if event.ID == events[4].ID {
			require.Nil(t, item.RelatedRecord)
			require.NotEmpty(t, item.RelatedRecordNotice)
		}
	}
}

func TestDashboardOutboxFailureBusinessListBoundsPreviewAndDetailRejectsChangedStateOrInvalidID(t *testing.T) {
	router, repo, db := createDashboardOutboxFailureClosedLoopTestRouter(t)
	var lastEventID uint
	for index := 0; index < 105; index++ {
		event := &outbox.Event{EventKey: fmt.Sprintf("bounded-list:%d", index), EventType: outbox.EventTypeOrderPaymentSucceeded, AggregateType: outbox.AggregateTypeOrder, AggregateID: "1", Status: "failed"}
		require.NoError(t, repo.CreateEvent(event))
		lastEventID = event.ID
	}
	recorder := requestDashboardOutboxFailureClosedLoopRoute(router, http.MethodGet, "/api/admin/dashboard/business-follow-ups", string(auth.RoleSupport), "")
	require.Equal(t, http.StatusOK, recorder.Code)
	var response service.DashboardCommandCenterResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.EqualValues(t, 105, *response.ActionCenter.OutboxFailures.TotalCount)
	require.Len(t, response.ActionCenter.OutboxFailures.Items, 100)
	recorder = requestDashboardOutboxFailureClosedLoopRoute(router, http.MethodGet, fmt.Sprintf("/api/admin/dashboard/business-follow-ups/%d", lastEventID), string(auth.RoleSupport), "")
	require.Equal(t, http.StatusOK, recorder.Code)
	recorder = requestDashboardOutboxFailureClosedLoopRoute(router, http.MethodPost, fmt.Sprintf("/api/admin/ops/outbox/failures/%d/retry", lastEventID), string(auth.RoleAdmin), `{"note":"checked actual result"}`)
	require.Equal(t, http.StatusOK, recorder.Code)
	recorder = requestDashboardOutboxFailureClosedLoopRoute(router, http.MethodGet, fmt.Sprintf("/api/admin/dashboard/business-follow-ups/%d", lastEventID), string(auth.RoleSupport), "")
	require.Equal(t, http.StatusNotFound, recorder.Code)
	for _, invalidID := range []string{"0", "invalid", "-1", "4294967296"} {
		recorder = requestDashboardOutboxFailureClosedLoopRoute(router, http.MethodGet, "/api/admin/dashboard/business-follow-ups/"+invalidID, string(auth.RoleSupport), "")
		require.Equal(t, http.StatusBadRequest, recorder.Code)
	}
	require.NoError(t, repo.CreateEvent(&outbox.Event{EventKey: "missing-identity-assigned-conversation", EventType: outbox.EventTypeCustomerServiceRealtime, AggregateType: outbox.AggregateTypeCustomerServiceConversation, AggregateID: "12", Status: "failed"}))
	missingIdentitySnapshot, err := repository.NewDashboardOutboxBusinessFollowUpReadRepository(db).GetDashboardOutboxFailureSnapshot(context.Background(), 100, repository.DashboardOutboxBusinessReadScope{AllowedEventTypes: []string{outbox.EventTypeCustomerServiceRealtime}, AssignedAgentUserID: new(uint)})
	require.NoError(t, err)
	require.Zero(t, missingIdentitySnapshot.TotalCount)
}

func TestDashboardOutboxFailureSnapshotExcludesDeletedEventsAndHonorsCancellation(t *testing.T) {
	db := newOutboxReconciliationTestDB(t)
	repo := repository.NewOutboxRepository(db)
	event := &outbox.Event{EventKey: "deleted-failure", EventType: "order.paid", AggregateType: "order", AggregateID: "1", Status: outbox.EventStatusFailed}
	require.NoError(t, repo.CreateEvent(event))
	require.NoError(t, db.Delete(event).Error)
	snapshot, err := repository.NewDashboardOutboxBusinessFollowUpReadRepository(db).GetDashboardOutboxFailureSnapshot(context.Background(), 5, repository.DashboardOutboxBusinessReadScope{CanViewAllEventTypes: true})
	require.NoError(t, err)
	require.Zero(t, snapshot.TotalCount)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = repository.NewDashboardOutboxBusinessFollowUpReadRepository(db).GetDashboardOutboxFailureSnapshot(ctx, 5, repository.DashboardOutboxBusinessReadScope{CanViewAllEventTypes: true})
	require.ErrorIs(t, err, context.Canceled)
}
