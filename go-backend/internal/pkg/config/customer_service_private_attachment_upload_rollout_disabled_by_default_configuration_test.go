package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCustomerServicePrivateAttachmentUploadRolloutDefaultsToDisabledAndRequiresExplicitEnvironmentOverride(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-at-least-thirty-two-characters")
	t.Setenv("MEDIA_UPLOAD_CUSTOMER_SERVICE_PRIVATE_ATTACHMENTS_ENABLED", "")
	cfg, err := Load()
	require.NoError(t, err)
	require.False(t, cfg.MediaUpload.CustomerServicePrivateAttachmentsEnabled)
	t.Setenv("MEDIA_UPLOAD_CUSTOMER_SERVICE_PRIVATE_ATTACHMENTS_ENABLED", "true")
	cfg, err = Load()
	require.NoError(t, err)
	require.True(t, cfg.MediaUpload.CustomerServicePrivateAttachmentsEnabled)
	t.Setenv("MEDIA_UPLOAD_CUSTOMER_SERVICE_PRIVATE_ATTACHMENTS_ENABLED", "false")
	cfg, err = Load()
	require.NoError(t, err)
	require.False(t, cfg.MediaUpload.CustomerServicePrivateAttachmentsEnabled)
}
