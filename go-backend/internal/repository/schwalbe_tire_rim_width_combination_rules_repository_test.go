package repository

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestListSchwalbeTireRimWidthCombinationRulesOrdersOfficialRanges(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	statements := []string{
		`CREATE TABLE schwalbe_tire_rim_width_combination_rules (id INTEGER PRIMARY KEY AUTOINCREMENT, tire_width_min_mm INTEGER NOT NULL, tire_width_max_mm INTEGER NOT NULL, inner_rim_width_min_mm INTEGER NOT NULL, inner_rim_width_max_mm INTEGER NOT NULL, source_basis TEXT NOT NULL, source_version TEXT NOT NULL, source_url TEXT NOT NULL, source_checked_at DATE NOT NULL)`,
		`INSERT INTO schwalbe_tire_rim_width_combination_rules (tire_width_min_mm, tire_width_max_mm, inner_rim_width_min_mm, inner_rim_width_max_mm, source_basis, source_version, source_url, source_checked_at) VALUES (47, 57, 17, 30, 'ETRTO Standard 2024; Schwalbe possible combinations guidance', '05/2024', 'https://example.test/rim-width.pdf', '2026-09-28')`,
		`INSERT INTO schwalbe_tire_rim_width_combination_rules (tire_width_min_mm, tire_width_max_mm, inner_rim_width_min_mm, inner_rim_width_max_mm, source_basis, source_version, source_url, source_checked_at) VALUES (20, 21, 15, 17, 'ETRTO Standard 2024; Schwalbe possible combinations guidance', '05/2024', 'https://example.test/rim-width.pdf', '2026-09-28')`,
	}
	for _, statement := range statements {
		require.NoError(t, db.Exec(statement).Error)
	}

	rules, err := NewProductRepository(db).ListSchwalbeTireRimWidthCombinationRules()
	require.NoError(t, err)
	require.Len(t, rules, 2)
	require.Equal(t, 20, rules[0].TireWidthMinMM)
	require.Equal(t, 17, rules[0].InnerRimWidthMaxMM)
	require.Equal(t, 47, rules[1].TireWidthMinMM)
	require.Equal(t, "05/2024", rules[1].SourceVersion)
}
