package localbackup

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderModuleConfiguredComponentInventoryRequiresValidOriginalAllocationAndDependency(t *testing.T) {
	for _, scenario := range []string{"original_component_dependency", "reused_primary_variant_dependency", "repeated_component_allocations", "original_dependency_missing", "original_dependency_marked_missing", "original_dependency_digest_invalid", "wrong_original_component_identity", "null_configuration", "array_configuration", "null_allocations", "object_allocations", "null_allocation_entry", "missing_variant", "negative_variant", "fractional_variant", "string_variant", "variant_outside_postgresql_bigint", "missing_quantity", "zero_quantity", "negative_quantity", "fractional_quantity", "string_quantity", "quantity_outside_postgresql_integer", "multiplied_quantity_overflow", "repeated_quantity_overflow"} {
		t.Run(scenario, func(t *testing.T) {
			snapshot := createOrderModuleSettledActiveOrderRestoreContractTestSnapshot()
			configuration := json.RawMessage(`{"inventory_allocations":[{"variant_id":12,"quantity":2,"group_slug":"spokes","value_key":"original"}]}`)
			quantity := 2
			dependency := selectedBusinessModuleExternalReference{Table: "product_variants", Values: map[string]string{"id": "12"}, RowSHA256: strings.Repeat("a", 64)}
			valid := scenario == "original_component_dependency" || scenario == "reused_primary_variant_dependency" || scenario == "repeated_component_allocations"
			switch scenario {
			case "reused_primary_variant_dependency":
				configuration = json.RawMessage(`{"inventory_allocations":[{"variant_id":5,"quantity":1}]}`)
				dependency.Values["id"] = "5"
			case "repeated_component_allocations":
				configuration = json.RawMessage(`{"inventory_allocations":[{"variant_id":12,"quantity":2},{"variant_id":12,"quantity":3}]}`)
			case "original_dependency_marked_missing":
				dependency.RowSHA256 = "missing"
			case "original_dependency_digest_invalid":
				dependency.RowSHA256 = "invalid"
			case "wrong_original_component_identity":
				dependency.Values["id"] = "13"
			case "null_configuration":
				configuration = json.RawMessage(`null`)
			case "array_configuration":
				configuration = json.RawMessage(`[]`)
			case "null_allocations":
				configuration = json.RawMessage(`{"inventory_allocations":null}`)
			case "object_allocations":
				configuration = json.RawMessage(`{"inventory_allocations":{}}`)
			case "null_allocation_entry":
				configuration = json.RawMessage(`{"inventory_allocations":[null]}`)
			case "missing_variant", "negative_variant", "fractional_variant", "string_variant", "variant_outside_postgresql_bigint":
				variantJSON := map[string]string{"missing_variant": "null", "negative_variant": "-1", "fractional_variant": "1.5", "string_variant": `"12"`, "variant_outside_postgresql_bigint": "9223372036854775808"}[scenario]
				configuration = json.RawMessage(`{"inventory_allocations":[{"variant_id":` + variantJSON + `,"quantity":2}]}`)
			case "missing_quantity", "zero_quantity", "negative_quantity", "fractional_quantity", "string_quantity", "quantity_outside_postgresql_integer":
				quantityJSON := map[string]string{"missing_quantity": "null", "zero_quantity": "0", "negative_quantity": "-1", "fractional_quantity": "1.5", "string_quantity": `"2"`, "quantity_outside_postgresql_integer": "2147483648"}[scenario]
				configuration = json.RawMessage(`{"inventory_allocations":[{"variant_id":12,"quantity":` + quantityJSON + `}]}`)
			case "multiplied_quantity_overflow":
				configuration = json.RawMessage(`{"inventory_allocations":[{"variant_id":12,"quantity":2147483647}]}`)
			case "repeated_quantity_overflow":
				configuration = json.RawMessage(`{"inventory_allocations":[{"variant_id":12,"quantity":1000000000},{"variant_id":12,"quantity":1000000000}]}`)
			}
			var item map[string]any
			require.NoError(t, json.Unmarshal(snapshot.Tables["order_items"].Rows[0], &item))
			item["configuration_snapshot"], item["quantity"] = configuration, quantity
			encoded, err := json.Marshal(item)
			require.NoError(t, err)
			snapshot.Tables["order_items"] = businessModuleCompleteTableSnapshot{Rows: []json.RawMessage{encoded}}
			if scenario != "original_dependency_missing" {
				snapshot.ExternalDependencies = []selectedBusinessModuleExternalReference{dependency}
			}
			result := inspectOrderModuleMissingOrderRecordsRestoreEligibility(snapshot, map[string][]json.RawMessage{"orders": snapshot.Tables["orders"].Rows, "order_items": snapshot.Tables["order_items"].Rows}, &SelectedBusinessModuleIsolatedRestorePreview{}, nil)
			require.Equal(t, valid, result.Eligible, "%v", result.Blockers)
			if valid {
				require.Equal(t, 1, result.SettledActiveOrderOwners)
			}
		})
	}
}

func TestOrderModulePendingOrderRestoreExplainsMissingReservationReleaseProof(t *testing.T) {
	snapshot := createOrderModuleSettledActiveOrderRestoreContractTestSnapshot()
	snapshot.Tables["orders"] = businessModuleCompleteTableSnapshot{Rows: []json.RawMessage{json.RawMessage(`{"id":1,"status":"pending","payment_status":"unpaid"}`)}}
	result := inspectOrderModuleMissingOrderRecordsRestoreEligibility(snapshot, map[string][]json.RawMessage{"orders": snapshot.Tables["orders"].Rows}, &SelectedBusinessModuleIsolatedRestorePreview{}, nil)
	require.False(t, result.Eligible)
	require.Contains(t, result.Blockers, "pending_order_inventory_reservation_requires_durable_release_reconciliation:orders")
}
