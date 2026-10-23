package localbackup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/lib/pq"
)

const minimumPostgreSQLFreeReserveBytes uint64 = 256 << 20

func (cfg Configuration) checkPostgreSQLRestoreCapacity(sourceDatabaseBytes int64) error {
	if sourceDatabaseBytes <= 0 || uint64(sourceDatabaseBytes) > (math.MaxUint64-minimumPostgreSQLFreeReserveBytes)/2 {
		return errors.New("package database size is missing or invalid; regenerate the backup with the current protocol")
	}
	// Keep the current target, stage database, indexes and WAL independently.
	// This is an estimate, never a reservation or a guarantee against other
	// applications filling the filesystem after the preflight.
	required := uint64(sourceDatabaseBytes)*2 + minimumPostgreSQLFreeReserveBytes
	return cfg.checkPostgreSQLAvailableBytes(required)
}

func (cfg Configuration) checkPostgreSQLAvailableBytes(required uint64) error {
	free, err := availableDiskBytes(cfg.PostgreSQLCapacityDirectory)
	if err != nil {
		return errors.New("cannot inspect the explicitly mapped PostgreSQL capacity filesystem")
	}
	if free < required {
		return fmt.Errorf("insufficient PostgreSQL volume capacity: need at least %d free bytes for recovery and reserve, available %d", required, free)
	}
	return nil
}

func findOwnedUniqueSequenceCollisionOrExhaustionGaps(ctx context.Context, db *sql.DB, manifest *Manifest) error {
	rows, err := db.QueryContext(ctx, `SELECT sequence.relname,owner.relname,attribute.attname FROM pg_depend dependency JOIN pg_class sequence ON sequence.oid=dependency.objid AND sequence.relkind='S' JOIN pg_class owner ON owner.oid=dependency.refobjid JOIN pg_attribute attribute ON attribute.attrelid=owner.oid AND attribute.attnum=dependency.refobjsubid WHERE dependency.classid='pg_class'::regclass AND dependency.refclassid='pg_class'::regclass AND dependency.deptype IN ('a','i') AND sequence.relnamespace='public'::regnamespace AND owner.relnamespace='public'::regnamespace AND EXISTS(SELECT 1 FROM pg_constraint constraint_record WHERE constraint_record.conrelid=owner.oid AND constraint_record.contype IN ('p','u') AND array_length(constraint_record.conkey,1)=1 AND attribute.attnum=ANY(constraint_record.conkey)) ORDER BY sequence.relname`)
	if err != nil {
		return err
	}
	var ownedSequences [][3]string
	for rows.Next() {
		var entry [3]string
		if err := rows.Scan(&entry[0], &entry[1], &entry[2]); err != nil {
			_ = rows.Close()
			return err
		}
		ownedSequences = append(ownedSequences, entry)
	}
	readErr := rows.Err()
	_ = rows.Close()
	if readErr != nil {
		return readErr
	}
	for _, entry := range ownedSequences {
		sequence := `public.` + pq.QuoteIdentifier(entry[0])
		table := `public.` + pq.QuoteIdentifier(entry[1])
		column := pq.QuoteIdentifier(entry[2])
		var unsafe bool
		query := `SELECT seqcycle OR CASE WHEN seqincrement>0 THEN candidate>seqmax OR coalesce(candidate<=(SELECT max(` + column + `)::numeric FROM ` + table + `),false) ELSE candidate<seqmin OR coalesce(candidate>=(SELECT min(` + column + `)::numeric FROM ` + table + `),false) END FROM pg_sequence CROSS JOIN LATERAL (SELECT last_value::numeric+CASE WHEN is_called THEN seqincrement::numeric ELSE 0 END candidate FROM ` + sequence + `) state WHERE seqrelid=$1::regclass`
		if err := db.QueryRowContext(ctx, query, sequence).Scan(&unsafe); err != nil {
			return err
		}
		if unsafe {
			manifest.Gaps = append(manifest.Gaps, "unsafe_owned_unique_sequence:"+entry[0]+":"+entry[1]+"."+entry[2])
		}
	}
	return nil
}

func appendPostgreSQLSequenceStateToDatabaseFingerprint(ctx context.Context, tx *sql.Tx, fingerprint io.Writer) error {
	rows, err := tx.QueryContext(ctx, `SELECT sequencename FROM pg_sequences WHERE schemaname='public' ORDER BY sequencename`)
	if err != nil {
		return err
	}
	var sequences []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			_ = rows.Close()
			return err
		}
		sequences = append(sequences, name)
	}
	readErr := rows.Err()
	_ = rows.Close()
	if readErr != nil {
		return readErr
	}
	for _, name := range sequences {
		var lastValue int64
		var called bool
		if err := tx.QueryRowContext(ctx, `SELECT last_value,is_called FROM public.`+pq.QuoteIdentifier(name)).Scan(&lastValue, &called); err != nil {
			return err
		}
		var definition string
		if err := tx.QueryRowContext(ctx, `SELECT jsonb_build_array(seqstart,seqincrement,seqmax,seqmin,seqcache,seqcycle)::text FROM pg_sequence WHERE seqrelid=$1::regclass`, pq.QuoteIdentifier("public")+"."+pq.QuoteIdentifier(name)).Scan(&definition); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(fingerprint, "sequence:%q:%d:%t:%s\n", name, lastValue, called, definition)
	}
	return nil
}
