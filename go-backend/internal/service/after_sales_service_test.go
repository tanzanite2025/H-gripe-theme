package service

import (
	"testing"

	"commerce-platform/internal/domain/aftersales"
	coupondomain "commerce-platform/internal/domain/coupon"
	"commerce-platform/internal/domain/order"
	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestAfterSalesServiceCreatesIndependentCaseAndRejectsInvalidItems(t *testing.T) {
	db, service := newAfterSalesService(t)
	orderRecord := seedAfterSalesOrder(t, db, 2)
	otherOrder := seedAfterSalesOrder(t, db, 1)

	created, err := service.CreateCase(CreateAfterSalesCaseInput{
		OrderID: orderRecord.ID,
		Type:    aftersales.TypeReturnRefund,
		Reason:  "Product arrived damaged",
		Items: []AfterSalesCaseItemInput{{
			OrderItemID: orderRecord.Items[0].ID,
			Quantity:    1,
		}},
		CreatedBy: 7,
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, aftersales.StatusRequested, created.Status)
	assert.Equal(t, orderRecord.ID, created.OrderID)
	require.Len(t, created.Items, 1)
	assert.Equal(t, orderRecord.Items[0].ProductName, created.Items[0].ProductName)
	require.Len(t, created.Events, 1)
	assert.Equal(t, "", created.Events[0].FromStatus)
	assert.Equal(t, aftersales.StatusRequested, created.Events[0].ToStatus)
	assert.Equal(t, "售后单创建", created.Events[0].Resolution)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	assert.Equal(t, "shipped", savedOrder.Status)

	_, err = service.CreateCase(CreateAfterSalesCaseInput{
		OrderID: orderRecord.ID,
		Type:    aftersales.TypeExchange,
		Reason:  "Wrong item",
		Items: []AfterSalesCaseItemInput{{
			OrderItemID: otherOrder.Items[0].ID,
			Quantity:    1,
		}},
	})
	require.ErrorIs(t, err, ErrAfterSalesItemOrderMismatch)
}

func TestAfterSalesServiceEnforcesTransitionsAndRemainingQuantity(t *testing.T) {
	db, service := newAfterSalesService(t)
	orderRecord := seedAfterSalesOrder(t, db, 2)

	created, err := service.CreateCase(CreateAfterSalesCaseInput{
		OrderID: orderRecord.ID,
		Type:    aftersales.TypeExchange,
		Reason:  "Size change",
		Items: []AfterSalesCaseItemInput{{
			OrderItemID: orderRecord.Items[0].ID,
			Quantity:    2,
		}},
	})
	require.NoError(t, err)

	_, err = service.CreateCase(CreateAfterSalesCaseInput{
		OrderID: orderRecord.ID,
		Type:    aftersales.TypeReturnRefund,
		Reason:  "Second request",
		Items: []AfterSalesCaseItemInput{{
			OrderItemID: orderRecord.Items[0].ID,
			Quantity:    1,
		}},
	})
	require.ErrorIs(t, err, ErrAfterSalesQuantityExceeded)

	_, err = service.UpdateStatus(created.ID, aftersales.StatusApproved, "", 7)
	require.ErrorIs(t, err, ErrAfterSalesTransitionInvalid)

	_, err = service.UpdateStatus(created.ID, aftersales.StatusReviewing, "", 7)
	require.NoError(t, err)
	updated, err := service.UpdateStatus(created.ID, aftersales.StatusApproved, "", 7)
	require.NoError(t, err)
	assert.Equal(t, aftersales.StatusApproved, updated.Status)
	require.Len(t, updated.Events, 3)
	assert.Equal(t, aftersales.StatusReviewing, updated.Events[2].FromStatus)
	assert.Equal(t, aftersales.StatusApproved, updated.Events[2].ToStatus)

	_, err = service.UpdateStatus(created.ID, aftersales.StatusResolving, "Refund approved", 7)
	require.NoError(t, err)
	_, err = service.UpdateStatus(created.ID, aftersales.StatusCompleted, "Refund completed", 7)
	require.NoError(t, err)

	_, err = service.CreateCase(CreateAfterSalesCaseInput{
		OrderID: orderRecord.ID,
		Type:    aftersales.TypeReturnRefund,
		Reason:  "Duplicate after completion",
		Items: []AfterSalesCaseItemInput{{
			OrderItemID: orderRecord.Items[0].ID,
			Quantity:    1,
		}},
	})
	require.ErrorIs(t, err, ErrAfterSalesQuantityExceeded)
}

func TestAfterSalesServiceListsAdminCasesWithOrderNumberAndFilters(t *testing.T) {
	db, service := newAfterSalesService(t)
	orderRecord := seedAfterSalesOrder(t, db, 1)

	created, err := service.CreateCase(CreateAfterSalesCaseInput{
		OrderID:     orderRecord.ID,
		Type:        aftersales.TypeReshipment,
		Reason:      "Package lost in transit",
		Description: "Carrier confirmed the parcel was lost.",
		Items: []AfterSalesCaseItemInput{{
			OrderItemID: orderRecord.Items[0].ID,
			Quantity:    1,
		}},
	})
	require.NoError(t, err)

	records, total, err := service.ListAdminCases(ListAfterSalesCasesInput{
		Page:     1,
		PageSize: 20,
		Status:   aftersales.StatusRequested,
		Type:     aftersales.TypeReshipment,
		Search:   orderRecord.OrderNumber,
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, records, 1)
	assert.Equal(t, created.ID, records[0].ID)
	assert.Equal(t, orderRecord.OrderNumber, records[0].OrderNumber)
	require.Len(t, records[0].Items, 1)

	_, _, err = service.ListAdminCases(ListAfterSalesCasesInput{Status: "not-a-status"})
	require.ErrorIs(t, err, ErrAfterSalesStatusInvalid)
	_, _, err = service.ListAdminCases(ListAfterSalesCasesInput{Type: "not-a-type"})
	require.ErrorIs(t, err, ErrAfterSalesTypeInvalid)
}

func TestAfterSalesServiceResolvesEventOperatorNames(t *testing.T) {
	db, service := newAfterSalesService(t)
	operator := &user.User{
		Email:     "after-sales-operator@example.com",
		Username:  "after-sales-operator",
		Password:  "test-password",
		FirstName: "Mina",
		LastName:  "Chen",
		Role:      "admin",
	}
	require.NoError(t, db.Create(operator).Error)
	orderRecord := seedAfterSalesOrder(t, db, 1)

	created, err := service.CreateCase(CreateAfterSalesCaseInput{
		OrderID: orderRecord.ID,
		Type:    aftersales.TypeRefundOnly,
		Reason:  "Duplicate charge",
		Items: []AfterSalesCaseItemInput{{
			OrderItemID: orderRecord.Items[0].ID,
			Quantity:    1,
		}},
		CreatedBy: operator.ID,
	})
	require.NoError(t, err)
	require.Len(t, created.Events, 1)
	assert.Equal(t, "Mina Chen", created.Events[0].OperatorName)

	updated, err := service.UpdateStatus(created.ID, aftersales.StatusReviewing, "Review started", operator.ID)
	require.NoError(t, err)
	require.Len(t, updated.Events, 2)
	assert.Equal(t, "Mina Chen", updated.Events[1].OperatorName)
}

func TestAfterSalesServiceRefundReviewDraftAndDecision(t *testing.T) {
	db, service := newAfterSalesService(t)
	orderRecord := seedAfterSalesOrder(t, db, 2)

	created, err := service.CreateCase(CreateAfterSalesCaseInput{
		OrderID: orderRecord.ID,
		Type:    aftersales.TypeRefundOnly,
		Reason:  "Refund for damaged item",
		Items: []AfterSalesCaseItemInput{{
			OrderItemID: orderRecord.Items[0].ID,
			Quantity:    1,
		}},
		CreatedBy: 7,
	})
	require.NoError(t, err)
	_, err = service.UpdateStatus(created.ID, aftersales.StatusReviewing, "Review started", 7)
	require.NoError(t, err)
	_, err = service.UpdateStatus(created.ID, aftersales.StatusApproved, "Refund path approved", 7)
	require.NoError(t, err)
	_, err = service.UpdateStatus(created.ID, aftersales.StatusResolving, "Ready for refund approval", 7)
	require.NoError(t, err)

	draft, err := service.SaveRefundReview(SaveAfterSalesRefundReviewInput{
		CaseID:         created.ID,
		ProposedAmount: 50,
		Currency:       "usd",
		RequestNotes:   "Selected item total is refundable.",
		UpdatedBy:      7,
	})
	require.NoError(t, err)
	assert.Equal(t, aftersales.RefundReviewStatusPending, draft.Status)
	assert.Equal(t, "USD", draft.Currency)
	assert.Equal(t, 50.0, draft.ProposedAmount)

	approved, err := service.DecideRefundReview(DecideAfterSalesRefundReviewInput{
		CaseID:        created.ID,
		Status:        aftersales.RefundReviewStatusApproved,
		DecisionNotes: "Approved for manual refund execution in the next workflow.",
		ReviewedBy:    8,
	})
	require.NoError(t, err)
	assert.Equal(t, aftersales.RefundReviewStatusApproved, approved.Status)
	require.NotNil(t, approved.ReviewedByID)
	assert.Equal(t, uint(8), *approved.ReviewedByID)

	_, err = service.SaveRefundReview(SaveAfterSalesRefundReviewInput{
		CaseID:         created.ID,
		ProposedAmount: 20,
		Currency:       "USD",
		RequestNotes:   "Attempt to overwrite approved decision.",
		UpdatedBy:      7,
	})
	require.ErrorIs(t, err, ErrAfterSalesRefundReviewFinalized)
}

func TestAfterSalesServiceRefundReviewValidatesAvailabilityAndAmount(t *testing.T) {
	db, service := newAfterSalesService(t)
	orderRecord := seedAfterSalesOrder(t, db, 2)

	created, err := service.CreateCase(CreateAfterSalesCaseInput{
		OrderID: orderRecord.ID,
		Type:    aftersales.TypeExchange,
		Reason:  "Exchange only",
		Items: []AfterSalesCaseItemInput{{
			OrderItemID: orderRecord.Items[0].ID,
			Quantity:    1,
		}},
	})
	require.NoError(t, err)
	_, err = service.SaveRefundReview(SaveAfterSalesRefundReviewInput{
		CaseID:         created.ID,
		ProposedAmount: 10,
		Currency:       "USD",
		RequestNotes:   "Not a refund case.",
		UpdatedBy:      7,
	})
	require.ErrorIs(t, err, ErrAfterSalesRefundReviewUnavailable)

	refundCase, err := service.CreateCase(CreateAfterSalesCaseInput{
		OrderID: orderRecord.ID,
		Type:    aftersales.TypeRefundOnly,
		Reason:  "Refund only",
		Items: []AfterSalesCaseItemInput{{
			OrderItemID: orderRecord.Items[0].ID,
			Quantity:    1,
		}},
	})
	require.NoError(t, err)
	_, err = service.UpdateStatus(refundCase.ID, aftersales.StatusReviewing, "Review started", 7)
	require.NoError(t, err)
	_, err = service.UpdateStatus(refundCase.ID, aftersales.StatusApproved, "Refund path approved", 7)
	require.NoError(t, err)
	_, err = service.UpdateStatus(refundCase.ID, aftersales.StatusResolving, "Ready for refund approval", 7)
	require.NoError(t, err)

	_, err = service.SaveRefundReview(SaveAfterSalesRefundReviewInput{
		CaseID:         refundCase.ID,
		ProposedAmount: 51,
		Currency:       "USD",
		RequestNotes:   "Exceeds selected item total.",
		UpdatedBy:      7,
	})
	require.ErrorIs(t, err, ErrAfterSalesRefundReviewAmountExceeded)
}

func TestAfterSalesServiceCreatesIdempotentPendingRefundFromApprovedReview(t *testing.T) {
	db, service := newAfterSalesService(t)
	orderRecord := seedAfterSalesOrder(t, db, 2)
	transaction := &paymentdomain.Transaction{
		OrderID:       orderRecord.ID,
		TransactionID: "as-refund-draft-transaction",
		PaymentMethod: "stripe",
		Amount:        100,
		Currency:      "USD",
		Status:        "completed",
	}
	require.NoError(t, db.Create(transaction).Error)

	created, err := service.CreateCase(CreateAfterSalesCaseInput{
		OrderID: orderRecord.ID,
		Type:    aftersales.TypeRefundOnly,
		Reason:  "Refund approved after inspection",
		Items: []AfterSalesCaseItemInput{{
			OrderItemID: orderRecord.Items[0].ID,
			Quantity:    1,
		}},
		CreatedBy: 7,
	})
	require.NoError(t, err)
	moveAfterSalesCaseToResolving(t, service, created.ID)

	_, err = service.SaveRefundReview(SaveAfterSalesRefundReviewInput{
		CaseID:         created.ID,
		ProposedAmount: 50,
		Currency:       "USD",
		RequestNotes:   "One selected item is refundable.",
		UpdatedBy:      7,
	})
	require.NoError(t, err)
	_, err = service.DecideRefundReview(DecideAfterSalesRefundReviewInput{
		CaseID:        created.ID,
		Status:        aftersales.RefundReviewStatusApproved,
		DecisionNotes: "Approved after inspection.",
		ReviewedBy:    8,
	})
	require.NoError(t, err)

	review, refund, err := service.CreatePendingRefundFromApprovedReview(CreateAfterSalesPendingRefundInput{
		CaseID:  created.ID,
		AdminID: 9,
	})
	require.NoError(t, err)
	require.NotNil(t, review)
	require.NotNil(t, refund)
	require.NotNil(t, review.LinkedRefundID)
	assert.Equal(t, refund.ID, *review.LinkedRefundID)
	assert.Equal(t, transaction.ID, refund.TransactionID)
	assert.Equal(t, "pending", refund.Status)
	assert.Equal(t, 50.0, refund.Amount)
	assert.Equal(t, 50.0, refund.RequestedAmount)
	require.Len(t, refund.LineItems, 1)
	assert.Equal(t, orderRecord.Items[0].ID, refund.LineItems[0].OrderItemID)
	assert.Equal(t, 1, refund.LineItems[0].Quantity)
	assert.Contains(t, refund.Reason, "After-sales case")

	repeatReview, repeatRefund, err := service.CreatePendingRefundFromApprovedReview(CreateAfterSalesPendingRefundInput{
		CaseID:  created.ID,
		AdminID: 9,
	})
	require.NoError(t, err)
	assert.Equal(t, refund.ID, repeatRefund.ID)
	require.NotNil(t, repeatReview.LinkedRefundID)
	assert.Equal(t, refund.ID, *repeatReview.LinkedRefundID)

	var refundCount int64
	require.NoError(t, db.Model(&paymentdomain.Refund{}).Count(&refundCount).Error)
	assert.Equal(t, int64(1), refundCount)
}

func TestAfterSalesServiceDoesNotCreateRefundBeforeApproval(t *testing.T) {
	db, service := newAfterSalesService(t)
	orderRecord := seedAfterSalesOrder(t, db, 1)
	require.NoError(t, db.Create(&paymentdomain.Transaction{
		OrderID:       orderRecord.ID,
		TransactionID: "as-unapproved-refund-transaction",
		PaymentMethod: "stripe",
		Amount:        100,
		Currency:      "USD",
		Status:        "completed",
	}).Error)

	created, err := service.CreateCase(CreateAfterSalesCaseInput{
		OrderID: orderRecord.ID,
		Type:    aftersales.TypeRefundOnly,
		Reason:  "Pending review",
		Items: []AfterSalesCaseItemInput{{
			OrderItemID: orderRecord.Items[0].ID,
			Quantity:    1,
		}},
	})
	require.NoError(t, err)
	moveAfterSalesCaseToResolving(t, service, created.ID)
	_, err = service.SaveRefundReview(SaveAfterSalesRefundReviewInput{
		CaseID:         created.ID,
		ProposedAmount: 100,
		Currency:       "USD",
		RequestNotes:   "Awaiting refund approval.",
		UpdatedBy:      7,
	})
	require.NoError(t, err)

	_, _, err = service.CreatePendingRefundFromApprovedReview(CreateAfterSalesPendingRefundInput{
		CaseID:  created.ID,
		AdminID: 9,
	})
	require.ErrorIs(t, err, ErrAfterSalesRefundReviewNotApproved)
}

func newAfterSalesService(t *testing.T) (*gorm.DB, *AfterSalesService) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(
		&order.Order{},
		&order.OrderItem{},
		&user.User{},
		&coupondomain.Coupon{},
		&coupondomain.CouponUsage{},
		&paymentdomain.Transaction{},
		&paymentdomain.Refund{},
		&paymentdomain.RefundLineItem{},
		&aftersales.AfterSalesCase{},
		&aftersales.AfterSalesCaseItem{},
		&aftersales.AfterSalesCaseEvent{},
		&aftersales.AfterSalesCaseAttachment{},
		&aftersales.AfterSalesRefundReview{},
	))

	orderRepo := repository.NewOrderRepository(db)
	refundReviewRepo := repository.NewAfterSalesRefundReviewRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	txManager := repository.NewTxManager(
		db,
		orderRepo,
		repository.NewProductRepository(db),
		repository.NewCouponRepository(db),
		repository.NewLoyaltyRepository(db),
		paymentRepo,
	)
	txManager.ConfigureAfterSalesRefundReviewRepository(refundReviewRepo)

	service := NewAfterSalesService(
		repository.NewAfterSalesCaseRepository(db),
		orderRepo,
		refundReviewRepo,
	)
	service.ConfigureUserRepository(repository.NewUserRepository(db))
	service.ConfigureTxManager(txManager)
	return db, service
}

func moveAfterSalesCaseToResolving(t *testing.T, service *AfterSalesService, caseID uint) {
	t.Helper()

	_, err := service.UpdateStatus(caseID, aftersales.StatusReviewing, "Review started", 7)
	require.NoError(t, err)
	_, err = service.UpdateStatus(caseID, aftersales.StatusApproved, "Refund path approved", 7)
	require.NoError(t, err)
	_, err = service.UpdateStatus(caseID, aftersales.StatusResolving, "Ready for refund approval", 7)
	require.NoError(t, err)
}

func seedAfterSalesOrder(t *testing.T, db *gorm.DB, quantity int) *order.Order {
	t.Helper()

	record := &order.Order{
		OrderNumber:    "AS-TEST-" + t.Name() + "-" + string(rune(quantity+'0')),
		UserID:         1,
		Status:         "shipped",
		PaymentStatus:  "paid",
		ShippingStatus: "shipped",
		SubtotalAmount: 100,
		TotalAmount:    100,
		Currency:       "USD",
		Items: []order.OrderItem{{
			ProductID:   10,
			VariantID:   uintPtrForAfterSalesTest(11),
			ProductName: "Test product",
			SKU:         "TEST-SKU",
			Quantity:    quantity,
			Price:       50,
			Subtotal:    100,
			Total:       100,
		}},
	}
	require.NoError(t, db.Create(record).Error)
	require.NoError(t, db.Preload("Items").First(record, record.ID).Error)
	return record
}

func uintPtrForAfterSalesTest(value uint) *uint {
	return &value
}
