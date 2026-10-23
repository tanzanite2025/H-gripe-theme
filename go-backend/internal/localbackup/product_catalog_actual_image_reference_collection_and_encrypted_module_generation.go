package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"commerce-platform/internal/pkg/safehtml"
)

var productCatalogEmbeddedMediaReference = regexp.MustCompile(`(?i)(?:https?://[^\s<>"'()]+|/(?:uploads/|uploads-private/|company/)[^\s<>"'()]+)`)
var productCatalogHTMLMediaSource = regexp.MustCompile(`(?i)\b(?:src|poster|srcset)\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s>]+))`)
var productCatalogCSSMediaSource = regexp.MustCompile(`(?i)url\(\s*["']?([^\s"')]+)["']?\s*\)`)

func (cfg Configuration) collectProductCatalogMediaReferences(snapshot productCatalogModuleSnapshot) (map[string]bool, error) {
	references := map[string]bool{}
	var inspectValue func(any, string) error
	inspectValue = func(value any, key string) error {
		switch typed := value.(type) {
		case map[string]any:
			for childKey, child := range typed {
				if err := inspectValue(child, childKey); err != nil {
					return err
				}
			}
		case []any:
			for _, child := range typed {
				if err := inspectValue(child, key); err != nil {
					return err
				}
			}
		case string:
			mediaField := key == "url" || key == "thumbnail_url" || key == "poster_url" || strings.Contains(key, "image") || strings.Contains(key, "swatch") || key == "logo"
			trimmed := strings.TrimSpace(typed)
			if mediaField && trimmed != "" {
				if strings.HasPrefix(trimmed, "data:") || strings.HasPrefix(trimmed, "blob:") {
					return errors.New("product image uses a nonportable inline or browser-only reference")
				}
				if strings.HasPrefix(trimmed, "/") || strings.Contains(trimmed, "://") {
					references[trimmed] = true
				} else {
					return errors.New("product media reference is not an explicit local or absolute URL")
				}
			}
			for _, match := range productCatalogHTMLMediaSource.FindAllStringSubmatch(typed, -1) {
				for _, candidate := range strings.Split(strings.Join(match[1:], ""), ",") {
					fields := strings.Fields(candidate)
					if len(fields) > 0 {
						references[html.UnescapeString(fields[0])] = true
					}
				}
			}
			for _, match := range productCatalogCSSMediaSource.FindAllStringSubmatch(typed, -1) {
				references[html.UnescapeString(match[1])] = true
			}
			for _, reference := range productCatalogEmbeddedMediaReference.FindAllString(typed, -1) {
				path := strings.ToLower(strings.Split(reference, "?")[0])
				if strings.HasPrefix(reference, "/uploads") || strings.HasPrefix(reference, "/company/") || strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".jpeg") || strings.HasSuffix(path, ".png") || strings.HasSuffix(path, ".webp") || strings.HasSuffix(path, ".avif") || strings.HasSuffix(path, ".gif") || strings.HasSuffix(path, ".svg") || strings.HasSuffix(path, ".mp4") {
					references[reference] = true
				}
			}
		}
		return nil
	}
	for name, table := range snapshot.Tables {
		for _, row := range table.Rows {
			var fields map[string]any
			if err := json.Unmarshal(row, &fields); err != nil {
				return nil, err
			}
			{
				columns, err := productCatalogPersistedRichTextMediaColumnsForTable(name, row)
				if err != nil {
					return nil, err
				}
				for _, column := range columns {
					value, err := productCatalogRowString(row, column)
					if err != nil {
						return nil, err
					}
					document, err := safehtml.ParseRichTextMediaSourceAndFileLinkAttributes(value)
					if err != nil {
						return nil, err
					}
					for _, reference := range document.References {
						if cfg.productCatalogRichTextAttributeRequiresLocalFile(reference) {
							references[*reference.URL] = true
						}
					}
					delete(fields, column)
				}
				for _, column := range productCatalogKnownPlainTextFieldsExcludedFromMediaDiscovery(name) {
					delete(fields, column)
				}
			}
			if err := inspectValue(fields, ""); err != nil {
				return nil, err
			}
		}
	}
	return references, nil
}

func (cfg Configuration) resolveProductCatalogImageSource(reference string) (string, string, error) {
	if root, key := cfg.resolveLocalMediaReference(reference); root == "uploads" && key != "" {
		return cfg.StorageRoots[root], "media/uploads/" + key, nil
	}
	parsed, err := url.Parse(reference)
	if err != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", "", errors.New("product image reference cannot be resolved to a durable local file")
	}
	if parsed.Host != "" {
		recognized := false
		for _, origin := range cfg.PublicOrigins {
			configured, err := url.Parse(origin)
			if err == nil && configured.Scheme == parsed.Scheme && strings.EqualFold(configured.Host, parsed.Host) {
				recognized = true
			}
		}
		if !recognized {
			return "", "", errors.New("remote product images must be registered and stored locally before backup")
		}
	}
	filename := "product-static/" + strings.TrimPrefix(parsed.Path, "/")
	if cfg.ProductStaticAssetsDirectory == "" || !isAllowedBackupFileForScope(productCatalogBackupScope, filename) || !strings.HasPrefix(filename, "product-static/company/") {
		return "", "", errors.New("product image is missing a registered local storage or company static-asset root")
	}
	return cfg.ProductStaticAssetsDirectory, filename, nil
}

func (cfg Configuration) collectProductCatalogImageSources(snapshot productCatalogModuleSnapshot) (map[string]string, error) {
	references, err := cfg.collectProductCatalogMediaReferences(snapshot)
	if err != nil {
		return nil, err
	}
	sources := map[string]string{}
	for reference := range references {
		root, packagePath, err := cfg.resolveProductCatalogImageSource(reference)
		if err != nil {
			return nil, err
		}
		relative := strings.TrimPrefix(strings.TrimPrefix(packagePath, "media/uploads/"), "product-static/")
		source := filepath.Join(root, filepath.FromSlash(relative))
		resolvedRoot, err := filepath.EvalSymlinks(root)
		if err != nil {
			return nil, err
		}
		resolvedSource, err := filepath.EvalSymlinks(source)
		if err != nil {
			return nil, errors.New("a referenced product image file is missing")
		}
		if !pathContains(resolvedRoot, resolvedSource) || filepath.Clean(source) != resolvedSource {
			return nil, errors.New("product image path is outside its registered root or traverses a symbolic link")
		}
		sources[packagePath] = source
	}
	// Stored keys are authoritative even when a historical asset has an empty URL.
	for _, table := range []string{"media_assets", "media_asset_derivatives"} {
		for _, row := range snapshot.Tables[table].Rows {
			key, err := productCatalogRowString(row, "storage_key")
			if err != nil {
				return nil, err
			}
			if key == "" {
				return nil, errors.New("product asset storage key is missing")
			}
			root, packagePath, err := cfg.resolveProductCatalogImageSource("/uploads/" + key)
			if err != nil {
				return nil, err
			}
			source := filepath.Join(root, filepath.FromSlash(key))
			resolved, err := filepath.EvalSymlinks(source)
			if err != nil || resolved != filepath.Clean(source) {
				return nil, errors.New("product asset file is missing or uses a symbolic link")
			}
			sources[packagePath] = source
		}
	}
	return sources, nil
}

func (manager *Manager) CreateProductCatalogBackup(actor, confirmation string) (*Task, error) {
	cfg := manager.Configuration
	if confirmation != "BACKUP PRODUCTS "+cfg.SiteID+"/"+cfg.Environment {
		return nil, errors.New("explicit product backup maintenance confirmation is required")
	}
	if err := cfg.validateRuntimeBoundary(); err != nil {
		return nil, err
	}
	if err := cfg.checkWorkspaceCapacity(64 << 20); err != nil {
		return nil, err
	}
	task, err := manager.createTask("product_catalog_backup", actor)
	if err != nil {
		return nil, err
	}
	if err := manager.startTaskOperation(task, "capturing_product_catalog_and_images", manager.generateProductCatalogEncryptedModule); err != nil {
		task.State = "failed"
		task.Error = err.Error()
		_ = manager.saveTask(task)
		return nil, err
	}
	return manager.GetTask(task.ID)
}

func (manager *Manager) generateProductCatalogEncryptedModule(ctx context.Context, task *Task) (resultErr error) {
	cfg := manager.Configuration
	lease, err := manager.acquireMaintenance(ctx, task)
	if err != nil {
		return err
	}
	defer lease.Close()
	defer func() { resultErr = errors.Join(resultErr, manager.releaseMaintenance(task)) }()
	if err := cfg.validateCompanyStaticSourceReleaseVersion(); err != nil {
		return err
	}
	db, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := assertNoUnfencedDatabaseSessions(ctx, db, cfg.DatabaseName); err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, beforeFingerprint, migration, err := captureDatabaseTransactionInventory(ctx, tx)
	if err != nil {
		return err
	}
	if migration != cfg.ExpectedMigrationVersion {
		return errors.New("product backup migration mismatch")
	}
	snapshot, err := cfg.captureProductCatalogModuleSnapshot(ctx, tx)
	if err != nil {
		return err
	}
	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	if len(data) > productCatalogMaximumJSONBytes {
		return errors.New("product JSON snapshot exceeds capacity")
	}
	directory, _ := manager.taskDirectory(task.ID)
	snapshotDirectory := filepath.Join(directory, "product-snapshot")
	if err := os.Mkdir(snapshotDirectory, 0700); err != nil {
		return err
	}
	defer os.RemoveAll(snapshotDirectory)
	if err := writeDurableFile(filepath.Join(snapshotDirectory, productCatalogRowsFilename), data); err != nil {
		return err
	}
	manifest := Manifest{FormatVersion: PackageVersion, PackageID: task.ID, SiteID: cfg.SiteID, Environment: cfg.Environment, ApplicationVersion: cfg.ApplicationVersion, MigrationVersion: migration, CapturedAt: time.Now().UTC(), Scope: productCatalogBackupScope, Tables: map[string]int64{}, Files: []FileIdentity{}, Gaps: []string{}, DatabaseFingerprint: productCatalogBytesSHA256(data)}
	for name, table := range snapshot.Tables {
		manifest.Tables[name] = int64(len(table.Rows))
	}
	entry, err := fileIdentity(filepath.Join(snapshotDirectory, productCatalogRowsFilename), productCatalogRowsFilename)
	if err != nil {
		return err
	}
	manifest.Files = append(manifest.Files, entry)
	sources, err := cfg.collectProductCatalogImageSources(snapshot)
	if err != nil {
		return err
	}
	paths := []string{}
	for filename := range sources {
		paths = append(paths, filename)
	}
	sort.Strings(paths)
	var total int64 = int64(len(data))
	for _, filename := range paths {
		if err := ctx.Err(); err != nil {
			return err
		}
		info, err := os.Stat(sources[filename])
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return errors.New("product image snapshot refuses non-regular files")
		}
		if info.Size() > cfg.MaxExpandedBytes-total || len(manifest.Files) >= cfg.MaxFiles {
			return errors.New("product image capacity exceeded")
		}
		if err := cfg.checkWorkspaceCapacity(info.Size()); err != nil {
			return err
		}
		entry, err := copyVerifiedRegularFile(sources[filename], filepath.Join(snapshotDirectory, filepath.FromSlash(filename)), filename)
		if err != nil {
			return err
		}
		total += entry.Size
		manifest.Files = append(manifest.Files, entry)
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	_, afterFingerprint, _, err := captureDatabaseInventory(ctx, db)
	if err != nil {
		return err
	}
	if beforeFingerprint != afterFingerprint {
		return errors.New("product source database changed during maintenance capture")
	}
	for _, entry := range manifest.Files {
		if source, ok := sources[entry.Path]; ok {
			current, err := fileIdentity(source, entry.Path)
			if err != nil || current != entry {
				return errors.New("product image changed during maintenance capture")
			}
		}
	}
	if err := validateProductCatalogMediaMetadataAgainstFiles(snapshot, manifest, cfg); err != nil {
		return err
	}
	passphrase, err := readProtectedCredential(cfg.PassphraseFile)
	if err != nil {
		return err
	}
	if err := cfg.checkWorkspaceCapacity(total + (64 << 20)); err != nil {
		return err
	}
	destination, _ := manager.taskPackagePath(task.ID)
	if err := WriteEncryptedPackage(destination, snapshotDirectory, manifest, passphrase, cfg.MaxPackageBytes); err != nil {
		return err
	}
	identity, err := fileIdentity(destination, "")
	if err != nil {
		return err
	}
	task.Manifest = &manifest
	task.Size = identity.Size
	task.CiphertextSHA256 = identity.SHA256
	task.State = "ready_to_download"
	task.Phase = "product_catalog_and_actual_images_encrypted"
	return manager.saveTask(task)
}

func validateProductCatalogPackageFileInventory(snapshot productCatalogModuleSnapshot, manifest Manifest, cfg Configuration) error {
	sources, err := cfg.collectProductCatalogPackagePathsWithoutReadingLiveFiles(snapshot)
	if err != nil {
		return err
	}
	sources[productCatalogRowsFilename] = true
	if len(sources) != len(manifest.Files) {
		return errors.New("product package file inventory contains missing or unrelated files")
	}
	for _, file := range manifest.Files {
		if !sources[file.Path] {
			return fmt.Errorf("unrelated file in product package: %s", file.Path)
		}
	}
	return validateProductCatalogMediaMetadataAgainstFiles(snapshot, manifest, cfg)
}

func validateProductCatalogMediaMetadataAgainstFiles(snapshot productCatalogModuleSnapshot, manifest Manifest, cfg Configuration) error {
	files := map[string]FileIdentity{}
	for _, file := range manifest.Files {
		files[file.Path] = file
	}
	for _, table := range []string{"media_assets", "media_asset_derivatives"} {
		for _, row := range snapshot.Tables[table].Rows {
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(row, &fields); err != nil {
				return err
			}
			if table == "media_assets" {
				visibility, err := productCatalogRowString(row, "visibility")
				if err != nil {
					return err
				}
				partition, err := productCatalogRowString(row, "library_partition")
				if err != nil {
					return err
				}
				if visibility != "public" || partition == "customer_service" {
					return errors.New("product package refuses private or customer-service media")
				}
			}
			key, err := productCatalogRowString(row, "storage_key")
			if err != nil {
				return err
			}
			mediaURL, err := productCatalogRowString(row, "url")
			if err != nil {
				return err
			}
			if mediaURL != "" {
				root, urlKey := cfg.resolveLocalMediaReference(mediaURL)
				if root != "uploads" || urlKey != key {
					return errors.New("product media URL disagrees with its registered storage key")
				}
			}
			_, filename, err := cfg.resolveProductCatalogImageSource("/uploads/" + key)
			if err != nil {
				return err
			}
			file, exists := files[filename]
			if !exists {
				return errors.New("registered product media file is missing")
			}
			if raw := fields["size"]; len(raw) > 0 && string(raw) != "null" {
				var size int64
				if err := json.Unmarshal(raw, &size); err != nil || size != file.Size {
					return errors.New("product media recorded size differs from actual file")
				}
			}
			if raw := fields["content_sha256"]; len(raw) > 0 && string(raw) != "null" {
				var digest string
				if err := json.Unmarshal(raw, &digest); err != nil {
					return err
				}
				if strings.TrimSpace(digest) != "" && strings.TrimSpace(digest) != file.SHA256 {
					return errors.New("product media recorded digest differs from actual file")
				}
			}
		}
	}
	return nil
}

func (cfg Configuration) collectProductCatalogPackagePathsWithoutReadingLiveFiles(snapshot productCatalogModuleSnapshot) (map[string]bool, error) {
	references, err := cfg.collectProductCatalogMediaReferences(snapshot)
	if err != nil {
		return nil, err
	}
	paths := map[string]bool{}
	for reference := range references {
		_, filename, err := cfg.resolveProductCatalogImageSource(reference)
		if err != nil {
			return nil, err
		}
		paths[filename] = true
	}
	for _, table := range []string{"media_assets", "media_asset_derivatives"} {
		for _, row := range snapshot.Tables[table].Rows {
			key, err := productCatalogRowString(row, "storage_key")
			if err != nil {
				return nil, err
			}
			_, filename, err := cfg.resolveProductCatalogImageSource("/uploads/" + key)
			if err != nil {
				return nil, err
			}
			paths[filename] = true
		}
	}
	return paths, nil
}
