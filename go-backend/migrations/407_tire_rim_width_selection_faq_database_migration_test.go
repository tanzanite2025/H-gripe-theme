package migrations_test

import (
	"context"
	"database/sql"
	"net"
	"net/url"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func createIsolatedTireRimWidthSelectionFAQMigrationTransaction(t *testing.T) *sql.Tx {
	t.Helper()
	host, username, password := os.Getenv("DB_HOST"), os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD")
	if host == "" || username == "" || password == "" {
		t.Skip("PostgreSQL integration environment is not configured")
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	databaseName := os.Getenv("DB_NAME")
	if databaseName == "" {
		databaseName = "postgres"
	}
	connectionURL := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(username, password),
		Host:     net.JoinHostPort(host, port),
		Path:     "/" + databaseName,
		RawQuery: "sslmode=disable&connect_timeout=5",
	}
	db, err := sql.Open("postgres", connectionURL.String())
	if err != nil {
		t.Fatal("open migration test database:", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	transaction, err := db.BeginTx(ctx, nil)
	if err != nil {
		cancel()
		_ = db.Close()
		t.Fatal("begin isolated FAQ migration transaction:", err)
	}
	t.Cleanup(func() {
		_ = transaction.Rollback()
		cancel()
		_ = db.Close()
	})
	// Temporary tables keep the test independent of the live FAQ data and schema.
	// No public sequences or real application rows are modified.
	_, err = transaction.ExecContext(ctx, `
		CREATE TEMP TABLE faq_pages (
			id BIGSERIAL PRIMARY KEY, page_id TEXT NOT NULL, route_path TEXT NOT NULL DEFAULT '',
			route_key TEXT, manifest_version TEXT, route_status TEXT NOT NULL DEFAULT 'current',
			last_seen_at TIMESTAMPTZ, domain TEXT, locale TEXT NOT NULL,
			title TEXT NOT NULL DEFAULT '', subtitle TEXT, sort_order INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'active', created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), deleted_at TIMESTAMPTZ,
			UNIQUE (page_id, locale)
		) ON COMMIT DROP;
		CREATE TEMP TABLE faqs (
			id BIGSERIAL PRIMARY KEY, page_id TEXT, locale TEXT, question TEXT NOT NULL,
			answer TEXT NOT NULL, answer_image_url TEXT NOT NULL DEFAULT '',
			answer_image_alt TEXT NOT NULL DEFAULT '', answer_image_width INTEGER NOT NULL DEFAULT 0,
			answer_image_height INTEGER NOT NULL DEFAULT 0, status TEXT NOT NULL DEFAULT 'published',
			"order" INTEGER NOT NULL DEFAULT 0, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), deleted_at TIMESTAMPTZ
		) ON COMMIT DROP;
		SET LOCAL search_path TO pg_temp;
	`)
	if err != nil {
		t.Fatal("create isolated FAQ migration tables:", err)
	}
	return transaction
}

func executeTireRimWidthSelectionFAQMigrationSQL(t *testing.T, transaction *sql.Tx, filename string) {
	t.Helper()
	if _, err := transaction.Exec(readMigrationFile(t, filename)); err != nil {
		t.Fatalf("execute FAQ migration %s: %v", filename, err)
	}
}

func TestTireRimWidthSelectionFAQMigrationPreservesAdminEditsAndSupportsSafeRollback(t *testing.T) {
	transaction := createIsolatedTireRimWidthSelectionFAQMigrationTransaction(t)
	_, err := transaction.Exec(`
		INSERT INTO faq_pages (page_id, locale, title, subtitle, status, sort_order)
		VALUES
			('guides-tireguides-choose', 'en', 'Administrator title', 'Administrator subtitle', 'hidden', 99),
			('guides-tireguides-choose', 'zh_cn', 'Tire and rim width reference calculator FAQs',
			 'Convert inch tire widths and resolve DT Swiss rim internal width references', 'active', 250);
		INSERT INTO faqs (page_id, locale, question, answer)
		VALUES ('other-guide', 'en', 'Keep another guide', '<p>Other content</p>');
	`)
	if err != nil {
		t.Fatal(err)
	}

	for run := 0; run < 2; run++ {
		executeTireRimWidthSelectionFAQMigrationSQL(t, transaction, "407_add_tire_rim_width_selection_editable_faq_content.up.sql")
	}
	var pageCount, itemCount int
	if err := transaction.QueryRow("SELECT count(*) FROM faq_pages").Scan(&pageCount); err != nil {
		t.Fatal(err)
	}
	if pageCount != 20 {
		t.Fatalf("expected 20 route-owned locale shells, got %d", pageCount)
	}
	if err := transaction.QueryRow("SELECT count(*) FROM faqs WHERE page_id = 'guides-tireguides-choose'").Scan(&itemCount); err != nil {
		t.Fatal(err)
	}
	if itemCount != 18 {
		t.Fatalf("expected nine FAQ items in each of two locales without duplicates, got %d", itemCount)
	}
	var title, subtitle, status, routePath string
	var sortOrder int
	if err := transaction.QueryRow("SELECT title, subtitle, status, sort_order, route_path FROM faq_pages WHERE locale = 'en'").
		Scan(&title, &subtitle, &status, &sortOrder, &routePath); err != nil {
		t.Fatal(err)
	}
	if title != "Administrator title" || subtitle != "Administrator subtitle" || status != "hidden" || sortOrder != 99 || routePath != "/guides/tireguides/choose" {
		t.Fatal("up migration changed administrator metadata or failed to bind the page route")
	}
	if err := transaction.QueryRow("SELECT title, subtitle FROM faq_pages WHERE locale = 'zh_cn'").Scan(&title, &subtitle); err != nil {
		t.Fatal(err)
	}
	if title != "外胎与车圈宽度选择 FAQ" || subtitle != "了解尺寸选择、标准正常配合范围、最佳宽度与模拟图示" {
		t.Fatal("legacy generated DT metadata was not replaced with the current Chinese page description")
	}

	_, err = transaction.Exec(`
		UPDATE faqs SET question = 'Administrator renamed question', answer = '<p>Edited answer</p>',
			updated_at = NOW() + INTERVAL '1 second'
			WHERE page_id = 'guides-tireguides-choose' AND locale = 'en' AND "order" = 110;
		UPDATE faqs SET status = 'draft'
			WHERE page_id = 'guides-tireguides-choose' AND locale = 'zh_cn' AND "order" = 120;
		UPDATE faqs SET answer_image_url = '/uploads/faq/admin-image.webp'
			WHERE page_id = 'guides-tireguides-choose' AND locale = 'zh_cn' AND "order" = 130;
		UPDATE faqs SET "order" = 9
			WHERE page_id = 'guides-tireguides-choose' AND locale = 'zh_cn' AND "order" = 140;
		UPDATE faqs SET updated_at = NOW() + INTERVAL '1 second'
			WHERE page_id = 'guides-tireguides-choose' AND locale = 'en' AND "order" = 150;
	`)
	if err != nil {
		t.Fatal(err)
	}
	executeTireRimWidthSelectionFAQMigrationSQL(t, transaction, "407_add_tire_rim_width_selection_editable_faq_content.up.sql")
	if err := transaction.QueryRow("SELECT count(*) FROM faqs WHERE page_id = 'guides-tireguides-choose'").Scan(&itemCount); err != nil {
		t.Fatal(err)
	}
	if itemCount != 18 {
		t.Fatalf("reapplying migration recreated an administrator-renamed question: %d items", itemCount)
	}
	executeTireRimWidthSelectionFAQMigrationSQL(t, transaction, "407_add_tire_rim_width_selection_editable_faq_content.down.sql")
	var preservedAdminItems, preservedOtherItems int
	if err := transaction.QueryRow("SELECT count(*) FROM faqs WHERE page_id = 'guides-tireguides-choose'").Scan(&preservedAdminItems); err != nil {
		t.Fatal(err)
	}
	if preservedAdminItems != 5 {
		t.Fatalf("safe rollback must retain all five edited/draft/image/reordered/touched entries, got %d", preservedAdminItems)
	}
	if err := transaction.QueryRow("SELECT count(*) FROM faqs WHERE page_id = 'other-guide'").Scan(&preservedOtherItems); err != nil {
		t.Fatal(err)
	}
	if preservedOtherItems != 1 {
		t.Fatal("rollback changed another guide's FAQ")
	}
	if err := transaction.QueryRow("SELECT count(*) FROM faq_pages").Scan(&pageCount); err != nil {
		t.Fatal(err)
	}
	if pageCount != 20 {
		t.Fatal("rollback removed route-owned FAQ page shells")
	}
}

func TestTireRimWidthSelectionFAQMigrationSeedsOnlyLocalesWithoutExistingContent(t *testing.T) {
	transaction := createIsolatedTireRimWidthSelectionFAQMigrationTransaction(t)
	_, err := transaction.Exec(`
		INSERT INTO faqs (page_id, locale, question, answer, status)
		VALUES ('guides-tireguides-choose', 'en', 'Existing admin question', '<p>Keep this answer</p>', 'draft');
	`)
	if err != nil {
		t.Fatal(err)
	}
	executeTireRimWidthSelectionFAQMigrationSQL(t, transaction, "407_add_tire_rim_width_selection_editable_faq_content.up.sql")
	var englishItems, chineseItems int
	if err := transaction.QueryRow("SELECT count(*) FROM faqs WHERE page_id = 'guides-tireguides-choose' AND locale = 'en'").Scan(&englishItems); err != nil {
		t.Fatal(err)
	}
	if err := transaction.QueryRow("SELECT count(*) FROM faqs WHERE page_id = 'guides-tireguides-choose' AND locale = 'zh_cn'").Scan(&chineseItems); err != nil {
		t.Fatal(err)
	}
	if englishItems != 1 || chineseItems != 9 {
		t.Fatalf("existing editorial content must be preserved and only empty locales seeded; en=%d zh_cn=%d", englishItems, chineseItems)
	}
}
