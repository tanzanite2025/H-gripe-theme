package admin

import (
	"context"
	"testing"

	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestDashboardBusinessFollowUpDefaultReadScopeDeniesSnapshotDetailAndIdentity(t *testing.T) {
	db := newOutboxReconciliationTestDB(t)
	require.NoError(t, db.Exec(`CREATE TABLE orders (id INTEGER PRIMARY KEY, order_number TEXT, deleted_at DATETIME)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO orders (id, order_number) VALUES (1, 'DEFAULT-SCOPE-PRIVATE')`).Error)
	event := &outbox.Event{EventKey: "default-scope-denies", EventType: outbox.EventTypeOrderPaid, AggregateType: outbox.AggregateTypeOrder, AggregateID: "1", Status: "failed"}
	require.NoError(t, repository.NewOutboxRepository(db).CreateEvent(event))
	reader := repository.NewDashboardOutboxBusinessFollowUpReadRepository(db)
	snapshot, err := reader.GetDashboardOutboxFailureSnapshot(t.Context(), 5, repository.DashboardOutboxBusinessReadScope{})
	require.NoError(t, err)
	require.Zero(t, snapshot.TotalCount)
	require.Empty(t, snapshot.Items)
	_, err = reader.FindDashboardBusinessFollowUpFailureByID(t.Context(), event.ID, repository.DashboardOutboxBusinessReadScope{})
	require.ErrorIs(t, err, repository.ErrOutboxEventNotFound)
	identities, err := repository.NewDashboardBusinessFollowUpBusinessIdentityReadRepository(db).ReadDashboardBusinessIdentitiesByAggregateType(t.Context(), outbox.AggregateTypeOrder, []uint{1}, repository.DashboardOutboxBusinessReadScope{})
	require.NoError(t, err)
	require.Empty(t, identities)
}

func TestDashboardBusinessFollowUpMissingConversationIdentityDeniesReadEvenWithAllowedEvent(t *testing.T) {
	db := newOutboxReconciliationTestDB(t)
	require.NoError(t, db.Exec(`CREATE TABLE tickets (id INTEGER PRIMARY KEY, ticket_number TEXT, assigned_to INTEGER, deleted_at DATETIME)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO tickets VALUES (12, 'MISSING-AGENT-PRIVATE', 12, NULL)`).Error)
	event := &outbox.Event{EventKey: "missing-agent-scope", EventType: outbox.EventTypeCustomerServiceRealtime, AggregateType: outbox.AggregateTypeCustomerServiceConversation, AggregateID: "12", Status: "failed"}
	require.NoError(t, repository.NewOutboxRepository(db).CreateEvent(event))
	scope := repository.DashboardOutboxBusinessReadScope{AllowedEventTypes: []string{outbox.EventTypeCustomerServiceRealtime}, AllowedBusinessAggregateTypes: []string{outbox.AggregateTypeCustomerServiceConversation}}
	snapshot, err := repository.NewDashboardOutboxBusinessFollowUpReadRepository(db).GetDashboardOutboxFailureSnapshot(t.Context(), 5, scope)
	require.NoError(t, err)
	require.Zero(t, snapshot.TotalCount)
	identities, err := repository.NewDashboardBusinessFollowUpBusinessIdentityReadRepository(db).ReadDashboardBusinessIdentitiesByAggregateType(t.Context(), outbox.AggregateTypeCustomerServiceConversation, []uint{12}, scope)
	require.NoError(t, err)
	require.Empty(t, identities)
}

func TestDashboardBusinessFollowUpPartialReferenceFailureKeepsOtherDomainsAndCounts(t *testing.T) {
	for _, missingDomain := range []string{"products", "orders"} {
		t.Run(missingDomain, func(t *testing.T) {
			db := newOutboxReconciliationTestDB(t)
			for _, statement := range []string{
				`CREATE TABLE orders (id INTEGER PRIMARY KEY, order_number TEXT, deleted_at DATETIME)`,
				`CREATE TABLE products (id INTEGER PRIMARY KEY, name TEXT, deleted_at DATETIME)`,
				`CREATE TABLE tickets (id INTEGER PRIMARY KEY, ticket_number TEXT, assigned_to INTEGER, deleted_at DATETIME)`,
				`CREATE TABLE refunds (id INTEGER PRIMARY KEY, order_id INTEGER, deleted_at DATETIME)`,
				`INSERT INTO orders VALUES (1, 'PARTIAL-ORDER', NULL)`,
				`INSERT INTO products VALUES (31, 'Partial product', NULL)`,
				`INSERT INTO tickets VALUES (12, 'PARTIAL-CONVERSATION', 12, NULL)`,
				`INSERT INTO refunds VALUES (21, 1, NULL)`,
			} {
				require.NoError(t, db.Exec(statement).Error)
			}
			events := []*outbox.Event{
				{EventKey: "partial-order", EventType: outbox.EventTypeOrderPaid, AggregateType: outbox.AggregateTypeOrder, AggregateID: "1", Status: "failed"},
				{EventKey: "partial-product", EventType: outbox.EventTypeMerchantProductUpsert, AggregateType: outbox.AggregateTypeProduct, AggregateID: "31", Status: "failed"},
				{EventKey: "partial-conversation", EventType: outbox.EventTypeCustomerServiceRealtime, AggregateType: outbox.AggregateTypeCustomerServiceConversation, AggregateID: "12", Status: "failed"},
				{EventKey: "partial-refund", EventType: outbox.EventTypePaymentRefundExecutionRequested, AggregateType: outbox.AggregateTypePayment, AggregateID: "21", Status: "failed"},
			}
			for _, event := range events {
				require.NoError(t, repository.NewOutboxRepository(db).CreateEvent(event))
			}
			require.NoError(t, db.Migrator().DropTable(missingDomain))
			reader := createDashboardBusinessFollowUpReadServiceForTestDatabase(db)
			response := reader.GetDashboardBusinessFollowUpList(context.Background(), auth.RoleAdmin, 1)
			require.Equal(t, "ready", response.ActionCenter.OutboxFailures.Status)
			require.EqualValues(t, 4, *response.ActionCenter.OutboxFailures.TotalCount)
			require.Contains(t, response.SystemDegradedNodes, "outbox_business_references")
			for index, event := range events {
				item, err := reader.GetDashboardBusinessFollowUpDetail(t.Context(), auth.RoleAdmin, 1, event.ID)
				require.NoError(t, err)
				affected := missingDomain == "products" && index == 1 || missingDomain == "orders" && (index == 0 || index == 3)
				if affected {
					require.Contains(t, item.RelatedRecordNotice, "待办数量仍有效")
					if item.RelatedRecord != nil {
						require.Empty(t, item.RelatedRecord.Path)
					}
				} else {
					require.NotNil(t, item.RelatedRecord)
					require.NotEmpty(t, item.RelatedRecord.Path)
					require.Empty(t, item.RelatedRecordNotice)
				}
			}
		})
	}
}

func TestDashboardOutboxPostgreSQLPartialReferenceFailureDoesNotAbortSharedTransaction(t *testing.T) {
	db := createRollbackOnlyPostgreSQLDashboardBusinessFollowUpDatabase(t)
	for _, event := range []*outbox.Event{
		{EventKey: "postgres-missing-products", EventType: outbox.EventTypeMerchantProductUpsert, AggregateType: outbox.AggregateTypeProduct, AggregateID: "41", Status: "failed"},
		{EventKey: "postgres-valid-order", EventType: outbox.EventTypeOrderPaid, AggregateType: outbox.AggregateTypeOrder, AggregateID: "1", Status: "failed"},
		{EventKey: "postgres-valid-conversation", EventType: outbox.EventTypeCustomerServiceRealtime, AggregateType: outbox.AggregateTypeCustomerServiceConversation, AggregateID: "12", Status: "failed"},
	} {
		require.NoError(t, repository.NewOutboxRepository(db).CreateEvent(event))
	}
	require.NoError(t, db.Exec("DROP TABLE products").Error)
	response := createDashboardBusinessFollowUpReadServiceForTestDatabase(db).GetDashboardBusinessFollowUpList(t.Context(), auth.RoleAdmin, 1)
	require.Equal(t, "ready", response.ActionCenter.OutboxFailures.Status)
	require.EqualValues(t, 3, *response.ActionCenter.OutboxFailures.TotalCount)
	require.Contains(t, response.SystemDegradedNodes, "outbox_business_references")
	var linkedItems int
	for _, item := range response.ActionCenter.OutboxFailures.Items {
		if item.RelatedRecord != nil && item.RelatedRecord.Path != "" {
			linkedItems++
		}
	}
	require.Equal(t, 2, linkedItems, "Order and conversation reads survive a preceding PostgreSQL SQL error")
	var transactionStillUsable int
	require.NoError(t, db.Raw("SELECT 1").Scan(&transactionStillUsable).Error)
	require.Equal(t, 1, transactionStillUsable)
}

func TestDashboardBusinessFollowUpProductReadDoesNotDependOnCustomerServiceTable(t *testing.T) {
	db := newOutboxReconciliationTestDB(t)
	require.NoError(t, db.Exec(`CREATE TABLE products (id INTEGER PRIMARY KEY, name TEXT, deleted_at DATETIME)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO products VALUES (31, 'INDEPENDENT-PRODUCT', NULL)`).Error)
	event := &outbox.Event{EventKey: "independent-product-snapshot", EventType: outbox.EventTypeMerchantProductUpsert, AggregateType: outbox.AggregateTypeProduct, AggregateID: "31", Status: "failed"}
	require.NoError(t, repository.NewOutboxRepository(db).CreateEvent(event))
	reader := createDashboardBusinessFollowUpReadServiceForTestDatabase(db)
	response := reader.GetDashboardBusinessFollowUpList(t.Context(), auth.RoleEditor, 5)
	require.Equal(t, "ready", response.ActionCenter.OutboxFailures.Status, "Product scope must not reference the absent tickets table")
	require.EqualValues(t, 1, *response.ActionCenter.OutboxFailures.TotalCount)
	require.Empty(t, response.SystemDegradedNodes)
	item, err := reader.GetDashboardBusinessFollowUpDetail(t.Context(), auth.RoleEditor, 5, event.ID)
	require.NoError(t, err)
	require.Equal(t, "/google-merchant", item.RelatedRecord.Path)

	// An event with a misleading type still cannot introduce a conversation
	// into a product-only scope, even when the conversation table is unavailable.
	conversationEvent := &outbox.Event{EventKey: "product-event-private-conversation", EventType: outbox.EventTypeMerchantProductUpsert, AggregateType: outbox.AggregateTypeCustomerServiceConversation, AggregateID: "12", Status: "failed"}
	require.NoError(t, repository.NewOutboxRepository(db).CreateEvent(conversationEvent))
	response = reader.GetDashboardBusinessFollowUpList(t.Context(), auth.RoleEditor, 5)
	require.Equal(t, "ready", response.ActionCenter.OutboxFailures.Status)
	require.EqualValues(t, 1, *response.ActionCenter.OutboxFailures.TotalCount)
	_, err = reader.GetDashboardBusinessFollowUpDetail(t.Context(), auth.RoleEditor, 5, conversationEvent.ID)
	require.ErrorIs(t, err, repository.ErrOutboxEventNotFound)
}

func TestDashboardOutboxPostgreSQLReferenceIsolationSurvivesDisabledAutomaticNestedTransactions(t *testing.T) {
	db := createRollbackOnlyPostgreSQLDashboardBusinessFollowUpDatabase(t)
	for _, event := range []*outbox.Event{
		{EventKey: "disabled-nested-missing-product", EventType: outbox.EventTypeMerchantProductUpsert, AggregateType: outbox.AggregateTypeProduct, AggregateID: "41", Status: "failed"},
		{EventKey: "disabled-nested-valid-order", EventType: outbox.EventTypeOrderPaid, AggregateType: outbox.AggregateTypeOrder, AggregateID: "1", Status: "failed"},
	} {
		require.NoError(t, repository.NewOutboxRepository(db).CreateEvent(event))
	}
	require.NoError(t, db.Exec("DROP TABLE products").Error)
	reader := createDashboardBusinessFollowUpReadServiceForTestDatabase(db.Session(&gorm.Session{DisableNestedTransaction: true}))
	response := reader.GetDashboardBusinessFollowUpList(t.Context(), auth.RoleAdmin, 1)
	require.Equal(t, "ready", response.ActionCenter.OutboxFailures.Status)
	var linkedOrder bool
	for _, item := range response.ActionCenter.OutboxFailures.Items {
		if item.RelatedRecord != nil && item.RelatedRecord.Path == "/orders/list?order_id=1" {
			linkedOrder = true
		}
	}
	require.True(t, linkedOrder, "Domain isolation must not depend on the caller's automatic nested transaction setting")
	var transactionStillUsable int
	require.NoError(t, db.Raw("SELECT 1").Scan(&transactionStillUsable).Error)
	require.Equal(t, 1, transactionStillUsable)
}
