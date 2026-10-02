package repository

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestListSchwalbeTireHooklessCompatibilitiesReturnsEmptyWhenTableIsMissing(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	items, err := NewProductRepository(db).ListSchwalbeTireHooklessCompatibilities()
	require.NoError(t, err)
	require.Empty(t, items)
}

func TestListSchwalbeTireHooklessCompatibilitiesOrdersArticleOverridesBeforeModelFallbacks(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE schwalbe_tire_hookless_compatibility (
		scope_key TEXT PRIMARY KEY,
		article_no TEXT,
		model_name TEXT,
		status TEXT NOT NULL,
		source_basis TEXT NOT NULL,
		source_version TEXT NOT NULL,
		source_url TEXT NOT NULL,
		source_checked_at DATE NOT NULL
	)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO schwalbe_tire_hookless_compatibility
		(scope_key, article_no, model_name, status, source_basis, source_version, source_url, source_checked_at)
		VALUES
		('model:Rocket Ron', NULL, 'Rocket Ron', 'unknown', 'conflicting official filters', '2026-09-28', 'https://example.test/no', '2026-09-28'),
		('article:11600385.03', '11600385.03', 'Rocket Ron', 'supported', 'article source', '2026-09-28', 'https://example.test/yes', '2026-09-28')`).Error)

	items, err := NewProductRepository(db).ListSchwalbeTireHooklessCompatibilities()
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.Equal(t, "11600385.03", *items[0].ArticleNo)
	require.Equal(t, "Rocket Ron", *items[1].ModelName)
}
