package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"commerce-platform/internal/domain/homevisualtile"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

var ErrMediaPhysicalCapacityAuditAlreadyRunning = errors.New("media physical capacity audit is already running")

const mediaPhysicalCapacityAuditMaximumRecords = 100000
const mediaPhysicalCapacityAuditMaximumObjectObservations = 200000

type MediaPhysicalCapacityMetric struct {
	Objects int64 `json:"objects"`
	Bytes   int64 `json:"bytes,string"`
}

type MediaPhysicalCapacityAuditReport struct {
	StartedAt                            time.Time                   `json:"started_at"`
	CompletedAt                          time.Time                   `json:"completed_at"`
	PhysicalScanComplete                 bool                        `json:"physical_scan_complete"`
	LocationsScanned                     int                         `json:"locations_scanned"`
	LocationsIncomplete                  int                         `json:"locations_incomplete"`
	Observed                             MediaPhysicalCapacityMetric `json:"observed"`
	RegisteredOriginals                  MediaPhysicalCapacityMetric `json:"registered_originals"`
	RegisteredDerivatives                MediaPhysicalCapacityMetric `json:"registered_derivatives"`
	CleanupReceiptOnly                   MediaPhysicalCapacityMetric `json:"cleanup_receipt_only"`
	UnmatchedRegistration                MediaPhysicalCapacityMetric `json:"unmatched_registration"`
	WithCleanupReceipt                   MediaPhysicalCapacityMetric `json:"with_cleanup_receipt"`
	WithUploadReceipt                    MediaPhysicalCapacityMetric `json:"with_upload_receipt"`
	WithSoftDeletedRegistration          MediaPhysicalCapacityMetric `json:"with_soft_deleted_registration"`
	RegisteredNotObserved                MediaPhysicalCapacityMetric `json:"registered_not_observed"`
	RegisteredAbsenceUnverified          int64                       `json:"registered_absence_unverified"`
	RegisteredSizeMismatches             int64                       `json:"registered_size_mismatches"`
	ConflictingRecordedSizes             int64                       `json:"conflicting_recorded_sizes"`
	RegisteredUnknownRecordedSizeObjects int64                       `json:"registered_unknown_recorded_size_objects"`
	UnobservedUploadReservations         MediaPhysicalCapacityMetric `json:"unobserved_upload_reservations"`
	UnknownUploadReservationObjects      int64                       `json:"unknown_upload_reservation_objects"`
	UnresolvedRecordLocations            int64                       `json:"unresolved_record_locations"`
	Issues                               []string                    `json:"issues"`
}

type mediaPhysicalCapacityObjectEvidence struct {
	address                                            storage.PhysicalCapacityObjectAddress
	original, derivative, softDeleted, cleanup, upload bool
	recordedSize, reservation                          *int64
	recordedConflict, reservationConflict              bool
}

type mediaPhysicalCapacityLocationScan struct {
	provider storage.ReadOnlyPhysicalCapacityMetadataProvider
	location storage.PhysicalCapacityLocation
}

// Configuration is performed once at startup. Extra providers include the
// independent logo store; shared buckets/roots and nested roots are deduplicated.
func (s *MediaService) ConfigureAdditionalPhysicalCapacityAuditStorageProviders(providers ...storage.StorageService) {
	if s != nil {
		s.capacityAuditAdditionalStorageProviders = append([]storage.StorageService(nil), providers...)
	}
}

func addMediaPhysicalCapacityMetric(metric *MediaPhysicalCapacityMetric, size int64) error {
	if size < 0 || metric.Bytes > math.MaxInt64-size {
		return fmt.Errorf("physical capacity bytes overflow")
	}
	metric.Objects++
	metric.Bytes += size
	return nil
}

func mergeMediaPhysicalCapacityRecordedSize(destination **int64, conflict *bool, size int64) {
	if size < 0 {
		*conflict = true
		return
	}
	if *destination == nil {
		copyOfSize := size
		*destination = &copyOfSize
		return
	}
	if **destination != size {
		*conflict = true
	}
}

func isOutstandingMediaPhysicalCapacityUploadReceipt(event outbox.Event) bool {
	if event.Status != outbox.EventStatusPending && event.Status != outbox.EventStatusProcessing && event.Status != outbox.EventStatusFailed {
		return false
	}
	for _, prefix := range []string{repository.PermanentMediaOriginalUploadReceiptEventKeyPrefix, repository.DedicatedImmediateOwnerUploadReceiptEventKeyPrefix,
		repository.OrderEvidenceAttachmentPrecommittedObjectWriteReceiptEventKeyPrefix, repository.PrivateOrderBusinessAttachmentReceiptEventKeyPrefix,
		repository.UGCShowcasePrecommittedObjectWriteReceiptEventKeyPrefix, repository.PublicMediaDerivativeUploadReceiptEventKeyPrefix,
		outbox.SuggestionAndPayPalAttachmentReceiptEventKeyPrefix, homevisualtile.VisualShowcaseDraftUploadReceiptEventKeyPrefix} {
		if strings.HasPrefix(event.EventKey, prefix) {
			return true
		}
	}
	return false
}

func (s *MediaService) AuditReadOnlyMediaPhysicalCapacity(ctx context.Context) (*MediaPhysicalCapacityAuditReport, error) {
	if s == nil || s.repo == nil || s.storage == nil || s.capacityAuditRunning == nil {
		return nil, ErrMediaStorageUnavailable
	}
	if !s.capacityAuditRunning.CompareAndSwap(false, true) {
		return nil, ErrMediaPhysicalCapacityAuditAlreadyRunning
	}
	defer s.capacityAuditRunning.Store(false)
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	started := time.Now().UTC()
	snapshot, err := s.repo.ReadOnlyMediaPhysicalCapacityRegistrationSnapshot(ctx, mediaPhysicalCapacityAuditMaximumRecords)
	if err != nil {
		return nil, err
	}
	return reconcileReadOnlyMediaPhysicalCapacitySnapshot(ctx, snapshot, s.storage, s.capacityAuditAdditionalStorageProviders, started)
}

// No key, owner identifier, root path, endpoint, provider error or credential
// leaves this function. Unmatched objects are accounting observations only.
func reconcileReadOnlyMediaPhysicalCapacitySnapshot(ctx context.Context, snapshot repository.MediaPhysicalCapacityRegistrationSnapshot, primary storage.StorageService, additional []storage.StorageService, started time.Time) (*MediaPhysicalCapacityAuditReport, error) {
	report := &MediaPhysicalCapacityAuditReport{StartedAt: started, PhysicalScanComplete: true, Issues: []string{}}
	issues := make(map[string]bool)
	providers := append([]storage.StorageService{primary}, additional...)
	locations := make(map[string]mediaPhysicalCapacityLocationScan)
	for _, provider := range providers {
		metadata, supported := provider.(storage.ReadOnlyPhysicalCapacityMetadataProvider)
		if !supported {
			report.PhysicalScanComplete = false
			report.LocationsIncomplete++
			issues["provider_metadata_listing_unavailable"] = true
			continue
		}
		configured, err := metadata.PhysicalCapacityLocations()
		if err != nil || len(configured) == 0 {
			report.PhysicalScanComplete = false
			report.LocationsIncomplete++
			issues["storage_location_identity_unavailable"] = true
			continue
		}
		for _, location := range configured {
			if _, exists := locations[location.ID]; !exists {
				locations[location.ID] = mediaPhysicalCapacityLocationScan{metadata, location}
			}
		}
	}
	evidence := make(map[string]*mediaPhysicalCapacityObjectEvidence)
	resolveEvidence := func(provider storage.StorageService, reference string) *mediaPhysicalCapacityObjectEvidence {
		if ctx.Err() != nil {
			report.UnresolvedRecordLocations++
			return nil
		}
		metadata, ok := provider.(storage.ReadOnlyPhysicalCapacityMetadataProvider)
		if !ok || reference == "" {
			report.UnresolvedRecordLocations++
			return nil
		}
		key, err := provider.ObjectKey(reference)
		if err != nil {
			report.UnresolvedRecordLocations++
			return nil
		}
		address, err := metadata.PhysicalCapacityObjectAddress(key)
		if err != nil {
			report.UnresolvedRecordLocations++
			return nil
		}
		entry := evidence[address.ObjectID]
		if entry == nil {
			if len(evidence) >= mediaPhysicalCapacityAuditMaximumObjectObservations {
				report.UnresolvedRecordLocations++
				issues["record_object_limit_reached"] = true
				return nil
			}
			entry = &mediaPhysicalCapacityObjectEvidence{address: address}
			evidence[address.ObjectID] = entry
		}
		return entry
	}
	resolveRegisteredEvidence := func(key, reference string) *mediaPhysicalCapacityObjectEvidence {
		if ctx.Err() != nil {
			report.UnresolvedRecordLocations++
			return nil
		}
		if reference != "" {
			resolvedKey, err := primary.ObjectKey(reference)
			if err != nil {
				report.UnresolvedRecordLocations++
				issues["registration_location_or_key_conflict"] = true
				return nil
			}
			if key == "" {
				key = resolvedKey
			}
			var verifiedKey string
			if storage.IsPrivateObjectKey(key) {
				verifiedKey, err = storage.ResolvePrivateBusinessAttachmentObjectKeyForAuthorizedRead(primary, reference)
			} else {
				verifiedKey, err = storage.TrustedBusinessMediaReferenceObjectKey(reference, primary, "")
			}
			if err != nil || verifiedKey == "" || verifiedKey != key {
				report.UnresolvedRecordLocations++
				issues["registration_location_or_key_conflict"] = true
				return nil
			}
		}
		return resolveEvidence(primary, key)
	}
	for _, asset := range snapshot.Originals {
		if entry := resolveRegisteredEvidence(asset.StorageKey, asset.URL); entry != nil {
			entry.original = true
			entry.softDeleted = entry.softDeleted || asset.DeletedAt.Valid
			if asset.Size > 0 {
				mergeMediaPhysicalCapacityRecordedSize(&entry.recordedSize, &entry.recordedConflict, asset.Size)
			}
		}
	}
	for _, derivative := range snapshot.Derivatives {
		if entry := resolveRegisteredEvidence(derivative.StorageKey, derivative.URL); entry != nil {
			entry.derivative = true
			entry.softDeleted = entry.softDeleted || derivative.DeletedAt.Valid
			if derivative.Size > 0 {
				mergeMediaPhysicalCapacityRecordedSize(&entry.recordedSize, &entry.recordedConflict, derivative.Size)
			}
		}
	}
	for _, event := range snapshot.CleanupReceipts {
		provider := primary
		var payload struct {
			outbox.ObjectStorageCleanupPayload
			URL                  string   `json:"url"`
			Size                 *int64   `json:"size"`
			ExpectedUploadBytes  *int64   `json:"expected_upload_bytes"`
			AttachmentReferences []string `json:"attachment_references"`
		}
		if json.Unmarshal(event.Payload, &payload) != nil {
			issues["cleanup_receipt_metadata_invalid"] = true
			report.UnresolvedRecordLocations++
			continue
		}
		if payload.ResourceType == objectCleanupResourceSiteLogo {
			if len(additional) == 0 {
				report.UnresolvedRecordLocations++
				issues["dedicated_logo_location_unavailable"] = true
				continue
			}
			provider = additional[0]
		}
		keys := payload.ObjectKeys
		if event.EventType == outbox.EventTypeCustomerServiceAvatarCleanup {
			keys = []string{payload.URL}
		}
		if event.EventType == outbox.EventTypeCustomerServiceRetentionCleanup {
			keys = payload.AttachmentReferences
		}
		if len(keys) == 0 {
			report.UnresolvedRecordLocations++
			issues["cleanup_receipt_has_no_exact_object_identity"] = true
			continue
		}
		for _, key := range keys {
			entry := resolveEvidence(provider, key)
			if entry == nil {
				continue
			}
			entry.cleanup = true
			if !isOutstandingMediaPhysicalCapacityUploadReceipt(event) {
				continue
			}
			entry.upload = true
			if size, known := payload.ExpectedUploadObjectBytes[key]; known {
				mergeMediaPhysicalCapacityRecordedSize(&entry.reservation, &entry.reservationConflict, size)
			} else if len(keys) == 1 {
				if event.EventType == outbox.EventTypeCustomerServiceAvatarCleanup && payload.ExpectedUploadBytes != nil {
					mergeMediaPhysicalCapacityRecordedSize(&entry.reservation, &entry.reservationConflict, *payload.ExpectedUploadBytes)
				} else if (strings.HasPrefix(event.EventKey, outbox.SuggestionAndPayPalAttachmentReceiptEventKeyPrefix) || strings.HasPrefix(event.EventKey, homevisualtile.VisualShowcaseDraftUploadReceiptEventKeyPrefix)) && payload.Size != nil {
					mergeMediaPhysicalCapacityRecordedSize(&entry.reservation, &entry.reservationConflict, *payload.Size)
				}
			}
		}
	}
	for _, session := range snapshot.PendingConversationUploads {
		if entry := resolveEvidence(primary, session.ObjectKey); entry != nil {
			entry.upload = true
			if session.ExpectedUploadBytes != nil {
				mergeMediaPhysicalCapacityRecordedSize(&entry.reservation, &entry.reservationConflict, *session.ExpectedUploadBytes)
			}
		}
	}
	for _, draft := range snapshot.BusinessDraftUploads {
		var keys []string
		if json.Unmarshal(draft.ObjectKeysJSON, &keys) != nil {
			report.UnresolvedRecordLocations++
			issues["draft_upload_metadata_invalid"] = true
			continue
		}
		for _, key := range keys {
			if entry := resolveEvidence(primary, key); entry != nil {
				entry.upload = true
				if key == draft.OriginalObjectKey && draft.Size > 0 {
					mergeMediaPhysicalCapacityRecordedSize(&entry.reservation, &entry.reservationConflict, draft.Size)
				}
			}
		}
	}

	observed := make(map[string]int64)
	completeLocations := make(map[string]bool)
	locationIDs := make([]string, 0, len(locations))
	for id := range locations {
		locationIDs = append(locationIDs, id)
	}
	sort.Strings(locationIDs)
	observations := 0
	for _, id := range locationIDs {
		scan := locations[id]
		seenInLocation := make(map[string]bool)
		err := scan.provider.WalkPhysicalCapacityObjectMetadata(ctx, scan.location, func(object storage.PhysicalCapacityObjectMetadata) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			observations++
			if observations > mediaPhysicalCapacityAuditMaximumObjectObservations {
				issues["object_observation_limit_reached"] = true
				return fmt.Errorf("capacity observation limit reached")
			}
			if object.Size < 0 || object.ObjectID == "" || object.LocationID != id {
				return fmt.Errorf("invalid capacity metadata")
			}
			if seenInLocation[object.ObjectID] {
				issues["duplicate_object_in_provider_listing"] = true
				return fmt.Errorf("capacity provider repeated an object")
			}
			seenInLocation[object.ObjectID] = true
			if previous, exists := observed[object.ObjectID]; exists {
				if previous != object.Size {
					issues["object_changed_during_scan"] = true
					return fmt.Errorf("capacity metadata changed during scan")
				}
				return nil
			}
			if err := addMediaPhysicalCapacityMetric(&report.Observed, object.Size); err != nil {
				return err
			}
			observed[object.ObjectID] = object.Size
			return nil
		})
		if err != nil {
			report.PhysicalScanComplete = false
			report.LocationsIncomplete++
			issues["storage_metadata_scan_incomplete"] = true
		} else {
			report.LocationsScanned++
			completeLocations[id] = true
		}
	}
	for objectID, size := range observed {
		entry := evidence[objectID]
		metric := &report.UnmatchedRegistration
		if entry != nil {
			switch {
			case entry.original:
				metric = &report.RegisteredOriginals
			case entry.derivative:
				metric = &report.RegisteredDerivatives
			case entry.cleanup:
				metric = &report.CleanupReceiptOnly
			}
			if entry.cleanup {
				if err := addMediaPhysicalCapacityMetric(&report.WithCleanupReceipt, size); err != nil {
					return nil, err
				}
			}
			if entry.upload {
				if err := addMediaPhysicalCapacityMetric(&report.WithUploadReceipt, size); err != nil {
					return nil, err
				}
			}
			if entry.softDeleted {
				if err := addMediaPhysicalCapacityMetric(&report.WithSoftDeletedRegistration, size); err != nil {
					return nil, err
				}
			}
			if entry.recordedSize != nil && !entry.recordedConflict && *entry.recordedSize != size {
				report.RegisteredSizeMismatches++
			}
		}
		if err := addMediaPhysicalCapacityMetric(metric, size); err != nil {
			return nil, err
		}
	}
	for objectID, entry := range evidence {
		if (entry.original || entry.derivative) && entry.recordedSize == nil {
			report.RegisteredUnknownRecordedSizeObjects++
		}
		if entry.recordedConflict {
			report.ConflictingRecordedSizes++
		}
		if entry.original && entry.derivative {
			issues["object_has_original_and_derivative_registration"] = true
		}
		if _, exists := observed[objectID]; exists {
			continue
		}
		if entry.original || entry.derivative {
			if !completeLocations[entry.address.LocationID] {
				report.RegisteredAbsenceUnverified++
				continue
			}
			if entry.recordedSize != nil && !entry.recordedConflict {
				if err := addMediaPhysicalCapacityMetric(&report.RegisteredNotObserved, *entry.recordedSize); err != nil {
					return nil, err
				}
			} else {
				report.RegisteredAbsenceUnverified++
			}
		} else if entry.upload {
			if entry.reservation == nil || entry.reservationConflict {
				report.UnknownUploadReservationObjects++
			} else {
				if err := addMediaPhysicalCapacityMetric(&report.UnobservedUploadReservations, *entry.reservation); err != nil {
					return nil, err
				}
			}
		}
	}
	if ctx.Err() != nil {
		report.PhysicalScanComplete = false
		issues["audit_cancelled_or_timed_out"] = true
	}
	if report.UnresolvedRecordLocations > 0 {
		issues["some_record_locations_unverified"] = true
	}
	if report.UnknownUploadReservationObjects > 0 {
		issues["some_upload_reservation_sizes_unknown"] = true
	}
	for issue := range issues {
		report.Issues = append(report.Issues, issue)
	}
	sort.Strings(report.Issues)
	report.CompletedAt = time.Now().UTC()
	return report, nil
}
