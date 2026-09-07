package repository

import (
	productdomain "commerce-platform/internal/domain/product"
	productrequirement "commerce-platform/internal/domain/productrequirement"
	"strings"

	"gorm.io/gorm"
)

type ProductQualityRequirementRepository struct {
	db *gorm.DB
}

func NewProductQualityRequirementRepository(db *gorm.DB) *ProductQualityRequirementRepository {
	return &ProductQualityRequirementRepository{db: db}
}

func (r *ProductQualityRequirementRepository) WithTx(tx *gorm.DB) *ProductQualityRequirementRepository {
	return &ProductQualityRequirementRepository{db: tx}
}

func (r *ProductQualityRequirementRepository) FindByID(id uint) (*productrequirement.ProductQualityRequirementRule, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	var record productrequirement.ProductQualityRequirementRule
	if err := r.db.First(&record, id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *ProductQualityRequirementRepository) FindByScope(
	productID uint,
	variantID *uint,
	requirementType string,
) (*productrequirement.ProductQualityRequirementRule, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	query := r.db.Where(
		"product_id = ? AND requirement_type = ?",
		productID,
		strings.TrimSpace(requirementType),
	)
	if variantID == nil {
		query = query.Where("variant_id IS NULL")
	} else {
		query = query.Where("variant_id = ?", *variantID)
	}

	var record productrequirement.ProductQualityRequirementRule
	if err := query.First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *ProductQualityRequirementRepository) ListByProduct(productID uint) ([]productrequirement.ProductQualityRequirementRule, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	var records []productrequirement.ProductQualityRequirementRule
	err := r.db.
		Where("product_id = ?", productID).
		Order("CASE WHEN variant_id IS NULL THEN 0 ELSE 1 END ASC, variant_id ASC, id ASC").
		Find(&records).Error
	return records, err
}

func (r *ProductQualityRequirementRepository) Create(record *productrequirement.ProductQualityRequirementRule) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	return r.db.Create(record).Error
}

func (r *ProductQualityRequirementRepository) Update(record *productrequirement.ProductQualityRequirementRule) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	return r.db.Save(record).Error
}

func (r *ProductQualityRequirementRepository) EnsureProductScope(productID uint, variantID *uint) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if productID == 0 {
		return gorm.ErrRecordNotFound
	}

	var productCount int64
	if err := r.db.Model(&productdomain.Product{}).
		Where("id = ?", productID).
		Count(&productCount).Error; err != nil {
		return err
	}
	if productCount == 0 {
		return gorm.ErrRecordNotFound
	}

	if variantID == nil {
		return nil
	}

	var variantCount int64
	if err := r.db.Model(&productdomain.ProductVariant{}).
		Where("id = ? AND product_id = ?", *variantID, productID).
		Count(&variantCount).Error; err != nil {
		return err
	}
	if variantCount == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
