package service

import (
	"errors"
	"testing"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newProductCategoryProtectionTestService(t *testing.T) (*ProductCategoryService, *gorm.DB) {
	t.Helper()

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
	return NewProductCategoryService(repository.NewProductCategoryRepository(db)), db
}

func TestProductCategoryDeleteProtectsWheelsetSystemCategory(t *testing.T) {
	service, db := newProductCategoryProtectionTestService(t)
	category := product.ProductCategory{
		Name:      "Wheelsets",
		Slug:      SystemProductCategoryWheelsetSlug,
		Depth:     1,
		IsEnabled: true,
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}

	err := service.Delete(category.ID)
	if !errors.Is(err, ErrProductCategorySystemProtected) {
		t.Fatalf("expected protected category error, got %v", err)
	}
}

func TestProductCategoryUpdateProtectsWheelsetIdentity(t *testing.T) {
	service, db := newProductCategoryProtectionTestService(t)
	parent := product.ProductCategory{
		Name:      "Wheels",
		Slug:      "wheels",
		Depth:     1,
		IsEnabled: true,
	}
	if err := db.Create(&parent).Error; err != nil {
		t.Fatalf("create parent category: %v", err)
	}
	category := product.ProductCategory{
		Name:      "Wheelsets",
		Slug:      SystemProductCategoryWheelsetSlug,
		Depth:     1,
		IsEnabled: true,
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}

	updated, err := service.Update(category.ID, ProductCategoryInput{
		ParentID:  &parent.ID,
		Name:      "Wheelsets",
		Slug:      SystemProductCategoryWheelsetSlug,
		IsEnabled: true,
	})
	if err != nil {
		t.Fatalf("expected protected category to move within the tree, got %v", err)
	}
	if updated.ParentID == nil || *updated.ParentID != parent.ID || updated.Depth != 2 {
		t.Fatalf("expected protected category to move under parent with depth 2, got parent=%v depth=%d", updated.ParentID, updated.Depth)
	}

	_, err = service.Update(category.ID, ProductCategoryInput{
		Name:      "Wheelsets",
		Slug:      "custom-wheelsets",
		IsEnabled: true,
	})
	if !errors.Is(err, ErrProductCategorySystemProtected) {
		t.Fatalf("expected protected category error for slug change, got %v", err)
	}

	_, err = service.Update(category.ID, ProductCategoryInput{
		Name:      "Wheelsets",
		Slug:      SystemProductCategoryWheelsetSlug,
		IsEnabled: false,
	})
	if !errors.Is(err, ErrProductCategorySystemProtected) {
		t.Fatalf("expected protected category error for disabling, got %v", err)
	}
}
