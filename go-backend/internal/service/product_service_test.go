package service

import (
	"errors"
	"net"
	"strconv"
	"sync"
	"testing"

	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/pkg/cache"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/repository"

	"github.com/alicebob/miniredis/v2"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestProductServiceCreateAdminProductPersistsTemplateSpecs(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)

	createdProduct, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductTypeID: &productType.ID,
		Name:          "Carbon Rim 45",
		Slug:          "carbon-rim-45",
		Status:        "active",
		Locale:        "en",
		SpecValues: map[string]string{
			"outer_width_mm": "30.5",
			"tubeless_ready": "yes",
		},
		Variants: []ProductVariantInput{
			{
				SKU:          "RIM-001-24H-DISC",
				OptionValues: map[string]string{"brake_type": "disc"},
				Price:        399,
				Stock:        5,
				IsDefault:    true,
				IsActive:     boolPtr(true),
			},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, createdProduct)
	assert.Equal(t, productType.ID, *createdProduct.ProductTypeID)
	require.Len(t, createdProduct.SpecValues, 2)
	require.Len(t, createdProduct.Variants, 1)
	assert.Equal(t, "RIM-001-24H-DISC", createdProduct.SKU)
	assert.Equal(t, 5, createdProduct.Stock)
	assert.Equal(t, "true", findSavedSpecValue(t, createdProduct, "tubeless_ready"))
	assert.Equal(t, "30.5", findSavedSpecValue(t, createdProduct, "outer_width_mm"))
	assert.Equal(t, "RIM-001-24H-DISC", createdProduct.Variants[0].SKU)
	assert.JSONEq(t, `{"brake_type":"disc"}`, createdProduct.Variants[0].OptionValues)
}

func TestProductServiceGetByIDCoalescesConcurrentCacheMisses(t *testing.T) {
	db, productService := newTestProductService(t)
	record := product.Product{
		SKU:      "SINGLEFLIGHT-001",
		Name:     "SingleFlight Product",
		Slug:     "singleflight-product",
		Currency: "USD",
		Price:    100,
		Status:   "active",
		Locale:   "en",
	}
	require.NoError(t, db.Create(&record).Error)

	var once sync.Once
	ready := make(chan struct{})
	release := make(chan struct{})
	db.Callback().Query().Before("gorm:query").Register("test_block_product_load", func(tx *gorm.DB) {
		if tx.Statement.Table != "products" {
			return
		}
		once.Do(func() {
			close(ready)
			<-release
		})
	})

	const callers = 16
	results := make(chan *product.Product, callers)
	errs := make(chan error, callers)
	var wg sync.WaitGroup
	wg.Add(callers)
	for i := 0; i < callers; i++ {
		go func() {
			defer wg.Done()
			result, err := productService.GetByID(record.ID)
			results <- result
			errs <- err
		}()
	}

	<-ready
	close(release)
	wg.Wait()
	close(results)
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}
	for result := range results {
		require.NotNil(t, result)
		assert.Equal(t, record.ID, result.ID)
	}

	var viewCount int
	require.NoError(t, db.Model(&product.Product{}).Where("id = ?", record.ID).Pluck("view_count", &viewCount).Error)
	assert.Equal(t, 1, viewCount)
}

func TestProductServiceDistributedCacheMissesShareOneDatabaseLoad(t *testing.T) {
	db, _ := newTestProductService(t)
	record := product.Product{
		SKU:      "DISTRIBUTED-SINGLEFLIGHT-001",
		Name:     "Distributed SingleFlight Product",
		Slug:     "distributed-singleflight-product",
		Currency: "USD",
		Price:    100,
		Status:   "active",
		Locale:   "en",
	}
	require.NoError(t, db.Create(&record).Error)

	redisServer := miniredis.RunT(t)
	host, portText, err := net.SplitHostPort(redisServer.Addr())
	require.NoError(t, err)
	port, err := strconv.Atoi(portText)
	require.NoError(t, err)

	redisCacheA, err := cache.Init(config.RedisConfig{Host: host, Port: port})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = redisCacheA.Close()
	})
	redisCacheB, err := cache.Init(config.RedisConfig{Host: host, Port: port})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = redisCacheB.Close()
	})

	productServiceA := NewProductServiceWithCacheOptions(repository.NewProductRepository(db), redisCacheA, 60, 2)
	productServiceB := NewProductServiceWithCacheOptions(repository.NewProductRepository(db), redisCacheB, 60, 2)

	var once sync.Once
	ready := make(chan struct{})
	release := make(chan struct{})
	db.Callback().Query().Before("gorm:query").Register("test_block_distributed_product_load", func(tx *gorm.DB) {
		if tx.Statement.Table != "products" {
			return
		}
		once.Do(func() {
			close(ready)
			<-release
		})
	})

	type result struct {
		product *product.Product
		err     error
	}
	results := make(chan result, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	for _, service := range []*ProductService{productServiceA, productServiceB} {
		go func(service *ProductService) {
			defer wg.Done()
			loadedProduct, loadErr := service.GetByID(record.ID)
			results <- result{product: loadedProduct, err: loadErr}
		}(service)
	}

	<-ready
	close(release)
	wg.Wait()
	close(results)

	for loaded := range results {
		require.NoError(t, loaded.err)
		require.NotNil(t, loaded.product)
		assert.Equal(t, record.ID, loaded.product.ID)
	}

	var viewCount int
	require.NoError(t, db.Model(&product.Product{}).Where("id = ?", record.ID).Pluck("view_count", &viewCount).Error)
	assert.Equal(t, 1, viewCount)
}

func TestProductServiceCreateAdminProductPersistsProductScopedVisualVariantOptions(t *testing.T) {
	db, productService := newTestProductService(t)

	productType := product.ProductType{
		Name:      "Finish Product",
		Slug:      "finish_product",
		IsEnabled: true,
	}
	require.NoError(t, db.Create(&productType).Error)

	finishSpec := product.SpecDefinition{
		ProductTypeID:   productType.ID,
		Group:           "Appearance",
		Name:            "Finish",
		Slug:            "finish",
		FieldType:       "select",
		Presentation:    "color",
		IsVisible:       true,
		IsFilterable:    true,
		IsVariantOption: true,
		SortOrder:       10,
		Options:         `["template_black"]`,
	}
	require.NoError(t, db.Create(&finishSpec).Error)

	createdProduct, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductTypeID: &productType.ID,
		Name:          "Ruby Finish Product",
		Slug:          "ruby-finish-product",
		Status:        "active",
		Locale:        "en",
		VariantOptionValues: []ProductVariantOptionValueInput{
			{
				SpecDefinitionID: finishSpec.ID,
				ValueKey:         "ruby_red",
				Label:            "Ruby Red",
				ColorHex:         "#8F2028",
				SwatchURL:        "/uploads/swatches/ruby-red.webp",
				IsEnabled:        boolPtr(true),
			},
		},
		Variants: []ProductVariantInput{
			{
				SKU:          "VISUAL-RUBY-001",
				OptionValues: map[string]string{"finish": "ruby_red"},
				Price:        399,
				Stock:        5,
				IsDefault:    true,
				IsActive:     boolPtr(true),
			},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, createdProduct)
	require.Len(t, createdProduct.VariantOptionValues, 1)
	assert.Equal(t, "ruby_red", createdProduct.VariantOptionValues[0].ValueKey)
	assert.Equal(t, "Ruby Red", createdProduct.VariantOptionValues[0].Label)
	assert.Equal(t, "#8F2028", createdProduct.VariantOptionValues[0].ColorHex)
	assert.Equal(t, "/uploads/swatches/ruby-red.webp", createdProduct.VariantOptionValues[0].SwatchURL)
	require.Len(t, createdProduct.Variants, 1)
	assert.JSONEq(t, `{"finish":"ruby_red"}`, createdProduct.Variants[0].OptionValues)

	publicProducts, total, err := productService.ListPublic("en", false, 1, 20)
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, publicProducts, 1)
	require.NotNil(t, publicProducts[0].ProductType)
	require.Len(t, publicProducts[0].ProductType.SpecDefinitions, 1)
	require.Len(t, publicProducts[0].VariantOptionValues, 1)
	assert.Equal(t, "ruby_red", publicProducts[0].VariantOptionValues[0].ValueKey)
}

func TestProductServiceCreateAdminProductRejectsMediaBoundToOtherProductVariantOptionValue(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)
	brakeSpec := findSpecDefinitionBySlug(t, db, productType.ID, "brake_type")

	sourceProduct, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductTypeID: &productType.ID,
		Name:          "Source Rim",
		Slug:          "source-rim",
		Status:        "active",
		Locale:        "en",
		SpecValues: map[string]string{
			"outer_width_mm": "30",
		},
		VariantOptionValues: []ProductVariantOptionValueInput{
			{
				SpecDefinitionID: brakeSpec.ID,
				ValueKey:         "disc",
				Label:            "Disc Brake",
				ColorHex:         "#2A2A2A",
				IsEnabled:        boolPtr(true),
			},
		},
		Variants: []ProductVariantInput{
			{
				SKU:          "SOURCE-RIM-DISC",
				OptionValues: map[string]string{"brake_type": "disc"},
				Price:        399,
				Stock:        5,
				IsDefault:    true,
				IsActive:     boolPtr(true),
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, sourceProduct.VariantOptionValues, 1)
	foreignOptionValueID := sourceProduct.VariantOptionValues[0].ID

	createdProduct, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductTypeID: &productType.ID,
		Name:          "Bad Media Binding Rim",
		Slug:          "bad-media-binding-rim",
		Status:        "active",
		Locale:        "en",
		SpecValues: map[string]string{
			"outer_width_mm": "30",
		},
		Variants: []ProductVariantInput{
			{
				SKU:          "BAD-MEDIA-BINDING-DISC",
				OptionValues: map[string]string{"brake_type": "disc"},
				Price:        399,
				Stock:        5,
				IsDefault:    true,
				IsActive:     boolPtr(true),
			},
		},
		Media: []ProductMediaInput{
			{
				VariantOptionValueID: &foreignOptionValueID,
				MediaType:            "image",
				URL:                  "/uploads/products/bad-binding.webp",
			},
		},
	})

	require.Nil(t, createdProduct)
	require.ErrorIs(t, err, ErrProductMediaInvalid)

	var count int64
	require.NoError(t, db.Model(&product.Product{}).Where("slug = ?", "bad-media-binding-rim").Count(&count).Error)
	assert.EqualValues(t, 0, count)
}

func TestProductServiceCreateAdminProductRejectsInvalidMediaLocale(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)

	createdProduct, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductTypeID: &productType.ID,
		Name:          "Bad Media Locale Rim",
		Slug:          "bad-media-locale-rim",
		Status:        "active",
		Locale:        "en",
		SpecValues: map[string]string{
			"outer_width_mm": "30",
		},
		Variants: []ProductVariantInput{
			{
				SKU:          "BAD-MEDIA-LOCALE-DISC",
				OptionValues: map[string]string{"brake_type": "disc"},
				Price:        399,
				Stock:        5,
				IsDefault:    true,
				IsActive:     boolPtr(true),
			},
		},
		Media: []ProductMediaInput{
			{
				MediaType: "image",
				URL:       "/uploads/products/bad-media-locale.webp",
				Locale:    "zz",
			},
		},
	})

	require.Nil(t, createdProduct)
	require.ErrorIs(t, err, ErrProductMediaInvalid)
}

func TestProductServiceCreateAdminProductNormalizesMediaLocaleAlias(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)

	createdProduct, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductTypeID: &productType.ID,
		Name:          "Media Locale Alias Rim",
		Slug:          "media-locale-alias-rim",
		Status:        "active",
		Locale:        "en",
		SpecValues: map[string]string{
			"outer_width_mm": "30",
		},
		Variants: []ProductVariantInput{
			{
				SKU:          "MEDIA-LOCALE-ALIAS-DISC",
				OptionValues: map[string]string{"brake_type": "disc"},
				Price:        399,
				Stock:        5,
				IsDefault:    true,
				IsActive:     boolPtr(true),
			},
		},
		Media: []ProductMediaInput{
			{
				MediaType: "image",
				URL:       "/uploads/products/media-locale-alias.webp",
				Locale:    "en-US",
			},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, createdProduct)
	require.Len(t, createdProduct.Media, 1)
	assert.Equal(t, "en", createdProduct.Media[0].Locale)
}

func TestProductServiceCreateAdminProductUsesPrimaryPricingCurrency(t *testing.T) {
	db, productService := newTestProductService(t)
	policy := NewCurrencyPolicyService(repository.NewSettingRepository(db))
	_, err := policy.UpdatePolicy(currency.Policy{
		PrimaryCurrency:   "CNY",
		DisplayCurrencies: []string{"USD", "EUR"},
	})
	require.NoError(t, err)
	productService.ConfigureCurrencyPolicy(policy)

	createdProduct, err := productService.CreateAdminProduct(ProductCreateInput{
		Name:     "Carbon Rim CNY",
		Slug:     "carbon-rim-cny",
		Currency: "CNY",
		Status:   "active",
		Locale:   "en",
		Variants: []ProductVariantInput{
			{
				SKU:       "RIM-CNY-001",
				Currency:  "CNY",
				Price:     699,
				Stock:     5,
				IsDefault: true,
				IsActive:  boolPtr(true),
			},
		},
	})

	require.NoError(t, err)
	require.Equal(t, "CNY", createdProduct.Currency)
	require.Len(t, createdProduct.Variants, 1)
	require.Equal(t, "CNY", createdProduct.Variants[0].Currency)
}

func TestProductServiceCreateAdminProductPersistsDisplayPriceSnapshots(t *testing.T) {
	db, productService := newTestProductService(t)
	policy := NewCurrencyPolicyService(repository.NewSettingRepository(db))
	_, err := policy.UpdatePolicy(currency.Policy{
		PrimaryCurrency:   "CNY",
		DisplayCurrencies: []string{"USD", "EUR"},
	})
	require.NoError(t, err)
	productService.ConfigureCurrencyPolicy(policy)

	createdProduct, err := productService.CreateAdminProduct(ProductCreateInput{
		Name:     "Carbon Rim Display Prices",
		Slug:     "carbon-rim-display-prices",
		Currency: "CNY",
		Status:   "active",
		Locale:   "en",
		Variants: []ProductVariantInput{
			{
				SKU:       "RIM-DISPLAY-001",
				Currency:  "CNY",
				Price:     699,
				Stock:     5,
				IsDefault: true,
				IsActive:  boolPtr(true),
				DisplayPrices: []currency.DisplayPriceSnapshot{
					{
						Amount:        96.8,
						Currency:      "USD",
						QuoteCurrency: "USD",
						Rate:          0.1385,
						Source:        "direct_rate",
						Converted:     true,
					},
					{
						Amount:        699,
						Currency:      "CNY",
						QuoteCurrency: "CNY",
						Rate:          1,
						Source:        "base_currency",
						Converted:     true,
					},
				},
			},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, createdProduct)
	require.Len(t, createdProduct.Variants, 1)

	variantSnapshots := currency.ParseDisplayPriceSnapshots(createdProduct.Variants[0].DisplayPriceData)
	require.Len(t, variantSnapshots, 1)
	assert.Equal(t, "USD", variantSnapshots[0].Currency)
	assert.Equal(t, "USD", variantSnapshots[0].QuoteCurrency)
	assert.InDelta(t, 96.8, variantSnapshots[0].Amount, 0.001)
	assert.InDelta(t, 0.1385, variantSnapshots[0].Rate, 0.000001)
	assert.Equal(t, "direct_rate", variantSnapshots[0].Source)
	assert.True(t, variantSnapshots[0].Converted)

	productSnapshots := currency.ParseDisplayPriceSnapshots(createdProduct.DisplayPriceData)
	require.Len(t, productSnapshots, 1)
	assert.Equal(t, variantSnapshots[0], productSnapshots[0])
}

func TestProductServiceCreateAdminProductRejectsInvalidTemplateSpec(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)

	createdProduct, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductTypeID: &productType.ID,
		Name:          "Invalid Carbon Rim",
		Slug:          "invalid-carbon-rim",
		Status:        "active",
		Locale:        "en",
		SpecValues: map[string]string{
			"outer_width_mm": "30",
		},
		Variants: []ProductVariantInput{
			{
				SKU:          "RIM-INVALID-CANTI",
				OptionValues: map[string]string{"brake_type": "cantilever"},
				Price:        399,
				Stock:        5,
				IsDefault:    true,
				IsActive:     boolPtr(true),
			},
		},
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrProductSpecInvalid) || errors.Is(err, ErrProductVariantInvalid))
	assert.Nil(t, createdProduct)

	var productCount int64
	require.NoError(t, db.Model(&product.Product{}).Count(&productCount).Error)
	assert.Equal(t, int64(0), productCount)
}

func TestProductServiceUpdateAdminProductAllowsKeepingDisabledInformationTemplate(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)
	createdProduct := createProductWithSpecs(t, productService, productType.ID, "RIM-DISABLED-TEMPLATE", "disabled-template-rim", map[string]string{
		"outer_width_mm": "30",
	}, map[string]string{"brake_type": "disc"})

	template := product.ProductInformationTemplate{
		Kind:      product.ProductInformationTemplateKindAfterSales,
		Name:      "Legacy After-sales",
		Slug:      "legacy-after-sales",
		Content:   "<p>Legacy content</p>",
		Locale:    "en",
		IsEnabled: false,
	}
	templateRepo := repository.NewProductInformationTemplateRepository(db)
	require.NoError(t, templateRepo.Create(&template))
	storedTemplate, err := templateRepo.FindByID(template.ID)
	require.NoError(t, err)
	require.False(t, storedTemplate.IsEnabled)
	require.NoError(t, db.Model(&product.Product{}).
		Where("id = ?", createdProduct.ID).
		Update("after_sales_template_id", template.ID).Error)

	name := "Updated Product Name"
	templateID := template.ID
	updatedProduct, err := productService.UpdateAdminProduct(createdProduct.ID, ProductUpdateInput{
		AfterSalesTemplateID:       &templateID,
		UpdateAfterSalesTemplateID: true,
		Name:                       &name,
	})

	require.NoError(t, err)
	require.Equal(t, name, updatedProduct.Name)
	require.NotNil(t, updatedProduct.AfterSalesTemplateID)
	require.Equal(t, template.ID, *updatedProduct.AfterSalesTemplateID)
}

func TestProductServiceUpdateAdminProductRejectsSwitchingToDisabledInformationTemplate(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)
	createdProduct := createProductWithSpecs(t, productService, productType.ID, "RIM-NEW-DISABLED-TEMPLATE", "new-disabled-template-rim", map[string]string{
		"outer_width_mm": "30",
	}, map[string]string{"brake_type": "disc"})

	template := product.ProductInformationTemplate{
		Kind:      product.ProductInformationTemplateKindPackaging,
		Name:      "Disabled Packaging",
		Slug:      "disabled-packaging",
		Content:   "<p>Disabled content</p>",
		Locale:    "en",
		IsEnabled: false,
	}
	templateRepo := repository.NewProductInformationTemplateRepository(db)
	require.NoError(t, templateRepo.Create(&template))
	storedTemplate, err := templateRepo.FindByID(template.ID)
	require.NoError(t, err)
	require.False(t, storedTemplate.IsEnabled)

	templateID := template.ID
	updatedProduct, err := productService.UpdateAdminProduct(createdProduct.ID, ProductUpdateInput{
		PackagingTemplateID:       &templateID,
		UpdatePackagingTemplateID: true,
	})

	require.Nil(t, updatedProduct)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrProductInformationTemplateInvalid))
}

func TestProductServiceUpdateAdminProductRejectsLocaleChange(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)
	createdProduct := createProductWithSpecs(t, productService, productType.ID, "RIM-LOCALE-LOCK", "locale-lock-rim", map[string]string{
		"outer_width_mm": "30",
	}, map[string]string{"brake_type": "disc"})

	nextLocale := "fr"
	updatedProduct, err := productService.UpdateAdminProduct(createdProduct.ID, ProductUpdateInput{
		Locale: &nextLocale,
	})

	require.Nil(t, updatedProduct)
	require.ErrorIs(t, err, ErrProductLocaleImmutable)

	var storedProduct product.Product
	require.NoError(t, db.First(&storedProduct, createdProduct.ID).Error)
	assert.Equal(t, "en", storedProduct.Locale)
}

func TestProductServiceUpdateAdminProductAcceptsSameLocaleAlias(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)
	createdProduct := createProductWithSpecs(t, productService, productType.ID, "RIM-LOCALE-ALIAS", "locale-alias-rim", map[string]string{
		"outer_width_mm": "30",
	}, map[string]string{"brake_type": "disc"})

	nextLocale := "en-US"
	updatedProduct, err := productService.UpdateAdminProduct(createdProduct.ID, ProductUpdateInput{
		Locale: &nextLocale,
	})

	require.NoError(t, err)
	require.NotNil(t, updatedProduct)
	assert.Equal(t, "en", updatedProduct.Locale)
}

func TestProductServiceUpdateAdminProductPreservesInactiveVariantWhenAnotherVariantIsActive(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)
	createdProduct := createProductWithSpecs(t, productService, productType.ID, "RIM-INACTIVE", "inactive-rim", map[string]string{
		"outer_width_mm": "30",
	}, map[string]string{"brake_type": "disc"})
	require.Len(t, createdProduct.Variants, 1)

	inactive := false
	variantID := createdProduct.Variants[0].ID
	updatedProduct, err := productService.UpdateAdminProduct(createdProduct.ID, ProductUpdateInput{
		Variants: []ProductVariantInput{
			{
				ID:           &variantID,
				SKU:          createdProduct.Variants[0].SKU,
				OptionValues: map[string]string{"brake_type": "disc"},
				Price:        399,
				Stock:        5,
				IsDefault:    true,
				IsActive:     &inactive,
			},
			{
				SKU:          "RIM-INACTIVE-ACTIVE-VAR",
				OptionValues: map[string]string{"brake_type": "rim"},
				Price:        419,
				Stock:        7,
				IsDefault:    false,
				IsActive:     boolPtr(true),
			},
		},
		UpdateVariants: true,
	})

	require.NoError(t, err)
	require.Len(t, updatedProduct.Variants, 2)
	assert.Equal(t, 7, updatedProduct.TotalVariantStock())

	var inactiveVariant, activeVariant *product.ProductVariant
	for i := range updatedProduct.Variants {
		switch updatedProduct.Variants[i].SKU {
		case createdProduct.Variants[0].SKU:
			inactiveVariant = &updatedProduct.Variants[i]
		case "RIM-INACTIVE-ACTIVE-VAR":
			activeVariant = &updatedProduct.Variants[i]
		}
	}
	require.NotNil(t, inactiveVariant)
	require.NotNil(t, activeVariant)
	assert.False(t, inactiveVariant.IsActive)
	assert.False(t, inactiveVariant.IsDefault)
	assert.True(t, activeVariant.IsActive)
	assert.True(t, activeVariant.IsDefault)
}

func TestProductServiceCreateAdminProductRejectsAllInactiveVariants(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)
	inactive := false

	createdProduct, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductTypeID: &productType.ID,
		Name:          "Inactive SKU Rim",
		Slug:          "inactive-sku-rim",
		Status:        "active",
		Locale:        "en",
		SpecValues: map[string]string{
			"outer_width_mm": "30",
		},
		Variants: []ProductVariantInput{
			{
				SKU:          "RIM-ALL-INACTIVE-DISC",
				OptionValues: map[string]string{"brake_type": "disc"},
				Price:        399,
				Stock:        15,
				IsDefault:    true,
				IsActive:     &inactive,
			},
		},
	})

	require.ErrorIs(t, err, ErrProductVariantInvalid)
	assert.Nil(t, createdProduct)

	var productCount int64
	require.NoError(t, db.Model(&product.Product{}).Count(&productCount).Error)
	assert.Equal(t, int64(0), productCount)
}

func TestProductServiceCreateAdminProductRequiresVariant(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)

	createdProduct, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductTypeID: &productType.ID,
		Name:          "No Variant Rim",
		Slug:          "no-variant-rim",
		Status:        "active",
		Locale:        "en",
		SpecValues: map[string]string{
			"outer_width_mm": "30",
		},
	})

	require.ErrorIs(t, err, ErrProductVariantInvalid)
	assert.Nil(t, createdProduct)

	var productCount int64
	require.NoError(t, db.Model(&product.Product{}).Count(&productCount).Error)
	assert.Equal(t, int64(0), productCount)
}

func TestProductServiceSearchPublicFiltersByTemplateSpec(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)

	discRim := createProductWithSpecs(t, productService, productType.ID, "RIM-DISC", "disc-rim", map[string]string{
		"outer_width_mm": "30",
	}, map[string]string{"brake_type": "disc"})
	createProductWithSpecs(t, productService, productType.ID, "RIM-RIM", "rim-brake-rim", map[string]string{
		"outer_width_mm": "25",
	}, map[string]string{"brake_type": "rim"})

	results, total, err := productService.SearchPublic(ProductSearchInput{
		Locale: "en",
		Status: "active",
		SpecFilters: map[string][]string{
			"brake_type": {"disc"},
		},
		Page:     1,
		PageSize: 20,
	})

	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, results, 1)
	assert.Equal(t, discRim.ID, results[0].ID)
}

func TestProductServicePublicAccessOnlyReturnsActiveProducts(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)

	activeProduct := createProductWithSpecs(t, productService, productType.ID, "RIM-PUBLIC-ACTIVE", "public-active", map[string]string{
		"outer_width_mm": "30",
	}, map[string]string{"brake_type": "disc"})
	inactiveProduct := createProductWithSpecs(t, productService, productType.ID, "RIM-PUBLIC-INACTIVE", "public-inactive", map[string]string{
		"outer_width_mm": "30",
	}, map[string]string{"brake_type": "disc"})
	require.NoError(t, db.Model(&product.Product{}).Where("id = ?", inactiveProduct.ID).Update("status", "inactive").Error)

	_, err := productService.GetPublicByID(inactiveProduct.ID)
	require.ErrorIs(t, err, ErrProductNotFound)

	products, total, err := productService.ListPublic("en", false, 1, 20)
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, products, 1)
	assert.Equal(t, activeProduct.ID, products[0].ID)

	results, total, err := productService.SearchPublic(ProductSearchInput{
		Locale:   "en",
		Status:   "inactive",
		Page:     1,
		PageSize: 20,
	})
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, results, 1)
	assert.Equal(t, activeProduct.ID, results[0].ID)
}

func TestProductRepositoryPurchasableVariantRejectsInactiveProduct(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)

	inactiveProduct := createProductWithSpecs(t, productService, productType.ID, "RIM-NOT-BUYABLE", "not-buyable", map[string]string{
		"outer_width_mm": "30",
	}, map[string]string{"brake_type": "disc"})
	require.NoError(t, db.Model(&product.Product{}).Where("id = ?", inactiveProduct.ID).Update("status", "inactive").Error)

	_, _, err := repository.NewProductRepository(db).FindPurchasableVariant(inactiveProduct.ID, nil)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestProductServicePublicCatalogRespectsProductLocale(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)
	createdProduct, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductTypeID: &productType.ID,
		Name:          "Unified Rim",
		Slug:          "global-rim",
		Status:        "active",
		Locale:        "zh_cn",
		SpecValues: map[string]string{
			"outer_width_mm": "30",
		},
		Variants: []ProductVariantInput{
			{
				SKU:          "RIM-GLOBAL-VAR",
				OptionValues: map[string]string{"brake_type": "disc"},
				Price:        399,
				Stock:        5,
				IsDefault:    true,
				IsActive:     boolPtr(true),
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, createdProduct)

	_, err = productService.GetPublicBySlug("global-rim", "en")
	require.Error(t, err)

	publicProduct, err := productService.GetPublicBySlug("global-rim", "zh_cn")
	require.NoError(t, err)
	assert.Equal(t, createdProduct.ID, publicProduct.ID)

	listed, listTotal, err := productService.ListPublic("en", false, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(0), listTotal)
	assert.Empty(t, listed)

	listed, listTotal, err = productService.ListPublic("zh_cn", false, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(1), listTotal)
	require.Len(t, listed, 1)
	assert.Equal(t, createdProduct.ID, listed[0].ID)

	searched, searchTotal, err := productService.SearchPublic(ProductSearchInput{
		Locale:   "en",
		Status:   "active",
		Page:     1,
		PageSize: 20,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(0), searchTotal)
	assert.Empty(t, searched)

	searched, searchTotal, err = productService.SearchPublic(ProductSearchInput{
		Locale:   "zh_cn",
		Status:   "active",
		Page:     1,
		PageSize: 20,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), searchTotal)
	require.Len(t, searched, 1)
	assert.Equal(t, createdProduct.ID, searched[0].ID)

	available, availableTotal, err := productService.ListPublicAvailable("en", 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(0), availableTotal)
	assert.Empty(t, available)

	available, availableTotal, err = productService.ListPublicAvailable("zh_cn", 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(1), availableTotal)
	require.Len(t, available, 1)
	assert.Equal(t, createdProduct.ID, available[0].ID)
}

func TestProductServicePublicProductReturnsOnlyActiveTranslationRoutes(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)

	root := createProductWithSpecs(t, productService, productType.ID, "RIM-TRANSLATION-EN", "translation-rim", map[string]string{
		"outer_width_mm": "30",
	}, map[string]string{"brake_type": "disc"})
	parentID := root.ID
	translated := createProductWithSpecs(t, productService, productType.ID, "RIM-TRANSLATION-ZH", "translated-rim", map[string]string{
		"outer_width_mm": "30",
	}, map[string]string{"brake_type": "disc"})
	require.NoError(t, db.Model(&product.Product{}).Where("id = ?", translated.ID).Updates(map[string]interface{}{
		"locale":    "zh_cn",
		"parent_id": parentID,
	}).Error)

	inactive := createProductWithSpecs(t, productService, productType.ID, "RIM-TRANSLATION-FR", "inactive-rim", map[string]string{
		"outer_width_mm": "30",
	}, map[string]string{"brake_type": "disc"})
	require.NoError(t, db.Model(&product.Product{}).Where("id = ?", inactive.ID).Updates(map[string]interface{}{
		"locale":    "fr",
		"parent_id": parentID,
		"status":    "inactive",
	}).Error)

	resolved, routes, err := productService.GetPublicBySlugWithRoutes("translated-rim", "zh_cn")
	require.NoError(t, err)
	require.Equal(t, translated.ID, resolved.ID)
	require.Len(t, routes, 2)
	assert.Equal(t, "en", routes[0].Locale)
	assert.Equal(t, "translation-rim", routes[0].Slug)
	assert.Equal(t, "zh_cn", routes[1].Locale)
	assert.Equal(t, "translated-rim", routes[1].Slug)
}

func TestProductServiceValidatesProductTranslationRelationships(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)

	root := createProductWithSpecs(t, productService, productType.ID, "RIM-RELATION-EN", "relation-rim", map[string]string{
		"outer_width_mm": "30",
	}, map[string]string{"brake_type": "disc"})

	createTranslation := func(locale, sku, slug string, parentID *uint) (*product.Product, error) {
		return productService.CreateAdminProduct(ProductCreateInput{
			ProductTypeID: &productType.ID,
			Name:          sku,
			Slug:          slug,
			Status:        "active",
			Locale:        locale,
			ParentID:      parentID,
			SpecValues: map[string]string{
				"outer_width_mm": "30",
			},
			Variants: []ProductVariantInput{
				{
					SKU:          sku + "-VAR",
					OptionValues: map[string]string{"brake_type": "disc"},
					Price:        399,
					Stock:        5,
					IsDefault:    true,
					IsActive:     boolPtr(true),
				},
			},
		})
	}

	translated, err := createTranslation("zh_cn", "RIM-RELATION-ZH", "relation-rim-zh", &root.ID)
	require.NoError(t, err)
	require.NotNil(t, translated)

	_, err = createTranslation("zh_cn", "RIM-RELATION-ZH-DUPLICATE", "relation-rim-zh-duplicate", &root.ID)
	require.ErrorIs(t, err, ErrProductTranslationInvalid)

	_, err = createTranslation("en", "RIM-RELATION-EN-DUPLICATE", "relation-rim-en-duplicate", &root.ID)
	require.ErrorIs(t, err, ErrProductTranslationInvalid)

	_, err = createTranslation("fr", "RIM-RELATION-FR-NESTED", "relation-rim-fr-nested", &translated.ID)
	require.ErrorIs(t, err, ErrProductTranslationInvalid)
}

func TestProductServiceCopyAdminProductTranslationCreatesGroupedCopyWithUniqueSlugAndSKU(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)
	brakeSpec := findSpecDefinitionBySlug(t, db, productType.ID, "brake_type")

	source, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductTypeID: &productType.ID,
		Name:          "Translation Copy Rim",
		Slug:          "translation-copy-rim",
		Status:        "active",
		Locale:        "en",
		SpecValues: map[string]string{
			"outer_width_mm": "30",
			"tubeless_ready": "yes",
		},
		VariantOptionValues: []ProductVariantOptionValueInput{
			{
				SpecDefinitionID: brakeSpec.ID,
				ValueKey:         "disc",
				Label:            "Disc Brake",
				ColorHex:         "#111111",
				IsEnabled:        boolPtr(true),
			},
		},
		Variants: []ProductVariantInput{
			{
				SKU:          "RIM-COPY-DISC",
				OptionValues: map[string]string{"brake_type": "disc"},
				Price:        399,
				Stock:        5,
				IsDefault:    true,
				IsActive:     boolPtr(true),
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, source.Variants, 1)
	require.Len(t, source.VariantOptionValues, 1)

	sourceVariantID := source.Variants[0].ID
	sourceOptionValueID := source.VariantOptionValues[0].ID
	require.NoError(t, db.Create(&product.ProductMedia{
		ProductID:            source.ID,
		VariantID:            &sourceVariantID,
		VariantOptionValueID: &sourceOptionValueID,
		MediaType:            "image",
		Role:                 "primary",
		URL:                  "/uploads/products/source-rim.webp",
		IsPrimary:            true,
		IsVisible:            true,
	}).Error)

	_, err = productService.CreateAdminProduct(ProductCreateInput{
		ProductTypeID: &productType.ID,
		Name:          "French Occupant",
		Slug:          "translation-copy-rim",
		Status:        "active",
		Locale:        "fr",
		SpecValues: map[string]string{
			"outer_width_mm": "30",
		},
		Variants: []ProductVariantInput{
			{
				SKU:          "RIM-COPY-DISC-fr",
				OptionValues: map[string]string{"brake_type": "disc"},
				Price:        399,
				Stock:        5,
				IsDefault:    true,
				IsActive:     boolPtr(true),
			},
		},
	})
	require.NoError(t, err)

	translated, group, err := productService.CopyAdminProductTranslation(source.ID, "fr")
	require.NoError(t, err)
	require.NotNil(t, translated)
	require.NotNil(t, group)
	require.NotNil(t, translated.ParentID)
	assert.Equal(t, source.ID, *translated.ParentID)
	assert.Equal(t, "fr", translated.Locale)
	assert.Equal(t, "translation-copy-rim-fr", translated.Slug)
	assert.Equal(t, "RIM-COPY-DISC-fr-2", translated.SKU)
	require.Len(t, translated.SpecValues, 2)
	require.Len(t, translated.Variants, 1)
	require.Len(t, translated.VariantOptionValues, 1)
	require.Len(t, translated.Media, 1)
	assert.Equal(t, "RIM-COPY-DISC-fr-2", translated.Variants[0].SKU)
	assert.NotEqual(t, sourceVariantID, translated.Variants[0].ID)
	assert.NotEqual(t, sourceOptionValueID, translated.VariantOptionValues[0].ID)
	require.NotNil(t, translated.Media[0].VariantID)
	require.NotNil(t, translated.Media[0].VariantOptionValueID)
	assert.Equal(t, translated.Variants[0].ID, *translated.Media[0].VariantID)
	assert.Equal(t, translated.VariantOptionValues[0].ID, *translated.Media[0].VariantOptionValueID)
	assert.NotContains(t, group.MissingLocales, "fr")
}

func TestProductServiceCopyAdminProductTranslationRejectsExistingGroupLocale(t *testing.T) {
	db, productService := newTestProductService(t)
	productType := seedCarbonRimType(t, db)

	root := createProductWithSpecs(t, productService, productType.ID, "RIM-COPY-EXISTS-EN", "copy-exists-rim", map[string]string{
		"outer_width_mm": "30",
	}, map[string]string{"brake_type": "disc"})
	parentID := root.ID
	_, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductTypeID: &productType.ID,
		Name:          "Existing French Rim",
		Slug:          "copy-exists-rim-fr",
		Status:        "active",
		Locale:        "fr",
		ParentID:      &parentID,
		SpecValues: map[string]string{
			"outer_width_mm": "30",
		},
		Variants: []ProductVariantInput{
			{
				SKU:          "RIM-COPY-EXISTS-FR-VAR",
				OptionValues: map[string]string{"brake_type": "disc"},
				Price:        399,
				Stock:        5,
				IsDefault:    true,
				IsActive:     boolPtr(true),
			},
		},
	})
	require.NoError(t, err)

	_, _, err = productService.CopyAdminProductTranslation(root.ID, "fr")
	require.ErrorIs(t, err, ErrProductTranslationExists)

	translated, err := productService.GetAdminProductTranslationGroup(root.ID)
	require.NoError(t, err)
	assert.NotContains(t, translated.MissingLocales, "fr")
	assert.NotContains(t, translated.MissingLocales, "en")
}

func newTestProductService(t *testing.T) (*gorm.DB, *ProductService) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(t, db.AutoMigrate(
		&product.ProductType{},
		&product.ProductTypeTranslation{},
		&product.SpecDefinition{},
		&product.ProductInformationTemplate{},
		&product.Product{},
		&product.ProductMedia{},
		&product.ProductSpecValue{},
		&product.ProductVariant{},
		&product.ProductVariantOptionValue{},
		&setting.Setting{},
	))

	productService := NewProductService(repository.NewProductRepository(db), nil, 0)
	productService.ConfigureInformationTemplateRepository(repository.NewProductInformationTemplateRepository(db))
	return db, productService
}

func seedCarbonRimType(t *testing.T, db *gorm.DB) product.ProductType {
	t.Helper()

	productType := product.ProductType{
		Name:      "Carbon Rim",
		Slug:      "carbon_rim",
		IsEnabled: true,
	}
	require.NoError(t, db.Create(&productType).Error)

	specDefinitions := []product.SpecDefinition{
		{
			ProductTypeID: productType.ID,
			Group:         "Dimensions",
			Name:          "Outer Width",
			Slug:          "outer_width_mm",
			FieldType:     "number",
			Unit:          "mm",
			IsRequired:    true,
			IsFilterable:  true,
			IsVisible:     true,
			SortOrder:     10,
		},
		{
			ProductTypeID:   productType.ID,
			Group:           "Compatibility",
			Name:            "Brake Type",
			Slug:            "brake_type",
			FieldType:       "select",
			IsRequired:      true,
			IsFilterable:    true,
			IsVisible:       true,
			IsVariantOption: true,
			SortOrder:       20,
			Options:         `["disc","rim"]`,
		},
		{
			ProductTypeID: productType.ID,
			Group:         "Compatibility",
			Name:          "Tubeless Ready",
			Slug:          "tubeless_ready",
			FieldType:     "boolean",
			IsFilterable:  true,
			IsVisible:     true,
			SortOrder:     30,
		},
	}
	require.NoError(t, db.Create(&specDefinitions).Error)

	return productType
}

func createProductWithSpecs(t *testing.T, productService *ProductService, productTypeID uint, sku, slug string, specs map[string]string, variantOptions map[string]string) *product.Product {
	t.Helper()

	createdProduct, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductTypeID: &productTypeID,
		Name:          sku,
		Slug:          slug,
		Status:        "active",
		Locale:        "en",
		SpecValues:    specs,
		Variants: []ProductVariantInput{
			{
				SKU:          sku + "-VAR",
				OptionValues: variantOptions,
				Price:        399,
				Stock:        5,
				IsDefault:    true,
				IsActive:     boolPtr(true),
			},
		},
	})
	require.NoError(t, err)
	return createdProduct
}

func findSavedSpecValue(t *testing.T, productRecord *product.Product, slug string) string {
	t.Helper()

	for _, specValue := range productRecord.SpecValues {
		if specValue.SpecDefinition != nil && specValue.SpecDefinition.Slug == slug {
			return specValue.Value
		}
	}

	t.Fatalf("spec value %q not found", slug)
	return ""
}

func findSpecDefinitionBySlug(t *testing.T, db *gorm.DB, productTypeID uint, slug string) product.SpecDefinition {
	t.Helper()

	var definition product.SpecDefinition
	require.NoError(t, db.Where("product_type_id = ? AND slug = ?", productTypeID, slug).First(&definition).Error)
	return definition
}
