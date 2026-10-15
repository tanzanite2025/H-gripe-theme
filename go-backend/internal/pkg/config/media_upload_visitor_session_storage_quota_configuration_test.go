package config

import "testing"

func TestValidateConfigRejectsNonpositiveMediaVisitorSessionStorageQuota(t *testing.T) {
	for _, quotaBytes := range []int64{0, -1} {
		cfg := validTestConfig()
		cfg.MediaUpload.VisitorSessionStorageQuotaBytes = quotaBytes
		if err := validateConfig(cfg); err == nil {
			t.Fatalf("accepted visitor session storage quota %d", quotaBytes)
		}
	}
}

func TestLoadMediaVisitorSessionStorageQuotaUsesDefaultAndEnvironmentOverride(t *testing.T) {
	t.Setenv("MEDIA_UPLOAD_VISITOR_SESSION_STORAGE_QUOTA_BYTES", "")
	t.Setenv("JWT_SECRET", "test-secret-at-least-thirty-two-characters")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MediaUpload.VisitorSessionStorageQuotaBytes != 20<<20 {
		t.Fatalf("default visitor quota = %d", cfg.MediaUpload.VisitorSessionStorageQuotaBytes)
	}
	t.Setenv("MEDIA_UPLOAD_VISITOR_SESSION_STORAGE_QUOTA_BYTES", "10485760")
	cfg, err = Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MediaUpload.VisitorSessionStorageQuotaBytes != 10<<20 {
		t.Fatalf("environment visitor quota = %d", cfg.MediaUpload.VisitorSessionStorageQuotaBytes)
	}
}
