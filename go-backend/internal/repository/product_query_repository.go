package repository

import (
	"context"
	"fmt"
	"strings"
	"tanzanite/internal/domain/product"

	"gorm.io/gorm"
)

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

func (r *ProductRepository) FindAllWithFilters(page, pageSize int, status, locale, search, featured string) ([]product.Product, int64, error) {
	var products []product.Product
	var total int64

	query := r.db.Model(&product.Product{}).Preload("Images", func(db *gorm.DB) *gorm.DB {
		return orderProductImages(db)
	}).Preload("Variants", func(db *gorm.DB) *gorm.DB {
		return orderProductVariants(db)
	})

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

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&products).Error

	return products, total, err
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
