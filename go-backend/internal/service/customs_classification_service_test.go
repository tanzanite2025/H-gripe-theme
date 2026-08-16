package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCustomsClassificationServiceCreateListAndValidate(t *testing.T) {
	db, customsService := newTestCustomsClassificationService(t)
	productSpecificationTemplate := product.ProductSpecificationTemplate{Name: "Rim", Slug: "rim", IsEnabled: true}
	require.NoError(t, db.Create(&productSpecificationTemplate).Error)

	created, err := customsService.Create(CustomsClassificationInput{
		ProductSpecificationTemplateID: &productSpecificationTemplate.ID,
		Name:                           "Carbon Rim",
		Slug:                           " Carbon Rim ",
		ComponentKind:                  "rim",
		Material:                       "Carbon Fiber",
		HSCode:                         "8714.99",
		CNCode:                         "87149990",
		CountryOfOrigin:                "cn",
		CustomsDescription:             "Bicycle carbon rim",
		Source:                         "US_HTS",
		SourceCode:                     "8714.99.80",
	})

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, productSpecificationTemplate.ID, *created.ProductSpecificationTemplateID)
	assert.Equal(t, "carbon-rim", created.Slug)
	assert.Equal(t, "871499", created.HSCode)
	assert.Equal(t, "87149990", created.CNCode)
	assert.Equal(t, "CN", created.CountryOfOrigin)
	assert.Equal(t, product.CustomsClassificationStatusActive, created.Status)

	_, err = customsService.Create(CustomsClassificationInput{
		ProductSpecificationTemplateID: &productSpecificationTemplate.ID,
		Name:                           "Aluminum Rim",
		Slug:                           "aluminum-rim",
		ComponentKind:                  "rim",
		Material:                       "Aluminum",
		HSCode:                         "871499",
		CustomsDescription:             "Bicycle aluminum rim",
		Status:                         product.CustomsClassificationStatusPaused,
	})
	require.NoError(t, err)

	activeItems, err := customsService.List(CustomsClassificationListInput{ProductSpecificationTemplateID: productSpecificationTemplate.ID})
	require.NoError(t, err)
	require.Len(t, activeItems, 1)
	assert.Equal(t, "Carbon Rim", activeItems[0].Name)

	filtered, err := customsService.List(CustomsClassificationListInput{
		ProductSpecificationTemplateID: productSpecificationTemplate.ID,
		ComponentKind:                  "RIM",
		Material:                       "carbon fiber",
	})
	require.NoError(t, err)
	require.Len(t, filtered, 1)
	assert.Equal(t, created.ID, filtered[0].ID)

	allItems, err := customsService.List(CustomsClassificationListInput{IncludePaused: true})
	require.NoError(t, err)
	assert.Len(t, allItems, 2)

	_, err = customsService.Create(CustomsClassificationInput{
		Name:   "Duplicate Rim",
		Slug:   "carbon-rim",
		HSCode: "871499",
	})
	require.ErrorIs(t, err, ErrCustomsClassificationSlugExists)

	_, err = customsService.Create(CustomsClassificationInput{
		Name:            "Invalid Origin",
		Slug:            "invalid-origin",
		HSCode:          "8714",
		CountryOfOrigin: "CHN",
	})
	require.ErrorIs(t, err, ErrCustomsClassificationInvalid)
}

func TestCustomsClassificationServiceLookupProviders(t *testing.T) {
	_, customsService := newTestCustomsClassificationService(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/us-search":
			if r.URL.Query().Get("keyword") != "carbon rim" {
				http.Error(w, "unexpected keyword", http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode([]map[string]string{
				{"htsno": "8714.99.80", "description": "Bicycle rims and parts", "general": "<p>Free</p>"},
			})
		case strings.HasPrefix(r.URL.Path, "/commodities/"):
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"attributes": map[string]interface{}{
						"goods_nomenclature_item_id": "8714999000",
						"description_plain":          "Bicycle parts of carbon fibre",
						"declarable":                 true,
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	customsService.ConfigureLookupBaseURLs(server.URL+"/us-search", server.URL+"/commodities")

	usCandidates, err := customsService.Lookup(CustomsClassificationLookupInput{
		Provider: CustomsLookupProviderUSHTS,
		Query:    "carbon rim",
		Limit:    1,
	})
	require.NoError(t, err)
	require.Len(t, usCandidates, 1)
	assert.Equal(t, "871499", usCandidates[0].HSCode)
	assert.Equal(t, "8714.99.80", usCandidates[0].SourceCode)
	assert.Equal(t, "Free", usCandidates[0].Duty)

	ukCandidates, err := customsService.Lookup(CustomsClassificationLookupInput{
		Provider: CustomsLookupProviderUKTradeTariff,
		Query:    "87149990",
	})
	require.NoError(t, err)
	require.Len(t, ukCandidates, 1)
	assert.Equal(t, "871499", ukCandidates[0].HSCode)
	assert.Equal(t, "87149990", ukCandidates[0].CNCode)
	assert.Equal(t, "Bicycle parts of carbon fibre", ukCandidates[0].CustomsDescription)

	_, err = customsService.Lookup(CustomsClassificationLookupInput{Provider: CustomsLookupProviderUKTradeTariff, Query: "rim"})
	require.ErrorIs(t, err, ErrCustomsLookupInvalid)
}

func TestCustomsClassificationServiceLookupUsesAPISettings(t *testing.T) {
	db, customsService := newTestCustomsClassificationService(t)
	require.NoError(t, db.AutoMigrate(&setting.Setting{}))
	customsService.ConfigureSettings(NewSettingService(repository.NewSettingRepository(db), nil, 0))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Customs-Key") != "customs-secret" {
			http.Error(w, "missing customs key", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]string{
			{"htsno": "8714.99.80", "description": "Configured bicycle rim", "general": "Free"},
		})
	}))
	t.Cleanup(server.Close)

	settings := []setting.Setting{
		{
			Key:      "customs_lookup_us_hts_enabled",
			Value:    "true",
			Type:     "boolean",
			Group:    "api",
			Locale:   "en",
			IsPublic: false,
		},
		{
			Key:      "customs_lookup_us_hts_endpoint",
			Value:    server.URL,
			Type:     "string",
			Group:    "api",
			Locale:   "en",
			IsPublic: false,
		},
		{
			Key:      "customs_lookup_us_hts_api_key",
			Value:    "customs-secret",
			Type:     "string",
			Group:    "api",
			Locale:   "en",
			IsPublic: false,
		},
		{
			Key:      "customs_lookup_us_hts_api_key_header",
			Value:    "X-Customs-Key",
			Type:     "string",
			Group:    "api",
			Locale:   "en",
			IsPublic: false,
		},
	}
	require.NoError(t, db.Create(&settings).Error)

	candidates, err := customsService.Lookup(CustomsClassificationLookupInput{
		Provider: CustomsLookupProviderUSHTS,
		Query:    "configured rim",
	})
	require.NoError(t, err)
	require.Len(t, candidates, 1)
	assert.Equal(t, "Configured bicycle rim", candidates[0].Description)
}

func newTestCustomsClassificationService(t *testing.T) (*gorm.DB, *CustomsClassificationService) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(t, db.AutoMigrate(
		&product.ProductSpecificationTemplate{},
		&product.CustomsClassificationProfile{},
	))

	return db, NewCustomsClassificationService(repository.NewCustomsClassificationRepository(db))
}
