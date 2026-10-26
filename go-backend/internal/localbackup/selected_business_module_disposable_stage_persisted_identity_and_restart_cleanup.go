package localbackup

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/lib/pq"
)

type SelectedBusinessModuleDisposableStageIdentity struct {
	DatabaseName string `json:"database_name"`
	DatabaseOID  int64  `json:"database_oid"`
}

// The identity exists only for cleanup, never as a switch candidate. A crash
// between database allocation and recording its OID requires inspection.
func (manager *Manager) cleanupSelectedBusinessModuleDisposableStageByPersistedIdentity(ctx context.Context, task *Task) error {
	identity := task.SelectedModuleDisposableStage
	if identity == nil {
		return nil
	}
	if !strings.HasPrefix(identity.DatabaseName, "local_backup_stage_") || !safeDatabaseName.MatchString(identity.DatabaseName) || identity.DatabaseName == manager.Configuration.DatabaseName {
		return errors.New("模块演练库清理身份无效")
	}
	control, err := manager.Configuration.connectDatabase(ctx, "postgres")
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(control.Close)
	var currentOID int64
	if err := readDatabaseOID(ctx, control, identity.DatabaseName, &currentOID); err != nil {
		return err
	}
	if currentOID != 0 {
		var owner string
		if err := control.QueryRowContext(ctx, `SELECT pg_get_userbyid(datdba) FROM pg_database WHERE oid=$1`, currentOID).Scan(&owner); err != nil {
			return err
		}
		if identity.DatabaseOID <= 0 || currentOID != identity.DatabaseOID || owner != "local_backup_isolated_owner" {
			return errors.New("模块演练库 OID 或归属未确认，保留身份供检查，拒绝猜测清理")
		}
		if _, err := control.ExecContext(ctx, `DROP DATABASE `+pq.QuoteIdentifier(identity.DatabaseName)+` WITH (FORCE)`); err != nil { // #nosec G701 -- The isolated database prefix, persisted OID and owner are checked above, and its identifier is quoted.
			return err
		}
	}
	task.SelectedModuleDisposableStage = nil
	return manager.saveTask(task)
}

func (manager *Manager) cleanupSelectedBusinessModuleDisposableWorkspaceDirectories(task *Task) error {
	directory, err := manager.taskDirectory(task.ID)
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "selected-module-disposable-restore-") || strings.HasPrefix(entry.Name(), "selected-module-authenticated-package-") {
			if err := os.RemoveAll(filepath.Join(directory, entry.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}
