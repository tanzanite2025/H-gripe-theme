package localbackup

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCustomerServiceAttachmentInspectionRejectsEscapesSymlinksAndUnexpectedFiles(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "attachment.txt"), []byte("original"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(outside, "private.txt"), []byte("outside private data"), 0600))
	identity, err := inspectBoundedModuleRecoveryRegularFileIdentity(context.Background(), root, "attachment.txt", 8)
	require.NoError(t, err)
	require.EqualValues(t, 8, identity.Size)
	require.Len(t, identity.SHA256, 64)
	for _, key := range []string{"../private.txt", "/private.txt", "..\\private.txt", "C:/private.txt", ".local-backup-forbidden/file"} {
		_, err := inspectBoundedModuleRecoveryRegularFileIdentity(context.Background(), root, key, 1024)
		require.Error(t, err, key)
	}
	_, err = inspectBoundedModuleRecoveryRegularFileIdentity(context.Background(), root, "attachment.txt", 7)
	require.ErrorIs(t, err, errModuleRecoveryAttachmentExceedsExpectedSize)
	require.NoError(t, os.Mkdir(filepath.Join(root, "directory"), 0700))
	_, err = inspectBoundedModuleRecoveryRegularFileIdentity(context.Background(), root, "directory", 1024)
	require.Error(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = inspectBoundedModuleRecoveryRegularFileIdentity(ctx, root, "attachment.txt", 8)
	require.ErrorIs(t, err, context.Canceled)
	t.Run("symlinks_cannot_escape_or_alias_files", func(t *testing.T) {
		if err := os.Symlink(outside, filepath.Join(root, "linked-directory")); err != nil {
			t.Skip("symbolic links are unavailable to this test process")
		}
		_, err := inspectBoundedModuleRecoveryRegularFileIdentity(context.Background(), root, "linked-directory/private.txt", 1024)
		require.Error(t, err)
		require.NoError(t, os.Symlink(filepath.Join(root, "attachment.txt"), filepath.Join(root, "linked-file")))
		_, err = inspectBoundedModuleRecoveryRegularFileIdentity(context.Background(), root, "linked-file", 1024)
		require.Error(t, err)
	})
}
