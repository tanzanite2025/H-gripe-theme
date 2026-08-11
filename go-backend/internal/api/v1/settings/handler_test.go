package settings

import (
	"testing"

	seodomain "commerce-platform/internal/domain/seo"
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/service"
)

func TestFilterPublicSettingsManagedByDomain(t *testing.T) {
	input := []setting.Setting{
		{Key: "site_name", Group: "site"},
		{Key: "tz_loyalty_checkin_base_points", Group: "loyalty"},
		{Key: "tz_redeem_exchange_rate", Group: "redeem"},
		{Key: "social_instagram", Group: "social"},
		{Key: seodomain.HomeKeys.MetaTitle, Group: "seo"},
		{Key: "google_analytics", Group: "analytics"},
	}

	actual := service.FilterDomainManagedSettings(input)

	if len(actual) != 2 {
		t.Fatalf("expected 2 public settings, got %d", len(actual))
	}
	if actual[0].Key != "site_name" || actual[1].Key != "social_instagram" {
		t.Fatalf("unexpected filtered settings: %#v", actual)
	}
}

func TestDomainManagedPublicSettingKeysAreBlocked(t *testing.T) {
	for _, key := range []string{
		"tz_loyalty_checkin_base_points",
		"tz_redeem_exchange_rate",
		seodomain.HomeKeys.MetaTitle,
		seodomain.HomeKeys.MetaDescription,
		"google_analytics",
	} {
		if !service.IsDomainManagedSettingKey(key) {
			t.Fatalf("expected %s to be domain managed", key)
		}
	}
	if service.IsDomainManagedSettingKey("site_name") {
		t.Fatal("site_name should not be domain managed")
	}
	if service.IsDomainManagedSettingKey("meta_title") {
		t.Fatal("legacy meta_title should not be domain managed")
	}
}
