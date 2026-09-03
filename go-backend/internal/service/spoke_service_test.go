package service

import (
	"testing"

	productdomain "commerce-platform/internal/domain/product"
	spokedomain "commerce-platform/internal/domain/spoke"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSpokeServiceListUserHistoryFiltersByOwner(t *testing.T) {
	db, spokeService := newTestSpokeService(t)

	userID := uint(7)
	otherUserID := uint(8)
	rimModel := "R 460"
	otherRimModel := "Hidden"

	require.NoError(t, db.Create(&spokedomain.History{
		UserID:   &userID,
		RimModel: &rimModel,
	}).Error)
	require.NoError(t, db.Create(&spokedomain.History{
		UserID:   &otherUserID,
		RimModel: &otherRimModel,
	}).Error)

	items, total, err := spokeService.ListUserHistory(userID, "", 1, 10)
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, items, 1)
	require.NotNil(t, items[0].RimModel)
	assert.Equal(t, rimModel, *items[0].RimModel)
}

func TestSpokeServiceReplaceCatalogAllowsMissingGeometry(t *testing.T) {
	_, spokeService := newTestSpokeService(t)
	frontLeft := 282.0
	rearRight := 284.0

	export, err := spokeService.ReplaceCatalog(spokedomain.ExportResponse{
		Rims: []spokedomain.RimBrand{
			{
				ID:   "dt_swiss",
				Name: "DT Swiss",
				Items: []spokedomain.RimModel{
					{ID: "unknown_erd", Name: "Unknown ERD", ERD: nil},
				},
			},
		},
		Hubs: []spokedomain.HubBrand{
			{
				ID:   "custom_hubs",
				Name: "Custom Hubs",
				Items: []spokedomain.HubModel{
					{ID: "partial_hub", Name: "Partial Hub", Front: &spokedomain.HubGeometry{}},
				},
			},
		},
		Presets: []spokedomain.WheelBuildPreset{
			{
				ID:            "partial_build",
				Name:          "Partial Build",
				RimBrandID:    "dt_swiss",
				RimModelID:    "unknown_erd",
				HubBrandID:    "custom_hubs",
				HubModelID:    "partial_hub",
				WheelPosition: "auto",
				SpokeCount:    24,
				Crossing:      2,
				NippleType:    "standard",
				ActualLengths: &spokedomain.WheelBuildActualLengths{
					FrontLeft: &frontLeft,
					RearRight: &rearRight,
					Notes:     "verified bench build",
				},
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, export.Rims, 1)
	require.Len(t, export.Rims[0].Items, 1)
	assert.Nil(t, export.Rims[0].Items[0].ERD)
	require.NotEmpty(t, export.Options.SpokeCounts)
	require.Len(t, export.Presets, 1)
	require.NotNil(t, export.Presets[0].ActualLengths)
	assert.Equal(t, frontLeft, *export.Presets[0].ActualLengths.FrontLeft)
	assert.Nil(t, export.Presets[0].ActualLengths.FrontRight)
	assert.Equal(t, rearRight, *export.Presets[0].ActualLengths.RearRight)
	assert.Equal(t, "verified bench build", export.Presets[0].ActualLengths.Notes)
}

func TestSpokeServiceUsesGlobalProductBrandIdentityForRims(t *testing.T) {
	_, spokeService := newTestSpokeService(t)
	erd := 598.0

	export, err := spokeService.ReplaceCatalog(spokedomain.ExportResponse{
		Rims: []spokedomain.RimBrand{{
			ID:   "dt_swiss",
			Name: "Local name must be ignored",
			Items: []spokedomain.RimModel{{
				ID:   "rr411_db",
				Name: "RR 411 db",
				ERD:  &erd,
			}},
		}},
	})
	require.NoError(t, err)
	require.Len(t, export.Rims, 1)
	assert.Equal(t, "dt_swiss", export.Rims[0].ID)
	assert.Equal(t, "DT Swiss", export.Rims[0].Name)

	reloaded, err := spokeService.GetExport()
	require.NoError(t, err)
	require.Len(t, reloaded.Rims, 1)
	assert.Equal(t, "dt_swiss", reloaded.Rims[0].ID)
	assert.Equal(t, "DT Swiss", reloaded.Rims[0].Name)
}

func TestSpokeServiceReplaceCatalogRejectsPresetOptionsOutsideCalculatorOptions(t *testing.T) {
	_, spokeService := newTestSpokeService(t)
	erd := 598.0
	left := 22.5
	right := 35.6
	pcd := 44.0

	_, err := spokeService.ReplaceCatalog(spokedomain.ExportResponse{
		Rims: []spokedomain.RimBrand{
			{
				ID:   "dt_swiss",
				Name: "DT Swiss",
				Items: []spokedomain.RimModel{
					{ID: "rr", Name: "RR", ERD: &erd},
				},
			},
		},
		Hubs: []spokedomain.HubBrand{
			{
				ID:   "custom_hubs",
				Name: "Custom Hubs",
				Items: []spokedomain.HubModel{
					{
						ID:   "hub",
						Name: "Hub",
						Front: &spokedomain.HubGeometry{
							LeftFlange:     &left,
							RightFlange:    &right,
							LeftFlangePCD:  &pcd,
							RightFlangePCD: &pcd,
						},
					},
				},
			},
		},
		Presets: []spokedomain.WheelBuildPreset{
			{
				ID:            "bad_build",
				Name:          "Bad Build",
				RimBrandID:    "dt_swiss",
				RimModelID:    "rr",
				HubBrandID:    "custom_hubs",
				HubModelID:    "hub",
				WheelPosition: "front",
				SpokeCount:    30,
				Crossing:      2,
				NippleType:    "standard",
			},
		},
	})
	require.ErrorIs(t, err, ErrInvalidSpokeCatalog)
}

func TestSpokeServiceReplaceCatalogRejectsInvalidActualLengths(t *testing.T) {
	_, spokeService := newTestSpokeService(t)
	erd := 598.0
	left := 22.5
	right := 35.6
	pcd := 44.0
	badLength := 0.0

	_, err := spokeService.ReplaceCatalog(spokedomain.ExportResponse{
		Rims: []spokedomain.RimBrand{
			{
				ID:   "dt_swiss",
				Name: "DT Swiss",
				Items: []spokedomain.RimModel{
					{ID: "rr", Name: "RR", ERD: &erd},
				},
			},
		},
		Hubs: []spokedomain.HubBrand{
			{
				ID:   "custom_hubs",
				Name: "Custom Hubs",
				Items: []spokedomain.HubModel{
					{
						ID:   "hub",
						Name: "Hub",
						Front: &spokedomain.HubGeometry{
							LeftFlange:     &left,
							RightFlange:    &right,
							LeftFlangePCD:  &pcd,
							RightFlangePCD: &pcd,
						},
					},
				},
			},
		},
		Presets: []spokedomain.WheelBuildPreset{
			{
				ID:            "bad_actual_length",
				Name:          "Bad Actual Length",
				RimBrandID:    "dt_swiss",
				RimModelID:    "rr",
				HubBrandID:    "custom_hubs",
				HubModelID:    "hub",
				WheelPosition: "front",
				SpokeCount:    24,
				Crossing:      2,
				NippleType:    "standard",
				ActualLengths: &spokedomain.WheelBuildActualLengths{
					FrontLeft: &badLength,
				},
			},
		},
	})
	require.ErrorIs(t, err, ErrInvalidSpokeCatalog)
}

func newTestSpokeService(t *testing.T) (*gorm.DB, *SpokeService) {
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
		&productdomain.ProductBrand{},
		&spokedomain.CatalogRimBrand{},
		&spokedomain.CatalogRimModel{},
		&spokedomain.CatalogHubBrand{},
		&spokedomain.CatalogHubModel{},
		&spokedomain.CatalogBuildPreset{},
		&spokedomain.History{},
	))

	seedSpokeProductBrands(t, db)
	spokeRepo := repository.NewSpokeRepository(db)
	spokeRepo.ConfigureProductBrandRepository(repository.NewProductBrandRepository(db))
	return db, NewSpokeService(spokeRepo)
}

func seedSpokeProductBrands(t *testing.T, db *gorm.DB) {
	t.Helper()

	require.NoError(t, db.Create(&productdomain.ProductBrand{
		Name:      "DT Swiss",
		Slug:      "dt_swiss",
		IsEnabled: true,
		SortOrder: 0,
	}).Error)
	require.NoError(t, db.Create(&productdomain.ProductBrand{
		Name:      "Mavic",
		Slug:      "mavic",
		IsEnabled: true,
		SortOrder: 1,
	}).Error)
	require.NoError(t, db.Create(&productdomain.ProductBrand{
		Name:      "Kinlin",
		Slug:      "kinlin",
		IsEnabled: true,
		SortOrder: 2,
	}).Error)
}
