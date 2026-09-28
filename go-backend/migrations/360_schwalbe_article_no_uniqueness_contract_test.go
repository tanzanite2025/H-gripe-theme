package migrations_test

import (
	"strings"
	"testing"
)

func TestSchwalbeArticleNoUniquenessMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "360_enforce_schwalbe_article_no_uniqueness.up.sql"))
	for _, required := range []string{
		"PRODUCT_SPECIFICATION_TEMPLATES",
		"SCHWALBE_TIRE",
		"ARTICLE_NO",
		"DUPLICATE SCHWALBE ARTICLE NO.",
		"CREATE UNIQUE INDEX IF NOT EXISTS UQ_SCHWALBE_PRODUCT_ARTICLE_NO",
		"LOWER(BTRIM(VALUE))",
		"SPEC_DEFINITION_ID",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("Article No. uniqueness migration is missing contract fragment %q", required)
		}
	}

	down := strings.ToUpper(strings.TrimSpace(readMigrationFile(t, "360_enforce_schwalbe_article_no_uniqueness.down.sql")))
	if !strings.Contains(down, "DROP INDEX IF EXISTS UQ_SCHWALBE_PRODUCT_ARTICLE_NO") {
		t.Fatal("Article No. uniqueness down migration must remove only its own index")
	}
}
