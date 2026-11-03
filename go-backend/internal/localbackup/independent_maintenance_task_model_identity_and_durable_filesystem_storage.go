package localbackup

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Persist the original task JSON and identities without changing field order or timestamps.
type Task struct {
	ConfigurationModuleSiteContentRestore        *BusinessModuleTransactionalRestoreProgress            `json:"configuration_site_content_restore,omitempty"`
	MediaModuleRestore                           *BusinessModuleTransactionalRestoreProgress            `json:"media_module_restore,omitempty"`
	CustomerProfileModuleRestore                 *BusinessModuleTransactionalRestoreProgress            `json:"customer_profile_module_restore,omitempty"`
	ID                                           string                                                 `json:"id"`
	Kind                                         string                                                 `json:"kind"`
	Actor                                        string                                                 `json:"actor"`
	State                                        string                                                 `json:"state"`
	Phase                                        string                                                 `json:"phase"`
	Error                                        string                                                 `json:"error,omitempty"`
	CreatedAt                                    time.Time                                              `json:"created_at"`
	UpdatedAt                                    time.Time                                              `json:"updated_at"`
	ExpiresAt                                    time.Time                                              `json:"expires_at"`
	Size                                         int64                                                  `json:"size"`
	UploadedBytes                                int64                                                  `json:"uploaded_bytes"`
	CiphertextSHA256                             string                                                 `json:"ciphertext_sha256,omitempty"`
	DownloadRequests                             int                                                    `json:"download_requests"`
	WorkCopyExpired                              bool                                                   `json:"work_copy_expired"`
	Manifest                                     *Manifest                                              `json:"manifest,omitempty"`
	SelectedBackupModules                        []string                                               `json:"selected_backup_modules,omitempty"`
	SelectedModuleRestorePreview                 *SelectedBusinessModuleIsolatedRestorePreview          `json:"selected_module_restore_preview,omitempty"`
	OrderModuleRestore                           *OrderModuleTransactionalHistoryRestoreProgress        `json:"order_module_restore,omitempty"`
	CustomerServiceModuleRestore                 *BusinessModuleTransactionalRestoreProgress            `json:"customer_service_module_restore,omitempty"`
	SelectedCustomerServiceRestoreChoice         *SelectedCustomerServiceModuleSavedRestoreChoice       `json:"selected_customer_service_restore_choice,omitempty"`
	SelectedCustomerServiceCandidateRehearsal    *SelectedBusinessModuleIsolatedRestorePreview          `json:"selected_customer_service_candidate_rehearsal,omitempty"`
	SelectedModuleDisposableStage                *SelectedBusinessModuleDisposableStageIdentity         `json:"selected_module_disposable_stage,omitempty"`
	PreviewTargetFingerprint                     string                                                 `json:"preview_target_fingerprint,omitempty"`
	PreviewTargetTables                          map[string]int64                                       `json:"preview_target_tables,omitempty"`
	StageDatabase                                string                                                 `json:"stage_database,omitempty"`
	StageFingerprint                             string                                                 `json:"stage_fingerprint,omitempty"`
	Journal                                      *RecoveryJournal                                       `json:"recovery_journal,omitempty"`
	Reconciliation                               *ReconciliationEvidence                                `json:"reconciliation,omitempty"`
	ExternalProjectionReconciliationContract     *ExternalProjectionReconciliationContract              `json:"external_projection_reconciliation_contract,omitempty"`
	ReconciliationConnectionsEnabled             bool                                                   `json:"reconciliation_connections_enabled"`
	ProductCatalogRecovery                       *ProductCatalogIsolatedRecoveryResult                  `json:"product_catalog_recovery,omitempty"`
	ProductCatalogPreviewBinding                 *productCatalogRecoveryPreviewBinding                  `json:"product_catalog_preview_binding,omitempty"`
	ProductViewStoppedDrain                      *ProductViewStoppedDrainAndCandidateRebuildProgress    `json:"product_view_stopped_drain,omitempty"`
	ProductCatalogStoppedHandoffPreparation      *ProductCatalogStoppedHandoffPreparationJournal        `json:"product_catalog_stopped_handoff_preparation,omitempty"`
	ProductCatalogStoppedDetailCacheInvalidation *ProductCatalogStoppedDetailCacheInvalidationProgress  `json:"product_catalog_stopped_detail_cache_invalidation,omitempty"`
	ProductCatalogStoppedProjectionEvidence      *ProductCatalogStoppedProjectionOperatorEvidenceRecord `json:"product_catalog_stopped_projection_evidence,omitempty"`
}

func (manager *Manager) taskDirectory(id string) (string, error) {
	if !safeTaskID.MatchString(id) {
		return "", errors.New("invalid task identity")
	}
	return filepath.Join(manager.Configuration.Workspace, id), nil
}

func (manager *Manager) taskPackagePath(id string) (string, error) {
	directory, err := manager.taskDirectory(id)
	return filepath.Join(directory, "application-backup.age"), err
}

func (manager *Manager) GetTask(id string) (*Task, error) {
	manager.stateMutex.Lock()
	defer manager.stateMutex.Unlock()
	directory, err := manager.taskDirectory(id)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(directory, "task.json")) // #nosec G304 -- The path is a fixed record under the configured private maintenance root or validated task directory.
	if err != nil {
		return nil, err
	}
	var task Task
	if err := decodeStrictJSON(data, &task); err != nil {
		return nil, err
	}
	if task.ID != id {
		return nil, errors.New("task identity mismatch")
	}
	return &task, nil
}

func (manager *Manager) ListTasks() ([]*Task, error) {
	entries, err := os.ReadDir(manager.Configuration.Workspace)
	if err != nil {
		return nil, err
	}
	tasks := make([]*Task, 0)
	for _, entry := range entries {
		if !entry.IsDir() || !safeTaskID.MatchString(entry.Name()) {
			continue
		}
		task, err := manager.GetTask(entry.Name())
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].CreatedAt.After(tasks[j].CreatedAt) })
	return tasks, nil
}

func (manager *Manager) saveTask(task *Task) error {
	manager.stateMutex.Lock()
	defer manager.stateMutex.Unlock()
	directory, err := manager.taskDirectory(task.ID)
	if err != nil {
		return err
	}
	task.UpdatedAt = time.Now().UTC()
	return writeDurableJSON(filepath.Join(directory, "task.json"), task)
}

func (manager *Manager) createTask(kind, actor string) (*Task, error) {
	task := &Task{ID: newTaskID(), Kind: kind, Actor: actor, State: "queued", Phase: "created", CreatedAt: time.Now().UTC()}
	task.ExpiresAt = task.CreatedAt.Add(time.Duration(manager.Configuration.WorkCopyLifetimeHours) * time.Hour)
	directory, _ := manager.taskDirectory(task.ID)
	if err := os.Mkdir(directory, 0700); err != nil {
		return nil, err
	}
	return task, manager.saveTask(task)
}
