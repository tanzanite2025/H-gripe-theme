package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"commerce-platform/internal/domain/aftersales"
	coupondomain "commerce-platform/internal/domain/coupon"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/order"
	outboxdomain "commerce-platform/internal/domain/outbox"
	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
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
	require.NoError(t, err)
}

func TestAfterSalesServicePersistsAndValidatesReturnShipmentLifecycle(t *testing.T) {
	db, service := newAfterSalesService(t)
	orderRecord := seedAfterSalesOrder(t, db, 1)
	created, err := service.CreateCase(CreateAfterSalesCaseInput{
		OrderID: orderRecord.ID,
		Type:    aftersales.TypeReturnRefund,
		Reason:  "Return damaged item",
		Items: []AfterSalesCaseItemInput{{
			OrderItemID: orderRecord.Items[0].ID,
			Quantity:    1,
		}},
	})
	require.NoError(t, err)
	_, err = service.UpdateStatus(created.ID, aftersales.StatusReviewing, "Reviewed", 7)
	require.NoError(t, err)
	_, err = service.UpdateStatus(created.ID, aftersales.StatusApproved, "Return approved", 7)
	require.NoError(t, err)

	_, err = service.UpdateStatusWithReturnShipment(created.ID, UpdateAfterSalesStatusInput{
		Status:           aftersales.StatusAwaitingReturn,
		Resolution:       "Send the item to the EU returns hub",
		UpdatedBy:        7,
		WarehouseName:    "EU Returns Hub",
		WarehouseAddress: "1 Returns Way, Amsterdam",
	})
	require.NoError(t, err)
	_, err = service.UpdateStatusWithReturnShipment(created.ID, UpdateAfterSalesStatusInput{
		Status:         aftersales.StatusReturnInTransit,
		Resolution:     "Customer shipped the package",
		UpdatedBy:      7,
		Carrier:        "DHL",
		TrackingNumber: "DHL-123",
		TrackingURL:    "https://tracking.example/DHL-123",
		LabelURL:       "https://labels.example/DHL-123.pdf",
	})
	require.NoError(t, err)

	withoutReceiver := uint(0)
	_, err = service.UpdateStatusWithReturnShipment(created.ID, UpdateAfterSalesStatusInput{
		Status:     aftersales.StatusReceived,
		Resolution: "Package arrived",
		UpdatedBy:  7,
		ReceivedBy: &withoutReceiver,
	})
	require.ErrorIs(t, err, ErrAfterSalesReturnReceiverRequired)

	receivedBy := uint(42)
	updated, err := service.UpdateStatusWithReturnShipment(created.ID, UpdateAfterSalesStatusInput{
		Status:     aftersales.StatusReceived,
		Resolution: "Package arrived",
		UpdatedBy:  7,
		ReceivedBy: &receivedBy,
	})
	require.NoError(t, err)
	assert.Equal(t, aftersales.StatusReceived, updated.Status)
	require.Len(t, updated.ReturnShipments, 1)
	assert.Equal(t, "DHL", updated.ReturnShipments[0].Carrier)
	assert.Equal(t, "DHL-123", updated.ReturnShipments[0].TrackingNumber)
	assert.Equal(t, "https://labels.example/DHL-123.pdf", updated.ReturnShipments[0].LabelURL)
	assert.NotNil(t, updated.ReturnShipments[0].ReceivedAt)
}

func TestAfterSalesServiceFindsCaseByReturnTrackingNumber(t *testing.T) {
	db, service := newAfterSalesService(t)
	orderRecord := seedAfterSalesOrder(t, db, 1)
	created, err := service.CreateCase(CreateAfterSalesCaseInput{
		OrderID: orderRecord.ID,
		Type:    aftersales.TypeReturnRefund,
		Reason:  "Warehouse scan lookup",
		Items: []AfterSalesCaseItemInput{{
			OrderItemID: orderRecord.Items[0].ID,
			Quantity:    1,
		}},
	})
	require.NoError(t, err)
	_, err = service.UpdateStatus(created.ID, aftersales.StatusReviewing, "Reviewed", 7)
	require.NoError(t, err)
	_, err = service.UpdateStatus(created.ID, aftersales.StatusApproved, "Approved", 7)
	require.NoError(t, err)
	_, err = service.UpdateStatusWithReturnShipment(created.ID, UpdateAfterSalesStatusInput{
		Status:           aftersales.StatusAwaitingReturn,
		UpdatedBy:        7,
		WarehouseName:    "EU Returns Hub",
		WarehouseAddress: "1 Returns Way",
	})
	require.NoError(t, err)
	_, err = service.UpdateStatusWithReturnShipment(created.ID, UpdateAfterSalesStatusInput{
		Status:         aftersales.StatusReturnInTransit,
		UpdatedBy:      7,
		Carrier:        "DHL",
		TrackingNumber: "RET-SCAN-42",
	})
	require.NoError(t, err)

	resolved, err := service.FindCaseByReturnTrackingNumber("  ret-scan-42 ")
	require.NoError(t, err)
	assert.Equal(t, created.ID, resolved.ID)
	assert.Equal(t, aftersales.StatusReturnInTransit, resolved.Status)
	require.Len(t, resolved.ReturnShipments, 1)
	assert.Equal(t, "RET-SCAN-42", resolved.ReturnShipments[0].TrackingNumber)
}

func TestAfterSalesStatusTransitionQueuesCanonicalOutboxFact(t *testing.T) {
	db, service := newAfterSalesService(t)
	require.NoError(t, db.AutoMigrate(&outboxdomain.Event{}))
	service.txManager.ConfigureOutboxRepository(repository.NewOutboxRepository(db))
	orderRecord := seedAfterSalesOrder(t, db, 1)
	created, err := service.CreateCase(CreateAfterSalesCaseInput{
		OrderID: orderRecord.ID,
		Type:    aftersales.TypeReturnRefund,
		Reason:  "Package arrived damaged",
		Items: []AfterSalesCaseItemInput{{
			OrderItemID: orderRecord.Items[0].ID,
			Quantity:    1,
		}},
	})
	require.NoError(t, err)
	_, err = service.UpdateStatus(created.ID, aftersales.StatusReviewing, "Review started", 7)
	require.NoError(t, err)

	var event outboxdomain.Event
	require.NoError(t, db.Where(
		"event_type = ? AND aggregate_id = ?",
		outboxdomain.EventTypeAfterSalesStatusChanged,
		fmt.Sprint(created.ID),
	).Order("id DESC").First(&event).Error)
	var payload outboxdomain.AfterSalesStatusChangedPayload
	require.NoError(t, json.Unmarshal(event.Payload, &payload))
	assert.Equal(t, created.ID, payload.CaseID)
	assert.Equal(t, aftersales.StatusRequested, payload.PreviousStatus)
	assert.Equal(t, aftersales.StatusReviewing, payload.NewStatus)
	assert.NotZero(t, payload.TransitionID)
}

func TestAfterSalesReturnShipmentFieldsReachCanonicalOutboxSnapshot(t *testing.T) {
	db, service := newAfterSalesService(t)
	require.NoError(t, db.AutoMigrate(&outboxdomain.Event{}))
	service.txManager.ConfigureOutboxRepository(repository.NewOutboxRepository(db))
	orderRecord := seedAfterSalesOrder(t, db, 1)
	created, err := service.CreateCase(CreateAfterSalesCaseInput{
		OrderID: orderRecord.ID,
		Type:    aftersales.TypeReturnRefund,
		Reason:  "Return package tracking",
		Items: []AfterSalesCaseItemInput{{
			OrderItemID: orderRecord.Items[0].ID,
			Quantity:    1,
		}},
	})
	require.NoError(t, err)
	_, err = service.UpdateStatus(created.ID, aftersales.StatusReviewing, "Reviewed", 7)
	require.NoError(t, err)
	_, err = service.UpdateStatus(created.ID, aftersales.StatusApproved, "Approved", 7)
	require.NoError(t, err)
	_, err = service.UpdateStatusWithReturnShipment(created.ID, UpdateAfterSalesStatusInput{
		Status:           aftersales.StatusAwaitingReturn,
		Resolution:       "Return instructions issued",
		UpdatedBy:        7,
		WarehouseName:    "EU Returns Hub",
		WarehouseAddress: "1 Returns Way",
		LabelURL:         "https://labels.example/return-1.pdf",
	})
	require.NoError(t, err)

	var event outboxdomain.Event
	require.NoError(t, db.Where(
		"event_type = ? AND aggregate_id = ?",
		outboxdomain.EventTypeAfterSalesStatusChanged,
		fmt.Sprint(created.ID),
	).Order("id DESC").First(&event).Error)
	var payload outboxdomain.AfterSalesStatusChangedPayload
	require.NoError(t, json.Unmarshal(event.Payload, &payload))
	assert.Equal(t, aftersales.StatusAwaitingReturn, payload.NewStatus)
	assert.Equal(t, "EU Returns Hub", payload.WarehouseName)
	assert.Equal(t, "1 Returns Way", payload.WarehouseAddress)
	assert.Equal(t, "https://labels.example/return-1.pdf", payload.LabelURL)
}

func TestAfterSalesCustomerRequestQueuesCreationOutboxFact(t *testing.T) {
	db, service := newAfterSalesService(t)
	require.NoError(t, db.AutoMigrate(&outboxdomain.Event{}))
	service.txManager.ConfigureOutboxRepository(repository.NewOutboxRepository(db))
	orderRecord := seedAfterSalesOrder(t, db, 1)

	created, err := service.CreateCustomerRequest(CreateCustomerAfterSalesRequestInput{
		OrderID:     orderRecord.ID,
		Reason:      "Wrong valve color",
		Description: "The delivered valve color does not match the order.",
		Items:       []AfterSalesCaseItemInput{{OrderItemID: orderRecord.Items[0].ID, Quantity: 1}},
		CreatedBy:   1,
	})
	require.NoError(t, err)

	var event outboxdomain.Event
	require.NoError(t, db.Where(
		"event_type = ? AND aggregate_id = ?",
		outboxdomain.EventTypeAfterSalesStatusChanged,
		fmt.Sprint(created.ID),
	).First(&event).Error)
	var payload outboxdomain.AfterSalesStatusChangedPayload
	require.NoError(t, json.Unmarshal(event.Payload, &payload))
	assert.Equal(t, created.ID, payload.CaseID)
	assert.Equal(t, aftersales.StatusRequested, payload.NewStatus)
	assert.Empty(t, payload.PreviousStatus)
	assert.NotZero(t, payload.TransitionID)
	assert.Equal(t, orderRecord.OrderNumber, payload.OrderNumber)
}

func TestAfterSalesServiceCustomerRequestDoesNotConsumeCompletedCaseQuantity(t *testing.T) {
	db, service := newAfterSalesService(t)
	orderRecord := seedAfterSalesOrder(t, db, 2)

	completedCase, err := service.CreateCase(CreateAfterSalesCaseInput{
		OrderID: orderRecord.ID,
		Type:    aftersales.TypeReturnRefund,
		Reason:  "First wheelset returned",
		Items: []AfterSalesCaseItemInput{{
			OrderItemID: orderRecord.Items[0].ID,
			Quantity:    1,
		}},
		CreatedBy: 7,
	})
	require.NoError(t, err)
	moveAfterSalesCaseToResolving(t, service, completedCase.ID)
	_, err = service.UpdateStatus(completedCase.ID, aftersales.StatusCompleted, "One wheelset returned", 7)
	require.NoError(t, err)

	request, err := service.CreateCustomerRequest(CreateCustomerAfterSalesRequestInput{
		OrderID:     orderRecord.ID,
		Reason:      "Second wheelset support request",
		Description: "The second wheelset now needs after-sales support.",
		Items:       []AfterSalesCaseItemInput{{OrderItemID: orderRecord.Items[0].ID, Quantity: 1}},
		CreatedBy:   1,
	})
	require.NoError(t, err)
	require.Len(t, request.Items, 1)
	assert.Equal(t, orderRecord.Items[0].ID, request.Items[0].OrderItemID)
	assert.Equal(t, 1, request.Items[0].Quantity)
}

func TestAfterSalesServiceCustomerRequestSnapshotDoesNotReserveEligibility(t *testing.T) {
	db, service := newAfterSalesService(t)
	orderRecord := seedAfterSalesOrder(t, db, 2)

	request, err := service.CreateCustomerRequest(CreateCustomerAfterSalesRequestInput{
		OrderID:     orderRecord.ID,
		Reason:      "Valve color is wrong",
		Description: "Please review the affected item.",
		CreatedBy:   1,
		Items:       []AfterSalesCaseItemInput{{OrderItemID: orderRecord.Items[0].ID, Quantity: 1}},
	})
	require.NoError(t, err)
	require.Len(t, request.Items, 1)
	assert.Equal(t, 1, request.Items[0].Quantity)

	// The customer snapshot contains the complete order for staff context, but
	// it must not prevent a later operator-created case from selecting the
	// actual affected line and quantity.
	created, err := service.CreateCase(CreateAfterSalesCaseInput{
		OrderID: orderRecord.ID,
		Type:    aftersales.TypeRefundOnly,
		Reason:  "Refund only the incorrectly colored valve",
		Items: []AfterSalesCaseItemInput{{
			OrderItemID: orderRecord.Items[0].ID,
			Quantity:    2,
		}},
		CreatedBy: 7,
	})
	require.NoError(t, err)
	require.Len(t, created.Items, 1)
	assert.Equal(t, 2, created.Items[0].Quantity)
}

func TestAfterSalesServiceAllowsCustomerRequestAlongsideOperatorCase(t *testing.T) {
	db, service := newAfterSalesService(t)
	orderRecord := seedAfterSalesOrder(t, db, 2)

	operatorCase, err := service.CreateCase(CreateAfterSalesCaseInput{
		OrderID: orderRecord.ID,
		Type:    aftersales.TypeReturnRefund,
		Reason:  "Return one wheelset",
		Items: []AfterSalesCaseItemInput{{
			OrderItemID: orderRecord.Items[0].ID,
			Quantity:    1,
		}},
		CreatedBy: 7,
	})
	require.NoError(t, err)

	customerCase, err := service.CreateCustomerRequest(CreateCustomerAfterSalesRequestInput{
		OrderID:     orderRecord.ID,
		Reason:      "The remaining wheelset needs support",
		Description: "Please review the other item while the first return is in transit.",
		Items:       []AfterSalesCaseItemInput{{OrderItemID: orderRecord.Items[0].ID, Quantity: 1}},
		CreatedBy:   1,
	})
	require.NoError(t, err)
	require.NotNil(t, customerCase)
	assert.Equal(t, aftersales.TypeCustomerRequest, customerCase.Type)
	require.Len(t, customerCase.Items, 1)
	assert.Equal(t, 1, customerCase.Items[0].Quantity)
	assert.NotEqual(t, operatorCase.ID, customerCase.ID)
}

func TestAfterSalesServiceListsCustomerCasesWithStatusHistory(t *testing.T) {
	db, service := newAfterSalesService(t)
	orderRecord := seedAfterSalesOrder(t, db, 1)

	created, err := service.CreateCustomerRequest(CreateCustomerAfterSalesRequestInput{
		OrderID:     orderRecord.ID,
		Reason:      "Need return instructions",
		Description: "Please share the return address and next steps.",
		Items:       []AfterSalesCaseItemInput{{OrderItemID: orderRecord.Items[0].ID, Quantity: 1}},
		CreatedBy:   1,
	})
	require.NoError(t, err)

	cases, err := service.ListCustomerCasesByOrder(orderRecord.ID, "")
	require.NoError(t, err)
	require.Len(t, cases, 1)
	assert.Equal(t, created.ID, cases[0].ID)
	assert.Equal(t, aftersales.StatusRequested, cases[0].Status)
	require.Len(t, cases[0].Items, 1)
	require.Len(t, cases[0].Events, 1)
	assert.Equal(t, aftersales.StatusRequested, cases[0].Events[0].ToStatus)
}

func TestAfterSalesServiceRejectsDuplicateActiveCustomerRequests(t *testing.T) {
	db, service := newAfterSalesService(t)
	orderRecord := seedAfterSalesOrder(t, db, 1)
	input := CreateCustomerAfterSalesRequestInput{
		OrderID:     orderRecord.ID,
		Reason:      "Package arrived damaged",
		Description: "The product needs support review.",
		Items:       []AfterSalesCaseItemInput{{OrderItemID: orderRecord.Items[0].ID, Quantity: 1}},
		CreatedBy:   1,
	}

	first, err := service.CreateCustomerRequest(input)
	require.NoError(t, err)
	require.NotNil(t, first)
	assert.Equal(t, aftersales.StatusRequested, first.Status)

	second, err := service.CreateCustomerRequest(input)
	require.ErrorIs(t, err, ErrAfterSalesRequestAlreadyExists)
	assert.Nil(t, second)

	var activeCount int64
	require.NoError(t, db.Model(&aftersales.AfterSalesCase{}).
		Where("order_id = ?", orderRecord.ID).
		Where("status NOT IN ?", []string{
			aftersales.StatusCompleted,
			aftersales.StatusRejected,
			aftersales.StatusCancelled,
		}).
		Count(&activeCount).Error)
	assert.Equal(t, int64(1), activeCount)
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
		ProposedAmount: mustTestMoney(t, 50, "usd"),
		RequestNotes:   "Selected item total is refundable.",
		UpdatedBy:      7,
	})
	require.NoError(t, err)
	assert.Equal(t, aftersales.RefundReviewStatusPending, draft.Status)
	assert.Equal(t, "USD", draft.Currency)
	assert.Equal(t, int64(5000), draft.ProposedAmountMinor)

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
		ProposedAmount: mustTestMoney(t, 20, "USD"),
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
		ProposedAmount: mustTestMoney(t, 10, "USD"),
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
		ProposedAmount: mustTestMoney(t, 51, "USD"),
		RequestNotes:   "Exceeds selected item total.",
		UpdatedBy:      7,
	})
	require.ErrorIs(t, err, ErrAfterSalesRefundReviewAmountExceeded)
}

func TestRefundReviewLimitAllocatesOrderDiscountAcrossSelectedItems(t *testing.T) {
	orderRecord := &order.Order{
		SubtotalAmountMinor: 20000,
		DiscountAmountMinor: 10000,
		TotalAmountMinor:    10000,
		Currency:            "USD",
		Items: []order.OrderItem{
			{ID: 1, Quantity: 1, Currency: "USD", SubtotalMinor: 10000, TotalMinor: 10000, PricingSnapshotData: datatypes.JSON([]byte(`{"net_amount_minor":10000}`))},
			{ID: 2, Quantity: 1, Currency: "USD", SubtotalMinor: 10000, TotalMinor: 10000, PricingSnapshotData: datatypes.JSON([]byte(`{"net_amount_minor":10000}`))},
		},
	}
	caseRecord := &aftersales.AfterSalesCase{
		Type: aftersales.TypeRefundOnly,
		Items: []aftersales.AfterSalesCaseItem{
			{OrderItemID: 1, Quantity: 1},
		},
	}

	amount := refundReviewLimit(caseRecord, orderRecord)

	assert.Equal(t, "USD", amount.Currency().String())
	assert.Equal(t, int64(5000), amount.AmountMinor())
}

func TestRefundReviewLimitKeepsUndiscountedAmountAndCapsAtOrderTotal(t *testing.T) {
	orderRecord := &order.Order{
		SubtotalAmountMinor: 20000,
		TotalAmountMinor:    3000,
		Currency:            "USD",
		Items: []order.OrderItem{
			{ID: 1, Quantity: 1, Currency: "USD", SubtotalMinor: 10000, TotalMinor: 10000, PricingSnapshotData: datatypes.JSON([]byte(`{"net_amount_minor":10000}`))},
			{ID: 2, Quantity: 1, Currency: "USD", SubtotalMinor: 10000, TotalMinor: 10000, PricingSnapshotData: datatypes.JSON([]byte(`{"net_amount_minor":10000}`))},
		},
	}
	caseRecord := &aftersales.AfterSalesCase{
		Type: aftersales.TypeRefundOnly,
		Items: []aftersales.AfterSalesCaseItem{
			{OrderItemID: 1, Quantity: 1},
		},
	}

	amount := refundReviewLimit(caseRecord, orderRecord)

	assert.Equal(t, int64(3000), amount.AmountMinor())
}

func TestAfterSalesServiceCreatesIdempotentPendingRefundFromApprovedReview(t *testing.T) {
	db, service := newAfterSalesService(t)
	orderRecord := seedAfterSalesOrder(t, db, 2)
	transaction := &paymentdomain.Transaction{
		OrderID:       orderRecord.ID,
		TransactionID: "as-refund-draft-transaction",
		PaymentMethod: "stripe",
		AmountMinor:   10000,
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
		ProposedAmount: mustTestMoney(t, 50, "USD"),
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
	assert.Equal(t, int64(5000), refund.AmountMinor)
	assert.Equal(t, int64(5000), refund.RequestedAmountMinor)
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

func TestAfterSalesServiceCreatesPendingRefundForFullyApprovedOrderCouponRefund(t *testing.T) {
	db, service := newAfterSalesService(t)
	orderRecord := seedAfterSalesOrder(t, db, 1)
	require.NoError(t, db.Model(&order.Order{}).Where("id = ?", orderRecord.ID).Updates(map[string]interface{}{
		"subtotal_amount_minor": 100000,
		"discount_amount_minor": 10000,
		"total_amount_minor":    90000,
		"coupon_code":           "SAVE100",
	}).Error)
	pricingSnapshot := refundPricingLineForTest(
		t,
		orderRecord.Items[0].ProductID,
		orderRecord.Items[0].VariantID,
		1,
		1000,
		100,
		"USD",
	)
	require.NoError(t, db.Model(&order.OrderItem{}).Where("id = ?", orderRecord.Items[0].ID).Updates(map[string]interface{}{
		"price_minor":      100000,
		"subtotal_minor":   100000,
		"total_minor":      100000,
		"pricing_snapshot": pricingSnapshot,
	}).Error)
	require.NoError(t, db.Create(&paymentdomain.Transaction{
		OrderID:       orderRecord.ID,
		TransactionID: "as-coupon-refund-transaction",
		PaymentMethod: "stripe",
		AmountMinor:   90000,
		Currency:      "USD",
		Status:        "completed",
	}).Error)
	promo := seedAfterSalesCoupon(t, db, "SAVE100", "fixed", 100, 1000, 0)
	require.NoError(t, db.Create(&coupondomain.CouponUsage{
		CouponID:      promo.ID,
		UserID:        orderRecord.UserID,
		OrderID:       orderRecord.ID,
		DiscountMinor: 10000,
	}).Error)

	created, err := service.CreateCase(CreateAfterSalesCaseInput{
		OrderID: orderRecord.ID,
		Type:    aftersales.TypeRefundOnly,
		Reason:  "Order-level coupon refund",
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
		ProposedAmount: mustTestMoney(t, 900, "USD"),
		RequestNotes:   "Approve the paid amount after coupon discount.",
		UpdatedBy:      7,
	})
	require.NoError(t, err)
	_, err = service.DecideRefundReview(DecideAfterSalesRefundReviewInput{
		CaseID:        created.ID,
		Status:        aftersales.RefundReviewStatusApproved,
		DecisionNotes: "Approved.",
		ReviewedBy:    8,
	})
	require.NoError(t, err)

	_, refund, err := service.CreatePendingRefundFromApprovedReview(CreateAfterSalesPendingRefundInput{
		CaseID:  created.ID,
		AdminID: 9,
	})
	require.NoError(t, err)
	require.NotNil(t, refund)
	assert.Equal(t, "pending", refund.Status)
	assert.Equal(t, int64(90000), refund.RequestedAmountMinor)
	assert.Equal(t, int64(0), refund.DiscountClawbackAmountMinor)
	assert.Equal(t, int64(90000), refund.AmountMinor)
	require.Len(t, refund.LineItems, 1)
	assert.Equal(t, int64(90000), refund.LineItems[0].LineTotalMinor)
}

func TestAfterSalesServiceDoesNotCreateRefundBeforeApproval(t *testing.T) {
	db, service := newAfterSalesService(t)
	orderRecord := seedAfterSalesOrder(t, db, 1)
	require.NoError(t, db.Create(&paymentdomain.Transaction{
		OrderID:       orderRecord.ID,
		TransactionID: "as-unapproved-refund-transaction",
		PaymentMethod: "stripe",
		AmountMinor:   10000,
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
		ProposedAmount: mustTestMoney(t, 100, "USD"),
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
		&aftersales.AfterSalesReturnShipment{},
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
	txManager.ConfigureAfterSalesCaseRepository(repository.NewAfterSalesCaseRepository(db))

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
		OrderNumber:         "AS-TEST-" + t.Name() + "-" + string(rune(quantity+'0')),
		UserID:              1,
		Status:              "shipped",
		PaymentStatus:       "paid",
		ShippingStatus:      "shipped",
		SubtotalAmountMinor: 10000,
		TotalAmountMinor:    10000,
		Currency:            "USD",
		Items: []order.OrderItem{{
			ProductID:     10,
			VariantID:     uintPtrForAfterSalesTest(11),
			ProductName:   "Test product",
			SKU:           "TEST-SKU",
			Quantity:      quantity,
			PriceMinor:    5000,
			SubtotalMinor: 10000,
			TotalMinor:    10000,
		}},
	}
	record.Items[0].PricingSnapshotData = refundPricingLineForTest(
		t,
		record.Items[0].ProductID,
		record.Items[0].VariantID,
		record.Items[0].Quantity,
		float64(record.Items[0].PriceMinor)/100,
		float64(record.Items[0].DiscountMinor)/100,
		record.Currency,
	)
	require.NoError(t, db.Create(record).Error)
	require.NoError(t, db.Preload("Items").First(record, record.ID).Error)
	return record
}

func seedAfterSalesCoupon(t *testing.T, db *gorm.DB, code string, couponType string, value float64, minAmount float64, maxDiscount float64) coupondomain.Coupon {
	t.Helper()

	record := coupondomain.Coupon{
		Code:      code,
		Type:      couponType,
		Enabled:   true,
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now().Add(24 * time.Hour),
	}
	if couponType == "percentage" {
		record.ValueRateDecimal = strconv.FormatFloat(value, 'f', -1, 64)
	} else {
		record.ValueMinor = majorTestMinor(t, value, "USD")
	}
	record.MinAmountMinor = majorTestMinor(t, minAmount, "USD")
	record.MaxDiscountMinor = majorTestMinor(t, maxDiscount, "USD")
	require.NoError(t, db.Create(&record).Error)
	return record
}

func majorTestMinor(t *testing.T, amount float64, code string) int64 {
	t.Helper()
	money, err := domainmoney.FromMajorFloat(amount, code)
	require.NoError(t, err)
	return money.AmountMinor()
}

func uintPtrForAfterSalesTest(value uint) *uint {
	return &value
}
