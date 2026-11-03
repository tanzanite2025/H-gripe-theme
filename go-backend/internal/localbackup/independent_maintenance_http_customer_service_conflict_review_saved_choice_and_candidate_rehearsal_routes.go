package localbackup

import (
	"net/http"
)

// Register each supported task action with its required permission and original request/response behavior.
func (manager *Manager) registeredMaintenanceCustomerServiceReviewAndSelectionHTTPTaskActions() map[string]maintenanceHTTPTaskActionDefinition {
	return map[string]maintenanceHTTPTaskActionDefinition{
		"customer-service-effective-selection-inspection": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {

			var body CustomerServiceEffectiveSelectionInspectionRequest
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			report, err := manager.InspectCustomerServiceRecoveryEffectiveSelection(request.Context(), segments[2], body)
			if err != nil {
				writeMaintenanceError(response, err)
				return
			}
			writeMaintenanceJSON(response, http.StatusOK, map[string]any{"report": report})
		}},
		"customer-service-conflict-review": {Permission: PermissionRestore, Methods: []string{http.MethodGet, http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {

			var review *CustomerServiceRecoveryConflictReview
			var err error
			if request.Method == http.MethodGet {
				review, err = manager.ReadCustomerServiceRecoveryConflictReview(segments[2])
			} else {
				var body CustomerServiceRecoveryConflictReviewRequest
				if !readMaintenanceJSON(response, request, &body) {
					return
				}
				review, err = manager.SaveCustomerServiceRecoveryConflictReview(request.Context(), segments[2], session.Subject, body)
			}
			if err != nil {
				writeMaintenanceError(response, err)
				return
			}
			writeMaintenanceJSON(response, http.StatusOK, map[string]any{"review": review})
		}},
		"selected-customer-service-restore-choice": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {
			id := segments[2]

			var body struct {
				PreviewSHA256 string `json:"preview_sha256"`
				Policy        string `json:"policy"`
			}
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			task, err := manager.SaveSelectedCustomerServiceModuleRestoreChoice(request.Context(), id, session.Subject, body.PreviewSHA256, body.Policy)
			writeMaintenanceTask(response, task, err)
		}},
		"selected-customer-service-candidate-rehearsal": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {
			id := segments[2]

			var body struct {
				ChoiceSHA256 string `json:"choice_sha256"`
				Confirmation string `json:"confirmation"`
			}
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			task, err := manager.StartSelectedCustomerServiceModuleCandidateRehearsal(id, body.ChoiceSHA256, body.Confirmation)
			writeMaintenanceTask(response, task, err)
		}},
	}
}
