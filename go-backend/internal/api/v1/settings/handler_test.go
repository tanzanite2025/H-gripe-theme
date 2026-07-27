package settings

import (
	"testing"

	"tanzanite/internal/domain/setting"
)

func TestFilterPublicSettingsManagedByDomain(t *testing.T) {
	input := []setting.Setting{
		{Key: "site_name", Group: "site"},
		{Key: "tz_loyalty_checkin_base_points", Group: "loyalty"},
		{Key: "tz_redeem_exchange_rate", Group: "redeem"},
		{Key: "social_instagram", Group: "social"},
	}

	actual := filterPublicSettingsManagedByDomain(input)

	if len(actual) != 2 {
		t.Fatalf("expected 2 public settings, got %d", len(actual))
	}
	if actual[0].Key != "site_name" || actual[1].Key != "social_instagram" {
		t.Fatalf("unexpected filtered settings: %#v", actual)
	}
}

func TestDomainManagedPublicSettingKeysAreBlocked(t *testing.T) {
	for _, key := range []string{"tz_loyalty_checkin_base_points", "tz_redeem_exchange_rate"} {
		if !isPublicSettingKeyManagedByDomain(key) {
			t.Fatalf("expected %s to be domain managed", key)
		}
	}
	if isPublicSettingKeyManagedByDomain("site_name") {
		t.Fatal("site_name should not be domain managed")
	}
}
