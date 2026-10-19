package repository

import (
	"context"
	"fmt"

	"commerce-platform/internal/pkg/storage"
	"gorm.io/gorm"
)

// Use the dedicated owner's database and URL resolver. The shared library may
// contain the same key in a different physical store and cannot grant access.
func findPublishedBusinessConsumerForDedicatedPublicObject(ctx context.Context, db *gorm.DB, namespace, key string, resolver PublicBusinessMediaReferenceObjectKeyResolver) (bool, error) {
	if db == nil {
		return false, gorm.ErrInvalidDB
	}
	if storage.DedicatedPublicObjectNamespaceForStorageKey(key) != namespace {
		return false, fmt.Errorf("dedicated object does not belong to %s", namespace)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	repo := NewMediaRepository(db.WithContext(ctx))
	repo.ConfigurePublicBusinessMediaReferenceIdentityResolver(resolver)
	// Private attachment JSON and conflicting registration metadata retain bytes
	// but cannot grant access. Public requests need not inspect those domains.
	state, err := repo.findDedicatedPublicObjectBusinessReferenceStateForRetentionOrPublicAccess(key, true)
	return state.HasPublishedConsumer && err == nil, err
}

func (r *SiteLogoRepository) HasPublishedBusinessConsumerForDedicatedPublicObject(ctx context.Context, key string, resolver PublicBusinessMediaReferenceObjectKeyResolver) (bool, error) {
	return findPublishedBusinessConsumerForDedicatedPublicObject(ctx, r.db, dedicatedSiteLogoObjectNamespace, key, resolver)
}

func (r *SiteFaviconRepository) HasPublishedBusinessConsumerForDedicatedPublicObject(ctx context.Context, key string, resolver PublicBusinessMediaReferenceObjectKeyResolver) (bool, error) {
	return findPublishedBusinessConsumerForDedicatedPublicObject(ctx, r.db, dedicatedSiteFaviconObjectNamespace, key, resolver)
}

func (r *HomeVisualTileRepository) HasPublishedBusinessConsumerForDedicatedPublicObject(ctx context.Context, key string, resolver PublicBusinessMediaReferenceObjectKeyResolver) (bool, error) {
	return findPublishedBusinessConsumerForDedicatedPublicObject(ctx, r.db, dedicatedHomeVisualObjectNamespace, key, resolver)
}
