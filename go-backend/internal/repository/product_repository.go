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

// FindProductsByIDs 鏍规嵁IDs鏌ユ壘浜у搧鍒楄〃
func (r *ProductRepository) FindVariantBySKU(sku string) (*product.ProductVariant, error) {
	var variant product.ProductVariant
	if err := r.db.Where("sku = ?", sku).First(&variant).Error; err != nil {
		return nil, err
	}
	return &variant, nil
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

// Delete 鍒犻櫎浜у搧锛堣蒋鍒犻櫎锛?
func syncProductSummaryFromVariants(p *product.Product, variants []product.ProductVariant) {
	if len(variants) == 0 {
		return
	}

	defaultIndex := -1
	totalStock := 0
	for i, variant := range variants {
		if variant.IsActive {
			totalStock += variant.Stock
		}
		if variant.IsActive && variant.IsDefault {
			defaultIndex = i
		}
	}
	if defaultIndex == -1 {
		for i, variant := range variants {
			if variant.IsActive {
				defaultIndex = i
				break
			}
		}
	}
	if defaultIndex == -1 {
		defaultIndex = 0
	}

	defaultVariant := variants[defaultIndex]
	p.SKU = defaultVariant.SKU
	p.Price = defaultVariant.Price
	p.SalePrice = defaultVariant.SalePrice
	p.Stock = totalStock
	if defaultVariant.Weight > 0 {
		p.Weight = defaultVariant.Weight
	}
}

func replaceProductVariants(tx *gorm.DB, productID uint, variants []product.ProductVariant) error {
	var existingVariants []product.ProductVariant
	if err := tx.Where("product_id = ?", productID).Find(&existingVariants).Error; err != nil {
		return err
	}

	existingByID := make(map[uint]product.ProductVariant, len(existingVariants))
	existingBySKU := make(map[string]product.ProductVariant, len(existingVariants))
	for _, variant := range existingVariants {
		existingByID[variant.ID] = variant
		existingBySKU[variant.SKU] = variant
	}

	keepIDs := make([]uint, 0, len(variants))
	for i := range variants {
		variants[i].ProductID = productID
		if variants[i].ID == 0 {
			if existing, ok := existingBySKU[variants[i].SKU]; ok {
				variants[i].ID = existing.ID
			}
		}

		if variants[i].ID != 0 {
			if _, ok := existingByID[variants[i].ID]; !ok {
				return fmt.Errorf("variant %d does not belong to product %d", variants[i].ID, productID)
			}
			if err := tx.Save(&variants[i]).Error; err != nil {
				return err
			}
			keepIDs = append(keepIDs, variants[i].ID)
			continue
		}

		if err := tx.Create(&variants[i]).Error; err != nil {
			return err
		}
		keepIDs = append(keepIDs, variants[i].ID)
	}

	deleteQuery := tx.Where("product_id = ?", productID)
	if len(keepIDs) > 0 {
		deleteQuery = deleteQuery.Where("id NOT IN ?", keepIDs)
	}
	return deleteQuery.Delete(&product.ProductVariant{}).Error
}

func (r *ProductRepository) Delete(id uint) error {
	return r.db.Delete(&product.Product{}, id).Error
}

// List 鑾峰彇浜у搧鍒楄〃
func activeVariantExistsSQL(alias string) string {
	return fmt.Sprintf(`EXISTS (
		SELECT 1 FROM product_variants %s
		WHERE %s.product_id = products.id
		  AND %s.deleted_at IS NULL
		  AND %s.is_active = TRUE
	)`, alias, alias, alias, alias)
}

func (r *ProductRepository) List(locale, status string, featured bool, offset, limit int) ([]product.Product, int64, error) {
	var products []product.Product
	var total int64

	query := r.db.Model(&product.Product{}).Preload("Images", func(db *gorm.DB) *gorm.DB {
		return orderProductImages(db)
	}).Preload("Variants", func(db *gorm.DB) *gorm.DB {
		return orderProductVariants(db)
	}).Where(activeVariantExistsSQL("pv_list"))

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
	}).Preload("Variants", func(db *gorm.DB) *gorm.DB {
		return orderProductVariants(db)
	}).Where(activeVariantExistsSQL("pv_public"))

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
		query = query.Where(`EXISTS (
			SELECT 1 FROM product_variants pv_price_min
			WHERE pv_price_min.product_id = products.id
			  AND pv_price_min.deleted_at IS NULL
			  AND pv_price_min.is_active = TRUE
			  AND COALESCE(pv_price_min.sale_price, pv_price_min.price) >= ?
		)`, *input.PriceMin)
	}
	if input.PriceMax != nil {
		query = query.Where(`EXISTS (
			SELECT 1 FROM product_variants pv_price_max
			WHERE pv_price_max.product_id = products.id
			  AND pv_price_max.deleted_at IS NULL
			  AND pv_price_max.is_active = TRUE
			  AND COALESCE(pv_price_max.sale_price, pv_price_max.price) <= ?
		)`, *input.PriceMax)
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
		variantAlias := fmt.Sprintf("pvv_%d", filterIndex)
		var variantConditions []string
		var args []interface{}
		args = append(args, slug, values)
		for _, value := range values {
			variantConditions = append(variantConditions, fmt.Sprintf("%s.option_values LIKE ?", variantAlias))
			args = append(args, fmt.Sprintf("%%\"%s\":\"%s\"%%", slug, value))
		}

		query = query.Where(fmt.Sprintf(`(
			EXISTS (
				SELECT 1
				FROM product_spec_values %s
				JOIN product_spec_definitions %s ON %s.id = %s.spec_definition_id
				WHERE %s.product_id = products.id
				  AND %s.slug = ?
				  AND %s.value IN ?
			)
			OR EXISTS (
				SELECT 1
				FROM product_variants %s
				WHERE %s.product_id = products.id
				  AND %s.deleted_at IS NULL
				  AND %s.is_active = TRUE
				  AND (%s)
			)
		)`, valueAlias, defAlias, defAlias, valueAlias, valueAlias, defAlias, valueAlias, variantAlias, variantAlias, variantAlias, variantAlias, strings.Join(variantConditions, " OR ")), args...)
		filterIndex++
	}

	if err := query.Distinct("products.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Distinct("products.*").Order("products.updated_at DESC").Offset(input.Offset).Limit(input.Limit).Find(&products).Error
	return products, total, err
}

// IncrementViewCount 澧炲姞娴忚娆℃暟
func (r *ProductRepository) FindPurchasableVariant(productID uint, variantID *uint) (*product.Product, *product.ProductVariant, error) {
	p, err := r.FindByID(productID)
	if err != nil {
		return nil, nil, err
	}

	activeVariants := p.ActiveVariants()
	if len(activeVariants) == 0 {
		return nil, nil, gorm.ErrRecordNotFound
	}

	if variantID != nil {
		for i := range activeVariants {
			if activeVariants[i].ID == *variantID {
				return p, &activeVariants[i], nil
			}
		}
		return nil, nil, gorm.ErrRecordNotFound
	}

	if variant := p.DefaultVariant(); variant != nil {
		return p, variant, nil
	}

	return nil, nil, gorm.ErrRecordNotFound
}

func (r *ProductRepository) DecrementVariantStocks(items map[uint]int) error {
	if len(items) == 0 {
		return nil
	}

	for variantID, quantity := range items {
		res := r.db.Model(&product.ProductVariant{}).
			Where("id = ? AND is_active = ? AND stock >= ?", variantID, true, quantity).
			UpdateColumn("stock", gorm.Expr("stock - ?", quantity))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("insufficient stock for variant %d or variant not found", variantID)
		}
	}

	return r.refreshProductStockForVariants(items)
}

func (r *ProductRepository) IncrementVariantStock(variantID uint, quantity int) error {
	res := r.db.Model(&product.ProductVariant{}).Where("id = ?", variantID).
		UpdateColumn("stock", gorm.Expr("stock + ?", quantity))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return r.refreshProductStockForVariants(map[uint]int{variantID: quantity})
}

func (r *ProductRepository) refreshProductStockForVariants(items map[uint]int) error {
	if len(items) == 0 {
		return nil
	}

	var productIDs []uint
	if err := r.db.Model(&product.ProductVariant{}).
		Where("id IN ?", uintMapKeys(items)).
		Distinct().
		Pluck("product_id", &productIDs).Error; err != nil {
		return err
	}

	for _, productID := range productIDs {
		var totalStock int64
		if err := r.db.Model(&product.ProductVariant{}).
			Where("product_id = ? AND is_active = ? AND deleted_at IS NULL", productID, true).
			Select("COALESCE(SUM(stock), 0)").
			Scan(&totalStock).Error; err != nil {
			return err
		}
		if err := r.db.Model(&product.Product{}).
			Where("id = ?", productID).
			Update("stock", totalStock).Error; err != nil {
			return err
		}
	}

	return nil
}

func uintMapKeys(items map[uint]int) []uint {
	keys := make([]uint, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	return keys
}

func (r *ProductRepository) IncrementViewCount(id uint) error {
	return r.db.Model(&product.Product{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}

// FindAllWithFilters 鏍规嵁绛涢€夋潯浠惰幏鍙栧晢鍝佸垪琛?
func (r *ProductRepository) FindAllWithFilters(page, pageSize int, status, locale, search, featured string) ([]product.Product, int64, error) {
	var products []product.Product
	var total int64

	query := r.db.Model(&product.Product{}).Preload("Images", func(db *gorm.DB) *gorm.DB {
		return orderProductImages(db)
	}).Preload("Variants", func(db *gorm.DB) *gorm.DB {
		return orderProductVariants(db)
	})

	// 搴旂敤绛涢€夋潯浠?
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

	// 鑾峰彇鎬绘暟
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 鍒嗛〉鏌ヨ
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&products).Error

	return products, total, err
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

// FindAttributeByID 鏍规嵁ID鏌ユ壘灞炴€?
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

// FindAttributeBySlug 鏍规嵁Slug鏌ユ壘灞炴€?
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

// FindAllAttributes 鑾峰彇鎵€鏈夊睘鎬у垪琛?
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

// CreateAttribute 鍒涘缓灞炴€?
func (r *ProductRepository) CreateAttribute(attr *product.ProductAttribute) error {
	return r.db.Create(attr).Error
}

// UpdateAttribute 鏇存柊灞炴€?
func (r *ProductRepository) UpdateAttribute(attr *product.ProductAttribute) error {
	return r.db.Save(attr).Error
}

// DeleteAttribute 鍒犻櫎灞炴€?
func (r *ProductRepository) DeleteAttribute(id uint) error {
	// 鍏堝垹闄ゅ睘鎬у叧鑱旂殑灞炴€у€?
	if err := r.db.Where("attribute_id = ?", id).Delete(&product.AttributeValue{}).Error; err != nil {
		return err
	}
	return r.db.Delete(&product.ProductAttribute{}, id).Error
}

// FindFilterableAttributes 鑾峰彇鍓嶅彴鍙繃婊ょ殑灞炴€?
func (r *ProductRepository) FindFilterableAttributes() ([]product.ProductAttribute, error) {
	var attrs []product.ProductAttribute
	err := r.db.Preload("Values", func(db *gorm.DB) *gorm.DB {
		return db.Where("is_enabled = ?", true).Order("product_attribute_values.sort_order ASC")
	}).Where("is_filterable = ? AND is_enabled = ?", true, true).Order("sort_order ASC").Find(&attrs).Error
	return attrs, err
}

// FindAttributeValueByID 鏍规嵁ID鑾峰彇灞炴€у€?
func (r *ProductRepository) FindAttributeValueByID(id uint) (*product.AttributeValue, error) {
	var val product.AttributeValue
	err := r.db.First(&val, id).Error
	if err != nil {
		return nil, err
	}
	return &val, nil
}

// CreateAttributeValue 鍒涘缓灞炴€у€?
func (r *ProductRepository) CreateAttributeValue(val *product.AttributeValue) error {
	return r.db.Create(val).Error
}

// UpdateAttributeValue 鏇存柊灞炴€у€?
func (r *ProductRepository) UpdateAttributeValue(val *product.AttributeValue) error {
	return r.db.Save(val).Error
}

// DeleteAttributeValue 鍒犻櫎灞炴€у€?
func (r *ProductRepository) DeleteAttributeValue(id uint) error {
	return r.db.Delete(&product.AttributeValue{}, id).Error
}

// FindValuesByAttributeID 鏍规嵁灞炴€D鏌ユ壘灞炴€у€?
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
