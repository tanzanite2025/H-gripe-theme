package localbackup

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"commerce-platform/internal/pkg/maintenancegate"
)

func (manager *Manager) BeginUpload(actor string, size int64) (*Task, error) {
	if size <= 0 || size > manager.Configuration.MaxPackageBytes {
		return nil, errors.New("upload size exceeds configured package capacity")
	}
	if !manager.operationMutex.TryLock() {
		return nil, errors.New("another backup or restore operation is running")
	}
	defer manager.operationMutex.Unlock()
	tasks, err := manager.ListTasks()
	if err != nil {
		return nil, err
	}
	reserved := size
	for _, task := range tasks {
		if taskHasUnresolvedProductCatalogStoppedHandoffPreparation(task) {
			return nil, errors.New("unresolved stopped product handoff preparation work copies must be retained")
		}
		if task.ProductViewStoppedDrain != nil && task.ProductViewStoppedDrain.Phase != "candidate_rebuilt" {
			return nil, errors.New("unresolved product view drain work copies must be retained for explicit retry")
		}
		if task.State == "uploading" {
			reserved += task.Size - task.UploadedBytes
			if reserved > manager.Configuration.MaxWorkspaceBytes {
				return nil, errors.New("active uploads exceed reserved capacity")
			}
		}
	}
	if err := manager.Configuration.checkWorkspaceCapacity(reserved); err != nil {
		return nil, err
	}
	task, err := manager.createTask("upload", actor)
	if err != nil {
		return nil, err
	}
	task.State = "uploading"
	task.Phase = "awaiting_chunks"
	task.Size = size
	return task, manager.saveTask(task)
}

func (manager *Manager) AppendUploadChunk(id string, offset int64, data []byte, expectedSHA256 string) (*Task, error) {
	if len(data) == 0 || int64(len(data)) > ChunkBytes {
		return nil, errors.New("invalid upload chunk length")
	}
	digest := sha256.Sum256(data)
	if hex.EncodeToString(digest[:]) != expectedSHA256 {
		return nil, errors.New("upload chunk checksum mismatch")
	}
	if !manager.operationMutex.TryLock() {
		return nil, errors.New("another backup or restore operation is running")
	}
	defer manager.operationMutex.Unlock()
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	if task.Kind != "upload" || task.State != "uploading" || offset < 0 || offset > task.UploadedBytes || int64(len(data)) > task.Size-offset {
		return nil, errors.New("upload offset or task state is invalid")
	}
	directory, _ := manager.taskDirectory(id)
	file, err := os.OpenFile(filepath.Join(directory, "upload.partial"), os.O_CREATE|os.O_RDWR, 0600) // #nosec G304 -- GetTask validates the task ID; both endpoints are fixed filenames in its private workspace.
	if err != nil {
		return nil, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(file.Close)
	if offset < task.UploadedBytes {
		if int64(len(data)) > task.UploadedBytes-offset {
			return nil, errors.New("chunk overlaps an uncommitted upload range")
		}
		previous := make([]byte, len(data))
		if _, err := file.ReadAt(previous, offset); err != nil || !bytes.Equal(previous, data) {
			return nil, errors.New("retry chunk differs from committed content")
		}
		return task, nil
	}
	if err := manager.Configuration.checkWorkspaceCapacity(int64(len(data))); err != nil {
		return nil, err
	}
	// Truncate bytes written before an interrupted metadata commit. The durable
	// journal offset is authoritative, so retry never appends duplicate bytes.
	if err := file.Truncate(task.UploadedBytes); err != nil {
		return nil, err
	}
	if _, err := file.WriteAt(data, offset); err != nil {
		return nil, err
	}
	if err := file.Sync(); err != nil {
		return nil, err
	}
	if err := syncDirectory(directory); err != nil {
		return nil, err
	}
	task.UploadedBytes += int64(len(data))
	task.Phase = "receiving_chunks"
	return task, manager.saveTask(task)
}

func (manager *Manager) CommitUpload(id string) (*Task, error) {
	if !manager.operationMutex.TryLock() {
		return nil, errors.New("another backup or restore operation is running")
	}
	defer manager.operationMutex.Unlock()
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	if task.Kind != "upload" || task.UploadedBytes != task.Size {
		return nil, errors.New("upload is incomplete")
	}
	if task.State == "uploaded" {
		return task, nil
	}
	if task.State != "uploading" {
		return nil, errors.New("upload is not writable")
	}
	directory, _ := manager.taskDirectory(id)
	destination, _ := manager.taskPackagePath(id)
	if _, err := os.Stat(destination); errors.Is(err, os.ErrNotExist) { // #nosec G703 -- GetTask validates the task ID; both endpoints are fixed filenames in its private workspace.
		if err := os.Rename(filepath.Join(directory, "upload.partial"), destination); err != nil { // #nosec G703 -- GetTask validates the task ID; both endpoints are fixed filenames in its private workspace.
			return nil, err
		}
	}
	info, err := os.Stat(destination) // #nosec G703 -- GetTask validates the task ID; both endpoints are fixed filenames in its private workspace.
	if err != nil || info.Size() != task.Size {
		return nil, errors.New("committed upload length mismatch")
	}
	if err := syncDirectory(directory); err != nil {
		return nil, err
	}
	task.State = "uploaded"
	task.Phase = "encrypted_package_received_not_yet_verified"
	return task, manager.saveTask(task)
}

func (manager *Manager) ExpireWorkCopies(ctx context.Context) error {
	if !manager.operationMutex.TryLock() {
		return nil
	}
	defer manager.operationMutex.Unlock()
	// Recovery fences and paired rollback targets are never removed by TTL.
	if fenced, err := maintenancegate.IsApplicationFenced(manager.Configuration.MaintenanceDirectory); err != nil || fenced {
		return err
	}
	tasks, err := manager.ListTasks()
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if taskHasUnresolvedProductCatalogStoppedHandoffPreparation(task) {
			return errors.New("unresolved stopped product handoff preparation work copies must be retained")
		}
		if task.ProductViewStoppedDrain != nil && task.ProductViewStoppedDrain.Phase != "candidate_rebuilt" {
			return errors.New("unresolved product view drain work copies must be retained for explicit retry")
		}
		if (task.Journal != nil && task.State != "completed" && task.State != "rolled_back") || task.State == "running" || task.State == "queued" || task.State == "expired" || task.WorkCopyExpired || time.Now().Before(task.ExpiresAt) {
			continue
		}
		if err := manager.requireCustomerServicePersistentCandidateDiscarded(task.ID); err != nil {
			return err
		}
		if task.StageDatabase != "" && task.State != "completed" {
			if taskHasProductCatalogContinuouslyStoppedSwitchJournal(task) {
				if err := manager.cleanupVerifiedRolledBackStoppedProductSwitchMediaVolumes(ctx, task); err != nil {
					return err
				}
			}
			if task.Journal != nil && task.Journal.ProductCatalog != nil {
				if err := manager.cleanupProductCatalogVerifiedPairedMediaVolumes(ctx, task); err != nil {
					return err
				}
			}
			var cleanupErr error
			if task.ProductCatalogPreviewBinding != nil {
				cleanupErr = manager.dropProductCatalogStageOnlyWithBoundIdentity(ctx, task)
			} else {
				cleanupErr = manager.Configuration.dropOwnedStageDatabase(ctx, task.StageDatabase)
			}
			if cleanupErr != nil {
				return cleanupErr
			}
		}
		if task.Journal == nil {
			if err := manager.cleanupUnswitchedPreparedMedia(task); err != nil {
				return err
			}
		}
		directory, _ := manager.taskDirectory(task.ID)
		for _, name := range []string{"application-backup.age", "upload.partial", "verified", "snapshot", "product-snapshot", "product-isolated-target"} {
			if err := os.RemoveAll(filepath.Join(directory, name)); err != nil {
				return err
			}
		}
		task.WorkCopyExpired = true
		if task.Journal == nil {
			task.State = "expired"
			task.Phase = "server_work_copy_expired_computer_copy_unaffected"
		}
		if err := manager.saveTask(task); err != nil {
			return err
		}
	}
	return nil
}

func (manager *Manager) recordDownloadRequest(id string) error {
	task, err := manager.GetTask(id)
	if err != nil {
		return err
	}
	if task.State != "ready_to_download" && task.State != "generated_with_gaps" {
		return errors.New("generated download is not available")
	}
	task.DownloadRequests++
	return manager.saveTask(task)
}

func readUploadChunk(reader io.Reader) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(reader, ChunkBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > ChunkBytes {
		return nil, errors.New("chunk body is too large")
	}
	return data, nil
}
