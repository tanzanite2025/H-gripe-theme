package localbackup

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

func validateBusinessModuleNormalForeignKeyEnforcementBeforeWriting(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot) error {
	var role string
	if err := tx.QueryRowContext(ctx, `SHOW session_replication_role`).Scan(&role); err != nil {
		return err
	}
	if role != "origin" {
		return errors.New("business module restore requires origin replication role and active database constraints")
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
		return errors.New("business module restore requires validated foreign keys with every normal enforcement trigger enabled")
	}
	return nil
}
