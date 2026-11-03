package localbackup

import (
	"net/http"
)

// Register each supported task action with its required permission and original request/response behavior.
func (manager *Manager) registeredMaintenancePackageValidationAndModuleRehearsalHTTPTaskActions() map[string]maintenanceHTTPTaskActionDefinition {
	return map[string]maintenanceHTTPTaskActionDefinition{
		"module-dependency-report": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {
			id := segments[2]

			var body struct {
				Module string `json:"module"`
			}
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			report, err := manager.InspectVerifiedPackageModuleDependencies(request.Context(), id, body.Module)
			if err != nil {
				writeMaintenanceError(response, err)
				return
			}
			writeMaintenanceJSON(response, http.StatusOK, map[string]any{"report": report})
		}},
		"verify": {Permission: PermissionUpload, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {
			id := segments[2]

			task, err := manager.VerifyUploadedPackage(id)
			writeMaintenanceTask(response, task, err)
		}},
		"selected-module-isolated-restore-preview": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {
			id := segments[2]

			var body struct {
				Confirmation string `json:"confirmation"`
			}
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			task, err := manager.StartSelectedBusinessModuleIsolatedRestorePreview(id, body.Confirmation)
			writeMaintenanceTask(response, task, err)
		}},
	}
}
