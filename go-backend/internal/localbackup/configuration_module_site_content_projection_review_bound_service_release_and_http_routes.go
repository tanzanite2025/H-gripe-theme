package localbackup

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"
)

// This is an operator attestation, not an automated CDN/cache purge claim.
// It is persisted against the already committed SQL/media result before release.
type ConfigurationContentProjectionReview struct {
	Operator                 string    `json:"operator"`
	Reference                string    `json:"reference"`
	AfterDatabaseFingerprint string    `json:"after_database_fingerprint"`
	AfterMediaFingerprint    string    `json:"after_media_fingerprint"`
	ReviewedAt               time.Time `json:"reviewed_at"`
}

type ConfigurationContentReviewedServiceReleaseRequest struct {
	Confirmation                   string `json:"confirmation"`
	ExpectedJournalSHA256          string `json:"expected_journal_sha256"`
	ContentAndCacheReviewReference string `json:"content_and_cache_review_reference"`
}

func validateConfigurationContentProjectionReviewBinding(task *Task) error {
	p := transactionalBusinessModuleRestoreProgressForTask(task)
	if p == nil {
		return errors.New("configuration content review requires one valid module journal")
	}
	review := p.ConfigurationContentProjectionReview
	if task.ConfigurationModuleSiteContentRestore == nil {
		if review != nil {
			return errors.New("configuration content review cannot authorize another module")
		}
		return nil
	}
	if len(p.Files) > 0 {
		return errors.New("configuration content entry cannot install or replace files")
	}
	if review == nil {
		if p.Outcome == "restored" && (p.Phase == "releasing" || p.Phase == "released") {
			return errors.New("configuration service release requires the committed content and cache review")
		}
		return nil
	}
	if (p.Phase != "rows_committed" && p.Phase != "releasing" && p.Phase != "released") || strings.TrimSpace(review.Operator) == "" || len(strings.TrimSpace(review.Reference)) < 4 || len(review.Reference) > 2000 || review.ReviewedAt.IsZero() || review.AfterDatabaseFingerprint != p.AfterDatabaseFingerprint || review.AfterMediaFingerprint != p.AfterMediaFingerprint || p.AuditID <= 0 {
		return errors.New("configuration content review is not bound to the committed result")
	}
	return nil
}

func (manager *Manager) StartConfigurationSiteContentReviewedServiceRelease(id, actor string, request ConfigurationContentReviewedServiceReleaseRequest) (*Task, error) {
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	if task.ConfigurationModuleSiteContentRestore == nil || validateTransactionalBusinessModuleRestoreProgress(task) != nil || task.ConfigurationModuleSiteContentRestore.Phase != "rows_committed" || (task.State != "configuration_site_content_restored_still_isolated" && task.State != "failed" && task.State != "interrupted") || strings.TrimSpace(actor) == "" || request.ExpectedJournalSHA256 != task.ConfigurationModuleSiteContentRestore.SHA256 || request.Confirmation != "RELEASE CONFIGURATION CONTENT "+manager.Configuration.SiteID+"/"+manager.Configuration.Environment+"/"+id || len(strings.TrimSpace(request.ContentAndCacheReviewReference)) < 4 || len(request.ContentAndCacheReviewReference) > 2000 {
		return nil, errors.New("请核对本次恢复的网站内容、文章缓存及页面/CDN缓存，填写核对记录后确认恢复服务")
	}
	err = manager.startTaskOperation(task, "reviewing_configuration_content_and_releasing_service", func(ctx context.Context, task *Task) error {
		if err := manager.verifyTransactionalBusinessModuleRestoreFenceAndRuntime(task); err != nil {
			return err
		}
		lease, err := manager.acquireMaintenance(ctx, task)
		if err != nil {
			return err
		}
		defer func(closeResource func() error) { _ = closeResource() }(lease.Close)
		live, err := manager.Configuration.connectDatabase(ctx, manager.Configuration.DatabaseName)
		if err != nil {
			return err
		}
		defer func(closeResource func() error) { _ = closeResource() }(live.Close)
		p := task.ConfigurationModuleSiteContentRestore
		p.ConfigurationContentProjectionReview = &ConfigurationContentProjectionReview{Operator: actor, Reference: strings.TrimSpace(request.ContentAndCacheReviewReference), AfterDatabaseFingerprint: p.AfterDatabaseFingerprint, AfterMediaFingerprint: p.AfterMediaFingerprint, ReviewedAt: time.Now().UTC()}
		if err := manager.saveTransactionalBusinessModuleRestoreProgress(task); err != nil {
			return err
		}
		return manager.completeOrCancelInterruptedTransactionalBusinessModuleRestoreWithHeldLease(ctx, task, live, nil)
	})
	if err != nil {
		return nil, err
	}
	return manager.GetTask(id)
}

func (manager *Manager) registeredMaintenanceConfigurationSiteContentHTTPTaskActions() map[string]maintenanceHTTPTaskActionDefinition {
	return map[string]maintenanceHTTPTaskActionDefinition{
		"configuration-site-content-formal-restore": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {
			var body ConfigurationModuleSiteContentFormalRestoreRequest
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			task, err := manager.StartConfigurationModuleSiteContentFormalRestore(segments[2], session.Subject, body)
			writeMaintenanceTask(response, task, err)
		}},
		"configuration-site-content-interrupted-restore-recovery": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {
			var body struct {
				Confirmation string `json:"confirmation"`
			}
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			task, err := manager.StartInterruptedConfigurationModuleSiteContentRestoreRecovery(segments[2], session.Subject, body.Confirmation)
			writeMaintenanceTask(response, task, err)
		}},
		"configuration-site-content-reviewed-service-release": {Permission: PermissionRestore, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {
			var body ConfigurationContentReviewedServiceReleaseRequest
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			task, err := manager.StartConfigurationSiteContentReviewedServiceRelease(segments[2], session.Subject, body)
			writeMaintenanceTask(response, task, err)
		}},
	}
}
