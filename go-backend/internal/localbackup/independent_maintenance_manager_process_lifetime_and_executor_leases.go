package localbackup

import (
	"errors"
	"path/filepath"
	"sync"

	"github.com/gofrs/flock"
)

// Acquire independent executor leases, dispatch persisted restart recovery and close the same leases on failure or shutdown.
type Manager struct {
	Configuration                           Configuration
	stateMutex                              sync.Mutex
	operationMutex                          sync.Mutex
	processLease                            *flock.Flock
	gateProcessLease                        *flock.Flock
	productCatalogStoppedHandoffLease       *flock.Flock
	productCatalogStoppedHandoffLeaseTaskID string
}

func NewManager(cfg Configuration) (*Manager, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	lease := flock.New(filepath.Join(cfg.Workspace, "maintenance-executor.lock"))
	locked, err := lease.TryLock()
	if err != nil || !locked {
		return nil, errors.New("another independent backup executor owns this workspace")
	}
	manager := &Manager{Configuration: cfg, processLease: lease}
	gateLease := flock.New(filepath.Join(cfg.MaintenanceDirectory, "independent-maintenance-executor.lock"))
	gateLocked, gateErr := gateLease.TryLock()
	if gateErr != nil || !gateLocked {
		_ = lease.Close()
		_ = gateLease.Close()
		return nil, errors.New("another backup executor owns the application maintenance gate")
	}
	manager.gateProcessLease = gateLease
	if err := manager.recoverPersistedMaintenanceTasksAfterExecutorRestart(); err != nil {
		_ = manager.Close()
		return nil, err
	}
	return manager, nil
}

func (manager *Manager) Close() error {
	manager.operationMutex.Lock()
	defer manager.operationMutex.Unlock()
	return errors.Join(manager.closeProductCatalogStoppedHandoffLease(), manager.processLease.Close(), manager.gateProcessLease.Close())
}
