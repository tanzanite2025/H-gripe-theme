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
	Locale                           string
	Status                           string
	Featured                         *bool
	Keyword                          string
	ProductSpecificationTemplateSlug string
	CategorySlug                     string
	BrandSlug                        string
	PriceMin                         *float64
	PriceMax                         *float64
	SpecFilters                      map[string][]string
	Offset                           int
	Limit                            int
}

type ProductQuickBuyCandidateQuery struct {
	Locale                          string
	ProductSpecificationTemplateIDs []uint
	ProductCategoryIDs              []uint
	Keyword                         string
	SpecFilters                     map[string][]string
	Offset                          int
	Limit                           int
}

type ProductRecommendationQuery struct {
	Locale                         string
	ProductSpecificationTemplateID *uint
	Keyword                        string
	ExcludeProductIDs              []uint
	Offset                         int
	Limit                          int
}

type ProductCategoryFilterableSpecification struct {
	Slug      string   `json:"slug"`
	Name      string   `json:"name"`
	FieldType string   `json:"field_type"`
	Unit      string   `json:"unit"`
	Options   []string `json:"options"`
	Values    []string `json:"values"`
}

type ProductCustomsSummary struct {
	Total                     int64 `json:"total"`
	Complete                  int64 `json:"complete"`
	Incomplete                int64 `json:"incomplete"`
	MissingHSCode             int64 `json:"missing_hs_code"`
	MissingCNCode             int64 `json:"missing_cn_code"`
	MissingCountryOfOrigin    int64 `json:"missing_country_of_origin"`
	MissingCustomsDescription int64 `json:"missing_customs_description"`
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

	if len(input.ProductSpecificationTemplateIDs) > 0 {
		query = query.Where("products.product_specification_template_id IN ?", input.ProductSpecificationTemplateIDs)
	}
	if len(input.ProductCategoryIDs) > 0 {
		query = query.Where(`products.product_category_id IN (
			WITH RECURSIVE quick_buy_category_tree(id) AS (
				SELECT id
				FROM product_categories
				WHERE id IN ? AND is_enabled = TRUE

				UNION ALL

				SELECT child.id
				FROM product_categories child
				JOIN quick_buy_category_tree parent ON child.parent_id = parent.id
				WHERE child.is_enabled = TRUE
			)
			SELECT id FROM quick_buy_category_tree
		)`, input.ProductCategoryIDs)
	}
	if input.Locale != "" {
		query = query.Where("products.locale = ?", input.Locale)
	}
	if input.Keyword != "" {
		pattern := "%" + strings.ToLower(input.Keyword) + "%"
		query = query.Joins("LEFT JOIN product_specification_templates quick_buy_product_specification_templates ON quick_buy_product_specification_templates.id = products.product_specification_template_id").
			Where(`
				LOWER(products.name) LIKE ?
				OR LOWER(products.sku) LIKE ?
				OR LOWER(products.short_desc) LIKE ?
				OR LOWER(products.description) LIKE ?
				OR LOWER(quick_buy_product_specification_templates.name) LIKE ?
				OR LOWER(quick_buy_product_specification_templates.slug) LIKE ?
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

	query := r.db.Model(&product.Product{}).Preload("Brand").Preload("ProductSpecificationTemplate").Preload("Media", func(db *gorm.DB) *gorm.DB {
		return orderProductMedia(db)
	}).Preload("ProductSpecificationTemplate.SpecDefinitions", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
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
		Preload("ProductSpecificationTemplate.SpecDefinitions", func(db *gorm.DB) *gorm.DB {
			return orderSpecDefinitions(db)
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
		Preload("ProductSpecificationTemplate.SpecDefinitions", orderSpecDefinitions).
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
	if input.ProductSpecificationTemplateID != nil && *input.ProductSpecificationTemplateID > 0 {
		query = query.Where("products.product_specification_template_id = ?", *input.ProductSpecificationTemplateID)
	}
	if len(input.ExcludeProductIDs) > 0 {
		query = query.Where("products.id NOT IN ?", input.ExcludeProductIDs)
	}
	if input.Keyword != "" {
		pattern := "%" + strings.ToLower(input.Keyword) + "%"
		query = query.Joins("LEFT JOIN product_specification_templates recommendation_product_specification_templates ON recommendation_product_specification_templates.id = products.product_specification_template_id").
			Where(`
				LOWER(products.name) LIKE ?
				OR LOWER(products.sku) LIKE ?
				OR LOWER(products.short_desc) LIKE ?
				OR LOWER(products.description) LIKE ?
				OR LOWER(recommendation_product_specification_templates.name) LIKE ?
				OR LOWER(recommendation_product_specification_templates.slug) LIKE ?
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
		Preload("ProductSpecificationTemplate.SpecDefinitions", orderSpecDefinitions).
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

func (r *ProductRepository) ListFilterableSpecificationsForCategory(categorySlug string) ([]ProductCategoryFilterableSpecification, error) {
	categorySlug = strings.TrimSpace(categorySlug)
	if categorySlug == "" {
		return []ProductCategoryFilterableSpecification{}, nil
	}

	var productIDs []uint
	if err := r.db.Table("products").
		Select("DISTINCT products.id").
		Where("products.deleted_at IS NULL").
		Where(`products.product_category_id IN (
			WITH RECURSIVE category_tree(id) AS (
				SELECT id FROM product_categories WHERE slug = ? AND is_enabled = TRUE

				UNION ALL

				SELECT child.id
				FROM product_categories child
				JOIN category_tree parent ON child.parent_id = parent.id
				WHERE child.is_enabled = TRUE
			)
			SELECT id FROM category_tree
		)`, categorySlug).
		Pluck("products.id", &productIDs).Error; err != nil {
		return nil, err
	}
	if len(productIDs) == 0 {
		return []ProductCategoryFilterableSpecification{}, nil
	}

	type definitionRow struct {
		Slug      string
		Name      string
		FieldType string
		Unit      string
		Options   string
	}
	var definitions []definitionRow
	if err := r.db.Table("product_spec_definitions AS definition").
		Select("DISTINCT definition.slug, definition.name, definition.field_type, definition.unit, definition.options").
		Joins("JOIN products ON products.product_specification_template_id = definition.product_specification_template_id").
		Where("products.id IN ?", productIDs).
		Where("definition.is_filterable = TRUE AND definition.is_visible = TRUE").
		Order("definition.slug ASC").
		Scan(&definitions).Error; err != nil {
		return nil, err
	}
	if len(definitions) == 0 {
		return []ProductCategoryFilterableSpecification{}, nil
	}

	result := make([]ProductCategoryFilterableSpecification, 0, len(definitions))
	valuesBySlug := make(map[string]map[string]struct{}, len(definitions))
	knownSlugs := make([]string, 0, len(definitions))
	for _, definition := range definitions {
		slug := strings.TrimSpace(definition.Slug)
		if slug == "" {
			continue
		}
		if _, exists := valuesBySlug[slug]; exists {
			continue
		}
		knownSlugs = append(knownSlugs, slug)
		valuesBySlug[slug] = make(map[string]struct{})
		options := []string{}
		if strings.TrimSpace(definition.Options) != "" {
			_ = json.Unmarshal([]byte(definition.Options), &options)
		}
		for _, option := range options {
			if value := strings.TrimSpace(option); value != "" {
				valuesBySlug[slug][value] = struct{}{}
			}
		}
		result = append(result, ProductCategoryFilterableSpecification{
			Slug:      slug,
			Name:      strings.TrimSpace(definition.Name),
			FieldType: strings.TrimSpace(definition.FieldType),
			Unit:      strings.TrimSpace(definition.Unit),
			Options:   normalizeSpecFilterValues(options),
		})
	}
	if len(knownSlugs) == 0 {
		return []ProductCategoryFilterableSpecification{}, nil
	}

	type valueRow struct {
		Slug  string
		Value string
	}
	var valueRows []valueRow
	if err := r.db.Table("product_spec_values AS value").
		Select("definition.slug, value.value").
		Joins("JOIN product_spec_definitions AS definition ON definition.id = value.spec_definition_id").
		Where("value.product_id IN ?", productIDs).
		Where("definition.slug IN ?", knownSlugs).
		Scan(&valueRows).Error; err != nil {
		return nil, err
	}
	for _, row := range valueRows {
		if values, exists := valuesBySlug[row.Slug]; exists {
			if value := strings.TrimSpace(row.Value); value != "" {
				values[value] = struct{}{}
			}
		}
	}

	var variantOptionValues []string
	if err := r.db.Model(&product.ProductVariant{}).
		Where("product_id IN ?", productIDs).
		Where("deleted_at IS NULL AND is_active = TRUE").
		Pluck("option_values", &variantOptionValues).Error; err != nil {
		return nil, err
	}
	for _, raw := range variantOptionValues {
		var values map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &values); err != nil {
			continue
		}
		for slug, rawValue := range values {
			knownValues, exists := valuesBySlug[slug]
			if !exists {
				continue
			}
			if value := strings.TrimSpace(fmt.Sprint(rawValue)); value != "" && value != "<nil>" {
				knownValues[value] = struct{}{}
			}
		}
	}

	for index := range result {
		values := make([]string, 0, len(valuesBySlug[result[index].Slug]))
		for value := range valuesBySlug[result[index].Slug] {
			values = append(values, value)
		}
		sort.Strings(values)
		result[index].Values = values
	}
	return result, nil
}

func (r *ProductRepository) ListFilterableSpecificationsWithDynamicValuesForCategory(categorySlug string) ([]ProductCategoryFilterableSpecification, error) {
	categorySlug = strings.TrimSpace(categorySlug)
	if categorySlug == "" {
		return []ProductCategoryFilterableSpecification{}, nil
	}

	var productIDs []uint
	if err := r.db.Table("products").
		Select("DISTINCT products.id").
		Where("products.deleted_at IS NULL").
		Where(`products.product_category_id IN (
			WITH RECURSIVE category_tree(id) AS (
				SELECT id FROM product_categories WHERE slug = ? AND is_enabled = TRUE

				UNION ALL

				SELECT child.id
				FROM product_categories child
				JOIN category_tree parent ON child.parent_id = parent.id
				WHERE child.is_enabled = TRUE
			)
			SELECT id FROM category_tree
		)`, categorySlug).
		Pluck("products.id", &productIDs).Error; err != nil {
		return nil, err
	}
	if len(productIDs) == 0 {
		return []ProductCategoryFilterableSpecification{}, nil
	}

	type definitionRow struct {
		Slug      string
		Name      string
		FieldType string
		Unit      string
	}
	var definitions []definitionRow
	if err := r.db.Table("product_spec_definitions AS definition").
		Select("DISTINCT definition.slug, definition.name, definition.field_type, definition.unit").
		Joins("JOIN products ON products.product_specification_template_id = definition.product_specification_template_id").
		Where("products.id IN ?", productIDs).
		Where("definition.is_filterable = TRUE AND definition.is_visible = TRUE").
		Order("definition.slug ASC").
		Scan(&definitions).Error; err != nil {
		return nil, err
	}
	if len(definitions) == 0 {
		return []ProductCategoryFilterableSpecification{}, nil
	}

	result := make([]ProductCategoryFilterableSpecification, 0, len(definitions))
	valuesBySlug := make(map[string]map[string]struct{}, len(definitions))
	knownSlugs := make([]string, 0, len(definitions))
	for _, definition := range definitions {
		slug := strings.TrimSpace(definition.Slug)
		if slug == "" {
			continue
		}
		if _, exists := valuesBySlug[slug]; exists {
			continue
		}
		knownSlugs = append(knownSlugs, slug)
		valuesBySlug[slug] = make(map[string]struct{})
		result = append(result, ProductCategoryFilterableSpecification{
			Slug:      slug,
			Name:      strings.TrimSpace(definition.Name),
			FieldType: strings.TrimSpace(definition.FieldType),
			Unit:      strings.TrimSpace(definition.Unit),
			Options:   []string{},
		})
	}
	if len(knownSlugs) == 0 {
		return []ProductCategoryFilterableSpecification{}, nil
	}

	type valueRow struct {
		Slug  string
		Value string
	}
	var valueRows []valueRow
	if err := r.db.Table("product_spec_values AS value").
		Select("definition.slug, value.value").
		Joins("JOIN product_spec_definitions AS definition ON definition.id = value.spec_definition_id").
		Where("value.product_id IN ?", productIDs).
		Where("definition.slug IN ?", knownSlugs).
		Scan(&valueRows).Error; err != nil {
		return nil, err
	}
	for _, row := range valueRows {
		if values, exists := valuesBySlug[row.Slug]; exists {
			if value := strings.TrimSpace(row.Value); value != "" {
				values[value] = struct{}{}
			}
		}
	}

	var variantOptionValues []string
	if err := r.db.Model(&product.ProductVariant{}).
		Where("product_id IN ?", productIDs).
		Where("deleted_at IS NULL AND is_active = TRUE").
		Pluck("option_values", &variantOptionValues).Error; err != nil {
		return nil, err
	}
	for _, raw := range variantOptionValues {
		var values map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &values); err != nil {
			continue
		}
		for slug, rawValue := range values {
			knownValues, exists := valuesBySlug[slug]
			if !exists {
				continue
			}
			if value := strings.TrimSpace(fmt.Sprint(rawValue)); value != "" && value != "<nil>" {
				knownValues[value] = struct{}{}
			}
		}
	}

	filtered := make([]ProductCategoryFilterableSpecification, 0, len(result))
	for _, item := range result {
		valuesSet := valuesBySlug[item.Slug]
		if len(valuesSet) == 0 {
			continue
		}
		values := make([]string, 0, len(valuesSet))
		for value := range valuesSet {
			values = append(values, value)
		}
		sort.Strings(values)
		item.Values = values
		filtered = append(filtered, item)
	}
	return filtered, nil
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

func (r *ProductRepository) ProductCategoryInQuickBuyScope(productCategoryID *uint, productCategoryIDs []uint) (bool, error) {
	if len(productCategoryIDs) == 0 {
		return true, nil
	}
	if productCategoryID == nil || *productCategoryID == 0 {
		return false, nil
	}

	var count int64
	err := r.db.Table("product_categories").
		Where("id = ?", *productCategoryID).
		Where(`id IN (
			WITH RECURSIVE quick_buy_category_tree(id) AS (
				SELECT id
				FROM product_categories
				WHERE id IN ? AND is_enabled = TRUE

				UNION ALL

				SELECT child.id
				FROM product_categories child
				JOIN quick_buy_category_tree parent ON child.parent_id = parent.id
				WHERE child.is_enabled = TRUE
			)
			SELECT id FROM quick_buy_category_tree
		)`, productCategoryIDs).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
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

	query := r.db.Model(&product.Product{}).Preload("Brand").Preload("ProductSpecificationTemplate").Preload("CustomsClassificationProfile").Preload("Media", func(db *gorm.DB) *gorm.DB {
		return orderProductMedia(db)
	}).Preload("ProductSpecificationTemplate.SpecDefinitions", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
	}).Preload("Variants", func(db *gorm.DB) *gorm.DB {
		return orderProductVariants(db)
	})
	query = r.preloadProductVariantOptionValues(query).
		Preload("AfterSalesTemplate").
		Preload("PackagingTemplate").
		Where(activeVariantExistsSQL("pv_public"))

	var err error
	query, err = applyPublicProductSearchFilters(query, input, r.db.Dialector.Name())
	if err != nil {
		return nil, 0, err
	}

	if err := query.Distinct("products.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err = query.Distinct("products.*").Order("products.updated_at DESC, products.id DESC").Offset(input.Offset).Limit(input.Limit).Find(&products).Error
	return products, total, err
}

// SearchPublicCompact loads only the relations needed to render a product
// card. The caller supplies a lookahead limit instead of paying for a full
// COUNT query and the detail-only product relations.
func (r *ProductRepository) SearchPublicCompact(input ProductSearchQuery) ([]product.Product, error) {
	var products []product.Product

	query := r.db.Model(&product.Product{}).
		Preload("Brand").
		Preload("Media", func(db *gorm.DB) *gorm.DB {
			return orderProductMedia(db).
				Where("product_media.media_type = ? AND product_media.is_visible = ?", "image", true)
		}).
		Preload("Variants", func(db *gorm.DB) *gorm.DB {
			return orderProductVariants(db).Where("product_variants.is_active = ?", true)
		}).
		Where(activeVariantExistsSQL("pv_public_compact"))

	var err error
	query, err = applyPublicProductSearchFilters(query, input, r.db.Dialector.Name())
	if err != nil {
		return nil, err
	}

	err = query.Distinct("products.*").
		Order("products.updated_at DESC, products.id DESC").
		Offset(input.Offset).
		Limit(input.Limit).
		Find(&products).Error
	return products, err
}

func applyPublicProductSearchFilters(query *gorm.DB, input ProductSearchQuery, dialect string) (*gorm.DB, error) {
	if input.Locale != "" {
		query = query.Where("products.locale = ?", input.Locale)
	}
	if input.Status != "" {
		query = query.Where("products.status = ?", input.Status)
	}
	if input.Featured != nil {
		query = query.Where("products.featured = ?", *input.Featured)
	}
	if input.ProductSpecificationTemplateSlug != "" {
		query = query.Joins("JOIN product_specification_templates ON product_specification_templates.id = products.product_specification_template_id AND product_specification_templates.slug = ?", input.ProductSpecificationTemplateSlug)
	}
	if input.CategorySlug != "" {
		query = query.Where(`products.product_category_id IN (
			WITH RECURSIVE category_tree AS (
				SELECT id
				FROM product_categories
				WHERE slug = ? AND is_enabled = TRUE

				UNION ALL

				SELECT child.id
				FROM product_categories child
				JOIN category_tree parent ON child.parent_id = parent.id
				WHERE child.is_enabled = TRUE
			)
			SELECT id FROM category_tree
		)`, input.CategorySlug)
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

	return applyProductSpecFilters(query, input.SpecFilters, dialect)
}

func (r *ProductRepository) FindAllWithFilters(page, pageSize int, status, locale, search, featured, customsStatus, productSpecificationTemplateID string) ([]product.Product, int64, error) {
	var products []product.Product
	var total int64

	query := r.preloadProductCategory(r.db.Model(&product.Product{}).Preload("Brand").Preload("ProductSpecificationTemplate").Preload("CustomsClassificationProfile").Preload("Media", func(db *gorm.DB) *gorm.DB {
		return orderProductMedia(db)
	}).Preload("Variants", func(db *gorm.DB) *gorm.DB {
		return orderProductVariants(db)
	}).Preload("AfterSalesTemplate").Preload("PackagingTemplate"))

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if locale != "" {
		query = query.Where("locale = ?", locale)
	}
	if productSpecificationTemplateID != "" {
		query = query.Where("product_specification_template_id = ?", productSpecificationTemplateID)
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
	switch customsStatus {
	case "complete":
		query = query.Where("COALESCE(hs_code, '') <> '' AND COALESCE(cn_code, '') <> '' AND COALESCE(country_of_origin, '') <> '' AND COALESCE(customs_description, '') <> ''")
	case "incomplete":
		query = query.Where("COALESCE(hs_code, '') = '' OR COALESCE(cn_code, '') = '' OR COALESCE(country_of_origin, '') = '' OR COALESCE(customs_description, '') = ''")
	case "missing_hs_code":
		query = query.Where("COALESCE(hs_code, '') = ''")
	case "missing_cn_code":
		query = query.Where("COALESCE(cn_code, '') = ''")
	case "missing_country_of_origin":
		query = query.Where("COALESCE(country_of_origin, '') = ''")
	case "missing_customs_description":
		query = query.Where("COALESCE(customs_description, '') = ''")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&products).Error

	return products, total, err
}

type ProductCurrencyMismatchSample struct {
	ID       uint   `json:"id"`
	SKU      string `json:"sku"`
	Name     string `json:"name"`
	Currency string `json:"currency"`
}

type ProductVariantCurrencyMismatchSample struct {
	ID        uint   `json:"id"`
	ProductID uint   `json:"product_id"`
	SKU       string `json:"sku"`
	Title     string `json:"title"`
	Currency  string `json:"currency"`
}

func (r *ProductRepository) CountProductsWithCurrencyMismatch(expectedCurrency string) (int64, error) {
	var total int64
	err := r.db.Model(&product.Product{}).
		Where("UPPER(COALESCE(currency, '')) <> ?", strings.ToUpper(strings.TrimSpace(expectedCurrency))).
		Count(&total).Error
	return total, err
}

func (r *ProductRepository) CountProductVariantsWithCurrencyMismatch(expectedCurrency string) (int64, error) {
	var total int64
	err := r.db.Model(&product.ProductVariant{}).
		Where("UPPER(COALESCE(currency, '')) <> ?", strings.ToUpper(strings.TrimSpace(expectedCurrency))).
		Count(&total).Error
	return total, err
}

func (r *ProductRepository) ListProductsWithCurrencyMismatch(expectedCurrency string, limit int) ([]ProductCurrencyMismatchSample, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	var samples []ProductCurrencyMismatchSample
	err := r.db.Model(&product.Product{}).
		Select("id", "sku", "name", "currency").
		Where("UPPER(COALESCE(currency, '')) <> ?", strings.ToUpper(strings.TrimSpace(expectedCurrency))).
		Order("updated_at DESC").
		Limit(limit).
		Scan(&samples).Error
	return samples, err
}

func (r *ProductRepository) ListProductVariantsWithCurrencyMismatch(expectedCurrency string, limit int) ([]ProductVariantCurrencyMismatchSample, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	var samples []ProductVariantCurrencyMismatchSample
	err := r.db.Model(&product.ProductVariant{}).
		Select("id", "product_id", "sku", "title", "currency").
		Where("UPPER(COALESCE(currency, '')) <> ?", strings.ToUpper(strings.TrimSpace(expectedCurrency))).
		Order("updated_at DESC").
		Limit(limit).
		Scan(&samples).Error
	return samples, err
}

func (r *ProductRepository) GetCustomsSummary(locale string) (ProductCustomsSummary, error) {
	var summary ProductCustomsSummary
	query := r.db.Model(&product.Product{}).Select(`
		COUNT(*) AS total,
		COALESCE(SUM(CASE WHEN COALESCE(hs_code, '') <> ''
			AND COALESCE(cn_code, '') <> ''
			AND COALESCE(country_of_origin, '') <> ''
			AND COALESCE(customs_description, '') <> '' THEN 1 ELSE 0 END), 0) AS complete,
		COALESCE(SUM(CASE WHEN COALESCE(hs_code, '') = ''
			OR COALESCE(cn_code, '') = ''
			OR COALESCE(country_of_origin, '') = ''
			OR COALESCE(customs_description, '') = '' THEN 1 ELSE 0 END), 0) AS incomplete,
		COALESCE(SUM(CASE WHEN COALESCE(hs_code, '') = '' THEN 1 ELSE 0 END), 0) AS missing_hs_code,
		COALESCE(SUM(CASE WHEN COALESCE(cn_code, '') = '' THEN 1 ELSE 0 END), 0) AS missing_cn_code,
		COALESCE(SUM(CASE WHEN COALESCE(country_of_origin, '') = '' THEN 1 ELSE 0 END), 0) AS missing_country_of_origin,
		COALESCE(SUM(CASE WHEN COALESCE(customs_description, '') = '' THEN 1 ELSE 0 END), 0) AS missing_customs_description
	`)
	if strings.TrimSpace(locale) != "" {
		query = query.Where("locale = ?", strings.TrimSpace(locale))
	}
	if err := query.Scan(&summary).Error; err != nil {
		return ProductCustomsSummary{}, err
	}
	return summary, nil
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
