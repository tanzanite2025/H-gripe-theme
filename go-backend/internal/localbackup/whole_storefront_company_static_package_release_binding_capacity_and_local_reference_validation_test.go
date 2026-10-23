package localbackup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func prepareCompanyStaticReleaseFilesForWholeStorefrontRecoveryTest(t *testing.T, cfg *Configuration) string {
	t.Helper()
	cfg.ProductStaticAssetsDirectory = t.TempDir()
	company := filepath.Join(cfg.ProductStaticAssetsDirectory, "company")
	require.NoError(t, os.MkdirAll(filepath.Join(company, "products"), 0700))
	require.NoError(t, os.WriteFile(filepath.Join(company, companyStaticReleaseVersionFilename), []byte(cfg.ApplicationVersion), 0600))
	filename := filepath.Join(company, "products", "backup-required-photo.txt")
	require.NoError(t, os.WriteFile(filename, []byte("original-company-photo-bytes"), 0600))
	return filename
}

func TestWholeStorefrontCompanyStaticSnapshotBindsReleaseCopiesActualBytesAndResolvesLocalReferences(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	prepareCompanyStaticReleaseFilesForWholeStorefrontRecoveryTest(t, &cfg)
	cfg.PublicOrigins = []string{"https://storefront.example.invalid"}
	manifest := Manifest{Scope: "whole_application", ApplicationVersion: cfg.ApplicationVersion}
	snapshot := t.TempDir()
	require.NoError(t, cfg.copyWholeApplicationCompanyStaticSnapshot(t.Context(), snapshot, &manifest))
	require.Equal(t, cfg.ApplicationVersion, manifest.StorefrontCompanyStaticReleaseVersion)
	require.Len(t, manifest.Files, 2)
	require.NoError(t, cfg.validateWholeApplicationCompanyStaticPackage(manifest, snapshot))
	actual, err := cfg.companyStaticReleaseFileFingerprint(t.Context())
	require.NoError(t, err)
	require.Equal(t, fingerprintCompanyStaticManifestFiles(manifest.Files), actual)
	for _, reference := range []string{"/company/products/backup-required-photo.txt", "https://storefront.example.invalid/company/products/backup-required-photo.txt"} {
		require.Equal(t, companyStaticPackagePrefix+"products/backup-required-photo.txt", cfg.wholeBackupBusinessReferencePackagePath(reference))
	}
	require.Empty(t, cfg.wholeBackupBusinessReferencePackagePath("https://foreign.example.invalid/company/products/backup-required-photo.txt"))
	for _, path := range []string{"product-static/company/../escape", "product-static/company/", "product-static/company/C:/escape", "product-static/other/private", "product-static/company/.local-backup-old-task/private"} {
		require.False(t, isAllowedBackupFileForScope("whole_application", path), path)
	}
	for _, scenario := range []string{"missing_release_binding", "wrong_release", "missing_target_static_volume", "wrong_marker_bytes", "legacy_package_with_static_target"} {
		t.Run(scenario, func(t *testing.T) {
			changed, target := manifest, cfg
			switch scenario {
			case "missing_release_binding":
				changed.StorefrontCompanyStaticReleaseVersion = ""
			case "wrong_release":
				changed.StorefrontCompanyStaticReleaseVersion = "different-release"
			case "missing_target_static_volume":
				target.ProductStaticAssetsDirectory = ""
			case "wrong_marker_bytes":
				marker := filepath.Join(snapshot, filepath.FromSlash(companyStaticPackagePrefix+companyStaticReleaseVersionFilename))
				require.NoError(t, os.WriteFile(marker, []byte("different-release"), 0600))
				t.Cleanup(func() { require.NoError(t, os.WriteFile(marker, []byte(cfg.ApplicationVersion), 0600)) })
			case "legacy_package_with_static_target":
				changed = Manifest{Scope: "whole_application"}
			}
			require.Error(t, target.validateWholeApplicationCompanyStaticPackage(changed, snapshot))
		})
	}
}

func TestCompanyStaticExportRejectsMissingWrongReleaseCapacityAndSymbolicLinks(t *testing.T) {
	for _, scenario := range []string{"missing_marker", "wrong_release", "capacity", "symbolic_link"} {
		t.Run(scenario, func(t *testing.T) {
			cfg := createIndependentRecoveryTestConfiguration(t)
			filename := prepareCompanyStaticReleaseFilesForWholeStorefrontRecoveryTest(t, &cfg)
			marker := filepath.Join(cfg.ProductStaticAssetsDirectory, "company", companyStaticReleaseVersionFilename)
			switch scenario {
			case "missing_marker":
				require.NoError(t, os.Remove(marker))
			case "wrong_release":
				require.NoError(t, os.WriteFile(marker, []byte("different-release"), 0600))
			case "capacity":
				cfg.MaxExpandedBytes = 1
			case "symbolic_link":
				if err := os.Symlink(filename, filepath.Join(filepath.Dir(filename), "symbolic-link.txt")); err != nil {
					t.Skip("symbolic links unavailable on this host")
				}
			}
			err := cfg.copyWholeApplicationCompanyStaticSnapshot(t.Context(), t.TempDir(), &Manifest{})
			require.Error(t, err)
			require.False(t, strings.Contains(err.Error(), cfg.ProductStaticAssetsDirectory), "validation errors should explain the boundary without exposing host paths")
		})
	}
}
