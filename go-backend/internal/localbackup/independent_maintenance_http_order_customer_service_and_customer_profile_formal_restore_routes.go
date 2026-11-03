package localbackup

import (
	"net/http"
)

// Register each supported task action with its required permission and original request/response behavior.
func (manager *Manager) registeredMaintenanceTransactionalBusinessModulesHTTPTaskActions() map[string]maintenanceHTTPTaskActionDefinition {
	return map[string]maintenanceHTTPTaskActionDefinition{
		"customer-profile-module-formal-restore": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {

			var body CustomerProfileModuleFormalRestoreRequest
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			task, err := manager.StartCustomerProfileModuleFormalRestore(segments[2], session.Subject, body)
			writeMaintenanceTask(response, task, err)
		}},
		"customer-profile-module-interrupted-restore-recovery": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {

			var body struct {
				Confirmation string `json:"confirmation"`
			}
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			task, err := manager.StartInterruptedCustomerProfileModuleRestoreRecovery(segments[2], session.Subject, body.Confirmation)
			writeMaintenanceTask(response, task, err)
		}},
		"customer-service-module-formal-restore": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {

			var body CustomerServiceModuleFormalRestoreRequest
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			task, err := manager.StartCustomerServiceModuleFormalRestore(segments[2], session.Subject, body)
			writeMaintenanceTask(response, task, err)
		}},
		"customer-service-module-interrupted-restore-recovery": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {

			var body struct {
				Confirmation string `json:"confirmation"`
			}
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			task, err := manager.StartInterruptedCustomerServiceModuleRestoreRecovery(segments[2], session.Subject, body.Confirmation)
			writeMaintenanceTask(response, task, err)
		}},
		"order-module-formal-history-restore": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {

			var body OrderModuleFormalHistoryRestoreRequest
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			task, err := manager.StartOrderModuleFormalHistoryRestore(segments[2], session.Subject, body)
			writeMaintenanceTask(response, task, err)
		}},
		"order-module-interrupted-restore-recovery": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {

			var body struct {
				Confirmation string `json:"confirmation"`
			}
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			task, err := manager.StartInterruptedOrderHistoryRestoreRecovery(segments[2], session.Subject, body.Confirmation)
			writeMaintenanceTask(response, task, err)
		}},
	}
}
