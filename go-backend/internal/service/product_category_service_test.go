package service

import (
	"testing"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/locales"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestProductCategoryListPublicLocalizesTreeAndFallsBackToDefaults(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&product.ProductCategory{},
		&product.ProductCategoryTranslation{},
	); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	root := product.ProductCategory{
		Name:        "Wheel Parts",
		Slug:        "wheel-parts",
		Description: "Default wheel parts description",
		Depth:       1,
		IsEnabled:   true,
	}
	if err := db.Create(&root).Error; err != nil {
		t.Fatalf("create root category: %v", err)
	}

	child := product.ProductCategory{
		ParentID:    &root.ID,
		Name:        "Rims",
		Slug:        "rims",
		Description: "Default rims description",
		Depth:       2,
		IsEnabled:   true,
	}
	if err := db.Create(&child).Error; err != nil {
		t.Fatalf("create child category: %v", err)
	}

	if err := db.Create(&product.ProductCategoryTranslation{
		ProductCategoryID: root.ID,
		Locale:            "zh_cn",
		Name:              "轮组部件",
		Description:       "轮组部件说明",
	}).Error; err != nil {
		t.Fatalf("create root translation: %v", err)
	}

	disabled := product.ProductCategory{
		Name:  "Hidden",
		Slug:  "hidden",
		Depth: 1,
	}
	if err := db.Create(&disabled).Error; err != nil {
		t.Fatalf("create disabled category: %v", err)
	}
	if err := db.Model(&product.ProductCategory{}).
		Where("id = ?", disabled.ID).
		Update("is_enabled", false).Error; err != nil {
		t.Fatalf("disable category: %v", err)
	}

	view, err := NewProductCategoryService(repository.NewProductCategoryRepository(db)).ListPublic("zh-CN")
	if err != nil {
		t.Fatalf("list public categories: %v", err)
	}

	if len(view.Tree) != 1 {
		t.Fatalf("expected one public root category, got %d", len(view.Tree))
	}
	if view.Tree[0].Name != "轮组部件" {
		t.Fatalf("expected translated root name, got %q", view.Tree[0].Name)
	}
	if view.Tree[0].RoutePath != "/zh_cn/shop/wheel-parts" {
		t.Fatalf("expected localized root route path, got %q", view.Tree[0].RoutePath)
	}
	if view.Tree[0].Description != "轮组部件说明" {
		t.Fatalf("expected translated root description, got %q", view.Tree[0].Description)
	}
	if len(view.Tree[0].Children) != 1 {
		t.Fatalf("expected one child category, got %d", len(view.Tree[0].Children))
	}
	if view.Tree[0].Children[0].RoutePath != "/zh_cn/shop/wheel-parts/rims" {
		t.Fatalf("expected localized child route path, got %q", view.Tree[0].Children[0].RoutePath)
	}
	if view.Tree[0].Children[0].Name != "Rims" {
		t.Fatalf("expected child name to fall back to default, got %q", view.Tree[0].Children[0].Name)
	}
	if len(view.Flat) != 2 {
		t.Fatalf("expected disabled category to be excluded, got %d flat categories", len(view.Flat))
	}
}

func TestProductCategoryListIncludesTranslationSummary(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&product.ProductCategory{},
		&product.ProductCategoryTranslation{},
	); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	category := product.ProductCategory{
		Name:      "Wheel Parts",
		Slug:      "wheel-parts",
		Depth:     1,
		IsEnabled: true,
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}
	if err := db.Create(&[]product.ProductCategoryTranslation{
		{ProductCategoryID: category.ID, Locale: "en", Name: "Wheel Parts"},
		{ProductCategoryID: category.ID, Locale: "zh_cn", Name: "轮组部件"},
	}).Error; err != nil {
		t.Fatalf("create translations: %v", err)
	}

	view, err := NewProductCategoryService(repository.NewProductCategoryRepository(db)).List(true)
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}
	if len(view.Flat) != 1 {
		t.Fatalf("expected one category, got %d", len(view.Flat))
	}

	categoryView := view.Flat[0]
	if categoryView.TranslationCompleted != 2 {
		t.Fatalf("expected two completed translations, got %d", categoryView.TranslationCompleted)
	}
	if categoryView.TranslationTotal != len(locales.EnabledLocaleCodes()) {
		t.Fatalf(
			"expected translation total %d, got %d",
			len(locales.EnabledLocaleCodes()),
			categoryView.TranslationTotal,
		)
	}
	if len(categoryView.TranslationMissingLocales) != categoryView.TranslationTotal-2 {
		t.Fatalf(
			"expected %d missing translations, got %d",
			categoryView.TranslationTotal-2,
			len(categoryView.TranslationMissingLocales),
		)
	}
}

func TestProductCategorySEOUpdateDoesNotPersistDescriptionAsIntro(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&product.ProductCategory{},
		&product.ProductCategoryTranslation{},
	); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	category := product.ProductCategory{
		Name:        "Road Wheels",
		Slug:        "road-wheels",
		Description: "Legacy category description",
		Depth:       1,
		IsEnabled:   true,
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}

	metaTitle := "Road Wheels | Tanzanite"
	view, err := NewProductCategoryService(repository.NewProductCategoryRepository(db)).
		UpdateSEO(category.ID, "en", &metaTitle, nil, nil)
	if err != nil {
		t.Fatalf("update category SEO: %v", err)
	}
	if view.Intro != "" {
		t.Fatalf("SEO view intro = %q, want empty SEO intro", view.Intro)
	}

	var stored product.ProductCategory
	if err := db.First(&stored, category.ID).Error; err != nil {
		t.Fatalf("reload category: %v", err)
	}
	if stored.SEOIntro != "" {
		t.Fatalf("stored SEO intro = %q, want empty SEO intro", stored.SEOIntro)
	}
	if stored.MetaTitle != metaTitle {
		t.Fatalf("stored meta title = %q, want %q", stored.MetaTitle, metaTitle)
	}
}

func TestBuildProductBreadcrumbUsesEnabledCategoryAncestors(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&product.ProductCategory{},
		&product.ProductCategoryTranslation{},
	); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	root := product.ProductCategory{
		Name:      "Road Wheels",
		Slug:      "road-wheels",
		Depth:     1,
		IsEnabled: true,
	}
	if err := db.Create(&root).Error; err != nil {
		t.Fatalf("create root category: %v", err)
	}
	child := product.ProductCategory{
		ParentID:  &root.ID,
		Name:      "Climbing",
		Slug:      "climbing",
		Depth:     2,
		IsEnabled: true,
	}
	if err := db.Create(&child).Error; err != nil {
		t.Fatalf("create child category: %v", err)
	}
	if err := db.Create(&[]product.ProductCategoryTranslation{
		{ProductCategoryID: root.ID, Locale: "zh_cn", Name: "公路轮组"},
		{ProductCategoryID: child.ID, Locale: "zh_cn", Name: "爬坡"},
	}).Error; err != nil {
		t.Fatalf("create category translations: %v", err)
	}

	categoryService := NewProductCategoryService(repository.NewProductCategoryRepository(db))
	item := product.Product{
		ID:                41,
		Name:              "Hyper 45",
		Slug:              "hyper-45",
		ProductCategoryID: &child.ID,
	}

	view, err := categoryService.BuildProductBreadcrumb(item, "zh-CN")
	if err != nil {
		t.Fatalf("build product breadcrumb: %v", err)
	}
	if view.Status != "ready" || view.Reason != "" {
		t.Fatalf("expected ready breadcrumb, got %#v", view)
	}
	if len(view.Items) != 5 {
		t.Fatalf("expected home, shop, two categories, and product, got %#v", view.Items)
	}
	if view.Items[2].Name != "公路轮组" || view.Items[2].Path != "/zh_cn/shop/road-wheels" {
		t.Fatalf("unexpected root breadcrumb item: %#v", view.Items[2])
	}
	if view.Items[3].Name != "爬坡" || view.Items[3].Path != "/zh_cn/shop/road-wheels/climbing" {
		t.Fatalf("unexpected child breadcrumb item: %#v", view.Items[3])
	}
	if view.Items[4].Path != "/zh_cn/products/hyper-45" {
		t.Fatalf("unexpected product breadcrumb item: %#v", view.Items[4])
	}
}

func TestBuildProductBreadcrumbDoesNotExposePartialDisabledCategoryPath(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&product.ProductCategory{},
		&product.ProductCategoryTranslation{},
	); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	root := product.ProductCategory{
		Name:      "Hidden Wheels",
		Slug:      "hidden-wheels",
		Depth:     1,
		IsEnabled: false,
	}
	if err := db.Create(&root).Error; err != nil {
		t.Fatalf("create root category: %v", err)
	}
	if err := db.Model(&product.ProductCategory{}).
		Where("id = ?", root.ID).
		Update("is_enabled", false).Error; err != nil {
		t.Fatalf("disable root category: %v", err)
	}
	item := product.Product{
		ID:                42,
		Name:              "Hidden Product",
		Slug:              "hidden-product",
		ProductCategoryID: &root.ID,
	}

	view, err := NewProductCategoryService(repository.NewProductCategoryRepository(db)).
		BuildProductBreadcrumb(item, "en")
	if err != nil {
		t.Fatalf("build product breadcrumb: %v", err)
	}
	if view.Status != "unavailable" || view.Reason != "category_path_incomplete" {
		t.Fatalf("expected incomplete category diagnostic, got %#v", view)
	}
	if len(view.Items) != 3 || view.Items[2].Path != "/products/hidden-product" {
		t.Fatalf("expected fallback home/shop/product items, got %#v", view.Items)
	}
}
