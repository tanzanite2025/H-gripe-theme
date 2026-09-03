package service

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	sitequalitydomain "commerce-platform/internal/domain/sitequality"
)

func siteQualityRenderedLinkAuditIssues(audit *siteQualityRenderedLinkAudit) []LighthouseRunnerIssue {
	if audit == nil || (!audit.Configured && strings.TrimSpace(audit.Status) == "") {
		return nil
	}
	status := strings.TrimSpace(audit.Status)
	if !audit.Configured || siteQualityRuntimeAuditSkipped(status) {
		return nil
	}
	if status != "complete" && !siteQualityRuntimeAuditTimeoutSkipped(status) {
		return []LighthouseRunnerIssue{
			siteQualityRuntimeScanFailureIssue(
				siteQualityLinkAuditFailedAuditID,
				"links",
				"Rendered link audit did not complete",
				audit.Error,
			),
		}
	}

	offenders := make([]siteQualityRenderedLink, 0)
	for _, link := range audit.Links {
		if link.TimeoutSkipped {
			continue
		}
		if siteQualityRenderedLinkBroken(link) {
			offenders = append(offenders, link)
		}
	}
	if len(offenders) == 0 {
		return nil
	}
	sort.SliceStable(offenders, func(i, j int) bool {
		if offenders[i].StatusCode != offenders[j].StatusCode {
			return offenders[i].StatusCode > offenders[j].StatusCode
		}
		return offenders[i].Href < offenders[j].Href
	})
	totalBrokenCount := len(offenders)
	severity := siteQualityBrokenLinkSeverity(offenders)
	evidenceLinks := offenders
	if len(evidenceLinks) > 12 {
		evidenceLinks = evidenceLinks[:12]
	}
	count := float64(totalBrokenCount)
	return []LighthouseRunnerIssue{
		{
			ID:           siteQualityBrokenLinkAuditID,
			Kind:         "links",
			RuleVersion:  siteQualityAuditRuleVersion,
			Title:        "Rendered page contains broken links",
			Description:  "The final browser-rendered DOM includes links that return HTTP errors, fail to respond, or redirect outside the configured storefront origin.",
			Severity:     severity,
			DisplayValue: fmt.Sprintf("%d broken link(s)", totalBrokenCount),
			NumericValue: &count,
			Links:        siteQualityBrokenLinkEvidence(evidenceLinks),
			Runtime:      siteQualityBrokenLinkRuntimeEvidence(evidenceLinks),
		},
	}
}

func siteQualityRenderedLinkBroken(link siteQualityRenderedLink) bool {
	if !link.OK {
		return true
	}
	if link.StatusCode == 0 || link.StatusCode >= http.StatusBadRequest {
		return true
	}
	return strings.TrimSpace(link.Error) != ""
}

func siteQualityBrokenLinkSeverity(links []siteQualityRenderedLink) string {
	for _, link := range links {
		if link.StatusCode == http.StatusNotFound || link.StatusCode == http.StatusGone {
			return "critical"
		}
	}
	return "high"
}

func siteQualityBrokenLinkEvidence(links []siteQualityRenderedLink) []sitequalitydomain.SiteQualityLinkEvidence {
	evidence := make([]sitequalitydomain.SiteQualityLinkEvidence, 0, len(links))
	for _, link := range links {
		evidence = append(evidence, sitequalitydomain.SiteQualityLinkEvidence{
			Href:     strings.TrimSpace(link.Href),
			Text:     strings.TrimSpace(link.Text),
			TextLang: strings.TrimSpace(link.TextLang),
		})
	}
	return evidence
}

func siteQualityBrokenLinkRuntimeEvidence(links []siteQualityRenderedLink) []sitequalitydomain.SiteQualityRuntimeEvidence {
	evidence := make([]sitequalitydomain.SiteQualityRuntimeEvidence, 0, len(links))
	for _, link := range links {
		evidence = append(evidence, sitequalitydomain.SiteQualityRuntimeEvidence{
			Selector:      strings.TrimSpace(link.Selector),
			Text:          strings.TrimSpace(link.Text),
			URL:           strings.TrimSpace(link.Href),
			FinalURL:      strings.TrimSpace(link.FinalURL),
			Status:        "failed",
			Error:         strings.TrimSpace(link.Error),
			StatusCode:    link.StatusCode,
			RedirectCount: link.RedirectCount,
		})
	}
	return evidence
}

func siteQualityInteractionAuditIssues(audit *siteQualityInteractionAudit) []LighthouseRunnerIssue {
	if audit == nil {
		return nil
	}
	status := strings.TrimSpace(audit.Status)
	if !audit.Configured || siteQualityRuntimeAuditSkipped(status) {
		return nil
	}
	if status != "complete" && !siteQualityRuntimeAuditTimeoutSkipped(status) {
		return []LighthouseRunnerIssue{
			siteQualityRuntimeScanFailureIssue(
				siteQualityInteractionAuditFailedID,
				"interaction",
				"Interaction latency audit did not complete",
				audit.Error,
			),
		}
	}

	failures := make([]siteQualityInteractionProbe, 0)
	slowInteractions := make([]siteQualityInteractionProbe, 0)
	for _, interaction := range audit.Interactions {
		interactionStatus := strings.TrimSpace(interaction.Status)
		if siteQualityRuntimeAuditTimeoutSkipped(interactionStatus) {
			continue
		}
		if interactionStatus != "complete" {
			failures = append(failures, interaction)
			continue
		}
		if interaction.Exceeded {
			slowInteractions = append(slowInteractions, interaction)
		}
	}
	if len(failures) == 0 && len(slowInteractions) == 0 {
		return nil
	}
	issues := make([]LighthouseRunnerIssue, 0, 2)
	if len(failures) > 0 {
		issues = append(issues, siteQualityInteractionProbeFailureIssue(failures))
	}
	if len(slowInteractions) > 0 {
		sort.SliceStable(slowInteractions, func(i, j int) bool {
			return siteQualityRuntimeFloatValue(slowInteractions[i].ResponseMilliseconds) >
				siteQualityRuntimeFloatValue(slowInteractions[j].ResponseMilliseconds)
		})
		maxResponse := siteQualityRuntimeFloatValue(slowInteractions[0].ResponseMilliseconds)
		issues = append(issues, LighthouseRunnerIssue{
			ID:           siteQualityInteractionLatencyAuditID,
			Kind:         "interaction",
			RuleVersion:  siteQualityAuditRuleVersion,
			Title:        "Configured user interaction is too slow",
			Description:  "A configured browser interaction exceeded its response budget after the page hydrated.",
			Severity:     siteQualityLatencySeverity(maxResponse),
			DisplayValue: fmt.Sprintf("Slowest interaction %.0f ms", maxResponse),
			NumericValue: &maxResponse,
			SavingsMS:    &maxResponse,
			Runtime:      siteQualityInteractionRuntimeEvidence(slowInteractions),
		})
	}
	return issues
}

func siteQualityInteractionProbeFailureIssue(interactions []siteQualityInteractionProbe) LighthouseRunnerIssue {
	count := float64(len(interactions))
	return LighthouseRunnerIssue{
		ID:           siteQualityInteractionAuditFailedID,
		Kind:         "interaction",
		RuleVersion:  siteQualityAuditRuleVersion,
		Title:        "Configured user interaction probe failed",
		Description:  "A configured browser interaction probe failed before response latency could be measured.",
		Severity:     "medium",
		DisplayValue: fmt.Sprintf("%d interaction probe(s) failed", len(interactions)),
		NumericValue: &count,
		Runtime:      siteQualityInteractionRuntimeEvidence(interactions),
	}
}

func siteQualityInteractionRuntimeEvidence(interactions []siteQualityInteractionProbe) []sitequalitydomain.SiteQualityRuntimeEvidence {
	evidence := make([]sitequalitydomain.SiteQualityRuntimeEvidence, 0, len(interactions))
	for _, interaction := range interactions {
		evidence = append(evidence, sitequalitydomain.SiteQualityRuntimeEvidence{
			Name:        strings.TrimSpace(interaction.Name),
			Selector:    strings.TrimSpace(interaction.Selector),
			Action:      strings.TrimSpace(interaction.Action),
			Status:      strings.TrimSpace(interaction.Status),
			Source:      strings.TrimSpace(interaction.MetricSource),
			Error:       strings.TrimSpace(interaction.Error),
			ResponseMS:  copyFloat64(interaction.ResponseMilliseconds),
			ThresholdMS: copyFloat64(interaction.ThresholdMilliseconds),
		})
	}
	return evidence
}

func siteQualitySoftNavigationAuditIssues(audit *siteQualitySoftNavigationAudit) []LighthouseRunnerIssue {
	if audit == nil {
		return nil
	}
	status := strings.TrimSpace(audit.Status)
	if !audit.Configured || siteQualityRuntimeAuditSkipped(status) {
		return nil
	}
	if status != "complete" && !siteQualityRuntimeAuditTimeoutSkipped(status) {
		return []LighthouseRunnerIssue{
			siteQualityRuntimeScanFailureIssue(
				siteQualitySoftNavigationAuditFailedID,
				"navigation",
				"Soft navigation audit did not complete",
				audit.Error,
			),
		}
	}

	failures := make([]siteQualitySoftNavigationResult, 0)
	regressions := make([]siteQualitySoftNavigationResult, 0)
	for _, navigation := range audit.Navigations {
		navigationStatus := strings.TrimSpace(navigation.Status)
		if siteQualityRuntimeAuditTimeoutSkipped(navigationStatus) {
			continue
		}
		if navigationStatus != "complete" {
			failures = append(failures, navigation)
			continue
		}
		if navigation.Exceeded || strings.TrimSpace(navigation.Mode) == "hard-navigation" {
			regressions = append(regressions, navigation)
		}
	}
	if len(failures) == 0 && len(regressions) == 0 {
		return nil
	}
	issues := make([]LighthouseRunnerIssue, 0, 2)
	if len(failures) > 0 {
		issues = append(issues, siteQualitySoftNavigationProbeFailureIssue(failures))
	}
	if len(regressions) > 0 {
		sort.SliceStable(regressions, func(i, j int) bool {
			return siteQualityRuntimeFloatValue(regressions[i].DurationMilliseconds) >
				siteQualityRuntimeFloatValue(regressions[j].DurationMilliseconds)
		})
		maxDuration := siteQualityRuntimeFloatValue(regressions[0].DurationMilliseconds)
		issues = append(issues, LighthouseRunnerIssue{
			ID:           siteQualitySoftNavigationRegressionID,
			Kind:         "navigation",
			RuleVersion:  siteQualityAuditRuleVersion,
			Title:        "Client-side route transition regressed",
			Description:  "A configured NuxtLink route transition was slow, leaked too much heap, or fell back to hard navigation.",
			Severity:     siteQualityLatencySeverity(maxDuration),
			DisplayValue: fmt.Sprintf("Slowest route transition %.0f ms", maxDuration),
			NumericValue: &maxDuration,
			SavingsMS:    &maxDuration,
			Runtime:      siteQualitySoftNavigationRuntimeEvidence(regressions),
		})
	}
	return issues
}

func siteQualitySoftNavigationProbeFailureIssue(navigations []siteQualitySoftNavigationResult) LighthouseRunnerIssue {
	count := float64(len(navigations))
	return LighthouseRunnerIssue{
		ID:           siteQualitySoftNavigationAuditFailedID,
		Kind:         "navigation",
		RuleVersion:  siteQualityAuditRuleVersion,
		Title:        "Configured soft navigation probe failed",
		Description:  "A configured NuxtLink route probe failed before route transition performance could be evaluated.",
		Severity:     "medium",
		DisplayValue: fmt.Sprintf("%d soft navigation probe(s) failed", len(navigations)),
		NumericValue: &count,
		Runtime:      siteQualitySoftNavigationRuntimeEvidence(navigations),
	}
}

func siteQualitySoftNavigationRuntimeEvidence(navigations []siteQualitySoftNavigationResult) []sitequalitydomain.SiteQualityRuntimeEvidence {
	evidence := make([]sitequalitydomain.SiteQualityRuntimeEvidence, 0, len(navigations))
	for _, navigation := range navigations {
		evidence = append(evidence, sitequalitydomain.SiteQualityRuntimeEvidence{
			Selector:             strings.TrimSpace(navigation.Selector),
			Text:                 strings.TrimSpace(navigation.Text),
			URL:                  strings.TrimSpace(navigation.FromURL),
			FinalURL:             strings.TrimSpace(navigation.ToURL),
			ExpectedURL:          strings.TrimSpace(navigation.ExpectedURL),
			Status:               strings.TrimSpace(navigation.Status),
			Mode:                 strings.TrimSpace(navigation.Mode),
			Error:                strings.TrimSpace(navigation.Error),
			DurationMS:           copyFloat64(navigation.DurationMilliseconds),
			ThresholdMS:          copyFloat64(navigation.ThresholdMilliseconds),
			JSHeapDeltaBytes:     copyInt64(navigation.JSHeapDeltaBytes),
			JSHeapThresholdBytes: copyInt64(navigation.JSHeapDeltaThresholdBytes),
		})
	}
	return evidence
}

func siteQualityRuntimeScanFailureIssue(id string, kind string, title string, reason string) LighthouseRunnerIssue {
	description := strings.TrimSpace(reason)
	if description == "" {
		description = "The browser-rendered runtime audit did not complete."
	}
	return LighthouseRunnerIssue{
		ID:          id,
		Kind:        kind,
		RuleVersion: siteQualityAuditRuleVersion,
		Title:       title,
		Description: description,
		Severity:    "medium",
		Runtime: []sitequalitydomain.SiteQualityRuntimeEvidence{
			{Status: "failed", Error: description},
		},
	}
}

func siteQualityRuntimeAuditSkipped(status string) bool {
	return status == "skipped"
}

func siteQualityRuntimeAuditTimeoutSkipped(status string) bool {
	return status == "timeout_skipped"
}

func siteQualityLatencySeverity(milliseconds float64) string {
	switch {
	case milliseconds >= 500:
		return "critical"
	case milliseconds >= 250:
		return "high"
	default:
		return "medium"
	}
}

func siteQualityRuntimeFloatValue(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}
