package localbackup

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSelectedBusinessModuleCustomerServiceSnapshotProjectionRejectsOversizedOrUnidentifiableRowsBeforeDatabaseAccess(t *testing.T) {
	for _, scenario := range []struct{ name, row, code string }{
		{"invalid_json", `{`, "module_snapshot_projection_invalid"},
		{"missing_primary_identity", `{"category":"customer_service"}`, "module_snapshot_projection_resource_limit"},
		{"oversized_structured_reference", `{"id":1,"metadata":{"reference":"` + strings.Repeat("x", 65537) + `"}}`, "module_snapshot_projection_resource_limit"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			observation := &moduleRecoveryDatabaseObservation{PrimaryKeys: map[string][]string{"ticket_messages": {"id"}}, Columns: map[string]map[string]bool{"ticket_messages": {"id": true, "metadata": true}}}
			snapshot := selectedBusinessModuleCompleteSnapshot{Tables: map[string]productCatalogTableSnapshot{"ticket_messages": {Rows: []json.RawMessage{json.RawMessage(scenario.row)}}}}
			checks := newSelectedCustomerServiceModuleDomainChecks()
			_, err := buildSelectedCustomerServiceModuleSnapshotRelationshipRows(context.Background(), nil, snapshot, observation, checks)
			require.NoError(t, err)
			require.False(t, checks.Complete)
			require.Positive(t, checks.IssueCounts[scenario.code])
			encoded, err := json.Marshal(checks)
			require.NoError(t, err)
			require.NotContains(t, string(encoded), strings.Repeat("x", 64))
		})
	}
}
