package product

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func TestSearchSchwalbeTireCatalogSelectorFiltersInnerRimWidthBeforePagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newSchwalbeTireCatalogHandlerTestDB(t)
	if err := db.Exec(`CREATE TABLE schwalbe_tire_rim_width_combination_rules (
		id INTEGER PRIMARY KEY,
		tire_width_min_mm INTEGER NOT NULL,
		tire_width_max_mm INTEGER NOT NULL,
		inner_rim_width_min_mm INTEGER NOT NULL,
		inner_rim_width_max_mm INTEGER NOT NULL,
		source_basis TEXT NOT NULL,
		source_version TEXT NOT NULL,
		source_url TEXT NOT NULL,
		source_checked_at DATE NOT NULL
	)`).Error; err != nil {
		t.Fatalf("create rim-width rule table: %v", err)
	}
	if err := db.Exec(`INSERT INTO schwalbe_tire_rim_width_combination_rules
		(id, tire_width_min_mm, tire_width_max_mm, inner_rim_width_min_mm, inner_rim_width_max_mm, source_basis, source_version, source_url, source_checked_at)
		VALUES (1, 35, 46, 17, 27, 'ETRTO Standard 2024; Schwalbe possible combinations guidance', '05/2024', 'https://example.invalid/rules', '2026-09-28')`).Error; err != nil {
		t.Fatalf("insert rim-width rule: %v", err)
	}

	for index := 1; index <= 21; index++ {
		articleNo := fmt.Sprintf("match-%02d", index)
		if err := db.Exec(`INSERT INTO schwalbe_tire_specifications
			(article_no, model_name, etrto, inch_designation, source_url, source_checked_at)
			VALUES (?, ?, '40-622', '28x1.50', 'https://example.invalid/catalog', '2026-09-28')`, articleNo, articleNo).Error; err != nil {
			t.Fatalf("insert matching row %s: %v", articleNo, err)
		}
	}
	if err := db.Exec(`INSERT INTO schwalbe_tire_specifications
		(article_no, model_name, etrto, inch_designation, source_url, source_checked_at)
		VALUES ('match-29', 'Match 29 inch', '40-622', '29x1.50', 'https://example.invalid/catalog', '2026-09-28')`).Error; err != nil {
		t.Fatalf("insert other-wheel matching row: %v", err)
	}
	if err := db.Exec(`INSERT INTO schwalbe_tire_specifications
		(article_no, model_name, etrto, inch_designation, source_url, source_checked_at)
		VALUES ('excluded', 'Excluded', '80-622', '28x3.00', 'https://example.invalid/catalog', '2026-09-28')`).Error; err != nil {
		t.Fatalf("insert excluded row: %v", err)
	}

	router := gin.New()
	handler := NewHandler(service.NewProductService(repository.NewProductRepository(db), nil, 0))
	router.GET("/products/schwalbe-tire-catalog/selector", handler.SearchSchwalbeTireCatalogSelector)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/products/schwalbe-tire-catalog/selector?inner_rim_width_mm=23.5&wheel_size=28-622&page=2&sort=article", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Data struct {
			Items []struct {
				ArticleNo        string `json:"article_no"`
				RimWidthGuidance []struct {
					InnerRimWidthMinMM int `json:"inner_rim_width_min_mm"`
					InnerRimWidthMaxMM int `json:"inner_rim_width_max_mm"`
				} `json:"rim_width_guidance"`
			} `json:"items"`
			Page            int `json:"page"`
			Total           int `json:"total"`
			TotalPages      int `json:"total_pages"`
			RimWidthContext struct {
				InnerRimWidthMM float64 `json:"inner_rim_width_mm"`
				GuidanceStatus  string  `json:"guidance_status"`
				SourceVersion   string  `json:"source_version"`
			} `json:"rim_width_context"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.Page != 2 || payload.Data.Total != 21 || payload.Data.TotalPages != 2 || len(payload.Data.Items) != 1 {
		t.Fatalf("expected 21 filtered rows across two pages, got page=%d total=%d pages=%d items=%d", payload.Data.Page, payload.Data.Total, payload.Data.TotalPages, len(payload.Data.Items))
	}
	if payload.Data.Items[0].ArticleNo != "match-21" {
		t.Fatalf("expected the last matching row on page 2, got %s", payload.Data.Items[0].ArticleNo)
	}
	if len(payload.Data.Items[0].RimWidthGuidance) != 1 ||
		payload.Data.Items[0].RimWidthGuidance[0].InnerRimWidthMinMM != 17 ||
		payload.Data.Items[0].RimWidthGuidance[0].InnerRimWidthMaxMM != 27 {
		t.Fatalf("expected the source-backed inner-width range on the item, got %+v", payload.Data.Items[0].RimWidthGuidance)
	}
	if payload.Data.RimWidthContext.InnerRimWidthMM != 23.5 || payload.Data.RimWidthContext.GuidanceStatus != "covered" || payload.Data.RimWidthContext.SourceVersion != "05/2024" {
		t.Fatalf("expected covered rim-width context, got %+v", payload.Data.RimWidthContext)
	}
	if strings.Contains(recorder.Body.String(), "source_url") || strings.Contains(recorder.Body.String(), "example.invalid") {
		t.Fatalf("selector response exposed source URL: %s", recorder.Body.String())
	}
}

func TestSearchSchwalbeTireCatalogSelectorReportsUncoveredInnerRimWidth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newSchwalbeTireCatalogHandlerTestDB(t)
	if err := db.Exec(`CREATE TABLE schwalbe_tire_rim_width_combination_rules (
		id INTEGER PRIMARY KEY,
		tire_width_min_mm INTEGER NOT NULL,
		tire_width_max_mm INTEGER NOT NULL,
		inner_rim_width_min_mm INTEGER NOT NULL,
		inner_rim_width_max_mm INTEGER NOT NULL,
		source_basis TEXT NOT NULL,
		source_version TEXT NOT NULL,
		source_url TEXT NOT NULL,
		source_checked_at DATE NOT NULL
	)`).Error; err != nil {
		t.Fatalf("create rim-width rule table: %v", err)
	}
	if err := db.Exec(`INSERT INTO schwalbe_tire_rim_width_combination_rules
		(id, tire_width_min_mm, tire_width_max_mm, inner_rim_width_min_mm, inner_rim_width_max_mm, source_basis, source_version, source_url, source_checked_at)
		VALUES (1, 35, 46, 17, 27, 'ETRTO Standard 2024', '05/2024', 'https://example.invalid/rules', '2026-09-28')`).Error; err != nil {
		t.Fatalf("insert rim-width rule: %v", err)
	}
	if err := db.Exec(`INSERT INTO schwalbe_tire_specifications (article_no, model_name, etrto, source_url, source_checked_at) VALUES ('candidate', 'Candidate', '40-622', 'https://example.invalid/catalog', '2026-09-28')`).Error; err != nil {
		t.Fatalf("insert catalog row: %v", err)
	}

	router := gin.New()
	handler := NewHandler(service.NewProductService(repository.NewProductRepository(db), nil, 0))
	router.GET("/products/schwalbe-tire-catalog/selector", handler.SearchSchwalbeTireCatalogSelector)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/products/schwalbe-tire-catalog/selector?inner_rim_width_mm=30&wheel_size=28-622", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Data struct {
			Items           []json.RawMessage `json:"items"`
			Total           int               `json:"total"`
			RimWidthContext struct {
				GuidanceStatus string `json:"guidance_status"`
			} `json:"rim_width_context"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.Total != 0 || len(payload.Data.Items) != 0 || payload.Data.RimWidthContext.GuidanceStatus != "no_coverage" {
		t.Fatalf("expected uncovered inner width to return no rows and explicit status, got total=%d items=%d status=%s", payload.Data.Total, len(payload.Data.Items), payload.Data.RimWidthContext.GuidanceStatus)
	}
}

func TestSearchSchwalbeTireCatalogSelectorRequiresWheelSizeForInnerRimWidth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newSchwalbeTireCatalogHandlerTestDB(t)
	if err := db.Exec(`CREATE TABLE schwalbe_tire_rim_width_combination_rules (
		id INTEGER PRIMARY KEY,
		tire_width_min_mm INTEGER NOT NULL,
		tire_width_max_mm INTEGER NOT NULL,
		inner_rim_width_min_mm INTEGER NOT NULL,
		inner_rim_width_max_mm INTEGER NOT NULL,
		source_basis TEXT NOT NULL,
		source_version TEXT NOT NULL,
		source_url TEXT NOT NULL,
		source_checked_at DATE NOT NULL
	)`).Error; err != nil {
		t.Fatalf("create rim-width rule table: %v", err)
	}
	if err := db.Exec(`INSERT INTO schwalbe_tire_rim_width_combination_rules
		(id, tire_width_min_mm, tire_width_max_mm, inner_rim_width_min_mm, inner_rim_width_max_mm, source_basis, source_version, source_url, source_checked_at)
		VALUES (1, 35, 46, 17, 27, 'ETRTO Standard 2024', '05/2024', 'https://example.invalid/rules', '2026-09-28')`).Error; err != nil {
		t.Fatalf("insert rim-width rule: %v", err)
	}
	if err := db.Exec(`INSERT INTO schwalbe_tire_specifications (article_no, model_name, etrto, inch_designation, source_url, source_checked_at) VALUES ('candidate', 'Candidate', '40-622', '28x1.50', 'https://example.invalid/catalog', '2026-09-28')`).Error; err != nil {
		t.Fatalf("insert catalog row: %v", err)
	}

	router := gin.New()
	handler := NewHandler(service.NewProductService(repository.NewProductRepository(db), nil, 0))
	router.GET("/products/schwalbe-tire-catalog/selector", handler.SearchSchwalbeTireCatalogSelector)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/products/schwalbe-tire-catalog/selector?inner_rim_width_mm=23", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Data struct {
			Items           []json.RawMessage `json:"items"`
			Total           int               `json:"total"`
			RimWidthContext struct {
				GuidanceStatus string `json:"guidance_status"`
			} `json:"rim_width_context"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.Total != 0 || len(payload.Data.Items) != 0 || payload.Data.RimWidthContext.GuidanceStatus != "wheel_size_required" {
		t.Fatalf("expected incomplete inner-width context to return no rows and an explicit wheel-size requirement, got total=%d items=%d status=%s", payload.Data.Total, len(payload.Data.Items), payload.Data.RimWidthContext.GuidanceStatus)
	}
}
