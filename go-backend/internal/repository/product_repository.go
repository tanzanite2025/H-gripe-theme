package repository

import (
	"tanzanite/internal/domain/product"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// WithTx 澶嶇敤浜嬪姟 db 瀹炰緥
func (r *ProductRepository) WithTx(tx *gorm.DB) *ProductRepository {
	return &ProductRepository{db: tx}
}

func orderProductImages(db *gorm.DB) *gorm.DB {
	return db.Order(clause.OrderByColumn{
		Column: clause.Column{Table: "product_images", Name: "order"},
	})
}

func orderSpecDefinitions(db *gorm.DB) *gorm.DB {
	return db.Order("product_spec_definitions.sort_order ASC, product_spec_definitions.id ASC")
}

func orderProductVariants(db *gorm.DB) *gorm.DB {
	return db.Order("product_variants.sort_order ASC, product_variants.id ASC")
}

// Create 鍒涘缓浜у搧
func (r *ProductRepository) Create(p *product.Product) error {
	return r.db.Create(p).Error
}

func (r *ProductRepository) CreateWithSpecValues(p *product.Product, specValues []product.ProductSpecValue) error {
	return r.CreateWithSpecValuesAndVariants(p, specValues, nil)
}

func (r *ProductRepository) CreateWithSpecValuesAndVariants(p *product.Product, specValues []product.ProductSpecValue, variants []product.ProductVariant) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		syncProductSummaryFromVariants(p, variants)
		if err := tx.Create(p).Error; err != nil {
			return err
		}

		if len(specValues) > 0 {
			for i := range specValues {
				specValues[i].ProductID = p.ID
			}
			if err := tx.Create(&specValues).Error; err != nil {
				return err
			}
		}

		if len(variants) > 0 {
			for i := range variants {
				variants[i].ProductID = p.ID
			}
			if err := tx.Create(&variants).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// FindByID 鏍规嵁ID鏌ユ壘浜у搧
func (r *ProductRepository) FindByID(id uint) (*product.Product, error) {
	var p product.Product
	err := r.db.Preload("Images", func(db *gorm.DB) *gorm.DB {
		return orderProductImages(db)
	}).Preload("ProductType.SpecDefinitions", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
	}).Preload("SpecValues.SpecDefinition", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
	}).Preload("Variants", func(db *gorm.DB) *gorm.DB {
		return orderProductVariants(db)
	}).First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// FindBySlug 鏍规嵁slug鍜岃瑷€鏌ユ壘浜у搧
func (r *ProductRepository) FindBySlug(slug, locale string) (*product.Product, error) {
	var p product.Product
	err := r.db.Preload("Images", func(db *gorm.DB) *gorm.DB {
		return orderProductImages(db)
	}).Preload("ProductType.SpecDefinitions", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
	}).Preload("SpecValues.SpecDefinition", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
	}).Preload("Variants", func(db *gorm.DB) *gorm.DB {
		return orderProductVariants(db)
	}).Where("slug = ? AND locale = ?", slug, locale).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// FindBySKU 鏍规嵁SKU鏌ユ壘浜у搧
func (r *ProductRepository) FindBySKU(sku string) (*product.Product, error) {
	var p product.Product
	err := r.db.Preload("Images").
		Preload("Variants", func(db *gorm.DB) *gorm.DB { return orderProductVariants(db) }).
		Where("sku = ?", sku).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepository) FindProductsByIDs(ids []uint) ([]product.Product, error) {
	var products []product.Product
	if len(ids) == 0 {
		return products, nil
	}
	err := r.db.Preload("Variants", func(db *gorm.DB) *gorm.DB {
		return orderProductVariants(db)
	}).Where("id IN ?", ids).Find(&products).Error
	return products, err
}

// Update 鏇存柊浜у搧
func (r *ProductRepository) Update(p *product.Product) error {
	return r.db.Save(p).Error
}

func (r *ProductRepository) UpdateWithSpecValues(p *product.Product, specValues []product.ProductSpecValue, replaceSpecs bool) error {
	return r.UpdateWithSpecValuesAndVariants(p, specValues, replaceSpecs, nil, false)
}

func (r *ProductRepository) UpdateWithSpecValuesAndVariants(p *product.Product, specValues []product.ProductSpecValue, replaceSpecs bool, variants []product.ProductVariant, replaceVariants bool) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if replaceVariants {
			syncProductSummaryFromVariants(p, variants)
		}
		if err := tx.Save(p).Error; err != nil {
			return err
		}

		if replaceSpecs {
			if err := tx.Where("product_id = ?", p.ID).Delete(&product.ProductSpecValue{}).Error; err != nil {
				return err
			}

			if len(specValues) > 0 {
				for i := range specValues {
					specValues[i].ProductID = p.ID
				}
				if err := tx.Create(&specValues).Error; err != nil {
					return err
				}
			}
		}

		if replaceVariants {
			if err := replaceProductVariants(tx, p.ID, variants); err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *ProductRepository) Delete(id uint) error {
	return r.db.Delete(&product.Product{}, id).Error
}

func (r *ProductRepository) IncrementViewCount(id uint) error {
	return r.db.Model(&product.Product{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}

// UpdateStatus 鏇存柊鍟嗗搧鐘舵€?
func (r *ProductRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&product.Product{}).Where("id = ?", id).Update("status", status).Error
}

// GetStats 鑾峰彇鍟嗗搧缁熻
func (r *ProductRepository) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 鎬诲晢鍝佹暟
	var total int64
	if err := r.db.Model(&product.Product{}).Count(&total).Error; err != nil {
		return nil, err
	}
	stats["total"] = total

	// 鎸夌姸鎬佺粺璁?
	var statusStats []struct {
		Status string
		Count  int64
	}
	if err := r.db.Model(&product.Product{}).Select("status, COUNT(*) as count").Group("status").Scan(&statusStats).Error; err != nil {
		return nil, err
	}

	for _, stat := range statusStats {
		stats[stat.Status] = stat.Count
	}

	// 绮鹃€夊晢鍝佹暟
	var featuredCount int64
	if err := r.db.Model(&product.Product{}).Where("featured = ?", true).Count(&featuredCount).Error; err != nil {
		return nil, err
	}
	stats["featured"] = featuredCount

	// 浣庡簱瀛樺晢鍝佹暟锛堝簱瀛?< 10锛?
	var lowStockCount int64
	if err := r.db.Model(&product.Product{}).Where("stock < ? AND stock > 0", 10).Count(&lowStockCount).Error; err != nil {
		return nil, err
	}
	stats["low_stock"] = lowStockCount

	// 缂鸿揣鍟嗗搧鏁?
	var outOfStockCount int64
	if err := r.db.Model(&product.Product{}).Where("stock = 0").Count(&outOfStockCount).Error; err != nil {
		return nil, err
	}
	stats["out_of_stock"] = outOfStockCount

	return stats, nil
}
