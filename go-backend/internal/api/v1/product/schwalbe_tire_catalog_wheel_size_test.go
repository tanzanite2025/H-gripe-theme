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

func TestSearchSchwalbeTireCatalogSelectorFiltersByWheelDiameterAndBSDPair(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newSchwalbeTireCatalogHandlerTestDB(t)
	rows := []struct {
		articleNo       string
		etrto           string
		inchDesignation string
	}{
		{articleNo: "wheel-28", etrto: "40-622", inchDesignation: "28x1.50"},
		{articleNo: "wheel-29", etrto: "57-622", inchDesignation: "29x2.25"},
		{articleNo: "wheel-26-559", etrto: "50-559", inchDesignation: "26x2.00"},
		{articleNo: "wheel-26-590", etrto: "37-590", inchDesignation: "26x1 3/8"},
	}
	for _, row := range rows {
		if err := db.Exec(
			`INSERT INTO schwalbe_tire_specifications (article_no, model_name, etrto, inch_designation, source_url, source_checked_at) VALUES (?, ?, ?, ?, 'https://www.schwalbe.com/example', '2026-09-28')`,
			row.articleNo,
			row.articleNo,
			row.etrto,
			row.inchDesignation,
		).Error; err != nil {
			t.Fatalf("insert %s: %v", row.articleNo, err)
		}
	}

	router := gin.New()
	handler := NewHandler(service.NewProductService(repository.NewProductRepository(db), nil, 0))
	router.GET("/products/schwalbe-tire-catalog/selector", handler.SearchSchwalbeTireCatalogSelector)

	var allPayload struct {
		Data struct {
			Items []struct {
				ArticleNo string `json:"article_no"`
				WheelSize *struct {
					Value           string `json:"value"`
					WheelDiameterIn string `json:"wheel_diameter_in"`
					BSD             int    `json:"bsd_mm"`
				} `json:"wheel_size"`
			} `json:"items"`
			FilterOptions struct {
				WheelSizes []struct {
					Value           string `json:"value"`
					WheelDiameterIn string `json:"wheel_diameter_in"`
					BSD             int    `json:"bsd_mm"`
				} `json:"wheel_sizes"`
			} `json:"filter_options"`
		} `json:"data"`
	}
	allRecorder := httptest.NewRecorder()
	router.ServeHTTP(allRecorder, httptest.NewRequestWithContext(context.Background(),
		http.MethodGet,
		"/products/schwalbe-tire-catalog/selector?sort=article",
		nil,
	))
	if allRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200 for all wheel sizes, got %d: %s", allRecorder.Code, allRecorder.Body.String())
	}
	if err := json.Unmarshal(allRecorder.Body.Bytes(), &allPayload); err != nil {
		t.Fatalf("decode all wheel sizes response: %v", err)
	}
	if len(allPayload.Data.FilterOptions.WheelSizes) != 4 {
		t.Fatalf("expected four distinct wheel diameter/BSD options, got %+v", allPayload.Data.FilterOptions.WheelSizes)
	}
	expectedOptions := []struct {
		value    string
		diameter string
		bsd      int
	}{
		{value: "26-559", diameter: "26", bsd: 559},
		{value: "26-590", diameter: "26", bsd: 590},
		{value: "28-622", diameter: "28", bsd: 622},
		{value: "29-622", diameter: "29", bsd: 622},
	}
	for index, expected := range expectedOptions {
		actual := allPayload.Data.FilterOptions.WheelSizes[index]
		if actual.Value != expected.value || actual.WheelDiameterIn != expected.diameter || actual.BSD != expected.bsd {
			t.Fatalf("wheel size option %d: expected %+v, got %+v", index, expected, actual)
		}
	}
	projectedSizes := make(map[string]string, len(allPayload.Data.Items))
	for _, item := range allPayload.Data.Items {
		if item.WheelSize == nil {
			t.Fatalf("expected wheel_size projection for %s", item.ArticleNo)
		}
		projectedSizes[item.ArticleNo] = item.WheelSize.Value
	}
	for articleNo, expectedValue := range map[string]string{
		"wheel-28":     "28-622",
		"wheel-29":     "29-622",
		"wheel-26-559": "26-559",
		"wheel-26-590": "26-590",
	} {
		if projectedSizes[articleNo] != expectedValue {
			t.Fatalf("expected wheel_size projection %s=%s, got %q", articleNo, expectedValue, projectedSizes[articleNo])
		}
	}

	tests := []struct {
		name        string
		wheelSize   string
		expectedArt string
	}{
		{name: "28 inch 622 BSD", wheelSize: "28-622", expectedArt: "wheel-28"},
		{name: "29 inch 622 BSD", wheelSize: "29-622", expectedArt: "wheel-29"},
		{name: "26 inch 559 BSD", wheelSize: "26-559", expectedArt: "wheel-26-559"},
		{name: "26 inch 590 BSD", wheelSize: "26-590", expectedArt: "wheel-26-590"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var payload struct {
				Data struct {
					Items []struct {
						ArticleNo string `json:"article_no"`
					} `json:"items"`
					Total int `json:"total"`
				} `json:"data"`
			}
			recorder := httptest.NewRecorder()
			request := httptest.NewRequestWithContext(context.Background(),
				http.MethodGet,
				"/products/schwalbe-tire-catalog/selector?sort=article&wheel_size="+testCase.wheelSize,
				nil,
			)
			router.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if payload.Data.Total != 1 || len(payload.Data.Items) != 1 || payload.Data.Items[0].ArticleNo != testCase.expectedArt {
				t.Fatalf("expected only %s for wheel_size=%s, got total=%d items=%v", testCase.expectedArt, testCase.wheelSize, payload.Data.Total, payload.Data.Items)
			}
		})
	}
}
