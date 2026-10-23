package localbackup

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRecoveryMediaPermissionsRejectExternalSymlinksAndPreserveOutsideFiles(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("requires POSIX permissions and unprivileged symbolic links")
	}
	tree, outside := t.TempDir(), t.TempDir()
	insideFile := filepath.Join(tree, "media.bin")
	outsideFile := filepath.Join(outside, "private.bin")
	require.NoError(t, os.WriteFile(insideFile, []byte("media"), 0600))
	require.NoError(t, os.WriteFile(outsideFile, []byte("private"), 0600))
	require.NoError(t, applyRecoveryMediaTreePermissionsWithinOpenedRoot(context.Background(), tree, 0644, 0755, false))
	insideInfo, err := os.Stat(insideFile)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0644), insideInfo.Mode().Perm())
	require.NoError(t, os.Symlink(outsideFile, filepath.Join(tree, "external-link")))
	require.Error(t, applyRecoveryMediaTreePermissionsWithinOpenedRoot(context.Background(), tree, 0644, 0755, false))
	outsideInfo, err := os.Stat(outsideFile)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0600), outsideInfo.Mode().Perm())
}
