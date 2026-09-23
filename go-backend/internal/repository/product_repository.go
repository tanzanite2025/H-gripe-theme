package repository

import (
	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/product"
	"context"
	"fmt"
	"sort"

	"gorm.io/datatypes"
	"gorm.io/gorm"
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

func orderProductMedia(db *gorm.DB) *gorm.DB {
	return db.Order("product_media.sort_order ASC, product_media.id ASC")
}

func orderSpecDefinitions(db *gorm.DB) *gorm.DB {
	return db.Order("product_spec_definitions.sort_order ASC, product_spec_definitions.id ASC")
}

func preloadProductSpecDefinitionOptionItems(db *gorm.DB) *gorm.DB {
	return db.Preload("ProductSpecificationTemplate.SpecDefinitions.OptionItems", func(db *gorm.DB) *gorm.DB {
		return db.Order("product_spec_option_items.sort_order ASC, product_spec_option_items.id ASC")
	})
}

func preloadSpecDefinitionOptionItems(db *gorm.DB) *gorm.DB {
	return db.Preload("SpecDefinitions.OptionItems", func(db *gorm.DB) *gorm.DB {
		return db.Order("product_spec_option_items.sort_order ASC, product_spec_option_items.id ASC")
	})
}

func orderProductVariants(db *gorm.DB) *gorm.DB {
	return db.Order("product_variants.sort_order ASC, product_variants.id ASC")
}

func (r *ProductRepository) preloadProductVariantOptionRules(query *gorm.DB) *gorm.DB {
	if r.db.Migrator().HasTable(&product.ProductOptionGroupVariantRule{}) {
		query = query.Preload("Variants.OptionGroupRules")
	}
	if r.db.Migrator().HasTable(&product.ProductOptionValueVariantRule{}) {
		query = query.Preload("Variants.OptionValueRules")
	}
	return query
}

func orderProductVariantOptionValues(db *gorm.DB) *gorm.DB {
	return db.Order("product_variant_option_values.sort_order ASC, product_variant_option_values.id ASC")
}

func (r *ProductRepository) preloadProductVariantOptionValues(db *gorm.DB) *gorm.DB {
	db = r.preloadProductOptionValueRelations(db)
	if r.db.Migrator().HasTable(&product.ProductVariantOptionValue{}) {
		query := db.Preload("VariantOptionValues", func(db *gorm.DB) *gorm.DB {
			return orderProductVariantOptionValues(db)
		})
		if r.db.Migrator().HasTable(&product.ProductCustomOptionPolicy{}) {
			query = query.Preload("VariantOptionValues.CustomOptionPolicy")
		}
		return query
	}
	return db
}

func (r *ProductRepository) preloadProductOptionValueRelations(db *gorm.DB) *gorm.DB {
	if !r.db.Migrator().HasTable(&product.ProductOptionValueRelation{}) {
		return db
	}
	return db.Preload("OptionValueRelations", func(db *gorm.DB) *gorm.DB {
		return db.Order("product_option_value_relations.relation_type ASC, product_option_value_relations.source_option_value_id ASC, product_option_value_relations.target_option_value_id ASC")
	})
}

func (r *ProductRepository) preloadProductCategory(db *gorm.DB) *gorm.DB {
	if r.db.Migrator().HasTable(&product.ProductCategory{}) {
		return db.Preload("ProductCategory")
	}
	return db
}

// Create 鍒涘缓浜у搧
func (r *ProductRepository) Create(p *product.Product) error {
	return r.db.Create(p).Error
}

func (r *ProductRepository) CreateWithSpecValues(p *product.Product, specValues []product.ProductSpecValue) error {
	return r.CreateWithSpecValuesAndVariants(p, specValues, nil)
}

func (r *ProductRepository) CreateWithSpecValuesAndVariants(p *product.Product, specValues []product.ProductSpecValue, variants []product.ProductVariant) error {
	return r.CreateWithSpecValuesVariantsOptionValuesAndMedia(p, specValues, variants, nil, nil)
}

func (r *ProductRepository) CreateWithSpecValuesVariantsAndMedia(p *product.Product, specValues []product.ProductSpecValue, variants []product.ProductVariant, mediaItems []product.ProductMedia) error {
	return r.CreateWithSpecValuesVariantsOptionValuesAndMedia(p, specValues, variants, nil, mediaItems)
}

func (r *ProductRepository) CreateWithSpecValuesVariantsOptionValuesAndMedia(
	p *product.Product,
	specValues []product.ProductSpecValue,
	variants []product.ProductVariant,
	optionValues []product.ProductVariantOptionValue,
	mediaItems []product.ProductMedia,
) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
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

		variantRules := make([]struct {
			groups []product.ProductOptionGroupVariantRule
			values []product.ProductOptionValueVariantRule
		}, len(variants))
		if len(variants) > 0 {
			requestedActive := make([]bool, len(variants))
			for i := range variants {
				variants[i].ProductID = p.ID
				requestedActive[i] = variants[i].IsActive
				variantRules[i].groups = variants[i].OptionGroupRules
				variantRules[i].values = variants[i].OptionValueRules
				variants[i].OptionGroupRules = nil
				variants[i].OptionValueRules = nil
			}
			if err := tx.Omit("DisplayPriceData").Create(&variants).Error; err != nil {
				return err
			}
			for i := range variants {
				if requestedActive[i] {
					continue
				}
				if err := tx.Model(&variants[i]).Update("is_active", false).Error; err != nil {
					return err
				}
				variants[i].IsActive = false
			}
		}

		if len(optionValues) > 0 {
			if err := replaceProductVariantOptionValues(tx, p.ID, optionValues); err != nil {
				return err
			}
		}
		if p.OptionValueRelationsDirty {
			if err := replaceProductOptionValueRelations(tx, p.ID, p.OptionValueRelations); err != nil {
				return err
			}
		}
		for i := range variants {
			if err := replaceVariantOptionRules(tx, variants[i].ID, variantRules[i].groups, variantRules[i].values); err != nil {
				return err
			}
		}

		if len(mediaItems) > 0 {
			if err := ensureProductMediaReferencesBelongToProduct(tx, p.ID, mediaItems); err != nil {
				return err
			}
			for i := range mediaItems {
				mediaItems[i].ProductID = p.ID
			}
			if err := tx.Create(&mediaItems).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// FindByID 鏍规嵁ID鏌ユ壘浜у搧
func (r *ProductRepository) FindByID(id uint) (*product.Product, error) {
	return r.FindByIDContext(context.Background(), id)
}

func (r *ProductRepository) FindByIDContext(ctx context.Context, id uint) (*product.Product, error) {
	var p product.Product
	query := r.preloadProductCategory(r.db.WithContext(ctx).Preload("Brand")).Preload("Media", func(db *gorm.DB) *gorm.DB {
		return orderProductMedia(db)
	}).Preload("ProductSpecificationTemplate.SpecDefinitions", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
	}).Preload("SpecValues.SpecDefinition", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
	}).Preload("Variants", func(db *gorm.DB) *gorm.DB {
		return orderProductVariants(db)
	})
	if r.db.Migrator().HasTable(&product.ProductSpecOptionItem{}) {
		query = preloadProductSpecDefinitionOptionItems(query)
	}
	query = r.preloadProductVariantOptionRules(query)
	query = r.preloadProductVariantOptionValues(query)
	err := query.Preload("AfterSalesTemplate").Preload("PackagingTemplate").Preload("CustomsClassificationProfile").First(&p, id).Error
	if err != nil {
		return nil, err
	}
	if err := r.attachProductDisplayPriceSnapshot(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

// FindBySlug finds a product by slug. When locale is empty, it treats products
// as a unified storefront catalog item instead of a translated content row.
func (r *ProductRepository) FindBySlug(slug, locale string) (*product.Product, error) {
	return r.FindBySlugContext(context.Background(), slug, locale)
}

func (r *ProductRepository) FindBySlugContext(ctx context.Context, slug, locale string) (*product.Product, error) {
	var p product.Product
	query := r.preloadProductCategory(r.db.WithContext(ctx).Preload("Brand")).Preload("Media", func(db *gorm.DB) *gorm.DB {
		return orderProductMedia(db)
	}).Preload("ProductSpecificationTemplate.SpecDefinitions", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
	}).Preload("SpecValues.SpecDefinition", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
	}).Preload("Variants", func(db *gorm.DB) *gorm.DB {
		return orderProductVariants(db)
	})
	if r.db.Migrator().HasTable(&product.ProductSpecOptionItem{}) {
		query = preloadProductSpecDefinitionOptionItems(query)
	}
	query = r.preloadProductVariantOptionRules(query)
	query = r.preloadProductVariantOptionValues(query).Preload("AfterSalesTemplate").Preload("PackagingTemplate").Preload("CustomsClassificationProfile").Where("slug = ?", slug)

	if locale != "" {
		query = query.Where("locale = ?", locale)
	}

	err := query.First(&p).Error
	if err != nil {
		return nil, err
	}
	if err := r.attachProductDisplayPriceSnapshot(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

// FindBySKU 鏍规嵁SKU鏌ユ壘浜у搧
func (r *ProductRepository) FindBySKU(sku string) (*product.Product, error) {
	var p product.Product
	query := r.db.Preload("Brand").Preload("Media", func(db *gorm.DB) *gorm.DB { return orderProductMedia(db) }).
		Preload("Variants", func(db *gorm.DB) *gorm.DB { return orderProductVariants(db) })
	query = r.preloadProductVariantOptionValues(query)
	err := query.Preload("AfterSalesTemplate").Preload("PackagingTemplate").Preload("CustomsClassificationProfile").
		Joins("JOIN product_variants product_sku_lookup ON product_sku_lookup.product_id = products.id AND product_sku_lookup.deleted_at IS NULL AND product_sku_lookup.sku = ?", sku).
		First(&p).Error
	if err != nil {
		return nil, err
	}
	if err := r.attachProductDisplayPriceSnapshot(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepository) FindProductsByIDs(ids []uint) ([]product.Product, error) {
	var products []product.Product
	if len(ids) == 0 {
		return products, nil
	}
	err := r.db.Preload("Brand").Preload("Variants", func(db *gorm.DB) *gorm.DB {
		return orderProductVariants(db)
	}).Where("id IN ?", ids).Find(&products).Error
	if err != nil {
		return nil, err
	}
	if err := r.attachProductDisplayPriceSnapshots(products); err != nil {
		return nil, err
	}
	return products, nil
}

// FindProductsByIDsForCustomerContext is a deliberately narrow projection
// reader. It loads only public product fields needed for customer context
// cards and does not expose supplier cost, inventory internals, or product
// attributes.
func (r *ProductRepository) FindProductsByIDsForCustomerContext(ids []uint) ([]product.Product, error) {
	var products []product.Product
	if len(ids) == 0 {
		return products, nil
	}
	query := r.db.Select("id", "name", "currency", "price_minor", "sale_price_minor").
		Preload("Media", func(db *gorm.DB) *gorm.DB {
			return orderProductMedia(db)
		}).
		Preload("Variants", func(db *gorm.DB) *gorm.DB {
			return orderProductVariants(db)
		}).Where("id IN ?", ids)
	if err := query.Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// ListProductsForDisplayPriceRefresh loads only the source price fields needed
// to rebuild customer-facing display price snapshots.
func (r *ProductRepository) ListProductsForDisplayPriceRefresh() ([]product.Product, error) {
	var products []product.Product
	query := r.db.
		Select("id", "currency", "price_minor", "sale_price_minor").
		Preload("Variants", func(db *gorm.DB) *gorm.DB {
			return db.
				// StartingPriceVariant is part of the snapshot source fingerprint.
				// Keep the fields it uses in this lightweight refresh projection so
				// read-model hydration selects the same source as the refresh job.
				Select("id", "product_id", "currency", "price_minor", "sale_price_minor", "is_active", "is_default", "sort_order").
				Order("sort_order ASC, id ASC")
		})
	if err := query.Find(&products).Error; err != nil {
		return nil, err
	}
	if err := r.attachProductDisplayPriceSnapshots(products); err != nil {
		return nil, err
	}
	return products, nil
}

type ProductVariantDisplayPriceSnapshotUpdate struct {
	VariantID            uint
	SourceCurrency       string
	SourcePriceMinor     int64
	SourceSalePriceMinor *int64
	DisplayPriceData     datatypes.JSON
}

type ProductDisplayPriceSnapshotUpdate struct {
	ProductID              uint
	SourceCurrency         string
	SourcePriceMinor       int64
	SourceSalePriceMinor   *int64
	UpdateProduct          bool
	DisplayPriceData       datatypes.JSON
	VariantSnapshotUpdates []ProductVariantDisplayPriceSnapshotUpdate
}

// UpdateDisplayPriceSnapshots updates only product/SKU display snapshot JSON.
// Source currencies and source amounts are deliberately excluded from this
// mutation.
func (r *ProductRepository) UpdateDisplayPriceSnapshots(updates []ProductDisplayPriceSnapshotUpdate) error {
	if len(updates) == 0 {
		return nil
	}
	const batchSize = 500
	for start := 0; start < len(updates); start += batchSize {
		end := start + batchSize
		if end > len(updates) {
			end = len(updates)
		}
		batch := updates[start:end]
		if err := r.db.Transaction(func(tx *gorm.DB) error {
			for _, update := range batch {
				if update.ProductID == 0 {
					continue
				}
				if update.UpdateProduct {
					if err := upsertProductDisplayPriceSnapshot(tx, product.ProductDisplayPriceSnapshot{
						ScopeKey:             fmt.Sprintf("product:%d", update.ProductID),
						ProductID:            update.ProductID,
						SourceCurrency:       update.SourceCurrency,
						SourcePriceMinor:     update.SourcePriceMinor,
						SourceSalePriceMinor: update.SourceSalePriceMinor,
						DisplayPriceData:     update.DisplayPriceData,
					}, update.SourcePriceMinor, update.SourceCurrency, update.SourceSalePriceMinor, nil); err != nil {
						return err
					}
				}
				for _, variantUpdate := range update.VariantSnapshotUpdates {
					if variantUpdate.VariantID == 0 {
						continue
					}
					if err := upsertProductDisplayPriceSnapshot(tx, product.ProductDisplayPriceSnapshot{
						ScopeKey:             fmt.Sprintf("variant:%d", variantUpdate.VariantID),
						ProductID:            update.ProductID,
						VariantID:            &variantUpdate.VariantID,
						SourceCurrency:       variantUpdate.SourceCurrency,
						SourcePriceMinor:     variantUpdate.SourcePriceMinor,
						SourceSalePriceMinor: variantUpdate.SourceSalePriceMinor,
						DisplayPriceData:     variantUpdate.DisplayPriceData,
					}, variantUpdate.SourcePriceMinor, variantUpdate.SourceCurrency, variantUpdate.SourceSalePriceMinor, &variantUpdate.VariantID); err != nil {
						return err
					}
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func upsertProductDisplayPriceSnapshot(tx *gorm.DB, snapshot product.ProductDisplayPriceSnapshot, sourceMinor int64, sourceCurrency string, sourceSaleMinor *int64, variantID *uint) error {
	var source struct {
		PriceMinor     int64
		SalePriceMinor *int64
		Currency       string
	}
	query := tx.Model(&product.Product{})
	if variantID == nil {
		if err := query.Select("price_minor, sale_price_minor, currency").Where("id = ?", snapshot.ProductID).Scan(&source).Error; err != nil {
			return err
		}
		// Product display snapshots use the same starting active variant as the
		// refresh service. Validate that source fingerprint before persisting.
		var variants []product.ProductVariant
		if err := tx.Where("product_id = ? AND deleted_at IS NULL", snapshot.ProductID).
			Select("id, product_id, currency, price_minor, sale_price_minor, is_active, is_default, sort_order").
			Order("sort_order ASC, id ASC").Find(&variants).Error; err != nil {
			return err
		}
		item := product.Product{ID: snapshot.ProductID, Currency: source.Currency, PriceMinor: source.PriceMinor, SalePriceMinor: source.SalePriceMinor, Variants: variants}
		if starting := item.StartingPriceVariant(); starting != nil {
			source.PriceMinor, source.SalePriceMinor, source.Currency = starting.PriceMinor, starting.SalePriceMinor, starting.Currency
		}
	} else {
		if err := tx.Model(&product.ProductVariant{}).Select("price_minor, sale_price_minor, currency").Where("id = ? AND product_id = ?", *variantID, snapshot.ProductID).Scan(&source).Error; err != nil {
			return err
		}
	}
	if source.PriceMinor != sourceMinor || currency.NormalizeCode(source.Currency) != currency.NormalizeCode(sourceCurrency) || !salePriceMinorEqual(source.SalePriceMinor, sourceSaleMinor) {
		return nil
	}
	return tx.Where("scope_key = ?", snapshot.ScopeKey).Assign(snapshot).FirstOrCreate(&snapshot).Error
}

func salePriceMinorEqual(left, right *int64) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func (r *ProductRepository) FindProductCacheIdentitiesByIDs(ids []uint) ([]product.Product, error) {
	var products []product.Product
	if len(ids) == 0 {
		return products, nil
	}
	err := r.db.Select("id", "slug", "locale").Where("id IN ?", ids).Find(&products).Error
	return products, err
}

func (r *ProductRepository) FindProductCacheIdentitiesByProductSpecificationTemplateID(productSpecificationTemplateID uint) ([]product.Product, error) {
	var products []product.Product
	if productSpecificationTemplateID == 0 {
		return products, nil
	}
	err := r.db.Select("id", "slug", "locale").Where("product_specification_template_id = ?", productSpecificationTemplateID).Find(&products).Error
	return products, err
}

func (r *ProductRepository) FindProductCacheIdentitiesByInformationTemplateID(templateID uint) ([]product.Product, error) {
	var products []product.Product
	if templateID == 0 {
		return products, nil
	}
	err := r.db.Select("id", "slug", "locale").
		Where("after_sales_template_id = ? OR packaging_template_id = ?", templateID, templateID).
		Find(&products).Error
	return products, err
}

func (r *ProductRepository) FindProductCacheIdentitiesByBrandID(brandID uint) ([]product.Product, error) {
	var products []product.Product
	if brandID == 0 {
		return products, nil
	}
	err := r.db.Select("id", "slug", "locale").
		Where("brand_id = ?", brandID).
		Find(&products).Error
	return products, err
}

func (r *ProductRepository) FindProductCacheIdentitiesByBrandIDPage(brandID, afterID uint, limit int) ([]product.Product, error) {
	var products []product.Product
	if brandID == 0 {
		return products, nil
	}
	if limit <= 0 || limit > 1000 {
		limit = 500
	}
	query := r.db.Select("id", "slug", "locale").
		Where("brand_id = ?", brandID).
		Order("id ASC").
		Limit(limit)
	if afterID > 0 {
		query = query.Where("id > ?", afterID)
	}
	err := query.Find(&products).Error
	return products, err
}

func (r *ProductRepository) FindProductSyncIdentitiesByBrandID(brandID uint) ([]product.Product, error) {
	var products []product.Product
	if brandID == 0 {
		return products, nil
	}
	err := r.db.Select("id", "status").
		Where("brand_id = ?", brandID).
		Find(&products).Error
	return products, err
}

func (r *ProductRepository) FindProductSyncIdentitiesByBrandIDPage(brandID, afterID uint, limit int) ([]product.Product, error) {
	var products []product.Product
	if brandID == 0 {
		return products, nil
	}
	if limit <= 0 || limit > 1000 {
		limit = 500
	}
	query := r.db.Select("id", "status").
		Where("brand_id = ?", brandID).
		Order("id ASC").
		Limit(limit)
	if afterID > 0 {
		query = query.Where("id > ?", afterID)
	}
	err := query.Find(&products).Error
	return products, err
}

// Update 鏇存柊浜у搧
func (r *ProductRepository) Update(p *product.Product) error {
	if p == nil || p.ID == 0 {
		return gorm.ErrInvalidData
	}
	return r.db.Model(&product.Product{}).Where("id = ?", p.ID).Updates(productUpdateColumns(p)).Error
}

// UpdateSEO mutates only SEO columns. SEO writes must never persist a stale
// product snapshot (especially stock, price, or view_count).
func (r *ProductRepository) UpdateSEO(id uint, metaTitle, metaDescription string) error {
	if id == 0 {
		return gorm.ErrInvalidData
	}
	return r.db.Model(&product.Product{}).Where("id = ?", id).Updates(map[string]interface{}{
		"meta_title": metaTitle,
		"meta_desc":  metaDescription,
	}).Error
}

func productUpdateColumns(p *product.Product) map[string]interface{} {
	columns := map[string]interface{}{
		"product_specification_template_id": p.ProductSpecificationTemplateID,
		"product_category_id":               p.ProductCategoryID,
		"brand_id":                          p.BrandID,
		"shipping_template_id":              p.ShippingTemplateID,
		"after_sales_template_id":           p.AfterSalesTemplateID,
		"packaging_template_id":             p.PackagingTemplateID,
		"customs_classification_profile_id": p.CustomsClassificationProfileID,
		"hs_code":                           p.HSCode,
		"cn_code":                           p.CNCode,
		"country_of_origin":                 p.CountryOfOrigin,
		"customs_description":               p.CustomsDescription,
		"name":                              p.Name,
		"slug":                              p.Slug,
		"description":                       p.Description,
		"short_desc":                        p.ShortDesc,
		"currency":                          p.Currency,
		"price_minor":                       p.PriceMinor,
		"sale_price_minor":                  p.SalePriceMinor,
		"fulfillment_mode":                  p.FulfillmentMode,
		"status":                            p.Status,
		"locale":                            p.Locale,
		"parent_id":                         p.ParentID,
		"featured":                          p.Featured,
		"meta_title":                        p.MetaTitle,
		"meta_desc":                         p.MetaDesc,
	}
	return columns
}

func (r *ProductRepository) UpdateWithSpecValues(p *product.Product, specValues []product.ProductSpecValue, replaceSpecs bool) error {
	return r.UpdateWithSpecValuesAndVariants(p, specValues, replaceSpecs, nil, false)
}

func (r *ProductRepository) UpdateWithSpecValuesAndVariants(p *product.Product, specValues []product.ProductSpecValue, replaceSpecs bool, variants []product.ProductVariant, replaceVariants bool) error {
	return r.UpdateWithSpecValuesVariantsAndMedia(p, specValues, replaceSpecs, variants, replaceVariants, nil, false)
}

func (r *ProductRepository) UpdateWithSpecValuesVariantsAndMedia(p *product.Product, specValues []product.ProductSpecValue, replaceSpecs bool, variants []product.ProductVariant, replaceVariants bool, mediaItems []product.ProductMedia, replaceMedia bool) error {
	return r.UpdateWithSpecValuesVariantsOptionValuesAndMedia(p, specValues, replaceSpecs, variants, replaceVariants, nil, false, mediaItems, replaceMedia)
}

func (r *ProductRepository) UpdateWithSpecValuesVariantsOptionValuesAndMedia(
	p *product.Product,
	specValues []product.ProductSpecValue,
	replaceSpecs bool,
	variants []product.ProductVariant,
	replaceVariants bool,
	optionValues []product.ProductVariantOptionValue,
	replaceOptionValues bool,
	mediaItems []product.ProductMedia,
	replaceMedia bool,
) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		variantRules := make([]struct {
			groups []product.ProductOptionGroupVariantRule
			values []product.ProductOptionValueVariantRule
		}, len(variants))
		if err := tx.Model(&product.Product{}).Where("id = ?", p.ID).Updates(productUpdateColumns(p)).Error; err != nil {
			return err
		}
		if p.ProductCategoryID == nil {
			if err := tx.Model(&product.Product{}).Where("id = ?", p.ID).UpdateColumn("product_category_id", nil).Error; err != nil {
				return err
			}
		}
		if p.CustomsClassificationProfileID == nil {
			if err := tx.Model(&product.Product{}).Where("id = ?", p.ID).UpdateColumn("customs_classification_profile_id", nil).Error; err != nil {
				return err
			}
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
			for i := range variants {
				variantRules[i].groups = variants[i].OptionGroupRules
				variantRules[i].values = variants[i].OptionValueRules
				variants[i].OptionGroupRules = nil
				variants[i].OptionValueRules = nil
			}
			if err := replaceProductVariants(tx, p.ID, variants); err != nil {
				return err
			}
		}

		if replaceOptionValues {
			if err := replaceProductVariantOptionValues(tx, p.ID, optionValues); err != nil {
				return err
			}
		}
		if p.OptionValueRelationsDirty {
			if err := replaceProductOptionValueRelations(tx, p.ID, p.OptionValueRelations); err != nil {
				return err
			}
		}
		if replaceVariants {
			for i := range variants {
				if err := replaceVariantOptionRules(tx, variants[i].ID, variantRules[i].groups, variantRules[i].values); err != nil {
					return err
				}
			}
		}

		if replaceMedia {
			if err := replaceProductMedia(tx, p.ID, mediaItems); err != nil {
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
	return r.IncrementViewCountContext(context.Background(), id)
}

func (r *ProductRepository) IncrementViewCountContext(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&product.Product{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}

// IncrementViewCounts applies buffered view count deltas in one transaction.
func (r *ProductRepository) IncrementViewCounts(ctx context.Context, counts map[uint]int64) error {
	if len(counts) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ids := make([]uint, 0, len(counts))
		for id := range counts {
			ids = append(ids, id)
		}
		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
		for _, id := range ids {
			delta := counts[id]
			if delta <= 0 {
				continue
			}
			if err := tx.Model(&product.Product{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + ?", delta)).Error; err != nil {
				return err
			}
		}
		return nil
	})
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

	// Inventory statistics come from the active inventory-owning variants.
	var inventoryStats struct {
		LowStock   int64 `gorm:"column:low_stock"`
		OutOfStock int64 `gorm:"column:out_of_stock"`
	}
	if err := r.db.Raw(`
		SELECT
			COALESCE(SUM(CASE WHEN inventory.total_stock > 0 AND inventory.total_stock < ? THEN 1 ELSE 0 END), 0) AS low_stock,
			COALESCE(SUM(CASE WHEN inventory.total_stock = 0 THEN 1 ELSE 0 END), 0) AS out_of_stock
		FROM (
			SELECT
				products.id,
				COALESCE(SUM(CASE
					WHEN variants.id IS NULL THEN 0
					WHEN variants.master_variant_id IS NULL THEN variants.stock
					ELSE COALESCE(master_variants.stock, 0)
				END), 0) AS total_stock
			FROM products
			LEFT JOIN product_variants AS variants
				ON variants.product_id = products.id
				AND variants.is_active = TRUE
				AND variants.deleted_at IS NULL
			LEFT JOIN product_variants AS master_variants
				ON master_variants.id = variants.master_variant_id
				AND master_variants.deleted_at IS NULL
			WHERE products.deleted_at IS NULL
				AND COALESCE(NULLIF(products.fulfillment_mode, ''), ?) = ?
			GROUP BY products.id
		) AS inventory
	`, 10, product.FulfillmentModeStock, product.FulfillmentModeStock).Scan(&inventoryStats).Error; err != nil {
		return nil, err
	}
	stats["low_stock"] = inventoryStats.LowStock
	stats["out_of_stock"] = inventoryStats.OutOfStock

	return stats, nil
}
