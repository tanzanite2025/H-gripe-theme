package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"path"
	"sort"
	"strings"

	"github.com/lib/pq"

	"commerce-platform/internal/pkg/safehtml"
)

type selectedBusinessModuleExternalReference struct {
	Table     string            `json:"table"`
	Values    map[string]string `json:"values"`
	RowSHA256 string            `json:"row_sha256"`
}

type selectedBusinessModuleCompleteSnapshot struct {
	ContractVersion      int                                       `json:"contract_version"`
	Modules              []string                                  `json:"modules"`
	Tables               map[string]productCatalogTableSnapshot    `json:"tables"`
	ExternalDependencies []selectedBusinessModuleExternalReference `json:"external_dependencies"`
	RequiredFilePaths    []string                                  `json:"required_file_paths"`
}

func readSelectedBusinessModuleCompleteTableSnapshot(ctx context.Context, tx *sql.Tx, table, predicate string, arguments ...any) (productCatalogTableSnapshot, error) {
	result := productCatalogTableSnapshot{Rows: []json.RawMessage{}}
	var err error
	result.SchemaSHA256, _, err = inspectProductCatalogTableSchema(ctx, tx, table)
	if err != nil {
		return result, err
	}
	result.ConstraintsSHA256, err = captureProductCatalogConstraintFingerprint(ctx, tx, table)
	if err != nil {
		return result, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT to_jsonb(r)::text FROM public.`+pq.QuoteIdentifier(table)+` r WHERE `+predicate+` ORDER BY to_jsonb(r)::text LIMIT 100001`, arguments...) // #nosec G202 -- Selectors are internal registry literals; data values are bound.
	if err != nil {
		return result, fmt.Errorf("无法保存模块完整记录：%s", table)
	}
	defer func(closeResource func() error) { _ = closeResource() }(rows.Close)
	totalBytes := 0
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return result, err
		}
		totalBytes += len(raw)
		if len(result.Rows) >= productCatalogMaximumRows || totalBytes > productCatalogMaximumJSONBytes {
			return result, errors.New("模块记录超过容量限制，未生成截断备份")
		}
		result.Rows = append(result.Rows, json.RawMessage(raw))
	}
	return result, rows.Err()
}

func (cfg Configuration) captureSelectedBusinessModuleCompleteSnapshot(ctx context.Context, tx *sql.Tx, modules []string) (selectedBusinessModuleCompleteSnapshot, error) {
	snapshot := selectedBusinessModuleCompleteSnapshot{ContractVersion: 1, Modules: modules, Tables: map[string]productCatalogTableSnapshot{}, RequiredFilePaths: []string{}}
	if err := requireRegisteredStorefrontTablesForProductCatalogModule(ctx, tx); err != nil {
		return snapshot, err
	}
	selectors := selectedBusinessModuleTableSelectors(modules)
	for _, module := range modules {
		if module == "products" {
			productSnapshot, err := cfg.captureProductCatalogModuleSnapshot(ctx, tx)
			if err != nil {
				return snapshot, err
			}
			for name, table := range productSnapshot.Tables {
				snapshot.Tables[name] = table
			}
			for reference := range selectors {
				if productCatalogAllowedTableNames()[reference] {
					delete(selectors, reference)
				}
			}
		}
	}
	// Explicit media choice includes all assets, overriding product-only assets.
	for _, module := range modules {
		if module == "media" {
			selectors["media_assets"], selectors["media_asset_derivatives"] = "true", "true"
		}
	}
	for table, predicate := range selectors {
		contents, err := readSelectedBusinessModuleCompleteTableSnapshot(ctx, tx, table, predicate)
		if err != nil {
			return snapshot, err
		}
		snapshot.Tables[table] = contents
		if err := validateSelectedBusinessModuleSnapshotAggregateCapacity(snapshot); err != nil {
			return snapshot, err
		}
	}
	if err := cfg.includeSelectedBusinessModuleReferencedMediaRows(ctx, tx, &snapshot); err != nil {
		return snapshot, err
	}
	paths, err := cfg.collectSelectedBusinessModuleRequiredFilePaths(snapshot)
	if err != nil {
		return snapshot, err
	}
	for path := range paths {
		snapshot.RequiredFilePaths = append(snapshot.RequiredFilePaths, path)
	}
	sort.Strings(snapshot.RequiredFilePaths)
	snapshot.ExternalDependencies, err = captureSelectedBusinessModuleExternalReferenceFingerprints(ctx, tx, snapshot)
	return snapshot, err
}

func validateSelectedBusinessModuleSnapshotAggregateCapacity(snapshot selectedBusinessModuleCompleteSnapshot) error {
	rows, bytes := 0, 0
	for _, table := range snapshot.Tables {
		rows += len(table.Rows)
		for _, row := range table.Rows {
			bytes += len(row)
		}
	}
	if rows > productCatalogMaximumRows || bytes > productCatalogMaximumJSONBytes {
		return errors.New("模块组合记录超过容量限制，请分别导出或使用整站备份")
	}
	return nil
}

func (cfg Configuration) includeSelectedBusinessModuleReferencedMediaRows(ctx context.Context, tx *sql.Tx, snapshot *selectedBusinessModuleCompleteSnapshot) error {
	assetIDs, objectKeys := map[string]bool{}, map[string]bool{}
	for _, row := range snapshot.Tables["media_assets"].Rows {
		id, _ := productCatalogRowString(row, "id")
		assetIDs[id] = true
	}
	// Asset columns are registered FK/business bindings, never arbitrary IDs.
	for _, contents := range snapshot.Tables {
		for _, row := range contents.Rows {
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(row, &fields); err != nil {
				return err
			}
			for _, column := range []string{"media_asset_id", "avatar_media_asset_id"} {
				if _, exists := fields[column]; exists {
					id, _ := productCatalogRowString(row, column)
					if id != "" && id != "0" {
						assetIDs[id] = true
					}
				}
			}
		}
	}
	paths, err := cfg.collectSelectedBusinessModuleRequiredFilePaths(*snapshot)
	if err != nil {
		return err
	}
	for path := range paths {
		for _, root := range mediaRoots {
			if strings.HasPrefix(path, "media/"+root+"/") {
				objectKeys[strings.TrimPrefix(path, "media/"+root+"/")] = true
			}
		}
	}
	ids, keys := []string{}, []string{}
	for id := range assetIDs {
		ids = append(ids, id)
	}
	for key := range objectKeys {
		keys = append(keys, key)
	}
	// URL-only legacy references also carry their original media registration.
	for _, query := range []struct{ table, predicate string }{
		{"media_assets", `r.id::text=ANY($1) OR r.storage_key=ANY($2) OR r.id IN (SELECT media_asset_id FROM media_asset_derivatives WHERE storage_key=ANY($2))`},
		{"media_asset_derivatives", `r.media_asset_id::text=ANY($1) OR r.media_asset_id IN (SELECT id FROM media_assets WHERE storage_key=ANY($2)) OR r.media_asset_id IN (SELECT media_asset_id FROM media_asset_derivatives WHERE storage_key=ANY($2))`},
	} {
		contents, err := readSelectedBusinessModuleCompleteTableSnapshot(ctx, tx, query.table, query.predicate, pq.Array(ids), pq.Array(keys))
		if err != nil {
			return err
		}
		// The supplied IDs already contain every explicitly selected asset.
		snapshot.Tables[query.table] = contents
		if err := validateSelectedBusinessModuleSnapshotAggregateCapacity(*snapshot); err != nil {
			return err
		}
	}
	return nil
}

func (cfg Configuration) collectSelectedBusinessModuleRequiredFilePaths(snapshot selectedBusinessModuleCompleteSnapshot) (map[string]bool, error) {
	paths := map[string]bool{}
	includeReference := func(reference string, required bool) error {
		reference = strings.TrimSpace(html.UnescapeString(reference))
		if reference == "" || reference == "null" {
			return nil
		}
		if strings.HasPrefix(reference, "data:") || strings.Contains(reference, "{{") || strings.Contains(reference, "{%") {
			return nil
		}
		packagePath := cfg.wholeBackupBusinessReferencePackagePath(reference)
		// Company navigation routes and image files share a URL prefix. A page
		// link is not a static file; explicit img/video sources still are.
		if !required && strings.HasPrefix(packagePath, companyStaticPackagePrefix) && path.Ext(packagePath) == "" {
			return nil
		}
		if packagePath == "" {
			if required || strings.Contains(reference, "/uploads/") || strings.Contains(reference, "/uploads-private/") || strings.Contains(reference, "/site-logo-uploads/") {
				return errors.New("模块中存在无法定位到本地实际文件的图片或附件，未生成缺文件备份")
			}
			return nil
		}
		paths[packagePath] = true
		return nil
	}
	var inspectValue func(any, bool) error
	inspectValue = func(value any, required bool) error {
		switch typed := value.(type) {
		case map[string]any:
			for key, child := range typed {
				mediaField := (required && (key == "url" || key == "path" || key == "storage_url")) || key == "thumbnail_url" || key == "poster_url" || key == "storage_url" || key == "image_url" || key == "avatar_url" || key == "avatar" || key == "featured_image" || key == "featured_img" || key == "swatch_url" || key == "file_path"
				// Payment, webhook and provider URL fields are not media.
				if err := inspectValue(child, mediaField); err != nil {
					return err
				}
			}
		case []any:
			for _, child := range typed {
				if err := inspectValue(child, required); err != nil {
					return err
				}
			}
		case string:
			trimmed := strings.TrimSpace(typed)
			if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
				var nested any
				if json.Unmarshal([]byte(trimmed), &nested) == nil {
					return inspectValue(nested, required)
				}
			}
			if required && trimmed != "" {
				if err := includeReference(typed, true); err != nil {
					return err
				}
			}
			if strings.Contains(typed, "<") {
				document, err := safehtml.ParseRichTextMediaSourceAndFileLinkAttributes(typed)
				if err != nil {
					return err
				}
				for _, reference := range document.References {
					if cfg.productCatalogRichTextAttributeRequiresLocalFile(reference) {
						if err := includeReference(*reference.URL, reference.ElementAndAttribute != "a.href"); err != nil {
							return err
						}
					}
				}
			}
			for _, match := range productCatalogCSSMediaSource.FindAllStringSubmatch(typed, -1) {
				if err := includeReference(match[1], true); err != nil {
					return err
				}
			}
			for _, match := range productCatalogEmbeddedMediaReference.FindAllString(typed, -1) {
				if err := includeReference(match, false); err != nil {
					return err
				}
			}
		}
		return nil
	}
	assets := map[string]string{}
	for _, row := range snapshot.Tables["media_assets"].Rows {
		id, _ := productCatalogRowString(row, "id")
		visibility, _ := productCatalogRowString(row, "visibility")
		assets[id] = visibility
	}
	productTables := productCatalogModuleSnapshot{Tables: map[string]productCatalogTableSnapshot{}}
	productOwned := map[string]bool{}
	for _, name := range productCatalogOwnedTableNames {
		productOwned[name] = true
		if table, exists := snapshot.Tables[name]; exists {
			productTables.Tables[name] = table
		}
	}
	productReferences, err := cfg.collectProductCatalogMediaReferences(productTables)
	if err != nil {
		return nil, err
	}
	for reference := range productReferences {
		if err := includeReference(reference, true); err != nil {
			return nil, err
		}
	}
	for table, contents := range snapshot.Tables {
		if productOwned[table] {
			continue
		}
		for _, row := range contents.Rows {
			var fields map[string]any
			if err := json.Unmarshal(row, &fields); err != nil {
				return nil, err
			}
			if raw, exists := fields["attachments"]; exists && raw != nil {
				if text, isText := raw.(string); isText && strings.TrimSpace(text) != "" {
					var decoded any
					if err := json.Unmarshal([]byte(text), &decoded); err != nil {
						return nil, fmt.Errorf("历史附件 JSON 无法解析：%s；请修复后重新备份", table)
					}
					raw = decoded
				}
				if raw != nil && raw != "" {
					if _, array := raw.([]any); !array {
						return nil, fmt.Errorf("附件字段必须为数组：%s", table)
					}
					if err := inspectValue(raw, true); err != nil {
						return nil, err
					}
				}
				delete(fields, "attachments")
			}
			if key, present := fields["storage_key"].(string); present && key != "" {
				visibility, _ := fields["visibility"].(string)
				root := registeredMediaStorageRootForObjectKeyAndVisibility(key, visibility)
				switch table {
				case "media_asset_derivatives":
					id, _ := productCatalogRowString(row, "media_asset_id")
					root = registeredMediaStorageRootForObjectKeyAndVisibility(key, assets[id])
				case "order_evidence_attachments":
					root = "uploads_private"
				case "site_logo_assets":
					root = "site_logo_uploads"
				case "media_assets", "site_favicon_assets", "visual_showcase_items":
				default:
					return nil, fmt.Errorf("未登记的模块文件存储字段：%s", table)
				}
				path := "media/" + root + "/" + key
				if !safePackagePath(path) {
					return nil, errors.New("模块附件路径越界")
				}
				paths[path] = true
				delete(fields, "url") // Authoritative storage key also covers private API URLs.
			}
			if value, exists := fields["url"].(string); exists {
				switch table {
				case "media", "media_assets", "media_asset_derivatives", "product_media", "gallery_images":
					if err := includeReference(value, true); err != nil {
						return nil, err
					}
				}
			}
			if err := inspectValue(fields, false); err != nil {
				return nil, fmt.Errorf("%s: %w", table, err)
			}
		}
	}
	return paths, nil
}

func captureSelectedBusinessModuleExternalReferenceFingerprints(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot) ([]selectedBusinessModuleExternalReference, error) {
	observation, err := inspectModuleRecoveryDatabaseSchemaWithoutRowFingerprints(ctx, tx)
	if err != nil {
		return nil, err
	}
	return captureSelectedBusinessModuleExternalReferenceFingerprintsWithObservedSchema(ctx, tx, snapshot, observation)
}

func captureSelectedBusinessModuleExternalReferenceFingerprintsWithObservedSchema(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, observation *moduleRecoveryDatabaseObservation) ([]selectedBusinessModuleExternalReference, error) {
	references := append([]ModuleRecoveryDependencyReference(nil), observation.References...)
	references = append(references, registeredNonForeignKeyModuleRecoveryReferences()...)
	references = append(references, registeredOrderTransactionBusinessReferences()...)
	dependencies := map[string]selectedBusinessModuleExternalReference{}
	selectedReferenceKeys := map[string]map[string]bool{}
	for _, reference := range references {
		for _, row := range snapshot.Tables[reference.SourceTable].Rows {
			if len(reference.SourceColumns) != len(reference.TargetColumns) || len(reference.SourceColumns) == 0 {
				return nil, errors.New("不支持的模块复合依赖")
			}
			values, predicates, arguments := map[string]string{}, []string{}, []any{}
			valid := true
			for index, column := range reference.SourceColumns {
				value, valueErr := productCatalogRowString(row, column)
				if valueErr != nil || value == "" || value == "0" || !observation.Columns[reference.TargetTable][reference.TargetColumns[index]] {
					valid = false
					break
				}
				values[reference.TargetColumns[index]] = value
				arguments = append(arguments, value)
				predicates = append(predicates, pq.QuoteIdentifier(reference.TargetColumns[index])+fmt.Sprintf("::text=$%d", index+1))
			}
			if !valid {
				continue
			}
			// A row already carried in the selected set needs no external record.
			indexName := reference.TargetTable + "/" + strings.Join(reference.TargetColumns, ",")
			index, indexed := selectedReferenceKeys[indexName]
			if !indexed {
				index = map[string]bool{}
				for _, candidate := range snapshot.Tables[reference.TargetTable].Rows {
					candidateValues := map[string]string{}
					for _, column := range reference.TargetColumns {
						candidateValues[column], _ = productCatalogRowString(candidate, column)
					}
					encoded, _ := json.Marshal(candidateValues)
					index[string(encoded)] = true
				}
				selectedReferenceKeys[indexName] = index
			}
			encoded, _ := json.Marshal(values)
			if index[string(encoded)] {
				continue
			}
			identity := reference.TargetTable + "/" + string(encoded)
			if _, exists := dependencies[identity]; exists {
				continue
			}
			var raw string
			err := tx.QueryRowContext(ctx, `SELECT to_jsonb(r)::text FROM public.`+pq.QuoteIdentifier(reference.TargetTable)+` r WHERE `+strings.Join(predicates, " AND "), arguments...).Scan(&raw) // #nosec G202 -- Reference identifiers originate in catalog metadata or explicit registry; values are bound.
			if err != nil && err != sql.ErrNoRows {
				return nil, err
			}
			digest := "missing"
			if err == nil {
				digest = productCatalogBytesSHA256([]byte(raw))
			}
			dependencies[identity] = selectedBusinessModuleExternalReference{reference.TargetTable, values, digest}
		}
	}
	identities := []string{}
	for identity := range dependencies {
		identities = append(identities, identity)
	}
	sort.Strings(identities)
	result := []selectedBusinessModuleExternalReference{}
	for _, identity := range identities {
		result = append(result, dependencies[identity])
	}
	return result, nil
}
