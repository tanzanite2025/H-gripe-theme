package localbackup

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lib/pq"
)

func (cfg Configuration) databaseConnectionURL(databaseName string) (string, error) {
	parsed, err := url.Parse(cfg.DatabaseURL)
	if err != nil || (parsed.Scheme != "postgres" && parsed.Scheme != "postgresql") || parsed.Host == "" {
		return "", errors.New("explicit PostgreSQL maintenance URL required")
	}
	parsed.Path = "/" + databaseName
	query := parsed.Query()
	query.Set("application_name", "local_computer_backup_recovery")
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func (cfg Configuration) connectDatabase(ctx context.Context, databaseName string) (*sql.DB, error) {
	connectionURL, err := cfg.databaseConnectionURL(databaseName)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("postgres", connectionURL)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(2)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, errors.New("cannot connect to configured PostgreSQL target")
	}
	return db, nil
}

func (cfg Configuration) runPostgreSQLTool(ctx context.Context, executable, databaseName string, args ...string) error {
	if executable != "pg_dump" && executable != "pg_restore" {
		return errors.New("unsupported PostgreSQL backup executable")
	}
	connectionURL, err := cfg.databaseConnectionURL(databaseName)
	if err != nil {
		return err
	}
	parsed, _ := url.Parse(connectionURL)
	password, _ := parsed.User.Password()
	parsed.User = url.User(parsed.User.Username())
	command := exec.CommandContext(ctx, executable, append([]string{"--dbname", parsed.String()}, args...)...) // #nosec G204 -- The executable is allowlisted to pg_dump/pg_restore; internal argv uses no shell and credentials stay in the environment.
	// Passwords never appear in argv, output, or task errors. No shell is used.
	command.Env = append(os.Environ(), "PGPASSWORD="+password, "PGOPTIONS=-c statement_timeout=900000 -c lock_timeout=10000")
	command.Stdout, command.Stderr = nil, nil
	if err := command.Run(); err != nil {
		return fmt.Errorf("%s failed; no online data was imported", executable)
	}
	return nil
}

func assertNoUnfencedDatabaseSessions(ctx context.Context, db *sql.DB, name string) error {
	var count int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM pg_stat_activity WHERE datname=$1 AND application_name <> 'local_computer_backup_recovery' AND backend_type='client backend'`, name).Scan(&count); err != nil {
		return err
	}
	if count != 0 {
		return errors.New("database still has unfenced client sessions; stop all writers before retrying")
	}
	return nil
}

func captureDatabaseInventory(ctx context.Context, db *sql.DB) (map[string]int64, string, int64, error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, "", 0, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(tx.Rollback)
	counts, fingerprint, migration, err := captureDatabaseTransactionInventory(ctx, tx)
	if err != nil {
		return nil, "", 0, err
	}
	return counts, fingerprint, migration, tx.Commit()
}

func captureDatabaseTransactionInventory(ctx context.Context, tx *sql.Tx) (map[string]int64, string, int64, error) {
	var unsupported bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM pg_tables WHERE schemaname NOT IN ('public','pg_catalog','information_schema') AND schemaname NOT LIKE 'pg_toast%') OR EXISTS(SELECT 1 FROM pg_extension WHERE extname IN ('pg_cron','pgagent'))`).Scan(&unsupported); err != nil {
		return nil, "", 0, err
	}
	if unsupported {
		return nil, "", 0, errors.New("unregistered schemas or database-side schedulers require a separately reviewed recovery contract")
	}
	var externalTablespaces bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=current_database() AND dattablespace<>(SELECT oid FROM pg_tablespace WHERE spcname='pg_default')) OR EXISTS(SELECT 1 FROM pg_class WHERE relnamespace='public'::regnamespace AND reltablespace NOT IN (0,(SELECT oid FROM pg_tablespace WHERE spcname='pg_default')))`).Scan(&externalTablespaces); err != nil {
		return nil, "", 0, err
	}
	if externalTablespaces {
		return nil, "", 0, errors.New("custom database tablespaces require separately mapped and validated storage capacity")
	}
	if _, err := tx.ExecContext(ctx, `SET LOCAL timezone='UTC'`); err != nil {
		return nil, "", 0, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT tablename FROM pg_tables WHERE schemaname='public' ORDER BY tablename`)
	if err != nil {
		return nil, "", 0, err
	}
	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			_ = rows.Close()
			return nil, "", 0, err
		}
		tables = append(tables, name)
	}
	readErr := rows.Err()
	_ = rows.Close()
	if readErr != nil {
		return nil, "", 0, readErr
	}
	counts := make(map[string]int64, len(tables))
	hash := sha256.New()
	for _, table := range tables {
		var count int64
		var first, second string
		query := `SELECT count(*), coalesce(sum(('x'||substr(md5(to_jsonb(record)::text),1,15))::bit(60)::bigint)::text,'0'), coalesce(sum(('x'||substr(md5(to_jsonb(record)::text),17,15))::bit(60)::bigint)::text,'0') FROM public.` + pq.QuoteIdentifier(table) + ` record`
		if err := tx.QueryRowContext(ctx, query).Scan(&count, &first, &second); err != nil {
			return nil, "", 0, err
		}
		counts[table] = count
		_, _ = fmt.Fprintf(hash, "%q:%d:%s:%s\n", table, count, first, second)
	}
	if err := appendPostgreSQLSequenceStateToDatabaseFingerprint(ctx, tx, hash); err != nil {
		return nil, "", 0, err
	}
	var migration int64
	var dirty bool
	if err := tx.QueryRowContext(ctx, `SELECT version,dirty FROM public.schema_migrations`).Scan(&migration, &dirty); err != nil {
		return nil, "", 0, errors.New("migration identity unavailable")
	}
	if dirty {
		return nil, "", 0, errors.New("dirty database migration cannot be backed up or restored")
	}
	return counts, hex.EncodeToString(hash.Sum(nil)), migration, nil
}

func (cfg Configuration) captureSnapshotAndDump(ctx context.Context, directory string, manifest *Manifest) error {
	db, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(db.Close)
	if err := assertNoUnfencedDatabaseSessions(ctx, db, cfg.DatabaseName); err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(tx.Rollback)
	if err := tx.QueryRowContext(ctx, `SELECT pg_export_snapshot()`).Scan(&manifest.SnapshotID); err != nil {
		return err
	}
	manifest.Tables, manifest.DatabaseFingerprint, manifest.MigrationVersion, err = captureDatabaseTransactionInventory(ctx, tx)
	if err != nil {
		return err
	}
	if err := tx.QueryRowContext(ctx, `SELECT pg_database_size(current_database())`).Scan(&manifest.DatabaseSizeBytes); err != nil {
		return err
	}
	if err := cfg.runPostgreSQLTool(ctx, "pg_dump", cfg.DatabaseName, "--format=custom", "--no-owner", "--no-privileges", "--no-tablespaces", "--snapshot="+manifest.SnapshotID, "--file="+filepath.Join(directory, "database.dump")); err != nil {
		return err
	}
	return tx.Commit()
}

func (cfg Configuration) restoreIsolatedDatabase(ctx context.Context, directory, name string, manifest Manifest) error {
	return cfg.restoreIsolatedDatabaseWithCreatedIdentityBinding(ctx, directory, name, manifest, nil)
}

func (cfg Configuration) restoreIsolatedDatabaseWithCreatedIdentityBinding(ctx context.Context, directory, name string, manifest Manifest, bindCreatedDatabaseIdentity func(int64) error) error {
	if err := cfg.checkPostgreSQLRestoreCapacity(manifest.DatabaseSizeBytes); err != nil {
		return err
	}
	control, err := cfg.connectDatabase(ctx, "postgres")
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(control.Close)
	if !strings.HasPrefix(name, "local_backup_stage_") || !safeDatabaseName.MatchString(name) {
		return errors.New("invalid isolated database name")
	}
	// Restored SQL runs as a dedicated unprivileged owner, never as the
	// maintenance superuser. A package cannot create superuser/event triggers.
	_, err = control.ExecContext(ctx, `DO $$ BEGIN IF NOT EXISTS(SELECT 1 FROM pg_roles WHERE rolname='local_backup_isolated_owner') THEN CREATE ROLE local_backup_isolated_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION; END IF; END $$`)
	if err != nil {
		return err
	}
	var privileged bool
	if err := control.QueryRowContext(ctx, `SELECT rolsuper OR rolcreatedb OR rolcreaterole OR rolreplication OR rolcanlogin OR rolbypassrls FROM pg_roles WHERE rolname='local_backup_isolated_owner'`).Scan(&privileged); err != nil || privileged {
		return errors.New("isolated restore owner has unsafe privileges")
	}
	if _, err = control.ExecContext(ctx, `CREATE DATABASE `+pq.QuoteIdentifier(name)+` WITH TEMPLATE template0 OWNER local_backup_isolated_owner`); err != nil { // #nosec G701 -- Dynamic database, role, table and column identifiers use pq.QuoteIdentifier; data values remain bound parameters.
		return err
	}
	// Bind only after this CREATE succeeds, before pg_restore or module writes.
	// A failed CREATE must never authorize cleanup of a same-named database.
	if bindCreatedDatabaseIdentity != nil {
		var oid int64
		if err := readDatabaseOID(ctx, control, name, &oid); err != nil {
			return err
		}
		if oid <= 0 {
			return errors.New("created isolated database identity is unavailable")
		}
		if err := bindCreatedDatabaseIdentity(oid); err != nil {
			return err
		}
	}
	if err := cfg.runPostgreSQLTool(ctx, "pg_restore", name, "--exit-on-error", "--single-transaction", "--no-owner", "--no-privileges", "--no-tablespaces", "--role=local_backup_isolated_owner", filepath.Join(directory, "database.dump")); err != nil {
		return err
	}
	restored, err := cfg.connectDatabase(ctx, name)
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(restored.Close)
	counts, restoredFingerprint, version, err := captureDatabaseInventory(ctx, restored)
	if err != nil {
		return err
	}
	if version != manifest.MigrationVersion || len(counts) != len(manifest.Tables) {
		return errors.New("isolated schema or migration mismatch")
	}
	if restoredFingerprint != manifest.DatabaseFingerprint {
		return errors.New("isolated database content fingerprint mismatch")
	}
	for table, count := range manifest.Tables {
		if counts[table] != count {
			return fmt.Errorf("isolated row count mismatch: %s", table)
		}
	}
	rows, err := restored.QueryContext(ctx, `SELECT conrelid::regclass::text,conname FROM pg_constraint WHERE contype IN ('f','c') AND connamespace='public'::regnamespace ORDER BY conrelid,conname`)
	if err != nil {
		return err
	}
	var constraints [][2]string
	for rows.Next() {
		var table, name string
		if err := rows.Scan(&table, &name); err != nil {
			_ = rows.Close()
			return err
		}
		constraints = append(constraints, [2]string{table, name})
	}
	readErr := rows.Err()
	_ = rows.Close()
	if readErr != nil {
		return readErr
	}
	for _, constraint := range constraints {
		if _, err := restored.ExecContext(ctx, `ALTER TABLE `+pq.QuoteIdentifier("public")+`.`+pq.QuoteIdentifier(strings.TrimPrefix(constraint[0], "public."))+` VALIDATE CONSTRAINT `+pq.QuoteIdentifier(constraint[1])); err != nil {
			return errors.New("isolated business constraint validation failed")
		}
	}
	return nil
}

func readRecoveryMaterials(directory string) (map[string]string, error) {
	data, err := os.ReadFile(filepath.Join(directory, "recovery-secrets.json")) // #nosec G304 -- This fixed recovery material filename is inside the validated private task workspace.
	if err != nil {
		return nil, err
	}
	var secrets map[string]string
	err = json.Unmarshal(data, &secrets)
	return secrets, err
}
