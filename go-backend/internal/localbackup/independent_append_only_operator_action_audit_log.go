package localbackup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

func (manager *Manager) recordMaintenanceOperatorAction(actor, method, path string) error {
	manager.stateMutex.Lock()
	defer manager.stateMutex.Unlock()
	file, err := os.OpenFile(filepath.Join(manager.Configuration.Workspace, "maintenance-operator-actions.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(file.Close)
	record := struct {
		OccurredAt time.Time `json:"occurred_at"`
		Actor      string    `json:"actor"`
		Method     string    `json:"method"`
		Path       string    `json:"path"`
		Event      string    `json:"event"`
	}{time.Now().UTC(), actor, method, path, "authorized_action_requested"}
	if err := json.NewEncoder(file).Encode(record); err != nil {
		return err
	}
	return file.Sync()
}
