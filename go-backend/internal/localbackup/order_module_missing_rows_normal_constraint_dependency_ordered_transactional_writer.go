package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/lib/pq"
)

// Use the same database-enforced writer for order rehearsals and live restores.
// Dependencies within a table are checked at the end of its single INSERT;
// dependencies between tables require a parent-first order. Cycles are refused
// rather than disabling triggers or changing persisted constraints.
func planOrderModuleMissingTableInsertsWithNormalForeignKeyEnforcement(missing map[string][]json.RawMessage, references []ModuleRecoveryDependencyReference) ([]string, error) {
	parents := map[string]map[string]bool{}
	for table, rows := range missing {
		if len(rows) > 0 {
			parents[table] = map[string]bool{}
		}
	}
	addDependency := func(child, parent string) {
		if child != parent && parents[child] != nil && parents[parent] != nil {
			parents[child][parent] = true
		}
	}
	for _, reference := range references {
		if reference.Kind == "foreign_key" {
			addDependency(reference.SourceTable, reference.TargetTable)
		}
	}
	// These registered ownership references are not all protected by SQL FKs.
	for _, reference := range [][2]string{
		{"order_attributions", "orders"}, {"order_idempotencies", "orders"},
		{"shipment_records", "orders"}, {"transactions", "orders"},
		{"refunds", "orders"}, {"refunds", "transactions"},
		{"refund_idempotencies", "refunds"}, {"payment_refund_executions", "refunds"},
		{"payment_refund_executions", "transactions"},
		{"outbox_events", "orders"}, {"outbox_events", "refunds"},
		{"email_delivery_records", "outbox_events"},
	} {
		addDependency(reference[0], reference[1])
	}
	var ordered []string
	for len(parents) > 0 {
		var ready []string
		for table, dependencies := range parents {
			if len(dependencies) == 0 {
				ready = append(ready, table)
			}
		}
		if len(ready) == 0 {
			return nil, errors.New("order restore has a cyclic table dependency without a normal-constraint insert contract")
		}
		sort.Strings(ready)
		ordered = append(ordered, ready...)
		for _, table := range ready {
			delete(parents, table)
			for _, dependencies := range parents {
				delete(dependencies, table)
			}
		}
	}
	return ordered, nil
}

// The caller owns the serializable transaction and commits rows, sequences and
// audit together. This function never changes replication role or constraint
// definitions and does not invoke registration, payment or notification jobs.
func restoreOrderModuleMissingRowsWithNormalConstraintsTransactionalSequencesAndVerifiedPreservation(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, observation *moduleRecoveryDatabaseObservation, missing map[string][]json.RawMessage) (int, error) {
	if !reflect.DeepEqual(snapshot.Modules, []string{"orders"}) {
		return 0, errors.New("normal-constraint order writer only accepts a standalone order package")
	}
	if err := validateBusinessModuleNormalForeignKeyEnforcementBeforeWriting(ctx, tx, snapshot); err != nil {
		return 0, err
	}
	ordered, err := planOrderModuleMissingTableInsertsWithNormalForeignKeyEnforcement(missing, observation.References)
	if err != nil {
		return 0, err
	}
	// The caller already inspected every package row in this transaction. A
	// no-op must not repeat all retained-table scans or advance any sequence.
	if len(ordered) == 0 {
		return 0, nil
	}
	for _, table := range ordered {
		if _, exists := snapshot.Tables[table]; !exists || len(observation.Columns[table]) == 0 || len(observation.PrimaryKeys[table]) == 0 {
			return 0, errors.New("order restore table has no authenticated full-row or primary-key contract")
		}
		if err := validateOrderModuleInsertOnlyWriteContract(ctx, tx, snapshot, table, observation); err != nil {
			return 0, err
		}
	}
	sequences, err := readBusinessModuleRestoreSequenceStates(ctx, tx)
	if err != nil {
		return 0, err
	}
	before := map[string]BusinessModuleRestoreTablePreservation{}
	for table := range observation.Tables {
		preserved, err := fingerprintBusinessModuleRetainedTargetTable(ctx, tx, table, observation.Shapes[table], missing[table])
		if err != nil {
			return 0, err
		}
		before[table] = preserved
	}
	restored := 0
	for _, table := range ordered {
		var columns []string
		for column := range observation.Columns[table] {
			columns = append(columns, pq.QuoteIdentifier(column))
		}
		sort.Strings(columns)
		list := strings.Join(columns, ",")
		rows, err := json.Marshal(missing[table])
		if err != nil {
			return 0, err
		}
		result, err := tx.ExecContext(ctx, `INSERT INTO public.`+pq.QuoteIdentifier(table)+` (`+list+`) SELECT `+list+` FROM jsonb_populate_recordset(NULL::public.`+pq.QuoteIdentifier(table)+`,$1::jsonb)`, string(rows)) // #nosec G202 -- Table and column identifiers use pq.QuoteIdentifier from the validated schema; complete row values use bound JSON parameters.
		if err != nil {
			return 0, fmt.Errorf("order restore violates an enforced database constraint: %s", table)
		}
		inserted, err := result.RowsAffected()
		if err != nil || inserted != int64(len(missing[table])) {
			return 0, errors.New("order restore full-row insert count differs from authenticated missing rows")
		}
		restored += int(inserted)
	}
	// Also force originally deferrable constraints to report any violation
	// before the result fingerprint and atomic audit receipt are prepared.
	if _, err := tx.ExecContext(ctx, `SET CONSTRAINTS ALL IMMEDIATE`); err != nil {
		return 0, errors.New("order restore violates a deferred database constraint")
	}
	for table, contents := range snapshot.Tables {
		for _, row := range contents.Rows {
			var same bool
			query := `SELECT to_jsonb(record)=$1::jsonb FROM public.` + pq.QuoteIdentifier(table) + ` record WHERE ` + businessModuleCompleteRowTypedPrimaryKeyPredicate(table, observation.PrimaryKeys[table])
			if err := tx.QueryRowContext(ctx, query, string(row)).Scan(&same); err != nil {
				return 0, fmt.Errorf("order restored complete row could not be verified: %s", table)
			}
			if same {
				continue
			}
			preserveRuntime, err := canPreserveAdvancedSelectedOrderPaymentRiskAlertState(ctx, tx, snapshot, table, row)
			if err != nil {
				return 0, err
			}
			preserveNotes, err := canPreserveCurrentRestorableOrderNotesWithOtherwiseIdenticalFacts(ctx, tx, snapshot, table, row)
			if err != nil {
				return 0, err
			}
			if !preserveRuntime && !preserveNotes {
				return 0, fmt.Errorf("order restored complete row differs from the package: %s", table)
			}
		}
	}
	for table, expected := range before {
		actual, err := fingerprintBusinessModuleRetainedTargetTable(ctx, tx, table, observation.Shapes[table], missing[table])
		if err != nil {
			return 0, err
		}
		if actual.ContentFingerprint != expected.ContentFingerprint || actual.RetainedRecords != expected.RetainedRecords {
			return 0, fmt.Errorf("order restore changed retained target rows: %s", table)
		}
	}
	if err := advanceBusinessModuleRestoreOwnedSequencesTransactionallyAndVerifyPreservation(ctx, tx, sequences, missing); err != nil {
		return 0, err
	}
	return restored, nil
}
