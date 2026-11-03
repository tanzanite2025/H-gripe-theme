package localbackup

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

// Retain bounded strict JSON decoding, error responses and the nonmutating HTTP task summary.
func readMaintenanceJSON(response http.ResponseWriter, request *http.Request, value any) bool {
	request.Body = http.MaxBytesReader(response, request.Body, 32<<10)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		writeMaintenanceError(response, errors.New("invalid request body"))
		return false
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		writeMaintenanceError(response, errors.New("unexpected trailing request body"))
		return false
	}
	return true
}

func writeMaintenanceJSON(response http.ResponseWriter, status int, value any) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(value)
}

func writeMaintenanceError(response http.ResponseWriter, err error) {
	writeMaintenanceJSON(response, http.StatusConflict, map[string]string{"error": err.Error()})
}

func writeMaintenanceTask(response http.ResponseWriter, task *Task, err error) {
	if err != nil {
		writeMaintenanceError(response, err)
		return
	}
	writeMaintenanceJSON(response, http.StatusOK, map[string]any{"task": summarizeTaskForMaintenanceHTTP(task)})
}

func summarizeTaskForMaintenanceHTTP(task *Task) *Task {
	if task == nil {
		return nil
	}
	summary := *task
	if task.Manifest != nil {
		manifest := *task.Manifest
		manifest.Files = nil
		manifest.MediaFileCount = 0
		manifest.MediaFileBytes = 0
		for _, entry := range task.Manifest.Files {
			if strings.HasPrefix(entry.Path, "media/") || strings.HasPrefix(entry.Path, "product-static/") {
				manifest.MediaFileCount++
				manifest.MediaFileBytes += entry.Size
			}
		}
		summary.Manifest = &manifest
	}
	return &summary
}
