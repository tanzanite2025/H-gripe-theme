package service

import (
	"testing"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
)

func TestCarbonRimCreationRejectsMissingWrongDisabledAndMisplacedSystemCategories(t *testing.T) {
	for _, scenario := range []string{"missing", "zero", "wrong", "disabled", "wrong-parent", "wrong-depth", "disabled-parent"} {
		t.Run(scenario, func(t *testing.T) {
			db, svc := newTestProductService(t)
			input := createCarbonRimProductServiceTestInput(t, db)
			switch scenario {
			case "missing":
				input.ProductCategoryID = nil
			case "zero":
				input.ProductCategoryID = new(uint)
			case "wrong":
				other := seedProductCategoryForProductServiceTest(t, db, "Other", "other", nil)
				input.ProductCategoryID = &other.ID
			case "disabled":
				require.NoError(t, db.Model(&product.ProductCategory{}).Where("id = ?", input.ProductCategoryID).Update("is_enabled", false).Error)
			case "wrong-parent":
				other := seedProductCategoryForProductServiceTest(t, db, "Other", "other", nil)
				require.NoError(t, db.Model(&product.ProductCategory{}).Where("id = ?", input.ProductCategoryID).Update("parent_id", other.ID).Error)
			case "wrong-depth":
				require.NoError(t, db.Model(&product.ProductCategory{}).Where("id = ?", input.ProductCategoryID).Update("depth", 3).Error)
			case "disabled-parent":
				require.NoError(t, db.Model(&product.ProductCategory{}).Where("slug = ?", SystemProductCategoryWheelComponentsSlug).Update("is_enabled", false).Error)
			}
			_, err := svc.CreateAdminProduct(input)
			require.ErrorIs(t, err, ErrProductCategoryInvalid)
			var count int64
			require.NoError(t, db.Model(&product.Product{}).Count(&count).Error)
			require.Zero(t, count)
		})
	}
}

func TestCarbonRimCategoryBindingRejectsGenericProductsClearingMovingAndMalformedTranslationCopies(t *testing.T) {
	db, svc := newTestProductService(t)
	input := createCarbonRimProductServiceTestInput(t, db)
	created, err := svc.CreateAdminProduct(input)
	require.NoError(t, err)
	other := seedProductCategoryForProductServiceTest(t, db, "Other", "other", nil)
	for _, categoryID := range []*uint{nil, &other.ID} {
		_, err = svc.UpdateAdminProduct(created.ID, ProductUpdateInput{UpdateProductCategoryID: true, ProductCategoryID: categoryID})
		require.ErrorIs(t, err, ErrProductCategoryInvalid)
	}
	_, err = svc.CreateAdminProduct(ProductCreateInput{Name: "Generic", Slug: "generic", Locale: "en", Currency: "USD", ProductCategoryID: input.ProductCategoryID})
	require.ErrorIs(t, err, ErrProductCategoryInvalid)
	ordinary := product.Product{Name: "Generic", Slug: "ordinary", Locale: "en", Currency: "USD"}
	require.NoError(t, db.Create(&ordinary).Error)
	_, err = svc.UpdateAdminProduct(ordinary.ID, ProductUpdateInput{UpdateProductCategoryID: true, ProductCategoryID: input.ProductCategoryID})
	require.ErrorIs(t, err, ErrProductCategoryInvalid)
	require.NoError(t, db.Model(&product.Product{}).Where("id = ?", created.ID).Update("product_category_id", other.ID).Error)
	_, _, err = svc.CopyAdminProductTranslation(created.ID, "fr")
	require.ErrorIs(t, err, ErrProductCategoryInvalid)
}

func TestCarbonRimSystemCategoryProtectsIdentityAndLocalizesActualBreadcrumbRelationship(t *testing.T) {
	svc, db := newProductCategoryProtectionTestService(t)
	root := seedProductCategoryForProductServiceTest(t, db, "Wheel Components", SystemProductCategoryWheelComponentsSlug, nil)
	category := seedProductCategoryForProductServiceTest(t, db, "Carbon Rim", SystemProductCategoryCarbonRimSlug, &root.ID)
	for _, input := range []ProductCategoryInput{
		{Name: category.Name, Slug: "renamed", ParentID: &root.ID, IsEnabled: true},
		{Name: category.Name, Slug: category.Slug, ParentID: &root.ID, IsEnabled: false},
		{Name: category.Name, Slug: category.Slug, IsEnabled: true},
	} {
		_, err := svc.Update(category.ID, input)
		require.ErrorIs(t, err, ErrProductCategorySystemProtected)
	}
	require.ErrorIs(t, svc.Delete(category.ID), ErrProductCategorySystemProtected)
	updated, err := svc.Update(category.ID, ProductCategoryInput{Name: "Carbon Fiber Rims", Slug: category.Slug, ParentID: &root.ID, Description: "Updated copy", IsEnabled: true})
	require.NoError(t, err)
	require.Equal(t, "Carbon Fiber Rims", updated.Name)
	require.NoError(t, db.Create(&[]product.ProductCategoryTranslation{
		{ProductCategoryID: root.ID, Locale: "zh_cn", Name: "车轮配件"},
		{ProductCategoryID: category.ID, Locale: "zh_cn", Name: "碳纤维车圈"},
	}).Error)
	breadcrumb, err := NewProductCategoryService(repository.NewProductCategoryRepository(db)).BuildProductBreadcrumb(product.Product{ID: 41, Name: "R45", Slug: "r45", ProductCategoryID: &category.ID}, "zh_cn")
	require.NoError(t, err)
	require.Equal(t, "ready", breadcrumb.Status)
	require.Len(t, breadcrumb.Items, 5)
	require.Equal(t, "车轮配件", breadcrumb.Items[2].Name)
	require.Equal(t, "碳纤维车圈", breadcrumb.Items[3].Name)
	require.Equal(t, "/zh_cn/shop/wheel-components/carbon-rim", breadcrumb.Items[3].Path)
}

func TestCarbonRimUpdatesAndTranslationCopiesRequireCompleteWheelSizeMaterialAndRidingUses(t *testing.T) {
	for _, missingField := range []string{"wheel_size", "material", "riding_uses"} {
		t.Run(missingField, func(t *testing.T) {
			db, svc := newTestProductService(t)
			created, err := svc.CreateAdminProduct(createCarbonRimProductServiceTestInput(t, db))
			require.NoError(t, err)
			incomplete := *created.CarbonRimSpecification
			switch missingField {
			case "wheel_size":
				incomplete.WheelSize = ""
			case "material":
				incomplete.Material = ""
				incomplete.MaterialDetails = ""
			case "riding_uses":
				incomplete.RidingUses = nil
			}
			_, err = svc.UpdateAdminProduct(created.ID, ProductUpdateInput{UpdateCarbonRimSpecification: true, CarbonRimSpecification: &incomplete})
			require.ErrorIs(t, err, ErrProductSpecInvalid)
			require.NoError(t, db.Model(&product.Product{}).Where("id = ?", created.ID).Select("CarbonRimSpecification").Updates(&product.Product{CarbonRimSpecification: &incomplete}).Error)
			_, err = svc.UpdateAdminProduct(created.ID, ProductUpdateInput{UpdateCarbonRimSpecification: true, CarbonRimSpecification: &incomplete})
			require.ErrorIs(t, err, ErrProductSpecInvalid)
			_, err = svc.UpdateAdminProduct(created.ID, ProductUpdateInput{})
			require.ErrorIs(t, err, ErrProductSpecInvalid)
			_, _, err = svc.CopyAdminProductTranslation(created.ID, "fr")
			require.ErrorIs(t, err, ErrProductSpecInvalid)
			_, err = svc.UpdateAdminProduct(created.ID, ProductUpdateInput{UpdateCarbonRimSpecification: true, CarbonRimSpecification: created.CarbonRimSpecification})
			require.NoError(t, err)
		})
	}
}
