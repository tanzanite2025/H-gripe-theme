package localbackup

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/lib/pq"
)

func readMediaModuleConfinedRegularFileIdentity(directory, key, packagePath string) (FileIdentity, error) {
	root, err := os.OpenRoot(directory)
	if err != nil {
		return FileIdentity{}, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(root.Close)
	partial := ""
	for _, component := range strings.Split(filepath.ToSlash(key), "/") {
		partial = filepath.Join(partial, component)
		info, err := root.Lstat(partial)
		if err != nil {
			return FileIdentity{}, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return FileIdentity{}, errors.New("媒体修复拒绝符号链接")
		}
	}
	file, err := root.Open(key)
	if err != nil {
		return FileIdentity{}, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(file.Close)
	before, err := file.Stat()
	if err != nil || !before.Mode().IsRegular() {
		return FileIdentity{}, errors.New("媒体修复只接受普通文件")
	}
	hash := sha256.New()
	size, err := io.Copy(hash, file)
	if err != nil {
		return FileIdentity{}, err
	}
	after, err := root.Lstat(key)
	if err != nil || !os.SameFile(before, after) || size != before.Size() || after.Size() != before.Size() || !after.ModTime().Equal(before.ModTime()) {
		return FileIdentity{}, errors.New("媒体修复读取期间文件已变化")
	}
	return FileIdentity{Path: packagePath, Size: size, SHA256: hex.EncodeToString(hash.Sum(nil))}, nil
}

// Repair bytes only when the current complete registration still equals the
// authenticated package. Missing or changed metadata never grants replacement.
func (cfg Configuration) inspectMediaModuleExistingRegisteredFileByteRepair(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, wanted, original FileIdentity) (*BusinessModuleTransactionalRestoreFile, error) {
	if !reflect.DeepEqual(snapshot.Modules, []string{"media"}) || !strings.HasPrefix(wanted.Path, "media/") || wanted.Path != original.Path || wanted == original || original.Size < 0 || original.Size > cfg.MaxExpandedBytes {
		return nil, nil
	}
	rootName, key, err := transactionalBusinessModuleRestoreMediaPath(wanted.Path)
	if err != nil {
		return nil, err
	}
	matches := 0
	for _, table := range []string{"media_assets", "media_asset_derivatives", "site_logo_assets", "site_favicon_assets"} {
		for _, row := range snapshot.Tables[table].Rows {
			field := func(column string) string { return readMediaModuleRestoreRowText(row, column) }
			if field("storage_key") != key || field("deleted_at") != "" {
				continue
			}
			visibility := field("visibility")
			if table == "media_asset_derivatives" {
				found := false
				for _, parent := range snapshot.Tables["media_assets"].Rows {
					if readMediaModuleRestoreRowText(parent, "id") != field("media_asset_id") {
						continue
					}
					if readMediaModuleRestoreRowText(parent, "deleted_at") != "" || readMediaModuleRestoreRowText(parent, "status") != "active" {
						return nil, nil
					}
					var sameParent bool
					if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM public.media_assets r WHERE to_jsonb(r)=$1::jsonb)`, string(parent)).Scan(&sameParent); err != nil {
						return nil, err
					}
					if !sameParent {
						return nil, nil
					}
					visibility, found = readMediaModuleRestoreRowText(parent, "visibility"), true
				}
				if !found {
					return nil, nil
				}
			}
			if registeredMediaStorageRootForObjectKeyAndVisibility(key, visibility) != rootName {
				continue
			}
			if table == "media_assets" && field("status") != "active" {
				return nil, nil
			}
			var same bool
			if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM public.`+pq.QuoteIdentifier(table)+` r WHERE to_jsonb(r)=$1::jsonb)`, string(row)).Scan(&same); err != nil {
				return nil, err
			}
			if !same {
				return nil, nil
			}
			matches++
		}
	}
	if matches != 1 {
		return nil, nil
	}
	return &BusinessModuleTransactionalRestoreFile{Identity: wanted, OriginalIdentity: &original}, nil
}

func mediaModuleExistingFileByteRepairPlanSHA256(files []BusinessModuleTransactionalRestoreFile) (string, error) {
	if len(files) == 0 {
		return "", nil
	}
	return businessModuleRecoveryCanonicalJSONSHA256(files)
}

func mediaModuleFileRepairStageRelativePath(task *Task, directory, key string) string {
	return filepath.Join(transactionalBusinessModuleRestoreContractForTask(task).AttachmentStagePrefix+task.ID, directory, filepath.FromSlash(key))
}

func (manager *Manager) installMediaModuleRegisteredFileRepairWithOriginalHardLink(task *Task, file *BusinessModuleTransactionalRestoreFile, checkpoint func(string) error) error {
	rootName, key, err := transactionalBusinessModuleRestoreMediaPath(file.Identity.Path)
	if err != nil {
		return err
	}
	directory := manager.Configuration.StorageRoots[rootName]
	root, err := os.OpenRoot(directory)
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(root.Close)
	originalKey := mediaModuleFileRepairStageRelativePath(task, "originals", key)
	replacementKey := mediaModuleFileRepairStageRelativePath(task, "replacements", key)
	for _, stagedKey := range []string{originalKey, replacementKey} {
		if err := root.MkdirAll(filepath.Dir(stagedKey), 0700); err != nil {
			return err
		}
	}
	actual, err := readMediaModuleConfinedRegularFileIdentity(directory, filepath.FromSlash(key), file.Identity.Path)
	if err != nil || file.OriginalIdentity == nil || actual != *file.OriginalIdentity {
		return errors.New("待修复文件已变化，请保留维护并重新核验")
	}
	if err := root.Link(filepath.FromSlash(key), originalKey); err != nil {
		return err
	}
	originalOpenFlags := os.O_RDONLY
	if runtime.GOOS == "windows" {
		originalOpenFlags = os.O_RDWR // FlushFileBuffers requires a writable handle.
	}
	originalFile, err := root.OpenFile(originalKey, originalOpenFlags, 0)
	if err != nil {
		return err
	}
	if err := errors.Join(originalFile.Sync(), originalFile.Close()); err != nil {
		return err
	}
	backupIdentity, err := readMediaModuleConfinedRegularFileIdentity(directory, originalKey, file.Identity.Path)
	if err != nil || backupIdentity != *file.OriginalIdentity {
		return errors.New("替换前的原文件备份无法核验")
	}
	for parent := filepath.Dir(originalKey); parent != "."; parent = filepath.Dir(parent) {
		if err := syncDirectory(filepath.Join(directory, parent)); err != nil {
			return err
		}
	}
	if checkpoint != nil {
		if err := checkpoint("media_original_backup_linked_before_acknowledgement"); err != nil {
			return err
		}
	}
	if err := root.Link(mediaModuleFileRepairStageRelativePath(task, "files", key), replacementKey); err != nil {
		return err
	}
	if err := root.Rename(replacementKey, filepath.FromSlash(key)); err != nil {
		return err
	}
	for _, parent := range []string{filepath.Dir(replacementKey), filepath.Dir(filepath.FromSlash(key))} {
		if err := syncDirectory(filepath.Join(directory, parent)); err != nil {
			return err
		}
	}
	if checkpoint != nil {
		return checkpoint("media_file_replaced_before_acknowledgement")
	}
	return nil
}

// Repeated rollback is safe even when replacement or its acknowledgement was
// interrupted. Unknown bytes or physical identities keep the store isolated.
func (manager *Manager) restoreOriginalMediaModuleFileAfterUncommittedByteRepair(task *Task, file BusinessModuleTransactionalRestoreFile) error {
	rootName, key, err := transactionalBusinessModuleRestoreMediaPath(file.Identity.Path)
	if err != nil {
		return err
	}
	directory := manager.Configuration.StorageRoots[rootName]
	root, err := os.OpenRoot(directory)
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(root.Close)
	actual, err := readMediaModuleConfinedRegularFileIdentity(directory, filepath.FromSlash(key), file.Identity.Path)
	if err != nil || file.OriginalIdentity == nil {
		return errors.New("修复文件或原字节凭据缺失，保留维护")
	}
	originalKey := mediaModuleFileRepairStageRelativePath(task, "originals", key)
	if actual == *file.OriginalIdentity {
		originalInfo, originalErr := root.Lstat(originalKey)
		if errors.Is(originalErr, os.ErrNotExist) {
			return nil // Intent saved before backup link; original still untouched.
		}
		currentInfo, err := root.Lstat(filepath.FromSlash(key))
		if originalErr != nil || err != nil || !os.SameFile(currentInfo, originalInfo) {
			return errors.New("原文件字节相同但物理身份不同，保留维护")
		}
		return nil
	}
	if actual != file.Identity {
		return errors.New("未提交修复后的文件已被其他内容替换，保留维护")
	}
	currentInfo, err := root.Lstat(filepath.FromSlash(key))
	if err != nil {
		return err
	}
	stagedInfo, err := root.Lstat(mediaModuleFileRepairStageRelativePath(task, "files", key))
	if err != nil || !os.SameFile(currentInfo, stagedInfo) {
		return errors.New("未提交修复文件的安装身份已变化")
	}
	original, err := readMediaModuleConfinedRegularFileIdentity(directory, originalKey, file.Identity.Path)
	if err != nil || original != *file.OriginalIdentity {
		return errors.New("未提交修复的原文件无法核验，保留维护")
	}
	rollbackKey := mediaModuleFileRepairStageRelativePath(task, "rollback", key)
	if err := root.MkdirAll(filepath.Dir(rollbackKey), 0700); err != nil {
		return err
	}
	if rollbackInfo, err := root.Lstat(rollbackKey); err == nil {
		originalInfo, err := root.Lstat(originalKey)
		if err != nil || !os.SameFile(rollbackInfo, originalInfo) {
			return errors.New("原文件回退暂存身份已变化")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	} else if err := root.Link(originalKey, rollbackKey); err != nil {
		return err
	}
	if err := root.Rename(rollbackKey, filepath.FromSlash(key)); err != nil {
		return err
	}
	return errors.Join(syncDirectory(filepath.Join(directory, filepath.Dir(filepath.FromSlash(key)))), syncDirectory(filepath.Join(directory, filepath.Dir(rollbackKey))))
}

func validateMediaModuleByteRepairJournalOriginalIdentity(task *Task, file BusinessModuleTransactionalRestoreFile) error {
	if file.OriginalIdentity == nil {
		return nil
	}
	if task.MediaModuleRestore == nil || file.OriginalIdentity.Path != file.Identity.Path || *file.OriginalIdentity == file.Identity || file.OriginalIdentity.Size < 0 || !moduleRecoverySideEffectReceiptRequestDigestIsValid(file.OriginalIdentity.SHA256) {
		return errors.New("媒体修复原字节凭据无效或模块越界")
	}
	return nil
}

// Complete rows never enter this report; counts and a bound plan digest suffice.
func applyMediaModuleExistingByteRepairCountsAndDigestToEligibility(eligibility *MediaModuleRestoreEligibility, repairs []BusinessModuleTransactionalRestoreFile) error {
	var err error
	eligibility.ExistingFileRepairs = len(repairs)
	eligibility.ExistingFileRepairPlanSHA256, err = mediaModuleExistingFileByteRepairPlanSHA256(repairs)
	return err
}

// Cleanup accepts these extra paths only for a journaled media byte repair.
func mediaModuleByteRepairAdditionalStagingDirectoryNames(file BusinessModuleTransactionalRestoreFile) []string {
	if file.OriginalIdentity == nil {
		return nil
	}
	return strings.Fields("originals replacements rollback")
}
