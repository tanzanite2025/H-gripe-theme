package repository

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"sync/atomic"
	"time"

	"commerce-platform/internal/domain/aftersales"
	"commerce-platform/internal/domain/merchant"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/payment"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/ticket"

	"gorm.io/gorm"
)

// DashboardBusinessFollowUpBusinessIdentity contains authoritative identity facts,
// never presentation labels, customer contacts or fields decoded from an event payload.
type DashboardBusinessFollowUpBusinessIdentity struct {
	Kind         string
	ID           uint
	OrderID      uint
	OrderNumber  string
	Name         string
	TicketNumber string
}

type DashboardBusinessFollowUpBusinessIdentityReadRepository struct {
	db *gorm.DB
}

var dashboardBusinessIdentitySavepointSequence atomic.Uint64

func NewDashboardBusinessFollowUpBusinessIdentityReadRepository(db *gorm.DB) *DashboardBusinessFollowUpBusinessIdentityReadRepository {
	return &DashboardBusinessFollowUpBusinessIdentityReadRepository{db: db}
}

// ReadDashboardBusinessIdentitiesByAggregateType performs one batch query for one
// authorized domain. A zero scope denies reads, including conversation identities.
func (r *DashboardBusinessFollowUpBusinessIdentityReadRepository) ReadDashboardBusinessIdentitiesByAggregateType(ctx context.Context, kind string, ids []uint, scope DashboardOutboxBusinessReadScope) (map[uint]DashboardBusinessFollowUpBusinessIdentity, error) {
	identities := map[uint]DashboardBusinessFollowUpBusinessIdentity{}
	if len(ids) == 0 || !slices.Contains(scope.AllowedBusinessAggregateTypes, kind) {
		return identities, nil
	}
	if r == nil || r.db == nil {
		return nil, errors.New("dashboard business identity repository is unavailable")
	}
	var rows []DashboardBusinessFollowUpBusinessIdentity
	readIdentities := func(db *gorm.DB) error {
		query := db
		switch kind {
		case outbox.AggregateTypeOrder:
			query = query.Model(&order.Order{}).Select("id, order_number")
		case outbox.AggregateTypeAfterSalesCase:
			query = query.Model(&aftersales.AfterSalesCase{}).Select("id, order_id")
		case outbox.AggregateTypePayment:
			query = query.Model(&payment.Refund{}).Select("id, order_id")
		case outbox.AggregateTypeProduct:
			query = query.Model(&product.Product{}).Select("id, name")
		case outbox.AggregateTypeMerchantOffer:
			query = query.Model(&merchant.GoogleMerchantOffer{}).Select("id, title AS name")
		case outbox.AggregateTypeCustomerServiceConversation:
			query = query.Model(&ticket.Ticket{}).Select("id, ticket_number")
			// Assignment may have changed since the failure snapshot was read.
			if !scope.CanViewAllEventTypes || scope.AssignedAgentUserID != nil {
				if scope.AssignedAgentUserID == nil || *scope.AssignedAgentUserID == 0 {
					return nil
				}
				query = query.Where("assigned_to = ?", *scope.AssignedAgentUserID)
			}
		default:
			return nil
		}
		return query.Where("id IN ?", ids).Scan(&rows).Error
	}
	queryDB := r.db.WithContext(ctx)
	var err error
	if _, sharedTransaction := queryDB.Statement.ConnPool.(gorm.TxCommitter); sharedTransaction && queryDB.Name() == "postgres" {
		err = readDashboardBusinessIdentitiesWithPostgreSQLSavepointIsolation(queryDB, readIdentities)
	} else {
		err = readIdentities(queryDB)
	}
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		row.Kind = kind
		if kind == outbox.AggregateTypeOrder {
			row.OrderID = row.ID
		}
		identities[row.ID] = row
	}
	return identities, nil
}

// Explicit savepoints isolate a domain SQL failure even when the caller disables
// GORM's automatic nested transactions. Names contain only an internal prefix
// and decimal digits; neither domain names nor user input become SQL identifiers.
func readDashboardBusinessIdentitiesWithPostgreSQLSavepointIsolation(db *gorm.DB, readIdentities func(*gorm.DB) error) error {
	savepointName := "dashboard_business_identity_" + strconv.FormatUint(dashboardBusinessIdentitySavepointSequence.Add(1), 10)
	// Exec checks the actual SQL error; the PostgreSQL driver's SavePoint and
	// RollbackTo helpers in the current dependency discard execution errors.
	if err := db.Exec("SAVEPOINT " + savepointName).Error; err != nil {
		return err
	}
	readError := readIdentities(db)
	// A canceled read still needs to restore the caller's transaction. Bound
	// cleanup independently so cancellation cannot skip rollback or release.
	cleanupContext, cancelCleanup := context.WithTimeout(context.WithoutCancel(db.Statement.Context), 5*time.Second)
	defer cancelCleanup()
	cleanupDB := db.WithContext(cleanupContext)
	if readError != nil {
		if rollbackError := cleanupDB.Exec("ROLLBACK TO SAVEPOINT " + savepointName).Error; rollbackError != nil {
			return errors.Join(readError, rollbackError)
		}
	}
	return errors.Join(readError, cleanupDB.Exec("RELEASE SAVEPOINT "+savepointName).Error)
}
