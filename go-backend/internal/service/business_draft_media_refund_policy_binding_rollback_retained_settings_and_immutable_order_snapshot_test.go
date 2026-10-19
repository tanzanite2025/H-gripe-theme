package service

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/refundcancellation"
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/repository"
)

func configureBusinessDraftMediaRefundPolicyTestService(t *testing.T, env businessDraftMediaReceiptTestEnvironment) (*RefundCancellationPolicyService, *repository.SettingRepository) {
	t.Helper()
	for _, model := range []any{&setting.Setting{}, &order.PolicyDisclosure{}} {
		if !env.db.Migrator().HasTable(model) {
			require.NoError(t, env.db.AutoMigrate(model))
		}
	}
	settings := repository.NewSettingRepository(env.db)
	settings.ConfigurePublicBusinessMediaReferenceIdentityResolver(env.service)
	return NewRefundCancellationPolicyService(settings), settings
}

func businessDraftMediaRefundPolicyTestContent(reference string) refundcancellation.Policy {
	policy := refundcancellation.DefaultPolicy()
	policy.Sections = []refundcancellation.Section{{ID: "proof", Title: "Proof", Body: "Keep this evidence", Image: &refundcancellation.Image{URL: reference, Alt: "Proof"}}}
	return policy
}

func TestBusinessDraftMediaRefundPolicyBindingAndSettingsWriteCommitOrRollBackTogether(t *testing.T) {
	for _, rollback := range []bool{false, true} {
		t.Run(map[bool]string{false: "committed", true: "rolled back"}[rollback], func(t *testing.T) {
			env := createBusinessDraftMediaReceiptTestEnvironment(t, nil)
			policies, _ := configureBusinessDraftMediaRefundPolicyTestService(t, env)
			_, err := policies.Update("en", refundcancellation.DefaultPolicy())
			require.NoError(t, err)
			asset := uploadBusinessDraftMediaReceiptTestAsset(t, env)
			if rollback {
				require.NoError(t, env.db.Callback().Update().Before("gorm:update").Register("reject_policy_setting_write", func(tx *gorm.DB) {
					if tx.Statement.Table == "settings" {
						tx.AddError(errors.New("policy write rejected"))
					}
				}))
			}
			_, err = policies.Update("en", businessDraftMediaRefundPolicyTestContent(asset.URL))
			if rollback {
				require.ErrorContains(t, err, "policy write rejected")
				require.NoError(t, env.db.Callback().Update().Remove("reject_policy_setting_write"))
			} else {
				require.NoError(t, err)
			}
			var receipt media.BusinessDraftMediaUploadReceipt
			require.NoError(t, env.db.First(&receipt).Error)
			if rollback {
				require.Equal(t, "ready", receipt.Status)
			} else {
				require.Equal(t, "bound", receipt.Status)
				require.ErrorIs(t, env.service.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
			}
		})
	}
}

func TestBusinessDraftMediaRefundPolicyHistoricalSettingsAndOrderSnapshotsRetainExactObjectsUntilDetached(t *testing.T) {
	for _, source := range []string{"soft deleted setting", "order policy snapshot"} {
		t.Run(source, func(t *testing.T) {
			env := createBusinessDraftMediaReceiptTestEnvironment(t, nil)
			_, _ = configureBusinessDraftMediaRefundPolicyTestService(t, env)
			asset := uploadBusinessDraftMediaReceiptTestAsset(t, env)
			encoded, err := json.Marshal(businessDraftMediaRefundPolicyTestContent(asset.URL))
			require.NoError(t, err)
			if source == "soft deleted setting" {
				row := setting.Setting{Key: refundcancellation.Key, Value: string(encoded), Locale: "fr", Group: refundcancellation.Group, Type: "json", IsPublic: false}
				require.NoError(t, env.db.Create(&row).Error)
				require.NoError(t, env.db.Delete(&row).Error)
			} else {
				require.NoError(t, env.db.Create(&order.PolicyDisclosure{OrderID: 91, PolicyKey: refundcancellation.Key, PolicyJSON: string(encoded)}).Error)
			}
			expireBusinessDraftMediaReceiptTestClock(t, env)
			result, err := env.service.ExpireUnboundBusinessDraftMediaUploads(t.Context(), 10)
			require.NoError(t, err)
			require.Equal(t, 1, result.Blocked)
			assertBusinessDraftMediaReceiptTestObjectsExist(t, env, asset, true)
			// Only remove synthetic history in the disposable fixture, never business data.
			if source == "soft deleted setting" {
				require.NoError(t, env.db.Unscoped().Where("key = ?", refundcancellation.Key).Delete(&setting.Setting{}).Error)
			} else {
				require.NoError(t, env.db.Where("order_id = ?", 91).Delete(&order.PolicyDisclosure{}).Error)
			}
			expireBusinessDraftMediaReceiptTestClock(t, env)
			result, err = env.service.ExpireUnboundBusinessDraftMediaUploads(t.Context(), 10)
			require.NoError(t, err)
			require.Equal(t, 1, result.Retired)
		})
	}
}

func TestBusinessDraftMediaRefundPolicyOrderSnapshotLocksActiveAssetWithoutChangingSnapshotBytesAndRejectsRetiredMedia(t *testing.T) {
	env := createBusinessDraftMediaReceiptTestEnvironment(t, nil)
	policies, settings := configureBusinessDraftMediaRefundPolicyTestService(t, env)
	asset := uploadBusinessDraftMediaReceiptTestAsset(t, env)
	_, err := policies.Update("en", businessDraftMediaRefundPolicyTestContent(asset.URL))
	require.NoError(t, err)
	snapshot, err := policies.BuildOrderDisclosure(settings, 91, "en", "", "", nil)
	require.NoError(t, err)
	originalJSON, originalHash := snapshot.PolicyJSON, snapshot.PolicyHash
	require.NoError(t, repository.NewOrderPolicyDisclosureRepository(env.db).Create(snapshot))
	require.Equal(t, originalJSON, snapshot.PolicyJSON)
	require.Equal(t, originalHash, snapshot.PolicyHash)
	_, err = policies.Update("en", refundcancellation.DefaultPolicy())
	require.NoError(t, err)
	require.ErrorIs(t, env.service.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
	require.NoError(t, env.db.Delete(snapshot).Error)
	require.NoError(t, env.service.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
	stale := *snapshot
	stale.ID = 0
	stale.OrderID = 92
	require.Error(t, repository.NewOrderPolicyDisclosureRepository(env.db).Create(&stale))
}

func TestBusinessDraftMediaRefundPolicyContactMediaLinkBindsAndRetainsItsImmutableOrderSnapshot(t *testing.T) {
	env := createBusinessDraftMediaReceiptTestEnvironment(t, nil)
	policies, settings := configureBusinessDraftMediaRefundPolicyTestService(t, env)
	asset := uploadBusinessDraftMediaReceiptTestAsset(t, env)
	policy := refundcancellation.DefaultPolicy()
	policy.ContactURL = asset.URL
	_, err := policies.Update("en", policy)
	require.NoError(t, err)
	var receipt media.BusinessDraftMediaUploadReceipt
	require.NoError(t, env.db.First(&receipt).Error)
	require.Equal(t, "bound", receipt.Status)
	snapshot, err := policies.BuildOrderDisclosure(settings, 91, "en", "", "", nil)
	require.NoError(t, err)
	require.NoError(t, repository.NewOrderPolicyDisclosureRepository(env.db).Create(snapshot))
	_, err = policies.Update("en", refundcancellation.DefaultPolicy())
	require.NoError(t, err)
	require.ErrorIs(t, env.service.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
}
