package admin

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/securecookie"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Opt in with DASHBOARD_OUTBOX_POSTGRES_CONFIG_FILE or DB_HOST/DB_USERNAME/DB_NAME.
// Credentials are read in process and never included in verification output.
func openConfiguredPostgreSQLForDashboardBusinessFollowUpVerification(t *testing.T) *gorm.DB {
	t.Helper()
	configurationPath := os.Getenv("DASHBOARD_OUTBOX_POSTGRES_CONFIG_FILE")
	if configurationPath == "" && os.Getenv("DB_HOST") == "" {
		t.Skip("PostgreSQL dashboard verification is not configured")
	}
	settings := viper.New()
	if configurationPath != "" {
		settings.SetConfigFile(configurationPath)
		if err := settings.ReadInConfig(); err != nil {
			t.Fatal("Cannot read the configured PostgreSQL settings")
		}
	}
	settings.SetDefault("database.port", 5432)
	for settingName, environmentName := range map[string]string{"host": "DB_HOST", "port": "DB_PORT", "database": "DB_NAME", "username": "DB_USERNAME", "password": "DB_PASSWORD"} {
		require.NoError(t, settings.BindEnv("database."+settingName, environmentName))
	}
	databaseConfiguration := config.DatabaseConfig{
		Host: settings.GetString("database.host"), Port: settings.GetInt("database.port"),
		Database: settings.GetString("database.database"), Username: settings.GetString("database.username"),
		Password: settings.GetString("database.password"),
	}
	if databaseConfiguration.Host == "" || databaseConfiguration.Username == "" || databaseConfiguration.Database == "" {
		t.Fatal("PostgreSQL host, database and user settings are required")
	}
	connectionURL := &url.URL{Scheme: "postgres", User: url.UserPassword(databaseConfiguration.Username, databaseConfiguration.Password), Host: net.JoinHostPort(databaseConfiguration.Host, strconv.Itoa(databaseConfiguration.Port)), Path: "/" + databaseConfiguration.Database, RawQuery: "sslmode=disable&connect_timeout=5"}
	db, err := gorm.Open(postgres.Open(connectionURL.String()), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
	if err != nil {
		failureCategory := "connection unavailable"
		switch {
		case strings.Contains(err.Error(), "password authentication failed"):
			failureCategory = "authentication rejected"
		case strings.Contains(err.Error(), "does not exist"):
			failureCategory = "configured database does not exist"
		case strings.Contains(err.Error(), "connectex"), strings.Contains(err.Error(), "connection refused"):
			failureCategory = "database address is not accepting connections"
		}
		t.Fatalf("Cannot connect to configured PostgreSQL: %s; connection details omitted", failureCategory)
	}
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(2)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	return db
}

func TestDashboardOutboxPostgreSQLExistingBusinessDataQueriesAreReadOnlyAndConsistent(t *testing.T) {
	// Generic CI database settings also enable the isolated-schema test below.
	// Inspect existing business tables only when that separate check is requested.
	if os.Getenv("DASHBOARD_OUTBOX_POSTGRES_CONFIG_FILE") == "" && os.Getenv("DASHBOARD_OUTBOX_VERIFY_EXISTING_BUSINESS_DATA") != "true" {
		t.Skip("Existing business data verification requires DASHBOARD_OUTBOX_POSTGRES_CONFIG_FILE or DASHBOARD_OUTBOX_VERIFY_EXISTING_BUSINESS_DATA=true")
	}
	db := openConfiguredPostgreSQLForDashboardBusinessFollowUpVerification(t)
	transaction := db.Begin(&sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	require.NoError(t, transaction.Error)
	t.Cleanup(func() { require.NoError(t, transaction.Rollback().Error) })
	dashboard := createDashboardBusinessFollowUpReadServiceForTestDatabase(transaction)
	var expectedCounts struct{ Total, Manual int64 }
	require.NoError(t, transaction.Raw(`SELECT COUNT(*) AS total, COUNT(*) FILTER (WHERE status = 'dead_letter' OR attempts >= max_attempts) AS manual FROM outbox_events WHERE status IN ('failed', 'dead_letter') AND deleted_at IS NULL`).Scan(&expectedCounts).Error)
	for _, role := range []auth.Role{auth.RoleAdmin, auth.RoleManager, auth.RoleEditor, auth.RoleSupport} {
		response := dashboard.GetDashboardBusinessFollowUpList(t.Context(), role, 0)
		failures := response.ActionCenter.OutboxFailures
		require.Equal(t, "ready", failures.Status, "PostgreSQL business query must be available")
		require.Empty(t, response.SystemDegradedNodes, "Existing business references must be readable")
		require.Equal(t, *failures.TotalCount, *failures.FailedCount+*failures.DeadLetterCount)
		require.Equal(t, *failures.TotalCount, *failures.ManualAttentionCount+*failures.AutomaticRetryCount)
		if role.HasPermission(auth.PermOpsView) {
			require.Equal(t, expectedCounts.Total, *failures.TotalCount)
			require.Equal(t, expectedCounts.Manual, *failures.ManualAttentionCount)
		}
		for _, item := range failures.Items {
			if item.RelatedRecord != nil && item.RelatedRecord.RequiredPermission != "" {
				require.True(t, role.HasPermission(auth.Permission(item.RelatedRecord.RequiredPermission)))
			}
		}
	}
	t.Log("Existing PostgreSQL dashboard queries passed inside a read-only transaction; no business identities were logged")
}

func createRollbackOnlyPostgreSQLDashboardBusinessFollowUpDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	db := openConfiguredPostgreSQLForDashboardBusinessFollowUpVerification(t)
	transaction := db.Begin()
	require.NoError(t, transaction.Error)
	schemaName := "test_dashboard_business_follow_up_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	t.Cleanup(func() {
		require.NoError(t, transaction.Rollback().Error)
		var remainingSchemas int64
		require.NoError(t, db.Raw("SELECT COUNT(*) FROM pg_namespace WHERE nspname = ?", schemaName).Scan(&remainingSchemas).Error)
		require.Zero(t, remainingSchemas, "Verification schema must disappear after rollback")
	})
	require.NoError(t, transaction.Exec("CREATE SCHEMA "+schemaName).Error)
	// No public fallback: every test mutation must resolve inside this schema.
	require.NoError(t, transaction.Exec("SET LOCAL search_path TO "+schemaName).Error)
	require.NoError(t, transaction.AutoMigrate(&outbox.Event{}, &user.User{}))
	for _, statement := range []string{
		`CREATE TABLE orders (id BIGINT PRIMARY KEY, order_number TEXT, deleted_at TIMESTAMPTZ)`,
		`CREATE TABLE tickets (id BIGINT PRIMARY KEY, ticket_number TEXT, assigned_to BIGINT, deleted_at TIMESTAMPTZ)`,
		`CREATE TABLE products (id BIGINT PRIMARY KEY, name TEXT, deleted_at TIMESTAMPTZ)`,
		`CREATE TABLE google_merchant_offers (id BIGINT PRIMARY KEY, title TEXT, deleted_at TIMESTAMPTZ)`,
		`CREATE TABLE refunds (id BIGINT PRIMARY KEY, order_id BIGINT, deleted_at TIMESTAMPTZ)`,
		`CREATE TABLE after_sales_cases (id BIGINT PRIMARY KEY, order_id BIGINT, deleted_at TIMESTAMPTZ)`,
		`INSERT INTO orders VALUES (1, 'POSTGRES-REVIEW-ORDER-1', NULL)`,
		`INSERT INTO tickets VALUES (12, 'POSTGRES-OWN-CONVERSATION', 12, NULL), (13, 'POSTGRES-PRIVATE-CONVERSATION', 13, NULL)`,
		`INSERT INTO refunds VALUES (21, 1, NULL)`,
		`INSERT INTO after_sales_cases VALUES (31, 1, NULL)`,
		`INSERT INTO products VALUES (41, 'PostgreSQL verified product', NULL)`,
		`INSERT INTO google_merchant_offers VALUES (51, 'PostgreSQL verified offer', NULL)`,
	} {
		require.NoError(t, transaction.Exec(statement).Error)
	}
	return transaction
}

func TestDashboardOutboxPostgreSQLAuthenticatedHTTPFailureRetryDispatchAndRecoveryClosedLoop(t *testing.T) {
	db := createRollbackOnlyPostgreSQLDashboardBusinessFollowUpDatabase(t)
	repo := repository.NewOutboxRepository(db)
	outboxService := service.NewOutboxService(repo)
	authService := service.NewAuthService(repository.NewUserRepository(db), config.JWTConfig{Secret: uuid.NewString(), ExpireHours: 1})
	tokens := map[auth.Role]string{}
	for role, id := range map[auth.Role]uint{auth.RoleAdmin: 1, auth.RoleManager: 2, auth.RoleViewer: 3, auth.RoleEditor: 5, auth.RoleSupport: 12, auth.RoleUser: 14} {
		account := &user.User{ID: id, Email: string(role) + "@outbox-review.invalid", Username: "outbox-review-" + string(role), Password: "isolated-unused-password", Role: string(role), Status: "active"}
		require.NoError(t, db.Create(account).Error)
		token, err := authService.GenerateToken(account)
		require.NoError(t, err)
		tokens[role] = token
	}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	adminGroup := router.Group("/api/admin", middleware.CSRFProtection(nil, securecookie.AdminCookieNames()))
	authenticated := adminGroup.Group("", middleware.AuthMiddleware(authService, securecookie.AdminCookieNames()), middleware.RequireBackofficeAccess())
	registerAuthenticatedCoreRoutes(authenticated, nil, NewDashboardHandler(nil, createDashboardBusinessFollowUpReadServiceForTestDatabase(db)), nil, nil)
	registerOperationsRoutes(authenticated, nil, nil, nil, NewOutboxReconciliationHandler(outboxService), nil, nil, nil, nil, nil, nil, nil)
	apiServer := httptest.NewServer(router)
	t.Cleanup(apiServer.Close)
	requestAuthenticatedBusinessFollowUp := func(method, path string, role auth.Role, includeCSRF bool) *httptest.ResponseRecorder {
		request, err := http.NewRequestWithContext(t.Context(), method, apiServer.URL+path, strings.NewReader(`{"note":"Verified receiver recovery in rollback-only integration"}`))
		require.NoError(t, err)
		request.Header.Set("Content-Type", "application/json")
		if token := tokens[role]; token != "" {
			request.AddCookie(&http.Cookie{Name: securecookie.AdminAuthTokenCookie, Value: token})
		}
		if includeCSRF {
			request.AddCookie(&http.Cookie{Name: securecookie.AdminCSRFTokenCookie, Value: "isolated-review-csrf"})
			request.Header.Set(securecookie.CSRFTokenHeader, "isolated-review-csrf")
		}
		response, err := apiServer.Client().Do(request)
		require.NoError(t, err)
		defer func() { require.NoError(t, response.Body.Close()) }()
		recordedResponse := httptest.NewRecorder()
		recordedResponse.Code = response.StatusCode
		for name, values := range response.Header {
			recordedResponse.Header()[name] = values
		}
		_, err = io.Copy(recordedResponse.Body, response.Body)
		require.NoError(t, err)
		return recordedResponse
	}
	readAuthenticatedBusinessFollowUpSnapshot := func(role auth.Role) service.DashboardCommandCenterResponse {
		response := requestAuthenticatedBusinessFollowUp(http.MethodGet, "/api/admin/dashboard/business-follow-ups", role, false)
		require.Equal(t, http.StatusOK, response.Code)
		require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
		var snapshot service.DashboardCommandCenterResponse
		require.NoError(t, json.NewDecoder(response.Body).Decode(&snapshot))
		require.Empty(t, snapshot.SystemDegradedNodes)
		return snapshot
	}
	for _, authorizationCase := range []struct {
		role auth.Role
		code int
	}{{"", http.StatusUnauthorized}, {auth.RoleViewer, http.StatusForbidden}, {auth.RoleUser, http.StatusForbidden}} {
		require.Equal(t, authorizationCase.code, requestAuthenticatedBusinessFollowUp(http.MethodGet, "/api/admin/dashboard/command-center", authorizationCase.role, false).Code)
	}
	require.Zero(t, *readAuthenticatedBusinessFollowUpSnapshot(auth.RoleSupport).ActionCenter.OutboxFailures.TotalCount)
	var receiverRecovered atomic.Bool
	var receiverFailureStatus atomic.Int32
	receiverFailureStatus.Store(http.StatusUnauthorized)
	var successfullyReceived atomic.Int32
	var invalidIdempotencyKeys atomic.Int32
	const reviewEventKey = "order.paid:postgresql-business-follow-up-review"
	const reviewUnknownEventKey = "order.paid:postgresql-uncertain-outcome-review"
	receiver := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Idempotency-Key") != reviewEventKey && request.Header.Get("Idempotency-Key") != reviewUnknownEventKey {
			invalidIdempotencyKeys.Add(1)
		}
		if !receiverRecovered.Load() {
			writer.WriteHeader(int(receiverFailureStatus.Load()))
			return
		}
		successfullyReceived.Add(1)
		writer.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(receiver.Close)
	t.Setenv("ORDER_PAID_OUTBOX_WEBHOOK_URL", receiver.URL)
	t.Setenv("ORDER_PAID_OUTBOX_WEBHOOK_TOKEN", "isolated-review-token")
	outboxService.RegisterHandler(outbox.EventTypeOrderPaid, service.NewOrderPaidOutboxWebhookHandlerFromEnv().Handle)
	now := time.Now().UTC()
	event := &outbox.Event{EventKey: reviewEventKey, EventType: outbox.EventTypeOrderPaid, AggregateType: outbox.AggregateTypeOrder, AggregateID: "1", Payload: []byte(`{"order_id":1}`), MaxAttempts: 1, AvailableAt: now.Add(-time.Minute)}
	require.NoError(t, repo.CreateEvent(event))
	failedDispatch, err := outboxService.ProcessPending(t.Context(), now, 5)
	require.NoError(t, err)
	require.Equal(t, 1, failedDispatch.DeadLetter)
	require.Zero(t, successfullyReceived.Load())
	failures := readAuthenticatedBusinessFollowUpSnapshot(auth.RoleSupport).ActionCenter.OutboxFailures
	require.EqualValues(t, 1, *failures.TotalCount)
	require.EqualValues(t, 1, *failures.ManualAttentionCount)
	require.Equal(t, "manual_attention", failures.Items[0].ProcessingState)
	require.Contains(t, failures.Items[0].BusinessImpact, "下游系统")
	require.Equal(t, "/orders/list?order_id=1", failures.Items[0].RelatedRecord.Path)
	retryPath := fmt.Sprintf("/api/admin/ops/outbox/failures/%d/retry", event.ID)
	require.Equal(t, http.StatusForbidden, requestAuthenticatedBusinessFollowUp(http.MethodPost, retryPath, auth.RoleSupport, true).Code)
	require.Equal(t, http.StatusForbidden, requestAuthenticatedBusinessFollowUp(http.MethodPost, retryPath, auth.RoleAdmin, false).Code)
	receiverRecovered.Store(true)
	require.Equal(t, http.StatusOK, requestAuthenticatedBusinessFollowUp(http.MethodPost, retryPath, auth.RoleAdmin, true).Code)
	require.Zero(t, *readAuthenticatedBusinessFollowUpSnapshot(auth.RoleSupport).ActionCenter.OutboxFailures.TotalCount)
	requeuedEvent, err := repo.FindEventByID(event.ID)
	require.NoError(t, err)
	require.Equal(t, outbox.EventStatusPending, requeuedEvent.Status)
	require.Zero(t, successfullyReceived.Load(), "Leaving the error list cannot claim delivery success")
	recoveredDispatch, err := outboxService.ProcessPending(t.Context(), now.Add(time.Minute), 5)
	require.NoError(t, err)
	require.Equal(t, 1, recoveredDispatch.Processed)
	require.EqualValues(t, 1, successfullyReceived.Load())
	require.Zero(t, invalidIdempotencyKeys.Load())
	processedEvent, err := repo.FindEventByID(event.ID)
	require.NoError(t, err)
	require.Equal(t, outbox.EventStatusProcessed, processedEvent.Status)
	require.NotNil(t, processedEvent.ProcessedAt)
	require.Zero(t, *readAuthenticatedBusinessFollowUpSnapshot(auth.RoleSupport).ActionCenter.OutboxFailures.TotalCount)
	require.Equal(t, http.StatusNotFound, requestAuthenticatedBusinessFollowUp(http.MethodGet, fmt.Sprintf("/api/admin/dashboard/business-follow-ups/%d", event.ID), auth.RoleSupport, false).Code)

	// A server error on an external write may have applied. The real dispatcher
	// keeps that result unknown; it must not offer the ordinary retry operation.
	receiverRecovered.Store(false)
	receiverFailureStatus.Store(http.StatusServiceUnavailable)
	unknownEvent := &outbox.Event{EventKey: reviewUnknownEventKey, EventType: outbox.EventTypeOrderPaid, AggregateType: outbox.AggregateTypeOrder, AggregateID: "1", Payload: []byte(`{"order_id":1}`), MaxAttempts: 1, AvailableAt: now.Add(-time.Minute)}
	require.NoError(t, repo.CreateEvent(unknownEvent))
	uncertainDispatch, err := outboxService.ProcessPending(t.Context(), now.Add(2*time.Minute), 5)
	require.NoError(t, err)
	require.Equal(t, 1, uncertainDispatch.Unknown)
	storedUnknownEvent, err := repo.FindEventByID(unknownEvent.ID)
	require.NoError(t, err)
	require.Equal(t, outbox.EventStatusUnknown, storedUnknownEvent.Status)
	require.Nil(t, storedUnknownEvent.ProcessedAt)
	require.Zero(t, *readAuthenticatedBusinessFollowUpSnapshot(auth.RoleSupport).ActionCenter.OutboxFailures.TotalCount)
	require.Equal(t, http.StatusConflict, requestAuthenticatedBusinessFollowUp(http.MethodPost, fmt.Sprintf("/api/admin/ops/outbox/failures/%d/retry", unknownEvent.ID), auth.RoleAdmin, true).Code)
	nextDispatch, err := outboxService.ProcessPending(t.Context(), now.Add(3*time.Minute), 5)
	require.NoError(t, err)
	require.Zero(t, nextDispatch.Claimed)

	// Verify PostgreSQL assignment casts and every authoritative lookup domain.
	for index, referenceEvent := range []*outbox.Event{
		{EventType: outbox.EventTypeCustomerServiceRealtime, AggregateType: outbox.AggregateTypeCustomerServiceConversation, AggregateID: "12"},
		{EventType: outbox.EventTypeCustomerServiceRealtime, AggregateType: outbox.AggregateTypeCustomerServiceConversation, AggregateID: "13"},
		{EventType: outbox.EventTypePaymentRefundExecutionRequested, AggregateType: outbox.AggregateTypePayment, AggregateID: "21"},
		{EventType: outbox.EventTypeAfterSalesStatusChanged, AggregateType: outbox.AggregateTypeAfterSalesCase, AggregateID: "31"},
		{EventType: outbox.EventTypeMerchantProductUpsert, AggregateType: outbox.AggregateTypeProduct, AggregateID: "41"},
		{EventType: outbox.EventTypeMerchantOfferRevalidate, AggregateType: outbox.AggregateTypeMerchantOffer, AggregateID: "51"},
	} {
		referenceEvent.EventKey = fmt.Sprintf("postgresql-review-reference:%d", index)
		referenceEvent.Status = outbox.EventStatusFailed
		referenceEvent.AvailableAt = now.Add(10 * time.Hour)
		require.NoError(t, repo.CreateEvent(referenceEvent))
	}
	salesSnapshot := readAuthenticatedBusinessFollowUpSnapshot(auth.RoleSupport)
	require.EqualValues(t, 3, *salesSnapshot.ActionCenter.OutboxFailures.TotalCount)
	serializedSalesSnapshot, err := json.Marshal(salesSnapshot)
	require.NoError(t, err)
	require.NotContains(t, string(serializedSalesSnapshot), "POSTGRES-PRIVATE-CONVERSATION")
	require.NotContains(t, string(serializedSalesSnapshot), "last_error")
	require.NotContains(t, string(serializedSalesSnapshot), "payload")
	managerSnapshot := readAuthenticatedBusinessFollowUpSnapshot(auth.RoleManager)
	require.EqualValues(t, 6, *managerSnapshot.ActionCenter.OutboxFailures.TotalCount)
	for _, item := range managerSnapshot.ActionCenter.OutboxFailures.Items {
		require.NotNil(t, item.RelatedRecord)
	}
	require.EqualValues(t, 2, *readAuthenticatedBusinessFollowUpSnapshot(auth.RoleEditor).ActionCenter.OutboxFailures.TotalCount)
	// A persisted account change must take effect even with an old signed cookie.
	require.NoError(t, db.Model(&user.User{}).Where("id = 12").Update("role", string(auth.RoleViewer)).Error)
	require.Equal(t, http.StatusForbidden, requestAuthenticatedBusinessFollowUp(http.MethodGet, "/api/admin/dashboard/business-follow-ups", auth.RoleSupport, false).Code)
	require.NoError(t, db.Model(&user.User{}).Where("id = 12").Updates(map[string]any{"role": string(auth.RoleSupport), "status": "suspended"}).Error)
	require.Equal(t, http.StatusUnauthorized, requestAuthenticatedBusinessFollowUp(http.MethodGet, "/api/admin/dashboard/business-follow-ups", auth.RoleSupport, false).Code)
	t.Log("Real PostgreSQL, signed admin cookies, CSRF, assigned conversations, HTTP dispatch failure/recovery and read-only sales loop passed; isolated schema is rolled back")
}
