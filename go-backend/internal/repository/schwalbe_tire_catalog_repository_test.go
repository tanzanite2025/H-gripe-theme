package repository

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestListSchwalbeTireCatalogKeepsCandidatesAndMarksMatchingProducts(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	statements := []string{
		`CREATE TABLE schwalbe_tire_specifications (article_no TEXT PRIMARY KEY, ean TEXT, model_name TEXT, etrto TEXT, inch_designation TEXT, weight_g REAL, version_label TEXT, compound TEXT, color TEXT, bead TEXT, e_bike_rating TEXT, epi INTEGER, load_kg REAL, seal TEXT, tread TEXT, min_pressure_bar REAL, max_pressure_bar REAL, min_pressure_psi REAL, max_pressure_psi REAL, source_url TEXT, source_checked_at DATE)`,
		`CREATE TABLE products (id INTEGER PRIMARY KEY, product_specification_template_id INTEGER)`,
		`CREATE TABLE product_specification_templates (id INTEGER PRIMARY KEY, slug TEXT)`,
		`CREATE TABLE product_spec_definitions (id INTEGER PRIMARY KEY, product_specification_template_id INTEGER, slug TEXT)`,
		`CREATE TABLE product_spec_values (product_id INTEGER, spec_definition_id INTEGER, value TEXT)`,
		`INSERT INTO schwalbe_tire_specifications (article_no, model_name, etrto, weight_g, source_url, source_checked_at) VALUES ('11111111', 'Green Marathon', '40-622', 650, 'https://www.schwalbe.com/example', '2026-09-28')`,
		`INSERT INTO schwalbe_tire_specifications (article_no, model_name, etrto, source_url, source_checked_at) VALUES ('22222222', 'Marathon Plus', '37-622', 'https://www.schwalbe.com/example', '2026-09-28')`,
		`INSERT INTO products (id, product_specification_template_id) VALUES (7, 1)`,
		`INSERT INTO product_specification_templates (id, slug) VALUES (1, 'schwalbe_tire')`,
		`INSERT INTO product_spec_definitions (id, product_specification_template_id, slug) VALUES (8, 1, 'article_no')`,
		`INSERT INTO product_spec_values (product_id, spec_definition_id, value) VALUES (7, 8, ' 11111111 ')`,
	}
	for _, statement := range statements {
		require.NoError(t, db.Exec(statement).Error)
	}

	items, err := NewProductRepository(db).ListSchwalbeTireCatalog("")
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.Equal(t, "11111111", items[0].ArticleNo)
	require.True(t, items[0].ProductExists)
	require.NotNil(t, items[0].WeightG)
	require.Equal(t, 650.0, *items[0].WeightG)
	require.False(t, items[1].ProductExists)

	filtered, err := NewProductRepository(db).ListSchwalbeTireCatalog("plus")
	require.NoError(t, err)
	require.Len(t, filtered, 1)
	require.Equal(t, "22222222", filtered[0].ArticleNo)
}

func TestListSchwalbeTireCatalogTreatsSearchLikeMetacharactersLiterally(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	statements := []string{
		`CREATE TABLE schwalbe_tire_specifications (article_no TEXT PRIMARY KEY, ean TEXT, model_name TEXT, etrto TEXT, inch_designation TEXT, weight_g REAL, version_label TEXT, compound TEXT, color TEXT, bead TEXT, e_bike_rating TEXT, epi INTEGER, load_kg REAL, seal TEXT, tread TEXT, min_pressure_bar REAL, max_pressure_bar REAL, min_pressure_psi REAL, max_pressure_psi REAL, source_url TEXT, source_checked_at DATE)`,
		`CREATE TABLE products (id INTEGER PRIMARY KEY, product_specification_template_id INTEGER)`,
		`CREATE TABLE product_specification_templates (id INTEGER PRIMARY KEY, slug TEXT)`,
		`CREATE TABLE product_spec_definitions (id INTEGER PRIMARY KEY, product_specification_template_id INTEGER, slug TEXT)`,
		`CREATE TABLE product_spec_values (product_id INTEGER, spec_definition_id INTEGER, value TEXT)`,
		`INSERT INTO schwalbe_tire_specifications (article_no, model_name, etrto, source_url, source_checked_at) VALUES ('10000001', 'Green Marathon', '40-622', 'https://example.test/1', '2026-09-28')`,
		`INSERT INTO schwalbe_tire_specifications (article_no, model_name, etrto, source_url, source_checked_at) VALUES ('10000002', 'Marathon_Plus', '40-622', 'https://example.test/2', '2026-09-28')`,
		`INSERT INTO schwalbe_tire_specifications (article_no, model_name, etrto, source_url, source_checked_at) VALUES ('10000003', 'Marathon 100%', '40-622', 'https://example.test/3', '2026-09-28')`,
		`INSERT INTO schwalbe_tire_specifications (article_no, model_name, etrto, source_url, source_checked_at) VALUES ('10000004', 'Ride!Line', '40-622', 'https://example.test/4', '2026-09-28')`,
	}
	for _, statement := range statements {
		require.NoError(t, db.Exec(statement).Error)
	}

	for searchTerm, expectedArticleNo := range map[string]string{
		"_": "10000002",
		"%": "10000003",
		"!": "10000004",
	} {
		items, err := NewProductRepository(db).ListSchwalbeTireCatalog(searchTerm)
		require.NoError(t, err)
		require.Len(t, items, 1, "search term %q should match only its literal occurrence", searchTerm)
		require.Equal(t, expectedArticleNo, items[0].ArticleNo)
	}
}
