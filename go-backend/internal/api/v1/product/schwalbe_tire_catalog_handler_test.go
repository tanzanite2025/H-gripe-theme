package product

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestListSchwalbeTireCatalogDoesNotExposeSourceURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newSchwalbeTireCatalogHandlerTestDB(t)
	if err := db.Exec(`INSERT INTO schwalbe_tire_specifications (article_no, model_name, etrto, source_url, source_checked_at) VALUES ('11111111', 'Green Marathon', '40-622', 'https://www.schwalbe.com/example', '2026-09-28')`).Error; err != nil {
		t.Fatalf("insert test data: %v", err)
	}

	router := gin.New()
	handler := NewHandler(service.NewProductService(repository.NewProductRepository(db), nil, 0))
	router.GET("/products/schwalbe-tire-catalog", handler.ListSchwalbeTireCatalog)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/products/schwalbe-tire-catalog", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if strings.Contains(body, "source_url") || strings.Contains(body, "www.schwalbe.com") {
		t.Fatalf("public catalog response exposed source URL: %s", body)
	}
	for _, expected := range []string{`"article_no":"11111111"`, `"model_name":"Green Marathon"`, `"source_checked_at"`} {
		if !strings.Contains(body, expected) {
			t.Fatalf("expected response to contain %s, got %s", expected, body)
		}
	}
}

func TestSearchSchwalbeTireCatalogSelectorKeepsFacetOptionsFromSearchResults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newSchwalbeTireCatalogHandlerTestDB(t)
	statements := []string{
		`INSERT INTO schwalbe_tire_specifications (article_no, model_name, etrto, version_label, compound, color, bead, e_bike_rating, seal, source_url, source_checked_at) VALUES ('11111111', 'Green Marathon', '40-622', 'Super Ground, Radial', 'ADDIX', 'Black', 'Folding', NULL, 'TLR', 'https://www.schwalbe.com/first', '2026-09-28')`,
		`INSERT INTO schwalbe_tire_specifications (article_no, model_name, etrto, version_label, compound, color, bead, e_bike_rating, seal, source_url, source_checked_at) VALUES ('22222222', 'Green Marathon Plus', '100-622', 'Super Race', 'ADDIX Eco', 'Blue', 'Wired', 'E-25', 'Tube', 'https://www.schwalbe.com/second', '2026-09-28')`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("insert test data: %v", err)
		}
	}

	router := gin.New()
	handler := NewHandler(service.NewProductService(repository.NewProductRepository(db), nil, 0))
	router.GET("/products/schwalbe-tire-catalog/selector", handler.SearchSchwalbeTireCatalogSelector)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/products/schwalbe-tire-catalog/selector?search=green&model=Green%20Marathon&tire_width_mm=40&bead_seat_diameter_mm=622&casing=Super%20Ground&radial=1&bead=Folding&seal=TLR&e_bike_rating=none&color=Black&compound=ADDIX&sort=article&page=1", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Data struct {
			Items         []json.RawMessage `json:"items"`
			Page          int               `json:"page"`
			Total         int               `json:"total"`
			FilterOptions struct {
				ModelNames []struct {
					Value string `json:"value"`
				} `json:"model_names"`
				NominalTireWidthsMM []struct {
					Value int `json:"value"`
				} `json:"nominal_tire_widths_mm"`
				Colors []struct {
					Value string `json:"value"`
				} `json:"colors"`
				EBikeRatings []struct {
					Value *string `json:"value"`
				} `json:"e_bike_ratings"`
			} `json:"filter_options"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.Page != 1 || payload.Data.Total != 1 || len(payload.Data.Items) != 1 {
		t.Fatalf("expected the selected facets to return one item on page 1, got page=%d total=%d items=%d", payload.Data.Page, payload.Data.Total, len(payload.Data.Items))
	}
	if len(payload.Data.FilterOptions.ModelNames) != 2 || len(payload.Data.FilterOptions.Colors) != 2 {
		t.Fatalf("expected filter options to reflect both text-search matches, got models=%v colors=%v", payload.Data.FilterOptions.ModelNames, payload.Data.FilterOptions.Colors)
	}
	if len(payload.Data.FilterOptions.NominalTireWidthsMM) != 2 || payload.Data.FilterOptions.NominalTireWidthsMM[0].Value != 40 || payload.Data.FilterOptions.NominalTireWidthsMM[1].Value != 100 {
		t.Fatalf("expected both text-search widths in options, got %v", payload.Data.FilterOptions.NominalTireWidthsMM)
	}
	if len(payload.Data.FilterOptions.EBikeRatings) != 2 || payload.Data.FilterOptions.EBikeRatings[1].Value != nil {
		t.Fatalf("expected the E-Bike options to include the published E-25 and unrated value, got %v", payload.Data.FilterOptions.EBikeRatings)
	}
	if strings.Contains(recorder.Body.String(), "source_url") || strings.Contains(recorder.Body.String(), "www.schwalbe.com") {
		t.Fatalf("selector response exposed source URL: %s", recorder.Body.String())
	}

	sortingRecorder := httptest.NewRecorder()
	sortingRequest := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/products/schwalbe-tire-catalog/selector?search=green&sort=etrto", nil)
	router.ServeHTTP(sortingRecorder, sortingRequest)
	var sortingPayload struct {
		Data struct {
			Items []struct {
				ETRTO string `json:"etrto"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(sortingRecorder.Body.Bytes(), &sortingPayload); err != nil {
		t.Fatalf("decode sorted response: %v", err)
	}
	if len(sortingPayload.Data.Items) != 2 || sortingPayload.Data.Items[0].ETRTO != "40-622" || sortingPayload.Data.Items[1].ETRTO != "100-622" {
		t.Fatalf("expected numeric ETRTO sort order 40-622 then 100-622, got %v", sortingPayload.Data.Items)
	}
}

func TestSearchSchwalbeTireCatalogSelectorReturnsOnlyRequestedPage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newSchwalbeTireCatalogHandlerTestDB(t)
	for index := 1; index <= 21; index++ {
		articleNo := fmt.Sprintf("%d", index)
		modelName := fmt.Sprintf("Model %02d", index)
		if err := db.Exec(`INSERT INTO schwalbe_tire_specifications (article_no, model_name, etrto, color, source_url, source_checked_at) VALUES (?, ?, '40-622', 'Black', 'https://www.schwalbe.com/example', '2026-09-28')`, articleNo, modelName).Error; err != nil {
			t.Fatalf("insert test row %d: %v", index, err)
		}
	}

	router := gin.New()
	handler := NewHandler(service.NewProductService(repository.NewProductRepository(db), nil, 0))
	router.GET("/products/schwalbe-tire-catalog/selector", handler.SearchSchwalbeTireCatalogSelector)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/products/schwalbe-tire-catalog/selector?page=2&sort=article", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Data struct {
			Items []struct {
				ArticleNo string `json:"article_no"`
			} `json:"items"`
			Page          int `json:"page"`
			PageSize      int `json:"page_size"`
			Total         int `json:"total"`
			TotalPages    int `json:"total_pages"`
			FilterOptions struct {
				ModelNames []struct {
					Value string `json:"value"`
				} `json:"model_names"`
			} `json:"filter_options"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.Page != 2 || payload.Data.PageSize != 20 || payload.Data.Total != 21 || payload.Data.TotalPages != 2 || len(payload.Data.Items) != 1 {
		t.Fatalf("expected only the last row on page 2, got page=%d size=%d total=%d pages=%d items=%d", payload.Data.Page, payload.Data.PageSize, payload.Data.Total, payload.Data.TotalPages, len(payload.Data.Items))
	}
	if len(payload.Data.FilterOptions.ModelNames) != 21 {
		t.Fatalf("expected options from the full search result, got %d names", len(payload.Data.FilterOptions.ModelNames))
	}
	if payload.Data.Items[0].ArticleNo != "21" {
		t.Fatalf("expected natural Article No. sorting to place article 21 on page 2, got %s", payload.Data.Items[0].ArticleNo)
	}
}

func TestSearchSchwalbeTireCatalogSelectorSortsByWeightWithUnknownValuesLast(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newSchwalbeTireCatalogHandlerTestDB(t)
	statements := []string{
		`INSERT INTO schwalbe_tire_specifications (article_no, model_name, etrto, weight_g, source_url, source_checked_at) VALUES ('light', 'Light tire', '40-622', 450, 'https://www.schwalbe.com/light', '2026-09-28')`,
		`INSERT INTO schwalbe_tire_specifications (article_no, model_name, etrto, weight_g, source_url, source_checked_at) VALUES ('heavy', 'Heavy tire', '40-622', 950, 'https://www.schwalbe.com/heavy', '2026-09-28')`,
		`INSERT INTO schwalbe_tire_specifications (article_no, model_name, etrto, weight_g, source_url, source_checked_at) VALUES ('unknown', 'Unknown tire', '40-622', NULL, 'https://www.schwalbe.com/unknown', '2026-09-28')`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("insert test data: %v", err)
		}
	}

	router := gin.New()
	handler := NewHandler(service.NewProductService(repository.NewProductRepository(db), nil, 0))
	router.GET("/products/schwalbe-tire-catalog/selector", handler.SearchSchwalbeTireCatalogSelector)

	readArticleOrder := func(sortValue string) []string {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/products/schwalbe-tire-catalog/selector?sort="+sortValue, nil)
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200 for sort %s, got %d: %s", sortValue, recorder.Code, recorder.Body.String())
		}
		var payload struct {
			Data struct {
				Items []struct {
					ArticleNo string `json:"article_no"`
				} `json:"items"`
			} `json:"data"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode sorted response for %s: %v", sortValue, err)
		}
		articles := make([]string, 0, len(payload.Data.Items))
		for _, item := range payload.Data.Items {
			articles = append(articles, item.ArticleNo)
		}
		return articles
	}

	if got := readArticleOrder("weight_asc"); strings.Join(got, ",") != "light,heavy,unknown" {
		t.Fatalf("expected light-to-heavy order with unknown last, got %v", got)
	}
	if got := readArticleOrder(""); strings.Join(got, ",") != "light,heavy,unknown" {
		t.Fatalf("expected the default order to be light-to-heavy with unknown last, got %v", got)
	}
	if got := readArticleOrder("weight_desc"); strings.Join(got, ",") != "heavy,light,unknown" {
		t.Fatalf("expected heavy-to-light order with unknown last, got %v", got)
	}
}

func TestSearchSchwalbeTireCatalogSelectorFiltersByMinimumLoadCapacity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newSchwalbeTireCatalogHandlerTestDB(t)
	statements := []string{
		`INSERT INTO schwalbe_tire_specifications (article_no, model_name, etrto, load_kg, source_url, source_checked_at) VALUES ('11111111', 'Below Threshold', '40-622', 90.49, 'https://www.schwalbe.com/below', '2026-09-28')`,
		`INSERT INTO schwalbe_tire_specifications (article_no, model_name, etrto, load_kg, source_url, source_checked_at) VALUES ('22222222', 'At Threshold', '40-622', 90.5, 'https://www.schwalbe.com/equal', '2026-09-28')`,
		`INSERT INTO schwalbe_tire_specifications (article_no, model_name, etrto, load_kg, source_url, source_checked_at) VALUES ('33333333', 'Above Threshold', '40-622', 100, 'https://www.schwalbe.com/above', '2026-09-28')`,
		`INSERT INTO schwalbe_tire_specifications (article_no, model_name, etrto, load_kg, source_url, source_checked_at) VALUES ('44444444', 'Unknown Load', '40-622', NULL, 'https://www.schwalbe.com/unknown', '2026-09-28')`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("insert test data: %v", err)
		}
	}

	router := gin.New()
	handler := NewHandler(service.NewProductService(repository.NewProductRepository(db), nil, 0))
	router.GET("/products/schwalbe-tire-catalog/selector", handler.SearchSchwalbeTireCatalogSelector)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/products/schwalbe-tire-catalog/selector?min_load_kg=90.5&sort=article", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Data struct {
			Items []struct {
				ArticleNo string   `json:"article_no"`
				LoadKG    *float64 `json:"load_kg"`
			} `json:"items"`
			Total int `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.Total != 2 || len(payload.Data.Items) != 2 {
		t.Fatalf("expected the equal and above-threshold rows, got total=%d items=%v", payload.Data.Total, payload.Data.Items)
	}
	if payload.Data.Items[0].ArticleNo != "22222222" || payload.Data.Items[0].LoadKG == nil || *payload.Data.Items[0].LoadKG != 90.5 {
		t.Fatalf("expected the row at the threshold to be included, got %+v", payload.Data.Items[0])
	}
	if payload.Data.Items[1].ArticleNo != "33333333" || payload.Data.Items[1].LoadKG == nil || *payload.Data.Items[1].LoadKG != 100 {
		t.Fatalf("expected the row above the threshold to be included, got %+v", payload.Data.Items[1])
	}
}

func newSchwalbeTireCatalogHandlerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	statements := []string{
		`CREATE TABLE schwalbe_tire_specifications (article_no TEXT PRIMARY KEY, ean TEXT, model_name TEXT, etrto TEXT, inch_designation TEXT, weight_g REAL, version_label TEXT, compound TEXT, color TEXT, bead TEXT, e_bike_rating TEXT, epi INTEGER, load_kg REAL, seal TEXT, tread TEXT, min_pressure_bar REAL, max_pressure_bar REAL, min_pressure_psi REAL, max_pressure_psi REAL, source_url TEXT, source_checked_at DATE)`,
		`CREATE TABLE products (id INTEGER PRIMARY KEY, product_specification_template_id INTEGER)`,
		`CREATE TABLE product_specification_templates (id INTEGER PRIMARY KEY, slug TEXT)`,
		`CREATE TABLE product_spec_definitions (id INTEGER PRIMARY KEY, product_specification_template_id INTEGER, slug TEXT)`,
		`CREATE TABLE product_spec_values (product_id INTEGER, spec_definition_id INTEGER, value TEXT)`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("prepare test schema: %v", err)
		}
	}
	return db
}
