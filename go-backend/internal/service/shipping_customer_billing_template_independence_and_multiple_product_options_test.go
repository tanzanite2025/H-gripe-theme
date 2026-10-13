package service

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	productdomain "commerce-platform/internal/domain/product"
	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"
)

func createIndependentCustomerShippingTemplateForTest(t *testing.T, db *gorm.DB, name string, fee int64, countries string) shippingdomain.ShippingTemplate {
	t.Helper()
	carrier := seedQuoteCarrier(t, db, name+" carrier", name)
	route := shippingdomain.CarrierService{CarrierID: carrier.ID, ServiceCode: name, ServiceName: name, Countries: countries, Currency: "CNY", BillingMode: "volumetric_weight", FirstWeightGrams: 9999, AdditionalWeightGrams: 9999, MinChargeWeightGrams: 99999, VolumetricDivisor: 1, FuelSurchargePercentDecimal: "95", RemoteSurchargeMinor: 99999, Enabled: true}
	require.NoError(t, db.Create(&route).Error)
	aggregate := shippingdomain.ShippingRouteGroup{Name: name + " aggregate", Enabled: true}
	repo := repository.NewShippingRepository(db)
	require.NoError(t, repo.CreateShippingRouteGroup(&aggregate, []uint{route.ID}))
	template := shippingdomain.ShippingTemplate{Name: name, Type: "quantity", Currency: "USD", Enabled: true, RouteGroupID: &aggregate.ID, CustomerBillingMode: "actual_weight", DefaultFeeMinor: fee}
	require.NoError(t, repo.CreateTemplateWithRules(&template, nil))
	return template
}

func TestCustomerShippingPriceNeverReadsCarrierInternalCosts(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := createIndependentCustomerShippingTemplateForTest(t, db, "independent", 1200, `["US"]`)
	item, variant := seedQuoteProduct(t, db, 50, 800, template.ID)
	input := ShippingQuoteInput{Country: "US", Currency: "USD", Items: []ShippingQuoteItemInput{{ProductID: item.ID, VariantID: &variant.ID, Quantity: 1}}}
	first, err := shippingService.QuoteCart(input)
	require.NoError(t, err)
	assert.Equal(t, int64(1200), first.ShippingFeeMinor)
	assert.Equal(t, "actual_weight", first.SelectedPlan.Legs[0].BillingMode)
	require.NoError(t, db.Model(&shippingdomain.CarrierService{}).Where("service_code = ?", "independent").Updates(map[string]interface{}{"remote_surcharge_minor": 555555, "fuel_surcharge_percent_decimal": "99", "min_charge_weight_grams": 55555}).Error)
	second, err := shippingService.QuoteCart(input)
	require.NoError(t, err)
	assert.Equal(t, first.ShippingFeeMinor, second.ShippingFeeMinor)
	assert.Equal(t, first.SelectedPlan.Legs[0].BillableWeightGrams, second.SelectedPlan.Legs[0].BillableWeightGrams)
}

func TestProductHasTwoCustomerShippingOptionsWithoutDoubleCharging(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	first := createIndependentCustomerShippingTemplateForTest(t, db, "standard", 1000, `["US"]`)
	second := createIndependentCustomerShippingTemplateForTest(t, db, "express", 2500, `["US"]`)
	item, variant := seedQuoteProduct(t, db, 50, 800, first.ID, second.ID)
	quote, err := shippingService.QuoteCart(ShippingQuoteInput{Country: "US", Currency: "USD", Items: []ShippingQuoteItemInput{{ProductID: item.ID, VariantID: &variant.ID, Quantity: 1}}})
	require.NoError(t, err)
	require.Len(t, quote.Plans, 2)
	assert.Equal(t, int64(1000), quote.Plans[0].ShippingFeeMinor)
	assert.Equal(t, int64(2500), quote.Plans[1].ShippingFeeMinor)
	for _, plan := range quote.Plans {
		require.Len(t, plan.Legs, 1)
		assert.Equal(t, []int{0}, plan.Legs[0].ItemIndexes)
	}
}

func TestSharedCustomerTemplateCanCoverOnlySelectedCartItems(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	shared := createIndependentCustomerShippingTemplateForTest(t, db, "shared", 1000, `["US"]`)
	alternate := createIndependentCustomerShippingTemplateForTest(t, db, "alternate", 2000, `["US"]`)
	require.NoError(t, db.Model(&shippingdomain.ShippingTemplate{}).Where("id = ?", shared.ID).Update("type", "quantity").Error)
	require.NoError(t, db.Create(&shippingdomain.ShippingRule{TemplateID: shared.ID, Region: "US", Currency: "USD", MinValue: 0, MaxValue: 0, FeeMinor: 1000, AdditionalMinor: 500}).Error)
	first, firstVariant := seedQuoteProduct(t, db, 50, 800, shared.ID, alternate.ID)
	second, secondVariant := seedQuoteProduct(t, db, 50, 800, shared.ID, alternate.ID)
	quote, err := shippingService.QuoteCart(ShippingQuoteInput{Country: "US", Currency: "USD", Items: []ShippingQuoteItemInput{{ProductID: first.ID, VariantID: &firstVariant.ID, Quantity: 1}, {ProductID: second.ID, VariantID: &secondVariant.ID, Quantity: 1}}})
	require.NoError(t, err)
	require.Len(t, quote.Plans, 4)
	fees := []int64{}
	for _, plan := range quote.Plans {
		fees = append(fees, plan.ShippingFeeMinor)
		covered := map[int]int{}
		for _, leg := range plan.Legs {
			for _, index := range leg.ItemIndexes {
				covered[index]++
			}
		}
		assert.Equal(t, map[int]int{0: 1, 1: 1}, covered)
	}
	assert.ElementsMatch(t, []int64{1500, 3000, 3000, 2000}, fees)
}

func TestSKUShippingBindingsOverrideProductAndEmptySelectionInherits(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	first := createIndependentCustomerShippingTemplateForTest(t, db, "product-default", 1000, `["US"]`)
	second := createIndependentCustomerShippingTemplateForTest(t, db, "sku-override", 2200, `["US"]`)
	item, variant := seedQuoteProduct(t, db, 50, 800, first.ID)
	repo := repository.NewProductRepository(db)
	require.NoError(t, repo.ReplaceProductShippingTemplateBindings(item.ID, &variant.ID, []uint{second.ID}))
	ids, err := repo.FindPurchasableShippingTemplateIDs(item.ID, &variant.ID)
	require.NoError(t, err)
	assert.Equal(t, []uint{second.ID}, ids)
	require.NoError(t, repo.ReplaceProductShippingTemplateBindings(item.ID, &variant.ID, nil))
	ids, err = repo.FindPurchasableShippingTemplateIDs(item.ID, &variant.ID)
	require.NoError(t, err)
	assert.Equal(t, []uint{first.ID}, ids)
	var count int64
	require.NoError(t, db.Model(&productdomain.ProductShippingTemplateBinding{}).Where("product_id = ? AND variant_id IS NULL", item.ID).Count(&count).Error)
	assert.Equal(t, int64(1), count)
	quote, err := shippingService.QuoteCart(ShippingQuoteInput{Country: "US", Currency: "USD", Items: []ShippingQuoteItemInput{{ProductID: item.ID, VariantID: &variant.ID, Quantity: 1}}})
	require.NoError(t, err)
	assert.Equal(t, int64(1000), quote.ShippingFeeMinor)
}

func TestSavingCustomerTemplateDoesNotChangeSharedRouteAggregate(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := createIndependentCustomerShippingTemplateForTest(t, db, "shared-identity", 1000, `["US"]`)
	original, err := shippingService.GetShippingRouteGroup(*template.RouteGroupID)
	require.NoError(t, err)
	require.Len(t, original.CarrierServices, 1)
	template.DefaultFeeMinor = 2300
	require.NoError(t, shippingService.UpdateTemplate(&template))
	after, err := shippingService.GetShippingRouteGroup(*template.RouteGroupID)
	require.NoError(t, err)
	require.Len(t, after.CarrierServices, 1)
	assert.Equal(t, original.CarrierServices[0].ID, after.CarrierServices[0].ID)
	assert.Error(t, shippingService.DeleteShippingRouteGroup(*template.RouteGroupID))
}

func TestCustomerShippingAlternativeOutsideCountryDoesNotBlockValidOption(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	allowed := createIndependentCustomerShippingTemplateForTest(t, db, "allowed", 1000, `["US"]`)
	unavailable := createIndependentCustomerShippingTemplateForTest(t, db, "unavailable", 2200, `["CA"]`)
	item, variant := seedQuoteProduct(t, db, 50, 800, allowed.ID, unavailable.ID)
	input := ShippingQuoteInput{Country: "US", Currency: "USD", Items: []ShippingQuoteItemInput{{ProductID: item.ID, VariantID: &variant.ID, Quantity: 1}}}
	quote, err := shippingService.QuoteCart(input)
	require.NoError(t, err)
	require.Len(t, quote.Plans, 1)
	assert.Equal(t, int64(1000), quote.ShippingFeeMinor)
	input.Country = "DE"
	_, err = shippingService.QuoteCart(input)
	assert.ErrorIs(t, err, ErrShippingQuotePlanUnavailable)
}

func TestUnavailableCustomerShippingAlternativeDoesNotBlockCompleteCartQuote(t *testing.T) {
	for _, routeBacked := range []bool{false, true} {
		for _, unavailableReason := range []string{"disabled", "no_matching_rate"} {
			t.Run(fmt.Sprintf("route_backed_%t/%s", routeBacked, unavailableReason), func(t *testing.T) {
				db, shippingService := newTestShippingQuoteService(t)
				valid := createIndependentCustomerShippingTemplateForTest(t, db, "valid", 1000, `["US"]`)
				unavailable := createIndependentCustomerShippingTemplateForTest(t, db, "unavailable", 2200, `["US"]`)
				if !routeBacked {
					require.NoError(t, db.Model(&shippingdomain.ShippingTemplate{}).Where("id IN ?", []uint{valid.ID, unavailable.ID}).Update("route_group_id", nil).Error)
				}
				if unavailableReason == "disabled" {
					require.NoError(t, db.Model(&shippingdomain.ShippingTemplate{}).Where("id = ?", unavailable.ID).Update("enabled", false).Error)
				} else {
					require.NoError(t, db.Model(&shippingdomain.ShippingTemplate{}).Where("id = ?", unavailable.ID).Update("default_fee_minor", 0).Error)
					require.NoError(t, db.Create(&shippingdomain.ShippingRule{TemplateID: unavailable.ID, Region: "US", Currency: "USD", MinValue: 10, MaxValue: 20, FeeMinor: 2200}).Error)
				}
				item, variant := seedQuoteProduct(t, db, 50, 800, unavailable.ID, valid.ID)
				input := ShippingQuoteInput{Country: "US", Currency: "USD", Items: []ShippingQuoteItemInput{{ProductID: item.ID, VariantID: &variant.ID, Quantity: 1}}}
				quote, err := shippingService.QuoteCart(input)
				require.NoError(t, err)
				require.Len(t, quote.Plans, 1)
				assert.Equal(t, int64(1000), quote.ShippingFeeMinor)
				assert.Equal(t, valid.ID, quote.Items[0].TemplateID)
				assert.Equal(t, valid.Name, quote.Items[0].TemplateName)

				// A second cart item with no usable alternative must fail the entire
				// quote, rather than disappear and leave a partial shipping plan.
				blocked, blockedVariant := seedQuoteProduct(t, db, 50, 800, unavailable.ID)
				input.Items = append(input.Items, ShippingQuoteItemInput{ProductID: blocked.ID, VariantID: &blockedVariant.ID, Quantity: 1})
				quote, err = shippingService.QuoteCart(input)
				require.Error(t, err)
				assert.Nil(t, quote)
				if unavailableReason == "disabled" {
					assert.ErrorIs(t, err, ErrShippingQuotePlanUnavailable)
				} else {
					assert.ErrorIs(t, err, ErrShippingRateUnavailable)
				}
			})
		}
	}
}

func TestCustomerShippingAlternativesDoNotHidePricingConfigurationFailures(t *testing.T) {
	for _, failure := range []string{"exchange_rate", "currency", "weight_scale", "billing_mode"} {
		t.Run(failure, func(t *testing.T) {
			db, shippingService := newTestShippingQuoteService(t)
			valid := createIndependentCustomerShippingTemplateForTest(t, db, "valid", 1000, `["US"]`)
			broken := createIndependentCustomerShippingTemplateForTest(t, db, "broken", 2200, `["US"]`)
			switch failure {
			case "exchange_rate":
				require.NoError(t, db.Exec("UPDATE shipping_templates SET currency = 'EUR' WHERE id = ?", broken.ID).Error)
			case "currency":
				require.NoError(t, db.Exec("UPDATE shipping_templates SET currency = 'INVALID' WHERE id = ?", broken.ID).Error)
			case "weight_scale":
				require.NoError(t, db.Model(&shippingdomain.ShippingTemplate{}).Where("id = ?", broken.ID).Update("type", "weight").Error)
				require.NoError(t, db.Create(&shippingdomain.ShippingRule{TemplateID: broken.ID, Region: "US", Currency: "USD", FeeMinor: 2200, AdditionalMinor: 500}).Error)
			case "billing_mode":
				require.NoError(t, db.Exec("UPDATE shipping_templates SET customer_billing_mode = 'INVALID' WHERE id = ?", broken.ID).Error)
			}
			item, variant := seedQuoteProduct(t, db, 50, 800, valid.ID, broken.ID)
			quote, err := shippingService.QuoteCart(ShippingQuoteInput{Country: "US", Currency: "USD", Items: []ShippingQuoteItemInput{{ProductID: item.ID, VariantID: &variant.ID, Quantity: 1}}})
			require.Error(t, err)
			assert.Nil(t, quote)
			if failure == "exchange_rate" {
				assert.ErrorIs(t, err, ErrShippingRateUnavailable)
			}
		})
	}
}

func TestProductAndSKUShippingBindingsCanBeSavedRepeatedlyWithoutDuplicates(t *testing.T) {
	db, _ := newTestShippingQuoteService(t)
	first := createIndependentCustomerShippingTemplateForTest(t, db, "repeat-standard", 1000, `["US"]`)
	second := createIndependentCustomerShippingTemplateForTest(t, db, "repeat-express", 2200, `["US"]`)
	item, variant := seedQuoteProduct(t, db, 50, 800, first.ID)
	repo := repository.NewProductRepository(db)
	for _, scope := range []*uint{nil, &variant.ID} {
		for attempt := 0; attempt < 3; attempt++ {
			require.NoError(t, repo.ReplaceProductShippingTemplateBindings(item.ID, scope, []uint{second.ID, first.ID, second.ID}))
			ids, err := repo.FindPurchasableShippingTemplateIDs(item.ID, scope)
			require.NoError(t, err)
			assert.Equal(t, []uint{second.ID, first.ID}, ids)
		}
		duplicate := productdomain.ProductShippingTemplateBinding{ProductID: item.ID, VariantID: scope, ShippingTemplateID: first.ID, Enabled: true}
		require.Error(t, db.Create(&duplicate).Error, "the database must reject duplicate active bindings in either scope")
	}
	require.Error(t, repo.ReplaceProductShippingTemplateBindings(item.ID, &variant.ID, []uint{999999}))
	ids, err := repo.FindPurchasableShippingTemplateIDs(item.ID, &variant.ID)
	require.NoError(t, err)
	assert.Equal(t, []uint{second.ID, first.ID}, ids, "a failed replacement must preserve the previous selection")
}

func TestShippingBindingStorageFailureCannotReportSuccessfulSave(t *testing.T) {
	db, _ := newTestShippingQuoteService(t)
	item, _ := seedQuoteProduct(t, db, 50, 800)
	repo := repository.NewProductRepository(db)
	require.NoError(t, db.Migrator().DropTable(&productdomain.ProductShippingTemplateBinding{}))
	require.Error(t, repo.ReplaceProductShippingTemplateBindings(item.ID, nil, nil))
	_, err := repo.FindPurchasableShippingTemplateIDs(item.ID, nil)
	require.Error(t, err)
}

func TestDeletingCustomerShippingTemplateRequiresRemovingProductBindings(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := createIndependentCustomerShippingTemplateForTest(t, db, "referenced", 1000, `["US"]`)
	item, _ := seedQuoteProduct(t, db, 50, 800, template.ID)
	require.ErrorContains(t, shippingService.DeleteTemplate(template.ID), "referenced by products or SKUs")
	require.NoError(t, repository.NewProductRepository(db).ReplaceProductShippingTemplateBindings(item.ID, nil, nil))
	require.NoError(t, shippingService.DeleteTemplate(template.ID))
}

func TestRemovingSKUAndProductClearsTheirCustomerShippingTemplateBindings(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	productTemplate := createIndependentCustomerShippingTemplateForTest(t, db, "product binding", 1000, `["US"]`)
	skuTemplate := createIndependentCustomerShippingTemplateForTest(t, db, "SKU binding", 2000, `["US"]`)
	item, variant := seedQuoteProduct(t, db, 50, 800, productTemplate.ID)
	productRepo := repository.NewProductRepository(db)
	require.NoError(t, productRepo.ReplaceProductShippingTemplateBindings(item.ID, &variant.ID, []uint{skuTemplate.ID}))
	require.Error(t, shippingService.DeleteTemplate(skuTemplate.ID))
	require.NoError(t, productRepo.UpdateWithSpecValuesAndVariants(&item, nil, false, nil, true))
	require.NoError(t, shippingService.DeleteTemplate(skuTemplate.ID))
	require.Error(t, shippingService.DeleteTemplate(productTemplate.ID))

	require.NoError(t, productRepo.Delete(item.ID))
	require.NoError(t, shippingService.DeleteTemplate(productTemplate.ID))
	var activeBindingCount int64
	require.NoError(t, db.Model(&productdomain.ProductShippingTemplateBinding{}).Where("product_id = ?", item.ID).Count(&activeBindingCount).Error)
	assert.Zero(t, activeBindingCount)
}

func TestDisabledRouteAggregateRejectsNewAndPreviouslyLockedShippingQuotes(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := createIndependentCustomerShippingTemplateForTest(t, db, "disable-aggregate", 1000, `["US"]`)
	item, variant := seedQuoteProduct(t, db, 50, 800, template.ID)
	input := ShippingQuoteInput{Country: "US", Currency: "USD", Items: []ShippingQuoteItemInput{{ProductID: item.ID, VariantID: &variant.ID, Quantity: 1}}}
	quote, err := shippingService.QuoteCart(input)
	require.NoError(t, err)
	require.NoError(t, db.Model(&shippingdomain.ShippingRouteGroup{}).Where("id = ?", *template.RouteGroupID).Update("enabled", false).Error)
	_, err = shippingService.QuoteCart(input)
	require.ErrorIs(t, err, ErrShippingQuotePlanUnavailable)
	input.ShippingQuoteID = quote.ID
	input.SelectedQuotePlanID = quote.SelectedPlan.ID
	_, err = shippingService.QuoteCart(input)
	require.ErrorIs(t, err, ErrShippingQuoteStale)
}

func TestLockedShippingQuoteChecksOnlyTheSelectedAlternative(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	selected := createIndependentCustomerShippingTemplateForTest(t, db, "lock-standard", 1000, `["US"]`)
	alternative := createIndependentCustomerShippingTemplateForTest(t, db, "lock-express", 2200, `["US"]`)
	item, variant := seedQuoteProduct(t, db, 50, 800, selected.ID, alternative.ID)
	input := ShippingQuoteInput{Country: "US", Currency: "USD", Items: []ShippingQuoteItemInput{{ProductID: item.ID, VariantID: &variant.ID, Quantity: 1}}}
	quote, err := shippingService.QuoteCart(input)
	require.NoError(t, err)
	require.Len(t, quote.Plans, 2)
	require.NoError(t, db.Model(&shippingdomain.ShippingTemplate{}).Where("id = ?", alternative.ID).Update("enabled", false).Error)
	input.ShippingQuoteID = quote.ID
	input.SelectedQuotePlanID = quote.Plans[0].ID
	restored, err := shippingService.QuoteCart(input)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), restored.ShippingFeeMinor)
	input.SelectedQuotePlanID = quote.Plans[1].ID
	_, err = shippingService.QuoteCart(input)
	require.ErrorIs(t, err, ErrShippingQuoteStale)
}

func TestOneCarrierRouteCanServeIndependentAggregatesAndCustomerPrices(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	first := createIndependentCustomerShippingTemplateForTest(t, db, "shared-provider", 1000, `["US"]`)
	shippingRepo := repository.NewShippingRepository(db)
	group, err := shippingRepo.FindShippingRouteGroupByID(*first.RouteGroupID)
	require.NoError(t, err)
	require.Len(t, group.CarrierServices, 1)
	secondGroup := shippingdomain.ShippingRouteGroup{Name: "same provider, different aggregate", Enabled: true}
	require.NoError(t, shippingRepo.CreateShippingRouteGroup(&secondGroup, []uint{group.CarrierServices[0].ID}))
	second := shippingdomain.ShippingTemplate{Name: "separate customer price", Type: "quantity", Currency: "USD", Enabled: true, RouteGroupID: &secondGroup.ID, DefaultFeeMinor: 2300}
	require.NoError(t, shippingService.CreateTemplate(&second))
	item, variant := seedQuoteProduct(t, db, 50, 800, first.ID, second.ID)
	quote, err := shippingService.QuoteCart(ShippingQuoteInput{Country: "US", Currency: "USD", Items: []ShippingQuoteItemInput{{ProductID: item.ID, VariantID: &variant.ID, Quantity: 1}}})
	require.NoError(t, err)
	require.Len(t, quote.Plans, 2)
	assert.Equal(t, int64(1000), quote.Plans[0].ShippingFeeMinor)
	assert.Equal(t, int64(2300), quote.Plans[1].ShippingFeeMinor)
	assert.Equal(t, quote.Plans[0].Legs[0].CarrierServiceID, quote.Plans[1].Legs[0].CarrierServiceID)
}

func TestAdminProductCreateNewSKUUpdateAndTranslationCopyPreserveShippingSelections(t *testing.T) {
	db, productService := newTestProductService(t)
	require.NoError(t, db.AutoMigrate(&shippingdomain.ShippingTemplate{}))
	productSpecificationTemplate := seedCarbonRimType(t, db)
	first := shippingdomain.ShippingTemplate{Name: "product selection", Type: "quantity", Currency: "USD", Enabled: true}
	second := shippingdomain.ShippingTemplate{Name: "SKU selection", Type: "quantity", Currency: "USD", Enabled: true}
	require.NoError(t, db.Create(&first).Error)
	require.NoError(t, db.Create(&second).Error)
	created, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductSpecificationTemplateID: &productSpecificationTemplate.ID,
		Name:                           "shipping selections", Slug: "shipping-selections", Status: "active", Locale: "en", ShippingTemplateIDs: []uint{first.ID, second.ID},
		SpecValues: map[string]string{"outer_width_mm": "30"},
		Variants:   []ProductVariantInput{{SKU: "SHIPPING-SELECTIONS-DEFAULT", OptionValues: map[string]string{"brake_type": "disc"}, PriceMinor: 5000, Stock: 10, IsDefault: true, IsActive: boolPtr(true)}},
	})
	require.NoError(t, err)
	require.Len(t, created.Variants, 1)
	updated, err := productService.UpdateAdminProduct(created.ID, ProductUpdateInput{
		UpdateVariants: true,
		Variants: []ProductVariantInput{
			{ID: &created.Variants[0].ID, SKU: created.Variants[0].SKU, OptionValues: map[string]string{"brake_type": "disc"}, PriceMinor: 5000, Stock: 10, IsDefault: true, IsActive: boolPtr(true)},
			{SKU: "SHIPPING-SELECTIONS-NEW", OptionValues: map[string]string{"brake_type": "rim"}, PriceMinor: 6000, Stock: 10, IsActive: boolPtr(true), ShippingTemplateIDs: []uint{second.ID}},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, []uint{first.ID, second.ID}, updated.ShippingTemplateIDs)
	require.Len(t, updated.Variants, 2)
	var newVariantID uint
	for _, variant := range updated.Variants {
		if variant.SKU == "SHIPPING-SELECTIONS-NEW" {
			newVariantID = variant.ID
			assert.Equal(t, []uint{second.ID}, variant.ShippingTemplateIDs)
		}
	}
	require.NotZero(t, newVariantID)
	translated, _, err := productService.CopyAdminProductTranslation(created.ID, "fr")
	require.NoError(t, err)
	assert.Equal(t, []uint{first.ID, second.ID}, translated.ShippingTemplateIDs)
	require.Len(t, translated.Variants, 2)
	for _, variant := range translated.Variants {
		if variant.MasterVariantID != nil && *variant.MasterVariantID == newVariantID {
			assert.Equal(t, []uint{second.ID}, variant.ShippingTemplateIDs)
		}
	}
}

func TestSavingWeightRulesRequiresCustomerWeightStepsAtEveryWriteEntry(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	template := createIndependentCustomerShippingTemplateForTest(t, db, "rule-validation", 1000, `["US"]`)
	template.Type = "weight"
	require.NoError(t, shippingService.UpdateTemplate(&template))
	rule := shippingdomain.ShippingRule{Region: "US", Currency: "USD", FeeMinor: 1000, AdditionalMinor: 100}
	require.ErrorIs(t, shippingService.CreateTemplateRule(template.ID, &rule), ErrShippingRateConfigurationInvalid)
	rule.AdditionalMinor = 0
	require.NoError(t, shippingService.CreateTemplateRule(template.ID, &rule))
	rule.AdditionalMinor = 100
	require.ErrorIs(t, shippingService.UpdateTemplateRule(template.ID, &rule), ErrShippingRateConfigurationInvalid)
	template.Rules = []shippingdomain.ShippingRule{rule}
	require.ErrorIs(t, shippingService.UpdateTemplate(&template), ErrShippingRateConfigurationInvalid)
	template.CustomerFirstWeightGrams = 500
	template.CustomerAdditionalWeightGrams = 100
	require.NoError(t, shippingService.UpdateTemplate(&template))
	require.NoError(t, shippingService.CreateTemplateRule(template.ID, &shippingdomain.ShippingRule{Region: "CA", Currency: "USD", FeeMinor: 1000, AdditionalMinor: 100}))
}

func TestCreatingCustomerTemplateDoesNotWriteSharedRouteAssociations(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	first := createIndependentCustomerShippingTemplateForTest(t, db, "shared-association", 1000, `["US"]`)
	group, err := shippingService.GetShippingRouteGroup(*first.RouteGroupID)
	require.NoError(t, err)
	originalName := group.Name
	group.Name = "must not be written by customer template creation"
	group.CarrierServices[0].ServiceName = "must not overwrite the published identity"
	second := shippingdomain.ShippingTemplate{Name: "second customer template", Type: "quantity", Currency: "USD", Enabled: true, RouteGroupID: first.RouteGroupID, RouteGroup: group, DefaultFeeMinor: 2000}
	require.NoError(t, shippingService.CreateTemplate(&second))
	unchanged, err := shippingService.GetShippingRouteGroup(*first.RouteGroupID)
	require.NoError(t, err)
	assert.Equal(t, originalName, unchanged.Name)
	require.Len(t, unchanged.CarrierServices, 1)
	assert.Equal(t, "shared-association", unchanged.CarrierServices[0].ServiceName)
}

func TestCreatingDisabledRouteAggregateCannotEnableCustomerCheckout(t *testing.T) {
	db, shippingService := newTestShippingQuoteService(t)
	first := createIndependentCustomerShippingTemplateForTest(t, db, "disabled-new-aggregate", 1000, `["US"]`)
	group, err := shippingService.GetShippingRouteGroup(*first.RouteGroupID)
	require.NoError(t, err)
	require.Len(t, group.CarrierServices, 1)
	disabled := shippingdomain.ShippingRouteGroup{Name: "created as disabled", Enabled: false}
	require.NoError(t, shippingService.CreateShippingRouteGroup(&disabled, []uint{group.CarrierServices[0].ID}))
	stored, err := shippingService.GetShippingRouteGroup(disabled.ID)
	require.NoError(t, err)
	assert.False(t, stored.Enabled)
	first.RouteGroupID = &disabled.ID
	require.NoError(t, shippingService.UpdateTemplate(&first))
	item, variant := seedQuoteProduct(t, db, 50, 800, first.ID)
	_, err = shippingService.QuoteCart(ShippingQuoteInput{Country: "US", Currency: "USD", Items: []ShippingQuoteItemInput{{ProductID: item.ID, VariantID: &variant.ID, Quantity: 1}}})
	require.ErrorIs(t, err, ErrShippingQuotePlanUnavailable)
}
