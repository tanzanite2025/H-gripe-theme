package localbackup

import (
	"context"
	"database/sql"
	"errors"
	"maps"

	"github.com/lib/pq"
)

// Domain contracts decide whether their INSERT-only restore can retain existing
// UPDATE-only guards. This shared check never grants that eligibility itself.
// INSERT triggers/rules, including disabled ones, remain prohibited. Historical
// diagnostic text is retained so this extraction does not alter task errors.
func validateBusinessModuleInsertOnlyTableContractWithUpdateOnlyGuardsPreserved(ctx context.Context, tx *sql.Tx, table string, observation *moduleRecoveryDatabaseObservation) error {
	var insertSideEffects bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM pg_trigger WHERE tgrelid=$1::regclass AND NOT tgisinternal AND (tgtype::int & 4)<>0) OR EXISTS(SELECT 1 FROM pg_rewrite WHERE ev_class=$1::regclass AND ev_type='3')`, "public."+pq.QuoteIdentifier(table)).Scan(&insertSideEffects); err != nil {
		return err
	}
	if insertSideEffects {
		return errors.New("order history table has INSERT triggers or rules without a recovery contract")
	}
	insertObservation := *observation
	insertObservation.WriteSideEffects = maps.Clone(observation.WriteSideEffects)
	delete(insertObservation.WriteSideEffects, table)
	return validateSelectedBusinessModuleIsolatedInsertWriteContract(ctx, tx, table, &insertObservation)
}
