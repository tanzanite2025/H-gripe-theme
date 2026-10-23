package localbackup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func prepareWholeBackupSiteIdentityFilesAndVerifyWrongRootDecoysCannotResolveMissingObjects(t *testing.T, cfg Configuration) string {
	t.Helper()
	db, err := cfg.connectDatabase(t.Context(), cfg.DatabaseName)
	require.NoError(t, err)
	defer db.Close()
	_, err = db.ExecContext(t.Context(), `
CREATE TABLE site_logo_assets(id bigint PRIMARY KEY,storage_key text,sha256 text,size_bytes bigint);
CREATE TABLE site_favicon_assets(id bigint PRIMARY KEY,storage_key text,sha256 text,size_bytes bigint);
INSERT INTO site_logo_assets VALUES(51,'site-logo/current.txt','',18);
INSERT INTO site_favicon_assets VALUES(52,'site-favicon/current.txt','',21);
INSERT INTO media_assets VALUES(53,'site-logo/current.txt','public','/uploads/site-logo/current.txt'),(54,'site-favicon/current.txt','public','/uploads/site-favicon/current.txt'),(55,'','public','https://synthetic-backup-media.invalid/uploads/site-favicon/current.txt');
INSERT INTO media_asset_derivatives VALUES(56,53,'site-logo/derived.txt'),(57,54,'site-favicon/derived.txt');`)
	require.NoError(t, err)
	for root, objects := range map[string]map[string]string{
		"uploads":           {"site-favicon/current.txt": "original-site-favicon", "site-favicon/derived.txt": "actual-shared-favicon-derivative", "site-logo/current.txt": "wrong-root-logo-decoy", "site-logo/derived.txt": "wrong-root-logo-derivative-decoy"},
		"site_logo_uploads": {"site-logo/current.txt": "original-site-logo", "site-logo/derived.txt": "actual-dedicated-logo-derivative", "site-favicon/current.txt": "wrong-root-favicon-decoy", "site-favicon/derived.txt": "wrong-root-favicon-derivative-decoy"},
	} {
		for key, content := range objects {
			filename := filepath.Join(cfg.StorageRoots[root], filepath.FromSlash(key))
			require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0700))
			require.NoError(t, os.WriteFile(filename, []byte(content), 0600))
		}
	}
	manifest := Manifest{Tables: map[string]int64{"media_assets": 4, "media_asset_derivatives": 2, "site_logo_assets": 1, "site_favicon_assets": 1}}
	require.NoError(t, cfg.copyMediaSnapshot(t.Context(), t.TempDir(), &manifest))
	require.NoError(t, cfg.findMediaReferenceGaps(t.Context(), db, &manifest))
	require.Empty(t, manifest.Gaps)
	for _, test := range []struct{ path, gap string }{
		{"media/uploads/site-favicon/current.txt", "missing_registered_business_attachment:site_favicon_assets:52"},
		{"media/site_logo_uploads/site-logo/current.txt", "missing_registered_business_attachment:site_logo_assets:51"},
		{"media/uploads/site-favicon/derived.txt", "missing_derivative:media_asset_derivatives:57"},
		{"media/site_logo_uploads/site-logo/derived.txt", "missing_derivative:media_asset_derivatives:56"},
	} {
		t.Run("wrong_root_decoy_cannot_replace_"+test.path, func(t *testing.T) {
			incomplete := Manifest{Tables: manifest.Tables}
			for _, identity := range manifest.Files {
				if identity.Path != test.path {
					incomplete.Files = append(incomplete.Files, identity)
				}
			}
			require.NoError(t, cfg.findMediaReferenceGaps(t.Context(), db, &incomplete))
			require.Contains(t, incomplete.Gaps, test.gap)
		})
	}
	return filepath.Join(cfg.StorageRoots["uploads"], "site-favicon", "current.txt")
}
