package service

import (
	"errors"
	"testing"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestProductInformationTemplateServiceRejectsDuplicateSlugPerKindAndLocale(t *testing.T) {
	db, templateService := newTestProductInformationTemplateService(t)

	_, err := templateService.Create(ProductInformationTemplateInput{
		Kind:      product.ProductInformationTemplateKindAfterSales,
		Name:      "Standard After-sales",
		Slug:      "standard",
		Content:   "<p>Standard content</p>",
		Locale:    "en",
		IsEnabled: true,
	})
	require.NoError(t, err)

	duplicate, err := templateService.Create(ProductInformationTemplateInput{
		Kind:      product.ProductInformationTemplateKindAfterSales,
		Name:      "Duplicate After-sales",
		Slug:      "standard",
		Content:   "<p>Duplicate content</p>",
		Locale:    "en",
		IsEnabled: true,
	})
	require.Nil(t, duplicate)
	require.True(t, errors.Is(err, ErrProductInformationTemplateSlugExists))

	var count int64
	require.NoError(t, db.Model(&product.ProductInformationTemplate{}).Count(&count).Error)
	require.Equal(t, int64(1), count)
}

func TestProductInformationTemplateServiceKeepsKindImmutable(t *testing.T) {
	_, templateService := newTestProductInformationTemplateService(t)

	template, err := templateService.Create(ProductInformationTemplateInput{
		Kind:      product.ProductInformationTemplateKindAfterSales,
		Name:      "Standard After-sales",
		Slug:      "standard-after-sales",
		Content:   "<p>Standard content</p>",
		Locale:    "en",
		IsEnabled: true,
	})
	require.NoError(t, err)

	updated, err := templateService.Update(template.ID, ProductInformationTemplateInput{
		Kind:      product.ProductInformationTemplateKindPackaging,
		Name:      "Changed Packaging",
		Slug:      "changed-packaging",
		Content:   "<p>Changed content</p>",
		Locale:    "en",
		IsEnabled: true,
	})
	require.Nil(t, updated)
	require.True(t, errors.Is(err, ErrProductInformationTemplateInvalid))

	unchanged, err := templateService.Get(template.ID)
	require.NoError(t, err)
	require.Equal(t, product.ProductInformationTemplateKindAfterSales, unchanged.Kind)
}

func TestProductInformationTemplateServiceRequiresSupportedLocale(t *testing.T) {
	_, templateService := newTestProductInformationTemplateService(t)

	template, err := templateService.Create(ProductInformationTemplateInput{
		Kind:      product.ProductInformationTemplateKindAfterSales,
		Name:      "Unsupported Locale",
		Slug:      "unsupported-locale",
		Content:   "<p>Unsupported content</p>",
		Locale:    "xx",
		IsEnabled: true,
	})
	require.Nil(t, template)
	require.True(t, errors.Is(err, ErrUnsupportedLocale))
}

func TestProductInformationTemplateServiceKeepsLocaleImmutable(t *testing.T) {
	_, templateService := newTestProductInformationTemplateService(t)

	template, err := templateService.Create(ProductInformationTemplateInput{
		Kind:      product.ProductInformationTemplateKindPackaging,
		Name:      "Standard Packaging",
		Slug:      "standard-packaging",
		Content:   "<p>Standard content</p>",
		Locale:    "en",
		IsEnabled: true,
	})
	require.NoError(t, err)

	updated, err := templateService.Update(template.ID, ProductInformationTemplateInput{
		Kind:      product.ProductInformationTemplateKindPackaging,
		Name:      "Standard Packaging",
		Slug:      "standard-packaging",
		Content:   "<p>Changed content</p>",
		Locale:    "fr",
		IsEnabled: true,
	})
	require.Nil(t, updated)
	require.True(t, errors.Is(err, ErrProductInformationTemplateInvalid))

	unchanged, err := templateService.Get(template.ID)
	require.NoError(t, err)
	require.Equal(t, "en", unchanged.Locale)
}

func TestProductInformationTemplateServiceAcceptsSameLocaleAliasOnUpdate(t *testing.T) {
	_, templateService := newTestProductInformationTemplateService(t)

	template, err := templateService.Create(ProductInformationTemplateInput{
		Kind:      product.ProductInformationTemplateKindPackaging,
		Name:      "Alias Packaging",
		Slug:      "alias-packaging",
		Content:   "<p>Standard content</p>",
		Locale:    "en",
		IsEnabled: true,
	})
	require.NoError(t, err)

	updated, err := templateService.Update(template.ID, ProductInformationTemplateInput{
		Kind:      product.ProductInformationTemplateKindPackaging,
		Name:      "Alias Packaging",
		Slug:      "alias-packaging",
		Content:   "<p>Changed content</p>",
		Locale:    "en-US",
		IsEnabled: true,
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	require.Equal(t, "en", updated.Locale)
}

func TestProductInformationTemplateServicePersistsDisabledTemplates(t *testing.T) {
	_, templateService := newTestProductInformationTemplateService(t)

	template, err := templateService.Create(ProductInformationTemplateInput{
		Kind:      product.ProductInformationTemplateKindPackaging,
		Name:      "Draft Packaging",
		Slug:      "draft-packaging",
		Content:   "<p>Draft content</p>",
		Locale:    "en",
		IsEnabled: false,
	})
	require.NoError(t, err)
	require.False(t, template.IsEnabled)

	enabledTemplates, err := templateService.List(product.ProductInformationTemplateKindPackaging, "en", false)
	require.NoError(t, err)
	require.Empty(t, enabledTemplates)

	allTemplates, err := templateService.List(product.ProductInformationTemplateKindPackaging, "en", true)
	require.NoError(t, err)
	require.Len(t, allTemplates, 1)
	require.False(t, allTemplates[0].IsEnabled)
}

func newTestProductInformationTemplateService(t *testing.T) (*gorm.DB, *ProductInformationTemplateService) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	require.NoError(t, db.AutoMigrate(&product.ProductInformationTemplate{}))
	return db, NewProductInformationTemplateService(repository.NewProductInformationTemplateRepository(db))
}
