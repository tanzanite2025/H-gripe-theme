package service

import (
	"errors"
	"time"

	"commerce-platform/internal/repository"
)

func (s *SiteQualityEngineService) CleanupTerminalJobs() (repository.SiteQualityJobCleanupResult, error) {
	var result repository.SiteQualityJobCleanupResult
	if s == nil || s.jobs == nil {
		return result, errors.New("SiteQuality quality engine is unavailable")
	}
	return s.jobs.DeleteTerminalJobs()
}

func (s *SiteQualityEngineService) CleanupOldFindings(cutoff time.Time) (repository.SiteQualityFindingCleanupResult, error) {
	var result repository.SiteQualityFindingCleanupResult
	if s == nil || s.findings == nil {
		return result, errors.New("SiteQuality quality engine is unavailable")
	}
	return s.findings.DeleteOld(cutoff)
}
