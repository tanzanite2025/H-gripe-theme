package product

import (
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

func TestListSchwalbeTireRimWidthCombinationRulesReturnsSourceBackedRules(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE schwalbe_tire_rim_width_combination_rules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tire_width_min_mm INTEGER NOT NULL,
			tire_width_max_mm INTEGER NOT NULL,
			inner_rim_width_min_mm INTEGER NOT NULL,
			inner_rim_width_max_mm INTEGER NOT NULL,
			source_basis TEXT NOT NULL,
			source_version TEXT NOT NULL,
			source_url TEXT NOT NULL,
			source_checked_at DATE NOT NULL
		);
		INSERT INTO schwalbe_tire_rim_width_combination_rules (
			tire_width_min_mm, tire_width_max_mm, inner_rim_width_min_mm,
			inner_rim_width_max_mm, source_basis, source_version, source_url,
			source_checked_at
		) VALUES (
			20, 21, 15, 17,
			'ETRTO Standard 2024; Schwalbe possible combinations guidance',
			'05/2024', 'https://example.test/rim-width.pdf', '2026-09-28'
		);
	`).Error; err != nil {
		t.Fatalf("seed test rule: %v", err)
	}

	router := gin.New()
	handler := NewHandler(service.NewProductService(repository.NewProductRepository(db), nil, 0))
	router.GET("/products/schwalbe-tire-rim-width-combination-rules", handler.ListSchwalbeTireRimWidthCombinationRules)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/products/schwalbe-tire-rim-width-combination-rules", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	for _, expected := range []string{
		`"code":0`,
		`"tire_width_min_mm":20`,
		`"inner_rim_width_max_mm":17`,
		`"source_version":"05/2024"`,
	} {
		if !strings.Contains(recorder.Body.String(), expected) {
			t.Fatalf("expected response to contain %s, got %s", expected, recorder.Body.String())
		}
	}
}
