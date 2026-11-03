package localbackup

import (
	"net/http"
)

// Register each supported task action with its required permission and original request/response behavior.
func (manager *Manager) registeredMaintenanceProductCatalogHTTPTaskActions() map[string]maintenanceHTTPTaskActionDefinition {
	return map[string]maintenanceHTTPTaskActionDefinition{
		"product-catalog-stopped-read-only-public-image-http-query": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {

			var body StoppedProductReadOnlyPublicImageHTTPRequest
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			result, err := manager.ExecuteStoppedProductReadOnlyPublicImageHTTPQuery(request.Context(), segments[2], body)
			if err != nil {
				writeMaintenanceError(response, err)
				return
			}
			writeMaintenanceJSON(response, http.StatusOK, map[string]any{"result": result})
		}},
		"product-catalog-stopped-read-only-storefront-http-query": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {

			var body StoppedProductReadOnlyStorefrontHTTPRequest
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			result, err := manager.ExecuteStoppedProductReadOnlyStorefrontHTTPQuery(request.Context(), segments[2], body)
			if err != nil {
				writeMaintenanceError(response, err)
				return
			}
			writeMaintenanceJSON(response, http.StatusOK, map[string]any{"result": result})
		}},
		"product-catalog-stopped-read-only-media-http-query": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {

			var body StoppedProductReadOnlyMediaHTTPRequest
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			result, err := manager.ExecuteStoppedProductReadOnlyMediaHTTPQuery(request.Context(), segments[2], body)
			if err != nil {
				writeMaintenanceError(response, err)
				return
			}
			writeMaintenanceJSON(response, http.StatusOK, map[string]any{"result": result})
		}},
		"product-catalog-stopped-read-only-product-http-query": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {

			var body StoppedProductReadOnlyPublicHTTPRequest
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			result, err := manager.ExecuteStoppedProductReadOnlyPublicHTTPQuery(request.Context(), segments[2], body)
			if err != nil {
				writeMaintenanceError(response, err)
				return
			}
			writeMaintenanceJSON(response, http.StatusOK, map[string]any{"result": result})
		}},
		"product-catalog-stopped-projection-inspection": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {

			var body struct{}
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			report, err := manager.InspectStoppedProductStorefrontProjectionBaseline(request.Context(), segments[2])
			if err != nil {
				writeMaintenanceError(response, err)
				return
			}
			writeMaintenanceJSON(response, http.StatusOK, map[string]any{"report": report})
		}},
		"product-catalog-stopped-projection-evidence": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {

			var body ProductCatalogStoppedProjectionEvidenceRequest
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			record, err := manager.SaveStoppedProductProjectionIndependentOperatorEvidence(request.Context(), segments[2], session.Subject, body)
			if err != nil {
				writeMaintenanceError(response, err)
				return
			}
			writeMaintenanceJSON(response, http.StatusOK, map[string]any{"record": record})
		}},
		"product-catalog-formal-restore-acceptance": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {

			var body ProductCatalogFormalRestoreAcceptanceRequest
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			task, err := manager.StartProductCatalogFormalRestoreAcceptance(segments[2], session.Subject, body)
			writeMaintenanceTask(response, task, err)
		}},
		"product-catalog-stopped-isolated-switch": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {

			var body ProductCatalogStoppedSwitchConfirmationRequest
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			var task *Task
			var err error
			if segments[3] == "product-catalog-stopped-isolated-switch" {
				task, err = manager.StartProductCatalogIsolatedSwitchWithinContinuouslyStoppedWindow(segments[2], session.Subject, body)
			} else {
				task, err = manager.StartVerifiedPairedRollbackOfStoppedProductCatalogSwitch(segments[2], session.Subject, body)
			}
			writeMaintenanceTask(response, task, err)
		}},
		"product-catalog-stopped-switch-rollback": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {

			var body ProductCatalogStoppedSwitchConfirmationRequest
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			var task *Task
			var err error
			if segments[3] == "product-catalog-stopped-isolated-switch" {
				task, err = manager.StartProductCatalogIsolatedSwitchWithinContinuouslyStoppedWindow(segments[2], session.Subject, body)
			} else {
				task, err = manager.StartVerifiedPairedRollbackOfStoppedProductCatalogSwitch(segments[2], session.Subject, body)
			}
			writeMaintenanceTask(response, task, err)
		}},
		"product-catalog-stopped-detail-cache-invalidation": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {

			var body ProductCatalogStoppedSwitchConfirmationRequest
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			task, err := manager.StartStoppedProductDetailCacheInvalidation(segments[2], session.Subject, body)
			writeMaintenanceTask(response, task, err)
		}},
		"product-catalog-stopped-post-switch-inspection": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {

			var body struct{}
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			report, err := manager.InspectStoppedProductCatalogPostSwitch(request.Context(), segments[2])
			if err != nil {
				writeMaintenanceError(response, err)
				return
			}
			writeMaintenanceJSON(response, http.StatusOK, map[string]any{"report": report})
		}},
		"product-catalog-stopped-handoff-preparation": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {

			var body ProductCatalogStoppedHandoffPreparationRequest
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			task, err := manager.StartProductCatalogContinuouslyStoppedHandoffPreparation(segments[2], session.Subject, body)
			writeMaintenanceTask(response, task, err)
		}},
		"product-catalog-stopped-handoff-cancellation": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {

			var body ProductCatalogStoppedHandoffCancellationRequest
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			task, err := manager.StartVerifiedCancellationOfProductCatalogStoppedHandoffPreparation(segments[2], session.Subject, body)
			writeMaintenanceTask(response, task, err)
		}},
		"product-view-stopped-drain-and-rebuild": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {

			var body ProductViewStoppedDrainAndCandidateRebuildRequest
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			task, err := manager.StartStoppedProductViewDrainAndFreshCandidateRebuild(segments[2], session.Subject, body)
			writeMaintenanceTask(response, task, err)
		}},
		"product-catalog-handoff-inspection": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {

			var body struct{}
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			report, err := manager.InspectProductCatalogCandidateHandoff(request.Context(), segments[2])
			if err != nil {
				writeMaintenanceError(response, err)
				return
			}
			writeMaintenanceJSON(response, http.StatusOK, map[string]any{"report": report})
		}},
	}
}
