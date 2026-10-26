package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
)

// Issues disclose registered names and counts, never reference values, hashes,
// account identifiers or conversation contents. Each reference counts once.
type SelectedBusinessModuleExternalDependencyIssue struct {
	Table             string   `json:"table"`
	ReferencedColumns []string `json:"referenced_columns"`
	Reason            string   `json:"reason"`
	Severity          string   `json:"severity"`
	References        int      `json:"references"`
}

func selectedBusinessModuleExternalReferenceIdentity(reference selectedBusinessModuleExternalReference) string {
	encoded, _ := json.Marshal(reference.Values)
	return reference.Table + "/" + string(encoded)
}

// A visitor browsing profile is optional context under the existing conversation
// contract. It is not the signed session identity stored on the ticket itself.
// Only a canonical hash, absent on both sides, reached by registered chat-only
// references qualifies. Actual account/FK loss and changed profiles still block.
func selectedChatExternalReferenceIsOptionalAbsentVisitorBrowsingContext(snapshot selectedBusinessModuleCompleteSnapshot, observation *moduleRecoveryDatabaseObservation, reference selectedBusinessModuleExternalReference) bool {
	if len(snapshot.Modules) != 1 || snapshot.Modules[0] != "customer_service" || reference.Table != "visitor_profiles" || len(reference.Values) != 1 {
		return false
	}
	hash := reference.Values["customer_service_visitor_hash"]
	if len(hash) != 64 || strings.IndexFunc(hash, func(character rune) bool {
		return (character < '0' || character > '9') && (character < 'a' || character > 'f')
	}) >= 0 {
		return false
	}
	for _, foreignKey := range observation.References {
		if foreignKey.TargetTable == "visitor_profiles" && len(snapshot.Tables[foreignKey.SourceTable].Rows) > 0 {
			return false
		}
	}
	foundOptionalReference := false
	for _, source := range append(registeredNonForeignKeyModuleRecoveryReferences(), registeredOrderTransactionBusinessReferences()...) {
		if source.TargetTable != "visitor_profiles" || len(source.TargetColumns) != 1 || source.TargetColumns[0] != "customer_service_visitor_hash" || len(source.SourceColumns) != 1 {
			continue
		}
		for _, row := range snapshot.Tables[source.SourceTable].Rows {
			value, err := productCatalogRowString(row, source.SourceColumns[0])
			if err != nil || value != hash {
				continue
			}
			if (source.SourceTable != "tickets" || source.SourceColumns[0] != "visitor_session_hash") && (source.SourceTable != "media_upload_sessions" || source.SourceColumns[0] != "upload_actor_visitor_session_hash") {
				return false
			}
			foundOptionalReference = true
		}
	}
	return foundOptionalReference
}

func classifySelectedBusinessModuleExternalDependencyIssues(snapshot selectedBusinessModuleCompleteSnapshot, observation *moduleRecoveryDatabaseObservation, actual []selectedBusinessModuleExternalReference, report *SelectedBusinessModuleIsolatedRestorePreview) {
	expected := map[string]selectedBusinessModuleExternalReference{}
	for _, reference := range snapshot.ExternalDependencies {
		expected[selectedBusinessModuleExternalReferenceIdentity(reference)] = reference
	}
	groups := map[string]*SelectedBusinessModuleExternalDependencyIssue{}
	recordIssue := func(reference selectedBusinessModuleExternalReference, reason, severity string) {
		columns := make([]string, 0, len(reference.Values))
		for column := range reference.Values {
			columns = append(columns, column)
		}
		sort.Strings(columns)
		key := reference.Table + "/" + strings.Join(columns, ",") + "/" + reason
		if groups[key] == nil {
			groups[key] = &SelectedBusinessModuleExternalDependencyIssue{Table: reference.Table, ReferencedColumns: columns, Reason: reason, Severity: severity}
		}
		groups[key].References++
		if severity == "blocking" {
			report.DependencyConflicts++
		} else {
			report.DependencyReviews++
		}
	}
	for _, reference := range actual {
		identity := selectedBusinessModuleExternalReferenceIdentity(reference)
		previous, exists := expected[identity]
		delete(expected, identity)
		switch {
		case !exists:
			recordIssue(reference, "unexpected_target_reference", "blocking")
		case reference.RowSHA256 == "missing" && previous.RowSHA256 == "missing":
			if selectedChatExternalReferenceIsOptionalAbsentVisitorBrowsingContext(snapshot, observation, reference) {
				recordIssue(reference, "visitor_browsing_context_absent_at_backup_and_target", "review")
			} else {
				recordIssue(reference, "dependency_absent_at_backup_and_target", "blocking")
			}
		case reference.RowSHA256 == "missing":
			recordIssue(reference, "target_dependency_missing", "blocking")
		case previous.RowSHA256 == "missing":
			recordIssue(reference, "package_dependency_missing", "blocking")
		case reference.RowSHA256 != previous.RowSHA256:
			recordIssue(reference, "dependency_content_changed", "blocking")
		}
	}
	for _, reference := range expected {
		recordIssue(reference, "expected_reference_not_reproduced", "blocking")
	}
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		report.DependencyIssues = append(report.DependencyIssues, *groups[key])
	}
}

// A newer scheduler evaluation at the same level must not rewind the alert
// deduplication state. Both snapshot identities, provider/level and computed
// times must agree; every other state field and all business rows remain strict.
func canPreserveAdvancedSelectedOrderPaymentRiskAlertState(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, table string, row json.RawMessage) (bool, error) {
	if table != "payment_risk_alert_states" {
		return false, nil
	}
	containsOrders := false
	for _, module := range snapshot.Modules {
		containsOrders = containsOrders || module == "orders"
	}
	if !containsOrders {
		return false, nil
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(row, &fields) != nil || len(fields) != 4 {
		return false, nil
	}
	sourceID, err := productCatalogRowString(row, "current_snapshot_id")
	if err != nil {
		return false, nil
	}
	identifier, err := strconv.ParseInt(sourceID, 10, 64)
	if err != nil || identifier <= 0 {
		return false, nil
	}
	var originalSnapshot json.RawMessage
	for _, candidate := range snapshot.Tables["payment_risk_snapshots"].Rows {
		id, _ := productCatalogRowString(candidate, "id")
		if id == sourceID {
			if originalSnapshot != nil {
				return false, nil
			}
			originalSnapshot = candidate
		}
	}
	if originalSnapshot == nil {
		return false, nil
	}
	var allowed bool
	err = tx.QueryRowContext(ctx, `WITH source_state AS (SELECT * FROM jsonb_populate_record(NULL::public.payment_risk_alert_states,$1::jsonb))
SELECT EXISTS (
 SELECT 1 FROM source_state source
 JOIN public.payment_risk_alert_states current_state ON current_state.provider=source.provider
 JOIN public.payment_risk_snapshots original ON original.id=source.current_snapshot_id
 JOIN public.payment_risk_snapshots latest ON latest.id=current_state.current_snapshot_id
 WHERE (to_jsonb(current_state)-'current_snapshot_id'-'updated_at')=(to_jsonb(source)-'current_snapshot_id'-'updated_at')
 AND source.current_level IN ('normal','warning','critical')
 AND current_state.current_snapshot_id>source.current_snapshot_id AND current_state.updated_at>source.updated_at
 AND original.provider=source.provider AND latest.provider=source.provider
 AND original.level=source.current_level AND latest.level=source.current_level
 AND original.computed_at=source.updated_at AND latest.computed_at=current_state.updated_at
 AND to_jsonb(original)=$2::jsonb
)`, string(row), string(originalSnapshot)).Scan(&allowed)
	return allowed, err
}
