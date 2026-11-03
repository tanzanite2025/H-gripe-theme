package localbackup

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// These checks compare independent export, domain-write and dependency contracts.
// They do not turn the dependency registry's OwnedTables into write permission.
func TestBusinessModuleExportCombinationsKeepReviewedTableUnionAndSharedRowFilters(t *testing.T) {
	options := availableBusinessModuleBackupOptions()
	require.Len(t, options, 6)
	registered := registeredStorefrontModuleRecoveryTableNames()
	for selection := 1; selection < 1<<len(options); selection++ {
		var modules []string
		for index, option := range options {
			if selection&(1<<index) != 0 {
				modules = append(modules, option.ID)
			}
		}
		t.Run(strings.Join(modules, "+"), func(t *testing.T) {
			normalized, err := normalizeSelectedBusinessBackupModules(modules)
			require.NoError(t, err)
			selectors := selectedBusinessModuleTableSelectors(normalized)
			expectedTables, selected := map[string]string{}, map[string]bool{}
			for _, module := range modules {
				selected[module] = true
				for table := range selectedBusinessModuleTableSelectors([]string{module}) {
					expectedTables[table] = "included"
				}
			}
			require.Equal(t, sortedBusinessModuleScopeTableNamesForArchitectureBoundaryChecks(expectedTables), sortedBusinessModuleScopeTableNamesForArchitectureBoundaryChecks(selectors))
			for table, predicate := range selectors {
				require.True(t, registered[table], "unregistered export table: %s", table)
				require.NotEmpty(t, predicate)
			}
			for _, table := range strings.Fields("schema_migrations audit_logs product_view_count_flush_receipts ops_connectors ops_deployment_workflow_runs") {
				require.NotContains(t, selectors, table, "module selection must not export recovery/operations bookkeeping")
			}
			for table, owner := range map[string]string{"users": "customers", "orders": "orders", "ticket_messages": "customer_service", "social_oauth_connections": "configuration"} {
				_, included := selectors[table]
				require.Equal(t, selected[owner], included, "%s must require an explicit %s export", table, owner)
			}
			if selected["customers"] {
				require.Equal(t, selectedBusinessModuleTableSelectors([]string{"customers"})["users"], selectors["users"], "combining modules cannot include staff accounts")
				require.NotEqual(t, "true", selectors["users"])
			}
			if selected["customer_service"] {
				for _, table := range []string{"tickets", "ticket_messages", "customer_service_inbox_states"} {
					require.Equal(t, selectedBusinessModuleTableSelectors([]string{"customer_service"})[table], selectors[table], "combining modules cannot include non-chat tickets")
					require.NotEqual(t, "true", selectors[table])
				}
			}
			if selected["media"] {
				for _, table := range []string{"media_upload_sessions", "ticket_message_media_attachments"} {
					require.Equal(t, "true", selectors[table], "explicit media selection includes all registered receipts/attachments")
				}
			} else if selected["customer_service"] {
				for _, table := range []string{"media_upload_sessions", "ticket_message_media_attachments"} {
					require.Equal(t, selectedBusinessModuleTableSelectors([]string{"customer_service"})[table], selectors[table])
					require.NotEqual(t, "true", selectors[table])
				}
			}
			for _, module := range []string{"orders", "customer_service"} {
				if selected[module] {
					predicate := selectedBusinessModuleTableSelectors([]string{module})["outbox_events"]
					require.Contains(t, selectors["outbox_events"], strings.TrimSuffix(strings.TrimPrefix(predicate, "("), ")"), "combining modules must preserve selected event families")
					require.NotEqual(t, "true", selectors["outbox_events"])
				}
			}
			snapshot := selectedBusinessModuleCompleteSnapshot{Modules: normalized}
			require.Equal(t, len(normalized) == 1 && normalized[0] == "customers", customerProfileModuleHasDedicatedMissingRowContract(snapshot, "users"))
			require.Equal(t, len(normalized) == 1 && normalized[0] == "orders", orderModuleHasDedicatedCompletedHistoryInsertContract(snapshot, "outbox_events"))
			eligibility := inspectOrderModuleMissingOrderRecordsRestoreEligibility(snapshot, nil, &SelectedBusinessModuleIsolatedRestorePreview{}, nil)
			if len(normalized) > 1 {
				require.False(t, eligibility.Eligible, "multi-select export cannot inherit standalone order writes")
				require.Contains(t, eligibility.Blockers, "only_a_separate_order_package_is_supported")
			}
		})
	}
	index := map[string]map[string]string{}
	for _, option := range options {
		index[option.ID] = selectedBusinessModuleTableSelectors([]string{option.ID})
	}
	reviewedOverlaps := map[string][]string{
		"configuration+media":     {"site_favicon_assets", "site_logo_assets"},
		"configuration+orders":    {"payment_protection_controls"},
		"customer_service+media":  {"media_upload_sessions", "ticket_message_media_attachments"},
		"customer_service+orders": {"outbox_events"},
		"media+products":          {"product_media"},
	}
	moduleNames := make([]string, 0, len(index))
	for module := range index {
		moduleNames = append(moduleNames, module)
	}
	sort.Strings(moduleNames)
	for left, first := range moduleNames {
		for _, second := range moduleNames[left+1:] {
			var overlap []string
			for table := range index[first] {
				if _, shared := index[second][table]; shared {
					overlap = append(overlap, table)
				}
			}
			require.ElementsMatch(t, reviewedOverlaps[first+"+"+second], overlap, "unreviewed cross-module export overlap: %s+%s", first, second)
		}
	}
	var attachedProductMedia []string
	for table := range productCatalogAllowedTableNames() {
		if _, directlySelected := index["products"][table]; !directlySelected {
			attachedProductMedia = append(attachedProductMedia, table)
		}
	}
	require.ElementsMatch(t, []string{"media_assets", "media_asset_derivatives"}, attachedProductMedia, "product payload adds referenced registrations, not arbitrary shared tables")
	encoded, err := json.Marshal(index)
	require.NoError(t, err)
	t.Logf("ARCH07_EXPORT_SELECTOR_INDEX=%s", encoded)
}

func TestCustomerProfileWriteTablesStayWithinExplicitCustomerExportAndExcludeSharedMedia(t *testing.T) {
	exported := selectedBusinessModuleTableSelectors([]string{"customers"})
	writable := customerProfileModuleOwnedTableNames()
	require.ElementsMatch(t, writable, sortedBusinessModuleScopeTableNamesForArchitectureBoundaryChecks(exported), "separate export/writer registrations must be reviewed together")
	require.Len(t, writable, 12)
	require.Len(t, exported, len(writable), "writer order must not contain duplicate names")
	for _, table := range writable {
		require.True(t, registeredStorefrontModuleRecoveryTableNames()[table])
	}
	for _, table := range []string{"media_assets", "media_asset_derivatives", "orders", "tickets", "outbox_events", "audit_logs"} {
		require.False(t, customerProfileModuleHasDedicatedMissingRowContract(selectedBusinessModuleCompleteSnapshot{Modules: []string{"customers"}}, table), "shared or audit table cannot become customer-owned: %s", table)
	}
}

func TestCustomerServiceConversationWritesStayNarrowerThanExportAndLegacyOwnedTables(t *testing.T) {
	contract := registeredOrderAndCustomerServiceRecoveryDependencyContracts()["customer_service"]
	exported := selectedBusinessModuleTableSelectors([]string{"customer_service"})
	registryAndEvent := append(append([]string{}, contract.OwnedTables...), "outbox_events")
	require.ElementsMatch(t, registryAndEvent, sortedBusinessModuleScopeTableNamesForArchitectureBoundaryChecks(exported))
	var writable, exportedReadOnly []string
	for table := range registeredStorefrontModuleRecoveryTableNames() {
		if customerServiceRecoveryPlanProposedRowReplacementAllowed(table) {
			writable = append(writable, table)
			require.Contains(t, exported, table)
		}
	}
	require.ElementsMatch(t, []string{"tickets", "ticket_messages", "customer_service_inbox_states", "ticket_message_media_attachments"}, writable)
	for table := range exported {
		if !customerServiceRecoveryPlanProposedRowReplacementAllowed(table) {
			exportedReadOnly = append(exportedReadOnly, table)
		}
	}
	require.ElementsMatch(t, strings.Fields("media_upload_sessions customer_service_retention_policies ticket_auto_replies customer_service_agent_profiles customer_service_agent_groups customer_service_agent_group_members outbox_events"), exportedReadOnly)
	for _, table := range contract.SharedTables {
		require.False(t, customerServiceRecoveryPlanProposedRowReplacementAllowed(table), "a shared dependency cannot acquire conversation replacement permission")
	}
}

func TestOrderExportExtrasAndSharedFactsDoNotAcquireMissingRowRestorePermission(t *testing.T) {
	contract := registeredOrderAndCustomerServiceRecoveryDependencyContracts()["orders"]
	exported := selectedBusinessModuleTableSelectors([]string{"orders"})
	owned := map[string]bool{}
	for _, table := range contract.OwnedTables {
		owned[table] = true
		require.Contains(t, exported, table)
	}
	var exportExtras, exportedShared []string
	for table := range exported {
		if !owned[table] {
			exportExtras = append(exportExtras, table)
		}
	}
	for _, table := range contract.SharedTables {
		if _, included := exported[table]; included {
			exportedShared = append(exportedShared, table)
		}
	}
	require.ElementsMatch(t, []string{"coupon_usage", "email_delivery_records", "outbox_events"}, exportExtras)
	require.ElementsMatch(t, []string{"coupon_usage", "outbox_events"}, exportedShared, "export inclusion and shared classification may overlap intentionally")
	for _, table := range contract.SharedTables {
		if table == "outbox_events" {
			continue // Completed order-owned events have their separate row/receipt contract.
		}
		t.Run(table, func(t *testing.T) {
			snapshot := selectedBusinessModuleCompleteSnapshot{Modules: []string{"orders"}, Tables: map[string]businessModuleCompleteTableSnapshot{
				"orders": {Rows: []json.RawMessage{json.RawMessage(`{"id":1,"status":"completed"}`)}},
			}}
			missing := map[string][]json.RawMessage{table: {json.RawMessage(`{"id":9,"order_id":1}`)}}
			eligibility := inspectOrderModuleMissingOrderRecordsRestoreEligibility(snapshot, missing, &SelectedBusinessModuleIsolatedRestorePreview{}, nil)
			require.False(t, eligibility.Eligible, "order export or dependency registration must not authorize shared-row insertion")
			require.Contains(t, eligibility.Blockers, "missing_execution_shared_or_unsupported_records:"+table)
		})
	}
}

func sortedBusinessModuleScopeTableNamesForArchitectureBoundaryChecks(selectors map[string]string) []string {
	names := make([]string, 0, len(selectors))
	for table := range selectors {
		names = append(names, table)
	}
	sort.Strings(names)
	return names
}
