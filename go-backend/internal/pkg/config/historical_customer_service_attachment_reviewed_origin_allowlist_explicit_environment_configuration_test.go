package config

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHistoricalCustomerServiceAttachmentReviewedOriginsRequireExplicitEnvironmentConfiguration(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-at-least-thirty-two-characters")
	t.Setenv("MEDIA_REVIEWED_HISTORICAL_CUSTOMER_SERVICE_ATTACHMENT_ORIGINS", "")
	cfg, err := Load()
	require.NoError(t, err)
	require.Empty(t, cfg.MediaUpload.ReviewedHistoricalCustomerServiceAttachmentOrigins)
	t.Setenv("MEDIA_REVIEWED_HISTORICAL_CUSTOMER_SERVICE_ATTACHMENT_ORIGINS", "http://localhost:9199,https://old-shop.example.test")
	cfg, err = Load()
	require.NoError(t, err)
	require.Equal(t, "http://localhost:9199,https://old-shop.example.test", cfg.MediaUpload.ReviewedHistoricalCustomerServiceAttachmentOrigins)
}
