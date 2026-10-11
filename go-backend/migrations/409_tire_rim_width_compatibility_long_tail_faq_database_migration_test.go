package migrations_test

import (
	"strings"
	"testing"

	"commerce-platform/internal/domain/tirerim"
)

func TestTireRimWidthCompatibilityLongTailFAQMigrationPreservesExistingContentAndAdministratorEdits(t *testing.T) {
	transaction := createIsolatedTireRimWidthSelectionFAQMigrationTransaction(t)
	_, err := transaction.Exec(`
		INSERT INTO faq_pages (page_id, locale, route_path, title, status, sort_order)
		VALUES
			('guides-tireguides-tire-rim-width-compatibility', 'en', '/guides/tireguides/tire-rim-width-compatibility', 'Administrator title', 'hidden', 42),
			('guides-tireguides-tire-rim-width-compatibility', 'zh_cn', '/guides/tireguides/tire-rim-width-compatibility', '中文标题', 'active', 250);
		INSERT INTO faqs (page_id, locale, question, answer, status, "order")
		VALUES
			('guides-tireguides-tire-rim-width-compatibility', 'en', 'Existing 105% question', '<p>Existing calculation explanation</p>', 'draft', 150),
			('guides-tireguides-tire-rim-width-compatibility', 'zh_cn', '现有 105% 问答', '<p>原计算说明</p>', 'published', 150),
			('other-guide', 'en', 'Other guide question', '<p>Keep another page</p>', 'published', 10);
	`)
	if err != nil {
		t.Fatal(err)
	}
	for run := 0; run < 2; run++ {
		executeTireRimWidthSelectionFAQMigrationSQL(t, transaction, "409_add_tire_rim_width_compatibility_long_tail_faq_content.up.sql")
	}
	var count int
	if err := transaction.QueryRow("SELECT count(*) FROM faqs WHERE page_id = 'guides-tireguides-tire-rim-width-compatibility'").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 8 {
		t.Fatalf("expected the two existing FAQs and six additions without duplicates, got %d", count)
	}
	var title, pageStatus string
	var pageOrder int
	if err := transaction.QueryRow("SELECT title, status, sort_order FROM faq_pages WHERE locale = 'en'").Scan(&title, &pageStatus, &pageOrder); err != nil {
		t.Fatal(err)
	}
	if title != "Administrator title" || pageStatus != "hidden" || pageOrder != 42 {
		t.Fatal("FAQ addition changed administrator page metadata or publication")
	}
	var recommendationAnswer string
	if err := transaction.QueryRow(`SELECT answer FROM faqs WHERE locale = 'en' AND "order" = 90`).Scan(&recommendationAnswer); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(recommendationAnswer, "21 mm") {
		t.Fatal("the seeded recommendation does not state the checked current result")
	}
	for _, rimSystem := range []tirerim.RimSystem{tirerim.RimSystemHooked, tirerim.RimSystemHookless} {
		recommendation, err := tirerim.ResolveTireRimWidthReference(28, rimSystem)
		if err != nil || recommendation.BestFitment == nil || recommendation.BestFitment.RimInnerWidthMM != 21 {
			t.Fatalf("28C FAQ needs review because the existing %s calculation changed: %#v, %v", rimSystem, recommendation, err)
		}
	}
	_, err = transaction.Exec(`
		UPDATE faqs SET question = 'Administrator renamed question', answer = '<p>Edited answer</p>',
			answer_image_url = '/uploads/faq/edited-answer.webp'
			WHERE locale = 'en' AND "order" = 80;
		UPDATE faqs SET status = 'draft' WHERE locale = 'zh_cn' AND "order" = 90;
		UPDATE faqs SET "order" = 8 WHERE locale = 'zh_cn' AND "order" = 100;
		UPDATE faqs SET updated_at = NOW() + INTERVAL '1 second' WHERE locale = 'en' AND "order" = 90;
	`)
	if err != nil {
		t.Fatal(err)
	}
	executeTireRimWidthSelectionFAQMigrationSQL(t, transaction, "409_add_tire_rim_width_compatibility_long_tail_faq_content.up.sql")
	if err := transaction.QueryRow("SELECT count(*) FROM faqs WHERE page_id = 'guides-tireguides-tire-rim-width-compatibility'").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 8 {
		t.Fatal("reapplying the migration duplicated an edited or reordered answer")
	}
	executeTireRimWidthSelectionFAQMigrationSQL(t, transaction, "409_add_tire_rim_width_compatibility_long_tail_faq_content.down.sql")
	if err := transaction.QueryRow("SELECT count(*) FROM faqs WHERE page_id = 'guides-tireguides-tire-rim-width-compatibility'").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 6 {
		t.Fatalf("rollback must preserve the two original FAQs and four administrator edits, got %d", count)
	}
	var question, answer, image string
	if err := transaction.QueryRow(`SELECT question, answer, answer_image_url FROM faqs WHERE locale = 'en' AND "order" = 80`).Scan(&question, &answer, &image); err != nil {
		t.Fatal(err)
	}
	if question != "Administrator renamed question" || answer != "<p>Edited answer</p>" || image != "/uploads/faq/edited-answer.webp" {
		t.Fatal("migration or rollback changed the administrator's edited answer")
	}
	if err := transaction.QueryRow(`SELECT count(*) FROM faqs WHERE "order" = 150 AND ((locale = 'en' AND question = 'Existing 105% question' AND answer = '<p>Existing calculation explanation</p>' AND status = 'draft') OR (locale = 'zh_cn' AND question = '现有 105% 问答' AND answer = '<p>原计算说明</p>' AND status = 'published'))`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatal("the existing 105% explanation was modified")
	}
	if err := transaction.QueryRow("SELECT count(*) FROM faqs WHERE page_id = 'other-guide'").Scan(&count); err != nil || count != 1 {
		t.Fatal("another guide's content was changed")
	}
}
