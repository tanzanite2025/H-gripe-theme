package service

import (
	"testing"

	fitmentcatalogdomain "commerce-platform/internal/domain/fitmentcatalog"
	productdomain "commerce-platform/internal/domain/product"
	spokedomain "commerce-platform/internal/domain/spoke"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSpokeServiceUsesFitmentHubSpecificationsAsOnlyHubSource(t *testing.T) {
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
		&fitmentcatalogdomain.HubSpecification{},
		&productdomain.ProductBrand{},
		&spokedomain.CatalogRimBrand{},
		&spokedomain.CatalogRimModel{},
		&spokedomain.CatalogHubBrand{},
		&spokedomain.CatalogHubModel{},
		&spokedomain.CatalogBuildPreset{},
	))

	require.NoError(t, db.Create(&fitmentcatalogdomain.HubSpecification{
		SpecCode:      "HUB-F-100",
		DisplayName:   "Front hub 100",
		Position:      fitmentcatalogdomain.HubPositionFront,
		AxleType:      fitmentcatalogdomain.HubAxleTypeQuickRelease,
		AxleSpacingMM: 100,
		WRMM:          spokeTestFloatPointer(35),
		WLMM:          spokeTestFloatPointer(22),
		PCDRMM:        spokeTestFloatPointer(44),
		PCDLMM:        spokeTestFloatPointer(42),
		IsEnabled:     true,
		SortOrder:     0,
	}).Error)
	require.NoError(t, db.Create(&fitmentcatalogdomain.HubSpecification{
		SpecCode:      "HUB-R-142",
		DisplayName:   "Rear hub 142",
		Position:      fitmentcatalogdomain.HubPositionRear,
		AxleType:      fitmentcatalogdomain.HubAxleTypeThruAxle,
		AxleSpacingMM: 142,
		WRMM:          spokeTestFloatPointer(20),
		WLMM:          spokeTestFloatPointer(33),
		PCDRMM:        spokeTestFloatPointer(46),
		PCDLMM:        spokeTestFloatPointer(46),
		IsEnabled:     true,
		SortOrder:     1,
	}).Error)
	require.NoError(t, db.Create(&fitmentcatalogdomain.HubSpecification{
		SpecCode:      "HUB-DISABLED",
		DisplayName:   "Disabled hub",
		Position:      fitmentcatalogdomain.HubPositionRear,
		AxleType:      fitmentcatalogdomain.HubAxleTypeThruAxle,
		AxleSpacingMM: 142,
		IsEnabled:     false,
		SortOrder:     2,
	}).Error)

	require.NoError(t, db.Create(&productdomain.ProductBrand{
		Name:      "Test rims",
		Slug:      "test-rims",
		IsEnabled: true,
		SortOrder: 0,
	}).Error)

	erd := 598.0
	spokeRepo := repository.NewSpokeRepository(
		db,
		repository.NewFitmentHubSpecificationRepository(db),
	)
	spokeRepo.ConfigureProductBrandRepository(repository.NewProductBrandRepository(db))
	spokeService := NewSpokeService(spokeRepo)

	export, err := spokeService.ReplaceCatalog(spokedomain.ExportResponse{
		Rims: []spokedomain.RimBrand{
			{
				ID:   "test-rims",
				Name: "Test rims",
				Items: []spokedomain.RimModel{
					{ID: "test-rim", Name: "Test rim", ERD: &erd},
				},
			},
		},
		Hubs: []spokedomain.HubBrand{
			{
				ID:   "manual-hubs",
				Name: "Manual hubs must be ignored",
				Items: []spokedomain.HubModel{
					{ID: "manual-hub", Name: "Manual hub"},
				},
			},
		},
		Presets: []spokedomain.WheelBuildPreset{
			{
				ID:            "rear-build",
				Name:          "Rear build",
				RimBrandID:    "test-rims",
				RimModelID:    "test-rim",
				HubBrandID:    "fitment_catalog",
				HubModelID:    "hub-r-142",
				WheelPosition: "rear",
				SpokeCount:    24,
				Crossing:      2,
				NippleType:    "standard",
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, export.Hubs, 1)
	require.Equal(t, "fitment_catalog", export.Hubs[0].ID)
	require.Len(t, export.Hubs[0].Items, 2)
	require.Equal(t, "hub-f-100", export.Hubs[0].Items[0].ID)
	require.Equal(t, "hub-r-142", export.Hubs[0].Items[1].ID)
	require.Nil(t, export.Hubs[0].Items[0].Rear)
	require.NotNil(t, export.Hubs[0].Items[0].Front)
	require.Equal(t, 35.0, *export.Hubs[0].Items[0].Front.RightFlange)
	require.Equal(t, 22.0, *export.Hubs[0].Items[0].Front.LeftFlange)
	require.Equal(t, 44.0, *export.Hubs[0].Items[0].Front.RightFlangePCD)
	require.Equal(t, 42.0, *export.Hubs[0].Items[0].Front.LeftFlangePCD)
	require.NotNil(t, export.Hubs[0].Items[1].Rear)
	require.Equal(t, 20.0, *export.Hubs[0].Items[1].Rear.RightFlange)
	require.Equal(t, 33.0, *export.Hubs[0].Items[1].Rear.LeftFlange)
	require.Len(t, export.Presets, 1)
	require.Equal(t, "fitment_catalog", export.Presets[0].HubBrandID)
	require.Equal(t, "hub-r-142", export.Presets[0].HubModelID)

	var projectedHubModels []spokedomain.CatalogHubModel
	require.NoError(t, db.Order("code ASC").Find(&projectedHubModels).Error)
	require.Len(t, projectedHubModels, 2)
	require.Equal(t, 20.0, *projectedHubModels[1].RearRightFlange)
	require.Equal(t, 33.0, *projectedHubModels[1].RearLeftFlange)

	require.NoError(t, db.Model(&fitmentcatalogdomain.HubSpecification{}).
		Where("spec_code = ?", "HUB-R-142").
		Update("is_enabled", false).Error)

	export, err = spokeService.GetExport()
	require.NoError(t, err)
	require.Len(t, export.Hubs, 1)
	require.Len(t, export.Hubs[0].Items, 1)
	require.Equal(t, "hub-f-100", export.Hubs[0].Items[0].ID)
	require.Empty(t, export.Presets)
}

func TestSpokeServiceKeepsPresetAfterFitmentHubCodeChange(t *testing.T) {
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
		&fitmentcatalogdomain.FrameFitmentEntry{},
		&fitmentcatalogdomain.FrameHubSpecification{},
		&fitmentcatalogdomain.HubSpecification{},
		&productdomain.ProductBrand{},
		&spokedomain.CatalogRimBrand{},
		&spokedomain.CatalogRimModel{},
		&spokedomain.CatalogHubBrand{},
		&spokedomain.CatalogHubModel{},
		&spokedomain.CatalogBuildPreset{},
	))

	hubRepository := repository.NewFitmentHubSpecificationRepository(db)
	spokeRepository := repository.NewSpokeRepository(db, hubRepository)
	spokeRepository.ConfigureProductBrandRepository(repository.NewProductBrandRepository(db))
	spokeService := NewSpokeService(spokeRepository)

	require.NoError(t, db.Create(&fitmentcatalogdomain.HubSpecification{
		SpecCode:      "HUB-R-142",
		DisplayName:   "Rear hub 142",
		Position:      fitmentcatalogdomain.HubPositionRear,
		AxleType:      fitmentcatalogdomain.HubAxleTypeThruAxle,
		AxleSpacingMM: 142,
		WRMM:          spokeTestFloatPointer(20),
		WLMM:          spokeTestFloatPointer(33),
		PCDRMM:        spokeTestFloatPointer(46),
		PCDLMM:        spokeTestFloatPointer(46),
		IsEnabled:     true,
	}).Error)

	require.NoError(t, db.Create(&productdomain.ProductBrand{
		Name:      "Test rims",
		Slug:      "test-rims",
		IsEnabled: true,
	}).Error)

	var specification fitmentcatalogdomain.HubSpecification
	require.NoError(t, db.First(&specification).Error)

	erd := 598.0
	_, err = spokeService.ReplaceCatalog(spokedomain.ExportResponse{
		Rims: []spokedomain.RimBrand{{
			ID:   "test-rims",
			Name: "Test rims",
			Items: []spokedomain.RimModel{{
				ID:   "test-rim",
				Name: "Test rim",
				ERD:  &erd,
			}},
		}},
		Presets: []spokedomain.WheelBuildPreset{{
			ID:            "rear-build",
			Name:          "Rear build",
			RimBrandID:    "test-rims",
			RimModelID:    "test-rim",
			HubBrandID:    "fitment_catalog",
			HubModelID:    "hub-r-142",
			WheelPosition: "rear",
			SpokeCount:    24,
			Crossing:      2,
			NippleType:    "standard",
		}},
	})
	require.NoError(t, err)

	require.NoError(t, db.Model(&fitmentcatalogdomain.HubSpecification{}).
		Where("id = ?", specification.ID).
		Update("spec_code", "HUB-R-142-V2").Error)

	export, err := spokeService.GetExport()
	require.NoError(t, err)
	require.Len(t, export.Presets, 1)
	require.Equal(t, "hub-r-142-v2", export.Presets[0].HubModelID)

	var projectedModel spokedomain.CatalogHubModel
	require.NoError(t, db.First(&projectedModel).Error)
	require.NotNil(t, projectedModel.FitmentHubSpecificationID)
	require.Equal(t, specification.ID, *projectedModel.FitmentHubSpecificationID)

	hubService := NewFitmentHubSpecificationService(hubRepository)
	hubService.ConfigureSpokeRepository(spokeRepository)
	require.ErrorIs(t, hubService.Delete(specification.ID), ErrFitmentHubSpecificationInUse)
}

func spokeTestFloatPointer(value float64) *float64 {
	return &value
}
