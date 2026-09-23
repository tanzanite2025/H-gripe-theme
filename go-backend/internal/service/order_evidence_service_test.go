package service

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/orderevidence"
	productrequirement "commerce-platform/internal/domain/productrequirement"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOrderEvidenceServiceCreatesScopedInitialPlan(t *testing.T) {
	db := newOrderEvidenceServiceTestDB(t)
	snapshot := servicePlanSnapshot(t, db, true)
	service := NewOrderEvidenceService()
	repos := repository.TxRepositories{
		OrderEvidence: repository.NewOrderEvidenceRepository(db),
	}

	pkg, err := service.CreateInitialPackage(repos, snapshot)
	require.NoError(t, err)
	require.NotNil(t, pkg)
	assert.Equal(t, snapshot.OrderID, pkg.OrderID)
	assert.Equal(t, snapshot.ID, pkg.SnapshotID)
	assert.Equal(t, orderevidence.PackageStatusIncomplete, pkg.Status)
	assert.True(t, pkg.IsHighValue)
	assert.True(t, pkg.HasSpokeTensionQC)

	items, err := repos.OrderEvidence.ListItemsByPackageID(pkg.ID)
	require.NoError(t, err)
	require.Len(t, items, 5)

	var configurationCount, identityCount, tensionCount, outboundCount, podCount int
	for _, item := range items {
		switch item.ItemType {
		case orderevidence.EvidenceItemTypeConfigurationConfirmation:
			configurationCount++
			assert.Equal(t, orderevidence.EvidenceItemStatusComplete, item.Status)
			assert.NotNil(t, item.SnapshotID)
			require.NoError(t, item.Validate())
		case orderevidence.EvidenceItemTypeProductIdentity:
			identityCount++
			assert.Equal(t, orderevidence.EvidenceItemStatusMissing, item.Status)
		case orderevidence.EvidenceItemTypeSpokeQCTension:
			tensionCount++
			assert.Equal(t, orderevidence.EvidenceRequiredReasonSpokeTensionQC, item.RequiredReason)
		case orderevidence.EvidenceItemTypeOutboundWeightPackaging:
			outboundCount++
			assert.Nil(t, item.OrderItemID)
		case orderevidence.EvidenceItemTypeSignedPOD:
			podCount++
			assert.Nil(t, item.OrderItemID)
		}
	}
	assert.Equal(t, 1, configurationCount)
	assert.Equal(t, 1, identityCount)
	assert.Equal(t, 1, tensionCount)
	assert.Equal(t, 1, outboundCount)
	assert.Equal(t, 1, podCount)
}

func TestOrderEvidenceServiceDoesNotCreateSpokeItemWithoutExplicitRequirement(t *testing.T) {
	db := newOrderEvidenceServiceTestDB(t)
	snapshot := servicePlanSnapshot(t, db, false)
	service := NewOrderEvidenceService()
	repos := repository.TxRepositories{
		OrderEvidence: repository.NewOrderEvidenceRepository(db),
	}

	pkg, err := service.CreateInitialPackage(repos, snapshot)
	require.NoError(t, err)
	items, err := repos.OrderEvidence.ListItemsByPackageID(pkg.ID)
	require.NoError(t, err)
	require.Len(t, items, 4)
	for _, item := range items {
		assert.NotEqual(t, orderevidence.EvidenceItemTypeSpokeQCTension, item.ItemType)
	}
}

func TestOrderEvidenceServiceRejectsSnapshotPayloadMismatch(t *testing.T) {
	db := newOrderEvidenceServiceTestDB(t)
	snapshot := servicePlanSnapshot(t, db, false)
	var payload orderevidence.OrderEvidenceSnapshotPayload
	require.NoError(t, json.Unmarshal(snapshot.SnapshotData, &payload))
	payload.Items[0].ProductName = "tampered"
	tampered, err := json.Marshal(payload)
	require.NoError(t, err)
	snapshot.SnapshotData = datatypes.JSON(tampered)

	_, err = NewOrderEvidenceService().CreateInitialPackage(repository.TxRepositories{
		OrderEvidence: repository.NewOrderEvidenceRepository(db),
	}, snapshot)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sha256 does not match")
}

func TestOrderEvidenceServiceRejectsSnapshotColumnPayloadDrift(t *testing.T) {
	db := newOrderEvidenceServiceTestDB(t)
	snapshot := servicePlanSnapshot(t, db, false)
	snapshot.IsHighValue = !snapshot.IsHighValue

	_, err := NewOrderEvidenceService().CreateInitialPackage(repository.TxRepositories{
		OrderEvidence: repository.NewOrderEvidenceRepository(db),
	}, snapshot)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "high-value flag does not match payload")
}

func TestOrderEvidenceAttachmentServiceEnforcesOrderScopedStoragePrefix(t *testing.T) {
	db := newOrderEvidenceServiceTestDB(t)
	snapshot := servicePlanSnapshot(t, db, false)
	pkg, err := NewOrderEvidenceService().CreateInitialPackage(repository.TxRepositories{
		OrderEvidence: repository.NewOrderEvidenceRepository(db),
	}, snapshot)
	require.NoError(t, err)
	items, err := repository.NewOrderEvidenceRepository(db).ListItemsByPackageID(pkg.ID)
	require.NoError(t, err)
	require.NotEmpty(t, items)

	attachmentService := NewOrderEvidenceAttachmentService()
	input := OrderEvidenceAttachmentReferenceInput{
		EvidenceItemID:   items[0].ID,
		StorageKey:       "order-evidence/999/photo.jpg",
		OriginalFilename: "photo.jpg",
		MimeType:         "image/jpeg",
		SizeBytes:        12,
		SHA256:           "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		UploadedBy:       7,
	}
	_, err = attachmentService.RegisterReference(repository.TxRepositories{
		OrderEvidence: repository.NewOrderEvidenceRepository(db),
	}, input)
	require.ErrorIs(t, err, ErrOrderEvidenceAttachmentOwnership)

	input.StorageKey = fmt.Sprintf("order-evidence/%d/photo.jpg", snapshot.OrderID)
	attachment, err := attachmentService.RegisterReference(repository.TxRepositories{
		OrderEvidence: repository.NewOrderEvidenceRepository(db),
	}, input)
	require.NoError(t, err)
	require.NotNil(t, attachment)
	assert.Equal(t, input.StorageKey, attachment.StorageKey)
}

func newOrderEvidenceServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&order.Order{},
		&order.OrderItem{},
		&orderevidence.OrderEvidenceSnapshot{},
		&orderevidence.OrderEvidencePackage{},
		&orderevidence.OrderEvidenceItem{},
		&orderevidence.OrderEvidenceAttachment{},
	))
	return db
}

func servicePlanSnapshot(t *testing.T, db *gorm.DB, requiresQC bool) *orderevidence.OrderEvidenceSnapshot {
	t.Helper()
	orderRecord := order.Order{
		OrderNumber:      "TZ-2026-PLAN-SERVICE",
		TotalAmountMinor: 80000,
		Currency:         "USD",
		FXSnapshotData:   servicePlanFXSnapshot(),
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	variantID := uint(22)
	item := order.OrderItem{
		OrderID:                   orderRecord.ID,
		ProductID:                 10,
		VariantID:                 &variantID,
		ProductName:               "Configured Product",
		SKU:                       "SKU-PLAN-10",
		Quantity:                  1,
		PriceMinor:                80000,
		WeightGrams:               9000,
		ConfigurationSnapshotData: datatypes.JSON([]byte(`{"finish":"black"}`)),
	}
	require.NoError(t, db.Create(&item).Error)
	orderRecord.Items = []order.OrderItem{item}

	var ruleID *uint
	if requiresQC {
		id := uint(900)
		ruleID = &id
	}
	resolution := productrequirement.SpokeTensionQCResolution{
		Required:        requiresQC,
		Matched:         requiresQC,
		RuleID:          ruleID,
		RuleVersion:     "plan-v1",
		Reason:          "explicit test rule",
		Source:          productrequirement.ResolutionSourceProduct,
		RequirementType: productrequirement.RequirementTypeSpokeTensionQC,
	}
	snapshot, err := orderevidence.BuildOrderEvidenceSnapshot(&orderRecord, []orderevidence.SnapshotItemInput{{
		Item:               item,
		ProductRequirement: resolution,
	}}, time.Now().UTC())
	require.NoError(t, err)
	require.NoError(t, db.Create(snapshot).Error)
	return snapshot
}

func servicePlanFXSnapshot() datatypes.JSON {
	return currency.OrderFXSnapshotJSON(currency.OrderFXSnapshot{
		Version:       currency.OrderFXSnapshotVersion,
		BaseCurrency:  "USD",
		OrderCurrency: "USD",
		RateDecimal:   "1",
		Source:        "test",
		CapturedAt:    time.Now().UTC(),
	})
}
