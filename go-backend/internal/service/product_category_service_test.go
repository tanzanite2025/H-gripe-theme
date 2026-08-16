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
	if view.Tree[0].Description != "轮组部件说明" {
		t.Fatalf("expected translated root description, got %q", view.Tree[0].Description)
	}
	if len(view.Tree[0].Children) != 1 {
		t.Fatalf("expected one child category, got %d", len(view.Tree[0].Children))
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
