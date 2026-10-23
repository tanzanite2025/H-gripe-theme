package localbackup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLiveMediaFingerprintMatchesSnapshotForAdjacentDirectoryAndFilenamePrefixesAndDetectsChangedBytes(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	for _, root := range mediaRoots {
		for _, key := range []string{"site-logo.txt", "site-logo/current.txt", "nested/item.txt", "nested/item/current.txt"} {
			filename := filepath.Join(cfg.StorageRoots[root], filepath.FromSlash(key))
			require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0700))
			require.NoError(t, os.WriteFile(filename, []byte(root+":"+key), 0600))
		}
	}
	manifest := Manifest{}
	require.NoError(t, cfg.copyMediaSnapshot(t.Context(), t.TempDir(), &manifest))
	liveFingerprint, err := cfg.mediaFingerprint(t.Context())
	require.NoError(t, err)
	require.Equal(t, fingerprintManifestMedia(manifest), liveFingerprint, "unchanged nested directories must not be mistaken for concurrent writes")
	for left, right := 0, len(manifest.Files)-1; left < right; left, right = left+1, right-1 {
		manifest.Files[left], manifest.Files[right] = manifest.Files[right], manifest.Files[left]
	}
	require.Equal(t, fingerprintManifestMedia(manifest), liveFingerprint, "manifest input order must not affect content identity")
	require.NoError(t, os.WriteFile(filepath.Join(cfg.StorageRoots["uploads"], "site-logo", "current.txt"), []byte("different bytes"), 0600))
	changedFingerprint, err := cfg.mediaFingerprint(t.Context())
	require.NoError(t, err)
	require.NotEqual(t, liveFingerprint, changedFingerprint)
}
