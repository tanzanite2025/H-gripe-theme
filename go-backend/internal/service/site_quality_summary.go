package service

import (
	"errors"
	"fmt"
	"time"

	sitequalitydomain "commerce-platform/internal/domain/sitequality"
	"commerce-platform/internal/repository"
)

func (s *SiteQualityEngineService) Summary(now time.Time) (*SiteQualityOperationalSummary, error) {
	if s == nil {
		return nil, errors.New("SiteQuality quality engine is unavailable")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	summary := &SiteQualityOperationalSummary{
		GeneratedAt:                    now,
		Status:                         "healthy",
		WorkerEnabled:                  s.cfg.WorkerEnabled,
		AutoScanEnabled:                s.cfg.AutoScanEnabled,
		WorkerIntervalSeconds:          int(s.cfg.WorkerInterval.Seconds()),
		ReleaseID:                      s.cfg.ReleaseID,
		SampleCount:                    s.cfg.SampleCount,
		RequiredConfirmations:          s.cfg.RequiredConfirmations,
		RequiredCleanEvaluations:       s.cfg.RequiredCleanEvaluations,
		WorkerBatchLimit:               s.cfg.WorkerBatchLimit,
		LeaseTimeoutSeconds:            int(s.cfg.LeaseTimeout.Seconds()),
		ProviderConcurrency:            s.cfg.ProviderConcurrency,
		ProviderRequestIntervalSeconds: int(s.cfg.ProviderRequestInterval.Seconds()),
	}
	warnings := make([]string, 0, 4)
	if !summary.WorkerEnabled {
		warnings = append(warnings, "Site Quality job worker is disabled; queued inspections will remain pending")
	}

	if s.lighthouseRunner != nil {
		runnerSummary, err := s.lighthouseRunner.List(repository.SiteQualityRunListFilter{
			Page:     1,
			PageSize: 1,
		})
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("runner summary unavailable: %v", err))
		} else {
			summary.RunnerConfigured = runnerSummary.RunnerConfigured
			summary.DefaultURL = runnerSummary.DefaultURL
			if defaultTargetURL, defaultErr := s.defaultSiteQualityTargetURL(); defaultErr == nil && defaultTargetURL != "" {
				summary.DefaultURL = defaultTargetURL
			}
			summary.RunCount = runnerSummary.Total
			if len(runnerSummary.Items) > 0 {
				latest := runnerSummary.Items[0]
				summary.LatestRun = &latest
				if latest.Status != sitequalitydomain.SiteQualityRunStatusSuccess {
					warnings = append(warnings, "latest Lighthouse sample failed")
				}
			}
			latestSuccessAt, err := s.lighthouseRunner.LatestSuccessfulAt()
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("latest successful sample is unavailable: %v", err))
			} else {
				summary.LatestSuccessAt = latestSuccessAt
			}
		}
	} else {
		warnings = append(warnings, "internal Lighthouse runner is unavailable")
	}

	if s.targets != nil {
		targetStats, err := s.targets.Stats(now)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("target stats unavailable: %v", err))
		} else {
			summary.Targets = targetStats
		}
	} else {
		warnings = append(warnings, "site quality targets are unavailable")
	}

	if s.jobs != nil {
		jobStats, err := s.jobs.Stats(now, s.cfg.LeaseTimeout)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("job stats unavailable: %v", err))
		} else {
			summary.Jobs = jobStats
			if jobStats.Failed > 0 {
				warnings = append(warnings, fmt.Sprintf("%d Site Quality job(s) are awaiting retry", jobStats.Failed))
			}
			if jobStats.DeadLetter > 0 {
				warnings = append(warnings, fmt.Sprintf("%d Site Quality job(s) reached dead letter", jobStats.DeadLetter))
			}
			if jobStats.StaleLeases > 0 {
				warnings = append(warnings, fmt.Sprintf("%d Site Quality job lease(s) are stale", jobStats.StaleLeases))
			}
		}
		slotStats, err := s.jobs.ProviderSlotStats("lighthouse_runner", s.cfg.ProviderConcurrency, now, s.cfg.LeaseTimeout)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("provider slot stats unavailable: %v", err))
		} else {
			summary.ProviderSlots = slotStats
			if slotStats.StaleLocked > 0 {
				warnings = append(warnings, fmt.Sprintf("%d Lighthouse runner slot lease(s) are stale", slotStats.StaleLocked))
			}
		}
	} else {
		warnings = append(warnings, "site quality jobs are unavailable")
	}

	if s.findings != nil {
		findingStats, err := s.findings.Stats()
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("finding stats unavailable: %v", err))
		} else {
			summary.Findings = findingStats
		}
	} else {
		warnings = append(warnings, "site quality findings are unavailable")
	}

	summary.Warnings = warnings
	summary.Status = siteQualityOperationalStatus(summary, warnings)
	return summary, nil
}

func siteQualityOperationalStatus(summary *SiteQualityOperationalSummary, warnings []string) string {
	if summary == nil {
		return "unavailable"
	}
	if !summary.RunnerConfigured {
		return "not_configured"
	}
	if !summary.WorkerEnabled {
		return "degraded"
	}
	if len(warnings) > 0 || summary.Jobs.DeadLetter > 0 || summary.Jobs.StaleLeases > 0 || summary.ProviderSlots.StaleLocked > 0 {
		return "degraded"
	}
	return "healthy"
}
