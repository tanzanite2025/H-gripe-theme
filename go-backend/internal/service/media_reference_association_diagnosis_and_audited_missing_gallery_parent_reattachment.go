package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/gallery"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"gorm.io/gorm"
)

var ErrMediaAssociationRepairInvalidInput = errors.New("repair requires positive IDs, a current preview and a reason of 3 to 1000 characters")

type MediaAssociationDiagnosisEntry struct {
	MediaUsageSummaryReference
	Diagnosis string `json:"diagnosis"`
}

type MediaAssociationDiagnosisReport struct {
	AssetID                 string                           `json:"asset_id"`
	Coverage                string                           `json:"coverage"`
	CheckedAt               time.Time                        `json:"checked_at"`
	HasRestrictedReferences bool                             `json:"has_restricted_references"`
	Entries                 []MediaAssociationDiagnosisEntry `json:"entries"`
	EntryLimit              int                              `json:"entry_limit"`
	EntriesLimited          bool                             `json:"entries_limited"`
}

type MissingGalleryParentRepairPreview struct {
	AssetID                string `json:"asset_id"`
	ImageID                string `json:"image_id"`
	MissingGalleryID       string `json:"missing_gallery_id"`
	DestinationGalleryID   string `json:"destination_gallery_id"`
	DestinationGalleryName string `json:"destination_gallery_name"`
	ImageTitle             string `json:"image_title"`
	Fingerprint            string `json:"fingerprint"`
}

func (s *MediaService) DiagnoseMediaReferenceAssociationsWithPermissionScopedPresentation(ctx context.Context, assetID uint, scope MediaUsageSummaryAccessScope) (MediaAssociationDiagnosisReport, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	report := MediaAssociationDiagnosisReport{AssetID: strconv.FormatUint(uint64(assetID), 10), Coverage: "recognized_business_sources_only", EntryLimit: 200, Entries: []MediaAssociationDiagnosisEntry{}}
	err := s.repo.ReadMediaAssociationDiagnosisSnapshot(ctx, func(repo *repository.MediaRepository, tx *gorm.DB) error {
		asset, err := repo.FindAssetByID(assetID)
		if err != nil {
			return err
		}
		references, err := repo.FindAssetReferences(asset)
		if err != nil {
			return err
		}
		if len(references) > 20000 {
			return repository.ErrMediaUsageSnapshotLimitExceeded
		}
		for _, reference := range references {
			if !scope.AllowedReferenceTypes[reference.ResourceType] {
				report.HasRestrictedReferences = true
				continue
			}
			diagnosis := ""
			switch {
			case reference.Category == media.ReferenceCategoryRegistration:
				diagnosis = "registration_conflict_requires_ownership_evidence"
				reference.ResourceID, reference.ParentResourceID, reference.Label = 0, 0, "同一对象有其他登记，继续保留并核查所有权"
			case reference.UsageState == "missing_parent":
				diagnosis = "missing_parent_requires_business_review"
				if reference.ResourceType == "gallery_image" {
					var image gallery.GalleryImage
					if err := tx.Unscoped().Select("id, deleted_at").First(&image, reference.ResourceID).Error; err != nil {
						return err
					}
					if image.DeletedAt.Valid {
						diagnosis = "deleted_history_must_remain_retained"
						reference.UsageState = "retained_deleted_placement"
					}
				}
			case reference.UsageState == "retained_deleted_parent" || reference.UsageState == "retained_deleted_placement":
				diagnosis = "deleted_history_must_remain_retained"
			}
			if diagnosis == "" {
				continue
			}
			if len(report.Entries) == report.EntryLimit {
				report.EntriesLimited = true
				continue
			}
			report.Entries = append(report.Entries, MediaAssociationDiagnosisEntry{MediaUsageSummaryReference: MediaUsageSummaryReference{
				Category: reference.Category, ResourceType: reference.ResourceType, ResourceID: strconv.FormatUint(uint64(reference.ResourceID), 10),
				ParentResourceID: strconv.FormatUint(uint64(reference.ParentResourceID), 10), Label: reference.Label, Field: reference.Field, UsageState: reference.UsageState}, Diagnosis: diagnosis})
		}
		return ctx.Err()
	})
	if err != nil {
		return MediaAssociationDiagnosisReport{}, err
	}
	report.CheckedAt = time.Now().UTC()
	return report, nil
}

func createMissingGalleryParentRepairPreview(asset *media.MediaAsset, image *gallery.GalleryImage, destination *gallery.Gallery) (MissingGalleryParentRepairPreview, error) {
	encoded, err := json.Marshal(struct {
		Asset       *media.MediaAsset
		Image       *gallery.GalleryImage
		Destination *gallery.Gallery
	}{asset, image, destination})
	if err != nil {
		return MissingGalleryParentRepairPreview{}, err
	}
	digest := sha256.Sum256(encoded)
	return MissingGalleryParentRepairPreview{AssetID: strconv.FormatUint(uint64(asset.ID), 10), ImageID: strconv.FormatUint(uint64(image.ID), 10),
		MissingGalleryID: strconv.FormatUint(uint64(image.GalleryID), 10), DestinationGalleryID: strconv.FormatUint(uint64(destination.ID), 10),
		DestinationGalleryName: destination.Name, ImageTitle: image.Title, Fingerprint: hex.EncodeToString(digest[:])}, nil
}

func (s *MediaService) validateMissingGalleryParentRepairReferences(repo *repository.MediaRepository, asset *media.MediaAsset, scope MediaUsageSummaryAccessScope) error {
	conversationAccess, err := repo.HistoricalUnclassifiedMediaRequiresConversationAccess(asset)
	if err != nil {
		return err
	}
	if conversationAccess {
		return repository.ErrMediaAssociationRepairConflict
	}
	key, err := s.ResolveTrustedMediaObjectKeyForPublicBusinessReference(asset.URL)
	if err != nil || key != asset.StorageKey || storage.IsPrivateObjectKey(key) || storage.DedicatedPublicObjectNamespaceForStorageKey(key) != "" {
		return repository.ErrMediaAssociationRepairConflict
	}
	refs, err := repo.FindAssetReferences(asset)
	if err != nil {
		return err
	}
	if len(refs) > 20000 {
		return repository.ErrMediaUsageSnapshotLimitExceeded
	}
	for _, reference := range refs {
		if !scope.AllowedReferenceTypes[reference.ResourceType] || reference.Category == media.ReferenceCategoryRegistration {
			return repository.ErrMediaAssociationRepairConflict
		}
	}
	// An unfinished draft cannot become permanent through this repair.
	return repo.ValidateMediaAssetHasNoUnboundBusinessDraftUploadReceipt(asset.ID)
}

func (s *MediaService) PreviewMissingGalleryParentReattachment(ctx context.Context, assetID, imageID, destinationID uint, scope MediaUsageSummaryAccessScope) (MissingGalleryParentRepairPreview, error) {
	if assetID == 0 || imageID == 0 || destinationID == 0 || !scope.AllowedReferenceTypes["gallery_image"] {
		return MissingGalleryParentRepairPreview{}, ErrMediaAssociationRepairInvalidInput
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var preview MissingGalleryParentRepairPreview
	err := s.repo.ReadMediaAssociationDiagnosisSnapshot(ctx, func(repo *repository.MediaRepository, tx *gorm.DB) error {
		asset, image, destination, err := repo.LoadMissingGalleryParentRepairRecords(assetID, imageID, destinationID, false)
		if err != nil {
			return err
		}
		if err = s.validateMissingGalleryParentRepairReferences(repo, asset, scope); err != nil {
			return err
		}
		preview, err = createMissingGalleryParentRepairPreview(asset, image, destination)
		return err
	})
	if err != nil {
		return MissingGalleryParentRepairPreview{}, err
	}
	return preview, nil
}

func (s *MediaService) ReattachMissingGalleryParentWithFreshPreviewAndRequiredAudit(ctx context.Context, assetID, imageID, destinationID uint, expectedFingerprint, reason string, scope MediaUsageSummaryAccessScope, auditFactory func() *audit.AuditLog) (MissingGalleryParentRepairPreview, error) {
	reason = strings.TrimSpace(reason)
	if assetID == 0 || imageID == 0 || destinationID == 0 || !scope.AllowedReferenceTypes["gallery_image"] || len(expectedFingerprint) != 64 || utf8.RuneCountInString(reason) < 3 || utf8.RuneCountInString(reason) > 1000 {
		return MissingGalleryParentRepairPreview{}, ErrMediaAssociationRepairInvalidInput
	}
	if _, err := hex.DecodeString(expectedFingerprint); err != nil || auditFactory == nil {
		return MissingGalleryParentRepairPreview{}, ErrMediaAssociationRepairInvalidInput
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	var result MissingGalleryParentRepairPreview
	err := s.repo.WithContextForMediaLibraryClassification(ctx).WithTransaction(func(repo *repository.MediaRepository, tx *gorm.DB) error {
		if err := repo.LockMissingGalleryParentRepairTables(); err != nil {
			return err
		}
		asset, image, destination, err := repo.LoadMissingGalleryParentRepairRecords(assetID, imageID, destinationID, true)
		if err != nil {
			return err
		}
		before, err := createMissingGalleryParentRepairPreview(asset, image, destination)
		if err != nil {
			return err
		}
		if before.Fingerprint != expectedFingerprint {
			return repository.ErrMediaAssociationRepairConflict
		}
		if err := s.validateMissingGalleryParentRepairReferences(repo, asset, scope); err != nil {
			return err
		}
		key, err := s.ResolveTrustedMediaObjectKeyForPublicBusinessReference(asset.URL)
		if err != nil || key != asset.StorageKey || storage.IsPrivateObjectKey(key) || storage.DedicatedPublicObjectNamespaceForStorageKey(key) != "" {
			return repository.ErrMediaAssociationRepairConflict
		}
		if err := s.ValidateMediaObjectExistsBeforeRegisteringOwnership(ctx, key); err != nil {
			return err
		}
		log := auditFactory()
		if log == nil || log.UserID == 0 {
			return ErrMediaLibraryClassificationAuditRequired
		}
		log.Action, log.Resource, log.ResourceID, log.Status = "repair_missing_parent", "media_gallery_image_association", image.ID, "success"
		log.OldValue = serviceAuditJSON(image)
		if err := repo.ReattachMissingGalleryParentWithoutChangingImageMetadata(image, destination); err != nil {
			return err
		}
		image.GalleryID = destination.ID
		log.NewValue = serviceAuditJSON(image)
		log.Changes = serviceAuditJSON(map[string]interface{}{"reason": reason, "review": before})
		if err := repository.NewAuditRepository(tx).CreateAuditLog(log); err != nil {
			return err
		}
		result = before
		return ctx.Err()
	})
	if err != nil {
		return MissingGalleryParentRepairPreview{}, err
	}
	return result, nil
}
