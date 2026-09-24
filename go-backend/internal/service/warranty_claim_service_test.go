package service

import (
	"net/url"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	orderdomain "commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/verification"
	"commerce-platform/internal/domain/warranty"
	"commerce-platform/internal/pkg/emailtoken"
	"commerce-platform/internal/repository"
)

func TestWarrantyServiceRecordValidationAndCreation(t *testing.T) {
	db, warrantyService := newTestWarrantyService(t)
	claim := seedWarrantyClaim(t, db, "TZ-WARRANTY-1", 7)

	_, err := warrantyService.CreateWarrantyServiceRecord(claim.ID, WarrantyServiceRecordInput{
		Summary: " ",
	}, 42)
	require.Error(t, err)
	assert.EqualError(t, err, "service summary is required")

	_, err = warrantyService.CreateWarrantyServiceRecord(claim.ID, WarrantyServiceRecordInput{
		ServiceType: "random",
		Status:      "open",
		Summary:     "checked",
	}, 42)
	require.Error(t, err)
	assert.EqualError(t, err, "invalid service record type")

	_, err = warrantyService.CreateWarrantyServiceRecord(claim.ID, WarrantyServiceRecordInput{
		ServiceType: "inspection",
		Status:      "random",
		Summary:     "checked",
	}, 42)
	require.Error(t, err)
	assert.EqualError(t, err, "invalid service record status")

	_, err = warrantyService.CreateWarrantyServiceRecord(claim.ID, WarrantyServiceRecordInput{
		ServiceType:     "inspection",
		Status:          "open",
		Summary:         "checked",
		CostAmountMinor: -1,
	}, 42)
	require.Error(t, err)
	assert.EqualError(t, err, "service cost amount cannot be negative")

	record, err := warrantyService.CreateWarrantyServiceRecord(claim.ID, WarrantyServiceRecordInput{
		ServiceType:     "Repair",
		Status:          "Processing",
		Summary:         " replaced bearing ",
		CostAmountMinor: 1250,
		Currency:        "usd",
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
	db, warrantyService := newTestWarrantyService(t)
	orderMismatchClaim := seedWarrantyClaim(t, db, "TZ-WARRANTY-2", 7)
	otherOrderItem := seedWarrantyOrderItem(t, db, "TZ-WARRANTY-OTHER", 7, 1002)

	err := warrantyService.BindWarrantyClaimOrderItem(orderMismatchClaim.ID, &otherOrderItem.ID)
	require.ErrorIs(t, err, ErrWarrantyOrderItemMismatch)

	userMismatchClaim := seedWarrantyClaim(t, db, "TZ-WARRANTY-USER", 7)
	otherUserItem := seedWarrantyOrderItem(t, db, "TZ-WARRANTY-USER", 8, 1003)
	err = warrantyService.BindWarrantyClaimOrderItem(userMismatchClaim.ID, &otherUserItem.ID)
	require.ErrorIs(t, err, ErrWarrantyOrderItemMismatch)

	matchingClaim := seedWarrantyClaim(t, db, "TZ-WARRANTY-MATCH", 7)
	matchingItem := seedWarrantyOrderItem(t, db, "TZ-WARRANTY-MATCH", 7, 1001)
	err = warrantyService.BindWarrantyClaimOrderItem(matchingClaim.ID, &matchingItem.ID)
	require.NoError(t, err)

	refreshed, err := repository.NewWarrantyRepository(db).FindWarrantyClaimByID(matchingClaim.ID)
	require.NoError(t, err)
	require.NotNil(t, refreshed.OrderItemID)
	assert.Equal(t, matchingItem.ID, *refreshed.OrderItemID)
}

func TestWarrantyOrderClaimRequiresVerifiedEmailChallenge(t *testing.T) {
	db, warrantyService := newTestWarrantyService(t)
	emailSender := &recordingEmailSender{}
	warrantyService.ConfigureEmailChallenges("test-email-secret")
	warrantyService.ConfigureEmailBaseURL("https://api.example.test")

	order := orderdomain.Order{
		OrderNumber:   "TZ-WARRANTY-VERIFIED",
		UserID:        7,
		Status:        "paid",
		PaymentStatus: "paid",
		TotalAmountMinor: 19900,
		Currency:      "USD",
	}
	order.ShippingAddress.Email = "rider@example.test"
	require.NoError(t, db.Create(&order).Error)

	_, err := warrantyService.CreateWarrantyClaimForOrder(WarrantyClaimByOrderInput{
		OrderNumber: "TZ-WARRANTY-VERIFIED",
		Email:       "rider@example.test",
	})
	require.ErrorIs(t, err, ErrWarrantyVerificationRequired)

	require.NoError(t, warrantyService.RequestWarrantyOrderVerification(order.OrderNumber, order.ShippingAddress.Email))
	processEmailChallengeDelivery(t, db, emailSender)
	require.Len(t, emailSender.bodies, 1)
	verificationURL := strings.TrimSpace(strings.Split(emailSender.bodies[0], "\n\n")[1])
	parsedVerificationURL, err := url.Parse(verificationURL)
	require.NoError(t, err)
	verificationToken := parsedVerificationURL.Query().Get("verification_token")
	require.NotEmpty(t, verificationToken)

	require.NoError(t, warrantyService.ValidateWarrantyOrderToken(verificationToken))
	claim, err := warrantyService.CreateWarrantyClaimForOrder(WarrantyClaimByOrderInput{
		OrderNumber:       order.OrderNumber,
		Email:             order.ShippingAddress.Email,
		VerificationToken: verificationToken,
		Description:       "rim issue",
	})
	require.NoError(t, err)
	assert.Equal(t, order.OrderNumber, claim.OrderNumber)

	_, err = warrantyService.CreateWarrantyClaimForOrder(WarrantyClaimByOrderInput{
		OrderNumber:       order.OrderNumber,
		Email:             order.ShippingAddress.Email,
		VerificationToken: verificationToken,
		Description:       "replay",
	})
	require.ErrorIs(t, err, ErrWarrantyVerificationRequired)
}

func TestWarrantyClaimWriteFailureDoesNotConsumeChallenge(t *testing.T) {
	db, warrantyService := newTestWarrantyService(t)
	emailSender := &recordingEmailSender{}
	warrantyService.ConfigureEmailChallenges("test-email-secret")
	warrantyService.ConfigureEmailBaseURL("https://api.example.test")

	orderRecord := orderdomain.Order{
		OrderNumber:   "TZ-WARRANTY-ROLLBACK",
		UserID:        7,
		Status:        "paid",
		PaymentStatus: "paid",
		TotalAmountMinor: 19900,
		Currency:      "USD",
	}
	orderRecord.ShippingAddress.Email = "rollback@example.test"
	require.NoError(t, db.Create(&orderRecord).Error)
	require.NoError(t, warrantyService.RequestWarrantyOrderVerification(orderRecord.OrderNumber, orderRecord.ShippingAddress.Email))
	processEmailChallengeDelivery(t, db, emailSender)
	verificationURL, err := url.Parse(strings.TrimSpace(strings.Split(emailSender.bodies[0], "\n\n")[1]))
	require.NoError(t, err)
	token := verificationURL.Query().Get("verification_token")
	require.NotEmpty(t, token)

	require.NoError(t, db.Migrator().DropTable(&warranty.WarrantyClaim{}))
	_, err = warrantyService.CreateWarrantyClaimForOrder(WarrantyClaimByOrderInput{
		OrderNumber:       orderRecord.OrderNumber,
		Email:             orderRecord.ShippingAddress.Email,
		VerificationToken: token,
		Description:       "retryable failure",
	})
	require.Error(t, err)

	challenge, err := repository.NewEmailChallengeRepository(db).Find(emailtoken.Hash(token), warrantyOrderChallengePurpose)
	require.NoError(t, err)
	require.Nil(t, challenge.UsedAt)
}

func TestGuestWarrantyClaimCanBeViewedWithAccessTokenOrMatchingAccountEmail(t *testing.T) {
	db, warrantyService := newTestWarrantyService(t)
	warrantyService.ConfigureEmailChallenges("test-email-secret")

	guestClaim := seedWarrantyClaim(t, db, "TZ-WARRANTY-GUEST", 0)
	guestClaim.Email = "guest@example.com"
	guestClaim.Resolution = "Send the wheel to the service address."
	require.NoError(t, db.Save(&guestClaim).Error)

	accessToken, err := warrantyService.IssueWarrantyClaimAccessToken(&guestClaim)
	require.NoError(t, err)
	require.NotEmpty(t, accessToken)

	viewed, err := warrantyService.GetWarrantyClaimForViewer(guestClaim.ID, 0, "", accessToken, false)
	require.NoError(t, err)
	require.Equal(t, guestClaim.Resolution, viewed.Resolution)

	_, err = warrantyService.GetWarrantyClaimForViewer(guestClaim.ID, 0, "", accessToken+"tampered", false)
	require.Error(t, err)

	viewed, err = warrantyService.GetWarrantyClaimForViewer(guestClaim.ID, 5, "GUEST@example.com", "", false)
	require.NoError(t, err)
	require.Equal(t, guestClaim.ID, viewed.ID)

	_, err = warrantyService.GetWarrantyClaimForViewer(guestClaim.ID, 5, "other@example.com", "", false)
	require.Error(t, err)
}

func newTestWarrantyService(t *testing.T) (*gorm.DB, *WarrantyService) {
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
		&warranty.WarrantyClaim{},
		&warranty.WarrantyServiceRecord{},
		&verification.EmailChallenge{},
		&outbox.Event{},
	))

	return db, NewWarrantyService(
		newTestEmailChallengeTxManager(db),
		repository.NewWarrantyRepository(db),
		repository.NewOrderRepository(db),
	)
}

func seedWarrantyClaim(t *testing.T, db *gorm.DB, orderNumber string, userID uint) warranty.WarrantyClaim {
	t.Helper()

	claim := warranty.WarrantyClaim{
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
		TotalAmountMinor: 19900,
		Currency:      "USD",
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
		PriceMinor:    19900,
		SubtotalMinor: 19900,
		TotalMinor:    19900,
	}
	require.NoError(t, db.Create(&item).Error)
	return item
}
