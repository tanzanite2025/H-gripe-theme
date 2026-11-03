package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
)

type orderModuleConfiguredComponentInventoryAllocation struct {
	VariantID uint64 `json:"variant_id"`
	Quantity  int64  `json:"quantity"`
}

// Keep the original JSON untouched. This reader validates only the quantities
// and identities used by the inventory services, without replaying allocations.
func readOrderModuleConfiguredComponentInventoryAllocationsForRestore(item json.RawMessage) ([]orderModuleConfiguredComponentInventoryAllocation, string) {
	var fields map[string]json.RawMessage
	var configuration map[string]json.RawMessage
	if json.Unmarshal(item, &fields) != nil || json.Unmarshal(fields["configuration_snapshot"], &configuration) != nil || configuration == nil {
		return nil, "active_order_component_inventory_configuration_invalid"
	}
	raw, present := configuration["inventory_allocations"]
	if !present {
		return nil, ""
	}
	var allocations []orderModuleConfiguredComponentInventoryAllocation
	if json.Unmarshal(raw, &allocations) != nil || strings.TrimSpace(string(raw)) == "null" || len(allocations) > 4096 {
		return nil, "active_order_component_inventory_configuration_invalid"
	}
	quantityText, _ := readBusinessModuleSnapshotRowColumnText(item, "quantity")
	itemQuantity, err := strconv.ParseInt(quantityText, 10, 64)
	if err != nil || itemQuantity <= 0 {
		return nil, "active_order_component_inventory_configuration_invalid"
	}
	totalByVariant := map[uint64]int64{}
	const maximumInventoryQuantity = int64(2147483647) // PostgreSQL stock is INTEGER.
	for _, allocation := range allocations {
		if allocation.VariantID == 0 || allocation.VariantID > (^uint64(0)>>1) || allocation.Quantity <= 0 || allocation.Quantity > maximumInventoryQuantity/itemQuantity {
			return nil, "active_order_component_inventory_configuration_invalid"
		}
		quantity := allocation.Quantity * itemQuantity
		if totalByVariant[allocation.VariantID] > maximumInventoryQuantity-quantity {
			return nil, "active_order_component_inventory_configuration_invalid"
		}
		totalByVariant[allocation.VariantID] += quantity
	}
	return allocations, ""
}

func orderModuleComponentInventoryAllocationsHaveRecordedVariantDependencies(snapshot selectedBusinessModuleCompleteSnapshot, allocations []orderModuleConfiguredComponentInventoryAllocation) bool {
	for _, allocation := range allocations {
		id := strconv.FormatUint(allocation.VariantID, 10)
		found := false
		for _, row := range snapshot.Tables["product_variants"].Rows {
			candidate, _ := readBusinessModuleSnapshotRowColumnText(row, "id")
			found = found || candidate == id
		}
		for _, reference := range snapshot.ExternalDependencies {
			if reference.Table == "product_variants" && len(reference.Values) == 1 && reference.Values["id"] == id && moduleRecoverySideEffectReceiptRequestDigestIsValid(reference.RowSHA256) {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Component references live inside configuration_snapshot, and translated SKU
// rows delegate stock to master_variant_id. Capture those precise dependencies
// as hashes, never as unselected product data. Ordinary dependency comparison
// then also rejects older packages that lack the original master/component hash.
func captureOrderModuleConfiguredComponentAndMasterInventoryDependencies(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, observation *moduleRecoveryDatabaseObservation, dependencies map[string]selectedBusinessModuleExternalReference) error {
	orderModuleSelected := false
	for _, module := range snapshot.Modules {
		orderModuleSelected = orderModuleSelected || module == "orders"
	}
	if !orderModuleSelected {
		return nil
	}
	selectedRows := indexOrderModuleHistoricalRowsByTableAndID(snapshot)
	readRows := map[string]json.RawMessage{}
	readInventoryDependencyRow := func(table, id string) (json.RawMessage, error) {
		identity := table + "/" + id
		if row, read := readRows[identity]; read {
			return row, nil
		}
		var raw string
		query := `SELECT to_jsonb(record)::text FROM public.product_variants record WHERE id::text=$1`
		if table == "products" {
			query = `SELECT to_jsonb(record)::text FROM public.products record WHERE id::text=$1`
		}
		err := tx.QueryRowContext(ctx, query, id).Scan(&raw)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		row := json.RawMessage(raw)
		readRows[identity] = row
		if len(selectedRows[table][id]) == 0 {
			digest := "missing"
			if err == nil {
				digest = businessModuleRecoveryBytesSHA256(row)
			}
			reference := selectedBusinessModuleExternalReference{Table: table, Values: map[string]string{"id": id}, RowSHA256: digest}
			dependencies[selectedBusinessModuleExternalReferenceIdentity(reference)] = reference
		}
		return row, nil
	}
	variantIDs := map[string]bool{}
	for _, item := range snapshot.Tables["order_items"].Rows {
		variantID, _ := readBusinessModuleSnapshotRowColumnText(item, "variant_id")
		if canonicalModuleRecoveryDurableEventPositiveID(variantID) {
			variantIDs[variantID] = true
		}
		allocations, blocker := readOrderModuleConfiguredComponentInventoryAllocationsForRestore(item)
		if blocker != "" {
			// Export the original malformed record; its restore remains blocked.
			continue
		}
		for _, allocation := range allocations {
			variantIDs[strconv.FormatUint(allocation.VariantID, 10)] = true
		}
	}
	if !observation.Columns["product_variants"]["id"] || !observation.Columns["products"]["id"] {
		return nil // Existing schema/row diagnostics explain missing dependencies.
	}
	for variantID := range variantIDs {
		variant, err := readInventoryDependencyRow("product_variants", variantID)
		if err != nil {
			return err
		}
		productID, _ := readBusinessModuleSnapshotRowColumnText(variant, "product_id")
		if canonicalModuleRecoveryDurableEventPositiveID(productID) {
			if _, err := readInventoryDependencyRow("products", productID); err != nil {
				return err
			}
		}
		masterID, _ := readBusinessModuleSnapshotRowColumnText(variant, "master_variant_id")
		if !canonicalModuleRecoveryDurableEventPositiveID(masterID) {
			continue
		}
		master, err := readInventoryDependencyRow("product_variants", masterID)
		if err != nil {
			return err
		}
		masterProductID, _ := readBusinessModuleSnapshotRowColumnText(master, "product_id")
		if canonicalModuleRecoveryDurableEventPositiveID(masterProductID) {
			if _, err := readInventoryDependencyRow("products", masterProductID); err != nil {
				return err
			}
		}
	}
	return nil
}
