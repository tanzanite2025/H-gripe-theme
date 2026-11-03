package localbackup

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

// Read full-row field values without numeric conversion and build typed primary-key predicates with quoted identifiers and bound JSON values.
func readBusinessModuleSnapshotRowColumnText(row json.RawMessage, column string) (string, error) {
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

func businessModuleCompleteRowTypedPrimaryKeyPredicate(table string, primaryKeys []string) string {
	var equalities []string
	for _, column := range primaryKeys {
		quoted := pq.QuoteIdentifier(column)
		equalities = append(equalities, "record."+quoted+"=(jsonb_populate_record(NULL::public."+pq.QuoteIdentifier(table)+",$1::jsonb))."+quoted)
	}
	return strings.Join(equalities, " AND ")
}
