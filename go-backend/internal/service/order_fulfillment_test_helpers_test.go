package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// seedReadyFulfillmentEvidenceForTest creates the same pre-dispatch state that
// the admin dialog is expected to produce. POD intentionally remains missing.
func seedReadyFulfillmentEvidenceForTest(t *testing.T, db *gorm.DB, orderRecord *order.Order) {
	t.Helper()
	require.NotNil(t, orderRecord)
	require.NotZero(t, orderRecord.ID)

	var storedOrder order.Order
	require.NoError(t, db.Preload("Items").First(&storedOrder, orderRecord.ID).Error)
	if storedOrder.Currency == "" {
		storedOrder.Currency = "USD"
	}
	if len(storedOrder.FXSnapshotData) == 0 || string(storedOrder.FXSnapshotData) == "{}" {
		storedOrder.FXSnapshotData = currency.OrderFXSnapshotJSON(currency.OrderFXSnapshot{
			Version:       currency.OrderFXSnapshotVersion,
			BaseCurrency:  "USD",
			OrderCurrency: "USD",
			RateDecimal:   "1",
			Source:        "fulfillment-test",
			CapturedAt:    time.Now().UTC(),
		})
	}
	require.NoError(t, db.Model(&order.Order{}).
		Where("id = ?", storedOrder.ID).
		Updates(map[string]interface{}{
			"currency":    storedOrder.Currency,
			"fx_snapshot": storedOrder.FXSnapshotData,
		}).Error)

	if len(storedOrder.Items) == 0 {
		itemPrice := storedOrder.TotalAmountMinor
		if itemPrice <= 0 {
			itemPrice = 100
		}
		variantID := uint(1)
		item := order.OrderItem{
			OrderID:                   storedOrder.ID,
			ProductID:                 1,
			VariantID:                 &variantID,
			ProductName:               "Fulfillment test product",
			SKU:                       fmt.Sprintf("FULFILL-TEST-%d", storedOrder.ID),
			Quantity:                  1,
			PriceMinor:                itemPrice,
			SubtotalMinor:             itemPrice,
			TotalMinor:                itemPrice,
			ConfigurationSnapshotData: datatypes.JSON([]byte("{}")),
			WeightGrams:               1000,
			FulfillmentMode:           order.FulfillmentModeStock,
		}
		require.NoError(t, db.Create(&item).Error)
		storedOrder.Items = []order.OrderItem{item}
	}

	for i := range storedOrder.Items {
		declaredValue := storedOrder.Items[i].TotalMinor
		if declaredValue <= 0 {
			declaredValue = storedOrder.Items[i].PriceMinor
		}
		if declaredValue <= 0 {
			declaredValue = 100
		}
		storedOrder.Items[i].DeclaredValueMinor = &declaredValue
		storedOrder.Items[i].DeclaredValueConfirmed = true
		require.NoError(t, db.Model(&order.OrderItem{}).
			Where("id = ?", storedOrder.Items[i].ID).
			Updates(map[string]interface{}{
				"declared_value_minor":     declaredValue,
				"declared_value_confirmed": true,
			}).Error)
	}

	snapshot, err := orderevidence.BuildOrderEvidenceSnapshot(
		&storedOrder,
		snapshotInputsForFulfillmentTest(storedOrder.Items),
		time.Now().UTC(),
	)
	require.NoError(t, err)
	require.NoError(t, db.Create(snapshot).Error)

	evidenceRepo := repository.NewOrderEvidenceRepository(db)
	_, err = NewOrderEvidenceService().CreateInitialPackage(
		repository.TxRepositories{OrderEvidence: evidenceRepo},
		snapshot,
	)
	require.NoError(t, err)

	pkg, err := evidenceRepo.FindLatestPackageByOrderID(storedOrder.ID)
	require.NoError(t, err)
	items, err := evidenceRepo.ListItemsByPackageID(pkg.ID)
	require.NoError(t, err)

	evidenceService := NewOrderEvidenceService()
	for _, item := range items {
		switch item.ItemType {
		case orderevidence.EvidenceItemTypeProductIdentity:
			registerFulfillmentTestAttachment(t, db, storedOrder.ID, item.ID, "product-identity.jpg", "identity")
			_, err = evidenceService.UpdateItem(
				repository.TxRepositories{OrderEvidence: evidenceRepo},
				OrderEvidenceItemUpdateInput{
					OrderID:    storedOrder.ID,
					ItemID:     item.ID,
					Status:     orderevidence.EvidenceItemStatusComplete,
					DataJSON:   datatypes.JSON([]byte(`{"capture_note":"dispatch product photo"}`)),
					CapturedBy: 7,
				},
			)
			require.NoError(t, err)
		case orderevidence.EvidenceItemTypeOutboundWeightPackaging:
			registerFulfillmentTestAttachment(t, db, storedOrder.ID, item.ID, "outbound-packaging.jpg", "outbound")
			_, err = evidenceService.UpdateItem(
				repository.TxRepositories{OrderEvidence: evidenceRepo},
				OrderEvidenceItemUpdateInput{
					OrderID: storedOrder.ID,
					ItemID:  item.ID,
					Status:  orderevidence.EvidenceItemStatusComplete,
					DataJSON: datatypes.JSON([]byte(`{
						"schema_version": 1,
						"gross_weight_g": 1200,
						"package_count": 1,
						"packaging_method": "carton"
					}`)),
					CapturedBy: 7,
				},
			)
			require.NoError(t, err)
		}
	}
}

func snapshotInputsForFulfillmentTest(items []order.OrderItem) []orderevidence.SnapshotItemInput {
	inputs := make([]orderevidence.SnapshotItemInput, 0, len(items))
	for _, item := range items {
		inputs = append(inputs, orderevidence.SnapshotItemInput{Item: item})
	}
	return inputs
}

func registerFulfillmentTestAttachment(
	t *testing.T,
	db *gorm.DB,
	orderID uint,
	itemID uint,
	filename string,
	content string,
) {
	t.Helper()
	hash := sha256.Sum256([]byte(content))
	attachment := orderevidence.OrderEvidenceAttachment{
		EvidenceItemID:   itemID,
		StorageKey:       fmt.Sprintf("order-evidence/%d/%d/%s", orderID, itemID, filename),
		OriginalFilename: filename,
		MimeType:         "image/jpeg",
		SizeBytes:        int64(len(content)),
		SHA256:           hex.EncodeToString(hash[:]),
		UploadedBy:       7,
	}
	require.NoError(t, db.Create(&attachment).Error)
}
