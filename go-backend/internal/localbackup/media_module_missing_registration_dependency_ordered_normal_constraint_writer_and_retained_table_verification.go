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

func validateMediaModuleNormalForeignKeyEnforcementBeforeWriting(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot) error {
	var role string
	if err := tx.QueryRowContext(ctx, `SHOW session_replication_role`).Scan(&role); err != nil {
		return err
	}
	if role != "origin" {
		return errors.New("media restore requires origin replication role and active database constraints")
	}
	var affectedTables []string
	for table := range snapshot.Tables {
		affectedTables = append(affectedTables, "public."+pq.QuoteIdentifier(table))
	}
	var unsafe bool
	// PostgreSQL ordinary-table FKs have two referencing and two referenced
	// internal triggers. A disabled/replica-only/missing guard is not repaired
	// silently, and existing constraints are never dropped and recreated here.
	query := `SELECT EXISTS(SELECT 1 FROM pg_constraint c WHERE c.contype='f' AND (c.conrelid=ANY($1::regclass[]) OR c.confrelid=ANY($1::regclass[])) AND (NOT c.convalidated OR (SELECT count(*) FROM pg_trigger t WHERE t.tgconstraint=c.oid AND t.tgisinternal AND t.tgenabled IN ('O','A'))<>4))`
	if err := tx.QueryRowContext(ctx, query, pq.Array(affectedTables)).Scan(&unsafe); err != nil {
		return err
	}
	if unsafe {
		return errors.New("media restore requires validated foreign keys with every normal enforcement trigger enabled")
	}
	return nil
}

// INSERT only, in dependency order, with ordinary FK/check/unique constraints
// and triggers enabled. No upload, derivative generation or cleanup service runs.
func restoreMediaModuleMissingRowsWithNormalConstraintsAndVerifiedPreservation(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, observation *moduleRecoveryDatabaseObservation, missing map[string][]json.RawMessage, transactionalSequences bool) (int, error) {
	if !reflect.DeepEqual(snapshot.Modules, []string{"media"}) {
		return 0, errors.New("单独媒体包才可使用媒体恢复 writer")
	}
	if err := validateMediaModuleNormalForeignKeyEnforcementBeforeWriting(ctx, tx, snapshot); err != nil {
		return 0, err
	}
	for table, rows := range missing {
		if len(rows) > 0 && !mediaModuleHasDedicatedMissingRowContract(snapshot, table) {
			return 0, errors.New("媒体恢复写入范围越界")
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
	for _, table := range mediaModuleOwnedTableNames() {
		if len(missing[table]) == 0 {
			continue
		}
		if err := validateMediaModuleInsertOnlyWriteContract(ctx, tx, snapshot, table, observation); err != nil {
			return 0, err
		}
		var columns []string
		for column := range observation.Columns[table] {
			columns = append(columns, pq.QuoteIdentifier(column))
		}
		sort.Strings(columns)
		list := strings.Join(columns, ",")
		for _, row := range missing[table] {
			if _, err := tx.ExecContext(ctx, `INSERT INTO public.`+pq.QuoteIdentifier(table)+` (`+list+`) SELECT `+list+` FROM jsonb_populate_record(NULL::public.`+pq.QuoteIdentifier(table)+`,$1::jsonb)`, string(row)); err != nil { // #nosec G202 -- Table and column identifiers use pq.QuoteIdentifier from the validated schema; complete row values use bound JSON parameters.
				return 0, fmt.Errorf("媒体完整行恢复违反身份或业务约束：%s", table)
			}
			var same bool
			if err := tx.QueryRowContext(ctx, `SELECT to_jsonb(record)=$1::jsonb FROM public.`+pq.QuoteIdentifier(table)+` record WHERE `+businessModuleCompleteRowTypedPrimaryKeyPredicate(table, observation.PrimaryKeys[table]), string(row)).Scan(&same); err != nil || !same {
				return 0, fmt.Errorf("媒体完整行写入后无法与原包核对一致：%s", table)
			}
			restored++
		}
	}
	if _, err := tx.ExecContext(ctx, `SET CONSTRAINTS ALL IMMEDIATE`); err != nil {
		return 0, err
	}
	if err := validateProductCatalogAffectedForeignKeys(ctx, tx, productCatalogModuleSnapshot{Tables: snapshot.Tables}); err != nil {
		return 0, err
	}
	for table, expected := range before {
		actual, err := fingerprintBusinessModuleRetainedTargetTable(ctx, tx, table, observation.Shapes[table], missing[table])
		if err != nil {
			return 0, err
		}
		if actual.ContentFingerprint != expected.ContentFingerprint || actual.RetainedRecords != expected.RetainedRecords {
			return 0, fmt.Errorf("媒体恢复改变了当前保留记录：%s", table)
		}
	}
	if transactionalSequences {
		err = advanceBusinessModuleRestoreOwnedSequencesTransactionallyAndVerifyPreservation(ctx, tx, sequences, missing)
	} else {
		_, err = advanceCustomerServiceCandidateOwnedSequencesWithoutRewinding(ctx, tx, sequences, missing)
	}
	return restored, err
}
