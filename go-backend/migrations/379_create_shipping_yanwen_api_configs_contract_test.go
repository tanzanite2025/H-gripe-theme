package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestCreateShippingYanwenAPIConfigsMigrationDefinesEncryptedEnvironmentCredentials(t *testing.T) {
	contents, err := os.ReadFile("379_create_shipping_yanwen_api_configs.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(contents))
	for _, fragment := range []string{
		"create table if not exists shipping_yanwen_api_configs",
		"environment varchar(16) not null default 'fat' unique",
		"endpoint varchar(500) not null",
		"user_id_encrypted text not null default ''",
		"api_token_encrypted text not null default ''",
		"enabled boolean not null default false",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration is missing required fragment %q", fragment)
		}
	}
}
