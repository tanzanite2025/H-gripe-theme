package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
	_, customsService := newTestCustomsClassificationService(t)

	created, err := customsService.Create(CustomsClassificationInput{
		Name:               "Carbon Rim",
		Slug:               " Carbon Rim ",
		ComponentKind:      "rim",
		Material:           "Carbon Fiber",
		HSCode:             "8714.99",
		CNCode:             "87149990",
		CountryOfOrigin:    "cn",
		CustomsDescription: "Bicycle carbon rim",
		Source:             "US_HTS",
		SourceCode:         "8714.99.80",
	})

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, "carbon-rim", created.Slug)
	assert.Equal(t, "871499", created.HSCode)
	assert.Equal(t, "87149990", created.CNCode)
	assert.Equal(t, "CN", created.CountryOfOrigin)
	assert.Equal(t, product.CustomsClassificationStatusActive, created.Status)

	_, err = customsService.Create(CustomsClassificationInput{
		Name:               "Aluminum Rim",
		Slug:               "aluminum-rim",
		ComponentKind:      "rim",
		Material:           "Aluminum",
		HSCode:             "871499",
		CustomsDescription: "Bicycle aluminum rim",
		Status:             product.CustomsClassificationStatusPaused,
	})
	require.NoError(t, err)

	activeItems, err := customsService.List(CustomsClassificationListInput{})
	require.NoError(t, err)
	require.Len(t, activeItems, 1)
	assert.Equal(t, "Carbon Rim", activeItems[0].Name)

	filtered, err := customsService.List(CustomsClassificationListInput{
		ComponentKind: "RIM",
		Material:      "carbon fiber",
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

func TestCustomsClassificationServiceNormalizesTradeRemedyRiskFields(t *testing.T) {
	_, customsService := newTestCustomsClassificationService(t)

	created, err := customsService.Create(CustomsClassificationInput{
		Name:                         "Sensitive bicycle wheelset",
		Slug:                         "sensitive-bicycle-wheelset",
		HSCode:                       "871499",
		CustomsDescription:           "Bicycle wheelset",
		TradeRemedyRiskLevel:         product.CustomsTradeRemedyRiskLevelHigh,
		TradeRemedyRiskTags:          []string{"US_SECTION_301_LIST_3_REVIEW", "eu_anti_dumping_attention", "eu_anti_dumping_attention"},
		TradeRemedyDeclarationAdvice: "Review current destination measures before shipment.",
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, product.CustomsTradeRemedyRiskLevelHigh, created.TradeRemedyRiskLevel)
	assert.Equal(t, []string{
		product.CustomsTradeRemedyRiskTagUSSection301List3Review,
		product.CustomsTradeRemedyRiskTagEUAntiDumpingAttention,
	}, []string(created.TradeRemedyRiskTags))
	assert.Equal(t, "Review current destination measures before shipment.", created.TradeRemedyDeclarationAdvice)

	_, err = customsService.Create(CustomsClassificationInput{
		Name:                 "Unsupported risk level",
		Slug:                 "unsupported-risk-level",
		HSCode:               "871499",
		TradeRemedyRiskLevel: "extreme",
	})
	require.ErrorIs(t, err, ErrCustomsClassificationInvalid)

	_, err = customsService.Create(CustomsClassificationInput{
		Name:                "Unsupported risk tag",
		Slug:                "unsupported-risk-tag",
		HSCode:              "871499",
		TradeRemedyRiskTags: []string{"invented_trade_measure"},
	})
	require.ErrorIs(t, err, ErrCustomsClassificationInvalid)
}

func TestCustomsClassificationServiceClearsVerificationDatesWhenTradeRemedyAdviceChanges(t *testing.T) {
	db, customsService := newTestCustomsClassificationService(t)
	created, err := customsService.Create(CustomsClassificationInput{
		Name:                 "Verified wheelset",
		Slug:                 "verified-wheelset",
		HSCode:               "871499",
		CustomsDescription:   "Bicycle wheelset",
		TradeRemedyRiskLevel: product.CustomsTradeRemedyRiskLevelHigh,
		TradeRemedyRiskTags:  []string{product.CustomsTradeRemedyRiskTagEUAntiDumpingAttention},
	})
	require.NoError(t, err)
	verifiedAt := time.Date(2026, time.October, 5, 0, 0, 0, 0, time.UTC)
	reviewDueAt := time.Date(2027, time.October, 5, 0, 0, 0, 0, time.UTC)
	require.NoError(t, db.Model(&product.CustomsClassificationProfile{}).Where("id = ?", created.ID).Updates(map[string]interface{}{
		"verified_at":   verifiedAt,
		"review_due_at": reviewDueAt,
	}).Error)

	updated, err := customsService.Update(created.ID, CustomsClassificationInput{
		Name:                         created.Name,
		Slug:                         created.Slug,
		HSCode:                       created.HSCode,
		CustomsDescription:           created.CustomsDescription,
		TradeRemedyRiskLevel:         created.TradeRemedyRiskLevel,
		TradeRemedyRiskTags:          []string(created.TradeRemedyRiskTags),
		TradeRemedyDeclarationAdvice: "Recheck EU and US measures before every shipment.",
		Status:                       created.Status,
	})
	require.NoError(t, err)
	assert.Nil(t, updated.VerifiedAt)
	assert.Nil(t, updated.ReviewDueAt)
}

func TestCustomsClassificationServiceSynchronizesLegacySourceURLWithUSAuthorityURL(t *testing.T) {
	_, customsService := newTestCustomsClassificationService(t)

	created, err := customsService.Create(CustomsClassificationInput{
		Name:               "Destination source matrix rim",
		Slug:               "destination-source-matrix-rim",
		HSCode:             "871492",
		Source:             "administrator",
		SourceCode:         "8714.92",
		SourceURLUS:        "https://hts.usitc.gov/search?query=8714.92",
		SourceURLEU:        "https://ec.europa.eu/taxation_customs/dds2/taric",
		SourceURLUK:        "https://www.gov.uk/trade-tariff/8714921000",
		CustomsDescription: "Bicycle rim",
	})
	require.NoError(t, err)
	assert.Equal(t, created.SourceURLUS, created.SourceURL)
	assert.Equal(t, "https://ec.europa.eu/taxation_customs/dds2/taric", created.SourceURLEU)
	assert.Equal(t, "https://www.gov.uk/trade-tariff/8714921000", created.SourceURLUK)

	updated, err := customsService.Update(created.ID, CustomsClassificationInput{
		Name:               created.Name,
		Slug:               created.Slug,
		HSCode:             created.HSCode,
		Source:             created.Source,
		SourceCode:         created.SourceCode,
		SourceURLUS:        "https://hts.usitc.gov/search?query=8714.92.10.00",
		SourceURLEU:        created.SourceURLEU,
		SourceURLUK:        created.SourceURLUK,
		CustomsDescription: created.CustomsDescription,
		Status:             created.Status,
	})
	require.NoError(t, err)
	assert.Equal(t, updated.SourceURLUS, updated.SourceURL)
	assert.Equal(t, "https://hts.usitc.gov/search?query=8714.92.10.00", updated.SourceURL)
}

func TestCustomsClassificationServicePreservesBuiltInTemplates(t *testing.T) {
	_, customsService := newTestCustomsClassificationService(t)

	created, err := customsService.Create(CustomsClassificationInput{
		Name:               "Built-in bicycle hub",
		Slug:               "built-in-bicycle-hub",
		HSCode:             "871499",
		Source:             "built_in",
		CustomsDescription: "Bicycle hub",
	})
	require.NoError(t, err)

	updated, err := customsService.Update(created.ID, CustomsClassificationInput{
		Name:               created.Name,
		Slug:               created.Slug,
		HSCode:             created.HSCode,
		Source:             "admin_override",
		CustomsDescription: created.CustomsDescription,
		Status:             created.Status,
	})
	require.NoError(t, err)
	assert.Equal(t, "built_in", updated.Source)

	err = customsService.Delete(created.ID)
	require.ErrorIs(t, err, ErrCustomsClassificationBuiltIn)

	remaining, err := customsService.Get(created.ID)
	require.NoError(t, err)
	assert.Equal(t, "built_in", remaining.Source)
}

func TestCustomsClassificationServicePreservesVerificationDatesForNotesAndClearsThemForClassificationChanges(t *testing.T) {
	db, customsService := newTestCustomsClassificationService(t)

	created, err := customsService.Create(CustomsClassificationInput{
		Name:               "Verified carbon rim",
		Slug:               "verified-carbon-rim",
		ComponentKind:      "rim",
		Material:           "carbon_fiber",
		HSCode:             "871492",
		CNCode:             "87149210",
		CustomsDescription: "Bicycle carbon rim",
		Source:             "built_in",
		SourceCode:         "8714.92.10.00",
		SourceURL:          "https://hts.usitc.gov/search?query=8714.92",
	})
	require.NoError(t, err)

	verifiedAt := time.Date(2026, time.October, 4, 0, 0, 0, 0, time.UTC)
	reviewDueAt := time.Date(2027, time.October, 4, 0, 0, 0, 0, time.UTC)
	require.NoError(t, db.Model(&product.CustomsClassificationProfile{}).Where("id = ?", created.ID).Updates(map[string]interface{}{
		"verified_at":   verifiedAt,
		"review_due_at": reviewDueAt,
	}).Error)

	updated, err := customsService.Update(created.ID, CustomsClassificationInput{
		Name:               created.Name,
		Slug:               created.Slug,
		ComponentKind:      created.ComponentKind,
		Material:           created.Material,
		HSCode:             created.HSCode,
		CNCode:             created.CNCode,
		CustomsDescription: created.CustomsDescription,
		Source:             "administrator_edit",
		SourceCode:         created.SourceCode,
		SourceURL:          created.SourceURL,
		Notes:              "Updated explanation only",
		Status:             created.Status,
	})
	require.NoError(t, err)
	require.NotNil(t, updated.VerifiedAt)
	require.NotNil(t, updated.ReviewDueAt)
	assert.True(t, updated.VerifiedAt.Equal(verifiedAt))
	assert.True(t, updated.ReviewDueAt.Equal(reviewDueAt))

	updated, err = customsService.Update(created.ID, CustomsClassificationInput{
		Name:               created.Name,
		Slug:               created.Slug,
		ComponentKind:      created.ComponentKind,
		Material:           created.Material,
		HSCode:             "871499",
		CNCode:             "87149990",
		CustomsDescription: created.CustomsDescription,
		Source:             "administrator_edit",
		SourceCode:         "8714.99.80.00",
		SourceURL:          "https://hts.usitc.gov/search?query=8714.99",
		Status:             created.Status,
	})
	require.NoError(t, err)
	assert.Nil(t, updated.VerifiedAt)
	assert.Nil(t, updated.ReviewDueAt)
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
	assert.Equal(t, "https://www.gov.uk/trade-tariff/8714999000", ukCandidates[0].SourceURL)

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

	require.NoError(t, db.AutoMigrate(&product.CustomsClassificationProfile{}))

	return db, NewCustomsClassificationService(repository.NewCustomsClassificationRepository(db))
}
