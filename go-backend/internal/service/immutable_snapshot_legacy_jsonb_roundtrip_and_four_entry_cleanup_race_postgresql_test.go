package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/repository"
)

func TestPostgreSQLImmutableSnapshotPayPalInvoiceExportRetainsSameOrderHistoryAfterReceiptConsumption(t *testing.T) {
	db, uploads, provider := createIsolatedPostgreSQLSuggestionAndPayPalAttachmentReceiptEnvironment(t)
	verifyPayPalInvoiceExportRetainsSameOrderHistoryAfterReceiptConsumption(t, db, uploads, provider)
}

func TestPostgreSQLImmutableSnapshotLegacyHashesSurviveJSONBRoundtripWithoutChangingStoredHistory(t *testing.T) {
	env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	db := env.db
	require.NoError(t, db.AutoMigrate(&orderevidence.OrderEvidenceSnapshot{}, &orderevidence.OrderEvidenceSubmissionSnapshot{}, &orderevidence.OrderEvidenceExportSnapshot{}))
	var rawJSON string
	require.NoError(t, db.Raw("SELECT CAST(? AS jsonb)", `{"long_key":9007199254740993,"z":1.23e-4}`).Scan(&rawJSON).Error)
	for _, kind := range []string{"initial", "stripe", "paypal", "export", "new canonical decimal"} {
		t.Run(kind, func(t *testing.T) {
			var payload any
			switch kind {
			case "initial":
				payload = orderevidence.OrderEvidenceSnapshotPayload{SchemaVersion: 1, Currency: "USD", Items: []orderevidence.OrderEvidenceSnapshotItem{{OrderItemID: 9007199254740993, SelectedSpecsJSON: json.RawMessage(`{"long_key":1,"z":2}`)}}}
			case "stripe":
				payload = stripeDisputeEvidenceSnapshotPayload{DisputeEvidenceSnapshotEnvelope: disputeEvidenceSnapshotEnvelope{Provider: "stripe", OrderID: 9007199254740993}, Request: stripeDisputeEvidenceSnapshotRequest{AdditionalStatement: "original"}}
			case "paypal":
				payload = paypalDisputeEvidenceSnapshotPayload{DisputeEvidenceSnapshotEnvelope: disputeEvidenceSnapshotEnvelope{Provider: "paypal", OrderID: 9007199254740993}, Request: paypalDisputeEvidenceSnapshotRequest{Notes: "original"}}
			default:
				payload = OrderEvidenceExportSnapshotPayload{Order: &order.Order{ID: 9007199254740993, ShippingPlanSnapshotData: []byte(rawJSON)}, Package: &OrderEvidenceExportPackage{Items: []OrderEvidenceExportItem{{DataJSON: []byte(rawJSON)}}}}
			}
			data, err := json.Marshal(payload)
			require.NoError(t, err)
			if kind == "new canonical decimal" {
				data, err = orderevidence.CanonicalImmutableSnapshotJSONForStorageIndependentIntegrity([]byte(`{"id":9007199254740993,"raw":{"decimal":1.2300e-4,"large":1e21}}`))
				require.NoError(t, err)
			}
			digest := fmt.Sprintf("%x", sha256.Sum256(data))
			var storedData []byte
			var storedDigest string
			switch kind {
			case "initial":
				snapshot := &orderevidence.OrderEvidenceSnapshot{OrderID: 44, SchemaVersion: 1, Currency: "USD", ConfirmedAt: time.Now(), SnapshotData: data, SnapshotSHA256: digest}
				require.NoError(t, db.Create(snapshot).Error)
				found, err := repository.NewOrderEvidenceSnapshotRepository(db).FindByOrderID(44)
				require.NoError(t, err)
				require.NoError(t, found.Validate())
				storedData, storedDigest = found.SnapshotData, found.SnapshotSHA256
			case "stripe", "paypal":
				snapshot := &orderevidence.OrderEvidenceSubmissionSnapshot{Provider: kind, DisputeID: 44, OrderID: 44, Version: 1, Status: "locked", SchemaVersion: 1, LockedAt: time.Now(), SnapshotData: data, SnapshotSHA256: digest}
				require.NoError(t, db.Create(snapshot).Error)
				found, err := repository.NewOrderEvidenceSubmissionSnapshotRepository(db).FindByID(snapshot.ID)
				require.NoError(t, err)
				require.NoError(t, found.Validate())
				storedData, storedDigest = found.SnapshotData, found.SnapshotSHA256
			default:
				packageID := uint(44)
				if kind == "new canonical decimal" {
					packageID = 45
				}
				snapshot := &orderevidence.OrderEvidenceExportSnapshot{OrderID: 44, EvidencePackageID: packageID, EvidencePackageVersion: 1, Version: 1, Status: "locked", SchemaVersion: 1, LockedAt: time.Now(), SnapshotData: data, SnapshotSHA256: digest}
				require.NoError(t, db.Create(snapshot).Error)
				found, err := repository.NewOrderEvidenceExportSnapshotRepository(db).FindByPackageID(packageID)
				require.NoError(t, err)
				require.NoError(t, found.Validate())
				storedData, storedDigest = found.SnapshotData, found.SnapshotSHA256
			}
			require.Equal(t, digest, storedDigest)
			require.NotEqual(t, string(data), string(storedData))
			tampered := append([]byte(nil), storedData...)
			for index := range tampered {
				if tampered[index] == '9' {
					tampered[index] = '8'
					break
				}
			}
			switch kind {
			case "initial":
				require.Error(t, (orderevidence.OrderEvidenceSnapshot{SnapshotData: tampered, SnapshotSHA256: storedDigest}).VerifyIntegrity())
			case "stripe", "paypal":
				require.Error(t, (orderevidence.OrderEvidenceSubmissionSnapshot{Provider: kind, SnapshotData: tampered, SnapshotSHA256: storedDigest}).VerifyIntegrity())
			default:
				require.Error(t, (orderevidence.OrderEvidenceExportSnapshot{SnapshotData: tampered, SnapshotSHA256: storedDigest}).VerifyIntegrity())
			}
		})
	}
}

func persistImmutableSnapshotThroughRepositoryEntryForCleanupRace(db *gorm.DB, mediaService *MediaService, entry string, data []byte) error {
	canonical, err := orderevidence.CanonicalImmutableSnapshotJSONForStorageIndependentIntegrity(data)
	if err != nil {
		return err
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(canonical))
	now := time.Now().UTC()
	switch entry {
	case "initial":
		repo := repository.NewOrderEvidenceSnapshotRepository(db)
		repo.ConfigureImmutableSnapshotMediaReferenceIdentityResolver(mediaService)
		return repo.Create(&orderevidence.OrderEvidenceSnapshot{OrderID: 44, SchemaVersion: 1, Currency: "USD", ConfirmedAt: now, SnapshotData: canonical, SnapshotSHA256: digest})
	case "submission create", "submission canonical":
		repo := repository.NewOrderEvidenceSubmissionSnapshotRepository(db)
		repo.ConfigureImmutableSnapshotMediaReferenceIdentityResolver(mediaService)
		snapshot := &orderevidence.OrderEvidenceSubmissionSnapshot{Provider: "stripe", DisputeID: 44, OrderID: 44, Version: 1, Status: "locked", SchemaVersion: 1, LockedAt: now, SnapshotData: canonical, SnapshotSHA256: digest}
		if entry == "submission create" {
			return repo.Create(snapshot)
		}
		_, err := repo.CreateOrGetLatest(snapshot)
		return err
	default:
		repo := repository.NewOrderEvidenceExportSnapshotRepository(db)
		repo.ConfigureImmutableSnapshotMediaReferenceIdentityResolver(mediaService)
		_, err := repo.CreateOrGetForPackage(&orderevidence.OrderEvidenceExportSnapshot{OrderID: 44, EvidencePackageID: 44, EvidencePackageVersion: 1, Version: 1, Status: "locked", SchemaVersion: 1, LockedAt: now, SnapshotData: canonical, SnapshotSHA256: digest})
		return err
	}
}

func TestPostgreSQLImmutableSnapshotFourWriteEntriesSerializeRichTextBindingAndPhysicalCleanup(t *testing.T) {
	for _, entry := range []string{"initial", "submission create", "submission canonical", "export"} {
		for _, firstOperation := range []string{"snapshot", "cleanup"} {
			t.Run(entry+"/"+firstOperation+" first", func(t *testing.T) {
				env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
				require.NoError(t, env.db.AutoMigrate(&orderevidence.OrderEvidenceSnapshot{}, &orderevidence.OrderEvidenceSubmissionSnapshot{}, &orderevidence.OrderEvidenceExportSnapshot{}))
				ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
				defer cancel()
				db := env.db.WithContext(ctx)
				assets := NewMediaService(repository.NewMediaRepository(db), env.storage.StorageService, nil, "https://shop.example.test", 20<<30)
				assets.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
				asset := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, env.storage.StorageService, "media/immutable-snapshot-race.png")
				writeMediaCleanupOwnershipConflictTestObjects(t, env.publicRoot, asset.StorageKey)
				data, err := json.Marshal(map[string]string{"content": `<p>proof<img src="` + asset.URL + `"></p>`})
				require.NoError(t, err)
				acquired, released := make(chan int, 1), make(chan struct{})
				var releaseOnce, pauseOnce sync.Once
				release := func() { releaseOnce.Do(func() { close(released) }) }
				defer release()
				pause := func(tx *gorm.DB) {
					pauseOnce.Do(func() { pausePostgreSQLMediaObjectOwnershipTransactionUntilReleased(ctx, t, tx, acquired, released) })
				}
				const callback = "pause_immutable_snapshot_cleanup_race"
				if firstOperation == "snapshot" {
					require.NoError(t, db.Callback().Create().After("gorm:create").Register(callback, func(tx *gorm.DB) {
						if tx.Statement.Table == "order_evidence_snapshots" || tx.Statement.Table == "order_evidence_submission_snapshots" || tx.Statement.Table == "order_evidence_export_snapshots" {
							pause(tx)
						}
					}))
					defer db.Callback().Create().Remove(callback)
				} else {
					require.NoError(t, db.Callback().Delete().After("gorm:delete").Register(callback, func(tx *gorm.DB) {
						if tx.Statement.Table == "media_assets" {
							pause(tx)
						}
					}))
					defer db.Callback().Delete().Remove(callback)
				}
				writer, cleaner := make(chan error, 1), make(chan error, 1)
				writeSnapshot := func() {
					writer <- persistImmutableSnapshotThroughRepositoryEntryForCleanupRace(db, assets, entry, data)
				}
				cleanup := func() { cleaner <- assets.DeleteAsset(ctx, asset.ID, MediaAssetDeleteConfirmation(asset.ID)) }
				if firstOperation == "snapshot" {
					go writeSnapshot()
				} else {
					go cleanup()
				}
				blocker := awaitPostgreSQLMediaObjectOwnershipTransactionPID(ctx, t, acquired)
				if firstOperation == "snapshot" {
					go cleanup()
				} else {
					go writeSnapshot()
				}
				require.Eventually(t, func() bool {
					var blocked bool
					err := db.Raw("SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE ? = ANY(pg_blocking_pids(pid)))", blocker).Scan(&blocked).Error
					return err == nil && blocked
				}, 4*time.Second, 20*time.Millisecond)
				release()
				writeErr := awaitPostgreSQLMediaObjectOwnershipOperationResult(ctx, t, writer)
				cleanupErr := awaitPostgreSQLMediaObjectOwnershipOperationResult(ctx, t, cleaner)
				if firstOperation == "snapshot" {
					require.NoError(t, writeErr)
					require.ErrorIs(t, cleanupErr, ErrMediaAssetInUse)
				} else {
					require.NoError(t, cleanupErr)
					require.Error(t, writeErr)
				}
			})
		}
	}
}
