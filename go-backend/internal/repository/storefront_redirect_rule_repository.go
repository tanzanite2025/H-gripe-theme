package repository

import (
	"errors"
	"time"

	urlmanagementdomain "commerce-platform/internal/domain/urlmanagement"

	"gorm.io/gorm"
)

type StorefrontRedirectRuleRepository struct {
	db *gorm.DB
}

func NewStorefrontRedirectRuleRepository(db *gorm.DB) *StorefrontRedirectRuleRepository {
	return &StorefrontRedirectRuleRepository{db: db}
}

func (r *StorefrontRedirectRuleRepository) List() ([]urlmanagementdomain.StorefrontRedirectRule, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("storefront redirect rule repository is unavailable")
	}

	var rules []urlmanagementdomain.StorefrontRedirectRule
	err := r.db.
		Order("state ASC").
		Order("source_path ASC").
		Find(&rules).Error
	return rules, err
}

func (r *StorefrontRedirectRuleRepository) ListPublished() ([]urlmanagementdomain.StorefrontPublishedRedirect, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("storefront redirect rule repository is unavailable")
	}

	var rules []urlmanagementdomain.StorefrontPublishedRedirect
	err := r.db.Model(&urlmanagementdomain.StorefrontRedirectRule{}).
		Select("source_path, target_path, status_code").
		Where("state = ?", urlmanagementdomain.RedirectRuleStatePublished).
		Order("source_path ASC").
		Scan(&rules).Error
	return rules, err
}

func (r *StorefrontRedirectRuleRepository) FindByID(id uint) (*urlmanagementdomain.StorefrontRedirectRule, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("storefront redirect rule repository is unavailable")
	}

	var rule urlmanagementdomain.StorefrontRedirectRule
	if err := r.db.First(&rule, id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *StorefrontRedirectRuleRepository) FindBySourcePath(sourcePath string) (*urlmanagementdomain.StorefrontRedirectRule, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("storefront redirect rule repository is unavailable")
	}

	var rule urlmanagementdomain.StorefrontRedirectRule
	if err := r.db.Where("source_path = ?", sourcePath).First(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *StorefrontRedirectRuleRepository) Create(rule *urlmanagementdomain.StorefrontRedirectRule) error {
	if r == nil || r.db == nil {
		return errors.New("storefront redirect rule repository is unavailable")
	}
	if rule == nil {
		return errors.New("storefront redirect rule is nil")
	}
	return r.db.Create(rule).Error
}

func (r *StorefrontRedirectRuleRepository) Publish(id uint, publishedByID uint, publishedAt time.Time) error {
	if r == nil || r.db == nil {
		return errors.New("storefront redirect rule repository is unavailable")
	}
	if publishedAt.IsZero() {
		publishedAt = time.Now().UTC()
	}

	return r.db.Model(&urlmanagementdomain.StorefrontRedirectRule{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"state":           urlmanagementdomain.RedirectRuleStatePublished,
			"published_by_id": publishedByID,
			"published_at":    publishedAt,
			"disabled_at":     nil,
			"updated_at":      publishedAt,
		}).Error
}

func (r *StorefrontRedirectRuleRepository) Disable(id uint, disabledAt time.Time) error {
	if r == nil || r.db == nil {
		return errors.New("storefront redirect rule repository is unavailable")
	}
	if disabledAt.IsZero() {
		disabledAt = time.Now().UTC()
	}

	return r.db.Model(&urlmanagementdomain.StorefrontRedirectRule{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"state":       urlmanagementdomain.RedirectRuleStateDisabled,
			"disabled_at": disabledAt,
			"updated_at":  disabledAt,
		}).Error
}
