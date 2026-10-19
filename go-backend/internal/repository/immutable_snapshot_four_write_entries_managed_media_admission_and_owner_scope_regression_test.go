package repository

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/pkg/storage"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type immutableSnapshotAdmissionObjectIdentityAndExistenceTestResolver struct{}

func (immutableSnapshotAdmissionObjectIdentityAndExistenceTestResolver) ResolveTrustedMediaObjectKeyForPublicBusinessReference(reference string) (string, error) {
	return storage.TrustedBusinessMediaReferenceObjectKey(reference, nil, "")
}

func (immutableSnapshotAdmissionObjectIdentityAndExistenceTestResolver) ValidateMediaObjectExistsBeforeRegisteringOwnership(_ context.Context, key string) error {
	if strings.Contains(key, "missing-object") {
		return errors.New("test object is missing")
	}
	return nil
}

func createImmutableSnapshotAdmissionRegressionDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	connection, err := db.DB()
	require.NoError(t, err)
	connection.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, connection.Close()) })
	require.NoError(t, db.AutoMigrate(&media.MediaAsset{}, &media.MediaAssetDerivative{}, &media.BusinessDraftMediaUploadReceipt{}, &orderevidence.OrderEvidenceSnapshot{}, &orderevidence.OrderEvidenceSubmissionSnapshot{}, &orderevidence.OrderEvidenceExportSnapshot{}))
	return db
}

func writeImmutableSnapshotThroughSelectedRepositoryEntry(t *testing.T, db *gorm.DB, entry string, orderID uint, data []byte) error {
	t.Helper()
	digest := fmt.Sprintf("%x", sha256.Sum256(data))
	now := time.Now().UTC()
	switch entry {
	case "initial create":
		repo := NewOrderEvidenceSnapshotRepository(db)
		repo.ConfigureImmutableSnapshotMediaReferenceIdentityResolver(immutableSnapshotAdmissionObjectIdentityAndExistenceTestResolver{})
		return repo.Create(&orderevidence.OrderEvidenceSnapshot{OrderID: orderID, SchemaVersion: 1, ConfirmedAt: now, Currency: "USD", SnapshotData: data, SnapshotSHA256: digest})
	case "submission create", "submission create or get":
		snapshot := &orderevidence.OrderEvidenceSubmissionSnapshot{Provider: "stripe", DisputeID: orderID, OrderID: orderID, Version: 1, Status: "locked", SchemaVersion: 1, LockedAt: now, SnapshotData: data, SnapshotSHA256: digest}
		repo := NewOrderEvidenceSubmissionSnapshotRepository(db)
		repo.ConfigureImmutableSnapshotMediaReferenceIdentityResolver(immutableSnapshotAdmissionObjectIdentityAndExistenceTestResolver{})
		if entry == "submission create" {
			return repo.Create(snapshot)
		}
		_, err := repo.CreateOrGetLatest(snapshot)
		return err
	case "export create or get":
		repo := NewOrderEvidenceExportSnapshotRepository(db)
		repo.ConfigureImmutableSnapshotMediaReferenceIdentityResolver(immutableSnapshotAdmissionObjectIdentityAndExistenceTestResolver{})
		_, err := repo.CreateOrGetForPackage(&orderevidence.OrderEvidenceExportSnapshot{OrderID: orderID, EvidencePackageID: orderID, EvidencePackageVersion: 1, Version: 1, Status: "locked", SchemaVersion: 1, LockedAt: now, SnapshotData: data, SnapshotSHA256: digest})
		return err
	}
	t.Fatalf("unknown snapshot entry %q", entry)
	return nil
}

func TestImmutableSnapshotFourWriteEntriesRejectMissingRetiredUnboundAndDedicatedMedia(t *testing.T) {
	for _, entry := range []string{"initial create", "submission create", "submission create or get", "export create or get"} {
		for _, state := range []string{"missing", "missing physical object", "retired", "unbound draft", "bound draft", "private", "dedicated shared collision", "active", "external same basename"} {
			t.Run(entry+"/"+state, func(t *testing.T) {
				db := createImmutableSnapshotAdmissionRegressionDatabase(t)
				key := "media/snapshot-proof.png"
				if state == "missing physical object" {
					key = "media/missing-object.png"
				}
				if state == "unbound draft" || state == "bound draft" {
					key = storage.BusinessDraftMediaObjectPrefix + "/11111111-1111-4111-8111-111111111111/22222222-2222-4222-8222-222222222222.png"
				}
				if state == "dedicated shared collision" {
					key = "site-logo/snapshot-proof.png"
				}
				asset := media.MediaAsset{Filename: "proof.png", StorageKey: key, URL: "/uploads/" + key, Visibility: "public", Status: "active", MediaType: "image"}
				if state == "private" {
					asset.Visibility = "private"
				}
				if state != "missing" {
					require.NoError(t, db.Create(&asset).Error)
				}
				if state == "retired" {
					require.NoError(t, db.Delete(&asset).Error)
				}
				if state == "unbound draft" || state == "bound draft" {
					status := "ready"
					if state == "bound draft" {
						status = "bound"
					}
					require.NoError(t, db.Create(&media.BusinessDraftMediaUploadReceipt{ID: "11111111-1111-4111-8111-111111111111", AssetID: &asset.ID, Status: status, ExpiresAt: time.Now().Add(-time.Hour), OriginalObjectKey: key, ObjectKeysJSON: []byte("[]")}).Error)
				}
				url := asset.URL
				if state == "external same basename" {
					url = "https://foreign.example.test/uploads/" + key
				}
				data, err := json.Marshal(map[string]string{"content": `<p>proof<img src="` + url + `"></p>`})
				require.NoError(t, err)
				err = writeImmutableSnapshotThroughSelectedRepositoryEntry(t, db, entry, 1, data)
				if state == "active" || state == "bound draft" || state == "external same basename" {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
				}
			})
		}
	}
}

func TestImmutableSnapshotDedicatedReferencesRequireSameOrderRetainedProvenanceAndKeepOriginalPayload(t *testing.T) {
	for _, entry := range []string{"initial create", "submission create", "submission create or get", "export create or get"} {
		t.Run(entry, func(t *testing.T) {
			db := createImmutableSnapshotAdmissionRegressionDatabase(t)
			data := []byte(`{"content":"<img src=\"/uploads/site-logo/retained.png\">"}`)
			digest := fmt.Sprintf("%x", sha256.Sum256(data))
			if entry == "initial create" {
				historical := &orderevidence.OrderEvidenceSubmissionSnapshot{Provider: "stripe", OrderID: 77, DisputeID: 99, Version: 1, Status: "locked", SchemaVersion: 1, LockedAt: time.Now(), SnapshotData: data, SnapshotSHA256: digest}
				require.NoError(t, db.Create(historical).Error)
			} else {
				historical := &orderevidence.OrderEvidenceSnapshot{OrderID: 77, SchemaVersion: 1, Currency: "USD", ConfirmedAt: time.Now(), SnapshotData: data, SnapshotSHA256: digest}
				require.NoError(t, db.Create(historical).Error)
			}
			require.ErrorIs(t, writeImmutableSnapshotThroughSelectedRepositoryEntry(t, db, entry, 78, data), ErrImmutableOrderEvidenceSnapshotMediaReferenceUnavailable)
			require.NoError(t, writeImmutableSnapshotThroughSelectedRepositoryEntry(t, db, entry, 77, data))
		})
	}
}

func TestImmutableSnapshotManagedReferenceRequiresPhysicalObjectVerifierAndTransactionCopyPreservesConfiguration(t *testing.T) {
	db := createImmutableSnapshotAdmissionRegressionDatabase(t)
	asset := media.MediaAsset{Filename: "proof.png", StorageKey: "media/proof.png", URL: "/uploads/media/proof.png", Visibility: "public", Status: "active", MediaType: "image"}
	require.NoError(t, db.Create(&asset).Error)
	data := []byte(`{"image":"/uploads/media/proof.png"}`)
	snapshot := &orderevidence.OrderEvidenceSnapshot{OrderID: 1, SchemaVersion: 1, ConfirmedAt: time.Now(), Currency: "USD", SnapshotData: data, SnapshotSHA256: fmt.Sprintf("%x", sha256.Sum256(data))}
	repo := NewOrderEvidenceSnapshotRepository(db)
	require.ErrorIs(t, repo.Create(snapshot), ErrMediaObjectRegistrationVerificationUnavailable)
	repo.ConfigureImmutableSnapshotMediaReferenceIdentityResolver(immutableSnapshotAdmissionObjectIdentityAndExistenceTestResolver{})
	require.NoError(t, db.Transaction(func(tx *gorm.DB) error { return repo.WithTx(tx).Create(snapshot) }))
	var stored orderevidence.OrderEvidenceSnapshot
	require.NoError(t, db.First(&stored, snapshot.ID).Error)
	require.Equal(t, string(data), string(stored.SnapshotData))
	require.Equal(t, snapshot.SnapshotSHA256, stored.SnapshotSHA256)
}
