package repository

import (
	"context"

	sitefavicondomain "commerce-platform/internal/domain/site_favicon"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SiteFaviconRepository struct {
	db                                   *gorm.DB
	validateObjectBeforeReplacingCurrent func(context.Context, *sitefavicondomain.Asset) error
}

func NewSiteFaviconRepository(db *gorm.DB) *SiteFaviconRepository {
	return &SiteFaviconRepository{db: db}
}

func (r *SiteFaviconRepository) WithTx(tx *gorm.DB) *SiteFaviconRepository {
	copy := *r
	copy.db = tx
	return &copy
}

func (r *SiteFaviconRepository) WithContext(ctx context.Context) *SiteFaviconRepository {
	copy := *r
	copy.db = r.db.WithContext(ctx)
	return &copy
}

func (r *SiteFaviconRepository) ConfigureTransactionalSiteFaviconObjectValidation(validate func(context.Context, *sitefavicondomain.Asset) error) {
	r.validateObjectBeforeReplacingCurrent = validate
}

func (r *SiteFaviconRepository) Current() (*sitefavicondomain.Asset, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	var asset sitefavicondomain.Asset
	err := r.db.Where("id = ?", sitefavicondomain.CurrentAssetID).First(&asset).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *SiteFaviconRepository) ReplaceCurrent(asset *sitefavicondomain.Asset, afterReplace func(*gorm.DB, *sitefavicondomain.Asset) error) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if asset == nil {
		return gorm.ErrInvalidData
	}
	asset.ID = sitefavicondomain.CurrentAssetID
	return runWithDedicatedPublicObjectNamespaceTransactionLock(r.db.Statement.Context, r.db, dedicatedSiteFaviconObjectNamespace, func(tx *gorm.DB) error {
		var current sitefavicondomain.Asset
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", asset.ID).First(&current).Error
		var previous *sitefavicondomain.Asset
		if err == nil {
			copy := current
			previous = &copy
		} else if err != gorm.ErrRecordNotFound {
			return err
		}
		if r.validateObjectBeforeReplacingCurrent == nil {
			return ErrDedicatedSiteIdentityTransactionalObjectValidationUnavailable
		}
		if err := r.validateObjectBeforeReplacingCurrent(tx.Statement.Context, asset); err != nil {
			return err
		}
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, UpdateAll: true}).Create(asset).Error; err != nil {
			return err
		}
		if afterReplace != nil {
			return afterReplace(tx, previous)
		}
		return nil
	})
}

func (r *SiteFaviconRepository) DeleteCurrent(afterDelete func(*gorm.DB, *sitefavicondomain.Asset) error) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	return runWithDedicatedPublicObjectNamespaceTransactionLock(r.db.Statement.Context, r.db, dedicatedSiteFaviconObjectNamespace, func(tx *gorm.DB) error {
		var current sitefavicondomain.Asset
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", sitefavicondomain.CurrentAssetID).First(&current).Error
		if err == gorm.ErrRecordNotFound {
			if afterDelete != nil {
				return afterDelete(tx, nil)
			}
			return nil
		}
		if err != nil {
			return err
		}
		if err := tx.Delete(&current).Error; err != nil {
			return err
		}
		if afterDelete != nil {
			return afterDelete(tx, &current)
		}
		return nil
	})
}

func (r *SiteFaviconRepository) WithLockedCurrentSiteFaviconForObjectCleanup(ctx context.Context, operation func(*sitefavicondomain.Asset, *MediaRepository) error) error {
	if r == nil {
		return gorm.ErrInvalidDB
	}
	return runDedicatedPublicObjectPhysicalCleanupTransactionHoldingNamespaceLockUntilProviderReturns(ctx, r.db, dedicatedSiteFaviconObjectNamespace, func(tx *gorm.DB) error {
		current, err := NewSiteFaviconRepository(tx).Current()
		if err != nil {
			return err
		}
		return operation(current, NewMediaRepository(tx))
	})
}
