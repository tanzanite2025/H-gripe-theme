package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/database"
	"commerce-platform/internal/pkg/safehtml"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"gorm.io/gorm"
)

const (
	verifiedSAPIMBrandName       = "SAPIM"
	verifiedSAPIMBrandSlug       = "sapim"
	verifiedSAPIMBrandWebsiteURL = "https://sapim.eu/"
)

type verifiedSAPIMSpokeProductSeed struct {
	name          string
	slug          string
	shortDesc     string
	description   string
	specification product.SpokesProductSpecification
}

func main() {
	if err := runCreateVerifiedSAPIMSpokeProductCatalog(); err != nil {
		log.Fatal(err)
	}
}

func runCreateVerifiedSAPIMSpokeProductCatalog() error {
	applicationConfig, err := config.Load()
	if err != nil {
		return fmt.Errorf("load application configuration: %w", err)
	}
	databaseConnection, err := database.Init(applicationConfig.Database)
	if err != nil {
		return fmt.Errorf("connect to the configured catalog database: %w", err)
	}
	sqlDatabase, err := databaseConnection.DB()
	if err != nil {
		return fmt.Errorf("get catalog database connection: %w", err)
	}
	defer func() {
		if err := sqlDatabase.Close(); err != nil {
			log.Printf("close catalog database: %v", err)
		}
	}()

	brandRepository := repository.NewProductBrandRepository(databaseConnection)
	productRepository := repository.NewProductRepository(databaseConnection)
	outboxRepository := repository.NewOutboxRepository(databaseConnection)
	productCachePublisher := service.NewProductCacheOutboxPublisher(outboxRepository)
	merchantPublisher := service.NewMerchantOutboxPublisher(outboxRepository)
	transactionManager := repository.NewTxManager(
		databaseConnection,
		repository.NewOrderRepository(databaseConnection),
		productRepository,
		repository.NewCouponRepository(databaseConnection),
		repository.NewLoyaltyRepository(databaseConnection),
		repository.NewPaymentRepository(databaseConnection),
	)
	transactionManager.ConfigureOutboxRepository(outboxRepository)
	transactionManager.ConfigureProductBrandRepository(brandRepository)

	brandService := service.NewProductBrandService(brandRepository)
	brandService.ConfigureProductDependencies(productRepository, productCachePublisher, merchantPublisher)
	brandService.ConfigureTxManager(transactionManager)
	brand, err := findOrCreateVerifiedSAPIMProductBrand(databaseConnection, brandService)
	if err != nil {
		return err
	}

	category, err := repository.NewProductCategoryRepository(databaseConnection).FindBySlug(product.SpokesProductCategorySlug, false)
	if err != nil {
		return fmt.Errorf("find enabled spokes product category: %w", err)
	}

	productService := service.NewProductService(productRepository, nil, 0)
	productService.ConfigureProductBrandRepository(brandRepository)
	productService.ConfigureProductCategoryRepository(repository.NewProductCategoryRepository(databaseConnection))
	productService.ConfigureProductCacheEventPublisher(productCachePublisher)
	productService.ConfigureMerchantEventPublisher(merchantPublisher)
	productService.ConfigureTxManager(transactionManager)

	for _, seed := range verifiedSAPIMSpokeProductSeeds() {
		created, alreadyPresent, err := createOrVerifySAPIMSpokeProduct(productService, databaseConnection, seed, brand.ID, category.ID)
		if err != nil {
			return err
		}
		state := "CREATED"
		if alreadyPresent {
			state = "VERIFIED"
		}
		if err := ensureSAPIMSpokeProductWithdrawalEvent(databaseConnection, merchantPublisher, created.ID); err != nil {
			return err
		}
		fmt.Printf("%s id=%d slug=%s product_type=%s status=%s variants=0 stock=0\n", state, created.ID, created.Slug, created.ProductType, created.Status)
	}
	return nil
}

func findOrCreateVerifiedSAPIMProductBrand(databaseConnection *gorm.DB, brandService *service.ProductBrandService) (*product.ProductBrand, error) {
	var brand product.ProductBrand
	err := databaseConnection.Where("slug = ?", verifiedSAPIMBrandSlug).First(&brand).Error
	if err == nil {
		if err := validateExistingVerifiedSAPIMBrandIdentity(brand); err != nil {
			return nil, err
		}
		return &brand, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("find SAPIM brand master data: %w", err)
	}
	created, err := brandService.Create(service.ProductBrandInput{
		Name:        verifiedSAPIMBrandName,
		Slug:        verifiedSAPIMBrandSlug,
		WebsiteURL:  verifiedSAPIMBrandWebsiteURL,
		Description: "Official SAPIM manufacturer brand record.",
		IsEnabled:   true,
	})
	if err != nil {
		return nil, fmt.Errorf("create SAPIM brand master data: %w", err)
	}
	return created, nil
}

func validateExistingVerifiedSAPIMBrandIdentity(brand product.ProductBrand) error {
	if brand.Name != verifiedSAPIMBrandName || brand.Slug != verifiedSAPIMBrandSlug || brand.WebsiteURL != verifiedSAPIMBrandWebsiteURL {
		return fmt.Errorf("SAPIM brand slug exists with a different name or official website; refusing to attach catalog products")
	}
	if !brand.IsEnabled {
		return fmt.Errorf("SAPIM brand exists but is disabled; refusing to attach catalog products")
	}
	return nil
}

func verifiedSAPIMSpokeProductSeeds() []verifiedSAPIMSpokeProductSeed {
	cxRaySource := "https://sapim.eu/wp-content/uploads/2026/03/CX-Ray_hr-scaled.png"
	dLightSource := "https://sapim.eu/wp-content/uploads/2026/03/D-Light_hr-scaled.png"
	raceSource := "https://sapim.eu/wp-content/uploads/2026/03/Race_hr-scaled.png"
	leaderSource := "https://sapim.eu/wp-content/uploads/2026/03/Leader_hr-scaled.png"
	spokesOverviewSource := "https://sapim.eu/spokes/"
	colors := []string{"Black oxidized", "Silver"}

	return []verifiedSAPIMSpokeProductSeed{
		{
			name: "SAPIM CX-Ray Spoke", slug: "sapim-cx-ray-spoke",
			shortDesc:   "Individual SAPIM CX-Ray bicycle spoke.",
			description: "Individual SAPIM CX-Ray bicycle spoke. Product-level technical facts are recorded from the manufacturer's official model chart.",
			specification: product.SpokesProductSpecification{
				SchemaVersion: 1, ModelName: "CX-Ray", ManufacturerFamily: "Aero spokes",
				GaugeOptions: []product.SpokeGaugeOption{{
					Gauge: "14G", HubEndDiameterMM: spokeMeasurementPointer(2.0), MiddleWidthMM: spokeMeasurementPointer(2.2),
					MiddleThicknessMM: spokeMeasurementPointer(0.9), ThreadEndDiameterMM: spokeMeasurementPointer(2.0),
					LengthMinMM: spokeMeasurementPointer(141), LengthMaxMM: spokeMeasurementPointer(310),
					WeightReference: &product.SpokeWeightReference{SpokeCount: 64, LengthMM: 260, TotalWeightG: 279},
				}},
				MiddleStrength: &product.SpokeMiddleStrengthRange{MinimumNPerMM2: 1600, MaximumNPerMM2: 1600},
				ColorOptions:   colors, RecommendedUses: []string{"Road racing", "Triathlon", "Downhill biking"},
				SourceURLs: []string{cxRaySource},
			},
		},
		{
			name: "SAPIM D-Light Spoke", slug: "sapim-d-light-spoke",
			shortDesc:   "Individual SAPIM D-Light bicycle spoke.",
			description: "Individual SAPIM D-Light bicycle spoke. Product-level technical facts are recorded from the manufacturer's official model chart.",
			specification: product.SpokesProductSpecification{
				SchemaVersion: 1, ModelName: "D-Light", ManufacturerFamily: "Butted spokes",
				GaugeOptions: []product.SpokeGaugeOption{{
					Gauge: "14G", MiddleDiameterMM: spokeMeasurementPointer(1.65), ThreadEndDiameterMM: spokeMeasurementPointer(2.0),
					LengthMinMM: spokeMeasurementPointer(131), LengthMaxMM: spokeMeasurementPointer(340),
					WeightReference: &product.SpokeWeightReference{SpokeCount: 64, LengthMM: 260, TotalWeightG: 307},
				}},
				MiddleStrength: &product.SpokeMiddleStrengthRange{MinimumNPerMM2: 1370, MaximumNPerMM2: 1370},
				ColorOptions:   colors, SourceURLs: []string{dLightSource},
			},
		},
		{
			name: "SAPIM Race Spoke", slug: "sapim-race-spoke",
			shortDesc:   "Individual SAPIM Race bicycle spoke.",
			description: "Individual SAPIM Race bicycle spoke. Product-level technical facts are recorded from the manufacturer's official model chart.",
			specification: product.SpokesProductSpecification{
				SchemaVersion: 1, ModelName: "Race", ManufacturerFamily: "Butted spokes",
				GaugeOptions: []product.SpokeGaugeOption{{
					Gauge: "14G", HubEndDiameterMM: spokeMeasurementPointer(2.0), MiddleDiameterMM: spokeMeasurementPointer(1.8),
					ThreadEndDiameterMM: spokeMeasurementPointer(2.0), LengthMinMM: spokeMeasurementPointer(131), LengthMaxMM: spokeMeasurementPointer(340),
					WeightReference: &product.SpokeWeightReference{SpokeCount: 64, LengthMM: 260, TotalWeightG: 363},
				}},
				MiddleStrength: &product.SpokeMiddleStrengthRange{MinimumNPerMM2: 1370, MaximumNPerMM2: 1370},
				ColorOptions:   colors, SourceURLs: []string{raceSource},
			},
		},
		{
			name: "SAPIM Leader Spoke", slug: "sapim-leader-spoke",
			shortDesc:   "Individual SAPIM Leader bicycle spoke.",
			description: "Individual SAPIM Leader bicycle spoke. Product-level technical facts are recorded from the manufacturer's official model chart.",
			specification: product.SpokesProductSpecification{
				SchemaVersion: 1, ModelName: "Leader", ManufacturerFamily: "Basic spokes",
				GaugeOptions: []product.SpokeGaugeOption{
					{Gauge: "15G", DiameterMM: spokeMeasurementPointer(1.8), LengthMinMM: spokeMeasurementPointer(77), LengthMaxMM: spokeMeasurementPointer(340), WeightReference: &product.SpokeWeightReference{SpokeCount: 64, LengthMM: 260, TotalWeightG: 388}},
					{Gauge: "14G", DiameterMM: spokeMeasurementPointer(2.0), LengthMinMM: spokeMeasurementPointer(77), LengthMaxMM: spokeMeasurementPointer(340), WeightReference: &product.SpokeWeightReference{SpokeCount: 64, LengthMM: 260, TotalWeightG: 431}},
					{Gauge: "13G", DiameterMM: spokeMeasurementPointer(2.3), LengthMinMM: spokeMeasurementPointer(77), LengthMaxMM: spokeMeasurementPointer(340), WeightReference: &product.SpokeWeightReference{SpokeCount: 64, LengthMM: 260, TotalWeightG: 572}},
					{Gauge: "12G", DiameterMM: spokeMeasurementPointer(2.6), LengthMinMM: spokeMeasurementPointer(141), LengthMaxMM: spokeMeasurementPointer(310), WeightReference: &product.SpokeWeightReference{SpokeCount: 64, LengthMM: 260, TotalWeightG: 635}},
					{Gauge: "11G", DiameterMM: spokeMeasurementPointer(2.9), LengthMinMM: spokeMeasurementPointer(141), LengthMaxMM: spokeMeasurementPointer(310), WeightReference: &product.SpokeWeightReference{SpokeCount: 64, LengthMM: 260, TotalWeightG: 706}},
				},
				MiddleStrength: &product.SpokeMiddleStrengthRange{MinimumNPerMM2: 1080, MaximumNPerMM2: 1180},
				ColorOptions:   colors, SourceURLs: []string{leaderSource},
			},
		},
		{
			name: "SAPIM CX-Leader Spoke", slug: "sapim-cx-leader-spoke",
			shortDesc:   "Individual SAPIM CX-Leader bicycle spoke.",
			description: "Individual SAPIM CX-Leader bicycle spoke. Other manufacturer specifications are left blank until an official model source can be verified.",
			specification: product.SpokesProductSpecification{
				SchemaVersion: 1, ModelName: "CX-Leader", SourceURLs: []string{spokesOverviewSource},
			},
		},
	}
}

func createOrVerifySAPIMSpokeProduct(productService *service.ProductService, databaseConnection *gorm.DB, seed verifiedSAPIMSpokeProductSeed, brandID, categoryID uint) (*product.Product, bool, error) {
	expectedShortDesc, err := safehtml.Sanitize(seed.shortDesc)
	if err != nil {
		return nil, false, fmt.Errorf("sanitize %s short description: %w", seed.specification.ModelName, err)
	}
	expectedDescription, err := safehtml.Sanitize(seed.description)
	if err != nil {
		return nil, false, fmt.Errorf("sanitize %s description: %w", seed.specification.ModelName, err)
	}
	var existing product.Product
	err = databaseConnection.Where("slug = ? AND locale = ?", seed.slug, "en").First(&existing).Error
	if err == nil {
		var variantCount int64
		if err := databaseConnection.Model(&product.ProductVariant{}).Where("product_id = ?", existing.ID).Count(&variantCount).Error; err != nil {
			return nil, false, fmt.Errorf("count variants for existing %s: %w", seed.slug, err)
		}
		matches := existing.Name == seed.name && existing.ShortDesc == expectedShortDesc && existing.Description == expectedDescription &&
			existing.ProductType == product.ProductTypeSpoke && existing.Status == "inactive" &&
			existing.BrandID != nil && *existing.BrandID == brandID &&
			existing.ProductCategoryID != nil && *existing.ProductCategoryID == categoryID &&
			existing.ProductSpecificationTemplateID == nil && existing.SKU == "" && variantCount == 0 &&
			existing.MerchantProductIdentifier == "" && existing.ManufacturerProductIdentifier == "" &&
			existing.StandardProductIdentifier == "" && equalSAPIMSpokeSeedSpecification(existing.SpokeSpecification, &seed.specification)
		if !matches {
			return nil, false, fmt.Errorf("product slug %q already exists with different data; refusing to overwrite it", seed.slug)
		}
		return &existing, true, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, fmt.Errorf("check existing product slug %q: %w", seed.slug, err)
	}

	spokeType := product.ProductTypeSpoke
	created, err := productService.CreateAdminProduct(service.ProductCreateInput{
		ProductType:        &spokeType,
		SpokeSpecification: &seed.specification,
		ProductCategoryID:  &categoryID,
		BrandID:            &brandID,
		Name:               seed.name,
		Slug:               seed.slug,
		ShortDesc:          seed.shortDesc,
		Description:        seed.description,
		Status:             "inactive",
		Locale:             "en",
	})
	if err != nil {
		return nil, false, fmt.Errorf("create %s product: %w", seed.specification.ModelName, err)
	}
	return created, false, nil
}

func ensureSAPIMSpokeProductWithdrawalEvent(databaseConnection *gorm.DB, merchantPublisher *service.MerchantOutboxPublisher, productID uint) error {
	aggregateID := fmt.Sprint(productID)
	var existingEvents int64
	err := databaseConnection.Model(&outbox.Event{}).
		Where("event_type = ? AND aggregate_type = ? AND aggregate_id = ?", outbox.EventTypeMerchantProductWithdraw, outbox.AggregateTypeProduct, aggregateID).
		Count(&existingEvents).Error
	if err != nil {
		return fmt.Errorf("check Merchant withdrawal event for spoke product %d: %w", productID, err)
	}
	if existingEvents > 0 {
		return nil
	}
	if err := merchantPublisher.EnqueueProductWithdraw(productID, "product_created"); err != nil {
		return fmt.Errorf("enqueue Merchant withdrawal event for inactive spoke product %d: %w", productID, err)
	}
	return nil
}

func equalSAPIMSpokeSeedSpecification(actual, expected *product.SpokesProductSpecification) bool {
	if actual == nil || expected == nil {
		return actual == expected
	}
	actualJSON, actualErr := json.Marshal(actual)
	expectedJSON, expectedErr := json.Marshal(expected)
	return actualErr == nil && expectedErr == nil && string(actualJSON) == string(expectedJSON)
}

func spokeMeasurementPointer(value float64) *float64 {
	return &value
}
