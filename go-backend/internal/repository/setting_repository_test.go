package repository

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newMockSettingRepository(t *testing.T) (*SettingRepository, sqlmock.Sqlmock, func()) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sql mock: %v", err)
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		_ = sqlDB.Close()
		t.Fatalf("failed to create gorm db: %v", err)
	}

	return NewSettingRepository(db), mock, func() {
		_ = sqlDB.Close()
	}
}

func TestGetByGroupQuotesReservedGroupColumnForPostgres(t *testing.T) {
	repo, mock, cleanup := newMockSettingRepository(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "settings" WHERE ("settings"."group" = $1 AND "settings"."locale" = $2) AND "settings"."deleted_at" IS NULL`)).
		WithArgs("site", "en").
		WillReturnRows(sqlmock.NewRows([]string{"id", "key", "value", "type", "locale", "group", "is_public", "description", "created_at", "updated_at", "deleted_at"}))

	settings, err := repo.GetByGroup("site", "en")
	if err != nil {
		t.Fatalf("GetByGroup returned error: %v", err)
	}
	if len(settings) != 0 {
		t.Fatalf("expected no settings, got %d", len(settings))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestGetGroupsQuotesReservedGroupColumnForPostgres(t *testing.T) {
	repo, mock, cleanup := newMockSettingRepository(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT DISTINCT "group" FROM "settings" WHERE "settings"."deleted_at" IS NULL`)).
		WillReturnRows(sqlmock.NewRows([]string{"group"}).AddRow("site").AddRow("seo"))

	groups, err := repo.GetGroups()
	if err != nil {
		t.Fatalf("GetGroups returned error: %v", err)
	}
	if len(groups) != 2 || groups[0] != "site" || groups[1] != "seo" {
		t.Fatalf("unexpected groups: %#v", groups)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}
