package migrations_test

import "testing"

func TestCarbonRimHolePatternAndDrillingFAQMigrationSeedsEmptyLocalesWithoutDuplicatesAndPreservesAdministratorEditsOnRollback(t *testing.T) {
	transaction := createIsolatedTireRimWidthSelectionFAQMigrationTransaction(t)
	_, err := transaction.ExecContext(t.Context(), `
		INSERT INTO faq_pages (page_id, route_path, locale, title, subtitle, status, sort_order)
		VALUES
			('company-about-hole-patterns', '/company/about/hole-patterns',
			 'en', 'Administrator title', 'Administrator subtitle', 'hidden', 42),
			('company-about-hole-patterns', '/company/about/hole-patterns',
			 'zh_cn', 'Hole Patterns FAQs', 'Rim drilling layouts and spoke hole customization', 'active', 250);
		INSERT INTO faqs (page_id, locale, question, answer)
		VALUES ('other-page', 'en', 'Keep another page', '<p>Other content</p>');
	`)
	if err != nil {
		t.Fatal(err)
	}
	const upMigration = "424_add_backend_editable_carbon_rim_hole_pattern_and_drilling_faq_content.up.sql"
	for run := 0; run < 2; run++ {
		executeTireRimWidthSelectionFAQMigrationSQL(t, transaction, upMigration)
	}
	var pageCount, questionCount int
	if err := transaction.QueryRowContext(t.Context(), "SELECT count(*) FROM faq_pages").Scan(&pageCount); err != nil {
		t.Fatal(err)
	}
	if pageCount != 20 {
		t.Fatalf("expected twenty locale page shells, got %d", pageCount)
	}
	if err := transaction.QueryRowContext(t.Context(), "SELECT count(*) FROM faqs WHERE page_id = 'company-about-hole-patterns'").Scan(&questionCount); err != nil {
		t.Fatal(err)
	}
	if questionCount != 12 {
		t.Fatalf("expected six questions per published locale without duplicates, got %d", questionCount)
	}
	var title, subtitle, status string
	var sortOrder int
	if err := transaction.QueryRowContext(t.Context(), "SELECT title, subtitle, status, sort_order FROM faq_pages WHERE locale = 'en'").Scan(&title, &subtitle, &status, &sortOrder); err != nil {
		t.Fatal(err)
	}
	if title != "Administrator title" || subtitle != "Administrator subtitle" || status != "hidden" || sortOrder != 42 {
		t.Fatal("migration changed administrator page metadata or visibility")
	}
	if err := transaction.QueryRowContext(t.Context(), "SELECT title, subtitle FROM faq_pages WHERE locale = 'zh_cn'").Scan(&title, &subtitle); err != nil {
		t.Fatal(err)
	}
	if title != "孔位与打孔角度常见问题" || subtitle != "了解孔数、2:1 / G3 编法、左右偏孔及定制所需资料" {
		t.Fatal("generated route metadata was not localized for the Chinese FAQ")
	}
	_, err = transaction.ExecContext(t.Context(), `
		UPDATE faqs SET question = 'Administrator question', answer = '<p>Edited answer</p>',
			updated_at = NOW() + INTERVAL '1 second'
			WHERE page_id = 'company-about-hole-patterns' AND locale = 'en' AND "order" = 10;
		UPDATE faqs SET status = 'draft'
			WHERE page_id = 'company-about-hole-patterns' AND locale = 'zh_cn' AND "order" = 20;
		UPDATE faqs SET answer_image_url = '/uploads/faq/admin-image.webp'
			WHERE page_id = 'company-about-hole-patterns' AND locale = 'zh_cn' AND "order" = 30;
		UPDATE faqs SET "order" = 9
			WHERE page_id = 'company-about-hole-patterns' AND locale = 'zh_cn' AND "order" = 40;
		UPDATE faqs SET deleted_at = NOW()
			WHERE page_id = 'company-about-hole-patterns' AND locale = 'en' AND "order" = 50;
		UPDATE faqs SET updated_at = NOW() + INTERVAL '1 second'
			WHERE page_id = 'company-about-hole-patterns' AND locale = 'en' AND "order" = 60;
	`)
	if err != nil {
		t.Fatal(err)
	}
	executeTireRimWidthSelectionFAQMigrationSQL(t, transaction, upMigration)
	executeTireRimWidthSelectionFAQMigrationSQL(t, transaction, "424_add_backend_editable_carbon_rim_hole_pattern_and_drilling_faq_content.down.sql")
	if err := transaction.QueryRowContext(t.Context(), "SELECT count(*) FROM faqs WHERE page_id = 'company-about-hole-patterns'").Scan(&questionCount); err != nil {
		t.Fatal(err)
	}
	if questionCount != 6 {
		t.Fatalf("rollback must retain edited, draft, illustrated, reordered, deleted and touched answers, got %d", questionCount)
	}
	if err := transaction.QueryRowContext(t.Context(), "SELECT count(*) FROM faqs WHERE page_id = 'other-page'").Scan(&questionCount); err != nil {
		t.Fatal(err)
	}
	if questionCount != 1 {
		t.Fatal("rollback changed another page's FAQ")
	}
}

func TestCarbonRimHolePatternAndDrillingFAQMigrationPreservesCustomPageIdentitiesDraftsAndDeletedContent(t *testing.T) {
	transaction := createIsolatedTireRimWidthSelectionFAQMigrationTransaction(t)
	_, err := transaction.ExecContext(t.Context(), `
		INSERT INTO faq_pages (page_id, route_path, locale, title, deleted_at)
		VALUES
			('custom-hole-pattern-content', '/company/about/hole-patterns', 'en', 'Custom page title', NULL),
			('company-about-hole-patterns', '/company/about/hole-patterns', 'zh_cn', 'Keep deleted page', NOW());
	`)
	if err != nil {
		t.Fatal(err)
	}
	const upMigration = "424_add_backend_editable_carbon_rim_hole_pattern_and_drilling_faq_content.up.sql"
	executeTireRimWidthSelectionFAQMigrationSQL(t, transaction, upMigration)
	var customPageQuestions, deletedPageQuestions, routePageCount int
	if err := transaction.QueryRowContext(t.Context(), "SELECT count(*) FROM faqs WHERE page_id = 'custom-hole-pattern-content'").Scan(&customPageQuestions); err != nil {
		t.Fatal(err)
	}
	if err := transaction.QueryRowContext(t.Context(), "SELECT count(*) FROM faqs WHERE locale = 'zh_cn'").Scan(&deletedPageQuestions); err != nil {
		t.Fatal(err)
	}
	if err := transaction.QueryRowContext(t.Context(), "SELECT count(*) FROM faq_pages WHERE route_path = '/company/about/hole-patterns' AND locale = 'en'").Scan(&routePageCount); err != nil {
		t.Fatal(err)
	}
	if customPageQuestions != 6 || deletedPageQuestions != 0 || routePageCount != 1 {
		t.Fatal("migration ignored the custom route identity or resurrected a deleted page")
	}
	_, err = transaction.ExecContext(t.Context(), `
		DELETE FROM faqs WHERE page_id = 'custom-hole-pattern-content';
		UPDATE faq_pages SET deleted_at = NULL WHERE locale = 'zh_cn';
		INSERT INTO faqs (page_id, locale, question, answer, status, deleted_at)
		VALUES
			('custom-hole-pattern-content', 'en', 'Existing draft question', '<p>Draft answer</p>', 'draft', NULL),
			('company-about-hole-patterns', 'zh_cn', 'Deleted question', '<p>Keep deletion</p>', 'published', NOW());
	`)
	if err != nil {
		t.Fatal(err)
	}
	executeTireRimWidthSelectionFAQMigrationSQL(t, transaction, upMigration)
	var questionCount int
	if err := transaction.QueryRowContext(t.Context(), "SELECT count(*) FROM faqs").Scan(&questionCount); err != nil {
		t.Fatal(err)
	}
	if questionCount != 2 {
		t.Fatal("migration supplemented administrator-owned or deliberately deleted FAQ content")
	}
}
