package service

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

func createReadOnlyPhysicalCapacityLocalTestProvider(t *testing.T) (storage.StorageService, string, string) {
	t.Helper()
	publicRoot, privateRoot := t.TempDir(), t.TempDir()
	provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: publicRoot, PrivateLocalPath: privateRoot, BaseURL: "https://capacity.example.test/uploads"})
	require.NoError(t, err)
	return provider, publicRoot, privateRoot
}

func writeReadOnlyPhysicalCapacityTestObject(t *testing.T, root, key string, size int) {
	t.Helper()
	target := filepath.Join(root, filepath.FromSlash(key))
	require.NoError(t, os.MkdirAll(filepath.Dir(target), 0700))
	require.NoError(t, os.WriteFile(target, make([]byte, size), 0600))
}

func createPhysicalCapacityTestCleanupReceipt(t *testing.T, key string, size *int64) outbox.Event {
	t.Helper()
	expected := map[string]int64{}
	if size != nil {
		expected[key] = *size
	}
	event, err := newObjectStorageCleanupEvent(objectCleanupResourceUnregisteredMediaUpload, "test", "media_asset", "test", []string{key}, expected)
	require.NoError(t, err)
	event.EventKey = repository.PermanentMediaOriginalUploadReceiptEventKeyPrefix + "test"
	event.Status = outbox.EventStatusPending
	return *event
}

func TestReadOnlyPhysicalCapacityDeduplicatesRegistrationsReceiptsAndObservedUploadPlansWithoutDeleting(t *testing.T) {
	provider, publicRoot, privateRoot := createReadOnlyPhysicalCapacityLocalTestProvider(t)
	privateKey := storage.CustomerServiceConversationAttachmentOriginalsPrefix + "/customer.png"
	for key, size := range map[string]int{"original.png": 10, "derivative.png": 4, "cleanup-only.png": 3, "unmatched.png": 2} {
		writeReadOnlyPhysicalCapacityTestObject(t, publicRoot, key, size)
	}
	writeReadOnlyPhysicalCapacityTestObject(t, privateRoot, privateKey, 6)
	known := int64(9)
	snapshot := repository.MediaPhysicalCapacityRegistrationSnapshot{
		Originals:                  []media.MediaAsset{{StorageKey: "original.png", Size: 10}, {StorageKey: "original.png", Size: 10, DeletedAt: gorm.DeletedAt{Time: time.Now(), Valid: true}}, {StorageKey: privateKey, Size: 6}, {StorageKey: "missing.png", Size: 7}},
		Derivatives:                []media.MediaAssetDerivative{{StorageKey: "derivative.png", Size: 4}},
		CleanupReceipts:            []outbox.Event{createPhysicalCapacityTestCleanupReceipt(t, "original.png", &known), createPhysicalCapacityTestCleanupReceipt(t, "cleanup-only.png", &known), createPhysicalCapacityTestCleanupReceipt(t, "planned.png", &known), createPhysicalCapacityTestCleanupReceipt(t, "unknown.png", nil)},
		PendingConversationUploads: []media.MediaUploadSession{{ObjectKey: "planned.png", ExpectedUploadBytes: &known}},
		BusinessDraftUploads:       []media.BusinessDraftMediaUploadReceipt{{OriginalObjectKey: "planned.png", Size: 9, ObjectKeysJSON: datatypes.JSON(`["planned.png"]`)}},
	}
	report, err := reconcileReadOnlyMediaPhysicalCapacitySnapshot(t.Context(), snapshot, provider, []storage.StorageService{provider}, time.Now().UTC())
	require.NoError(t, err)
	require.True(t, report.PhysicalScanComplete)
	require.Equal(t, 2, report.LocationsScanned)
	require.Equal(t, MediaPhysicalCapacityMetric{Objects: 5, Bytes: 25}, report.Observed)
	require.Equal(t, MediaPhysicalCapacityMetric{Objects: 2, Bytes: 16}, report.RegisteredOriginals)
	require.Equal(t, MediaPhysicalCapacityMetric{Objects: 1, Bytes: 4}, report.RegisteredDerivatives)
	require.Equal(t, MediaPhysicalCapacityMetric{Objects: 1, Bytes: 3}, report.CleanupReceiptOnly)
	require.Equal(t, MediaPhysicalCapacityMetric{Objects: 1, Bytes: 2}, report.UnmatchedRegistration)
	require.Equal(t, MediaPhysicalCapacityMetric{Objects: 2, Bytes: 13}, report.WithCleanupReceipt)
	require.Equal(t, MediaPhysicalCapacityMetric{Objects: 1, Bytes: 10}, report.WithSoftDeletedRegistration)
	require.Equal(t, MediaPhysicalCapacityMetric{Objects: 1, Bytes: 7}, report.RegisteredNotObserved)
	require.Equal(t, MediaPhysicalCapacityMetric{Objects: 1, Bytes: 9}, report.UnobservedUploadReservations)
	require.Equal(t, int64(1), report.UnknownUploadReservationObjects)
	for _, key := range []string{"original.png", "derivative.png", "cleanup-only.png", "unmatched.png"} {
		_, err := os.Stat(filepath.Join(publicRoot, key))
		require.NoError(t, err)
	}
	encoded, err := json.Marshal(report)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"bytes":"25"`)
	for _, sensitive := range []string{publicRoot, privateRoot, privateKey, "original.png"} {
		require.NotContains(t, string(encoded), sensitive)
	}
}

type readOnlyPhysicalCapacityFailingMetadataTestProvider struct {
	storage.StorageService
	metadata             storage.ReadOnlyPhysicalCapacityMetadataProvider
	fail                 bool
	duplicate            bool
	objectKeyResolutions int
}

func (p *readOnlyPhysicalCapacityFailingMetadataTestProvider) ObjectKey(reference string) (string, error) {
	p.objectKeyResolutions++
	return p.StorageService.ObjectKey(reference)
}

func (p *readOnlyPhysicalCapacityFailingMetadataTestProvider) PhysicalCapacityLocations() ([]storage.PhysicalCapacityLocation, error) {
	return p.metadata.PhysicalCapacityLocations()
}
func (p *readOnlyPhysicalCapacityFailingMetadataTestProvider) PhysicalCapacityObjectAddress(key string) (storage.PhysicalCapacityObjectAddress, error) {
	return p.metadata.PhysicalCapacityObjectAddress(key)
}
func (p *readOnlyPhysicalCapacityFailingMetadataTestProvider) WalkPhysicalCapacityObjectMetadata(ctx context.Context, location storage.PhysicalCapacityLocation, visit func(storage.PhysicalCapacityObjectMetadata) error) error {
	if p.fail {
		return errors.New("secret endpoint /private/customers/password")
	}
	return p.metadata.WalkPhysicalCapacityObjectMetadata(ctx, location, func(object storage.PhysicalCapacityObjectMetadata) error {
		if err := visit(object); err != nil {
			return err
		}
		if p.duplicate {
			return visit(object)
		}
		return nil
	})
}

func TestReadOnlyPhysicalCapacityIncompleteListingNeverClaimsRegisteredObjectsAreMissing(t *testing.T) {
	provider, _, _ := createReadOnlyPhysicalCapacityLocalTestProvider(t)
	failed := &readOnlyPhysicalCapacityFailingMetadataTestProvider{StorageService: provider, metadata: provider.(storage.ReadOnlyPhysicalCapacityMetadataProvider), fail: true}
	snapshot := repository.MediaPhysicalCapacityRegistrationSnapshot{Originals: []media.MediaAsset{{StorageKey: "not-listed.png", Size: 123}}}
	report, err := reconcileReadOnlyMediaPhysicalCapacitySnapshot(t.Context(), snapshot, failed, nil, time.Now().UTC())
	require.NoError(t, err)
	require.False(t, report.PhysicalScanComplete)
	require.Zero(t, report.RegisteredNotObserved.Objects)
	require.Equal(t, int64(1), report.RegisteredAbsenceUnverified)
	encoded, err := json.Marshal(report)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "password")
	require.NotContains(t, string(encoded), "/private/customers")
}

func TestReadOnlyPhysicalCapacityRepeatedProviderObjectsMakeScanIncompleteWithoutInflatingBytes(t *testing.T) {
	provider, publicRoot, _ := createReadOnlyPhysicalCapacityLocalTestProvider(t)
	writeReadOnlyPhysicalCapacityTestObject(t, publicRoot, "same.png", 17)
	duplicate := &readOnlyPhysicalCapacityFailingMetadataTestProvider{StorageService: provider, metadata: provider.(storage.ReadOnlyPhysicalCapacityMetadataProvider), duplicate: true}
	report, err := reconcileReadOnlyMediaPhysicalCapacitySnapshot(t.Context(), repository.MediaPhysicalCapacityRegistrationSnapshot{}, duplicate, nil, time.Now().UTC())
	require.NoError(t, err)
	require.False(t, report.PhysicalScanComplete)
	require.Equal(t, MediaPhysicalCapacityMetric{Objects: 1, Bytes: 17}, report.Observed)
	require.Contains(t, report.Issues, "duplicate_object_in_provider_listing")
}

func TestReadOnlyPhysicalCapacityIndependentNestedLogoStoreCountsOneObjectAndMatchesItsOwnReceipt(t *testing.T) {
	provider, publicRoot, _ := createReadOnlyPhysicalCapacityLocalTestProvider(t)
	logoRoot := filepath.Join(publicRoot, "logos")
	logo, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: logoRoot, BaseURL: "https://logos.example.test"})
	require.NoError(t, err)
	writeReadOnlyPhysicalCapacityTestObject(t, logoRoot, "site-logos/current.png", 8)
	receipt, err := newObjectStorageCleanupEvent(objectCleanupResourceSiteLogo, "logo", "site_logo", "logo", []string{"site-logos/current.png"})
	require.NoError(t, err)
	report, err := reconcileReadOnlyMediaPhysicalCapacitySnapshot(t.Context(), repository.MediaPhysicalCapacityRegistrationSnapshot{CleanupReceipts: []outbox.Event{*receipt}}, provider, []storage.StorageService{logo}, time.Now().UTC())
	require.NoError(t, err)
	require.True(t, report.PhysicalScanComplete)
	require.Equal(t, MediaPhysicalCapacityMetric{Objects: 1, Bytes: 8}, report.Observed)
	require.Equal(t, report.Observed, report.CleanupReceiptOnly)
}

func TestReadOnlyPhysicalCapacityConflictingSizesAndHistoricalUnknownReservationsStayExplicit(t *testing.T) {
	provider, publicRoot, _ := createReadOnlyPhysicalCapacityLocalTestProvider(t)
	writeReadOnlyPhysicalCapacityTestObject(t, publicRoot, "mismatch.png", 10)
	first, second := int64(3), int64(4)
	snapshot := repository.MediaPhysicalCapacityRegistrationSnapshot{
		Originals:       []media.MediaAsset{{StorageKey: "mismatch.png", Size: 8}, {StorageKey: "conflict.png", Size: 5}, {StorageKey: "conflict.png", Size: 6}, {StorageKey: "https://foreign.example.test/x.png", Size: 5}},
		CleanupReceipts: []outbox.Event{createPhysicalCapacityTestCleanupReceipt(t, "conflicting-plan.png", &first), createPhysicalCapacityTestCleanupReceipt(t, "conflicting-plan.png", &second), createPhysicalCapacityTestCleanupReceipt(t, "historical.png", nil)},
	}
	report, err := reconcileReadOnlyMediaPhysicalCapacitySnapshot(t.Context(), snapshot, provider, nil, time.Now().UTC())
	require.NoError(t, err)
	require.Equal(t, int64(1), report.RegisteredSizeMismatches)
	require.Equal(t, int64(1), report.ConflictingRecordedSizes)
	require.Equal(t, int64(2), report.UnknownUploadReservationObjects)
	require.Zero(t, report.UnobservedUploadReservations.Bytes)
	require.Equal(t, int64(1), report.RegisteredAbsenceUnverified)
	require.Equal(t, int64(1), report.UnresolvedRecordLocations)
}

func TestReadOnlyPhysicalCapacityHistoricalZeroSizeAndForeignOrConflictingURLsRemainUnverified(t *testing.T) {
	provider, publicRoot, _ := createReadOnlyPhysicalCapacityLocalTestProvider(t)
	writeReadOnlyPhysicalCapacityTestObject(t, publicRoot, "measured.png", 5)
	writeReadOnlyPhysicalCapacityTestObject(t, publicRoot, "same-key.png", 7)
	snapshot := repository.MediaPhysicalCapacityRegistrationSnapshot{Originals: []media.MediaAsset{
		{StorageKey: "measured.png", Size: 0}, {StorageKey: "absent-unknown.png", Size: 0},
		{StorageKey: "same-key.png", URL: "https://retired-storage.example.test/uploads/same-key.png", Size: 7},
		{StorageKey: "conflict.png", URL: provider.GetURL("different.png"), Size: 3},
	}}
	report, err := reconcileReadOnlyMediaPhysicalCapacitySnapshot(t.Context(), snapshot, provider, nil, time.Now().UTC())
	require.NoError(t, err)
	require.Equal(t, int64(2), report.RegisteredUnknownRecordedSizeObjects)
	require.Equal(t, int64(1), report.RegisteredAbsenceUnverified)
	require.Equal(t, int64(2), report.UnresolvedRecordLocations)
	require.Equal(t, MediaPhysicalCapacityMetric{Objects: 1, Bytes: 7}, report.UnmatchedRegistration)
	require.Zero(t, report.RegisteredSizeMismatches)
	require.Contains(t, report.Issues, "registration_location_or_key_conflict")
}

func TestReadOnlyPhysicalCapacityCopiedMediaServicesShareAuditGuardAndSnapshotFailuresReleaseIt(t *testing.T) {
	provider, _, _ := createReadOnlyPhysicalCapacityLocalTestProvider(t)
	service := NewMediaService(repository.NewMediaRepository(nil), provider, nil, "", 10)
	transactionalCopy := *service
	service.capacityAuditRunning.Store(true)
	_, err := transactionalCopy.AuditReadOnlyMediaPhysicalCapacity(t.Context())
	require.ErrorIs(t, err, ErrMediaPhysicalCapacityAuditAlreadyRunning)
	service.capacityAuditRunning.Store(false)
	for range 2 {
		_, err := service.AuditReadOnlyMediaPhysicalCapacity(t.Context())
		require.ErrorIs(t, err, gorm.ErrInvalidDB)
		require.False(t, transactionalCopy.capacityAuditRunning.Load())
	}
}

func TestReadOnlyPhysicalCapacityUploadByteMetadataDoesNotAddCleanupKeysOrRewriteLegacyPayloads(t *testing.T) {
	event, err := newObjectStorageCleanupEvent("test", "test", "media_asset", "test", []string{"owned.png"}, map[string]int64{"owned.png": 9, "foreign.png": 99})
	require.NoError(t, err)
	var payload outbox.ObjectStorageCleanupPayload
	require.NoError(t, json.Unmarshal(event.Payload, &payload))
	require.Equal(t, []string{"owned.png"}, payload.ObjectKeys)
	require.Equal(t, map[string]int64{"owned.png": 9}, payload.ExpectedUploadObjectBytes)
	legacy, err := newObjectStorageCleanupEvent("test", "legacy", "media_asset", "legacy", []string{"old.png"})
	require.NoError(t, err)
	require.NotContains(t, string(legacy.Payload), "expected_upload_object_bytes")
}

func TestReadOnlyPhysicalCapacityCancellationStopsResolvingRemainingRegistrationLocations(t *testing.T) {
	provider, _, _ := createReadOnlyPhysicalCapacityLocalTestProvider(t)
	tracked := &readOnlyPhysicalCapacityFailingMetadataTestProvider{StorageService: provider, metadata: provider.(storage.ReadOnlyPhysicalCapacityMetadataProvider)}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	snapshot := repository.MediaPhysicalCapacityRegistrationSnapshot{Originals: []media.MediaAsset{{StorageKey: "known.png", URL: provider.GetURL("known.png"), Size: 1}}}
	report, err := reconcileReadOnlyMediaPhysicalCapacitySnapshot(ctx, snapshot, tracked, nil, time.Now().UTC())
	require.NoError(t, err)
	require.Zero(t, tracked.objectKeyResolutions)
	require.False(t, report.PhysicalScanComplete)
	require.Contains(t, report.Issues, "audit_cancelled_or_timed_out")
}
