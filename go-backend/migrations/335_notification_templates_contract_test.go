package migrations_test

import (
	"strings"
	"testing"
)

func TestNotificationTemplatesMigrationCreatesCurrentAndVersionedRows(t *testing.T) {
	sql := strings.ToLower(readMigrationFile(t, "335_notification_templates.up.sql"))
	for _, fragment := range []string{
		"create table if not exists email_templates",
		"subject_template varchar(255)",
		"allowed_variables jsonb",
		"required_variables jsonb",
		"constraint uq_email_templates_code_locale unique (code, locale)",
		"create table if not exists email_template_versions",
		"references email_templates(id) on delete cascade",
		"unique (template_id, version)",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration missing %q", fragment)
		}
	}
}

func TestNotificationTemplatesMigrationRollbackDropsVersionBeforeCurrentTable(t *testing.T) {
	sql := strings.ToLower(readMigrationFile(t, "335_notification_templates.down.sql"))
	if strings.Index(sql, "drop table if exists email_template_versions") > strings.Index(sql, "drop table if exists email_templates") {
		t.Fatal("rollback must drop template versions before current templates")
	}
}
