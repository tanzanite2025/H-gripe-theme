package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"gorm.io/gorm"

	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/pkg/safehtml"
	"commerce-platform/internal/pkg/storage"
)

var ErrImmutableOrderEvidenceSnapshotMediaReferenceUnavailable = errors.New("immutable order evidence snapshot media reference is unavailable or outside its owner scope")

type immutableOrderEvidenceSnapshotMediaAdmissionScope struct {
	orderID                uint
	provider               string
	disputeID              uint
	bindNewInvoiceReceipts func(*gorm.DB, []string) error
}

// Snapshot insertion and physical retirement share the existing owner locks.
// The immutable body and its original digest are never rewritten here.
func admitImmutableOrderEvidenceSnapshotMediaReferencesWithinTransaction(tx *gorm.DB, resolver PublicBusinessMediaReferenceObjectKeyResolver, data []byte, scope immutableOrderEvidenceSnapshotMediaAdmissionScope) error {
	references, err := collectImmutableSnapshotExactAndRichTextMediaReferences(data)
	if err != nil {
		return err
	}
	repo := NewMediaRepository(tx)
	repo.ConfigurePublicBusinessMediaReferenceIdentityResolver(resolver)
	keys := make(map[string]string, len(references))
	namespaces := make(map[string]bool)
	var publicReferences []string
	for _, reference := range references {
		key, resolveErr := storage.TrustedBusinessMediaReferenceObjectKey(reference, nil, "")
		if resolver != nil {
			if privateResolver, ok := resolver.(interface{ ResolveRegisteredMediaObjectKeyForCleanupOwnership(string) (string, error) }); ok {
				key, resolveErr = privateResolver.ResolveRegisteredMediaObjectKeyForCleanupOwnership(reference)
			} else {
				key, resolveErr = resolver.ResolveTrustedMediaObjectKeyForPublicBusinessReference(reference)
			}
		}
		if normalized, valid := storage.NormalizeObjectKey(reference); valid && storage.DedicatedPublicObjectNamespaceForStorageKey(normalized) != "" {
			key, resolveErr = normalized, nil
		}
		if resolveErr != nil {
			return ErrImmutableOrderEvidenceSnapshotMediaReferenceUnavailable
		}
		if key == "" && repo.hasTable("media_assets") {
			identity, err := repo.resolveRegisteredBusinessMediaObjectIdentity(reference, "", true, repo.hasAssetDerivativesTable())
			if err != nil {
				return err
			}
			if identity.AssetID != 0 {
				var row struct{ StorageKey string }
				if identity.DerivativeID != 0 {
					if err := tx.Table("media_asset_derivatives").Select("storage_key").Where("id = ?", identity.DerivativeID).Take(&row).Error; err != nil {
						return err
					}
				} else if err := tx.Table("media_assets").Select("storage_key").Where("id = ?", identity.AssetID).Take(&row).Error; err != nil {
					return err
				}
				if storage.DedicatedPublicObjectNamespaceForStorageKey(row.StorageKey) != "" {
					key = row.StorageKey
				}
			}
		}
		keys[reference] = key
		if namespace := immutableSnapshotLifecycleNamespaceForObjectKey(key); namespace != "" {
			namespaces[namespace] = true
		} else if key != "" || repo.hasTable("media_assets") {
			publicReferences = append(publicReferences, reference)
		}
	}
	// The PayPal upload adapter already holds this namespace. Acquire it first
	// everywhere, before any evidence namespace or sorted shared asset rows.
	orderedNamespaces := make([]string, 0, len(namespaces))
	if namespaces[suggestionAndPayPalAttachmentLifecycleNamespace] {
		orderedNamespaces = append(orderedNamespaces, suggestionAndPayPalAttachmentLifecycleNamespace)
		delete(namespaces, suggestionAndPayPalAttachmentLifecycleNamespace)
	}
	remainingNamespaces := make([]string, 0, len(namespaces))
	for namespace := range namespaces {
		remainingNamespaces = append(remainingNamespaces, namespace)
	}
	sort.Strings(remainingNamespaces)
	orderedNamespaces = append(orderedNamespaces, remainingNamespaces...)
	for _, namespace := range orderedNamespaces {
		if err := acquireDedicatedPublicObjectNamespaceLockWithinTransaction(tx, namespace); err != nil {
			return err
		}
	}
	var newInvoiceKeys []string
	for _, reference := range references {
		key := keys[reference]
		if immutableSnapshotLifecycleNamespaceForObjectKey(key) == "" {
			continue
		}
		if namespace, ownerID, _, valid := storage.SuggestionOrPayPalAttachmentObjectIdentity(key); valid && namespace == storage.PayPalDisputeInvoiceAttachmentNamespace && scope.provider != "" && (scope.provider != "paypal" || ownerID != scope.disputeID) {
			return ErrImmutableOrderEvidenceSnapshotMediaReferenceUnavailable
		}
		inherited, err := immutableSnapshotHasRetainedSameOrderReference(tx, resolver, scope, reference, key)
		if err != nil {
			return err
		}
		if inherited {
			validator, ok := resolver.(registeredMediaObjectExistenceValidator)
			if !ok {
				return ErrMediaObjectRegistrationVerificationUnavailable
			}
			if err := validator.ValidateMediaObjectExistsBeforeRegisteringOwnership(tx.Statement.Context, key); err != nil {
				return errors.Join(ErrImmutableOrderEvidenceSnapshotMediaReferenceUnavailable, err)
			}
			continue
		}
		namespace, ownerID, _, valid := storage.SuggestionOrPayPalAttachmentObjectIdentity(key)
		if valid && namespace == storage.PayPalDisputeInvoiceAttachmentNamespace && scope.provider == "paypal" && ownerID == scope.disputeID && scope.bindNewInvoiceReceipts != nil {
			newInvoiceKeys = append(newInvoiceKeys, key)
			continue
		}
		return ErrImmutableOrderEvidenceSnapshotMediaReferenceUnavailable
	}
	resolved, err := lockAndResolvePublicBusinessMediaReferenceIdentities(tx, resolver, publicReferences, nil)
	if err != nil {
		return errors.Join(ErrImmutableOrderEvidenceSnapshotMediaReferenceUnavailable, err)
	}
	for _, reference := range publicReferences {
		identity := resolved.identities[reference]
		if keys[reference] != "" && identity.AssetID == 0 {
			return ErrImmutableOrderEvidenceSnapshotMediaReferenceUnavailable
		}
		if identity.AssetID == 0 {
			continue
		} // Unmanaged external links are facts, not local ownership.
		asset := resolved.assets[identity.AssetID]
		if asset == nil {
			return ErrImmutableOrderEvidenceSnapshotMediaReferenceUnavailable
		}
		if err := repo.ValidateBusinessDraftMediaDerivativeGenerationIsPermanentlyBound(asset); err != nil {
			return err
		}
		// Shared registrations can never authorize another lifecycle's bytes.
		if storage.DedicatedPublicObjectNamespaceForStorageKey(asset.StorageKey) != "" {
			return ErrImmutableOrderEvidenceSnapshotMediaReferenceUnavailable
		}
		objectKey := keys[reference]
		if objectKey == "" {
			objectKey = asset.StorageKey
		}
		if identity.DerivativeID != 0 && keys[reference] == "" {
			var derivative struct{ StorageKey string }
			if err := tx.Table("media_asset_derivatives").Select("storage_key").Where("id = ?", identity.DerivativeID).Take(&derivative).Error; err != nil {
				return err
			}
			objectKey = derivative.StorageKey
		}
		validator, ok := resolver.(registeredMediaObjectExistenceValidator)
		if !ok || objectKey == "" {
			return ErrMediaObjectRegistrationVerificationUnavailable
		}
		if err := validator.ValidateMediaObjectExistsBeforeRegisteringOwnership(tx.Statement.Context, objectKey); err != nil {
			return errors.Join(ErrImmutableOrderEvidenceSnapshotMediaReferenceUnavailable, err)
		}
	}
	if len(newInvoiceKeys) > 0 {
		sort.Strings(newInvoiceKeys)
		unique := newInvoiceKeys[:0]
		for _, key := range newInvoiceKeys {
			if len(unique) == 0 || unique[len(unique)-1] != key {
				unique = append(unique, key)
			}
		}
		if err := scope.bindNewInvoiceReceipts(tx, unique); err != nil {
			return err
		}
	}
	return nil
}

func lockImmutableSnapshotSubmissionSelectionBeforeReadingOrBindingReceipts(tx *gorm.DB, provider string, disputeID uint) error {
	if strings.EqualFold(strings.TrimSpace(provider), "paypal") {
		if err := acquireDedicatedPublicObjectNamespaceLockWithinTransaction(tx, suggestionAndPayPalAttachmentLifecycleNamespace); err != nil {
			return err
		}
	}
	return acquireDedicatedPublicObjectNamespaceLockWithinTransaction(tx, fmt.Sprintf("immutable-evidence-submission-selection/%s/%d", strings.ToLower(strings.TrimSpace(provider)), disputeID))
}

func immutableSnapshotLifecycleNamespaceForObjectKey(key string) string {
	namespace := storage.DedicatedPublicObjectNamespaceForStorageKey(key)
	switch namespace {
	case storage.SuggestionFeedbackAttachmentNamespace, storage.PayPalDisputeInvoiceAttachmentNamespace:
		return suggestionAndPayPalAttachmentLifecycleNamespace
	case "showcase/pending", "showcase/approved":
		return dedicatedUGCShowcaseLifecycleNamespace
	}
	return namespace
}

func collectImmutableSnapshotExactAndRichTextMediaReferences(data []byte) ([]string, error) {
	var document any
	if err := json.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("decode snapshot media references: invalid JSON")
	}
	unique := make(map[string]bool)
	var collectReferences func(any) error
	collectReferences = func(value any) error {
		switch value := value.(type) {
		case string:
			trimmed := strings.TrimSpace(value)
			if strings.HasPrefix(trimmed, "https://") || strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "/") || strings.HasPrefix(trimmed, "uploads/") || storage.DedicatedPublicObjectNamespaceForStorageKey(trimmed) != "" {
				unique[value] = true
			}
			if strings.Contains(value, "<") {
				parsed, err := safehtml.ParseRichTextMediaSourceAndFileLinkAttributes(value)
				if err != nil {
					return err
				}
				for _, reference := range parsed.References {
					unique[*reference.URL] = true
				}
			}
		case []any:
			for _, child := range value {
				if err := collectReferences(child); err != nil {
					return err
				}
			}
		case map[string]any:
			for _, child := range value {
				if err := collectReferences(child); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := collectReferences(document); err != nil {
		return nil, err
	}
	references := make([]string, 0, len(unique))
	for reference := range unique {
		references = append(references, reference)
	}
	sort.Strings(references)
	return references, nil
}

func immutableSnapshotHasRetainedSameOrderReference(tx *gorm.DB, resolver PublicBusinessMediaReferenceObjectKeyResolver, scope immutableOrderEvidenceSnapshotMediaAdmissionScope, reference, key string) (bool, error) {
	repo := NewMediaRepository(tx)
	repo.ConfigurePublicBusinessMediaReferenceIdentityResolver(resolver)
	query := mediaAssetReferenceQuery{URLs: []string{reference}, ObjectKeys: []string{key}}
	for _, table := range []string{"order_evidence_snapshots", "order_evidence_submission_snapshots", "order_evidence_export_snapshots"} {
		if !repo.hasTable(table) {
			continue
		}
		column := "snapshot_data"
		var rows []struct {
			ID    uint
			Value string
		}
		selected := tx.Table(table).Select("id, "+column+" AS value").Where("order_id = ?", scope.orderID)
		if scope.provider != "" && strings.HasPrefix(key, storage.PayPalDisputeInvoiceAttachmentNamespace+"/") && table == "order_evidence_submission_snapshots" {
			selected = selected.Where("provider = ? AND dispute_id = ?", scope.provider, scope.disputeID)
		}
		if err := selected.Find(&rows).Error; err != nil {
			return false, err
		}
		for _, row := range rows {
			matches, err := repo.decodedLegacyMediaJSONContainsExactReference(query, row.Value, table, column, row.ID)
			if err != nil {
				return false, err
			}
			if matches {
				var integrityErr error
				switch table {
				case "order_evidence_snapshots":
					var stored orderevidence.OrderEvidenceSnapshot
					if err := tx.First(&stored, row.ID).Error; err != nil {
						return false, err
					}
					integrityErr = stored.Validate()
				case "order_evidence_submission_snapshots":
					var stored orderevidence.OrderEvidenceSubmissionSnapshot
					if err := tx.First(&stored, row.ID).Error; err != nil {
						return false, err
					}
					integrityErr = stored.Validate()
				case "order_evidence_export_snapshots":
					var stored orderevidence.OrderEvidenceExportSnapshot
					if err := tx.First(&stored, row.ID).Error; err != nil {
						return false, err
					}
					integrityErr = stored.Validate()
				}
				return integrityErr == nil, integrityErr
			}
		}
	}
	if strings.HasPrefix(key, orderEvidenceAttachmentLifecycleNamespace+"/") && repo.hasTable("order_evidence_attachments") && repo.hasTable("order_evidence_items") {
		var count int64
		err := tx.Table("order_evidence_attachments AS attachment").Joins("JOIN order_evidence_items AS item ON item.id = attachment.evidence_item_id").Where("item.order_id = ? AND attachment.storage_key = ?", scope.orderID, key).Count(&count).Error
		return count > 0, err
	}
	return false, nil
}
