package repository

import (
	"commerce-platform/internal/domain/product"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"gorm.io/gorm"
)

type ProductSearchQuery struct {
	Locale      string
	Status      string
	Keyword     string
	TypeSlug    string
	BrandSlug   string
	PriceMin    *float64
	PriceMax    *float64
	SpecFilters map[string][]string
	Offset      int
	Limit       int
}

type ProductQuickBuyCandidateQuery struct {
	Locale         string
	ProductTypeIDs []uint
	Keyword        string
	SpecFilters    map[string][]string
	Offset         int
	Limit          int
}

type ProductRecommendationQuery struct {
	Locale            string
	ProductTypeID     *uint
	Keyword           string
	ExcludeProductIDs []uint
	Offset            int
	Limit             int
}

func activeVariantExistsSQL(alias string) string {
	return fmt.Sprintf(`EXISTS (
		SELECT 1 FROM product_variants %s
		WHERE %s.product_id = products.id
		  AND %s.deleted_at IS NULL
		  AND %s.is_active = TRUE
	)`, alias, alias, alias, alias)
}

func applyQuickBuyCandidateScope(query *gorm.DB, input ProductQuickBuyCandidateQuery) *gorm.DB {
	query = query.
		Where("products.status = ?", "active").
		Where(activeVariantExistsSQL("pv_quick_buy_candidate")).
		Where(`EXISTS (
			SELECT 1
			FROM product_variants pv_quick_buy_candidate_stock
			WHERE pv_quick_buy_candidate_stock.product_id = products.id
			  AND pv_quick_buy_candidate_stock.deleted_at IS NULL
			  AND pv_quick_buy_candidate_stock.is_active = TRUE
			  AND pv_quick_buy_candidate_stock.stock > 0
		)`)

	if len(input.ProductTypeIDs) > 0 {
		query = query.Where("products.product_type_id IN ?", input.ProductTypeIDs)
	}
	if input.Locale != "" {
		query = query.Where("products.locale = ?", input.Locale)
	}
	if input.Keyword != "" {
		pattern := "%" + strings.ToLower(input.Keyword) + "%"
		query = query.Joins("LEFT JOIN product_types quick_buy_product_types ON quick_buy_product_types.id = products.product_type_id").
			Where(`
				LOWER(products.name) LIKE ?
				OR LOWER(products.sku) LIKE ?
				OR LOWER(products.short_desc) LIKE ?
				OR LOWER(products.description) LIKE ?
				OR LOWER(quick_buy_product_types.name) LIKE ?
				OR LOWER(quick_buy_product_types.slug) LIKE ?
			`, pattern, pattern, pattern, pattern, pattern, pattern)
	}
	return query
}

func normalizeSpecFilterValues(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, raw := range values {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func applyProductSpecFilters(query *gorm.DB, specFilters map[string][]string, dialect string) (*gorm.DB, error) {
	slugs := make([]string, 0, len(specFilters))
	valuesBySlug := make(map[string][]string, len(specFilters))
	for rawSlug, rawValues := range specFilters {
		slug := strings.TrimSpace(rawSlug)
		values := normalizeSpecFilterValues(rawValues)
		if slug == "" || len(values) == 0 {
			continue
		}
		slugs = append(slugs, slug)
		valuesBySlug[slug] = values
	}
	sort.Strings(slugs)

	for index, slug := range slugs {
		values := valuesBySlug[slug]
		valueAlias := fmt.Sprintf("psv_%d", index)
		defAlias := fmt.Sprintf("psd_%d", index)
		variantAlias := fmt.Sprintf("pvv_%d", index)
		variantConditions := make([]string, 0, len(values))
		var variantArgs []interface{}
		for _, value := range values {
			if dialect == "postgres" {
				filter, err := json.Marshal(map[string]string{slug: value})
				if err != nil {
					return nil, fmt.Errorf("marshal variant option filter: %w", err)
				}
				variantConditions = append(variantConditions, fmt.Sprintf("(%s.option_values)::jsonb @> ?::jsonb", variantAlias))
				variantArgs = append(variantArgs, string(filter))
				continue
			}

			variantConditions = append(variantConditions, fmt.Sprintf("%s.option_values LIKE ?", variantAlias))
			variantArgs = append(variantArgs, fmt.Sprintf("%%\"%s\":\"%s\"%%", slug, value))
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
		)`, valueAlias, defAlias, defAlias, valueAlias, valueAlias, defAlias, valueAlias, variantAlias, variantAlias, variantAlias, variantAlias, strings.Join(variantConditions, " OR ")), append([]interface{}{slug, values}, variantArgs...)...)
	}
	return query, nil
}

func (r *ProductRepository) List(locale, status string, featured bool, offset, limit int) ([]product.Product, int64, error) {
	var products []product.Product
	var total int64

	query := r.db.Model(&product.Product{}).Preload("Brand").Preload("Media", func(db *gorm.DB) *gorm.DB {
		return orderProductMedia(db)
	}).Preload("ProductType.SpecDefinitions", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
	}).Preload("ProductType.Translations", func(db *gorm.DB) *gorm.DB {
		return db.Order("locale ASC, id ASC")
	}).Preload("Variants", func(db *gorm.DB) *gorm.DB {
		return orderProductVariants(db)
	})
	query = r.preloadProductVariantOptionValues(query).
		Preload("AfterSalesTemplate").
		Preload("PackagingTemplate").
		Where(activeVariantExistsSQL("pv_list"))

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

// ListPublicAvailable returns active products with at least one active variant
// that can currently be purchased.
func (r *ProductRepository) ListPublicAvailable(locale string, offset, limit int) ([]product.Product, int64, error) {
	var products []product.Product
	var total int64

	query := r.db.Model(&product.Product{}).
		Preload("Brand").
		Preload("Media", func(db *gorm.DB) *gorm.DB {
			return orderProductMedia(db)
		}).
		Preload("ProductType.SpecDefinitions", func(db *gorm.DB) *gorm.DB {
			return orderSpecDefinitions(db)
		}).
		Preload("ProductType.Translations", func(db *gorm.DB) *gorm.DB {
			return db.Order("locale ASC, id ASC")
		}).
		Preload("Variants", func(db *gorm.DB) *gorm.DB {
			return orderProductVariants(db)
		})
	query = r.preloadProductVariantOptionValues(query).
		Preload("AfterSalesTemplate").
		Preload("PackagingTemplate").
		Where("products.status = ?", "active").
		Where(activeVariantExistsSQL("pv_recommendation")).
		Where(`EXISTS (
			SELECT 1
			FROM product_variants pv_recommendation_stock
			WHERE pv_recommendation_stock.product_id = products.id
			  AND pv_recommendation_stock.deleted_at IS NULL
			  AND pv_recommendation_stock.is_active = TRUE
			  AND pv_recommendation_stock.stock > 0
		)`)

	if locale != "" {
		query = query.Where("products.locale = ?", locale)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("products.featured DESC").
		Order("products.view_count DESC").
		Order("products.created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&products).Error
	return products, total, err
}

func (r *ProductRepository) ListRecommendationCandidates(input ProductRecommendationQuery) ([]product.Product, int64, error) {
	var products []product.Product
	var total int64

	query := r.db.Model(&product.Product{}).
		Preload("Brand").
		Preload("Media", orderProductMedia).
		Preload("ProductType.SpecDefinitions", orderSpecDefinitions).
		Preload("ProductType.Translations", func(db *gorm.DB) *gorm.DB {
			return db.Order("locale ASC, id ASC")
		}).
		Preload("SpecValues.SpecDefinition", orderSpecDefinitions).
		Preload("Variants", orderProductVariants)
	query = r.preloadProductVariantOptionValues(query).
		Preload("AfterSalesTemplate").
		Preload("PackagingTemplate").
		Where("products.status = ?", "active").
		Where(activeVariantExistsSQL("pv_recommendation_candidate")).
		Where(`EXISTS (
			SELECT 1
			FROM product_variants pv_recommendation_candidate_stock
			WHERE pv_recommendation_candidate_stock.product_id = products.id
			  AND pv_recommendation_candidate_stock.deleted_at IS NULL
			  AND pv_recommendation_candidate_stock.is_active = TRUE
			  AND pv_recommendation_candidate_stock.stock > 0
		)`)

	if input.Locale != "" {
		query = query.Where("products.locale = ?", input.Locale)
	}
	if input.ProductTypeID != nil && *input.ProductTypeID > 0 {
		query = query.Where("products.product_type_id = ?", *input.ProductTypeID)
	}
	if len(input.ExcludeProductIDs) > 0 {
		query = query.Where("products.id NOT IN ?", input.ExcludeProductIDs)
	}
	if input.Keyword != "" {
		pattern := "%" + strings.ToLower(input.Keyword) + "%"
		query = query.Joins("LEFT JOIN product_types recommendation_product_types ON recommendation_product_types.id = products.product_type_id").
			Where(`
				LOWER(products.name) LIKE ?
				OR LOWER(products.sku) LIKE ?
				OR LOWER(products.short_desc) LIKE ?
				OR LOWER(products.description) LIKE ?
				OR LOWER(recommendation_product_types.name) LIKE ?
				OR LOWER(recommendation_product_types.slug) LIKE ?
			`, pattern, pattern, pattern, pattern, pattern, pattern)
	}

	if err := query.Distinct("products.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Distinct("products.*").
		Order("products.featured DESC").
		Order("products.view_count DESC").
		Order("products.created_at DESC").
		Offset(input.Offset).
		Limit(input.Limit).
		Find(&products).Error
	return products, total, err
}

func (r *ProductRepository) ListQuickBuyCandidates(input ProductQuickBuyCandidateQuery) ([]product.Product, int64, error) {
	products := []product.Product{}
	var total int64

	query := r.db.Model(&product.Product{}).
		Preload("Brand").
		Preload("Media", orderProductMedia).
		Preload("ProductType.SpecDefinitions", orderSpecDefinitions).
		Preload("ProductType.Translations", func(db *gorm.DB) *gorm.DB {
			return db.Order("locale ASC, id ASC")
		}).
		Preload("SpecValues.SpecDefinition", orderSpecDefinitions).
		Preload("Variants", orderProductVariants)
	query = r.preloadProductVariantOptionValues(query).
		Preload("AfterSalesTemplate").
		Preload("PackagingTemplate")

	query = applyQuickBuyCandidateScope(query, input)
	var err error
	query, err = applyProductSpecFilters(query, input.SpecFilters, r.db.Dialector.Name())
	if err != nil {
		return nil, 0, err
	}

	if err := query.Distinct("products.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err = query.
		Distinct("products.*").
		Order("products.featured DESC").
		Order("products.view_count DESC").
		Order("products.updated_at DESC").
		Order("products.id DESC").
		Offset(input.Offset).
		Limit(input.Limit).
		Find(&products).Error
	return products, total, err
}

func (r *ProductRepository) ListQuickBuyFilterValues(input ProductQuickBuyCandidateQuery, slugs []string) (map[string][]string, error) {
	result := make(map[string][]string)
	normalizedSlugs := make([]string, 0, len(slugs))
	seenSlugs := make(map[string]struct{}, len(slugs))
	for _, rawSlug := range slugs {
		slug := strings.TrimSpace(rawSlug)
		if slug == "" {
			continue
		}
		if _, exists := seenSlugs[slug]; exists {
			continue
		}
		seenSlugs[slug] = struct{}{}
		normalizedSlugs = append(normalizedSlugs, slug)
		result[slug] = []string{}
	}
	if len(normalizedSlugs) == 0 {
		return result, nil
	}

	valuesBySlug := make(map[string]map[string]struct{}, len(normalizedSlugs))
	for _, slug := range normalizedSlugs {
		valuesBySlug[slug] = make(map[string]struct{})
	}
	for _, slug := range normalizedSlugs {
		productIDs, err := r.quickBuyCandidateProductIDs(productQuickBuyCandidateQueryWithoutSpecFilter(input, slug))
		if err != nil {
			return nil, err
		}
		if len(productIDs) == 0 {
			continue
		}

		var specValues []string
		if err := r.db.Table("product_spec_values AS psv").
			Select("psv.value").
			Joins("JOIN product_spec_definitions AS psd ON psd.id = psv.spec_definition_id").
			Where("psv.product_id IN ?", productIDs).
			Where("psd.slug = ?", slug).
			Pluck("psv.value", &specValues).Error; err != nil {
			return nil, err
		}
		for _, rawValue := range specValues {
			value := strings.TrimSpace(rawValue)
			if value != "" {
				valuesBySlug[slug][value] = struct{}{}
			}
		}

		var rawVariantValues []string
		if err := r.db.Model(&product.ProductVariant{}).
			Where("product_id IN ?", productIDs).
			Where("deleted_at IS NULL").
			Where("is_active = ?", true).
			Pluck("option_values", &rawVariantValues).Error; err != nil {
			return nil, err
		}
		for _, raw := range rawVariantValues {
			var object map[string]interface{}
			if err := json.Unmarshal([]byte(raw), &object); err != nil {
				continue
			}
			value := strings.TrimSpace(fmt.Sprint(object[slug]))
			if value != "" && value != "<nil>" {
				valuesBySlug[slug][value] = struct{}{}
			}
		}
	}

	for _, slug := range normalizedSlugs {
		values := make([]string, 0, len(valuesBySlug[slug]))
		for value := range valuesBySlug[slug] {
			values = append(values, value)
		}
		sort.Strings(values)
		result[slug] = values
	}
	return result, nil
}

func (r *ProductRepository) quickBuyCandidateProductIDs(input ProductQuickBuyCandidateQuery) ([]uint, error) {
	query := applyQuickBuyCandidateScope(r.db.Model(&product.Product{}), input)
	filteredQuery, err := applyProductSpecFilters(query, input.SpecFilters, r.db.Dialector.Name())
	if err != nil {
		return nil, err
	}
	var productIDs []uint
	if err := filteredQuery.Select("DISTINCT products.id").Pluck("products.id", &productIDs).Error; err != nil {
		return nil, err
	}
	return productIDs, nil
}

func productQuickBuyCandidateQueryWithoutSpecFilter(input ProductQuickBuyCandidateQuery, slug string) ProductQuickBuyCandidateQuery {
	if len(input.SpecFilters) == 0 {
		return input
	}
	next := input
	next.SpecFilters = make(map[string][]string, len(input.SpecFilters))
	for key, values := range input.SpecFilters {
		if key == slug {
			continue
		}
		next.SpecFilters[key] = values
	}
	return next
}

func (r *ProductRepository) SearchPublic(input ProductSearchQuery) ([]product.Product, int64, error) {
	var products []product.Product
	var total int64

	query := r.db.Model(&product.Product{}).Preload("Brand").Preload("Media", func(db *gorm.DB) *gorm.DB {
		return orderProductMedia(db)
	}).Preload("ProductType.SpecDefinitions", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
	}).Preload("ProductType.Translations", func(db *gorm.DB) *gorm.DB {
		return db.Order("locale ASC, id ASC")
	}).Preload("Variants", func(db *gorm.DB) *gorm.DB {
		return orderProductVariants(db)
	})
	query = r.preloadProductVariantOptionValues(query).
		Preload("AfterSalesTemplate").
		Preload("PackagingTemplate").
		Where(activeVariantExistsSQL("pv_public"))

	if input.Locale != "" {
		query = query.Where("products.locale = ?", input.Locale)
	}
	if input.Status != "" {
		query = query.Where("products.status = ?", input.Status)
	}
	if input.TypeSlug != "" {
		query = query.Joins("JOIN product_types ON product_types.id = products.product_type_id AND product_types.slug = ?", input.TypeSlug)
	}
	if input.BrandSlug != "" {
		query = query.Joins("JOIN product_brands ON product_brands.id = products.brand_id AND product_brands.slug = ? AND product_brands.is_enabled = TRUE", input.BrandSlug)
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

	var err error
	query, err = applyProductSpecFilters(query, input.SpecFilters, r.db.Dialector.Name())
	if err != nil {
		return nil, 0, err
	}

	if err := query.Distinct("products.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err = query.Distinct("products.*").Order("products.updated_at DESC").Offset(input.Offset).Limit(input.Limit).Find(&products).Error
	return products, total, err
}

func (r *ProductRepository) FindAllWithFilters(page, pageSize int, status, locale, search, featured string) ([]product.Product, int64, error) {
	var products []product.Product
	var total int64

	query := r.db.Model(&product.Product{}).Preload("Brand").Preload("Media", func(db *gorm.DB) *gorm.DB {
		return orderProductMedia(db)
	}).Preload("Variants", func(db *gorm.DB) *gorm.DB {
		return orderProductVariants(db)
	}).Preload("AfterSalesTemplate").Preload("PackagingTemplate")

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
