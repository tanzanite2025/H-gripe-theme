package localbackup

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocalBackupLoadsExistingDeploymentOriginsWithoutDNSOrAdditionalDomainConfiguration(t *testing.T) {
	t.Setenv("SERVER_BASE_URL", "http://media.deployment.invalid:8080/uploads")
	t.Setenv("STOREFRONT_BASE_URL", "https://storefront.deployment.invalid/")
	cfg := createIndependentRecoveryTestConfiguration(t)
	cfg.PublicOrigins = []string{"https://old-storefront.deployment.invalid/", "http://MEDIA.deployment.invalid:8080"}
	configurationPath := filepath.Join(t.TempDir(), "synthetic-backup-configuration-with-existing-deployment-urls.json")
	require.NoError(t, writeDurableJSON(configurationPath, cfg))
	loaded, err := LoadConfiguration(configurationPath)
	require.NoError(t, err)
	require.Equal(t, []string{"https://old-storefront.deployment.invalid", "http://media.deployment.invalid:8080", "https://storefront.deployment.invalid"}, loaded.PublicOrigins)
	for _, reference := range []string{"/uploads/catalog/photo.png", "http://media.deployment.invalid:8080/uploads/catalog/photo.png", "https://storefront.deployment.invalid/uploads/catalog/photo.png", "https://old-storefront.deployment.invalid/uploads/catalog/photo.png"} {
		root, key := loaded.resolveLocalMediaReference(reference)
		require.Equal(t, "uploads", root)
		require.Equal(t, "catalog/photo.png", key)
	}
	for _, reference := range []string{"https://third-party.deployment.invalid/uploads/catalog/photo.png", "https://media.deployment.invalid:8080/uploads/catalog/photo.png"} {
		root, key := loaded.resolveLocalMediaReference(reference)
		require.Empty(t, root)
		require.Empty(t, key)
	}
}

func TestLocalBackupRelativeMediaWorksWithoutDeploymentOriginsAndRejectsInvalidURLsWithoutExposingCredentials(t *testing.T) {
	t.Setenv("SERVER_BASE_URL", "")
	t.Setenv("STOREFRONT_BASE_URL", "")
	cfg := Configuration{}
	require.NoError(t, cfg.inheritMediaOriginsFromExistingServerAndStorefrontDeploymentURLs())
	root, key := cfg.resolveLocalMediaReference("/uploads/site-favicon/current.png")
	require.Equal(t, "uploads", root)
	require.Equal(t, "site-favicon/current.png", key)
	for _, value := range []string{"https://username:synthetic-private-value@deployment.invalid", "https://deployment.invalid?secret=synthetic-private-value", "file:///synthetic-private-value", "synthetic-private-value"} {
		t.Run(value[:4], func(t *testing.T) {
			t.Setenv("SERVER_BASE_URL", value)
			err := cfg.inheritMediaOriginsFromExistingServerAndStorefrontDeploymentURLs()
			require.ErrorContains(t, err, "SERVER_BASE_URL")
			require.NotContains(t, err.Error(), "synthetic-private-value")
		})
	}
}

func TestLocalBackupSiteIdentityURLsAndCustomerServiceRegistrationsUseActualSeparateStorageBindings(t *testing.T) {
	cfg := Configuration{PublicOrigins: []string{"https://media.deployment.invalid"}}
	for _, test := range []struct{ reference, root, key string }{
		{"/uploads/site-favicon/current.png", "uploads", "site-favicon/current.png"},
		{"https://media.deployment.invalid/uploads/site-favicon/current.png", "uploads", "site-favicon/current.png"},
		{"/uploads/site-logo/current.png", "site_logo_uploads", "site-logo/current.png"},
		{"/site-logo-uploads/site-logo/current.png", "site_logo_uploads", "site-logo/current.png"},
		{"/uploads-private/conversation/current.png", "uploads_private", "conversation/current.png"},
	} {
		t.Run(test.reference, func(t *testing.T) {
			root, key := cfg.resolveLocalMediaReference(test.reference)
			require.Equal(t, test.root, root)
			require.Equal(t, test.key, key)
			for _, useURLFallback := range []bool{false, true} {
				storedKey, visibility := test.key, "public"
				if test.root == "uploads_private" {
					visibility = "private"
				}
				if useURLFallback {
					storedKey = ""
				}
				values := map[string]json.RawMessage{}
				for name, value := range map[string]string{"storage_key": storedKey, "visibility": visibility, "url": test.reference} {
					values[name], _ = json.Marshal(value)
				}
				root, key = cfg.customerServiceMediaAssetLocalIdentity(&moduleRecoveryDependencyRow{Values: values})
				require.Equal(t, test.root, root)
				require.Equal(t, test.key, key)
			}
		})
	}
}
