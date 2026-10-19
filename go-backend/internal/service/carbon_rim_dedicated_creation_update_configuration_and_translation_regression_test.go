package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"commerce-platform/internal/domain/product"
)

func createCarbonRimProductServiceTestInput() ProductCreateInput {
	kind := product.ProductTypeCarbonRim
	tubeless := true
	return ProductCreateInput{
		ProductType: &kind, Name: "Carbon rim 45", Slug: "dedicated-carbon-rim-45", Locale: "en", Status: "active", Currency: "USD",
		CarbonRimSpecification: &product.CarbonRimSpecification{SchemaVersion: 1, Model: "R45", WheelSize: "700c", Material: "carbon_fiber", MaterialDetails: "T700 / T800", RidingUses: []string{"road", "gravel"}, BeadSeatDiameterMM: 622, DepthMM: 45, InternalWidthMM: 25, ExternalWidthMM: 32, BeadStructure: "hooked", TubelessReady: &tubeless},
		Variants: []ProductVariantInput{
			{SKU: "DEDICATED-R45-24", PriceMinor: 19900, Stock: 4, Weight: 450, IsDefault: true, CarbonRimSpecification: &product.CarbonRimVariantSpecification{SpokeHoleCount: 24, DrillingPattern: "uniform", FinishCode: "matte-black", NetWeightGrams: 430}},
			{SKU: "DEDICATED-R45-28", PriceMinor: 20900, Stock: 8, Weight: 455, CarbonRimSpecification: &product.CarbonRimVariantSpecification{SpokeHoleCount: 28, DrillingPattern: "uniform", FinishCode: "matte-black", NetWeightGrams: 435}},
		},
	}
}

func TestDedicatedCarbonRimCreationUpdateRoundTripAndOrderConfigurationEvidence(t *testing.T) {
	_, svc := newTestProductService(t)
	input := createCarbonRimProductServiceTestInput()
	created, err := svc.CreateAdminProduct(input)
	require.NoError(t, err)
	require.Equal(t, product.ProductTypeCarbonRim, created.ProductType)
	require.Nil(t, created.ProductSpecificationTemplateID)
	require.Equal(t, input.CarbonRimSpecification, created.CarbonRimSpecification)
	require.Equal(t, "700c", created.CarbonRimSpecification.WheelSize)
	require.Equal(t, []string{"road", "gravel"}, created.CarbonRimSpecification.RidingUses)
	require.Len(t, created.Variants, 2)
	require.Equal(t, 28, created.Variants[1].CarbonRimSpecification.SpokeHoleCount)
	require.JSONEq(t, `{"rim_configuration":"24H / uniform / matte-black"}`, created.Variants[0].OptionValues)

	newName := "Updated dedicated rim"
	updated, err := svc.UpdateAdminProduct(created.ID, ProductUpdateInput{Name: &newName})
	require.NoError(t, err)
	require.Equal(t, created.CarbonRimSpecification, updated.CarbonRimSpecification)
	require.Equal(t, created.Variants[1].CarbonRimSpecification, updated.Variants[1].CarbonRimSpecification)

	specification := *created.CarbonRimSpecification
	specification.DepthMM = 46
	updated, err = svc.UpdateAdminProduct(created.ID, ProductUpdateInput{UpdateCarbonRimSpecification: true, CarbonRimSpecification: &specification})
	require.NoError(t, err)
	require.Equal(t, 46.0, updated.CarbonRimSpecification.DepthMM)
	variants := input.Variants
	for index := range variants {
		variants[index].ID = &created.Variants[index].ID
		variants[index].CarbonRimSpecification = nil
	}
	variants[0].Stock = 9
	updated, err = svc.UpdateAdminProduct(created.ID, ProductUpdateInput{UpdateVariants: true, Variants: variants})
	require.NoError(t, err)
	require.Equal(t, 9, updated.Variants[0].Stock)
	require.Equal(t, 8, updated.Variants[1].Stock)
	require.NotNil(t, updated.Variants[0].CarbonRimSpecification)
	configuration, err := ResolveProductConfiguration(updated, &updated.Variants[0], nil)
	require.NoError(t, err)
	require.Equal(t, 46.0, configuration.Snapshot.CarbonRimSpecification.DepthMM)
	require.Equal(t, 24, configuration.Snapshot.CarbonRimVariantSpecification.SpokeHoleCount)
	_, err = ResolveProductConfiguration(updated, &updated.Variants[0], []SelectedOption{{GroupSlug: "made-up-option", ValueKeys: []string{"fake"}}})
	require.Error(t, err)
}

func TestDedicatedCarbonRimRejectsDataLossAndCrossProductVariantMutation(t *testing.T) {
	db, svc := newTestProductService(t)
	input := createCarbonRimProductServiceTestInput()
	created, err := svc.CreateAdminProduct(input)
	require.NoError(t, err)

	_, err = svc.UpdateAdminProduct(created.ID, ProductUpdateInput{UpdateCarbonRimSpecification: true})
	require.ErrorIs(t, err, ErrProductSpecInvalid)
	templateID := uint(999)
	_, err = svc.UpdateAdminProduct(created.ID, ProductUpdateInput{ProductSpecificationTemplateID: &templateID, UpdateProductSpecificationTemplateID: true})
	require.ErrorIs(t, err, ErrProductSpecInvalid)
	_, err = svc.UpdateAdminProduct(created.ID, ProductUpdateInput{UpdateVariants: true})
	require.ErrorIs(t, err, ErrProductVariantInvalid)
	for index := range input.Variants {
		input.Variants[index].ID = &created.Variants[index].ID
	}
	foreignID := uint(987654)
	input.Variants[0].ID = &foreignID
	newName := "Must not persist"
	_, err = svc.UpdateAdminProduct(created.ID, ProductUpdateInput{Name: &newName, UpdateVariants: true, Variants: input.Variants})
	require.ErrorIs(t, err, ErrProductVariantInvalid)
	var persisted product.Product
	require.NoError(t, db.First(&persisted, created.ID).Error)
	require.Equal(t, created.Name, persisted.Name)
	input.Variants[0].ID = &created.Variants[0].ID
	input.Variants[0].CarbonRimSpecification.SpokeHoleCount = 32
	_, err = svc.UpdateAdminProduct(created.ID, ProductUpdateInput{UpdateVariants: true, Variants: input.Variants})
	require.ErrorIs(t, err, ErrProductVariantInvalid)
}

func TestDedicatedCarbonRimRejectsInvalidCreationWithoutPersistingProduct(t *testing.T) {
	for _, scenario := range []string{"duplicate-configuration", "wrong-width", "missing-tubeless", "missing-sku-specification", "generic-type", "template-binding", "two-defaults", "missing-wheel-size", "wheel-size-bsd-conflict", "missing-riding-uses", "wrong-material"} {
		t.Run(scenario, func(t *testing.T) {
			db, svc := newTestProductService(t)
			input := createCarbonRimProductServiceTestInput()
			switch scenario {
			case "missing-wheel-size":
				input.CarbonRimSpecification.WheelSize = ""
			case "wheel-size-bsd-conflict":
				input.CarbonRimSpecification.WheelSize = "650b"
			case "missing-riding-uses":
				input.CarbonRimSpecification.RidingUses = nil
			case "wrong-material":
				input.CarbonRimSpecification.Material = "aluminum"
			case "duplicate-configuration":
				input.Variants[1].CarbonRimSpecification.SpokeHoleCount = 24
			case "wrong-width":
				input.CarbonRimSpecification.ExternalWidthMM = 20
			case "missing-tubeless":
				input.CarbonRimSpecification.TubelessReady = nil
			case "missing-sku-specification":
				input.Variants[1].CarbonRimSpecification = nil
			case "generic-type":
				input.ProductType = nil
			case "template-binding":
				id := uint(1)
				input.ProductSpecificationTemplateID = &id
			case "two-defaults":
				input.Variants[1].IsDefault = true
			}
			_, err := svc.CreateAdminProduct(input)
			require.Error(t, err)
			var count int64
			require.NoError(t, db.Model(&product.Product{}).Count(&count).Error)
			require.Zero(t, count)
		})
	}
}

func TestDedicatedCarbonRimTranslationUsesSourceEngineeringFactsAndMasterInventory(t *testing.T) {
	_, svc := newTestProductService(t)
	input := createCarbonRimProductServiceTestInput()
	source, err := svc.CreateAdminProduct(input)
	require.NoError(t, err)
	translated, _, err := svc.CopyAdminProductTranslation(source.ID, "fr")
	require.NoError(t, err)
	require.Equal(t, product.ProductTypeCarbonRim, translated.ProductType)
	require.Equal(t, source.CarbonRimSpecification, translated.CarbonRimSpecification)
	require.Equal(t, source.Variants[0].ID, *translated.Variants[0].MasterVariantID)
	sourceSpecification := *source.CarbonRimSpecification
	sourceSpecification.DepthMM = 47
	for index := range input.Variants {
		input.Variants[index].ID = &source.Variants[index].ID
	}
	input.Variants[0].Stock = 12
	input.Variants[0].CarbonRimSpecification.NetWeightGrams = 432
	_, err = svc.UpdateAdminProduct(source.ID, ProductUpdateInput{UpdateCarbonRimSpecification: true, CarbonRimSpecification: &sourceSpecification, UpdateVariants: true, Variants: input.Variants})
	require.NoError(t, err)
	translated, err = svc.GetAdminProduct(translated.ID)
	require.NoError(t, err)
	require.Equal(t, 47.0, translated.CarbonRimSpecification.DepthMM)
	require.Equal(t, 432.0, translated.Variants[0].CarbonRimSpecification.NetWeightGrams)
	require.Equal(t, 12, translated.Variants[0].Stock)
	changed := *translated.CarbonRimSpecification
	changed.DepthMM = 50
	_, err = svc.UpdateAdminProduct(translated.ID, ProductUpdateInput{UpdateCarbonRimSpecification: true, CarbonRimSpecification: &changed})
	require.ErrorIs(t, err, ErrProductSpecInvalid)
}

func TestDedicatedCarbonRimLateVariantDatabaseFailureRollsBackProductAndSpecifications(t *testing.T) {
	db, svc := newTestProductService(t)
	input := createCarbonRimProductServiceTestInput()
	created, err := svc.CreateAdminProduct(input)
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TRIGGER reject_carbon_rim_variant_update BEFORE UPDATE ON product_variants BEGIN SELECT RAISE(ABORT, 'simulated late variant write failure'); END`).Error)
	for index := range input.Variants {
		input.Variants[index].ID = &created.Variants[index].ID
	}
	input.Variants[0].Stock = 99
	changedSpecification := *created.CarbonRimSpecification
	changedSpecification.DepthMM = 48
	changedName := "Must be rolled back"
	_, err = svc.UpdateAdminProduct(created.ID, ProductUpdateInput{Name: &changedName, UpdateCarbonRimSpecification: true, CarbonRimSpecification: &changedSpecification, UpdateVariants: true, Variants: input.Variants})
	require.ErrorContains(t, err, "simulated late variant write failure")
	persisted, err := svc.GetAdminProduct(created.ID)
	require.NoError(t, err)
	require.Equal(t, created.Name, persisted.Name)
	require.Equal(t, created.CarbonRimSpecification, persisted.CarbonRimSpecification)
	require.Equal(t, created.Variants[0].Stock, persisted.Variants[0].Stock)
}
