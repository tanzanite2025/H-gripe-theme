package localbackup

import (
	"net/http"
)

// Register each supported task action with its required permission and original request/response behavior.
func (manager *Manager) registeredMaintenanceWholeStorefrontRecoveryHTTPTaskActions() map[string]maintenanceHTTPTaskActionDefinition {
	return map[string]maintenanceHTTPTaskActionDefinition{
		"reconciliation-report": {Permission: PermissionRestore, Methods: []string{http.MethodGet}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {

			report, err := manager.InspectReconciliation(segments[2])
			if err != nil {
				writeMaintenanceError(response, err)
				return
			}
			writeMaintenanceJSON(response, http.StatusOK, map[string]any{"report": report})
		}},
		"reconciliation-connections": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {
			id := segments[2]
			var body struct {
				Confirmation string                 `json:"confirmation"`
				Evidence     ReconciliationEvidence `json:"evidence"`
			}
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			err := manager.OpenReconciliationConnections(id, body.Confirmation)
			writeMaintenanceTask(response, nil, err)
		}},
		"apply": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {
			id := segments[2]
			var body struct {
				Confirmation string                 `json:"confirmation"`
				Evidence     ReconciliationEvidence `json:"evidence"`
			}
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			task, err := manager.ApplyVerifiedRecovery(id, body.Confirmation)
			writeMaintenanceTask(response, task, err)
		}},
		"rollback": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {
			id := segments[2]
			var body struct {
				Confirmation string                 `json:"confirmation"`
				Evidence     ReconciliationEvidence `json:"evidence"`
			}
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			task, err := manager.RollbackRecovery(id, body.Confirmation)
			writeMaintenanceTask(response, task, err)
		}},
		"complete": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {
			id := segments[2]
			var body struct {
				Confirmation string                 `json:"confirmation"`
				Evidence     ReconciliationEvidence `json:"evidence"`
			}
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			body.Evidence.Operator = session.Subject
			err := manager.CompleteReconciledRecovery(id, body.Confirmation, body.Evidence)
			writeMaintenanceTask(response, nil, err)
		}},
		"release": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {
			id := segments[2]
			var body struct {
				Confirmation string                 `json:"confirmation"`
				Evidence     ReconciliationEvidence `json:"evidence"`
			}
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			err := manager.ReleaseUnmodifiedMaintenance(id)
			writeMaintenanceTask(response, nil, err)
		}},
	}
}
