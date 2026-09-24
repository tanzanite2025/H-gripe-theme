package service

import (
	"context"
	"testing"
	"time"

	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/orderevidence"
	productrequirement "commerce-platform/internal/domain/productrequirement"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func TestStartAndCompleteProductionForMadeToOrder(t *testing.T) {
	db, orderService := newTestOrderService(t)
	orderRecord := order.Order{
		OrderNumber:         "ORD-PRODUCTION-1",
		UserID:              42,
		Status:              "processing",
		PaymentStatus:       "paid",
		FulfillmentMode:     order.FulfillmentModeMadeToOrder,
		ProductionStatus:    order.ProductionStatusNotStarted,
		SubtotalAmountMinor: 12000,
		TotalAmountMinor:    12000,
		Currency:            "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)

	started, err := orderService.StartProduction(orderRecord.ID)
	require.NoError(t, err)
	assert.Equal(t, order.ProductionStatusStarted, started.ProductionStatus)
	require.NotNil(t, started.ProductionStartedAt)
	assert.True(t, started.ProductionStartedAt.After(time.Time{}))

	_, err = orderService.StartProduction(orderRecord.ID)
	require.ErrorIs(t, err, ErrOrderProductionAlreadyStarted)

	completed, err := orderService.CompleteProduction(orderRecord.ID)
	require.NoError(t, err)
	assert.Equal(t, order.ProductionStatusCompleted, completed.ProductionStatus)
	require.NotNil(t, completed.ProductionCompletedAt)
	assert.False(t, completed.ProductionCompletedAt.Before(*started.ProductionStartedAt))
}

func TestCompleteProductionDoesNotDependOnSnapshottedSpokeTensionEvidence(t *testing.T) {
	db, orderService := newTestOrderService(t)
	orderRecord, tensionItem := seedProductionEvidenceOrder(t, db)
	evidenceRepo := repository.NewOrderEvidenceRepository(db)

	_, err := NewOrderEvidenceService().UpdateItem(
		repository.TxRepositories{OrderEvidence: evidenceRepo},
		OrderEvidenceItemUpdateInput{
			ItemID:     tensionItem.ID,
			Status:     orderevidence.EvidenceItemStatusDraft,
			DataJSON:   datatypes.JSON([]byte(`{"manual_conclusion":"fail","measured_value":140,"reference_range":{"minimum":105,"maximum":130}}`)),
			CapturedBy: 7,
		},
	)
	require.NoError(t, err)

	completed, err := orderService.CompleteProduction(orderRecord.ID)
	require.NoError(t, err)
	assert.Equal(t, order.ProductionStatusCompleted, completed.ProductionStatus)

	var stored order.Order
	require.NoError(t, db.First(&stored, orderRecord.ID).Error)
	assert.Equal(t, order.ProductionStatusCompleted, stored.ProductionStatus)
	assert.NotNil(t, stored.ProductionCompletedAt)

	pkg, err := evidenceRepo.FindLatestPackageByOrderID(orderRecord.ID)
	require.NoError(t, err)
	assert.Equal(t, orderevidence.PackageStatusIncomplete, pkg.Status)
	var storedTensionItem orderevidence.OrderEvidenceItem
	for _, item := range pkg.Items {
		if item.ID == tensionItem.ID {
			storedTensionItem = item
			break
		}
	}
	assert.Equal(t, orderevidence.EvidenceItemStatusDraft, storedTensionItem.Status)
}

func TestCompleteProductionDoesNotInferSpokeQCFromHighValueOrder(t *testing.T) {
	db, orderService := newTestOrderService(t)
	orderRecord := order.Order{
		OrderNumber:      "ORD-PRODUCTION-HIGH-VALUE-NON-WHEEL",
		UserID:           42,
		Status:           "processing",
		PaymentStatus:    "paid",
		FulfillmentMode:  order.FulfillmentModeMadeToOrder,
		ProductionStatus: order.ProductionStatusStarted,
		TotalAmountMinor: 120000,
		Currency:         "USD",
		FXSnapshotData:   productionEvidenceFXSnapshot(),
	}
	require.NoError(t, db.Create(&orderRecord).Error)

	variantID := uint(23)
	orderItem := order.OrderItem{
		OrderID:       orderRecord.ID,
		ProductID:     11,
		VariantID:     &variantID,
		ProductName:   "High-value component",
		SKU:           "SKU-HIGH-VALUE-COMPONENT",
		Quantity:      1,
		PriceMinor:    120000,
		SubtotalMinor: 120000,
		TotalMinor:    120000,
		WeightGrams:   1000,
	}
	require.NoError(t, db.Create(&orderItem).Error)
	orderRecord.Items = []order.OrderItem{orderItem}
	snapshot, err := orderevidence.BuildOrderEvidenceSnapshot(&orderRecord, []orderevidence.SnapshotItemInput{{
		Item: orderItem,
	}}, time.Now().UTC())
	require.NoError(t, err)
	require.NoError(t, db.Create(snapshot).Error)
	evidenceRepo := repository.NewOrderEvidenceRepository(db)
	_, err = NewOrderEvidenceService().CreateInitialPackage(
		repository.TxRepositories{OrderEvidence: evidenceRepo},
		snapshot,
	)
	require.NoError(t, err)

	completed, err := orderService.CompleteProduction(orderRecord.ID)
	require.NoError(t, err)
	assert.Equal(t, order.ProductionStatusCompleted, completed.ProductionStatus)

	pkg, err := evidenceRepo.FindLatestPackageByOrderID(orderRecord.ID)
	require.NoError(t, err)
	for _, item := range pkg.Items {
		assert.NotEqual(t, orderevidence.EvidenceItemTypeSpokeQCTension, item.ItemType)
	}
}

func TestProductionWorkflowRejectsStockAndUnpaidOrders(t *testing.T) {
	db, orderService := newTestOrderService(t)
	stockOrder := order.Order{
		OrderNumber:      "ORD-PRODUCTION-STOCK",
		UserID:           42,
		Status:           "processing",
		PaymentStatus:    "paid",
		FulfillmentMode:  order.FulfillmentModeStock,
		TotalAmountMinor: 10000,
		Currency:         "USD",
	}
	require.NoError(t, db.Create(&stockOrder).Error)
	_, err := orderService.StartProduction(stockOrder.ID)
	require.ErrorIs(t, err, ErrOrderProductionNotRequired)

	unpaidOrder := order.Order{
		OrderNumber:      "ORD-PRODUCTION-UNPAID",
		UserID:           42,
		Status:           "pending",
		PaymentStatus:    "unpaid",
		FulfillmentMode:  order.FulfillmentModeMadeToOrder,
		ProductionStatus: order.ProductionStatusNotStarted,
		TotalAmountMinor: 10000,
		Currency:         "USD",
	}
	require.NoError(t, db.Create(&unpaidOrder).Error)
	_, err = orderService.StartProduction(unpaidOrder.ID)
	require.ErrorIs(t, err, ErrOrderProductionPaymentRequired)
}

func TestMadeToOrderFulfillmentRequiresCompletedProduction(t *testing.T) {
	db, orderService := newTestOrderService(t)
	orderRecord := order.Order{
		OrderNumber:      "ORD-PRODUCTION-FULFILL-GATE",
		UserID:           42,
		Status:           "processing",
		PaymentStatus:    "paid",
		FulfillmentMode:  order.FulfillmentModeMadeToOrder,
		ProductionStatus: order.ProductionStatusStarted,
		TotalAmountMinor: 10000,
		Currency:         "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)

	_, err := orderService.FulfillOrder(context.Background(), orderRecord.ID, OrderTrackingUpdateInput{})
	require.ErrorIs(t, err, ErrOrderProductionNotCompleted)
}

func TestMadeToOrderCancellationIsBlockedAfterProductionStarts(t *testing.T) {
	db, orderService := newTestOrderService(t)
	orderRecord := order.Order{
		OrderNumber:      "ORD-PRODUCTION-CANCEL-GATE",
		UserID:           42,
		Status:           "processing",
		PaymentStatus:    "paid",
		FulfillmentMode:  order.FulfillmentModeMadeToOrder,
		ProductionStatus: order.ProductionStatusStarted,
		TotalAmountMinor: 10000,
		Currency:         "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)

	require.ErrorIs(t, orderService.CancelOrder(orderRecord.ID, orderRecord.UserID), ErrProductionStartedCancellationNotAllowed)
}

func seedProductionEvidenceOrder(t *testing.T, db *gorm.DB) (order.Order, orderevidence.OrderEvidenceItem) {
	t.Helper()
	orderRecord := order.Order{
		OrderNumber:      "ORD-PRODUCTION-TENSION",
		UserID:           42,
		Status:           "processing",
		PaymentStatus:    "paid",
		FulfillmentMode:  order.FulfillmentModeMadeToOrder,
		ProductionStatus: order.ProductionStatusStarted,
		TotalAmountMinor: 80000,
		Currency:         "USD",
		FXSnapshotData:   productionEvidenceFXSnapshot(),
	}
	require.NoError(t, db.Create(&orderRecord).Error)

	variantID := uint(22)
	orderItem := order.OrderItem{
		OrderID:                   orderRecord.ID,
		ProductID:                 10,
		VariantID:                 &variantID,
		ProductName:               "Assembly Product",
		SKU:                       "SKU-PRODUCTION-TENSION",
		Quantity:                  1,
		PriceMinor:                80000,
		SubtotalMinor:             80000,
		TotalMinor:                80000,
		WeightGrams:               9000,
		ConfigurationSnapshotData: datatypes.JSON([]byte(`{"assembly":"wheelset"}`)),
	}
	require.NoError(t, db.Create(&orderItem).Error)
	orderRecord.Items = []order.OrderItem{orderItem}

	ruleID := uint(900)
	snapshot, err := orderevidence.BuildOrderEvidenceSnapshot(&orderRecord, []orderevidence.SnapshotItemInput{{
		Item: orderItem,
		ProductRequirement: productrequirement.SpokeTensionQCResolution{
			Required:        true,
			Matched:         true,
			RuleID:          &ruleID,
			RuleVersion:     "test-v1",
			Reason:          "explicit assembly rule",
			Source:          productrequirement.ResolutionSourceProduct,
			RequirementType: productrequirement.RequirementTypeSpokeTensionQC,
		},
	}}, time.Now().UTC())
	require.NoError(t, err)
	require.NoError(t, db.Create(snapshot).Error)

	evidenceRepo := repository.NewOrderEvidenceRepository(db)
	_, err = NewOrderEvidenceService().CreateInitialPackage(
		repository.TxRepositories{OrderEvidence: evidenceRepo},
		snapshot,
	)
	require.NoError(t, err)

	pkg, err := evidenceRepo.FindLatestPackageByOrderID(orderRecord.ID)
	require.NoError(t, err)
	for _, item := range pkg.Items {
		if item.ItemType == orderevidence.EvidenceItemTypeSpokeQCTension {
			return orderRecord, item
		}
	}
	t.Fatal("spoke tension evidence item was not created")
	return order.Order{}, orderevidence.OrderEvidenceItem{}
}

func productionEvidenceFXSnapshot() datatypes.JSON {
	return currency.OrderFXSnapshotJSON(currency.OrderFXSnapshot{
		Version:       currency.OrderFXSnapshotVersion,
		BaseCurrency:  "USD",
		OrderCurrency: "USD",
		RateDecimal:   "1",
		Source:        "production-test",
		CapturedAt:    time.Now().UTC(),
	})
}
