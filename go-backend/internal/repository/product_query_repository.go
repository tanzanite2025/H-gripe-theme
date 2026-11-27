package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"commerce-platform/internal/domain/product"

	"gorm.io/gorm"
)

type ProductSearchQuery struct {
	Locale        string
	Status        string
	Featured      *bool
	Keyword       string
	CategorySlug  string
	BrandSlug     string
	PriceMin      *int64
	PriceMax      *int64
	PriceCurrency string
	SpecFilters   map[string][]string
	Offset        int
	Limit         int
}

type ProductQuickBuyCandidateQuery struct {
	Locale             string
	ProductCategoryIDs []uint
	Keyword            string
	SpecFilters        map[string][]string
	Offset             int
	Limit              int
}

type ProductRecommendationQuery struct {
	Locale            string
	ProductCategoryID *uint
	Keyword           string
	ExcludeProductIDs []uint
	Offset            int
	Limit             int
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

func activeVariantInStockSQL(alias string) string {
	return fmt.Sprintf(`EXISTS (
		SELECT 1 FROM product_variants %s
		LEFT JOIN product_variants %s_master ON %s_master.id = %s.master_variant_id AND %s_master.deleted_at IS NULL
		WHERE %s.product_id = products.id
		  AND %s.deleted_at IS NULL
		  AND %s.is_active = TRUE
		  AND COALESCE(%s_master.stock, %s.stock) > 0
	)`, alias, alias, alias, alias, alias, alias, alias, alias, alias, alias)
}

func applyQuickBuyCandidateScope(query *gorm.DB, input ProductQuickBuyCandidateQuery) *gorm.DB {
	query = query.
		Where("products.status = ?", "active").
		Where(activeVariantExistsSQL("pv_quick_buy_candidate")).
		Where(`(
			products.fulfillment_mode = ?
			OR `+activeVariantInStockSQL("pv_quick_buy_candidate_stock")+`
		)`, product.FulfillmentModeMadeToOrder)

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
		query = query.Where(`
				LOWER(products.name) LIKE ?
				OR EXISTS (
					SELECT 1 FROM product_variants pv_quick_buy_search
					WHERE pv_quick_buy_search.product_id = products.id
					  AND pv_quick_buy_search.deleted_at IS NULL
					  AND LOWER(pv_quick_buy_search.sku) LIKE ?
				)
				OR LOWER(products.short_description) LIKE ?
				OR LOWER(products.description) LIKE ?
			`, pattern, pattern, pattern, pattern)
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

func wheelsetSpecificationJSONPathExpression(path []string, dialect string) string {
	if dialect == "postgres" {
		return fmt.Sprintf("products.wheelset_specification #>> '{%s}'", strings.Join(path, ","))
	}
	return fmt.Sprintf("json_extract(products.wheelset_specification, '$.%s')", strings.Join(path, "."))
}

func wheelsetSpecificationFilterSQLCondition(definition product.WheelsetSpecificationFilterDefinition, values []string, dialect string) (string, []interface{}) {
	valueConditions := make([]string, 0, len(values))
	args := []interface{}{product.ProductTypeWheelset}
	pathExpressions := make([]string, 0, len(definition.JSONPaths))
	for _, path := range definition.JSONPaths {
		pathExpressions = append(pathExpressions, wheelsetSpecificationJSONPathExpression(path, dialect))
	}

	for _, value := range values {
		pathConditions := make([]string, 0, len(pathExpressions))
		for _, pathExpression := range pathExpressions {
			pathConditions = append(pathConditions, pathExpression+" = ?")
			args = append(args, value)
		}
		joinOperator := " OR "
		if definition.RequiresSameValueInPaths {
			joinOperator = " AND "
		}
		valueConditions = append(valueConditions, "("+strings.Join(pathConditions, joinOperator)+")")
	}

	return "(products.product_type = ? AND products.wheelset_specification IS NOT NULL AND (" + strings.Join(valueConditions, " OR ") + "))", args
}

func wheelsetSpecificationFilterValuesFromJSON(rawJSON []byte, definition product.WheelsetSpecificationFilterDefinition) []string {
	var object map[string]interface{}
	if err := json.Unmarshal(rawJSON, &object); err != nil {
		return nil
	}

	valuesByPath := make([]string, 0, len(definition.JSONPaths))
	for _, path := range definition.JSONPaths {
		var current interface{} = object
		for _, segment := range path {
			properties, ok := current.(map[string]interface{})
			if !ok {
				current = nil
				break
			}
			current = properties[segment]
		}
		value, ok := current.(string)
		if !ok || strings.TrimSpace(value) == "" {
			valuesByPath = append(valuesByPath, "")
			continue
		}
		valuesByPath = append(valuesByPath, strings.TrimSpace(value))
	}

	if definition.RequiresSameValueInPaths {
		if len(valuesByPath) == 0 || valuesByPath[0] == "" {
			return nil
		}
		for _, value := range valuesByPath[1:] {
			if value != valuesByPath[0] {
				return nil
			}
		}
		return []string{valuesByPath[0]}
	}

	result := make([]string, 0, len(valuesByPath))
	for _, value := range valuesByPath {
		if value != "" {
			result = append(result, value)
		}
	}
	return normalizeSpecFilterValues(result)
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

	definitionsBySlug := make(map[string]product.WheelsetSpecificationFilterDefinition)
	for _, definition := range product.WheelsetProductLevelSpecificationFilterDefinitions() {
		definitionsBySlug[definition.Slug] = definition
	}
	for _, slug := range slugs {
		values := valuesBySlug[slug]
		definition, known := definitionsBySlug[slug]
		if !known {
			query = query.Where("1 = 0")
			continue
		}
		condition, args := wheelsetSpecificationFilterSQLCondition(definition, values, dialect)
		query = query.Where(condition, args...)
	}
	return query, nil
}

func (r *ProductRepository) List(locale, status string, featured bool, offset, limit int) ([]product.Product, int64, error) {
	var products []product.Product
	var total int64

	query := r.preloadProductCategory(r.db.Model(&product.Product{}).
		Preload("Brand").
		Preload("Media", orderProductMedia).
		Preload("Variants", orderProductVariants))
	query = r.preloadSpokeRepairKitModels(query).
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
	if err != nil {
		return nil, 0, err
	}
	if err := r.attachProductDisplayPriceSnapshots(products); err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

// FindPublished returns the complete set of storefront products that can be
// exposed publicly.  Sitemap generation intentionally uses a lightweight
// projection (no relation preloads) and does not impose the paged-list limit
// used by the admin/public API endpoints.
func (r *ProductRepository) FindPublished() ([]product.Product, error) {
	return r.FindPublishedByLocale("")
}

// FindPublishedByLocale returns storefront products for one locale.  An empty
// locale returns all locales.  Products without an active variant are not
// routable by the public storefront and are therefore excluded.
func (r *ProductRepository) FindPublishedByLocale(locale string) ([]product.Product, error) {
	var products []product.Product
	query := r.db.Model(&product.Product{}).
		Where("products.status = ?", "active").
		Where(activeVariantExistsSQL("pv_sitemap")).
		Where("TRIM(products.slug) <> ''").
		Where("TRIM(products.locale) <> ''")
	if strings.TrimSpace(locale) != "" {
		query = query.Where("products.locale = ?", strings.TrimSpace(locale))
	}
	if err := query.Order("products.locale ASC").Order("products.id ASC").Find(&products).Error; err != nil {
		return nil, err
	}
	if err := r.attachProductDisplayPriceSnapshots(products); err != nil {
		return nil, err
	}
	return products, nil
}

// ListPublicAvailable returns active products with at least one active variant
// that can currently be purchased.
func (r *ProductRepository) ListPublicAvailable(locale string, offset, limit int) ([]product.Product, int64, error) {
	var products []product.Product
	var total int64

	query := r.preloadProductCategory(r.db.Model(&product.Product{}).
		Preload("Brand").
		Preload("Media", func(db *gorm.DB) *gorm.DB {
			return orderProductMedia(db)
		}).
		Preload("Variants", func(db *gorm.DB) *gorm.DB {
			return orderProductVariants(db)
		}))
	query = r.preloadSpokeRepairKitModels(query).
		Preload("AfterSalesTemplate").
		Preload("PackagingTemplate").
		Where("products.status = ?", "active").
		Where(activeVariantExistsSQL("pv_recommendation")).
		Where(`(
			products.fulfillment_mode = ?
			OR `+activeVariantInStockSQL("pv_recommendation_stock")+`
		)`, product.FulfillmentModeMadeToOrder)

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
	if err != nil {
		return nil, 0, err
	}
	if err := r.attachProductDisplayPriceSnapshots(products); err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

func (r *ProductRepository) ListRecommendationCandidates(input ProductRecommendationQuery) ([]product.Product, int64, error) {
	var products []product.Product
	var total int64

	query := r.preloadProductCategory(r.db.Model(&product.Product{}).
		Preload("Brand").
		Preload("Media", orderProductMedia).
		Preload("Variants", orderProductVariants))
	query = r.preloadSpokeRepairKitModels(query).
		Preload("AfterSalesTemplate").
		Preload("PackagingTemplate").
		Where("products.status = ?", "active").
		Where(activeVariantExistsSQL("pv_recommendation_candidate")).
		Where(`(
			products.fulfillment_mode = ?
			OR `+activeVariantInStockSQL("pv_recommendation_candidate_stock")+`
		)`, product.FulfillmentModeMadeToOrder)

	if input.Locale != "" {
		query = query.Where("products.locale = ?", input.Locale)
	}
	if input.ProductCategoryID != nil && *input.ProductCategoryID > 0 {
		query = query.Where(`products.product_category_id IN (
			WITH RECURSIVE recommendation_category_tree(id) AS (
				SELECT id FROM product_categories WHERE id = ? AND is_enabled = TRUE

				UNION ALL

				SELECT child.id
				FROM product_categories child
				JOIN recommendation_category_tree parent ON child.parent_id = parent.id
				WHERE child.is_enabled = TRUE
			)
			SELECT id FROM recommendation_category_tree
		)`, *input.ProductCategoryID)
	}
	if len(input.ExcludeProductIDs) > 0 {
		query = query.Where("products.id NOT IN ?", input.ExcludeProductIDs)
	}
	if input.Keyword != "" {
		pattern := "%" + strings.ToLower(input.Keyword) + "%"
		query = query.Where(`
				LOWER(products.name) LIKE ?
				OR EXISTS (
					SELECT 1 FROM product_variants pv_recommendation_search
					WHERE pv_recommendation_search.product_id = products.id
					  AND pv_recommendation_search.deleted_at IS NULL
					  AND LOWER(pv_recommendation_search.sku) LIKE ?
				)
				OR LOWER(products.short_description) LIKE ?
				OR LOWER(products.description) LIKE ?
			`, pattern, pattern, pattern, pattern)
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
	if err != nil {
		return nil, 0, err
	}
	if err := r.attachProductDisplayPriceSnapshots(products); err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

func (r *ProductRepository) ListQuickBuyCandidates(input ProductQuickBuyCandidateQuery) ([]product.Product, int64, error) {
	products := []product.Product{}
	var total int64

	query := r.db.Model(&product.Product{}).
		Preload("Brand").
		Preload("Media", orderProductMedia).
		Preload("Variants", orderProductVariants)
	query = r.preloadSpokeRepairKitModels(query).
		Preload("AfterSalesTemplate").
		Preload("PackagingTemplate")

	query = applyQuickBuyCandidateScope(query, input)
	query = r.excludeSpokeRepairKitProductsFromQuickBuy(query)
	var err error
	query, err = applyProductSpecFilters(query, input.SpecFilters, r.db.Name())
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
	if err != nil {
		return nil, 0, err
	}
	if err := r.attachProductDisplayPriceSnapshots(products); err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

// excludeSpokeRepairKitProductsFromQuickBuy keeps the repair-kit purchase
// contract on the product detail page, where the buyer must choose one
// compatible wheelset model. The table-existence guard keeps lightweight
// repository tests and pre-migration read paths compatible with older schemas.
func (r *ProductRepository) excludeSpokeRepairKitProductsFromQuickBuy(query *gorm.DB) *gorm.DB {
	if r == nil || r.db == nil || query == nil || !r.db.Migrator().HasTable(&product.ProductCategory{}) {
		return query
	}
	return query.Where(`(
		products.product_category_id IS NULL
		OR products.product_category_id NOT IN (
		WITH RECURSIVE spoke_repair_kit_category_tree(id) AS (
			SELECT id
			FROM product_categories
			WHERE slug = ?

			UNION ALL

			SELECT child.id
			FROM product_categories child
			JOIN spoke_repair_kit_category_tree parent ON child.parent_id = parent.id
		)
		SELECT id FROM spoke_repair_kit_category_tree
		)
	)`, product.SpokeRepairKitProductCategorySlug)
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
		if err := r.addWheelsetSpecificationFilterValues(productIDs, []string{slug}, valuesBySlug); err != nil {
			return nil, err
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

func (r *ProductRepository) ListFilterableSpecificationsWithDynamicValuesForCategory(categorySlug string) ([]ProductCategoryFilterableSpecification, error) {
	categorySlug = strings.TrimSpace(categorySlug)
	if categorySlug == "" {
		return []ProductCategoryFilterableSpecification{}, nil
	}
	wheelsetCategory := strings.EqualFold(categorySlug, product.ProductTypeWheelset)
	if !wheelsetCategory {
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
	filterValueProductIDs := []uint{}
	if len(productIDs) > 0 {
		if err := r.db.Table("products").
			Where("id IN ?", productIDs).
			Where("product_type = ?", product.ProductTypeWheelset).
			Pluck("id", &filterValueProductIDs).Error; err != nil {
			return nil, err
		}
	}
	definitions := product.WheelsetProductLevelSpecificationFilterDefinitions()
	result := make([]ProductCategoryFilterableSpecification, 0, len(definitions))
	valuesBySlug := make(map[string]map[string]struct{}, len(definitions))
	filterSlugs := make([]string, 0, len(definitions))
	for _, definition := range definitions {
		filterSlugs = append(filterSlugs, definition.Slug)
		valuesBySlug[definition.Slug] = make(map[string]struct{})
		result = append(result, ProductCategoryFilterableSpecification{
			Slug:      definition.Slug,
			Name:      definition.Name,
			FieldType: definition.FieldType,
			Unit:      definition.Unit,
			Options:   []string{},
		})
	}
	if err := r.addWheelsetSpecificationFilterValues(filterValueProductIDs, filterSlugs, valuesBySlug); err != nil {
		return nil, err
	}

	filtered := make([]ProductCategoryFilterableSpecification, 0, len(result))
	for _, item := range result {
		valuesSet := valuesBySlug[item.Slug]
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

func (r *ProductRepository) addWheelsetSpecificationFilterValues(productIDs []uint, slugs []string, valuesBySlug map[string]map[string]struct{}) error {
	if len(productIDs) == 0 || len(slugs) == 0 {
		return nil
	}
	requestedSlugs := make(map[string]struct{}, len(slugs))
	for _, slug := range slugs {
		requestedSlugs[strings.TrimSpace(slug)] = struct{}{}
	}
	definitionsBySlug := make(map[string]product.WheelsetSpecificationFilterDefinition, len(slugs))
	for _, item := range product.WheelsetProductLevelSpecificationFilterDefinitions() {
		if _, requested := requestedSlugs[item.Slug]; requested {
			definitionsBySlug[item.Slug] = item
		}
	}
	if len(definitionsBySlug) == 0 {
		return nil
	}

	var specifications [][]byte
	if err := r.db.Table("products").
		Where("id IN ?", productIDs).
		Where("product_type = ?", product.ProductTypeWheelset).
		Where("wheelset_specification IS NOT NULL").
		Pluck("wheelset_specification", &specifications).Error; err != nil {
		return err
	}
	for _, rawJSON := range specifications {
		for slug, definition := range definitionsBySlug {
			values := valuesBySlug[slug]
			if values == nil {
				continue
			}
			for _, value := range wheelsetSpecificationFilterValuesFromJSON(rawJSON, definition) {
				values[value] = struct{}{}
			}
		}
	}
	return nil
}

func (r *ProductRepository) quickBuyCandidateProductIDs(input ProductQuickBuyCandidateQuery) ([]uint, error) {
	query := applyQuickBuyCandidateScope(r.db.Model(&product.Product{}), input)
	query = r.excludeSpokeRepairKitProductsFromQuickBuy(query)
	filteredQuery, err := applyProductSpecFilters(query, input.SpecFilters, r.db.Name())
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

	query := r.preloadProductCategory(r.db.Model(&product.Product{}).
		Preload("Brand").
		Preload("CustomsClassificationProfile").
		Preload("Media", orderProductMedia).
		Preload("Variants", orderProductVariants))
	query = r.preloadSpokeRepairKitModels(query).
		Preload("AfterSalesTemplate").
		Preload("PackagingTemplate").
		Where(activeVariantExistsSQL("pv_public"))

	var err error
	query, err = applyPublicProductSearchFilters(query, input, r.db.Name())
	if err != nil {
		return nil, 0, err
	}

	if err := query.Distinct("products.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err = query.Distinct("products.*").Order("products.updated_at DESC, products.id DESC").Offset(input.Offset).Limit(input.Limit).Find(&products).Error
	if err != nil {
		return nil, 0, err
	}
	if err := r.attachProductDisplayPriceSnapshots(products); err != nil {
		return nil, 0, err
	}
	return products, total, nil
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
	query = r.preloadProductCategory(query)
	query = r.preloadSpokeRepairKitModels(query)

	var err error
	query, err = applyPublicProductSearchFilters(query, input, r.db.Name())
	if err != nil {
		return nil, err
	}

	err = query.Distinct("products.*").
		Order("products.updated_at DESC, products.id DESC").
		Offset(input.Offset).
		Limit(input.Limit).
		Find(&products).Error
	if err != nil {
		return nil, err
	}
	if err := r.attachProductDisplayPriceSnapshots(products); err != nil {
		return nil, err
	}
	return products, nil
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
			AND UPPER(COALESCE(pv_price_min.currency, 'USD')) = UPPER(?)
			AND pv_price_min.price_minor >= ?
		)`, input.PriceCurrency, *input.PriceMin)
	}
	if input.PriceMax != nil {
		query = query.Where(`EXISTS (
			SELECT 1 FROM product_variants pv_price_max
			WHERE pv_price_max.product_id = products.id
			  AND pv_price_max.deleted_at IS NULL
			  AND pv_price_max.is_active = TRUE
			AND UPPER(COALESCE(pv_price_max.currency, 'USD')) = UPPER(?)
			AND pv_price_max.price_minor <= ?
		)`, input.PriceCurrency, *input.PriceMax)
	}
	if input.Keyword != "" {
		pattern := "%" + strings.ToLower(input.Keyword) + "%"
		query = query.Where(`
			LOWER(products.name) LIKE ?
			OR EXISTS (
				SELECT 1 FROM product_variants pv_search
				WHERE pv_search.product_id = products.id
				  AND pv_search.deleted_at IS NULL
				  AND LOWER(pv_search.sku) LIKE ?
			)
			OR LOWER(products.short_description) LIKE ?
			OR LOWER(products.description) LIKE ?
		`, pattern, pattern, pattern, pattern)
	}

	return applyProductSpecFilters(query, input.SpecFilters, dialect)
}

func (r *ProductRepository) FindAllWithFilters(page, pageSize int, status, locale, search, featured, customsStatus string) ([]product.Product, int64, error) {
	var products []product.Product
	var total int64

	query := r.preloadProductCategory(r.db.Model(&product.Product{}).Preload("Brand").Preload("CustomsClassificationProfile").Preload("Media", orderProductMedia).Preload("Variants", orderProductVariants).Preload("AfterSalesTemplate").Preload("PackagingTemplate"))
	query = r.preloadSpokeRepairKitModels(query)

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if locale != "" {
		query = query.Where("locale = ?", locale)
	}
	if search != "" {
		pattern := "%" + search + "%"
		query = query.Where(`
			name LIKE ?
			OR description LIKE ?
			OR EXISTS (
				SELECT 1 FROM product_variants pv_admin_search
				WHERE pv_admin_search.product_id = products.id
				  AND pv_admin_search.deleted_at IS NULL
				  AND pv_admin_search.sku LIKE ?
			)`, pattern, pattern, pattern)
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
	if err != nil {
		return nil, 0, err
	}
	if err := r.attachProductDisplayPriceSnapshots(products); err != nil {
		return nil, 0, err
	}
	return products, total, nil
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
		Select(`products.id, products.name, products.currency,
			COALESCE((SELECT pv.sku FROM product_variants pv
				WHERE pv.product_id = products.id AND pv.deleted_at IS NULL
				ORDER BY pv.is_default DESC, pv.sort_order ASC, pv.id ASC LIMIT 1), '') AS sku`).
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
