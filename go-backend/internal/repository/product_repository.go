package repository

import (
	"fmt"
	"strings"
	"tanzanite/internal/domain/product"

	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// WithTx 复用事务 db 实例
func (r *ProductRepository) WithTx(tx *gorm.DB) *ProductRepository {
	return &ProductRepository{db: tx}
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

// FindProductsByIDs 根据IDs查找产品列表
func (r *ProductRepository) FindProductsByIDs(ids []uint) ([]product.Product, error) {
	var products []product.Product
	if len(ids) == 0 {
		return products, nil
	}
	err := r.db.Where("id IN ?", ids).Find(&products).Error
	return products, err
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

func (r *ProductRepository) SearchPublic(locale, status, keyword string, offset, limit int) ([]product.Product, int64, error) {
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
	if keyword != "" {
		pattern := "%" + strings.ToLower(keyword) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(sku) LIKE ? OR LOWER(short_desc) LIKE ? OR LOWER(description) LIKE ?", pattern, pattern, pattern, pattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("updated_at DESC").Offset(offset).Limit(limit).Find(&products).Error
	return products, total, err
}

// UpdateStock 更新库存 (绝对值，后台管理用)
func (r *ProductRepository) UpdateStock(id uint, quantity int) error {
	return r.db.Model(&product.Product{}).Where("id = ?", id).Update("stock", quantity).Error
}

// DecrementStock 原子扣减库存
func (r *ProductRepository) DecrementStock(id uint, quantity int) error {
	res := r.db.Model(&product.Product{}).Where("id = ? AND stock >= ?", id, quantity).
		UpdateColumn("stock", gorm.Expr("stock - ?", quantity))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // 库存不足或记录不存在
	}
	return nil
}

// DecrementStocks 批量原子扣减库存
func (r *ProductRepository) DecrementStocks(items map[uint]int) error {
	if len(items) == 0 {
		return nil
	}

	var cases []string
	var args []interface{}
	var stockCases []string
	var stockArgs []interface{}
	var ids []uint

	for id, quantity := range items {
		ids = append(ids, id)
		cases = append(cases, "WHEN ? THEN stock - ?")
		args = append(args, id, quantity)

		stockCases = append(stockCases, "WHEN ? THEN ?")
		stockArgs = append(stockArgs, id, quantity)
	}

	query := fmt.Sprintf("UPDATE products SET stock = CASE id %s END WHERE id IN ? AND stock >= CASE id %s END",
		strings.Join(cases, " "),
		strings.Join(stockCases, " "))

	finalArgs := append(args, ids)
	finalArgs = append(finalArgs, stockArgs...)

	res := r.db.Exec(query, finalArgs...)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected < int64(len(items)) {
		return fmt.Errorf("insufficient stock for some items or items not found")
	}
	return nil
}

// IncrementStock 原子增加库存
func (r *ProductRepository) IncrementStock(id uint, quantity int) error {
	return r.db.Model(&product.Product{}).Where("id = ?", id).
		UpdateColumn("stock", gorm.Expr("stock + ?", quantity)).Error
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

// FindAttributeByID 根据ID查找属性
func (r *ProductRepository) FindAttributeByID(id uint) (*product.ProductAttribute, error) {
	var attr product.ProductAttribute
	err := r.db.Preload("Values", func(db *gorm.DB) *gorm.DB {
		return db.Order("product_attribute_values.sort_order ASC")
	}).First(&attr, id).Error
	if err != nil {
		return nil, err
	}
	return &attr, nil
}

// FindAttributeBySlug 根据Slug查找属性
func (r *ProductRepository) FindAttributeBySlug(slug string) (*product.ProductAttribute, error) {
	var attr product.ProductAttribute
	err := r.db.Preload("Values", func(db *gorm.DB) *gorm.DB {
		return db.Order("product_attribute_values.sort_order ASC")
	}).Where("slug = ?", slug).First(&attr).Error
	if err != nil {
		return nil, err
	}
	return &attr, nil
}

// FindAllAttributes 获取所有属性列表
func (r *ProductRepository) FindAllAttributes(page, pageSize int) ([]product.ProductAttribute, int64, error) {
	var attrs []product.ProductAttribute
	var total int64

	query := r.db.Model(&product.ProductAttribute{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Preload("Values", func(db *gorm.DB) *gorm.DB {
		return db.Order("product_attribute_values.sort_order ASC")
	}).Order("sort_order ASC, id ASC").Offset(offset).Limit(pageSize).Find(&attrs).Error

	return attrs, total, err
}

// CreateAttribute 创建属性
func (r *ProductRepository) CreateAttribute(attr *product.ProductAttribute) error {
	return r.db.Create(attr).Error
}

// UpdateAttribute 更新属性
func (r *ProductRepository) UpdateAttribute(attr *product.ProductAttribute) error {
	return r.db.Save(attr).Error
}

// DeleteAttribute 删除属性
func (r *ProductRepository) DeleteAttribute(id uint) error {
	// 先删除属性关联的属性值
	if err := r.db.Where("attribute_id = ?", id).Delete(&product.AttributeValue{}).Error; err != nil {
		return err
	}
	return r.db.Delete(&product.ProductAttribute{}, id).Error
}

// FindFilterableAttributes 获取前台可过滤的属性
func (r *ProductRepository) FindFilterableAttributes() ([]product.ProductAttribute, error) {
	var attrs []product.ProductAttribute
	err := r.db.Preload("Values", func(db *gorm.DB) *gorm.DB {
		return db.Where("is_enabled = ?", true).Order("product_attribute_values.sort_order ASC")
	}).Where("is_filterable = ? AND is_enabled = ?", true, true).Order("sort_order ASC").Find(&attrs).Error
	return attrs, err
}

// FindAttributeValueByID 根据ID获取属性值
func (r *ProductRepository) FindAttributeValueByID(id uint) (*product.AttributeValue, error) {
	var val product.AttributeValue
	err := r.db.First(&val, id).Error
	if err != nil {
		return nil, err
	}
	return &val, nil
}

// CreateAttributeValue 创建属性值
func (r *ProductRepository) CreateAttributeValue(val *product.AttributeValue) error {
	return r.db.Create(val).Error
}

// UpdateAttributeValue 更新属性值
func (r *ProductRepository) UpdateAttributeValue(val *product.AttributeValue) error {
	return r.db.Save(val).Error
}

// DeleteAttributeValue 删除属性值
func (r *ProductRepository) DeleteAttributeValue(id uint) error {
	return r.db.Delete(&product.AttributeValue{}, id).Error
}

// FindValuesByAttributeID 根据属性ID查找属性值
func (r *ProductRepository) FindValuesByAttributeID(attrID uint) ([]product.AttributeValue, error) {
	var values []product.AttributeValue
	err := r.db.Where("attribute_id = ?", attrID).Order("sort_order ASC").Find(&values).Error
	return values, err
}
