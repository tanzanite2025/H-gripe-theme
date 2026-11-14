package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migrationpostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"commerce-platform/internal/domain/shipping"
)

func createDisposableTrackingSyncLeaseDatabaseAtPreviousSQLMigration(t *testing.T) (*gorm.DB, *migrate.Migrate, string) {
	t.Helper()
	connectionURL, directory := os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL"), os.Getenv("LOCAL_BACKUP_INTEGRATION_MIGRATIONS_DIRECTORY")
	if connectionURL == "" || directory == "" || os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" {
		t.Skip("requires explicitly disposable PostgreSQL and complete SQL migrations")
	}
	control, err := sql.Open("postgres", connectionURL)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, control.Close()) })
	databaseName := fmt.Sprintf("local_backup_tracking_lease_%d", time.Now().UnixNano())
	_, err = control.ExecContext(t.Context(), `CREATE DATABASE `+pq.QuoteIdentifier(databaseName))
	require.NoError(t, err)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_, err := control.ExecContext(ctx, `DROP DATABASE `+pq.QuoteIdentifier(databaseName)+` WITH (FORCE)`)
		require.NoError(t, err)
	})
	parsed, err := url.Parse(connectionURL)
	require.NoError(t, err)
	parsed.Path = "/" + databaseName
	migrationPool, err := sql.Open("postgres", parsed.String())
	require.NoError(t, err)
	driver, err := migrationpostgres.WithInstance(migrationPool, &migrationpostgres.Config{})
	require.NoError(t, err)
	migrator, err := migrate.NewWithDatabaseInstance("file://"+filepath.ToSlash(directory), "postgres", driver)
	require.NoError(t, err)
	t.Cleanup(func() {
		sourceError, databaseError := migrator.Close()
		require.NoError(t, sourceError)
		require.NoError(t, databaseError)
	})
	require.NoError(t, migrator.Migrate(436))
	db, err := gorm.Open(postgres.Open(parsed.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	pool, err := db.DB()
	require.NoError(t, err)
	pool.SetMaxOpenConns(8)
	t.Cleanup(func() { require.NoError(t, pool.Close()) })
	return db, migrator, directory
}

func readTrackingSyncLeaseMigrationShipmentRows(t *testing.T, db *gorm.DB, omitLeaseColumns bool) string {
	t.Helper()
	row := "to_jsonb(s)"
	if omitLeaseColumns {
		row += " - ARRAY['sync_lease_owner','sync_lease_generation','sync_lease_expires_at']::text[]"
	}
	var result string
	require.NoError(t, db.Raw(`SELECT COALESCE(jsonb_agg(`+row+` ORDER BY id),'[]'::jsonb)::text FROM shipping_tracking_shipments s`).Row().Scan(&result))
	return result
}

func verifyTrackingSyncLeaseRollbackRefusesToDiscardOwnershipOrGenerations(t *testing.T, db *gorm.DB, down string) {
	t.Helper()
	before := readTrackingSyncLeaseMigrationShipmentRows(t, db, false)
	err := db.Transaction(func(tx *gorm.DB) error { return tx.Exec(down).Error })
	require.ErrorContains(t, err, "Cannot remove tracking sync leases")
	require.Equal(t, before, readTrackingSyncLeaseMigrationShipmentRows(t, db, false))
}

func TestPostgreSQLTrackingSyncLeaseSQLMigrationPreservesLegacyRowsAndEnforcesConcurrentWorkerFencing(t *testing.T) {
	db, migrator, directory := createDisposableTrackingSyncLeaseDatabaseAtPreviousSQLMigration(t)
	require.NoError(t, db.Exec(`
INSERT INTO shipping_tracking_providers(id,provider_code,provider_name,enabled) VALUES(900210,'MOCK-MIGRATION','Original tracking provider',true);
INSERT INTO shipping_tracking_shipments(id,order_id,tracking_provider_id,tracking_number,provider_carrier_code,registration_status,sync_status,event_count,last_synced_at,next_sync_at,last_error,enabled,deleted_at,created_at,updated_at)
VALUES(900210,900210,900210,'LEGACY-SYNCING','ORIGINAL','registered','syncing',7,'2025-01-02T03:04:05Z','2099-01-01','Original failed polling',true,null,'2025-01-01','2025-01-02'),
(900211,900211,900210,'PENDING-CONCURRENT','ORIGINAL','pending','pending',0,null,null,'Pending registration',true,null,'2025-01-01','2025-01-02'),
(900212,900212,900210,'SYNCED-FUTURE','ORIGINAL','registered','synced',9,'2025-01-02T03:04:05Z','2099-01-01','',true,null,'2025-01-01','2025-01-02'),
(900213,900213,900210,'DISABLED','ORIGINAL','pending','pending',0,null,null,'Disabled original',false,null,'2025-01-01','2025-01-02'),
(900214,900214,900210,'SOFT-DELETED','ORIGINAL','pending','pending',0,null,null,'Deleted original',true,'2025-06-01','2025-01-01','2025-01-02');
`).Error)
	before := readTrackingSyncLeaseMigrationShipmentRows(t, db, false)
	repo := NewShippingRepository(db)
	now := time.Now().UTC().Truncate(time.Microsecond)
	_, err := repo.ClaimDueTrackingShipments(1, now, "pre-migration-worker", time.Minute)
	var undefinedColumn interface{ SQLState() string }
	require.ErrorAs(t, err, &undefinedColumn)
	require.Equal(t, "42703", undefinedColumn.SQLState(), "reproduce missing fields without AutoMigrate")
	require.Equal(t, before, readTrackingSyncLeaseMigrationShipmentRows(t, db, false))
	require.NoError(t, migrator.Steps(1))
	version, dirty, err := migrator.Version()
	require.NoError(t, err)
	require.Equal(t, uint(437), version)
	require.False(t, dirty)
	require.Equal(t, before, readTrackingSyncLeaseMigrationShipmentRows(t, db, true), "all legacy business fields survive")
	var defaults int
	require.NoError(t, db.Raw(`SELECT count(*) FROM shipping_tracking_shipments WHERE sync_lease_owner='' AND sync_lease_generation=0 AND sync_lease_expires_at IS NULL`).Row().Scan(&defaults))
	require.Equal(t, 5, defaults)
	for _, column := range []string{"sync_lease_owner", "sync_lease_generation", "sync_lease_expires_at"} {
		require.True(t, db.Migrator().HasColumn(&shipping.TrackingShipment{}, column))
		require.True(t, db.Migrator().HasIndex(&shipping.TrackingShipment{}, "idx_shipping_tracking_shipments_"+column))
	}
	up, err := os.ReadFile(filepath.Join(directory, "437_add_shipping_tracking_sync_lease_owner_generation_and_expiry.up.sql"))
	require.NoError(t, err)
	down, err := os.ReadFile(filepath.Join(directory, "437_add_shipping_tracking_sync_lease_owner_generation_and_expiry.down.sql"))
	require.NoError(t, err)
	verifyTrackingSyncLeaseRollbackRefusesToDiscardOwnershipOrGenerations(t, db, string(down))
	claimed, err := repo.ClaimDueTrackingShipments(1, now, "first-recovery-worker", time.Minute)
	require.NoError(t, err)
	require.Len(t, claimed, 1)
	require.Equal(t, uint(900210), claimed[0].ID, "legacy syncing row without ownership can be recovered")
	require.Equal(t, int64(1), claimed[0].SyncLeaseGeneration)
	require.Equal(t, "first-recovery-worker", claimed[0].SyncLeaseOwner)
	claimedRows := readTrackingSyncLeaseMigrationShipmentRows(t, db, false)
	require.NoError(t, db.Exec(string(up)).Error)
	require.Equal(t, claimedRows, readTrackingSyncLeaseMigrationShipmentRows(t, db, false), "rerunning migration preserves active leases")
	_, err = repo.ClaimTrackingShipmentForSync(900210, "LEGACY-SYNCING", now.Add(10*time.Second), "competing-worker", time.Minute)
	require.ErrorIs(t, err, ErrTrackingSyncLeaseUnavailable)

	t.Run("concurrent_manual_claim_has_one_owner", func(t *testing.T) {
		type claimResult struct {
			shipment *shipping.TrackingShipment
			err      error
		}
		start, results := make(chan struct{}), make(chan claimResult, 2)
		for _, owner := range []string{"concurrent-worker-one", "concurrent-worker-two"} {
			go func() {
				<-start
				shipment, err := repo.ClaimTrackingShipmentForSync(900211, "PENDING-CONCURRENT", now, owner, time.Minute)
				results <- claimResult{shipment, err}
			}()
		}
		close(start)
		var successful []*shipping.TrackingShipment
		for range 2 {
			result := <-results
			if result.err == nil {
				successful = append(successful, result.shipment)
			} else {
				require.True(t, errors.Is(result.err, ErrTrackingSyncLeaseUnavailable) || errors.Is(result.err, gorm.ErrRecordNotFound), "%v", result.err)
			}
		}
		require.Len(t, successful, 1)
		require.Equal(t, int64(1), successful[0].SyncLeaseGeneration)
		require.NoError(t, repo.RenewTrackingShipmentSyncLease(successful[0], now.Add(20*time.Second), time.Minute))
		require.NoError(t, repo.ApplyTrackingWebhookSyncSuccess(900211, "PENDING-CONCURRENT", 11, nil, now.Add(30*time.Second)))
		beforeLateWorker := readTrackingSyncLeaseMigrationShipmentRows(t, db, false)
		require.ErrorIs(t, repo.CompleteTrackingShipmentSync(successful[0], 99, nil, nil, now.Add(40*time.Second)), ErrTrackingSyncLeaseLost)
		require.Equal(t, beforeLateWorker, readTrackingSyncLeaseMigrationShipmentRows(t, db, false))
	})

	t.Run("expired_owner_cannot_overwrite_reclaimed_result", func(t *testing.T) {
		replacement, err := repo.ClaimTrackingShipmentForSync(900210, "LEGACY-SYNCING", now.Add(2*time.Minute), "replacement-worker", time.Minute)
		require.NoError(t, err)
		require.Equal(t, int64(2), replacement.SyncLeaseGeneration)
		beforeStaleWrites := readTrackingSyncLeaseMigrationShipmentRows(t, db, false)
		require.ErrorIs(t, repo.RenewTrackingShipmentSyncLease(&claimed[0], now.Add(2*time.Minute), time.Minute), ErrTrackingSyncLeaseLost)
		require.ErrorIs(t, repo.CompleteTrackingShipmentSync(&claimed[0], 99, nil, nil, now.Add(2*time.Minute)), ErrTrackingSyncLeaseLost)
		require.ErrorIs(t, repo.FailTrackingShipmentSync(&claimed[0], "stale failure", nil, now.Add(2*time.Minute)), ErrTrackingSyncLeaseLost)
		require.Equal(t, beforeStaleWrites, readTrackingSyncLeaseMigrationShipmentRows(t, db, false))
		require.NoError(t, repo.CompleteTrackingShipmentSync(replacement, 8, nil, nil, now.Add(2*time.Minute+time.Second)))
		verifyTrackingSyncLeaseRollbackRefusesToDiscardOwnershipOrGenerations(t, db, string(down))
	})

	t.Run("due_scan_preserves_future_disabled_and_deleted_shipments", func(t *testing.T) {
		beforeScan := readTrackingSyncLeaseMigrationShipmentRows(t, db, false)
		due, err := repo.ClaimDueTrackingShipments(100, now.Add(5*time.Minute), "later-worker", time.Minute)
		require.NoError(t, err)
		require.Empty(t, due)
		require.Equal(t, beforeScan, readTrackingSyncLeaseMigrationShipmentRows(t, db, false))
	})
}

func TestPostgreSQLTrackingSyncLeaseSQLMigrationAllowsUnusedRollbackAndReapplicationWithoutChangingBusinessRows(t *testing.T) {
	db, migrator, _ := createDisposableTrackingSyncLeaseDatabaseAtPreviousSQLMigration(t)
	require.NoError(t, db.Exec(`
INSERT INTO shipping_tracking_providers(id,provider_code,provider_name) VALUES(900210,'MOCK-MIGRATION','Original provider');
INSERT INTO shipping_tracking_shipments(id,order_id,tracking_provider_id,tracking_number,provider_carrier_code,sync_status,event_count,last_error)
VALUES(900210,900210,900210,'UNUSED-LEASE','ORIGINAL','synced',7,'Original error');
`).Error)
	before := readTrackingSyncLeaseMigrationShipmentRows(t, db, false)
	require.NoError(t, migrator.Steps(1))
	require.NoError(t, migrator.Steps(-1))
	version, dirty, err := migrator.Version()
	require.NoError(t, err)
	require.Equal(t, uint(436), version)
	require.False(t, dirty)
	require.Equal(t, before, readTrackingSyncLeaseMigrationShipmentRows(t, db, false))
	for _, column := range []string{"sync_lease_owner", "sync_lease_generation", "sync_lease_expires_at"} {
		require.False(t, db.Migrator().HasColumn(&shipping.TrackingShipment{}, column))
	}
	require.NoError(t, migrator.Steps(1))
	require.Equal(t, before, readTrackingSyncLeaseMigrationShipmentRows(t, db, true))
	claim, err := NewShippingRepository(db).ClaimTrackingShipmentForSync(900210, "UNUSED-LEASE", time.Now().UTC(), "after-reapplication", time.Minute)
	require.NoError(t, err)
	require.Equal(t, int64(1), claim.SyncLeaseGeneration)
}
