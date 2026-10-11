package migrations_test

import "testing"

func TestTireRimWidthSelectionFAQRouteAndIdentityRenamePreservesEditorialContentAndSupportsRollback(t *testing.T) {
	transaction := createIsolatedTireRimWidthSelectionFAQMigrationTransaction(t)
	_, err := transaction.Exec(`
		CREATE TEMP TABLE storefront_route_catalog_entries (
			route_key TEXT UNIQUE, source_key TEXT, source_type TEXT, locale TEXT,
			path TEXT, canonical_path TEXT, summary TEXT, manifest_version TEXT,
			updated_at TIMESTAMPTZ
		) ON COMMIT DROP;
		INSERT INTO storefront_route_catalog_entries (route_key, source_key, source_type, locale, path, canonical_path)
		VALUES
			('manifest:guides-tireguides-choose:en', 'guides-tireguides-choose', 'static', 'en', '/guides/tireguides/choose', '/guides/tireguides/choose'),
			('manifest:guides-tireguides-choose:zh_cn', 'guides-tireguides-choose', 'static', 'zh_cn', '/zh_cn/guides/tireguides/choose', '/zh_cn/guides/tireguides/choose');
		CREATE UNIQUE INDEX faq_pages_route_locale_test_unique ON faq_pages (route_path, locale)
			WHERE deleted_at IS NULL AND route_path <> '';
		INSERT INTO faq_pages (page_id, locale, route_path, title, subtitle, status, sort_order)
		VALUES ('guides-tireguides-choose', 'en', '/guides/tireguides/choose',
			'Administrator title', 'Administrator description', 'hidden', 42);
		INSERT INTO faqs (page_id, locale, question, answer, status, "order", answer_image_url)
		VALUES ('guides-tireguides-choose', 'en', 'Edited question', '<p>Edited answer</p>',
			'draft', 71, '/uploads/faq/edited-answer.webp');
	`)
	if err != nil {
		t.Fatal(err)
	}
	for run := 0; run < 2; run++ {
		executeTireRimWidthSelectionFAQMigrationSQL(t, transaction, "408_move_tire_rim_width_selection_faq_to_descriptive_guide_route.up.sql")
	}
	var routePath, pageID, routeKey, title, subtitle, pageStatus string
	var sortOrder int
	if err := transaction.QueryRow("SELECT route_path, page_id, route_key, title, subtitle, status, sort_order FROM faq_pages").
		Scan(&routePath, &pageID, &routeKey, &title, &subtitle, &pageStatus, &sortOrder); err != nil {
		t.Fatal(err)
	}
	if routePath != "/guides/tireguides/tire-rim-width-compatibility" ||
		pageID != "guides-tireguides-tire-rim-width-compatibility" || routeKey != pageID ||
		title != "Administrator title" || subtitle != "Administrator description" || pageStatus != "hidden" || sortOrder != 42 {
		t.Fatal("route migration failed to align the new identity or changed editorial metadata")
	}
	var questionPageID string
	if err := transaction.QueryRow("SELECT page_id FROM faqs").Scan(&questionPageID); err != nil {
		t.Fatal(err)
	}
	if questionPageID != pageID {
		t.Fatal("FAQ questions were not moved to the new page identity")
	}
	var oldIdentityCount int
	if err := transaction.QueryRow("SELECT count(*) FROM faq_pages WHERE page_id = 'guides-tireguides-choose' OR route_path = '/guides/tireguides/choose'").Scan(&oldIdentityCount); err != nil {
		t.Fatal(err)
	}
	if oldIdentityCount != 0 {
		t.Fatal("route migration retained an old route or compatibility page")
	}
	var renamedCatalogCount int
	if err := transaction.QueryRow(`SELECT count(*) FROM storefront_route_catalog_entries
		WHERE route_key = 'manifest:guides-tireguides-tire-rim-width-compatibility:' || locale
		AND path = CASE WHEN locale = 'en' THEN '' ELSE '/' || locale END || '/guides/tireguides/tire-rim-width-compatibility'
		AND canonical_path = path`).Scan(&renamedCatalogCount); err != nil {
		t.Fatal(err)
	}
	if renamedCatalogCount != 2 {
		t.Fatal("the sitemap catalog did not align the English and Chinese route identities")
	}
	executeTireRimWidthSelectionFAQMigrationSQL(t, transaction, "408_move_tire_rim_width_selection_faq_to_descriptive_guide_route.down.sql")
	if err := transaction.QueryRow("SELECT route_path FROM faq_pages WHERE page_id = 'guides-tireguides-choose'").Scan(&routePath); err != nil {
		t.Fatal(err)
	}
	if routePath != "/guides/tireguides/choose" {
		t.Fatal("rollback failed to restore the original route and page identity")
	}
	var question, answer, faqStatus, image string
	var order int
	if err := transaction.QueryRow("SELECT page_id, question, answer, status, answer_image_url, \"order\" FROM faqs").
		Scan(&questionPageID, &question, &answer, &faqStatus, &image, &order); err != nil {
		t.Fatal(err)
	}
	if questionPageID != "guides-tireguides-choose" || question != "Edited question" || answer != "<p>Edited answer</p>" || faqStatus != "draft" ||
		image != "/uploads/faq/edited-answer.webp" || order != 71 {
		t.Fatal("route migration or rollback changed administrator FAQ content")
	}
}
