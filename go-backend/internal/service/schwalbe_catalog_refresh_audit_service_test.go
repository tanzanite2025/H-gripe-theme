package service

import (
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestLoadSchwalbeActiveSalesProductsReadsDedicatedProductSchema(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	statements := []string{
		`CREATE TABLE products (id INTEGER PRIMARY KEY, name TEXT, product_type TEXT, brand_id INTEGER, status TEXT, deleted_at DATETIME, parent_id INTEGER, bicycle_tire_specification TEXT, standard_product_identifier TEXT)`,
		`CREATE TABLE product_brands (id INTEGER PRIMARY KEY, slug TEXT)`,
		`CREATE TABLE product_variants (id INTEGER PRIMARY KEY, product_id INTEGER, is_active BOOLEAN, deleted_at DATETIME)`,
		`INSERT INTO product_brands VALUES (1, 'schwalbe'), (2, 'other-brand')`,
		`INSERT INTO products VALUES
			(1, 'Schwalbe Tire', 'bicycle_tire', 1, 'active', NULL, NULL, '{"schema_version":1,"article_no":" 11100001 ","model_name":"Example","etrto":"40-622","weight_g":650,"epi":67}', ' 4026495000000 '),
			(2, 'Other Brand Tire', 'bicycle_tire', 2, 'active', NULL, NULL, '{"schema_version":1,"article_no":"22200002"}', '4026495000001'),
			(3, 'Inactive Schwalbe Tire', 'bicycle_tire', 1, 'inactive', NULL, NULL, '{"schema_version":1,"article_no":"33300003"}', '')`,
		`INSERT INTO product_variants VALUES (1, 1, TRUE, NULL), (2, 2, TRUE, NULL), (3, 3, TRUE, NULL)`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}

	var products []schwalbeSalesProduct
	if err := loadSchwalbeActiveSalesProducts(db, &products); err != nil {
		t.Fatal(err)
	}
	if len(products) != 1 {
		t.Fatalf("expected only one active root Schwalbe tire product, got %#v", products)
	}
	if products[0].Values["article_no"] == nil || *products[0].Values["article_no"] != " 11100001 " {
		t.Fatalf("dedicated Article No. was not read from product JSON: %#v", products[0].Values)
	}
	if products[0].Values["ean"] == nil || *products[0].Values["ean"] != "4026495000000" {
		t.Fatalf("EAN was not read from standard_product_identifier: %#v", products[0].Values)
	}
	if products[0].Values["weight_g"] == nil || *products[0].Values["weight_g"] != "650" {
		t.Fatalf("numeric JSON specification was not read from the dedicated product schema: %#v", products[0].Values)
	}
}

func TestBuildSchwalbeCatalogRefreshAuditReportSeparatesCatalogAndSalesChanges(t *testing.T) {
	current := schwalbeAuditRow("11100001", 650, "white", "37–622", "https://old.example/1", "2026-09-28")
	incoming := schwalbeAuditRow("11100001", 650.5, "White", "37-622", "https://new.example/1", "2026-09-29")
	added := schwalbeAuditRow("11100002", 700, "black", "40-622", "https://new.example/2", "2026-09-29")
	removed := schwalbeAuditRow("11100003", 710, "black", "42-622", "https://old.example/3", "2026-09-28")

	currentByArticle, err := indexSchwalbeRefreshSnapshot([]SchwalbeCatalogRefreshSnapshotRow{current, removed}, false)
	if err != nil {
		t.Fatal(err)
	}
	incomingByArticle, err := indexSchwalbeRefreshSnapshot([]SchwalbeCatalogRefreshSnapshotRow{incoming, added}, true)
	if err != nil {
		t.Fatal(err)
	}
	product := schwalbeSalesProduct{
		ID:   42,
		Name: "Schwalbe Example 37-622",
		Values: map[string]*string{
			"article_no":       schwalbeAuditStringPtr("11100001"),
			"weight_g":         schwalbeAuditStringPtr("650.00"),
			"epi":              schwalbeAuditStringPtr("67"),
			"color":            schwalbeAuditStringPtr("white"),
			"etrto":            schwalbeAuditStringPtr("37-622"),
			"model_name":       schwalbeAuditStringPtr("Example"),
			"inch_designation": nil,
			"version_label":    nil,
			"compound":         nil,
			"bead":             nil,
			"e_bike_rating":    nil,
			"load_kg":          nil,
			"ean":              nil,
			"seal":             nil,
			"tread":            nil,
			"min_pressure_bar": nil,
			"max_pressure_bar": nil,
			"min_pressure_psi": nil,
			"max_pressure_psi": nil,
		},
	}
	missingSnapshotProduct := schwalbeSalesProduct{
		ID: 43, Name: "Schwalbe Removed 42-622", Values: schwalbeCatalogValues(removed),
	}
	missingSnapshotProduct.Values["article_no"] = schwalbeAuditStringPtr("11100003")

	report := buildSchwalbeCatalogRefreshAuditReport(currentByArticle, incomingByArticle, []schwalbeSalesProduct{product, missingSnapshotProduct})
	if len(report.Catalog.Added) != 1 || report.Catalog.Added[0] != "11100002" {
		t.Fatalf("expected added Article No. 11100002, got %#v", report.Catalog.Added)
	}
	if len(report.Catalog.NotInIncomingSnapshot) != 1 || report.Catalog.NotInIncomingSnapshot[0] != "11100003" {
		t.Fatalf("expected Article No. 11100003 absent from snapshot, got %#v", report.Catalog.NotInIncomingSnapshot)
	}
	if len(report.Catalog.Changed) != 1 || len(report.Catalog.Changed[0].Fields) != 2 {
		t.Fatalf("expected weight and case-sensitive color changes, got %#v", report.Catalog.Changed)
	}
	if report.Catalog.Changed[0].Fields[0].Field != "weight_g" || report.Catalog.Changed[0].Fields[1].Field != "color" {
		t.Fatalf("catalog field changes are not in contract order: %#v", report.Catalog.Changed[0].Fields)
	}
	if len(report.Catalog.SourceMetadata.SourceURLChanges) != 1 || report.Catalog.SourceMetadata.CheckedAtChangedArticles != 1 {
		t.Fatalf("expected provenance changes to be reported separately: %#v", report.Catalog.SourceMetadata)
	}
	if len(report.Sales.SpecificationDifferences) != 2 {
		t.Fatalf("expected both changed and missing-snapshot sales products to be listed, got %#v", report.Sales.SpecificationDifferences)
	}
	difference := report.Sales.SpecificationDifferences[0]
	if len(difference.DiffersFromCurrentCatalog) != 0 {
		t.Fatalf("product should match the current official snapshot, got %#v", difference.DiffersFromCurrentCatalog)
	}
	if !difference.RequiresSalesProductReview || len(difference.DiffersFromIncomingSnapshot) != 2 {
		t.Fatalf("incoming changes should require sales product review: %#v", difference)
	}
	if len(difference.OfficialFieldsChangedByRefresh) != 2 {
		t.Fatalf("expected both incoming official changes to be tied to the sales product: %#v", difference.OfficialFieldsChangedByRefresh)
	}
	missingSnapshotDifference := report.Sales.SpecificationDifferences[1]
	if !missingSnapshotDifference.IncomingSnapshotMissing || !missingSnapshotDifference.RequiresSalesProductReview {
		t.Fatalf("a sales Article No. absent from the incoming snapshot must require review: %#v", missingSnapshotDifference)
	}
	if len(missingSnapshotDifference.DiffersFromCurrentCatalog) != 0 {
		t.Fatalf("the second product should match its current catalog row: %#v", missingSnapshotDifference.DiffersFromCurrentCatalog)
	}
}

func TestSchwalbeAuditValueComparisonUsesNumericValuesAndOfficialText(t *testing.T) {
	if !auditValuesEqual("weight_g", schwalbeAuditStringPtr("650"), schwalbeAuditStringPtr("650.00")) {
		t.Fatal("numeric formatting should not create a difference")
	}
	if !auditValuesEqual("etrto", schwalbeAuditStringPtr("37–622"), schwalbeAuditStringPtr("37-622")) {
		t.Fatal("ETRTO dash variants should compare as the same normalized size")
	}
	if auditValuesEqual("color", schwalbeAuditStringPtr("Black"), schwalbeAuditStringPtr("black")) {
		t.Fatal("official text comparison must preserve case differences")
	}
	if auditValuesEqual("weight_g", nil, schwalbeAuditStringPtr("0")) {
		t.Fatal("an unknown measurement must differ from a populated zero")
	}
}

func TestIndexSchwalbeRefreshSnapshotRejectsUnsafeInput(t *testing.T) {
	if _, err := indexSchwalbeRefreshSnapshot(nil, true); err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("expected an empty snapshot error, got %v", err)
	}

	duplicateRows := []SchwalbeCatalogRefreshSnapshotRow{
		schwalbeAuditRow("11100001", 650, "black", "37-622", "https://example/1", "2026-09-29"),
		schwalbeAuditRow("\uFEFF11100001\u200B", 650, "black", "37-622", "https://example/2", "2026-09-29"),
	}
	if _, err := indexSchwalbeRefreshSnapshot(duplicateRows, true); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected a duplicate Article No. error, got %v", err)
	}

	zeroWeight := schwalbeAuditRow("11100002", 0, "black", "37-622", "https://example/2", "2026-09-29")
	if _, err := indexSchwalbeRefreshSnapshot([]SchwalbeCatalogRefreshSnapshotRow{zeroWeight}, true); err == nil || !strings.Contains(err.Error(), "positive measurement") {
		t.Fatalf("expected a positive measurement error, got %v", err)
	}

	fractionalEPI := schwalbeAuditRow("11100003", 650, "black", "37-622", "https://example/3", "2026-09-29")
	fractionalEPI.EPI = floatValue(67.5)
	if _, err := indexSchwalbeRefreshSnapshot([]SchwalbeCatalogRefreshSnapshotRow{fractionalEPI}, true); err == nil || !strings.Contains(err.Error(), "non-integer EPI") {
		t.Fatalf("expected a non-integer EPI error, got %v", err)
	}
}

func schwalbeAuditRow(articleNo string, weight float64, color, etrto, sourceURL, checkedAt string) SchwalbeCatalogRefreshSnapshotRow {
	return SchwalbeCatalogRefreshSnapshotRow{
		ArticleNo: articleNo, EAN: nil, ModelName: "Example", ETRTO: etrto, InchDesignation: nil,
		WeightG: floatValue(weight), VersionLabel: nil, Compound: nil, Color: schwalbeAuditStringPtr(color), Bead: nil,
		EBikeRating: nil, EPI: floatValue(67), LoadKG: nil, Seal: nil, Tread: nil,
		MinPressureBar: nil, MaxPressureBar: nil, MinPressurePSI: nil, MaxPressurePSI: nil,
		SourceURL: schwalbeAuditStringPtr(sourceURL), SourceCheckedAt: checkedAt,
	}
}

func floatValue(value float64) *float64 {
	return &value
}
