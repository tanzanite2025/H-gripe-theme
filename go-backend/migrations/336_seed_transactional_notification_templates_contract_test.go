package migrations_test

import (
	"strings"
	"testing"
)

func TestSeedTransactionalNotificationTemplatesCoversOrderAndAfterSalesCodes(t *testing.T) {
	sql := strings.ToLower(readMigrationFile(t, "336_seed_transactional_notification_templates.up.sql"))
	for _, code := range []string{
		"order_confirmation",
		"order_payment_expired",
		"order_cancelled",
		"order_shipping_notification",
		"order_delivered",
		"order_completed",
		"order_refunded",
		"after_sales_requested",
		"after_sales_approved",
		"after_sales_awaiting_return",
		"after_sales_return_in_transit",
		"after_sales_received",
		"after_sales_resolving",
		"after_sales_completed",
		"after_sales_rejected",
	} {
		if !strings.Contains(sql, "'"+code+"'") {
			t.Fatalf("template seed missing %q", code)
		}
	}
	for _, fragment := range []string{
		"on conflict (code, locale) do nothing",
		"insert into email_template_versions",
		"on conflict (template_id, version) do nothing",
		"allowed_variables",
		"required_variables",
		"body_html",
		"body_text",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("template seed missing %q", fragment)
		}
	}
}

func TestSeedTransactionalNotificationTemplatesRollbackPreservesEditedVersions(t *testing.T) {
	sql := strings.ToLower(readMigrationFile(t, "336_seed_transactional_notification_templates.down.sql"))
	if !strings.Contains(sql, "version = 1") {
		t.Fatal("rollback must not delete edited template versions")
	}
	if !strings.Contains(sql, "after_sales_completed") {
		t.Fatal("rollback missing after-sales seed codes")
	}
}
