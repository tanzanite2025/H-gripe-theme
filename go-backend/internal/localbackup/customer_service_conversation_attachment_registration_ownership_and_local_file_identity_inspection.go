package localbackup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type ModuleRecoveryFileInspection struct {
	PackageReferencedFiles  int `json:"package_referenced_files"`
	TargetReferencedFiles   int `json:"target_referenced_files"`
	PackageVerifiedFiles    int `json:"package_verified_files"`
	TargetMatchingFiles     int `json:"target_matching_files"`
	TargetMissingFiles      int `json:"target_missing_files"`
	TargetDifferentFiles    int `json:"target_different_files"`
	TargetOnlyVerifiedFiles int `json:"target_only_verified_files"`
}

type moduleRecoveryFileReference struct {
	Root, Key, Table, RowIdentity string
	Digest                        string
	Size                          int64
	SizeKnown                     bool
}

type moduleRecoveryInspectedFileSnapshot struct {
	Identity           FileIdentity
	Available, Missing bool
	ContentConflict    bool
}

func (reference moduleRecoveryFileReference) packagePath() string {
	return "media/" + reference.Root + "/" + reference.Key
}

func (cfg Configuration) customerServiceMediaAssetLocalIdentity(row *moduleRecoveryDependencyRow) (string, string) {
	key := row.referenceColumnText("storage_key")
	root := registeredMediaStorageRootForObjectKeyAndVisibility(key, row.referenceColumnText("visibility"))
	if key == "" {
		root, key = cfg.resolveLocalMediaReference(row.referenceColumnText("url"))
	}
	if !safePackagePath("media/" + root + "/" + key) {
		return "", ""
	}
	return root, key
}

func (cfg Configuration) includeCustomerServiceLegacyAttachmentAssets(ctx context.Context, graph *moduleRecoveryRowDependencyGraph) error {
	paths := map[string]bool{}
	keySet := map[string]map[string]json.RawMessage{}
	urlSet := map[string]map[string]json.RawMessage{}
	if err := cfg.includeCustomerServiceStructuredMetadataReferences(ctx, graph); err != nil {
		return err
	}
	for _, reference := range graph.LegacyFiles {
		paths[reference.packagePath()] = true
		keyJSON, _ := json.Marshal(reference.Key)
		keySet[reference.Key] = map[string]json.RawMessage{"storage_key": keyJSON}
	}
	for _, message := range graph.Rows["ticket_messages"] {
		if !graph.Scoped["ticket_messages"][message.Identity] {
			continue
		}
		identity := moduleRecoveryDependencyRowPublicIdentity(message, graph.Observation.PrimaryKeys["ticket_messages"])
		raw := strings.TrimSpace(message.referenceColumnText("attachments"))
		format := message.referenceColumnText("attachment_reference_format")
		if format != "" && format != "legacy_urls" && format != "media_asset_bindings" {
			graph.markIncomplete("attachment_reference_format_unsupported", "ticket_messages", "存在未支持的附件格式。")
			continue
		}
		if raw == "" || raw == "[]" || raw == "null" {
			continue
		}
		if format == "media_asset_bindings" {
			graph.Report.recordIssue("attachment_binding_legacy_field_conflict", "blocking", graph.Side, "ticket_messages", identity, "媒体绑定消息仍带历史附件字段，须核对实际归属。")
			continue
		}
		var references []string
		if json.Unmarshal([]byte(raw), &references) != nil {
			graph.Report.Complete = false
			graph.Report.recordIssue("legacy_attachment_json_invalid", "blocking", graph.Side, "ticket_messages", identity, "旧附件 JSON 解析失败；正文仍保留，但无法确定附件范围，不能据此清理对象。")
			continue
		}
		for _, reference := range references {
			root, key := cfg.resolveLocalMediaReference(reference)
			if key == "" {
				graph.Report.Complete = false
				graph.Report.recordIssue("legacy_attachment_reference_unresolved", "blocking", graph.Side, "ticket_messages", identity, "历史附件引用无法映射到本商城本地文件，须核对原始对象。")
				continue
			}
			if len(graph.LegacyFiles) >= graph.Limits.Rows {
				graph.markIncomplete("attachment_reference_resource_limit", "ticket_messages", "历史附件引用超过上限，检查未完成。")
				break
			}
			fileReference := moduleRecoveryFileReference{Root: root, Key: key, Table: "ticket_messages", RowIdentity: identity}
			graph.LegacyFiles = append(graph.LegacyFiles, fileReference)
			paths[fileReference.packagePath()] = true
			keyJSON, _ := json.Marshal(key)
			urlJSON, _ := json.Marshal(reference)
			keySet[key] = map[string]json.RawMessage{"storage_key": keyJSON}
			urlSet[reference] = map[string]json.RawMessage{"url": urlJSON}
		}
	}
	registeredPaths := map[string]bool{}
	for _, lookup := range []struct {
		column string
		keys   map[string]map[string]json.RawMessage
	}{{"storage_key", keySet}, {"url", urlSet}} {
		var keys []map[string]json.RawMessage
		var sortedKeys []string
		for key := range lookup.keys {
			sortedKeys = append(sortedKeys, key)
		}
		sort.Strings(sortedKeys)
		for _, key := range sortedKeys {
			keys = append(keys, lookup.keys[key])
		}
		for start := 0; start < len(keys) && !graph.Exhausted; start += 200 {
			assets, err := graph.readRowsMatchingReferenceKeys(ctx, "media_assets", []string{lookup.column}, keys[start:min(start+200, len(keys))], graph.Limits.Rows-graph.Count)
			if err != nil {
				return err
			}
			for _, asset := range assets {
				root, key := cfg.customerServiceMediaAssetLocalIdentity(asset)
				path := "media/" + root + "/" + key
				if paths[path] {
					registeredPaths[path] = true
					graph.includeRowsInDependencyGraph([]*moduleRecoveryDependencyRow{asset}, "media_assets", false)
				}
			}
			if !graph.Observation.Columns["media_asset_derivatives"][lookup.column] {
				continue
			}
			derivatives, err := graph.readRowsMatchingReferenceKeys(ctx, "media_asset_derivatives", []string{lookup.column}, keys[start:min(start+200, len(keys))], graph.Limits.Rows-graph.Count)
			if err != nil {
				return err
			}
			for _, derivative := range derivatives {
				parents, err := graph.readRowsMatchingReferenceKeys(ctx, "media_assets", []string{"id"}, []map[string]json.RawMessage{{"id": derivative.Values["media_asset_id"]}}, graph.Limits.Rows-graph.Count)
				if err != nil {
					return err
				}
				for _, parent := range parents {
					root, _ := cfg.customerServiceMediaAssetLocalIdentity(parent)
					key := derivative.referenceColumnText("storage_key")
					if key == "" {
						root, key = cfg.resolveLocalMediaReference(derivative.referenceColumnText("url"))
					}
					path := "media/" + root + "/" + key
					if paths[path] {
						registeredPaths[path] = true
						graph.includeRowsInDependencyGraph([]*moduleRecoveryDependencyRow{parent}, "media_assets", false)
						graph.includeRowsInDependencyGraph([]*moduleRecoveryDependencyRow{derivative}, "media_asset_derivatives", false)
					}
				}
			}
		}
	}
	for _, reference := range graph.LegacyFiles {
		if !registeredPaths[reference.packagePath()] {
			graph.Report.recordIssue("legacy_attachment_registration_missing", "blocking", graph.Side, reference.Table, reference.RowIdentity, "历史附件文件缺少可核对的媒体登记，不能推断其所有权或执行删除。")
		}
	}
	return nil
}

func (cfg Configuration) collectCustomerServiceConversationFileReferences(graph *moduleRecoveryRowDependencyGraph) []moduleRecoveryFileReference {
	references := append([]moduleRecoveryFileReference{}, graph.LegacyFiles...)
	for _, asset := range graph.Rows["media_assets"] {
		identity := moduleRecoveryDependencyRowPublicIdentity(asset, graph.Observation.PrimaryKeys["media_assets"])
		partition, purpose, classification := asset.referenceColumnText("library_partition"), asset.referenceColumnText("customer_service_media_purpose"), asset.referenceColumnText("classification_method")
		if partition != "customer_service" || purpose != "conversation_attachment" || (classification != "recorded_upload_context" && classification != "manual_reviewed") {
			code, message := "attachment_media_ownership_review_required", "附件属于共享、复用或未分类媒体，保留原归属，不能随会话整体覆盖媒体行。"
			if classification == "historical_unknown" || classification == "" || partition == "unclassified" {
				code, message = "historical_attachment_media_ownership_unknown", "历史媒体归属未确认，禁止按 URL、路径前缀或当前引用猜测其所有权。"
			}
			graph.Report.recordIssue(code, "blocking", graph.Side, "media_assets", identity, message)
		}
		if asset.referenceColumnText("customer_service_derivative_write_token") != "" || asset.referenceColumnText("customer_service_derivative_write_expires_at") != "" {
			graph.Report.recordIssue("attachment_derivative_write_lease_present", "blocking", graph.Side, "media_assets", identity, "媒体仍有派生写入租约，须先确认写入状态及文件版本。")
		}
		root, key := cfg.customerServiceMediaAssetLocalIdentity(asset)
		if key == "" {
			graph.Report.recordIssue("attachment_local_object_unresolved", "blocking", graph.Side, "media_assets", identity, "媒体登记无法安全映射到本商城的本地文件。")
			continue
		}
		size, _ := strconv.ParseInt(asset.referenceColumnText("size"), 10, 64)
		references = append(references, moduleRecoveryFileReference{Root: root, Key: key, Table: "media_assets", RowIdentity: identity, Digest: strings.ToLower(asset.referenceColumnText("content_sha256")), Size: size})
	}
	for _, derivative := range graph.Rows["media_asset_derivatives"] {
		identity := moduleRecoveryDependencyRowPublicIdentity(derivative, graph.Observation.PrimaryKeys["media_asset_derivatives"])
		asset := findCustomerServiceDependencyRowByID(graph, "media_assets", derivative.Values["media_asset_id"])
		root := "uploads"
		if asset != nil && asset.referenceColumnText("visibility") == "private" {
			root = "uploads_private"
		}
		key := derivative.referenceColumnText("storage_key")
		if key == "" {
			root, key = cfg.resolveLocalMediaReference(derivative.referenceColumnText("url"))
		}
		if !safePackagePath("media/" + root + "/" + key) {
			graph.Report.recordIssue("attachment_derivative_path_unresolved", "blocking", graph.Side, "media_asset_derivatives", identity, "派生文件路径无法安全映射。")
			continue
		}
		size, _ := strconv.ParseInt(derivative.referenceColumnText("size"), 10, 64)
		references = append(references, moduleRecoveryFileReference{Root: root, Key: key, Table: "media_asset_derivatives", RowIdentity: identity, Size: size})
	}
	for _, session := range graph.Rows["media_upload_sessions"] {
		if session.referenceColumnText("status") != "ready" && session.referenceColumnText("status") != "bound" {
			continue
		}
		identity := moduleRecoveryDependencyRowPublicIdentity(session, graph.Observation.PrimaryKeys["media_upload_sessions"])
		key := session.referenceColumnText("object_key")
		if !safePackagePath("media/uploads_private/" + key) {
			graph.Report.recordIssue("upload_receipt_path_unresolved", "blocking", graph.Side, "media_upload_sessions", identity, "上传收据原件路径无法安全映射。")
			continue
		}
		asset := findCustomerServiceDependencyRowByID(graph, "media_assets", session.Values["media_asset_id"])
		if asset != nil {
			root, assetKey := cfg.customerServiceMediaAssetLocalIdentity(asset)
			if root != "uploads_private" || assetKey != key {
				graph.Report.recordIssue("upload_receipt_asset_path_conflict", "blocking", graph.Side, "media_upload_sessions", identity, "上传收据路径与登记的私有原件不一致。")
			}
		}
		references = append(references, moduleRecoveryFileReference{Root: "uploads_private", Key: key, Table: "media_upload_sessions", RowIdentity: identity})
	}
	return references
}

func (cfg Configuration) inspectCustomerServiceConversationAttachmentFiles(ctx context.Context, graphs []*moduleRecoveryRowDependencyGraph, manifest *Manifest, verifiedDirectory string, report *ModuleRecoveryRowInspection) error {
	return cfg.inspectModuleRecoveryAttachmentFiles(ctx, graphs, manifest, verifiedDirectory, report, cfg.collectCustomerServiceConversationFileReferences)
}

func (cfg Configuration) inspectModuleRecoveryAttachmentFiles(ctx context.Context, graphs []*moduleRecoveryRowDependencyGraph, manifest *Manifest, verifiedDirectory string, report *ModuleRecoveryRowInspection, collectReferences func(*moduleRecoveryRowDependencyGraph) []moduleRecoveryFileReference) error {
	manifestFiles := make(map[string]FileIdentity, len(manifest.Files))
	for _, entry := range manifest.Files {
		manifestFiles[entry.Path] = entry
	}
	var packageReferences []moduleRecoveryFileReference
	for _, graph := range graphs {
		references := collectReferences(graph)
		if graph.Side == "package" {
			packageReferences = references
		} else {
			references = append(references, packageReferences...)
		}
		seen := map[string]bool{}
		for _, reference := range references {
			if err := ctx.Err(); err != nil {
				return err
			}
			path := reference.packagePath()
			expected, inManifest := manifestFiles[path]
			if graph.Side == "package" && inManifest && (reference.Digest != "" && reference.Digest != expected.SHA256 || (reference.SizeKnown || reference.Size > 0) && reference.Size != expected.Size) {
				report.recordIssue("attachment_registration_manifest_conflict", "blocking", graph.Side, reference.Table, reference.RowIdentity, "媒体登记的大小或摘要与包内文件清单不符。")
			}
			if seen[path] {
				continue
			}
			seen[path] = true
			graph.InspectedFiles[path] = moduleRecoveryInspectedFileSnapshot{}
			if len(seen) > graph.Limits.Rows {
				graph.markIncomplete("attachment_file_inspection_resource_limit", reference.Table, "文件引用超过上限，检查未完成。")
				break
			}
			root := cfg.StorageRoots[reference.Root]
			fileKey := reference.Key
			if graph.Side == "package" {
				report.Files.PackageReferencedFiles++
				root, fileKey = verifiedDirectory, path
			} else {
				report.Files.TargetReferencedFiles++
			}
			if graph.Side == "package" && !inManifest {
				report.Complete = false
				report.recordIssue("attachment_missing_from_package_manifest", "blocking", graph.Side, reference.Table, reference.RowIdentity, "模块引用的文件未登记在包内清单中。")
				continue
			}
			maximum := cfg.MaxExpandedBytes
			if inManifest {
				maximum = expected.Size
			}
			actual, err := inspectBoundedModuleRecoveryRegularFileIdentity(ctx, root, fileKey, maximum)
			if err != nil {
				code := "target_attachment_file_unavailable"
				if graph.Side == "package" {
					code = "verified_package_attachment_file_unavailable"
					report.Complete = false
				}
				if errors.Is(err, os.ErrNotExist) {
					graph.InspectedFiles[path] = moduleRecoveryInspectedFileSnapshot{Missing: true}
					if graph.Side == "target" {
						report.Files.TargetMissingFiles++
						code = "target_attachment_file_missing"
					}
				}
				if errors.Is(err, errModuleRecoveryAttachmentExceedsExpectedSize) && inManifest {
					graph.InspectedFiles[path] = moduleRecoveryInspectedFileSnapshot{ContentConflict: true}
					if graph.Side == "target" {
						code = "target_attachment_file_content_conflict"
						report.Files.TargetDifferentFiles++
					} else {
						code = "verified_package_attachment_changed_after_validation"
					}
				}
				report.recordIssue(code, "blocking", graph.Side, reference.Table, reference.RowIdentity, "附件文件缺失、不可读、超出大小或存在不安全的文件类型，未读取或替换其他路径。")
				continue
			}
			graph.InspectedFiles[path] = moduleRecoveryInspectedFileSnapshot{Identity: actual, Available: true}
			if inManifest && (actual.Size != expected.Size || actual.SHA256 != expected.SHA256) {
				code := "target_attachment_file_content_conflict"
				if graph.Side == "package" {
					code = "verified_package_attachment_changed_after_validation"
					report.Complete = false
				} else {
					report.Files.TargetDifferentFiles++
				}
				report.recordIssue(code, "blocking", graph.Side, reference.Table, reference.RowIdentity, "同一文件路径的字节与包内恢复点不同，不能静默覆盖。")
			} else if graph.Side == "package" {
				report.Files.PackageVerifiedFiles++
			} else if inManifest {
				report.Files.TargetMatchingFiles++
			} else {
				report.Files.TargetOnlyVerifiedFiles++
			}
			if graph.Side == "target" && (reference.Digest != "" && reference.Digest != actual.SHA256 || (reference.SizeKnown || reference.Size > 0) && reference.Size != actual.Size) {
				report.recordIssue("target_attachment_registration_content_conflict", "blocking", graph.Side, reference.Table, reference.RowIdentity, "目标媒体登记与其实际文件大小或摘要不一致。")
			}
		}
	}
	return nil
}

var errModuleRecoveryAttachmentExceedsExpectedSize = errors.New("attachment exceeds expected size")

// Root confines traversal even if a directory is renamed while observing files.
// Each path component must also be free of symlinks, and reads are byte bounded.
func inspectBoundedModuleRecoveryRegularFileIdentity(ctx context.Context, directory, key string, maximum int64) (FileIdentity, error) {
	if !safePackagePath("media/uploads/"+key) || maximum < 0 {
		return FileIdentity{}, errors.New("unsafe attachment identity request")
	}
	rootInfo, err := os.Lstat(directory)
	if err != nil {
		return FileIdentity{}, err
	}
	if !rootInfo.IsDir() || rootInfo.Mode()&os.ModeSymlink != 0 {
		return FileIdentity{}, errors.New("unsafe attachment root")
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		return FileIdentity{}, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(root.Close)
	var componentPath string
	for _, component := range strings.Split(key, "/") {
		componentPath = filepath.Join(componentPath, component)
		info, err := root.Lstat(componentPath)
		if err != nil {
			return FileIdentity{}, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return FileIdentity{}, errors.New("attachment symlinks are refused")
		}
		if componentPath == filepath.FromSlash(key) && !info.Mode().IsRegular() {
			return FileIdentity{}, errors.New("attachment must be a regular file")
		}
	}
	file, err := root.Open(filepath.FromSlash(key))
	if err != nil {
		return FileIdentity{}, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(file.Close)
	before, err := file.Stat()
	if err != nil {
		return FileIdentity{}, err
	}
	if !before.Mode().IsRegular() {
		return FileIdentity{}, errors.New("unsafe attachment file type")
	}
	if before.Size() > maximum {
		return FileIdentity{}, errModuleRecoveryAttachmentExceedsExpectedSize
	}
	digest := sha256.New()
	buffer := make([]byte, 64<<10)
	var size int64
	for {
		if err := ctx.Err(); err != nil {
			return FileIdentity{}, err
		}
		count, readErr := file.Read(buffer)
		size += int64(count)
		if size > maximum {
			return FileIdentity{}, errModuleRecoveryAttachmentExceedsExpectedSize
		}
		if _, err := digest.Write(buffer[:count]); err != nil {
			return FileIdentity{}, err
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return FileIdentity{}, readErr
		}
	}
	after, err := root.Lstat(filepath.FromSlash(key))
	if err != nil {
		return FileIdentity{}, err
	}
	if !after.Mode().IsRegular() || !os.SameFile(before, after) || before.Size() != size || after.Size() != size || !before.ModTime().Equal(after.ModTime()) {
		return FileIdentity{}, errors.New("attachment changed while inspecting")
	}
	return FileIdentity{Size: size, SHA256: hex.EncodeToString(digest.Sum(nil))}, nil
}
