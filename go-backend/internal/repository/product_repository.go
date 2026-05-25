package repository

import (
	"tanzanite/internal/domain/product"

	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Create 创建产品
func (r *ProductRepository) Create(p *product.Product) error {
	return r.db.Create(p).Error
}

// FindByID 根据ID查找产品
func (r *ProductRepository) FindByID(id uint) (*product.Product, error) {
	var p product.Product
	err := r.db.Preload("Images", func(db *gorm.DB) *gorm.DB {
		return db.Order("product_images.order ASC")
	}).First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// FindBySlug 根据slug和语言查找产品
func (r *ProductRepository) FindBySlug(slug, locale string) (*product.Product, error) {
	var p product.Product
	err := r.db.Preload("Images", func(db *gorm.DB) *gorm.DB {
		return db.Order("product_images.order ASC")
	}).Where("slug = ? AND locale = ?", slug, locale).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// FindBySKU 根据SKU查找产品
func (r *ProductRepository) FindBySKU(sku string) (*product.Product, error) {
	var p product.Product
	err := r.db.Preload("Images").Where("sku = ?", sku).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Update 更新产品
func (r *ProductRepository) Update(p *product.Product) error {
	return r.db.Save(p).Error
}

// Delete 删除产品（软删除）
func (r *ProductRepository) Delete(id uint) error {
	return r.db.Delete(&product.Product{}, id).Error
}

// List 获取产品列表
func (r *ProductRepository) List(locale, status string, featured bool, offset, limit int) ([]product.Product, int64, error) {
	var products []product.Product
	var total int64

	query := r.db.Model(&product.Product{}).Preload("Images", func(db *gorm.DB) *gorm.DB {
		return db.Order("product_images.order ASC")
	})

	if locale != "" {
		query = query.Where("locale = ?", locale)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if featured {
		query = query.Where("featured = ?", true)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&products).Error
	return products, total, err
}

// UpdateStock 更新库存
func (r *ProductRepository) UpdateStock(id uint, quantity int) error {
	return r.db.Model(&product.Product{}).Where("id = ?", id).Update("stock", quantity).Error
}

// IncrementViewCount 增加浏览次数
func (r *ProductRepository) IncrementViewCount(id uint) error {
	return r.db.Model(&product.Product{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}

// FindAllWithFilters 根据筛选条件获取商品列表
func (r *ProductRepository) FindAllWithFilters(page, pageSize int, status, locale, search, featured string) ([]product.Product, int64, error) {
	var products []product.Product
	var total int64

	query := r.db.Model(&product.Product{}).Preload("Images", func(db *gorm.DB) *gorm.DB {
		return db.Order("product_images.order ASC")
	})

	// 应用筛选条件
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if locale != "" {
		query = query.Where("locale = ?", locale)
	}
	if search != "" {
		query = query.Where("name LIKE ? OR sku LIKE ? OR description LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if featured == "true" {
		query = query.Where("featured = ?", true)
	} else if featured == "false" {
		query = query.Where("featured = ?", false)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&products).Error

	return products, total, err
}

// UpdateStatus 更新商品状态
func (r *ProductRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&product.Product{}).Where("id = ?", id).Update("status", status).Error
}

// GetStats 获取商品统计
func (r *ProductRepository) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总商品数
	var total int64
	if err := r.db.Model(&product.Product{}).Count(&total).Error; err != nil {
		return nil, err
	}
	stats["total"] = total

	// 按状态统计
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

	// 精选商品数
	var featuredCount int64
	if err := r.db.Model(&product.Product{}).Where("featured = ?", true).Count(&featuredCount).Error; err != nil {
		return nil, err
	}
	stats["featured"] = featuredCount

	// 低库存商品数（库存 < 10）
	var lowStockCount int64
	if err := r.db.Model(&product.Product{}).Where("stock < ? AND stock > 0", 10).Count(&lowStockCount).Error; err != nil {
		return nil, err
	}
	stats["low_stock"] = lowStockCount

	// 缺货商品数
	var outOfStockCount int64
	if err := r.db.Model(&product.Product{}).Where("stock = 0").Count(&outOfStockCount).Error; err != nil {
		return nil, err
	}
	stats["out_of_stock"] = outOfStockCount

	return stats, nil
}
