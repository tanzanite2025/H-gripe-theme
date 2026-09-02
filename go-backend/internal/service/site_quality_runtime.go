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
	if !audit.Configured || strings.TrimSpace(audit.Status) == "skipped" {
		return nil
	}
	if strings.TrimSpace(audit.Status) != "complete" {
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
	if len(offenders) > 12 {
		offenders = offenders[:12]
	}
	count := float64(len(offenders))
	return []LighthouseRunnerIssue{
		{
			ID:           siteQualityBrokenLinkAuditID,
			Kind:         "links",
			RuleVersion:  siteQualityAuditRuleVersion,
			Title:        "Rendered page contains broken links",
			Description:  "The final browser-rendered DOM includes links that return HTTP errors, fail to respond, or redirect outside the configured storefront origin.",
			Severity:     siteQualityBrokenLinkSeverity(offenders),
			DisplayValue: fmt.Sprintf("%d broken link(s)", len(offenders)),
			NumericValue: &count,
			Links:        siteQualityBrokenLinkEvidence(offenders),
			Runtime:      siteQualityBrokenLinkRuntimeEvidence(offenders),
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
	if audit == nil || !audit.Configured || strings.TrimSpace(audit.Status) == "skipped" {
		return nil
	}
	if strings.TrimSpace(audit.Status) != "complete" {
		return []LighthouseRunnerIssue{
			siteQualityRuntimeScanFailureIssue(
				siteQualityInteractionAuditFailedID,
				"interaction",
				"Interaction latency audit did not complete",
				audit.Error,
			),
		}
	}

	offenders := make([]siteQualityInteractionProbe, 0)
	for _, interaction := range audit.Interactions {
		if strings.TrimSpace(interaction.Status) != "complete" || interaction.Exceeded {
			offenders = append(offenders, interaction)
		}
	}
	if len(offenders) == 0 {
		return nil
	}
	sort.SliceStable(offenders, func(i, j int) bool {
		return siteQualityRuntimeFloatValue(offenders[i].ResponseMilliseconds) >
			siteQualityRuntimeFloatValue(offenders[j].ResponseMilliseconds)
	})
	maxResponse := siteQualityRuntimeFloatValue(offenders[0].ResponseMilliseconds)
	return []LighthouseRunnerIssue{
		{
			ID:           siteQualityInteractionLatencyAuditID,
			Kind:         "interaction",
			RuleVersion:  siteQualityAuditRuleVersion,
			Title:        "Configured user interaction is too slow",
			Description:  "A configured browser interaction exceeded its response budget or failed to complete after the page hydrated.",
			Severity:     siteQualityLatencySeverity(maxResponse),
			DisplayValue: fmt.Sprintf("Slowest interaction %.0f ms", maxResponse),
			NumericValue: &maxResponse,
			SavingsMS:    &maxResponse,
			Runtime:      siteQualityInteractionRuntimeEvidence(offenders),
		},
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
	if audit == nil || !audit.Configured || strings.TrimSpace(audit.Status) == "skipped" {
		return nil
	}
	if strings.TrimSpace(audit.Status) != "complete" {
		return []LighthouseRunnerIssue{
			siteQualityRuntimeScanFailureIssue(
				siteQualitySoftNavigationAuditFailedID,
				"navigation",
				"Soft navigation audit did not complete",
				audit.Error,
			),
		}
	}

	offenders := make([]siteQualitySoftNavigationResult, 0)
	for _, navigation := range audit.Navigations {
		if strings.TrimSpace(navigation.Status) != "complete" ||
			navigation.Exceeded ||
			strings.TrimSpace(navigation.Mode) == "hard-navigation" {
			offenders = append(offenders, navigation)
		}
	}
	if len(offenders) == 0 {
		return nil
	}
	sort.SliceStable(offenders, func(i, j int) bool {
		return siteQualityRuntimeFloatValue(offenders[i].DurationMilliseconds) >
			siteQualityRuntimeFloatValue(offenders[j].DurationMilliseconds)
	})
	maxDuration := siteQualityRuntimeFloatValue(offenders[0].DurationMilliseconds)
	return []LighthouseRunnerIssue{
		{
			ID:           siteQualitySoftNavigationRegressionID,
			Kind:         "navigation",
			RuleVersion:  siteQualityAuditRuleVersion,
			Title:        "Client-side route transition regressed",
			Description:  "A configured NuxtLink route transition was slow, leaked too much heap, fell back to hard navigation, or failed to change route.",
			Severity:     siteQualityLatencySeverity(maxDuration),
			DisplayValue: fmt.Sprintf("Slowest route transition %.0f ms", maxDuration),
			NumericValue: &maxDuration,
			SavingsMS:    &maxDuration,
			Runtime:      siteQualitySoftNavigationRuntimeEvidence(offenders),
		},
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
