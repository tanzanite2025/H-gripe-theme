package localbackup

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

type ReconciliationReport struct {
	CapturedAt                       time.Time                                 `json:"captured_at"`
	DatabaseFingerprint              string                                    `json:"database_fingerprint"`
	Tables                           map[string]int64                          `json:"tables"`
	MigrationVersion                 int64                                     `json:"migration_version"`
	FinancialTotalsByCurrency        map[string]map[string]string              `json:"financial_totals_by_currency"`
	Gaps                             []string                                  `json:"gaps"`
	MediaFingerprint                 string                                    `json:"media_fingerprint"`
	ExternalProjectionReconciliation *ExternalProjectionReconciliationContract `json:"external_projection_reconciliation"`
}

func (manager *Manager) OpenReconciliationConnections(id, confirmation string) error {
	if !manager.operationMutex.TryLock() {
		return errors.New("an operation is running")
	}
	defer manager.operationMutex.Unlock()
	task, err := manager.GetTask(id)
	if err != nil {
		return err
	}
	cfg := manager.Configuration
	if task.State != "awaiting_reconciliation" || !taskHasSwitchedWholeApplicationRecoveryJournal(task) || confirmation != "RECONCILE "+cfg.SiteID+"/"+cfg.Environment+"/"+id {
		return errors.New("controlled reconciliation confirmation required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	lease, err := manager.acquireMaintenance(ctx, task)
	if err != nil {
		return err
	}
	defer lease.Close()
	control, err := cfg.connectDatabase(ctx, "postgres")
	if err != nil {
		return err
	}
	defer control.Close()
	var oid int64
	if err := readDatabaseOID(ctx, control, cfg.DatabaseName, &oid); err != nil {
		return err
	}
	if oid != task.Journal.NewDatabaseOID {
		return errors.New("switched database identity changed")
	}
	// The application process fence remains in place. Only deliberately opened
	// administrative SQL sessions may reconcile facts; all must close on resume.
	task.ReconciliationConnectionsEnabled = true
	task.Phase = "administrative_reconciliation_connections_enabled_application_stopped"
	if err := manager.saveTask(task); err != nil {
		return err
	}
	_, err = control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName)+` ALLOW_CONNECTIONS true`)
	return err
}

func (manager *Manager) InspectReconciliation(id string) (*ReconciliationReport, error) {
	if !manager.operationMutex.TryLock() {
		return nil, errors.New("an operation is running")
	}
	defer manager.operationMutex.Unlock()
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	if task.State != "awaiting_reconciliation" || !taskHasSwitchedWholeApplicationRecoveryJournal(task) {
		return nil, errors.New("a paired recovery awaiting reconciliation is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	lease, err := manager.acquireMaintenance(ctx, task)
	if err != nil {
		return nil, err
	}
	defer lease.Close()
	return manager.inspectReconciliationUnderMaintenance(ctx, task)
}

func (manager *Manager) inspectReconciliationUnderMaintenance(ctx context.Context, task *Task) (*ReconciliationReport, error) {
	cfg := manager.Configuration
	control, err := cfg.connectDatabase(ctx, "postgres")
	if err != nil {
		return nil, err
	}
	defer control.Close()
	var oid int64
	if err := readDatabaseOID(ctx, control, cfg.DatabaseName, &oid); err != nil {
		return nil, err
	}
	if oid != task.Journal.NewDatabaseOID {
		return nil, errors.New("switched database identity changed")
	}
	if err := assertNoUnfencedDatabaseSessions(ctx, control, cfg.DatabaseName); err != nil {
		return nil, err
	}
	if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName)+` ALLOW_CONNECTIONS true`); err != nil {
		return nil, err
	}
	db, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	// Sealing affects new logins, not this already-open inspection connection.
	// It prevents the final report from racing another administrative session.
	if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName)+` ALLOW_CONNECTIONS false`); err != nil {
		return nil, err
	}
	defer func() {
		if task.ReconciliationConnectionsEnabled {
			reopenContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			control.ExecContext(reopenContext, `ALTER DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName)+` ALLOW_CONNECTIONS true`)
		}
	}()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	report := &ReconciliationReport{CapturedAt: time.Now().UTC(), FinancialTotalsByCurrency: map[string]map[string]string{}, Gaps: []string{}}
	report.Tables, report.DatabaseFingerprint, report.MigrationVersion, err = captureDatabaseTransactionInventory(ctx, tx)
	if err != nil {
		return nil, err
	}
	for table, column := range map[string]string{"orders": "total_minor", "transactions": "amount_minor", "refunds": "amount_minor", "payment_refund_executions": "amount_minor"} {
		if _, exists := report.Tables[table]; !exists {
			continue
		}
		rows, err := tx.QueryContext(ctx, `SELECT coalesce(to_jsonb(record)->>'currency','unspecified'),coalesce(sum((to_jsonb(record)->>$1)::numeric),0)::text FROM public.`+pq.QuoteIdentifier(table)+` record GROUP BY 1`, column)
		if err != nil {
			return nil, err
		}
		totals := map[string]string{}
		for rows.Next() {
			var currency, total string
			if err := rows.Scan(&currency, &total); err != nil {
				rows.Close()
				return nil, err
			}
			totals[currency] = total
		}
		readErr := rows.Err()
		rows.Close()
		if readErr != nil {
			return nil, readErr
		}
		report.FinancialTotalsByCurrency[table+"."+column] = totals
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	inspection := *task.Manifest
	inspection.Gaps = []string{}
	inspection.Tables = report.Tables
	if err := cfg.findMediaReferenceGaps(ctx, db, &inspection); err != nil {
		return nil, err
	}
	if err := findBusinessRelationshipGaps(ctx, db, &inspection); err != nil {
		return nil, err
	}
	report.Gaps = inspection.Gaps
	report.MediaFingerprint, err = cfg.mediaFingerprint(ctx)
	if err != nil {
		return nil, err
	}
	if report.MediaFingerprint != fingerprintManifestMedia(*task.Manifest) {
		report.Gaps = append(report.Gaps, "restored_media_differs_from_verified_package")
	}
	if wholeApplicationPackageIncludesCompanyStatic(task.Manifest) {
		fingerprint, err := cfg.companyStaticReleaseFileFingerprint(ctx)
		if err != nil {
			return nil, err
		}
		if fingerprint != fingerprintCompanyStaticManifestFiles(task.Manifest.Files) {
			report.Gaps = append(report.Gaps, "restored_company_static_files_differ_from_verified_package")
		}
	}
	report.ExternalProjectionReconciliation = manager.buildTaskBoundExternalProjectionReconciliationContract(task, report)
	return report, nil
}

func findBusinessRelationshipGaps(ctx context.Context, db *sql.DB, manifest *Manifest) error {
	for _, relationship := range [][3]string{{"order_items", "order_id", "orders"}, {"orders", "user_id", "users"}, {"transactions", "order_id", "orders"}, {"refunds", "order_id", "orders"}, {"payment_refund_executions", "order_id", "orders"}, {"ticket_messages", "ticket_id", "tickets"}, {"ticket_message_media_attachments", "ticket_message_id", "ticket_messages"}, {"ticket_message_media_attachments", "media_asset_id", "media_assets"}, {"media_asset_derivatives", "media_asset_id", "media_assets"}, {"order_evidence_attachments", "evidence_item_id", "order_evidence_items"}, {"after_sales_case_attachments", "case_id", "after_sales_cases"}} {
		if _, ok := manifest.Tables[relationship[0]]; !ok {
			continue
		}
		if _, ok := manifest.Tables[relationship[2]]; !ok {
			manifest.Gaps = append(manifest.Gaps, "missing_dependency_table:"+relationship[2])
			continue
		}
		var missing int64
		optionalZero := ""
		if relationship[0] == "orders" && relationship[1] == "user_id" {
			optionalZero = ` AND child.user_id<>0`
		}
		query := `SELECT count(*) FROM public.` + pq.QuoteIdentifier(relationship[0]) + ` child LEFT JOIN public.` + pq.QuoteIdentifier(relationship[2]) + ` parent ON parent.id=child.` + pq.QuoteIdentifier(relationship[1]) + ` WHERE child.` + pq.QuoteIdentifier(relationship[1]) + ` IS NOT NULL` + optionalZero + ` AND parent.id IS NULL`
		if err := db.QueryRowContext(ctx, query).Scan(&missing); err != nil {
			return err
		}
		if missing > 0 {
			manifest.Gaps = append(manifest.Gaps, "orphan_business_relationship:"+relationship[0]+"."+relationship[1])
		}
	}
	return findOwnedUniqueSequenceCollisionOrExhaustionGaps(ctx, db, manifest)
}
