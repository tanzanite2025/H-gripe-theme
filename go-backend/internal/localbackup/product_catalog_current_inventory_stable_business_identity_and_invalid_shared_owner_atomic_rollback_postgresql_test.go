package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// Run against the same complete-migrations stage that was restored through the
// encrypted HTTP upload. No test writes to a production or developer database.
func verifyProductCatalogCurrentInventoryStableIdentitiesAndInvalidSharedOwnerAtomicRollback(t *testing.T, cfg Configuration, stage *sql.DB, original productCatalogModuleSnapshot) {
	t.Helper()
	ctx := context.Background()
	findRow := func(snapshot *productCatalogModuleSnapshot, tableName, field, value string) (int, map[string]json.RawMessage) {
		t.Helper()
		for index, row := range snapshot.Tables[tableName].Rows {
			actual, err := productCatalogRowString(row, field)
			require.NoError(t, err)
			if actual == value {
				var fields map[string]json.RawMessage
				require.NoError(t, json.Unmarshal(row, &fields))
				return index, fields
			}
		}
		t.Fatalf("missing explicit product fixture row in %s", tableName)
		return 0, nil
	}
	for _, scenario := range []struct {
		name, table, key, value, changedColumn, changedJSON, expectedError string
	}{
		{"existing_sku_cannot_be_reassigned", "product_variants", "sku", "product-backup-sku", "sku", `"another-physical-sku"`, "identity or inventory owner differs"},
		{"existing_variant_cannot_move_to_another_product", "product_variants", "sku", "product-backup-sku", "product_id", `"another-selected-product"`, "identity or inventory owner differs"},
		{"existing_inventory_owner_cannot_change", "product_variants", "sku", "product-backup-translated-sku", "master_variant_id", `null`, "identity or inventory owner differs"},
		{"existing_options_cannot_change_physical_identity", "product_variants", "sku", "product-backup-sku", "option_values", `{"size":"different-physical-item"}`, "identity or inventory owner differs"},
		{"reused_primary_key_with_different_creation_identity_is_rejected", "product_variants", "sku", "product-backup-sku", "created_at", `"2000-01-01T00:00:00Z"`, "identity or inventory owner differs"},
		{"existing_product_slug_cannot_be_reassigned", "products", "locale", "en", "slug", `"another-catalog-item"`, "identity or inventory owner differs"},
		{"existing_translation_parent_cannot_change", "products", "locale", "zh", "parent_id", `null`, "identity or inventory owner differs"},
		{"negative_snapshot_stock_is_rejected_before_merge", "product_variants", "sku", "product-backup-sku", "stock", `-1`, "invalid inventory quantity"},
		{"null_snapshot_stock_is_rejected_before_merge", "product_variants", "sku", "product-backup-sku", "stock", `null`, "invalid inventory quantity"},
		{"overflowing_snapshot_stock_is_rejected_before_merge", "product_variants", "sku", "product-backup-sku", "stock", `2147483648`, "invalid inventory quantity"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			data, err := json.Marshal(original)
			require.NoError(t, err)
			var changed productCatalogModuleSnapshot
			require.NoError(t, json.Unmarshal(data, &changed))
			index, fields := findRow(&changed, scenario.table, scenario.key, scenario.value)
			fields[scenario.changedColumn] = json.RawMessage(scenario.changedJSON)
			if scenario.name == "existing_variant_cannot_move_to_another_product" {
				_, anotherProduct := findRow(&changed, "products", "locale", "zh")
				fields["product_id"] = anotherProduct["id"]
			}
			row, err := json.Marshal(fields)
			require.NoError(t, err)
			changed.Tables[scenario.table].Rows[index] = row
			_, before, _, err := captureDatabaseInventory(ctx, stage)
			require.NoError(t, err)
			_, _, result, err := cfg.applyProductCatalogRowsOnlyToIsolatedDatabase(ctx, stage, changed)
			require.ErrorContains(t, err, scenario.expectedError)
			require.Nil(t, result, "failed merges must not return a successful inventory result")
			_, after, _, err := captureDatabaseInventory(ctx, stage)
			require.NoError(t, err)
			require.Equal(t, before, after, "failed merges preserve every row and sequence")
		})
	}
	for _, ownerCase := range []string{"new_variant_cannot_own_itself", "new_variant_cannot_own_translated_inventory", "new_variant_owner_must_belong_to_the_translation_parent"} {
		t.Run(ownerCase, func(t *testing.T) {
			data, err := json.Marshal(original)
			require.NoError(t, err)
			var changed productCatalogModuleSnapshot
			require.NoError(t, json.Unmarshal(data, &changed))
			_, fields := findRow(&changed, "product_variants", "sku", "product-backup-new-sku")
			_, root := findRow(&changed, "product_variants", "sku", "product-backup-sku")
			_, translation := findRow(&changed, "product_variants", "sku", "product-backup-translated-sku")
			fields["id"] = json.RawMessage(`9223372036854700000`)
			fields["sku"] = json.RawMessage(`"invalid-new-inventory-owner"`)
			fields["stock"] = json.RawMessage(`0`)
			fields["option_values"] = json.RawMessage(`{"size":"invalid-new-owner"}`)
			switch ownerCase {
			case "new_variant_cannot_own_itself":
				fields["master_variant_id"] = fields["id"]
			case "new_variant_cannot_own_translated_inventory":
				fields["master_variant_id"] = translation["id"]
			default:
				fields["master_variant_id"] = root["id"]
			}
			row, err := json.Marshal(fields)
			require.NoError(t, err)
			table := changed.Tables["product_variants"]
			table.Rows = append(table.Rows, row)
			changed.Tables["product_variants"] = table
			_, before, _, err := captureDatabaseInventory(ctx, stage)
			require.NoError(t, err)
			_, _, result, err := cfg.applyProductCatalogRowsOnlyToIsolatedDatabase(ctx, stage, changed)
			require.ErrorContains(t, err, "invalid shared inventory owner")
			require.Nil(t, result)
			_, after, _, err := captureDatabaseInventory(ctx, stage)
			require.NoError(t, err)
			require.Equal(t, before, after, "invalid inventory ownership rolls back even prior row inserts and preserves sequences")
		})
	}
	t.Run("current_zero_stock_is_preserved_without_replaying_order_events", func(t *testing.T) {
		_, err := stage.ExecContext(ctx, `UPDATE product_variants SET stock=0 WHERE sku='product-backup-sku'`)
		require.NoError(t, err)
		_, _, summary, err := cfg.applyProductCatalogRowsOnlyToIsolatedDatabase(ctx, stage, original)
		require.NoError(t, err)
		require.GreaterOrEqual(t, summary.DifferentSnapshotStockValuesNotApplied, 1)
		var quantity int
		require.NoError(t, stage.QueryRowContext(ctx, `SELECT stock FROM product_variants WHERE sku='product-backup-sku'`).Scan(&quantity))
		require.Zero(t, quantity)
		var processedEvents int
		require.NoError(t, stage.QueryRowContext(ctx, `SELECT count(*) FROM outbox_events WHERE event_key='product-backup-completed' AND status='processed' AND processed_at IS NOT NULL`).Scan(&processedEvents))
		require.Equal(t, 1, processedEvents)
		_, err = stage.ExecContext(ctx, `UPDATE product_variants SET stock=7 WHERE sku='product-backup-sku'`)
		require.NoError(t, err)
	})
	t.Run("target_only_translated_inventory_reference_is_validated", func(t *testing.T) {
		var retainedID int64
		require.NoError(t, stage.QueryRowContext(ctx, `SELECT id FROM product_variants WHERE sku='target-only-translated-sku'`).Scan(&retainedID))
		_, err := stage.ExecContext(ctx, `UPDATE product_variants SET option_values='{"retained":"mismatched-owner"}' WHERE id=$1`, retainedID)
		require.NoError(t, err)
		_, before, _, err := captureDatabaseInventory(ctx, stage)
		require.NoError(t, err)
		_, _, _, err = cfg.applyProductCatalogRowsOnlyToIsolatedDatabase(ctx, stage, original)
		require.ErrorContains(t, err, "invalid shared inventory owner")
		_, after, _, err := captureDatabaseInventory(ctx, stage)
		require.NoError(t, err)
		require.Equal(t, before, after)
		_, err = stage.ExecContext(ctx, `UPDATE product_variants SET option_values='{}' WHERE id=$1`, retainedID)
		require.NoError(t, err)
	})
}
