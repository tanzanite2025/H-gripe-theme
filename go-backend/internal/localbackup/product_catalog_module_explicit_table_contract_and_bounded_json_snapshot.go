package localbackup

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/lib/pq"
)

const productCatalogBackupScope = "product_catalog_with_images"
const productCatalogRowsFilename = "product-catalog-rows.json"
const productCatalogMaximumJSONBytes = 64 << 20
const productCatalogMaximumRows = 100000

// Explicit ownership is intentional: reviews, carts, orders, customer identities,
// and feeds are separate modules even when they reference a product.
var productCatalogOwnedTableNames = strings.Fields(`products product_variants product_media
product_categories product_category_translations product_brands product_attributes product_attribute_values
product_specification_templates product_spec_definitions product_spec_values product_spec_option_items
product_variant_option_values product_option_value_relations product_option_group_variant_rules
product_option_value_variant_rules product_custom_option_policies product_information_templates
product_shipping_template_bindings product_quality_requirement_rules product_spoke_repair_kit_models
product_display_price_snapshots product_procurement_records product_profit_calculations`)

type productCatalogTableSnapshot struct {
	SchemaSHA256      string            `json:"schema_sha256"`
	ConstraintsSHA256 string            `json:"constraints_sha256"`
	Rows              []json.RawMessage `json:"rows"`
}

type productCatalogExternalDependency struct {
	Table     string `json:"table"`
	Column    string `json:"column"`
	Value     string `json:"value"`
	RowSHA256 string `json:"row_sha256"`
}

type productCatalogModuleSnapshot struct {
	ContractVersion      int                                    `json:"contract_version"`
	Tables               map[string]productCatalogTableSnapshot `json:"tables"`
	ExternalDependencies []productCatalogExternalDependency     `json:"external_dependencies"`
}

type productCatalogForeignKey struct{ Table, Column, ParentTable, ParentColumn string }

type ProductCatalogIsolatedRecoveryResult struct {
	RestoredRows             int                                        `json:"restored_rows"`
	VerifiedImageFiles       int                                        `json:"verified_image_files"`
	PreservedTables          int                                        `json:"preserved_tables"`
	PreservedTargetOnlyRows  bool                                       `json:"preserved_target_only_rows"`
	OnlineOverwriteAvailable bool                                       `json:"online_overwrite_available"`
	InventoryPreservation    *ProductCatalogInventoryPreservationResult `json:"inventory_preservation,omitempty"`
}

func isAllowedBackupFileForScope(scope, filename string) bool {
	if scope == "whole_application" {
		return safePackagePath(filename) || isAllowedCompanyStaticPackagePath(filename)
	}
	if scope != productCatalogBackupScope {
		return false
	}
	return filename == productCatalogRowsFilename || (strings.HasPrefix(filename, "media/uploads/") && safePackagePath(filename)) ||
		(strings.HasPrefix(filename, "product-static/company/") && safePackagePath("media/uploads/"+strings.TrimPrefix(filename, "product-static/")))
}

func productCatalogAllowedTableNames() map[string]bool {
	allowed := map[string]bool{"media_assets": true, "media_asset_derivatives": true}
	for _, name := range productCatalogOwnedTableNames {
		allowed[name] = true
	}
	return allowed
}

func productCatalogBytesSHA256(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func inspectProductCatalogTableSchema(ctx context.Context, tx *sql.Tx, table string) (string, []string, error) {
	rows, err := tx.QueryContext(ctx, `SELECT a.attname,format_type(a.atttypid,a.atttypmod),a.attnotnull,a.attidentity,a.attgenerated,COALESCE(pg_get_expr(d.adbin,d.adrelid),'') FROM pg_attribute a LEFT JOIN pg_attrdef d ON d.adrelid=a.attrelid AND d.adnum=a.attnum WHERE a.attrelid=to_regclass($1) AND a.attnum>0 AND NOT a.attisdropped ORDER BY a.attnum`, "public."+table)
	if err != nil {
		return "", nil, err
	}
	var columns []string
	var contract strings.Builder
	for rows.Next() {
		var name, typ, identity, generated, defaultValue string
		var required bool
		if err := rows.Scan(&name, &typ, &required, &identity, &generated, &defaultValue); err != nil {
			_ = rows.Close()
			return "", nil, err
		}
		if generated != "" || identity != "" {
			_ = rows.Close()
			return "", nil, errors.New("product module requires an explicit generated-column contract")
		}
		columns = append(columns, name)
		fmt.Fprintf(&contract, "%q:%q:%t:%q\n", name, typ, required, defaultValue)
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return "", nil, err
	}
	if len(columns) == 0 {
		return "", nil, fmt.Errorf("product table missing: %s", table)
	}
	return productCatalogBytesSHA256([]byte(contract.String())), columns, nil
}

func captureProductCatalogConstraintFingerprint(ctx context.Context, tx *sql.Tx, table string) (string, error) {
	var constraints string
	err := tx.QueryRowContext(ctx, `SELECT COALESCE(string_agg(pg_get_constraintdef(oid),E'\n' ORDER BY conname),'') FROM pg_constraint WHERE conrelid=to_regclass($1)`, "public."+table).Scan(&constraints)
	return productCatalogBytesSHA256([]byte(constraints)), err
}

func readProductCatalogRows(ctx context.Context, tx *sql.Tx, table, predicate string, arguments ...any) ([]json.RawMessage, error) {
	rows, err := tx.QueryContext(ctx, `SELECT to_jsonb(r)::text FROM public.`+pq.QuoteIdentifier(table)+` r WHERE `+predicate+` ORDER BY r.id LIMIT 100001`, arguments...) // #nosec G202 -- The predicate comes from internal fixed recovery selectors; table identifiers are quoted and values are bound.
	if err != nil {
		return nil, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(rows.Close)
	result := []json.RawMessage{}
	total := 0
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		total += len(raw)
		if len(result) >= productCatalogMaximumRows || total > productCatalogMaximumJSONBytes {
			return nil, errors.New("product module row or JSON capacity exceeded")
		}
		result = append(result, json.RawMessage(raw))
	}
	return result, rows.Err()
}

func inspectProductCatalogForeignKeys(ctx context.Context, tx *sql.Tx) ([]productCatalogForeignKey, error) {
	rows, err := tx.QueryContext(ctx, `SELECT t.relname,a.attname,p.relname,b.attname,cardinality(c.conkey),pn.nspname FROM pg_constraint c JOIN pg_class t ON t.oid=c.conrelid JOIN pg_namespace tn ON tn.oid=t.relnamespace JOIN pg_class p ON p.oid=c.confrelid JOIN pg_namespace pn ON pn.oid=p.relnamespace JOIN pg_attribute a ON a.attrelid=t.oid AND a.attnum=c.conkey[1] JOIN pg_attribute b ON b.attrelid=p.oid AND b.attnum=c.confkey[1] WHERE c.contype='f' AND tn.nspname='public' ORDER BY t.relname,a.attname`)
	if err != nil {
		return nil, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(rows.Close)
	var keys []productCatalogForeignKey
	for rows.Next() {
		var key productCatalogForeignKey
		var count int
		var namespace string
		if err := rows.Scan(&key.Table, &key.Column, &key.ParentTable, &key.ParentColumn, &count, &namespace); err != nil {
			return nil, err
		}
		if productCatalogAllowedTableNames()[key.Table] && (count != 1 || namespace != "public") {
			return nil, errors.New("unregistered product composite or cross-schema dependency")
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

func productCatalogRowString(row json.RawMessage, column string) (string, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(row, &fields); err != nil {
		return "", err
	}
	raw, exists := fields[column]
	if !exists {
		return "", fmt.Errorf("product row is missing column %s", column)
	}
	if string(raw) == "null" {
		return "", nil
	}
	var value string
	if len(raw) > 0 && raw[0] == '"' {
		if err := json.Unmarshal(raw, &value); err != nil {
			return "", err
		}
		return value, nil
	}
	return string(raw), nil
}

func captureProductCatalogExternalDependencies(ctx context.Context, tx *sql.Tx, snapshot *productCatalogModuleSnapshot) ([]productCatalogExternalDependency, error) {
	keys, err := inspectProductCatalogForeignKeys(ctx, tx)
	if err != nil {
		return nil, err
	}
	dependencies := map[string]productCatalogExternalDependency{}
	selectedParentValues := map[string]map[string]bool{}
	for _, key := range keys {
		table, selected := snapshot.Tables[key.Table]
		if !selected {
			continue
		}
		for _, row := range table.Rows {
			value, err := productCatalogRowString(row, key.Column)
			if err != nil {
				return nil, err
			}
			if value == "" {
				continue
			}
			if parent, selected := snapshot.Tables[key.ParentTable]; selected {
				parentIdentity := key.ParentTable + "/" + key.ParentColumn
				values, exists := selectedParentValues[parentIdentity]
				if !exists {
					values = map[string]bool{}
					for _, candidate := range parent.Rows {
						parentValue, err := productCatalogRowString(candidate, key.ParentColumn)
						if err != nil {
							return nil, err
						}
						values[parentValue] = true
					}
					selectedParentValues[parentIdentity] = values
				}
				if !values[value] {
					return nil, fmt.Errorf("product package dependency missing: %s.%s", key.Table, key.Column)
				}
				continue
			}
			identity := key.ParentTable + "/" + key.ParentColumn + "/" + value
			if _, exists := dependencies[identity]; exists {
				continue
			}
			var raw string
			err = tx.QueryRowContext(ctx, `SELECT to_jsonb(r)::text FROM public.`+pq.QuoteIdentifier(key.ParentTable)+` r WHERE `+pq.QuoteIdentifier(key.ParentColumn)+`::text=$1`, value).Scan(&raw)
			if err != nil {
				return nil, fmt.Errorf("shared product dependency unavailable: %s.%s", key.Table, key.Column)
			}
			dependencies[identity] = productCatalogExternalDependency{key.ParentTable, key.ParentColumn, value, productCatalogBytesSHA256([]byte(raw))}
		}
	}
	identities := make([]string, 0, len(dependencies))
	for identity := range dependencies {
		identities = append(identities, identity)
	}
	sort.Strings(identities)
	result := []productCatalogExternalDependency{}
	for _, identity := range identities {
		result = append(result, dependencies[identity])
	}
	return result, nil
}

func (cfg Configuration) captureProductCatalogModuleSnapshot(ctx context.Context, tx *sql.Tx) (productCatalogModuleSnapshot, error) {
	snapshot := productCatalogModuleSnapshot{ContractVersion: 1, Tables: map[string]productCatalogTableSnapshot{}}
	if err := requireRegisteredStorefrontTablesForProductCatalogModule(ctx, tx); err != nil {
		return snapshot, err
	}
	if err := validateProductCatalogStandardTriggersAndRejectUnregisteredRewriteRules(ctx, tx); err != nil {
		return snapshot, err
	}
	if _, err := inspectProductCatalogRequiredSystemCategoriesAndCarbonRimBindings(ctx, tx); err != nil {
		return snapshot, err
	}
	totalRows, totalBytes := 0, 0
	for _, name := range productCatalogOwnedTableNames {
		schema, _, err := inspectProductCatalogTableSchema(ctx, tx, name)
		if err != nil {
			return snapshot, err
		}
		rows, err := readProductCatalogRows(ctx, tx, name, "true")
		if err != nil {
			return snapshot, err
		}
		totalRows += len(rows)
		for _, row := range rows {
			totalBytes += len(row)
		}
		if totalRows > productCatalogMaximumRows || totalBytes > productCatalogMaximumJSONBytes {
			return snapshot, errors.New("product snapshot aggregate capacity exceeded")
		}
		constraints, err := captureProductCatalogConstraintFingerprint(ctx, tx, name)
		if err != nil {
			return snapshot, err
		}
		snapshot.Tables[name] = productCatalogTableSnapshot{SchemaSHA256: schema, ConstraintsSHA256: constraints, Rows: rows}
	}
	keys, err := inspectProductCatalogForeignKeys(ctx, tx)
	if err != nil {
		return snapshot, err
	}
	assetIDs := map[string]bool{}
	for _, key := range keys {
		if key.ParentTable != "media_assets" {
			continue
		}
		for _, row := range snapshot.Tables[key.Table].Rows {
			value, err := productCatalogRowString(row, key.Column)
			if err != nil {
				return snapshot, err
			}
			if value != "" {
				assetIDs[value] = true
			}
		}
	}
	// URL-only historical product links must pull in their registered asset too.
	references, err := cfg.collectProductCatalogMediaReferences(snapshot)
	if err != nil {
		return snapshot, err
	}
	for reference := range references {
		root, key := cfg.resolveLocalMediaReference(reference)
		if root == "uploads" && key != "" {
			rows, err := tx.QueryContext(ctx, `SELECT id::text FROM media_assets WHERE storage_key=$1 OR url=$2 UNION SELECT media_asset_id::text FROM media_asset_derivatives WHERE storage_key=$1 OR url=$2`, key, reference)
			if err != nil {
				return snapshot, err
			}
			for rows.Next() {
				var id string
				if err := rows.Scan(&id); err != nil {
					_ = rows.Close()
					return snapshot, err
				}
				assetIDs[id] = true
			}
			err = rows.Err()
			_ = rows.Close()
			if err != nil {
				return snapshot, err
			}
		}
	}
	ids := []string{}
	for id := range assetIDs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, name := range []string{"media_assets", "media_asset_derivatives"} {
		schema, _, err := inspectProductCatalogTableSchema(ctx, tx, name)
		if err != nil {
			return snapshot, err
		}
		column := "id"
		if name == "media_asset_derivatives" {
			column = "media_asset_id"
		}
		rows, err := readProductCatalogRows(ctx, tx, name, pq.QuoteIdentifier(column)+"::text=ANY($1)", pq.Array(ids))
		if err != nil {
			return snapshot, err
		}
		totalRows += len(rows)
		for _, row := range rows {
			totalBytes += len(row)
		}
		if totalRows > productCatalogMaximumRows || totalBytes > productCatalogMaximumJSONBytes {
			return snapshot, errors.New("product snapshot aggregate capacity exceeded")
		}
		constraints, err := captureProductCatalogConstraintFingerprint(ctx, tx, name)
		if err != nil {
			return snapshot, err
		}
		snapshot.Tables[name] = productCatalogTableSnapshot{SchemaSHA256: schema, ConstraintsSHA256: constraints, Rows: rows}
	}
	for _, row := range snapshot.Tables["media_assets"].Rows {
		visibility, err := productCatalogRowString(row, "visibility")
		if err != nil {
			return snapshot, err
		}
		partition, err := productCatalogRowString(row, "library_partition")
		if err != nil {
			return snapshot, err
		}
		if visibility != "public" || partition == "customer_service" {
			return snapshot, errors.New("product package refuses private or customer-service media")
		}
	}
	snapshot.ExternalDependencies, err = captureProductCatalogExternalDependencies(ctx, tx, &snapshot)
	if err == nil {
		err = cfg.validateProductCatalogCompleteMediaPlacementOwnershipAndAssetBindings(snapshot)
	}
	if err == nil {
		err = cfg.validateProductCatalogCompleteRichTextRegisteredMediaTypes(snapshot)
	}
	return snapshot, err
}

func requireRegisteredStorefrontTablesForProductCatalogModule(ctx context.Context, tx *sql.Tx) error {
	registered := registeredStorefrontModuleRecoveryTableNames()
	rows, err := tx.QueryContext(ctx, `SELECT tablename FROM pg_tables WHERE schemaname='public' ORDER BY tablename`)
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(rows.Close)
	count := 0
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		if !registered[name] {
			return fmt.Errorf("unregistered storefront table blocks product module backup or restore: %s", name)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if count != len(registered) {
		return errors.New("product module requires the complete registered storefront schema")
	}
	return nil
}
