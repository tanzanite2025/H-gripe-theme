package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func verifyProductCatalogStandardTriggerContractsBeforeAnyWriteAndSystemCategoryAtomicRollback(t *testing.T, manager *Manager, stage *sql.DB, snapshot productCatalogModuleSnapshot, packageData []byte) {
	t.Helper()
	ctx, cfg := context.Background(), manager.Configuration
	const categoryTrigger = "protect_carbon_rim_system_category_identity_and_tree_position"
	const productTrigger = "require_carbon_rim_product_type_and_system_category_binding"
	originalFunctions, originalTriggers := []string{}, []string{}
	for _, contract := range productCatalogApprovedStandardTriggerContracts() {
		var function, trigger string
		require.NoError(t, stage.QueryRowContext(context.Background(), `SELECT pg_get_functiondef(tgfoid),pg_get_triggerdef(oid) FROM pg_trigger WHERE tgrelid=$1::regclass AND tgname=$2`, "public."+contract.Table, contract.Name).Scan(&function, &trigger))
		originalFunctions = append(originalFunctions, function)
		originalTriggers = append(originalTriggers, trigger)
	}
	for _, scenario := range []struct{ name, setupSQL string }{
		{"missing_category_trigger", `DROP TRIGGER ` + categoryTrigger + ` ON product_categories`},
		{"missing_product_trigger", `DROP TRIGGER ` + productTrigger + ` ON products`},
		{"disabled_category_trigger", `ALTER TABLE product_categories DISABLE TRIGGER ` + categoryTrigger},
		{"disabled_product_trigger", `ALTER TABLE products DISABLE TRIGGER ` + productTrigger},
		{"always_category_trigger", `ALTER TABLE product_categories ENABLE ALWAYS TRIGGER ` + categoryTrigger},
		{"replica_product_trigger", `ALTER TABLE products ENABLE REPLICA TRIGGER ` + productTrigger},
		{"same_name_changed_function_body", `CREATE OR REPLACE FUNCTION ` + productTrigger + `() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN PERFORM nextval('products_id_seq'); RETURN NEW; END $$`},
		{"security_definer_function", `ALTER FUNCTION ` + productTrigger + `() SECURITY DEFINER`},
		{"function_local_configuration", `ALTER FUNCTION ` + categoryTrigger + `() SET search_path=public`},
		{"stable_function", `ALTER FUNCTION ` + productTrigger + `() STABLE`},
		{"parallel_safe_function", `ALTER FUNCTION ` + productTrigger + `() PARALLEL SAFE`},
		{"leakproof_function", `ALTER FUNCTION ` + productTrigger + `() LEAKPROOF`},
		{"different_trigger_function_identity", `CREATE FUNCTION synthetic_product_restore_side_effect_guard() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RETURN NEW; END $$; DROP TRIGGER ` + productTrigger + ` ON products; CREATE TRIGGER ` + productTrigger + ` BEFORE INSERT OR UPDATE OF product_type,product_category_id ON products FOR EACH ROW EXECUTE FUNCTION synthetic_product_restore_side_effect_guard()`},
		{"changed_update_columns", `DROP TRIGGER ` + productTrigger + ` ON products; CREATE TRIGGER ` + productTrigger + ` BEFORE INSERT OR UPDATE OF product_type ON products FOR EACH ROW EXECUTE FUNCTION ` + productTrigger + `()`},
		{"changed_trigger_events", `DROP TRIGGER ` + categoryTrigger + ` ON product_categories; CREATE TRIGGER ` + categoryTrigger + ` BEFORE INSERT OR UPDATE ON product_categories FOR EACH ROW EXECUTE FUNCTION ` + categoryTrigger + `()`},
		{"changed_trigger_timing", `DROP TRIGGER ` + productTrigger + ` ON products; CREATE TRIGGER ` + productTrigger + ` AFTER INSERT OR UPDATE OF product_type,product_category_id ON products FOR EACH ROW EXECUTE FUNCTION ` + productTrigger + `()`},
		{"conditional_trigger", `DROP TRIGGER ` + productTrigger + ` ON products; CREATE TRIGGER ` + productTrigger + ` BEFORE INSERT OR UPDATE OF product_type,product_category_id ON products FOR EACH ROW WHEN (NEW.id>0) EXECUTE FUNCTION ` + productTrigger + `()`},
		{"trigger_arguments", `DROP TRIGGER ` + productTrigger + ` ON products; CREATE TRIGGER ` + productTrigger + ` BEFORE INSERT OR UPDATE OF product_type,product_category_id ON products FOR EACH ROW EXECUTE FUNCTION ` + productTrigger + `('unexpected')`},
		{"unregistered_product_always_sequence_side_effect", `CREATE FUNCTION synthetic_product_restore_side_effect_guard() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN PERFORM nextval('products_id_seq'); RETURN NEW; END $$; CREATE TRIGGER synthetic_product_restore_guard BEFORE INSERT OR UPDATE ON products FOR EACH ROW EXECUTE FUNCTION synthetic_product_restore_side_effect_guard(); ALTER TABLE products ENABLE ALWAYS TRIGGER synthetic_product_restore_guard`},
		{"unregistered_media_replica_sequence_side_effect", `CREATE FUNCTION synthetic_product_restore_side_effect_guard() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN PERFORM nextval('products_id_seq'); RETURN NEW; END $$; CREATE TRIGGER synthetic_product_restore_guard BEFORE INSERT OR UPDATE ON media_assets FOR EACH ROW EXECUTE FUNCTION synthetic_product_restore_side_effect_guard(); ALTER TABLE media_assets ENABLE REPLICA TRIGGER synthetic_product_restore_guard`},
		{"disabled_custom_media_trigger", `CREATE FUNCTION synthetic_product_restore_side_effect_guard() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RETURN NEW; END $$; CREATE TRIGGER synthetic_product_restore_guard BEFORE INSERT OR UPDATE ON media_asset_derivatives FOR EACH ROW EXECUTE FUNCTION synthetic_product_restore_side_effect_guard(); ALTER TABLE media_asset_derivatives DISABLE TRIGGER synthetic_product_restore_guard`},
		{"product_rewrite_rule", `CREATE RULE synthetic_product_restore_rule AS ON UPDATE TO products DO ALSO SELECT nextval('products_id_seq')`},
		{"disabled_media_rewrite_rule", `CREATE RULE synthetic_product_restore_rule AS ON UPDATE TO media_assets DO ALSO SELECT nextval('products_id_seq'); ALTER TABLE media_assets DISABLE RULE synthetic_product_restore_rule`},
	} {
		t.Run("product_standard_trigger_"+scenario.name, func(t *testing.T) {
			t.Cleanup(func() {
				for _, table := range []string{"products", "product_categories", "media_assets", "media_asset_derivatives"} {
					_, err := stage.ExecContext(context.Background(), `DROP TRIGGER IF EXISTS synthetic_product_restore_guard ON public.`+pq.QuoteIdentifier(table)+`; DROP RULE IF EXISTS synthetic_product_restore_rule ON public.`+pq.QuoteIdentifier(table))
					require.NoError(t, err)
				}
				for i, contract := range productCatalogApprovedStandardTriggerContracts() {
					_, err := stage.ExecContext(context.Background(), `DROP TRIGGER IF EXISTS `+pq.QuoteIdentifier(contract.Name)+` ON public.`+pq.QuoteIdentifier(contract.Table)+`; `+originalFunctions[i]+`; `+originalTriggers[i])
					require.NoError(t, err)
				}
				_, err := stage.ExecContext(context.Background(), `DROP FUNCTION IF EXISTS synthetic_product_restore_side_effect_guard()`)
				require.NoError(t, err)
			})
			_, err := stage.ExecContext(context.Background(), scenario.setupSQL)
			require.NoError(t, err)
			_, before, _, err := captureDatabaseInventory(ctx, stage)
			require.NoError(t, err)
			_, _, result, err := cfg.applyProductCatalogRowsOnlyToIsolatedDatabase(ctx, stage, snapshot)
			require.Error(t, err)
			require.Contains(t, err.Error(), "trigger", "reject the write contract before merging, rather than incidentally failing later")
			require.Nil(t, result)
			_, after, _, err := captureDatabaseInventory(ctx, stage)
			require.NoError(t, err)
			require.Equal(t, before, after, "rejection must precede ALWAYS/REPLICA side effects, including nontransactional sequence advancement")
			tx, err := stage.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
			require.NoError(t, err)
			_, err = cfg.captureProductCatalogModuleSnapshot(ctx, tx)
			require.Error(t, err, "export must reject the same unsupported contract")
			require.NoError(t, tx.Rollback())
		})
	}

	var rootID, carbonID, originalProductID int64
	require.NoError(t, stage.QueryRowContext(context.Background(), `SELECT id FROM product_categories WHERE slug='wheel-components'`).Scan(&rootID))
	require.NoError(t, stage.QueryRowContext(context.Background(), `SELECT id FROM product_categories WHERE slug='carbon-rim'`).Scan(&carbonID))
	require.NoError(t, stage.QueryRowContext(context.Background(), `SELECT id FROM products WHERE slug='product-backup-original' AND locale='en'`).Scan(&originalProductID))
	for _, scenario := range []struct {
		name, table, slug, mutationJSON, currentSQL, cleanupSQL, expectedError string
	}{
		{name: "disabled_root", table: "product_categories", slug: "wheel-components", mutationJSON: `{"is_enabled":false}`, expectedError: "enabled wheel-components"},
		{name: "disabled_carbon_category", table: "product_categories", slug: "carbon-rim", mutationJSON: `{"is_enabled":false}`, expectedError: "enabled wheel-components"},
		{name: "wrong_root_depth", table: "product_categories", slug: "wheel-components", mutationJSON: `{"depth":2}`, expectedError: "enabled wheel-components"},
		{name: "wrong_root_parent", table: "product_categories", slug: "wheel-components", mutationJSON: fmt.Sprintf(`{"parent_id":%d}`, carbonID), expectedError: "enabled wheel-components"},
		{name: "wrong_carbon_depth", table: "product_categories", slug: "carbon-rim", mutationJSON: `{"depth":3}`, expectedError: "enabled wheel-components"},
		{name: "wrong_carbon_parent", table: "product_categories", slug: "carbon-rim", mutationJSON: `{"parent_id":null}`, expectedError: "enabled wheel-components"},
		{name: "renamed_root_slug", table: "product_categories", slug: "wheel-components", mutationJSON: `{"slug":"renamed-system-root"}`, expectedError: "enabled wheel-components"},
		{name: "renamed_carbon_slug", table: "product_categories", slug: "carbon-rim", mutationJSON: `{"slug":"renamed-system-carbon"}`, expectedError: "enabled wheel-components"},
		{name: "existing_normal_cannot_become_carbon", table: "products", slug: "product-backup-original", mutationJSON: fmt.Sprintf(`{"product_type":"carbon_rim","product_category_id":%d}`, carbonID), expectedError: "cannot change an existing dedicated"},
		{name: "existing_carbon_cannot_become_normal", currentSQL: fmt.Sprintf(`BEGIN; SET LOCAL session_replication_role=replica; UPDATE products SET product_type='carbon_rim',carbon_rim_specification='{"schema_version":1}',product_category_id=%d WHERE id=%d; COMMIT`, carbonID, originalProductID), cleanupSQL: fmt.Sprintf(`BEGIN; SET LOCAL session_replication_role=replica; UPDATE products SET product_type='',carbon_rim_specification=NULL,product_category_id=NULL WHERE id=%d; COMMIT`, originalProductID), expectedError: "cannot change an existing dedicated"},
		{name: "retained_product_wrong_binding", currentSQL: fmt.Sprintf(`BEGIN; SET LOCAL session_replication_role=replica; UPDATE products SET product_category_id=%d WHERE slug='target-only-product'; COMMIT`, carbonID), cleanupSQL: `BEGIN; SET LOCAL session_replication_role=replica; UPDATE products SET product_category_id=NULL WHERE slug='target-only-product'; COMMIT`, expectedError: "system category bindings"},
		{name: "retained_carbon_without_category", currentSQL: `BEGIN; SET LOCAL session_replication_role=replica; UPDATE products SET product_type='carbon_rim',carbon_rim_specification='{"schema_version":1}' WHERE slug='target-only-product'; COMMIT`, cleanupSQL: `BEGIN; SET LOCAL session_replication_role=replica; UPDATE products SET product_type='',carbon_rim_specification=NULL WHERE slug='target-only-product'; COMMIT`, expectedError: "system category bindings"},
		{name: "preexisting_invalid_system_tree", currentSQL: `BEGIN; SET LOCAL session_replication_role=replica; UPDATE product_categories SET is_enabled=false WHERE slug='carbon-rim'; COMMIT`, cleanupSQL: `BEGIN; SET LOCAL session_replication_role=replica; UPDATE product_categories SET is_enabled=true WHERE slug='carbon-rim'; COMMIT`, expectedError: "enabled wheel-components"},
	} {
		t.Run("product_system_category_"+scenario.name, func(t *testing.T) {
			if scenario.cleanupSQL != "" {
				t.Cleanup(func() {
					_, err := stage.ExecContext(context.Background(), scenario.cleanupSQL)
					require.NoError(t, err)
				})
			}
			if scenario.currentSQL != "" {
				_, err := stage.ExecContext(context.Background(), scenario.currentSQL)
				require.NoError(t, err)
			}
			changed := copyProductCatalogSnapshotWithMatchingRowJSONFieldsChanged(t, snapshot, scenario.table, scenario.slug, scenario.mutationJSON)
			_, before, _, err := captureDatabaseInventory(ctx, stage)
			require.NoError(t, err)
			_, _, result, err := cfg.applyProductCatalogRowsOnlyToIsolatedDatabase(ctx, stage, changed)
			require.ErrorContains(t, err, scenario.expectedError)
			require.Nil(t, result)
			_, after, _, err := captureDatabaseInventory(ctx, stage)
			require.NoError(t, err)
			require.Equal(t, before, after, "invalid final business state must roll back merged rows without changing sequences")
		})
	}

	for _, systemSlug := range []string{"wheel-components", "carbon-rim"} {
		t.Run("product_system_category_cannot_reassign_"+systemSlug, func(t *testing.T) {
			changed := copyProductCatalogSnapshotWithMatchingRowJSONFieldsChanged(t, snapshot, "product_categories", systemSlug, `{"slug":"synthetic-retired-system-identity"}`)
			if systemSlug == "wheel-components" {
				changed = copyProductCatalogSnapshotWithMatchingRowJSONFieldsChanged(t, changed, "product_categories", "carbon-rim", `{"parent_id":900004}`)
			}
			var replacement string
			require.NoError(t, stage.QueryRowContext(context.Background(), `SELECT (to_jsonb(c)||'{"id":900004}'::jsonb)::text FROM product_categories c WHERE slug=$1`, systemSlug).Scan(&replacement))
			table := changed.Tables["product_categories"]
			table.Rows = append(table.Rows, json.RawMessage(replacement))
			changed.Tables["product_categories"] = table
			_, before, _, err := captureDatabaseInventory(ctx, stage)
			require.NoError(t, err)
			_, _, _, err = cfg.applyProductCatalogRowsOnlyToIsolatedDatabase(ctx, stage, changed)
			require.ErrorContains(t, err, "cannot reassign protected system category identities")
			_, after, _, err := captureDatabaseInventory(ctx, stage)
			require.NoError(t, err)
			require.Equal(t, before, after)
		})
	}

	for _, validBinding := range []bool{false, true} {
		t.Run(fmt.Sprintf("product_system_category_new_carbon_valid_binding_%t", validBinding), func(t *testing.T) {
			changed := copyProductCatalogSnapshotWithMatchingRowJSONFieldsChanged(t, snapshot, "", "", "")
			var rawRow string
			require.NoError(t, stage.QueryRowContext(context.Background(), `SELECT (to_jsonb(p)||jsonb_build_object('id',900003,'slug','synthetic-new-carbon-product','parent_id',NULL,'product_type','carbon_rim','carbon_rim_specification','{"schema_version":1}'::jsonb,'product_category_id',$1::bigint))::text FROM products p WHERE id=$2`, carbonID, originalProductID).Scan(&rawRow))
			row := json.RawMessage(rawRow)
			if !validBinding {
				row = copyProductCatalogRowWithJSONFieldsChanged(t, row, fmt.Sprintf(`{"product_category_id":%d}`, rootID))
			}
			table := changed.Tables["products"]
			table.Rows = append(table.Rows, row)
			changed.Tables["products"] = table
			t.Cleanup(func() {
				_, err := stage.ExecContext(context.Background(), `DELETE FROM products WHERE id=900003`)
				require.NoError(t, err)
			})
			_, before, _, err := captureDatabaseInventory(ctx, stage)
			require.NoError(t, err)
			_, _, result, err := cfg.applyProductCatalogRowsOnlyToIsolatedDatabase(ctx, stage, changed)
			if validBinding {
				require.NoError(t, err)
				require.NotNil(t, result)
				var restoredType string
				require.NoError(t, stage.QueryRowContext(context.Background(), `SELECT product_type FROM products WHERE id=900003 AND product_category_id=$1`, carbonID).Scan(&restoredType))
				require.Equal(t, "carbon_rim", restoredType)
			} else {
				require.ErrorContains(t, err, "system category bindings")
				_, after, _, err := captureDatabaseInventory(ctx, stage)
				require.NoError(t, err)
				require.Equal(t, before, after)
			}
		})
	}

	for _, version := range []int{0, productCatalogStandardWriteGuardVersion + 1} {
		t.Run(fmt.Sprintf("product_standard_trigger_stopped_guard_version_%d", version), func(t *testing.T) {
			task := uploadProductCatalogPackageForIntegrationTest(t, manager, packageData)
			require.Equal(t, "product_module_isolated_verified", task.State, task.Error)
			require.Equal(t, productCatalogStandardWriteGuardVersion, task.ProductCatalogPreviewBinding.RequiredImageCleanupProtection.StandardWriteGuardVersion)
			task.ProductCatalogPreviewBinding.RequiredImageCleanupProtection.StandardWriteGuardVersion = version
			require.NoError(t, manager.saveTask(task))
			err := manager.rehearseProductCatalogPairedSwitchAndVerifiedRollback(ctx, task.ID, nil)
			require.ErrorContains(t, err, "cleanup protection is missing or changed")
			task, err = manager.GetTask(task.ID)
			require.NoError(t, err)
			require.Nil(t, task.Journal)
		})
	}
}

func copyProductCatalogSnapshotWithMatchingRowJSONFieldsChanged(t *testing.T, snapshot productCatalogModuleSnapshot, tableName, slug, mutationJSON string) productCatalogModuleSnapshot {
	t.Helper()
	encoded, err := json.Marshal(snapshot)
	require.NoError(t, err)
	var copied productCatalogModuleSnapshot
	require.NoError(t, json.Unmarshal(encoded, &copied))
	if tableName == "" {
		return copied
	}
	table := copied.Tables[tableName]
	matched := false
	for i, row := range table.Rows {
		var fields map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(row, &fields))
		var actualSlug string
		require.NoError(t, json.Unmarshal(fields["slug"], &actualSlug))
		if actualSlug == slug {
			table.Rows[i] = copyProductCatalogRowWithJSONFieldsChanged(t, row, mutationJSON)
			matched = true
		}
	}
	require.True(t, matched, "the mutation must exercise an actual source row")
	copied.Tables[tableName] = table
	return copied
}

func copyProductCatalogRowWithJSONFieldsChanged(t *testing.T, row json.RawMessage, mutationJSON string) json.RawMessage {
	t.Helper()
	var fields, mutations map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(row, &fields))
	require.NoError(t, json.Unmarshal([]byte(mutationJSON), &mutations))
	for field, value := range mutations {
		fields[field] = value
	}
	encoded, err := json.Marshal(fields)
	require.NoError(t, err)
	return encoded
}
