package repository

import (
	"context"
	"fmt"
	"strings"
	"tanzanite/internal/domain/product"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProductRepository struct {
	db *gorm.DB
}

type ProductSearchQuery struct {
	Locale      string
	Status      string
	Keyword     string
	TypeSlug    string
	PriceMin    *float64
	PriceMax    *float64
	SpecFilters map[string][]string
	Offset      int
	Limit       int
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// WithTx 复用事务 db 实例
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

// Create 创建产品
func (r *ProductRepository) Create(p *product.Product) error {
	return r.db.Create(p).Error
}

func (r *ProductRepository) CreateWithSpecValues(p *product.Product, specValues []product.ProductSpecValue) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(p).Error; err != nil {
			return err
		}

		if len(specValues) == 0 {
			return nil
		}

		for i := range specValues {
			specValues[i].ProductID = p.ID
		}

		return tx.Create(&specValues).Error
	})
}

// FindByID 根据ID查找产品
func (r *ProductRepository) FindByID(id uint) (*product.Product, error) {
	var p product.Product
	err := r.db.Preload("Images", func(db *gorm.DB) *gorm.DB {
		return orderProductImages(db)
	}).Preload("ProductType.SpecDefinitions", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
	}).Preload("SpecValues.SpecDefinition", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
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
		return orderProductImages(db)
	}).Preload("ProductType.SpecDefinitions", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
	}).Preload("SpecValues.SpecDefinition", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
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

func (r *ProductRepository) UpdateWithSpecValues(p *product.Product, specValues []product.ProductSpecValue, replaceSpecs bool) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(p).Error; err != nil {
			return err
		}

		if !replaceSpecs {
			return nil
		}

		if err := tx.Where("product_id = ?", p.ID).Delete(&product.ProductSpecValue{}).Error; err != nil {
			return err
		}

		if len(specValues) == 0 {
			return nil
		}

		for i := range specValues {
			specValues[i].ProductID = p.ID
		}

		return tx.Create(&specValues).Error
	})
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
		return orderProductImages(db)
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

func (r *ProductRepository) SearchPublic(input ProductSearchQuery) ([]product.Product, int64, error) {
	var products []product.Product
	var total int64

	query := r.db.Model(&product.Product{}).Preload("Images", func(db *gorm.DB) *gorm.DB {
		return orderProductImages(db)
	})

	if input.Locale != "" {
		query = query.Where("products.locale = ?", input.Locale)
	}
	if input.Status != "" {
		query = query.Where("products.status = ?", input.Status)
	}
	if input.TypeSlug != "" {
		query = query.Joins("JOIN product_types ON product_types.id = products.product_type_id AND product_types.slug = ?", input.TypeSlug)
	}
	if input.PriceMin != nil {
		query = query.Where("COALESCE(products.sale_price, products.price) >= ?", *input.PriceMin)
	}
	if input.PriceMax != nil {
		query = query.Where("COALESCE(products.sale_price, products.price) <= ?", *input.PriceMax)
	}
	if input.Keyword != "" {
		pattern := "%" + strings.ToLower(input.Keyword) + "%"
		query = query.Where("LOWER(products.name) LIKE ? OR LOWER(products.sku) LIKE ? OR LOWER(products.short_desc) LIKE ? OR LOWER(products.description) LIKE ?", pattern, pattern, pattern, pattern)
	}

	filterIndex := 0
	for slug, values := range input.SpecFilters {
		if slug == "" || len(values) == 0 {
			continue
		}

		valueAlias := fmt.Sprintf("psv_%d", filterIndex)
		defAlias := fmt.Sprintf("psd_%d", filterIndex)
		query = query.
			Joins(fmt.Sprintf("JOIN product_spec_values %s ON %s.product_id = products.id", valueAlias, valueAlias)).
			Joins(fmt.Sprintf("JOIN product_spec_definitions %s ON %s.id = %s.spec_definition_id AND %s.slug = ?", defAlias, defAlias, valueAlias, defAlias), slug).
			Where(fmt.Sprintf("%s.value IN ?", valueAlias), values)
		filterIndex++
	}

	if err := query.Distinct("products.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Distinct("products.*").Order("products.updated_at DESC").Offset(input.Offset).Limit(input.Limit).Find(&products).Error
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

	finalArgs := make([]interface{}, 0, len(args)+1+len(stockArgs))
	finalArgs = append(finalArgs, args...)
	finalArgs = append(finalArgs, ids)
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
		return orderProductImages(db)
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
	switch featured {
	case "true":
		query = query.Where("featured = ?", true)
	case "false":
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

func (r *ProductRepository) FindAllProductTypes(includeDisabled bool) ([]product.ProductType, error) {
	var productTypes []product.ProductType
	query := r.db.Preload("SpecDefinitions", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
	})
	if !includeDisabled {
		query = query.Where("is_enabled = ?", true)
	}

	err := query.Order("sort_order ASC, id ASC").Find(&productTypes).Error
	return productTypes, err
}

func (r *ProductRepository) FindProductTypeByID(id uint) (*product.ProductType, error) {
	var productType product.ProductType
	err := r.db.Preload("SpecDefinitions", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
	}).First(&productType, id).Error
	if err != nil {
		return nil, err
	}
	return &productType, nil
}

func (r *ProductRepository) FindProductTypeBySlug(slug string) (*product.ProductType, error) {
	var productType product.ProductType
	err := r.db.Preload("SpecDefinitions", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
	}).Where("slug = ?", slug).First(&productType).Error
	if err != nil {
		return nil, err
	}
	return &productType, nil
}

// SemanticSearchPublic performs a vector similarity search using pgvector (Stub)
func (r *ProductRepository) SemanticSearchPublic(ctx context.Context, query string) ([]product.Product, error) {
	// Stub: This is a placeholder for actual OpenAI embedding generation and pgvector search.
	// 1. Generate embedding using openai.Client (e.g. client.CreateEmbeddings(ctx, req))
	// 2. Search using GORM and pgvector's <=> operator:
	// r.db.WithContext(ctx).Order("embedding <=> ?", embedding).Limit(10).Find(&products)

	var products []product.Product
	return products, nil
}
