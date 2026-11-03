package localbackup

import "net/http"

func (manager *Manager) registeredMaintenanceMediaModuleHTTPTaskActions() map[string]maintenanceHTTPTaskActionDefinition {
	return map[string]maintenanceHTTPTaskActionDefinition{
		"media-module-formal-restore": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {

			var body MediaModuleFormalRestoreRequest
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			task, err := manager.StartMediaModuleFormalRestore(segments[2], session.Subject, body)
			writeMaintenanceTask(response, task, err)
		}},
		"media-module-interrupted-restore-recovery": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {

			var body struct {
				Confirmation string `json:"confirmation"`
			}
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			task, err := manager.StartInterruptedMediaModuleRestoreRecovery(segments[2], session.Subject, body.Confirmation)
			writeMaintenanceTask(response, task, err)
		}},
	}
}
