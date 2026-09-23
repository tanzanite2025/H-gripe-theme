package service

import (
	"encoding/json"
	"testing"
	"time"

	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/orderevidence"
	productdomain "commerce-platform/internal/domain/product"
	productrequirement "commerce-platform/internal/domain/productrequirement"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOrderEvidenceSnapshotServiceFreezesRequirementHistory(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&productdomain.Product{},
		&productdomain.ProductVariant{},
		&productrequirement.ProductQualityRequirementRule{},
		&order.Order{},
		&order.OrderItem{},
		&orderevidence.OrderEvidenceSnapshot{},
	))

	productRecord := productdomain.Product{
		SKU:        "SKU-EVIDENCE-SERVICE",
		Name:       "Configurable Product",
		Slug:       "evidence-service-product",
		Currency:   "USD",
		PriceMinor: 75000,
		Stock:      5,
	}
	require.NoError(t, db.Create(&productRecord).Error)
	variant := productdomain.ProductVariant{
		ProductID:    productRecord.ID,
		SKU:          "SKU-EVIDENCE-SERVICE-V1",
		Title:        "Configured",
		OptionValues: `{"finish":"black"}`,
		Currency:     "USD",
		PriceMinor:   75000,
		Stock:        5,
		Weight:       9000,
		IsDefault:    true,
		IsActive:     true,
	}
	require.NoError(t, db.Create(&variant).Error)
	rule := productrequirement.ProductQualityRequirementRule{
		ProductID:              productRecord.ID,
		RequirementType:        productrequirement.RequirementTypeSpokeTensionQC,
		SpokeTensionQCRequired: true,
		Status:                 productrequirement.RuleStatusActive,
		RuleVersion:            "product-v1",
		Reason:                 "assembly product",
	}
	require.NoError(t, db.Create(&rule).Error)

	orderRecord := order.Order{
		OrderNumber:      "TZ-2026-SNAPSHOT-SERVICE",
		Status:           "pending",
		PaymentStatus:    "unpaid",
		TotalAmountMinor: 75000,
		Currency:         "USD",
		FXSnapshotData:   serviceTestFXSnapshot(),
		Items: []order.OrderItem{{
			ProductID:                 productRecord.ID,
			VariantID:                 &variant.ID,
			ProductName:               productRecord.Name,
			SKU:                       variant.SKU,
			Quantity:                  1,
			PriceMinor:                75000,
			SubtotalMinor:             75000,
			TotalMinor:                75000,
			ConfigurationSnapshotData: datatypes.JSON([]byte(variant.OptionValues)),
			WeightGrams:               variant.Weight,
		}},
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	require.NotZero(t, orderRecord.Items[0].ID)

	service := NewOrderEvidenceSnapshotService()
	repos := repository.TxRepositories{
		ProductQualityRequirement: repository.NewProductQualityRequirementRepository(db),
		OrderEvidenceSnapshot:     repository.NewOrderEvidenceSnapshotRepository(db),
	}
	created, err := service.CreateForOrder(repos, &orderRecord)
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.True(t, created.IsHighValue)
	assert.True(t, created.HasSpokeTensionQC)

	require.NoError(t, db.Model(&rule).Updates(map[string]interface{}{
		"spoke_tension_qc_required": false,
		"status":                    productrequirement.RuleStatusInactive,
		"rule_version":              "product-v2",
	}).Error)
	require.NoError(t, db.Model(&productdomain.Product{}).
		Where("id = ?", productRecord.ID).
		Updates(map[string]interface{}{"name": "Changed current catalog name"}).Error)

	found, err := repository.NewOrderEvidenceSnapshotRepository(db).FindByOrderID(orderRecord.ID)
	require.NoError(t, err)
	var payload orderevidence.OrderEvidenceSnapshotPayload
	require.NoError(t, json.Unmarshal(found.SnapshotData, &payload))
	require.Len(t, payload.Items, 1)
	assert.True(t, payload.Items[0].ProductRequirementSnapshot.Required)
	assert.Equal(t, rule.ID, *payload.Items[0].ProductRequirementSnapshot.RuleID)
	assert.Equal(t, "product-v1", payload.Items[0].ProductRequirementSnapshot.RuleVersion)
	assert.Equal(t, "Configurable Product", payload.Items[0].ProductName)
}

func TestOrderEvidenceSnapshotServiceRequiresTransactionalStores(t *testing.T) {
	service := NewOrderEvidenceSnapshotService()
	orderRecord := &order.Order{
		ID:               1,
		TotalAmountMinor: 10000,
		Currency:         "USD",
		FXSnapshotData:   serviceTestFXSnapshot(),
		Items: []order.OrderItem{{
			ID:          2,
			ProductID:   3,
			VariantID:   uintPtrForEvidenceTest(4),
			Quantity:    1,
			WeightGrams: 100,
		}},
	}

	_, err := service.CreateForOrder(repository.TxRepositories{}, orderRecord)
	require.ErrorIs(t, err, ErrOrderEvidenceSnapshotStoreUnavailable)

	_, err = service.CreateForOrder(repository.TxRepositories{
		OrderEvidenceSnapshot: repository.NewOrderEvidenceSnapshotRepository(nil),
	}, orderRecord)
	require.ErrorIs(t, err, ErrOrderEvidenceSnapshotRequirementUnavailable)
}

func serviceTestFXSnapshot() datatypes.JSON {
	return currency.OrderFXSnapshotJSON(currency.OrderFXSnapshot{
		Version:       currency.OrderFXSnapshotVersion,
		BaseCurrency:  "USD",
		OrderCurrency: "USD",
		RateDecimal:   "1",
		Source:        "test",
		CapturedAt:    time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC),
	})
}

func uintPtrForEvidenceTest(value uint) *uint {
	return &value
}
