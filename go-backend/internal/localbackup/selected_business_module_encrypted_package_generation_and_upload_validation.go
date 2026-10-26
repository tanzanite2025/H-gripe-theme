package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (manager *Manager) CreateSelectedBusinessModuleBackup(actor, confirmation string, modules []string) (*Task, error) {
	modules, err := normalizeSelectedBusinessBackupModules(modules)
	if err != nil {
		return nil, err
	}
	cfg := manager.Configuration
	if confirmation != "BACKUP MODULES "+cfg.SiteID+"/"+cfg.Environment+"/"+strings.Join(modules, ",") {
		return nil, errors.New("请核对所选模块并输入对应的备份确认文字")
	}
	// Preserve the existing, stronger product-only package and isolated restore.
	if len(modules) == 1 && modules[0] == "products" {
		return manager.CreateProductCatalogBackup(actor, "BACKUP PRODUCTS "+cfg.SiteID+"/"+cfg.Environment)
	}
	if err := cfg.validateRuntimeBoundary(); err != nil {
		return nil, err
	}
	if err := cfg.checkWorkspaceCapacity(64 << 20); err != nil {
		return nil, err
	}
	task, err := manager.createTask("selected_business_modules_backup", actor)
	if err != nil {
		return nil, err
	}
	task.SelectedBackupModules = modules
	if err := manager.startTaskOperation(task, "capturing_selected_business_modules", manager.generateSelectedBusinessModuleEncryptedPackage); err != nil {
		task.State, task.Error = "failed", err.Error()
		_ = manager.saveTask(task)
		return nil, err
	}
	return manager.GetTask(task.ID)
}

func (manager *Manager) generateSelectedBusinessModuleEncryptedPackage(ctx context.Context, task *Task) (resultErr error) {
	cfg := manager.Configuration
	modules, err := normalizeSelectedBusinessBackupModules(task.SelectedBackupModules)
	if err != nil {
		return err
	}
	lease, err := manager.acquireMaintenance(ctx, task)
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(lease.Close)
	defer func() { resultErr = errors.Join(resultErr, manager.releaseMaintenance(task)) }()
	db, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(db.Close)
	if err := assertNoUnfencedDatabaseSessions(ctx, db, cfg.DatabaseName); err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(tx.Rollback)
	_, beforeFingerprint, migration, err := captureDatabaseTransactionInventory(ctx, tx)
	if err != nil {
		return err
	}
	if migration != cfg.ExpectedMigrationVersion || !moduleDependencyRegistrySupportsAuditedMigrationVersion(migration) {
		return errors.New("模块备份与当前数据库版本不兼容")
	}
	snapshot, err := cfg.captureSelectedBusinessModuleCompleteSnapshot(ctx, tx, modules)
	if err != nil {
		return err
	}
	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	if len(data) > productCatalogMaximumJSONBytes {
		return errors.New("模块组合超过 64 MiB 记录容量，请分别导出或使用整站备份")
	}
	manifest := Manifest{FormatVersion: PackageVersion, PackageID: task.ID, SiteID: cfg.SiteID, Environment: cfg.Environment, ApplicationVersion: cfg.ApplicationVersion, MigrationVersion: migration, CapturedAt: time.Now().UTC(), Scope: selectedBusinessModulesBackupScope, SelectedModules: modules, Tables: map[string]int64{}, Files: []FileIdentity{}, Gaps: []string{}, DatabaseFingerprint: productCatalogBytesSHA256(data)}
	if err := tx.QueryRowContext(ctx, `SELECT pg_export_snapshot()`).Scan(&manifest.SnapshotID); err != nil {
		return err
	}
	rowCount := 0
	for name, table := range snapshot.Tables {
		manifest.Tables[name] = int64(len(table.Rows))
		rowCount += len(table.Rows)
	}
	if rowCount > productCatalogMaximumRows {
		return errors.New("模块组合超过 100000 条记录，请分别导出或使用整站备份")
	}
	directory, _ := manager.taskDirectory(task.ID)
	staging := filepath.Join(directory, "selected-business-module-snapshot")
	if err := os.Mkdir(staging, 0700); err != nil {
		return err
	}
	defer func(cleanupDirectory string) { _ = os.RemoveAll(cleanupDirectory) }(staging)
	if err := writeDurableFile(filepath.Join(staging, selectedBusinessModulesRowsFilename), data); err != nil {
		return err
	}
	entry, err := fileIdentity(filepath.Join(staging, selectedBusinessModulesRowsFilename), selectedBusinessModulesRowsFilename)
	if err != nil {
		return err
	}
	manifest.Files = append(manifest.Files, entry)
	allMedia := false
	for _, module := range modules {
		allMedia = allMedia || module == "media"
	}
	if allMedia {
		requiredCapacity, err := cfg.estimateBackupWorkspaceBytes(ctx)
		if err != nil {
			return err
		}
		if err := cfg.checkWorkspaceCapacity(requiredCapacity); err != nil {
			return err
		}
		if err := cfg.copyMediaSnapshot(ctx, staging, &manifest); err != nil {
			return err
		}
		staticManifest := Manifest{Files: manifest.Files}
		if err := cfg.copyWholeApplicationCompanyStaticSnapshot(ctx, staging, &staticManifest); err != nil {
			return err
		}
		manifest.Files = staticManifest.Files
	} else {
		var expanded = int64(len(data))
		for _, filename := range snapshot.RequiredFilePaths {
			if err := ctx.Err(); err != nil {
				return err
			}
			source, err := cfg.resolveSelectedBusinessModuleConfinedFileSource(filename)
			if err != nil {
				return err
			}
			info, err := os.Stat(source)
			if err != nil || !info.Mode().IsRegular() {
				return errors.New("所选模块缺少实际附件或图片文件")
			}
			if info.Size() > cfg.MaxExpandedBytes-expanded || len(manifest.Files) >= cfg.MaxFiles {
				return errors.New("模块文件超过备份容量限制")
			}
			if err := cfg.checkWorkspaceCapacity(info.Size()); err != nil {
				return err
			}
			entry, err := copyVerifiedRegularFile(source, filepath.Join(staging, filepath.FromSlash(filename)), filename)
			if err != nil {
				return err
			}
			expanded += entry.Size
			manifest.Files = append(manifest.Files, entry)
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	_, afterFingerprint, _, err := captureDatabaseInventory(ctx, db)
	if err != nil {
		return err
	}
	if beforeFingerprint != afterFingerprint {
		return errors.New("维护期间业务数据发生变化，备份已拒绝")
	}
	for _, file := range manifest.Files {
		if file.Path == selectedBusinessModulesRowsFilename {
			continue
		}
		source, err := cfg.resolveSelectedBusinessModuleConfinedFileSource(file.Path)
		if err != nil {
			return err
		}
		current, err := fileIdentity(source, file.Path)
		if err != nil || current != file {
			return errors.New("维护期间模块附件发生变化，备份已拒绝")
		}
	}
	if allMedia {
		current, err := cfg.mediaFingerprint(ctx)
		if err != nil || current != fingerprintManifestMedia(manifest) {
			return errors.New("模块媒体清单在维护期间发生变化")
		}
		if cfg.ProductStaticAssetsDirectory != "" {
			current, err := cfg.companyStaticReleaseFileFingerprint(ctx)
			if err != nil || current != fingerprintCompanyStaticManifestFiles(manifest.Files) {
				return errors.New("模块 company 静态文件清单在维护期间发生变化")
			}
		}
	}
	cfg.sortedManifestFiles(&manifest)
	if err := cfg.validateSelectedBusinessModulePackageContents(snapshot, manifest); err != nil {
		return err
	}
	passphrase, err := readProtectedCredential(cfg.PassphraseFile)
	if err != nil {
		return err
	}
	var total int64
	for _, file := range manifest.Files {
		if file.Size > cfg.MaxExpandedBytes-total || len(manifest.Files) > cfg.MaxFiles {
			return errors.New("模块组合总容量超过限制")
		}
		total += file.Size
	}
	if err := cfg.checkWorkspaceCapacity(total + (64 << 20)); err != nil {
		return err
	}
	destination, _ := manager.taskPackagePath(task.ID)
	if err := WriteEncryptedPackage(destination, staging, manifest, passphrase, cfg.MaxPackageBytes); err != nil {
		return err
	}
	identity, err := fileIdentity(destination, "")
	if err != nil {
		return err
	}
	task.Manifest, task.Size, task.CiphertextSHA256 = &manifest, identity.Size, identity.SHA256
	task.State, task.Phase = "ready_to_download", "selected_business_modules_encrypted"
	return manager.saveTask(task)
}

func (cfg Configuration) resolveSelectedBusinessModuleConfinedFileSource(filename string) (string, error) {
	if !isAllowedBackupFileForScope(selectedBusinessModulesBackupScope, filename) || filename == selectedBusinessModulesRowsFilename {
		return "", errors.New("模块附件路径不在允许范围")
	}
	root, relative := "", ""
	for _, name := range mediaRoots {
		prefix := "media/" + name + "/"
		if strings.HasPrefix(filename, prefix) {
			root, relative = cfg.StorageRoots[name], strings.TrimPrefix(filename, prefix)
		}
	}
	if strings.HasPrefix(filename, companyStaticPackagePrefix) {
		root, relative = cfg.ProductStaticAssetsDirectory, strings.TrimPrefix(filename, "product-static/")
	}
	if root == "" {
		return "", errors.New("模块文件缺少本地存储目录")
	}
	source := filepath.Join(root, filepath.FromSlash(relative))
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	resolvedSource, err := filepath.EvalSymlinks(source)
	if err != nil || resolvedSource != filepath.Clean(source) || !pathContains(resolvedRoot, resolvedSource) {
		return "", fmt.Errorf("模块图片或附件缺失、越界或包含符号链接：%s", filename)
	}
	return source, nil
}

func (cfg Configuration) validateSelectedBusinessModulePackageContents(snapshot selectedBusinessModuleCompleteSnapshot, manifest Manifest) error {
	modules, err := normalizeSelectedBusinessBackupModules(snapshot.Modules)
	if err != nil || snapshot.ContractVersion != 1 || !reflect.DeepEqual(modules, snapshot.Modules) || !reflect.DeepEqual(modules, manifest.SelectedModules) {
		return errors.New("模块清单与备份范围不一致")
	}
	allowed := selectedBusinessModuleTableSelectors(modules)
	allowed["media_assets"], allowed["media_asset_derivatives"] = "referenced", "referenced"
	if len(allowed) != len(snapshot.Tables) || len(manifest.Tables) != len(snapshot.Tables) {
		return errors.New("模块表清单缺失或包含未选择的数据")
	}
	rows := 0
	for name, contents := range snapshot.Tables {
		if _, exists := allowed[name]; !exists || manifest.Tables[name] != int64(len(contents.Rows)) || len(contents.SchemaSHA256) != 64 || len(contents.ConstraintsSHA256) != 64 {
			return errors.New("模块数据表范围或内容清单不匹配")
		}
		rows += len(contents.Rows)
		if rows > productCatalogMaximumRows {
			return errors.New("模块记录超过容量限制")
		}
		for _, raw := range contents.Rows {
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
				return errors.New("模块完整记录格式无效")
			}
		}
	}
	paths, err := cfg.collectSelectedBusinessModuleRequiredFilePaths(snapshot)
	if err != nil {
		return err
	}
	required := []string{}
	for path := range paths {
		required = append(required, path)
	}
	sort.Strings(required)
	if !reflect.DeepEqual(required, snapshot.RequiredFilePaths) {
		return errors.New("模块实际文件引用清单不匹配")
	}
	allMedia := false
	for _, module := range modules {
		allMedia = allMedia || module == "media"
	}
	files := map[string]FileIdentity{}
	for _, file := range manifest.Files {
		if file.Path != selectedBusinessModulesRowsFilename && !paths[file.Path] && !allMedia {
			return errors.New("模块包夹带了未选择的文件")
		}
		files[file.Path] = file
	}
	for path := range paths {
		if _, exists := files[path]; !exists {
			return errors.New("模块包缺少引用的实际文件")
		}
	}
	assetVisibility := map[string]string{}
	for _, row := range snapshot.Tables["media_assets"].Rows {
		id, _ := productCatalogRowString(row, "id")
		assetVisibility[id], _ = productCatalogRowString(row, "visibility")
	}
	for _, table := range []string{"media_assets", "media_asset_derivatives", "order_evidence_attachments"} {
		for _, row := range snapshot.Tables[table].Rows {
			key, _ := productCatalogRowString(row, "storage_key")
			visibility, _ := productCatalogRowString(row, "visibility")
			if table == "media_asset_derivatives" {
				assetID, _ := productCatalogRowString(row, "media_asset_id")
				visibility = assetVisibility[assetID]
			}
			root := registeredMediaStorageRootForObjectKeyAndVisibility(key, visibility)
			if table == "order_evidence_attachments" {
				root = "uploads_private"
			}
			path := "media/" + root + "/" + key
			digest, _ := productCatalogRowString(row, "content_sha256")
			if digest == "" {
				digest, _ = productCatalogRowString(row, "sha256")
			}
			size, _ := productCatalogRowString(row, "size")
			if size == "" {
				size, _ = productCatalogRowString(row, "size_bytes")
			}
			declaredSize, parseErr := strconv.ParseInt(size, 10, 64)
			if key != "" && ((digest != "" && files[path].SHA256 != strings.ToLower(digest)) || (parseErr == nil && declaredSize > 0 && files[path].Size != declaredSize)) {
				return errors.New("模块媒体实际内容与登记大小或摘要不一致")
			}
		}
	}
	return nil
}

func (manager *Manager) readValidatedSelectedBusinessModuleSnapshot(task *Task, extracted string) (selectedBusinessModuleCompleteSnapshot, error) {
	var snapshot selectedBusinessModuleCompleteSnapshot
	filename := filepath.Join(extracted, selectedBusinessModulesRowsFilename)
	info, err := os.Stat(filename)
	if err != nil || info.Size() > productCatalogMaximumJSONBytes {
		return snapshot, errors.New("模块数据文件缺失或超过容量")
	}
	data, err := os.ReadFile(filename) // #nosec G304 -- Fixed filename in the authenticated private extraction workspace.
	if err != nil {
		return snapshot, err
	}
	if err := decodeStrictJSON(data, &snapshot); err != nil {
		return snapshot, err
	}
	if task.Manifest.DatabaseFingerprint != productCatalogBytesSHA256(data) {
		return snapshot, errors.New("模块记录内容摘要不一致")
	}
	if err := manager.Configuration.validateSelectedBusinessModulePackageContents(snapshot, *task.Manifest); err != nil {
		return snapshot, err
	}
	return snapshot, nil
}

func (manager *Manager) validateUploadedSelectedBusinessModulePackage(task *Task, extracted string) error {
	snapshot, err := manager.readValidatedSelectedBusinessModuleSnapshot(task, extracted)
	if err != nil {
		return err
	}
	task.SelectedBackupModules = snapshot.Modules
	// Deliberately distinct from "verified": no isolated database restore or
	// online apply has occurred, and the whole-storefront apply gate stays shut.
	task.State, task.Phase = "module_package_integrity_verified", "selected_module_contents_verified_restore_unavailable"
	return manager.saveTask(task)
}
