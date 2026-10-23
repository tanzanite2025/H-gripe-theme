package localbackup

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/lib/pq"
)

// A SELECT-only union presents target rows plus approved source projections.
// PostgreSQL performs typed equality for compound references and natural keys.
// Full source content stays in the isolated database; its observed digest is
// carried separately, never recomputed from a partial relationship projection.
func (graph *moduleRecoveryRowDependencyGraph) customerServiceEffectiveSelectionReadOnlyRelation(table string, arguments []any) (string, string, []any, error) {
	relation := "public." + pq.QuoteIdentifier(table) + " record"
	digest := "encode(sha256(convert_to(to_jsonb(record)::text,'UTF8')),'hex')"
	selected := graph.SelectedCustomerServiceRows[table]
	if len(selected) == 0 {
		return relation, digest, arguments, nil
	}
	if !customerServiceRecoveryPlanProposedRowReplacementAllowed(table) || len(graph.Observation.PrimaryKeys[table]) == 0 {
		return "", "", nil, errors.New("effective selection cannot overlay a shared table or a table without a primary key")
	}
	type selectedRowProjection struct {
		Values map[string]json.RawMessage `json:"values"`
		Digest string                     `json:"digest"`
	}
	var identities []string
	for identity := range selected {
		identities = append(identities, identity)
	}
	sort.Strings(identities)
	projections := make([]selectedRowProjection, 0, len(identities))
	for _, identity := range identities {
		row := selected[identity]
		values := map[string]json.RawMessage{}
		for column, value := range row.Values {
			if graph.Observation.Columns[table][column] {
				values[column] = value
			}
		}
		for _, column := range []string{"message_type", "metadata"} {
			if value, exists := row.Values["_structured_"+column]; exists && graph.Observation.Columns[table][column] {
				values[column] = value
			}
		}
		projections = append(projections, selectedRowProjection{values, row.Digest})
	}
	encoded, err := json.Marshal(projections)
	if err != nil {
		return "", "", nil, err
	}
	arguments = append(arguments, string(encoded))
	parameter := fmt.Sprintf("$%d::jsonb", len(arguments))
	var equalities []string
	for _, column := range graph.Observation.PrimaryKeys[table] {
		quoted := pq.QuoteIdentifier(column)
		equalities = append(equalities, "original."+quoted+"=(jsonb_populate_record(NULL::public."+pq.QuoteIdentifier(table)+",selected->'values'))."+quoted)
	}
	relation = `(SELECT original AS effective_record, encode(sha256(convert_to(to_jsonb(original)::text,'UTF8')),'hex') AS effective_digest FROM public.` + pq.QuoteIdentifier(table) + ` original WHERE NOT EXISTS(SELECT 1 FROM jsonb_array_elements(` + parameter + `) selected WHERE ` + strings.Join(equalities, " AND ") + `) UNION ALL SELECT jsonb_populate_record(NULL::public.` + pq.QuoteIdentifier(table) + `,selected->'values'),selected->>'digest' FROM jsonb_array_elements(` + parameter + `) selected) effective CROSS JOIN LATERAL (SELECT (effective.effective_record).*) record`
	return relation, "effective.effective_digest", arguments, nil
}
