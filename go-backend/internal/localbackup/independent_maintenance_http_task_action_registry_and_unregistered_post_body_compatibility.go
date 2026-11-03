package localbackup

import (
	"net/http"
)

// Keep action/permission/handler registration together and preserve strict-body errors for unknown task POST paths.
type maintenanceHTTPTaskActionDefinition struct {
	Permission string
	Methods    []string
	Handle     func(http.ResponseWriter, *http.Request, *MaintenanceSession, []string)
}

func (manager *Manager) registeredMaintenanceHTTPTaskActions() map[string]maintenanceHTTPTaskActionDefinition {
	actions := map[string]maintenanceHTTPTaskActionDefinition{}
	for _, group := range []map[string]maintenanceHTTPTaskActionDefinition{
		manager.registeredMaintenanceProductCatalogHTTPTaskActions(),
		manager.registeredMaintenanceTransactionalBusinessModulesHTTPTaskActions(),
		manager.registeredMaintenanceMediaModuleHTTPTaskActions(),
		manager.registeredMaintenanceConfigurationSiteContentHTTPTaskActions(),
		manager.registeredMaintenanceCustomerServiceReviewAndSelectionHTTPTaskActions(),
		manager.registeredMaintenanceWholeStorefrontRecoveryHTTPTaskActions(),
		manager.registeredMaintenancePackageValidationAndModuleRehearsalHTTPTaskActions(),
		manager.registeredMaintenancePackageDownloadsHTTPTaskActions(),
	} {
		for action, definition := range group {
			if _, exists := actions[action]; exists {
				panic("duplicate independent maintenance HTTP task action: " + action)
			}
			actions[action] = definition
		}
	}
	return actions
}

func serveUnregisteredMaintenanceTaskPostBodyCompatibility(response http.ResponseWriter, request *http.Request, segments []string) bool {
	if len(segments) != 4 || segments[0] != "api" || segments[1] != "tasks" || request.Method != http.MethodPost {
		return false
	}
	var body struct {
		Confirmation string                 `json:"confirmation"`
		Evidence     ReconciliationEvidence `json:"evidence"`
	}
	return !readMaintenanceJSON(response, request, &body)
}
