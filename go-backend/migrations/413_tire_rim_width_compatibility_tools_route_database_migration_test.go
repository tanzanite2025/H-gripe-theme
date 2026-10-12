package migrations_test

import "testing"

func TestTireRimWidthCompatibilityToolsRouteMigrationPreservesEditorialFAQContentAndSupportsRollback(t *testing.T) {
	transaction := createIsolatedTireRimWidthSelectionFAQMigrationTransaction(t)
	_, err := transaction.ExecContext(t.Context(), `
		CREATE TEMP TABLE storefront_route_catalog_entries (
			route_key TEXT UNIQUE, source_key TEXT, source_type TEXT, locale TEXT,
			path TEXT, canonical_path TEXT, summary TEXT, manifest_version TEXT,
			updated_at TIMESTAMPTZ
		) ON COMMIT DROP;
		INSERT INTO storefront_route_catalog_entries (route_key, source_key, source_type, locale, path, canonical_path)
		VALUES
			('manifest:guides-tireguides-tire-rim-width-compatibility:en', 'guides-tireguides-tire-rim-width-compatibility', 'static', 'en', '/guides/tireguides/tire-rim-width-compatibility', '/guides/tireguides/tire-rim-width-compatibility'),
			('manifest:guides-tireguides-tire-rim-width-compatibility:zh_cn', 'guides-tireguides-tire-rim-width-compatibility', 'static', 'zh_cn', '/zh_cn/guides/tireguides/tire-rim-width-compatibility', '/zh_cn/guides/tireguides/tire-rim-width-compatibility');
		CREATE UNIQUE INDEX faq_pages_route_locale_test_unique ON faq_pages (route_path, locale)
			WHERE deleted_at IS NULL AND route_path <> '';
		INSERT INTO faq_pages (page_id, locale, route_path, title, subtitle, status, sort_order)
		VALUES ('guides-tireguides-tire-rim-width-compatibility', 'en', '/guides/tireguides/tire-rim-width-compatibility',
			'Administrator title', 'Administrator description', 'hidden', 42);
		INSERT INTO faqs (page_id, locale, question, answer, status, "order", answer_image_url)
		VALUES ('guides-tireguides-tire-rim-width-compatibility', 'en', 'Edited question', '<p>Edited answer</p>',
			'draft', 71, '/uploads/faq/edited-answer.webp');
	`)
	if err != nil {
		t.Fatal(err)
	}

	for run := 0; run < 2; run++ {
		executeTireRimWidthSelectionFAQMigrationSQL(t, transaction, "413_move_tire_rim_width_compatibility_faq_to_tools.up.sql")
	}

	var routePath, routeKey, title, subtitle, status string
	var sortOrder int
	if err := transaction.QueryRowContext(t.Context(), "SELECT route_path, route_key, title, subtitle, status, sort_order FROM faq_pages").
		Scan(&routePath, &routeKey, &title, &subtitle, &status, &sortOrder); err != nil {
		t.Fatal(err)
	}
	if routePath != "/tools/tire-rim-width-compatibility-engine" || routeKey != "guides-tireguides-tire-rim-width-compatibility" ||
		title != "Administrator title" || subtitle != "Administrator description" || status != "hidden" || sortOrder != 42 {
		t.Fatal("tool route migration changed the FAQ page identity or editorial metadata")
	}

	var question, answer, faqStatus, image string
	var order int
	if err := transaction.QueryRowContext(t.Context(), "SELECT page_id, question, answer, status, answer_image_url, \"order\" FROM faqs").
		Scan(&routeKey, &question, &answer, &faqStatus, &image, &order); err != nil {
		t.Fatal(err)
	}
	if routeKey != "guides-tireguides-tire-rim-width-compatibility" || question != "Edited question" || answer != "<p>Edited answer</p>" ||
		faqStatus != "draft" || image != "/uploads/faq/edited-answer.webp" || order != 71 {
		t.Fatal("tool route migration changed administrator FAQ content")
	}

	var renamedCatalogCount int
	if err := transaction.QueryRowContext(t.Context(), `SELECT count(*) FROM storefront_route_catalog_entries
		WHERE source_key = 'guides-tireguides-tire-rim-width-compatibility'
		AND path = CASE WHEN locale = 'en' THEN '' ELSE '/' || locale END || '/tools/tire-rim-width-compatibility-engine'
		AND canonical_path = path`).Scan(&renamedCatalogCount); err != nil {
		t.Fatal(err)
	}
	if renamedCatalogCount != 2 {
		t.Fatal("the sitemap catalog did not align the English and Chinese tool routes")
	}

	executeTireRimWidthSelectionFAQMigrationSQL(t, transaction, "413_move_tire_rim_width_compatibility_faq_to_tools.down.sql")
	if err := transaction.QueryRowContext(t.Context(), "SELECT route_path FROM faq_pages").Scan(&routePath); err != nil {
		t.Fatal(err)
	}
	if routePath != "/guides/tireguides/tire-rim-width-compatibility" {
		t.Fatal("rollback failed to restore the former guide route")
	}

	var restoredFAQCount int
	if err := transaction.QueryRowContext(t.Context(), "SELECT count(*) FROM faqs WHERE page_id = 'guides-tireguides-tire-rim-width-compatibility'").Scan(&restoredFAQCount); err != nil {
		t.Fatal(err)
	}
	if restoredFAQCount != 1 {
		t.Fatal("rollback did not preserve the existing FAQ answer")
	}

}
