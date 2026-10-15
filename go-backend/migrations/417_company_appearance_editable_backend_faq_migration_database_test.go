package migrations_test

import "testing"

func TestCompanyAppearanceBackendFAQMigrationSeedsEditableContentAndPreservesChangesDuringRollback(t *testing.T) {
	transaction := createIsolatedTireRimWidthSelectionFAQMigrationTransaction(t)
	_, err := transaction.ExecContext(t.Context(), `
		INSERT INTO faq_pages (page_id, route_path, locale, title, subtitle, status, sort_order)
		VALUES ('company-about-appearance', '/company/about/appearance', 'en',
			'Administrator title', 'Administrator subtitle', 'hidden', 42);
		INSERT INTO faqs (page_id, locale, question, answer)
		VALUES ('other-page', 'en', 'Keep another page', '<p>Other content</p>');
	`)
	if err != nil {
		t.Fatal(err)
	}
	const upMigration = "417_move_company_appearance_questions_from_frontend_to_editable_backend_faq.up.sql"
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
	if err := transaction.QueryRowContext(t.Context(), "SELECT count(*) FROM faqs WHERE page_id = 'company-about-appearance'").Scan(&questionCount); err != nil {
		t.Fatal(err)
	}
	if questionCount != 10 {
		t.Fatalf("expected five questions in each of two locales without duplicates, got %d", questionCount)
	}
	var title, subtitle, status string
	var sortOrder int
	if err := transaction.QueryRowContext(t.Context(), "SELECT title, subtitle, status, sort_order FROM faq_pages WHERE locale = 'en'").Scan(&title, &subtitle, &status, &sortOrder); err != nil {
		t.Fatal(err)
	}
	if title != "Administrator title" || subtitle != "Administrator subtitle" || status != "hidden" || sortOrder != 42 {
		t.Fatal("migration changed administrator page metadata")
	}
	_, err = transaction.ExecContext(t.Context(), `
		UPDATE faqs SET question = 'Administrator question', answer = '<p>Edited answer</p>',
			updated_at = NOW() + INTERVAL '1 second'
			WHERE page_id = 'company-about-appearance' AND locale = 'en' AND "order" = 10;
		UPDATE faqs SET status = 'draft'
			WHERE page_id = 'company-about-appearance' AND locale = 'zh_cn' AND "order" = 20;
		UPDATE faqs SET answer_image_url = '/uploads/faq/admin-image.webp'
			WHERE page_id = 'company-about-appearance' AND locale = 'zh_cn' AND "order" = 30;
		UPDATE faqs SET "order" = 9
			WHERE page_id = 'company-about-appearance' AND locale = 'zh_cn' AND "order" = 40;
		UPDATE faqs SET deleted_at = NOW()
			WHERE page_id = 'company-about-appearance' AND locale = 'en' AND "order" = 50;
	`)
	if err != nil {
		t.Fatal(err)
	}
	executeTireRimWidthSelectionFAQMigrationSQL(t, transaction, upMigration)
	executeTireRimWidthSelectionFAQMigrationSQL(t, transaction, "417_move_company_appearance_questions_from_frontend_to_editable_backend_faq.down.sql")
	if err := transaction.QueryRowContext(t.Context(), "SELECT count(*) FROM faqs WHERE page_id = 'company-about-appearance'").Scan(&questionCount); err != nil {
		t.Fatal(err)
	}
	if questionCount != 5 {
		t.Fatalf("rollback must retain the five edited, draft, illustrated, reordered or deleted entries, got %d", questionCount)
	}
	if err := transaction.QueryRowContext(t.Context(), "SELECT count(*) FROM faqs WHERE page_id = 'other-page'").Scan(&questionCount); err != nil {
		t.Fatal(err)
	}
	if questionCount != 1 {
		t.Fatal("rollback changed another page's FAQ")
	}
}

func TestCompanyAppearanceBackendFAQMigrationPreservesCustomRouteIdentityAndExistingOrDeletedQuestions(t *testing.T) {
	transaction := createIsolatedTireRimWidthSelectionFAQMigrationTransaction(t)
	_, err := transaction.ExecContext(t.Context(), `
		INSERT INTO faq_pages (page_id, route_path, locale, title)
		VALUES
			('custom-appearance-content', '/company/about/appearance', 'en', 'Custom page title'),
			('company-about-appearance', '/company/about/appearance', 'zh_cn', '中文自定义标题');
		INSERT INTO faqs (page_id, locale, question, answer, status, deleted_at)
		VALUES
			('custom-appearance-content', 'en', 'Existing draft question', '<p>Draft answer</p>', 'draft', NULL),
			('company-about-appearance', 'zh_cn', 'Deleted question', '<p>Keep deletion</p>', 'published', NOW());
	`)
	if err != nil {
		t.Fatal(err)
	}
	executeTireRimWidthSelectionFAQMigrationSQL(t, transaction, "417_move_company_appearance_questions_from_frontend_to_editable_backend_faq.up.sql")
	var questionCount, routePageCount int
	if err := transaction.QueryRowContext(t.Context(), "SELECT count(*) FROM faqs").Scan(&questionCount); err != nil {
		t.Fatal(err)
	}
	if questionCount != 2 {
		t.Fatal("migration supplemented an administrator-owned or deliberately deleted FAQ")
	}
	if err := transaction.QueryRowContext(t.Context(), "SELECT count(*) FROM faq_pages WHERE route_path = '/company/about/appearance' AND locale = 'en'").Scan(&routePageCount); err != nil {
		t.Fatal(err)
	}
	if routePageCount != 1 {
		t.Fatal("migration created a second page identity for the same route")
	}
}
