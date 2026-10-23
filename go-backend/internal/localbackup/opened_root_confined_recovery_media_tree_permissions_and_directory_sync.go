package localbackup

import (
	"context"
	"errors"
	"io/fs"
	"os"
)

// A zero fileMode leaves regular files unchanged. All metadata updates use
// opened handles confined to the recovery tree, including during directory walks.
func applyRecoveryMediaTreePermissionsWithinOpenedRoot(ctx context.Context, treeDirectory string, fileMode, directoryMode os.FileMode, skipRoot bool) error {
	root, err := os.OpenRoot(treeDirectory)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()
	return fs.WalkDir(root.FS(), ".", func(relativePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return errors.New("recovery media permission tree contains a symbolic link")
		}
		if skipRoot && relativePath == "." {
			return nil
		}
		mode := fileMode
		if entry.IsDir() {
			mode = directoryMode
		} else if fileMode == 0 {
			return nil
		}
		file, err := root.Open(relativePath)
		if err != nil {
			return err
		}
		permissionErr := file.Chmod(mode) // #nosec G122 -- Chmod acts on the file handle opened through os.Root, so no path is resolved again during the walk.
		if permissionErr == nil && entry.IsDir() {
			permissionErr = syncDirectoryFile(file)
		}
		return errors.Join(permissionErr, file.Close())
	})
}
