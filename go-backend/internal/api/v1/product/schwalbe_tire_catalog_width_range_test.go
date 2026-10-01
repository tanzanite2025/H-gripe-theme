package product

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func TestSearchSchwalbeTireCatalogSelectorFiltersInclusiveNominalTireWidthRange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newSchwalbeTireCatalogHandlerTestDB(t)
	rows := []struct {
		articleNo string
		etrto     string
	}{
		{articleNo: "11111111", etrto: "35-622"},
		{articleNo: "22222222", etrto: "40-622"},
		{articleNo: "33333333", etrto: "57-622"},
		{articleNo: "44444444", etrto: "60-622"},
		{articleNo: "55555555", etrto: "unknown"},
	}
	for _, row := range rows {
		if err := db.Exec(
			`INSERT INTO schwalbe_tire_specifications (article_no, model_name, etrto, source_url, source_checked_at) VALUES (?, ?, ?, ?, '2026-09-28')`,
			row.articleNo,
			"Range Test",
			row.etrto,
			"https://www.schwalbe.com/"+row.articleNo,
		).Error; err != nil {
			t.Fatalf("insert %s: %v", row.articleNo, err)
		}
	}

	router := gin.New()
	handler := NewHandler(service.NewProductService(repository.NewProductRepository(db), nil, 0))
	router.GET("/products/schwalbe-tire-catalog/selector", handler.SearchSchwalbeTireCatalogSelector)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(context.Background(),
		http.MethodGet,
		"/products/schwalbe-tire-catalog/selector?tire_width_min_mm=40&tire_width_max_mm=57&sort=etrto",
		nil,
	)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Data struct {
			Items []struct {
				ArticleNo string `json:"article_no"`
				ETRTO     string `json:"etrto"`
			} `json:"items"`
			Total int `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.Total != 2 || len(payload.Data.Items) != 2 {
		t.Fatalf("expected only 40 mm and 57 mm rows, got total=%d items=%v", payload.Data.Total, payload.Data.Items)
	}
	if payload.Data.Items[0].ETRTO != "40-622" || payload.Data.Items[1].ETRTO != "57-622" {
		t.Fatalf("expected inclusive width endpoints in numeric order, got %v", payload.Data.Items)
	}
}

func TestSearchSchwalbeTireCatalogSelectorRangeTakesPrecedenceOverLegacyExactWidth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newSchwalbeTireCatalogHandlerTestDB(t)
	for _, row := range []struct {
		articleNo string
		etrto     string
	}{
		{articleNo: "11111111", etrto: "35-622"},
		{articleNo: "22222222", etrto: "40-622"},
	} {
		if err := db.Exec(
			`INSERT INTO schwalbe_tire_specifications (article_no, model_name, etrto, source_url, source_checked_at) VALUES (?, 'Precedence Test', ?, ?, '2026-09-28')`,
			row.articleNo,
			row.etrto,
			"https://www.schwalbe.com/"+row.articleNo,
		).Error; err != nil {
			t.Fatalf("insert %s: %v", row.articleNo, err)
		}
	}

	router := gin.New()
	handler := NewHandler(service.NewProductService(repository.NewProductRepository(db), nil, 0))
	router.GET("/products/schwalbe-tire-catalog/selector", handler.SearchSchwalbeTireCatalogSelector)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(context.Background(),
		http.MethodGet,
		"/products/schwalbe-tire-catalog/selector?tire_width_min_mm=40&tire_width_mm=35",
		nil,
	)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Data struct {
			Items []struct {
				ETRTO string `json:"etrto"`
			} `json:"items"`
			Total int `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.Total != 1 || len(payload.Data.Items) != 1 || payload.Data.Items[0].ETRTO != "40-622" {
		t.Fatalf("expected range to take precedence over legacy exact width, got %+v", payload.Data)
	}
}

func TestSearchSchwalbeTireCatalogSelectorNormalizesReversedWidthRange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newSchwalbeTireCatalogHandlerTestDB(t)
	for _, row := range []struct {
		articleNo string
		etrto     string
	}{
		{articleNo: "11111111", etrto: "35-622"},
		{articleNo: "22222222", etrto: "57-622"},
		{articleNo: "33333333", etrto: "60-622"},
	} {
		if err := db.Exec(
			`INSERT INTO schwalbe_tire_specifications (article_no, model_name, etrto, source_url, source_checked_at) VALUES (?, 'Reverse Range Test', ?, ?, '2026-09-28')`,
			row.articleNo,
			row.etrto,
			"https://www.schwalbe.com/"+row.articleNo,
		).Error; err != nil {
			t.Fatalf("insert %s: %v", row.articleNo, err)
		}
	}

	router := gin.New()
	handler := NewHandler(service.NewProductService(repository.NewProductRepository(db), nil, 0))
	router.GET("/products/schwalbe-tire-catalog/selector", handler.SearchSchwalbeTireCatalogSelector)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(context.Background(),
		http.MethodGet,
		"/products/schwalbe-tire-catalog/selector?tire_width_min_mm=57&tire_width_max_mm=35&sort=etrto",
		nil,
	)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Data struct {
			Items []struct {
				ETRTO string `json:"etrto"`
			} `json:"items"`
			Total int `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.Total != 2 || len(payload.Data.Items) != 2 {
		t.Fatalf("expected reversed range to normalize to 35..57, got total=%d items=%v", payload.Data.Total, payload.Data.Items)
	}
}
