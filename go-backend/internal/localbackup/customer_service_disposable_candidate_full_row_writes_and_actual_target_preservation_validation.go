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

const customerServiceCandidateFullRowByteLimit = 64 << 10
const customerServiceCandidateTotalFullRowByteLimit = 16 << 20

type customerServiceCandidateSelectedFullRow struct {
	Table, Identity, Digest string
	Content                 json.RawMessage
}

type customerServiceCandidateSequenceState struct {
	Name, Table, Column, Definition        string
	LastValue, Increment, Minimum, Maximum int64
	Called, Cycle                          bool
}

type customerServiceCandidateSequenceAdjustment struct {
	Sequence       string `json:"sequence"`
	PreviousValue  int64  `json:"previous_value"`
	PreviousCalled bool   `json:"previous_called"`
	CandidateValue int64  `json:"candidate_value"`
}

func customerServiceCandidateTypedPrimaryKeyPredicate(table string, primaryKeys []string) string {
	var equalities []string
	for _, column := range primaryKeys {
		quoted := pq.QuoteIdentifier(column)
		equalities = append(equalities, "record."+quoted+"=(jsonb_populate_record(NULL::public."+pq.QuoteIdentifier(table)+",$1::jsonb))."+quoted)
	}
	return strings.Join(equalities, " AND ")
}

// Full content is bounded independently of the relationship projection budget.
// Only PostgreSQL's original JSON representation is copied; report projections
// must never be used to reconstruct a business row.
func readCustomerServiceCandidateSelectedFullRows(ctx context.Context, source *moduleRecoveryRowDependencyGraph, selection *CustomerServiceEffectiveSelectionInspection) ([]customerServiceCandidateSelectedFullRow, map[string][]json.RawMessage, error) {
	if err := assertCustomerServiceCandidateColumnReferenceContractsRegistered(source.Observation); err != nil {
		return nil, nil, err
	}
	if err := assertCustomerServiceRetainedMediaOwnerColumnReferenceContractsRegistered(source.Observation); err != nil {
		return nil, nil, err
	}
	lookup := map[string]*moduleRecoveryDependencyRow{}
	for table, rows := range source.Rows {
		for identity, row := range rows {
			digest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(json.RawMessage(identity))
			if err != nil {
				return nil, nil, err
			}
			lookup[customerServiceRecoveryContentDecisionKey(table, digest)] = row
		}
	}
	var rows []customerServiceCandidateSelectedFullRow
	keys := map[string][]json.RawMessage{}
	totalBytes := 0
	for _, action := range selection.EffectiveRows {
		if action.Action != "selected_package_insert" && action.Action != "selected_package_replacement" {
			continue
		}
		row := lookup[customerServiceRecoveryContentDecisionKey(action.Table, action.IdentitySHA256)]
		if !customerServiceRecoveryPlanProposedRowReplacementAllowed(action.Table) || row == nil || !source.Scoped[action.Table][row.Identity] || row.Digest != action.SourceSHA256 || len(source.Observation.PrimaryKeys[action.Table]) == 0 {
			return nil, nil, errors.New("candidate selection is not a scoped, observed customer-service row")
		}
		var content, digest string
		// The size condition executes on the server before a body enters Go memory.
		query := `SELECT to_jsonb(record)::text,encode(sha256(convert_to(to_jsonb(record)::text,'UTF8')),'hex') FROM public.` + pq.QuoteIdentifier(action.Table) + ` record WHERE ` + customerServiceCandidateTypedPrimaryKeyPredicate(action.Table, source.Observation.PrimaryKeys[action.Table]) + ` AND octet_length(to_jsonb(record)::text)<=$2`
		if err := source.Transaction.QueryRowContext(ctx, query, row.Identity, customerServiceCandidateFullRowByteLimit).Scan(&content, &digest); err != nil {
			return nil, nil, errors.New("selected full row is absent or exceeds the candidate row size limit")
		}
		if digest != action.SourceSHA256 {
			return nil, nil, errors.New("selected full row digest changed")
		}
		totalBytes += len(content)
		if len(rows) >= 20000 || totalBytes > customerServiceCandidateTotalFullRowByteLimit {
			return nil, nil, errors.New("selected full rows exceed the candidate resource budget")
		}
		rows = append(rows, customerServiceCandidateSelectedFullRow{action.Table, row.Identity, digest, json.RawMessage(content)})
		keys[action.Table] = append(keys[action.Table], json.RawMessage(row.Identity))
	}
	order := map[string]int{"tickets": 0, "ticket_messages": 1, "customer_service_inbox_states": 2, "ticket_message_media_attachments": 3}
	sort.Slice(rows, func(i, j int) bool {
		if order[rows[i].Table] != order[rows[j].Table] {
			return order[rows[i].Table] < order[rows[j].Table]
		}
		return rows[i].Identity < rows[j].Identity
	})
	return rows, keys, nil
}

// Refuse unreviewed database code on candidate write paths. Explicit complete
// rows avoid default expressions; custom types, RLS, generated fields,
// inheritance and user functions in constraints/indexes need another contract.
func assertCustomerServiceCandidateTablesHaveSupportedWriteContracts(ctx context.Context, transaction *sql.Tx, observation *moduleRecoveryDatabaseObservation) error {
	if err := assertCustomerServiceCandidateColumnReferenceContractsRegistered(observation); err != nil {
		return err
	}
	if err := assertCustomerServiceRetainedMediaOwnerColumnReferenceContractsRegistered(observation); err != nil {
		return err
	}
	var unsupportedConstraintState bool
	if err := transaction.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM pg_constraint WHERE connamespace='public'::regnamespace AND contype IN ('f','c') AND NOT convalidated) OR EXISTS(SELECT 1 FROM pg_trigger trigger_record JOIN pg_class relation ON relation.oid=trigger_record.tgrelid WHERE relation.relnamespace='public'::regnamespace AND trigger_record.tgisinternal AND trigger_record.tgenabled NOT IN ('O','A'))`).Scan(&unsupportedConstraintState); err != nil {
		return err
	}
	if unsupportedConstraintState {
		return errors.New("candidate baseline has unvalidated constraints or disabled internal constraint triggers")
	}
	for _, table := range []string{"tickets", "ticket_messages", "customer_service_inbox_states", "ticket_message_media_attachments"} {
		if len(observation.WriteSideEffects[table]) != 0 {
			return errors.New("candidate write trigger or rewrite rule is not supported")
		}
		var unsupported bool
		query := `SELECT EXISTS(SELECT 1 FROM pg_class relation WHERE relation.oid=$1::regclass AND (relation.relkind<>'r' OR relation.relrowsecurity OR relation.relforcerowsecurity)) OR EXISTS(SELECT 1 FROM pg_inherits WHERE inhrelid=$1::regclass OR inhparent=$1::regclass) OR EXISTS(SELECT 1 FROM pg_attribute attribute JOIN pg_type type_record ON type_record.oid=attribute.atttypid WHERE attribute.attrelid=$1::regclass AND attribute.attnum>0 AND NOT attribute.attisdropped AND (attribute.attgenerated<>'' OR type_record.typnamespace<>'pg_catalog'::regnamespace)) OR EXISTS(SELECT 1 FROM pg_depend dependency JOIN pg_proc function_record ON dependency.refclassid='pg_proc'::regclass AND function_record.oid=dependency.refobjid WHERE function_record.pronamespace<>'pg_catalog'::regnamespace AND ((dependency.classid='pg_constraint'::regclass AND dependency.objid IN (SELECT oid FROM pg_constraint WHERE conrelid=$1::regclass)) OR (dependency.classid='pg_class'::regclass AND dependency.objid IN (SELECT indexrelid FROM pg_index WHERE indrelid=$1::regclass))))`
		if err := transaction.QueryRowContext(ctx, query, "public."+pq.QuoteIdentifier(table)).Scan(&unsupported); err != nil {
			return err
		}
		if unsupported {
			return errors.New("candidate table has an unsupported type, generated field, policy, inheritance or write expression")
		}
	}
	return nil
}

func writeCustomerServiceSelectedFullRowsAndValidateRetainedTables(ctx context.Context, transaction *sql.Tx, observation *moduleRecoveryDatabaseObservation, selection *CustomerServiceEffectiveSelectionInspection, rows []customerServiceCandidateSelectedFullRow, selectedKeys map[string][]json.RawMessage) error {
	if err := assertCustomerServiceCandidateTablesHaveSupportedWriteContracts(ctx, transaction, observation); err != nil {
		return err
	}
	for _, row := range rows {
		if !customerServiceRecoveryPlanProposedRowReplacementAllowed(row.Table) || len(observation.PrimaryKeys[row.Table]) == 0 {
			return errors.New("candidate writer refuses a shared table or a table without a primary key")
		}
		var columns, updates []string
		for column := range observation.Columns[row.Table] {
			columns = append(columns, pq.QuoteIdentifier(column))
		}
		sort.Strings(columns)
		for _, column := range columns {
			updates = append(updates, column+"=EXCLUDED."+column)
		}
		var primaryKeys []string
		for _, column := range observation.PrimaryKeys[row.Table] {
			primaryKeys = append(primaryKeys, pq.QuoteIdentifier(column))
		}
		query := `INSERT INTO public.` + pq.QuoteIdentifier(row.Table) + ` (` + strings.Join(columns, ",") + `) OVERRIDING SYSTEM VALUE SELECT ` + strings.Join(columns, ",") + ` FROM jsonb_populate_record(NULL::public.` + pq.QuoteIdentifier(row.Table) + `,$1::jsonb) ON CONFLICT (` + strings.Join(primaryKeys, ",") + `) DO UPDATE SET ` + strings.Join(updates, ",") // #nosec G202 -- Dynamic database, role, table and column identifiers use pq.QuoteIdentifier; data values remain bound parameters.
		if _, err := transaction.ExecContext(ctx, query, string(row.Content)); err != nil {
			return errors.New("candidate selected row violates a database constraint; rehearsal rejected")
		}
	}
	if _, err := transaction.ExecContext(ctx, `SET CONSTRAINTS ALL IMMEDIATE`); err != nil {
		return errors.New("candidate deferred constraints failed")
	}
	for _, row := range rows {
		var digest string
		if err := transaction.QueryRowContext(ctx, `SELECT encode(sha256(convert_to(to_jsonb(record)::text,'UTF8')),'hex') FROM public.`+pq.QuoteIdentifier(row.Table)+` record WHERE `+customerServiceCandidateTypedPrimaryKeyPredicate(row.Table, observation.PrimaryKeys[row.Table]), row.Identity).Scan(&digest); err != nil || digest != row.Digest {
			return errors.New("candidate full row does not match the selected source digest")
		}
	}
	if len(selection.TargetPreservation) != len(observation.Tables) {
		return errors.New("candidate retained table coverage is incomplete")
	}
	seen := map[string]bool{}
	for _, expected := range selection.TargetPreservation {
		if _, exists := observation.Tables[expected.Table]; !exists || seen[expected.Table] {
			return errors.New("candidate retained table is absent or repeated")
		}
		seen[expected.Table] = true
		actual, err := fingerprintCustomerServiceRecoveryPlanRetainedTargetTable(ctx, transaction, expected.Table, observation.Shapes[expected.Table], selectedKeys[expected.Table])
		if err != nil {
			return err
		}
		// The candidate also excludes newly inserted keys. Compare retained content
		// and count, not the number of exclusions used to obtain that content.
		if actual.RetainedRecords != expected.RetainedRecords || actual.ContentFingerprint != expected.ContentFingerprint {
			return fmt.Errorf("candidate changed retained target rows in %s", expected.Table)
		}
	}
	return nil
}

func readCustomerServiceCandidateSequenceStates(ctx context.Context, transaction *sql.Tx) ([]customerServiceCandidateSequenceState, error) {
	query := `SELECT sequence.relname,coalesce(owner.relname,''),coalesce(attribute.attname,''),jsonb_build_array(definition.seqstart,definition.seqincrement,definition.seqmax,definition.seqmin,definition.seqcache,definition.seqcycle)::text,definition.seqincrement,definition.seqmin,definition.seqmax,definition.seqcycle FROM pg_class sequence JOIN pg_sequence definition ON definition.seqrelid=sequence.oid LEFT JOIN pg_depend dependency ON dependency.classid='pg_class'::regclass AND dependency.objid=sequence.oid AND dependency.refclassid='pg_class'::regclass AND dependency.deptype IN ('a','i') LEFT JOIN pg_class owner ON owner.oid=dependency.refobjid LEFT JOIN pg_attribute attribute ON attribute.attrelid=owner.oid AND attribute.attnum=dependency.refobjsubid WHERE sequence.relnamespace='public'::regnamespace ORDER BY sequence.relname`
	result, err := transaction.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	var states []customerServiceCandidateSequenceState
	for result.Next() {
		var state customerServiceCandidateSequenceState
		if err := result.Scan(&state.Name, &state.Table, &state.Column, &state.Definition, &state.Increment, &state.Minimum, &state.Maximum, &state.Cycle); err != nil {
			_ = result.Close()
			return nil, err
		}
		states = append(states, state)
	}
	err = result.Err()
	_ = result.Close()
	if err != nil {
		return nil, err
	}
	for index := range states {
		state := &states[index]
		if err := transaction.QueryRowContext(ctx, `SELECT last_value,is_called FROM public.`+pq.QuoteIdentifier(state.Name)).Scan(&state.LastValue, &state.Called); err != nil {
			return nil, err
		}
	}
	return states, nil
}

// Advance only a sequence owned by a selected table/column, and only when an
// explicitly restored value would collide. Every other sequence stays exact.
// setval is nontransactional: any later error requires discarding this database.
func advanceCustomerServiceCandidateOwnedSequencesWithoutRewinding(ctx context.Context, transaction *sql.Tx, baseline []customerServiceCandidateSequenceState, selectedKeys map[string][]json.RawMessage) ([]customerServiceCandidateSequenceAdjustment, error) {
	before, err := readCustomerServiceCandidateSequenceStates(ctx, transaction)
	if err != nil {
		return nil, err
	}
	if !reflect.DeepEqual(before, baseline) {
		return nil, errors.New("candidate sequences changed while writing explicit rows")
	}
	var adjustments []customerServiceCandidateSequenceAdjustment
	for _, state := range before {
		if len(selectedKeys[state.Table]) == 0 {
			continue
		}
		if state.Cycle || state.Increment == 0 {
			return nil, errors.New("selected candidate sequence cycles or has an invalid increment")
		}
		sequence := "public." + pq.QuoteIdentifier(state.Name)
		aggregate := "max"
		comparison := "<="
		if state.Increment < 0 {
			aggregate = "min"
			comparison = ">="
		}
		var unsafe bool
		var bound sql.NullInt64
		query := `SELECT coalesce(last_value::numeric+CASE WHEN is_called THEN $1::numeric ELSE 0 END ` + comparison + ` boundary,false),boundary::bigint FROM public.` + pq.QuoteIdentifier(state.Name) + ` CROSS JOIN (SELECT ` + aggregate + `(` + pq.QuoteIdentifier(state.Column) + `)::numeric boundary FROM public.` + pq.QuoteIdentifier(state.Table) + `) extent`
		if err := transaction.QueryRowContext(ctx, query, state.Increment).Scan(&unsafe, &bound); err != nil {
			return nil, err
		}
		if !unsafe {
			continue
		}
		var exhausted bool
		if err := transaction.QueryRowContext(ctx, `SELECT $1::numeric+$2::numeric NOT BETWEEN $3::numeric AND $4::numeric`, bound.Int64, state.Increment, state.Minimum, state.Maximum).Scan(&exhausted); err != nil {
			return nil, err
		}
		if exhausted || !bound.Valid {
			return nil, errors.New("candidate sequence would be exhausted after selected rows")
		}
		if _, err := transaction.ExecContext(ctx, `SELECT setval($1::regclass,$2,true)`, sequence, bound.Int64); err != nil {
			return nil, err
		}
		adjustments = append(adjustments, customerServiceCandidateSequenceAdjustment{state.Name, state.LastValue, state.Called, bound.Int64})
	}
	after, err := readCustomerServiceCandidateSequenceStates(ctx, transaction)
	if err != nil {
		return nil, err
	}
	expected := append([]customerServiceCandidateSequenceState(nil), baseline...)
	for index := range expected {
		for _, adjustment := range adjustments {
			if expected[index].Name == adjustment.Sequence {
				expected[index].LastValue = adjustment.CandidateValue
				expected[index].Called = true
			}
		}
	}
	if !reflect.DeepEqual(after, expected) {
		return nil, errors.New("candidate changed an unselected sequence")
	}
	return adjustments, nil
}
