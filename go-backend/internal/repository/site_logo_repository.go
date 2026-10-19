package repository

import (
	"context"

	sitelogodomain "commerce-platform/internal/domain/site_logo"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SiteLogoRepository struct {
	db                                   *gorm.DB
	validateObjectBeforeReplacingCurrent func(context.Context, *sitelogodomain.Asset) error
}

func NewSiteLogoRepository(db *gorm.DB) *SiteLogoRepository {
	return &SiteLogoRepository{db: db}
}

func (r *SiteLogoRepository) WithTx(tx *gorm.DB) *SiteLogoRepository {
	copy := *r
	copy.db = tx
	return &copy
}

func (r *SiteLogoRepository) WithContext(ctx context.Context) *SiteLogoRepository {
	copy := *r
	copy.db = r.db.WithContext(ctx)
	return &copy
}

func (r *SiteLogoRepository) ConfigureTransactionalSiteLogoObjectValidation(validate func(context.Context, *sitelogodomain.Asset) error) {
	r.validateObjectBeforeReplacingCurrent = validate
}

func (r *SiteLogoRepository) Current() (*sitelogodomain.Asset, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}

	var asset sitelogodomain.Asset
	err := r.db.Where("id = ?", sitelogodomain.CurrentAssetID).First(&asset).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &asset, nil
}

func (r *SiteLogoRepository) ReplaceCurrent(
	asset *sitelogodomain.Asset,
	afterReplace func(tx *gorm.DB, previous *sitelogodomain.Asset) error,
) (*sitelogodomain.Asset, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if asset == nil {
		return nil, gorm.ErrInvalidData
	}

	asset.ID = sitelogodomain.CurrentAssetID
	var previous *sitelogodomain.Asset
	err := runWithDedicatedPublicObjectNamespaceTransactionLock(r.db.Statement.Context, r.db, dedicatedSiteLogoObjectNamespace, func(tx *gorm.DB) error {
		var current sitelogodomain.Asset
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", sitelogodomain.CurrentAssetID).
			First(&current).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if err == nil {
			previousCopy := current
			previous = &previousCopy
		}

		if r.validateObjectBeforeReplacingCurrent == nil {
			return ErrDedicatedSiteIdentityTransactionalObjectValidationUnavailable
		}
		if err := r.validateObjectBeforeReplacingCurrent(tx.Statement.Context, asset); err != nil {
			return err
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			UpdateAll: true,
		}).Create(asset).Error; err != nil {
			return err
		}
		if afterReplace != nil {
			return afterReplace(tx, previous)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return previous, nil
}

func (r *SiteLogoRepository) DeleteCurrent(
	afterDelete func(tx *gorm.DB, previous *sitelogodomain.Asset) error,
) (*sitelogodomain.Asset, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}

	var previous *sitelogodomain.Asset
	err := runWithDedicatedPublicObjectNamespaceTransactionLock(r.db.Statement.Context, r.db, dedicatedSiteLogoObjectNamespace, func(tx *gorm.DB) error {
		var current sitelogodomain.Asset
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", sitelogodomain.CurrentAssetID).
			First(&current).Error
		if err == gorm.ErrRecordNotFound {
			if afterDelete != nil {
				return afterDelete(tx, nil)
			}
			return nil
		}
		if err != nil {
			return err
		}

		previousCopy := current
		previous = &previousCopy
		if err := tx.Delete(&current).Error; err != nil {
			return err
		}
		if afterDelete != nil {
			return afterDelete(tx, previous)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return previous, nil
}

func (r *SiteLogoRepository) WithLockedCurrentSiteLogoForObjectCleanup(ctx context.Context, operation func(*sitelogodomain.Asset, *MediaRepository) error) error {
	if r == nil {
		return gorm.ErrInvalidDB
	}
	return runDedicatedPublicObjectPhysicalCleanupTransactionHoldingNamespaceLockUntilProviderReturns(ctx, r.db, dedicatedSiteLogoObjectNamespace, func(tx *gorm.DB) error {
		current, err := NewSiteLogoRepository(tx).Current()
		if err != nil {
			return err
		}
		return operation(current, NewMediaRepository(tx))
	})
}
