package service

import (
	"encoding/json"
	"errors"
	"sort"
	"time"

	sitequalitydomain "commerce-platform/internal/domain/sitequality"

	"gorm.io/gorm"
)

func (s *SiteQualityEngineService) applyJobEvaluation(
	job sitequalitydomain.SiteQualityJob,
	target sitequalitydomain.SiteQualityTarget,
	runs []LighthouseRunnerRunView,
) error {
	now := time.Now().UTC()
	decision, detections := evaluateSiteQualityRuns(target, job, runs)
	if job.FindingID != nil && *job.FindingID != 0 {
		if s.findings == nil {
			return errors.New("SiteQuality finding repository is unavailable")
		}
		finding, err := s.findings.FindByID(*job.FindingID)
		if err != nil {
			return err
		}
		if finding.TargetID == nil || *finding.TargetID != target.ID ||
			finding.Strategy != job.Strategy {
			return errors.New("SiteQuality recheck finding does not match its target or strategy")
		}
		decision, detections = restrictSiteQualityDecisionToFinding(
			decision,
			detections,
			finding.AuditID,
		)
	}
	status := sitequalitydomain.SiteQualityEvaluationStatusCompleted
	if len(runs) < job.SampleCount {
		status = sitequalitydomain.SiteQualityEvaluationStatusInsufficientSamples
	}
	confirmedAuditIDs := make([]string, 0, len(decision.Confirmed))
	for _, item := range decision.Confirmed {
		confirmedAuditIDs = append(confirmedAuditIDs, item.AuditID)
	}
	confirmedJSON, err := json.Marshal(confirmedAuditIDs)
	if err != nil {
		return err
	}
	cleanJSON, err := json.Marshal(decision.Clean)
	if err != nil {
		return err
	}
	decisionJSON, err := json.Marshal(decision)
	if err != nil {
		return err
	}
	latestRunID := uint(0)
	for _, run := range runs {
		if run.ID > latestRunID {
			latestRunID = run.ID
		}
	}
	if latestRunID == 0 {
		return errors.New("SiteQuality quality evaluation has no run evidence")
	}

	return s.jobs.Transaction(func(tx *gorm.DB) error {
		if err := s.jobs.WithTx(tx).AssertLease(tx, job.ID, s.workerID, job.LeaseGeneration); err != nil {
			return err
		}
		evaluation := sitequalitydomain.SiteQualityEvaluation{
			JobID:                 job.ID,
			TargetID:              target.ID,
			Strategy:              job.Strategy,
			Status:                status,
			SampleCount:           job.SampleCount,
			SuccessfulSamples:     len(runs),
			RequiredConfirmations: job.RequiredConfirmations,
			ConfirmedAuditIDs:     string(confirmedJSON),
			CleanAuditIDs:         string(cleanJSON),
			DecisionJSON:          string(decisionJSON),
			DecidedAt:             now,
		}
		if err := s.jobs.WithTx(tx).CreateEvaluationForLease(tx, &evaluation, s.workerID, job.LeaseGeneration); err != nil {
			return err
		}
		if err := s.findings.WithTx(tx).ApplyEvaluation(tx, sitequalitydomain.SiteQualityFindingEvaluationInput{
			TargetID:                 target.ID,
			FindingID:                job.FindingID,
			TargetURL:                target.CanonicalURL,
			Strategy:                 job.Strategy,
			Detections:               detections,
			CleanAuditIDs:            decision.Clean,
			ObservedAuditIDs:         decision.Observed,
			RequiredCleanEvaluations: s.cfg.RequiredCleanEvaluations,
			LatestRunID:              latestRunID,
			DecidedAt:                now,
		}); err != nil {
			return err
		}
		if err := s.jobs.WithTx(tx).MarkSucceeded(job.ID, s.workerID, job.LeaseGeneration, now); err != nil {
			return err
		}
		return s.targets.WithTx(tx).MarkCompleted(target.ID, now)
	})
}

func restrictSiteQualityDecisionToFinding(
	decision siteQualityEvaluationDecision,
	detections []sitequalitydomain.SiteQualityFindingDetection,
	auditID string,
) (siteQualityEvaluationDecision, []sitequalitydomain.SiteQualityFindingDetection) {
	filteredDetections := make([]sitequalitydomain.SiteQualityFindingDetection, 0, 1)
	for _, detection := range detections {
		if detection.AuditID == auditID {
			filteredDetections = append(filteredDetections, detection)
		}
	}

	filtered := siteQualityEvaluationDecision{
		Confirmed: make([]siteQualityDecision, 0, 1),
		Runs:      decision.Runs,
	}
	for _, item := range decision.Confirmed {
		if item.AuditID == auditID {
			filtered.Confirmed = append(filtered.Confirmed, item)
		}
	}
	for _, observed := range decision.Observed {
		if observed == auditID {
			filtered.Observed = append(filtered.Observed, observed)
		}
	}
	for _, clean := range decision.Clean {
		if clean == auditID {
			filtered.Clean = append(filtered.Clean, clean)
		}
	}
	return filtered, filteredDetections
}

func evaluateSiteQualityRuns(
	target sitequalitydomain.SiteQualityTarget,
	job sitequalitydomain.SiteQualityJob,
	runs []LighthouseRunnerRunView,
) (siteQualityEvaluationDecision, []sitequalitydomain.SiteQualityFindingDetection) {
	byAudit := map[string][]LighthouseRunnerIssue{}
	runIDs := make([]uint, 0, len(runs))
	for _, run := range runs {
		runIDs = append(runIDs, run.ID)
		for _, issue := range run.Issues {
			if _, ok := siteQualityLookupAuditRule(issue.ID); !ok {
				continue
			}
			byAudit[issue.ID] = append(byAudit[issue.ID], issue)
		}
	}

	observed := make([]string, 0, len(byAudit))
	for auditID := range byAudit {
		observed = append(observed, auditID)
	}
	sort.Strings(observed)

	confirmed := make([]siteQualityDecision, 0)
	detections := make([]sitequalitydomain.SiteQualityFindingDetection, 0)
	for _, auditID := range observed {
		issues := byAudit[auditID]
		if len(issues) < job.RequiredConfirmations {
			continue
		}
		decision := siteQualityDecisionFromIssues(auditID, issues, job.SampleCount)
		confirmed = append(confirmed, decision)
		targetID := target.ID
		detections = append(detections, sitequalitydomain.SiteQualityFindingDetection{
			TargetID:           &targetID,
			TargetURL:          target.CanonicalURL,
			Strategy:           job.Strategy,
			AuditID:            auditID,
			RuleID:             decision.RuleID,
			ProviderAuditID:    decision.ProviderAuditID,
			FindingKind:        decision.Kind,
			RuleVersion:        decision.RuleVersion,
			Confidence:         decision.Confidence,
			SampleCount:        decision.SampleCount,
			Confirmations:      decision.Confirmations,
			Title:              decision.Title,
			Description:        decision.Description,
			Severity:           decision.Severity,
			LatestRunID:        lastSiteQualityRunID(runIDs),
			LatestSavingsMS:    decision.MedianMS,
			LatestSavingsBytes: decision.MedianBytes,
			ResourceCount:      siteQualityDecisionEvidenceCount(decision),
			LatestEvidence:     mustEncodeSiteQualityFindingEvidence(decision),
		})
	}

	observedSet := make(map[string]struct{}, len(observed))
	for _, auditID := range observed {
		observedSet[auditID] = struct{}{}
	}
	clean := make([]string, 0)
	for auditID := range siteQualityActionableAuditRules {
		if _, ok := observedSet[auditID]; !ok {
			clean = append(clean, auditID)
		}
	}
	sort.Strings(clean)
	return siteQualityEvaluationDecision{
		Confirmed: confirmed,
		Clean:     clean,
		Observed:  observed,
		Runs:      runIDs,
	}, detections
}

func siteQualityDecisionFromIssues(
	auditID string,
	issues []LighthouseRunnerIssue,
	sampleCount int,
) siteQualityDecision {
	rule, _ := siteQualityLookupAuditRule(auditID)
	scores := make([]float64, 0)
	savingsMS := make([]float64, 0)
	savingsBytes := make([]int64, 0)
	best := issues[0]
	resourceByURL := map[string]LighthouseRunnerResource{}
	for _, issue := range issues {
		if issue.Score != nil {
			scores = append(scores, *issue.Score)
		}
		if issue.SavingsMS != nil {
			savingsMS = append(savingsMS, *issue.SavingsMS)
		}
		if issue.SavingsBytes != nil {
			savingsBytes = append(savingsBytes, *issue.SavingsBytes)
		}
		if siteQualityIssueRank(issue) > siteQualityIssueRank(best) {
			best = issue
		}
		for _, resource := range issue.Resources {
			resourceByURL[resource.URL] = resource
		}
	}
	resources := make([]LighthouseRunnerResource, 0, len(resourceByURL))
	for _, resource := range resourceByURL {
		resources = append(resources, resource)
	}
	sort.Slice(resources, func(i, j int) bool {
		return siteQualityResourceWastedMS(resources[i]) > siteQualityResourceWastedMS(resources[j])
	})
	return siteQualityDecision{
		AuditID:         auditID,
		RuleID:          siteQualityRuleIDForAudit(auditID),
		ProviderAuditID: siteQualityProviderAuditIDForAudit(auditID),
		Kind:            rule.Kind,
		RuleVersion:     siteQualityAuditRuleVersion,
		Title:           best.Title,
		Description:     best.Description,
		Severity:        siteQualityConfirmedSeverity(best, savingsMS, savingsBytes),
		Confirmations:   len(issues),
		SampleCount:     sampleCount,
		Confidence:      float64(len(issues)) / float64(sampleCount),
		MedianScore:     medianFloat64(scores),
		MedianMS:        medianFloat64(savingsMS),
		MedianBytes:     medianInt64(savingsBytes),
		Resources:       resources,
		Links:           best.Links,
		Headings:        best.Headings,
		StructuredData:  best.StructuredData,
		DisplayValue:    best.DisplayValue,
		NumericValue:    copyFloat64(best.NumericValue),
	}
}

func siteQualityDecisionEvidenceCount(decision siteQualityDecision) int {
	if len(decision.Headings) > 0 {
		return len(decision.Headings)
	}
	if len(decision.StructuredData) > 0 {
		return len(decision.StructuredData)
	}
	if len(decision.Links) > 0 {
		return len(decision.Links)
	}
	return len(decision.Resources)
}

func siteQualityConfirmedSeverity(
	best LighthouseRunnerIssue,
	savingsMS []float64,
	savingsBytes []int64,
) string {
	medianMS := medianFloat64(savingsMS)
	if medianMS != nil {
		switch {
		case *medianMS >= 1000:
			return "high"
		case *medianMS >= 250:
			return "medium"
		default:
			return "low"
		}
	}
	medianBytesValue := medianInt64(savingsBytes)
	if medianBytesValue != nil {
		switch {
		case *medianBytesValue >= 512*1024:
			return "high"
		case *medianBytesValue >= 64*1024:
			return "medium"
		default:
			return "low"
		}
	}
	return best.Severity
}

func medianFloat64(values []float64) *float64 {
	if len(values) == 0 {
		return nil
	}
	sort.Float64s(values)
	index := len(values) / 2
	var result float64
	if len(values)%2 == 0 {
		result = (values[index-1] + values[index]) / 2
	} else {
		result = values[index]
	}
	return &result
}

func medianInt64(values []int64) *int64 {
	if len(values) == 0 {
		return nil
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	index := len(values) / 2
	var result int64
	if len(values)%2 == 0 {
		result = (values[index-1] + values[index]) / 2
	} else {
		result = values[index]
	}
	return &result
}

func lastSiteQualityRunID(runIDs []uint) uint {
	var latest uint
	for _, id := range runIDs {
		if id > latest {
			latest = id
		}
	}
	return latest
}

func mustEncodeSiteQualityFindingEvidence(decision siteQualityDecision) string {
	evidence := sitequalitydomain.SiteQualityFindingEvidence{
		AuditID:         decision.AuditID,
		RuleID:          decision.RuleID,
		ProviderAuditID: decision.ProviderAuditID,
		Title:           decision.Title,
		Description:     decision.Description,
		Score:           decision.MedianScore,
		DisplayValue:    decision.DisplayValue,
		NumericValue:    decision.NumericValue,
		SavingsMS:       decision.MedianMS,
		SavingsBytes:    decision.MedianBytes,
		Resources:       siteQualityFindingResourcesFromLighthouse(decision.Resources),
		Links:           decision.Links,
		Headings:        decision.Headings,
		StructuredData:  decision.StructuredData,
	}
	encoded, err := json.Marshal(evidence)
	if err != nil {
		return "{}"
	}
	return string(encoded)
}

func siteQualityFindingResourcesFromLighthouse(
	resources []LighthouseRunnerResource,
) []sitequalitydomain.SiteQualityFindingResource {
	result := make([]sitequalitydomain.SiteQualityFindingResource, 0, len(resources))
	for _, resource := range resources {
		result = append(result, sitequalitydomain.SiteQualityFindingResource{
			URL:        resource.URL,
			TotalBytes: copyInt64(resource.TotalBytes),
			WastedMS:   copyFloat64(resource.WastedMS),
		})
	}
	return result
}
