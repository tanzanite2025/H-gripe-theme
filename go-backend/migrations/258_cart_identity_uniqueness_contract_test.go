package migrations_test

import (
	"strings"
	"testing"
)

func TestCartIdentityUniquenessUpMigrationContract(t *testing.T) {
	sql := strings.ToLower(readMigrationFile(t, "258_cart_identity_uniqueness.up.sql"))
	for _, fragment := range []string{
		"update carts",
		"join lateral",
		"set deleted_at = current_timestamp",
		"create unique index if not exists uq_carts_active_user_id",
		"on carts(user_id)",
		"create unique index if not exists uq_carts_active_session_id",
		"on carts(session_id)",
		"where user_id is null",
		"and deleted_at is null",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("cart identity migration is missing contract fragment %q", fragment)
		}
	}
}

func TestCartIdentityUniquenessDownMigrationContract(t *testing.T) {
	sql := strings.ToLower(readMigrationFile(t, "258_cart_identity_uniqueness.down.sql"))
	for _, fragment := range []string{
		"drop index if exists uq_carts_active_session_id",
		"drop index if exists uq_carts_active_user_id",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("cart identity rollback is missing contract fragment %q", fragment)
		}
	}
}
