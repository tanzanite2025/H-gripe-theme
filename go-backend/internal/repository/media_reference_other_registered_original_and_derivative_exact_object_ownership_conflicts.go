package repository

import (
	"fmt"
	"strings"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
)

// Cleanup also owns private objects, so its identity resolver must not use the
// public-business eligibility rules to discard a private registration.
type registeredMediaObjectCleanupIdentityResolver interface {
	ResolveRegisteredMediaObjectKeyForCleanupOwnership(reference string) (string, error)
}

func (r *MediaRepository) otherRegisteredMediaOriginalAndDerivativeObjectOwnershipReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	if query.AssetID == 0 && !query.IncludeAllExactObjectRegistrations {
		return nil, nil
	}
	for _, reference := range query.URLs {
		if key := r.resolveRegisteredMediaObjectKeyForCleanupOwnership(reference); key != "" {
			query.ObjectKeys = append(query.ObjectKeys, key)
		}
	}
	var references []media.AssetReference
	for _, source := range []struct {
		table, ownerColumn, resourceType, role string
	}{
		{"media_assets", "id", "media_asset_registration", "原图"},
		{"media_asset_derivatives", "media_asset_id", "media_derivative_registration", "衍生图"},
	} {
		if !r.hasTable(source.table) {
			continue
		}
		for afterID := uint(0); ; {
			var rows []struct {
				ID, OwnerID     uint
				URL, StorageKey string
				RetainedDeleted bool
			}
			// Retained registrations still own bytes, even when their parent or
			// derivative is deleted. Never choose one duplicate as the winner.
			scan := r.db.Table(source.table).
				Select("id, "+source.ownerColumn+" AS owner_id, url, storage_key, (deleted_at IS NOT NULL) AS retained_deleted").
				Where("id > ?", afterID)
			if !query.IncludeAllExactObjectRegistrations {
				scan = scan.Where(source.ownerColumn+" <> ?", query.AssetID)
			} else if source.table == "media_asset_derivatives" && query.ExcludedRetiredDerivativeID != 0 {
				scan = scan.Where("id <> ?", query.ExcludedRetiredDerivativeID)
			}
			if err := scan.Order("id ASC").Limit(256).Find(&rows).Error; err != nil {
				return nil, err
			}
			for _, row := range rows {
				matches := registeredMediaReferenceQueryContainsExactObjectKey(query, row.StorageKey) ||
					r.matchesExactRegisteredOrTrustedMediaReference(query, row.URL)
				for _, value := range []string{row.StorageKey, row.URL} {
					key := r.resolveRegisteredMediaObjectKeyForCleanupOwnership(value)
					matches = matches || (key != "" && (registeredMediaReferenceQueryContainsExactObjectKey(query, key) ||
						(query.RegisteredObjectCleanupNamespacePrefix != "" && strings.HasPrefix(key, query.RegisteredObjectCleanupNamespacePrefix))))
				}
				if !matches {
					continue
				}
				reference := newMediaReference(media.ReferenceCategoryRegistration, source.resourceType, row.ID, row.OwnerID,
					fmt.Sprintf("媒体 #%d 的%s登记占用同一存储对象，需核查登记", row.OwnerID, source.role), "storage_key, url")
				if row.RetainedDeleted {
					reference.UsageState = "retained_deleted_placement"
				}
				references = append(references, reference)
			}
			if len(rows) < 256 {
				break
			}
			afterID = rows[len(rows)-1].ID
		}
	}
	return references, nil
}

func (r *MediaRepository) resolveRegisteredMediaObjectKeyForCleanupOwnership(reference string) string {
	if resolver, ok := r.businessMediaReferenceResolver.(registeredMediaObjectCleanupIdentityResolver); ok {
		key, err := resolver.ResolveRegisteredMediaObjectKeyForCleanupOwnership(reference)
		if err == nil {
			return key
		}
		return ""
	}
	key, err := storage.TrustedBusinessMediaReferenceObjectKey(reference, nil, "")
	if err == nil && r.businessMediaReferenceResolver != nil {
		key, err = r.businessMediaReferenceResolver.ResolveTrustedMediaObjectKeyForPublicBusinessReference(reference)
	}
	if err != nil {
		return ""
	}
	return key
}
