package service

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	orderdomain "tanzanite/internal/domain/order"
	"tanzanite/internal/domain/registration"
	"tanzanite/internal/repository"
)

func TestWarrantyServiceRecordValidationAndCreation(t *testing.T) {
	db, registrationService := newTestWarrantyRegistrationService(t)
	claim := seedWarrantyClaim(t, db, "TZ-WARRANTY-1", 7)

	_, err := registrationService.CreateWarrantyServiceRecord(claim.ID, WarrantyServiceRecordInput{
		Summary: " ",
	}, 42)
	require.Error(t, err)
	assert.EqualError(t, err, "service summary is required")

	_, err = registrationService.CreateWarrantyServiceRecord(claim.ID, WarrantyServiceRecordInput{
		ServiceType: "random",
		Status:      "open",
		Summary:     "checked",
	}, 42)
	require.Error(t, err)
	assert.EqualError(t, err, "invalid service record type")

	_, err = registrationService.CreateWarrantyServiceRecord(claim.ID, WarrantyServiceRecordInput{
		ServiceType: "inspection",
		Status:      "random",
		Summary:     "checked",
	}, 42)
	require.Error(t, err)
	assert.EqualError(t, err, "invalid service record status")

	_, err = registrationService.CreateWarrantyServiceRecord(claim.ID, WarrantyServiceRecordInput{
		ServiceType: "inspection",
		Status:      "open",
		Summary:     "checked",
		CostAmount:  -1,
	}, 42)
	require.Error(t, err)
	assert.EqualError(t, err, "service cost amount cannot be negative")

	record, err := registrationService.CreateWarrantyServiceRecord(claim.ID, WarrantyServiceRecordInput{
		ServiceType: "Repair",
		Status:      "Processing",
		Summary:     " replaced bearing ",
		CostAmount:  12.5,
		Currency:    "usd",
	}, 42)
	require.NoError(t, err)
	assert.Equal(t, claim.ID, record.ClaimID)
	assert.Equal(t, "repair", record.ServiceType)
	assert.Equal(t, "processing", record.Status)
	assert.Equal(t, "replaced bearing", record.Summary)
	assert.Equal(t, "USD", record.Currency)
	assert.Equal(t, uint(42), record.CreatedBy)
}

func TestBindWarrantyClaimOrderItemRequiresMatchingOrderAndUser(t *testing.T) {
	db, registrationService := newTestWarrantyRegistrationService(t)
	orderMismatchClaim := seedWarrantyClaim(t, db, "TZ-WARRANTY-2", 7)
	otherOrderItem := seedWarrantyOrderItem(t, db, "TZ-WARRANTY-OTHER", 7, 1002)

	err := registrationService.BindWarrantyClaimOrderItem(orderMismatchClaim.ID, &otherOrderItem.ID)
	require.ErrorIs(t, err, ErrWarrantyOrderItemMismatch)

	userMismatchClaim := seedWarrantyClaim(t, db, "TZ-WARRANTY-USER", 7)
	otherUserItem := seedWarrantyOrderItem(t, db, "TZ-WARRANTY-USER", 8, 1003)
	err = registrationService.BindWarrantyClaimOrderItem(userMismatchClaim.ID, &otherUserItem.ID)
	require.ErrorIs(t, err, ErrWarrantyOrderItemMismatch)

	matchingClaim := seedWarrantyClaim(t, db, "TZ-WARRANTY-MATCH", 7)
	matchingItem := seedWarrantyOrderItem(t, db, "TZ-WARRANTY-MATCH", 7, 1001)
	err = registrationService.BindWarrantyClaimOrderItem(matchingClaim.ID, &matchingItem.ID)
	require.NoError(t, err)

	refreshed, err := repository.NewRegistrationRepository(db).FindWarrantyClaimByID(matchingClaim.ID)
	require.NoError(t, err)
	require.NotNil(t, refreshed.OrderItemID)
	assert.Equal(t, matchingItem.ID, *refreshed.OrderItemID)
}

func newTestWarrantyRegistrationService(t *testing.T) (*gorm.DB, *RegistrationService) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(t, db.AutoMigrate(
		&orderdomain.Order{},
		&orderdomain.OrderItem{},
		&registration.ProductRegistration{},
		&registration.WarrantyClaim{},
		&registration.WarrantyServiceRecord{},
	))

	return db, NewRegistrationService(
		repository.NewRegistrationRepository(db),
		nil,
		repository.NewOrderRepository(db),
	)
}

func seedWarrantyClaim(t *testing.T, db *gorm.DB, orderNumber string, userID uint) registration.WarrantyClaim {
	t.Helper()

	claim := registration.WarrantyClaim{
		UserID:      userID,
		IssueType:   "warranty",
		Description: "wheel issue",
		OrderNumber: orderNumber,
		Email:       "rider@example.com",
		Status:      "submitted",
	}
	require.NoError(t, db.Create(&claim).Error)
	return claim
}

func seedWarrantyOrderItem(t *testing.T, db *gorm.DB, orderNumber string, userID uint, productID uint) orderdomain.OrderItem {
	t.Helper()

	order := orderdomain.Order{
		OrderNumber:   orderNumber,
		UserID:        userID,
		Status:        "paid",
		PaymentStatus: "paid",
		TotalAmount:   199,
	}
	require.NoError(t, db.Create(&order).Error)

	variantID := uint(1)
	item := orderdomain.OrderItem{
		OrderID:     order.ID,
		ProductID:   productID,
		VariantID:   &variantID,
		ProductName: "Carbon Wheel",
		SKU:         "CW-001",
		Quantity:    1,
		Price:       199,
		Subtotal:    199,
		Total:       199,
	}
	require.NoError(t, db.Create(&item).Error)
	return item
}
