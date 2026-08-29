package service

import (
	"errors"

	"commerce-platform/internal/repository"
)

func (s *SiteQualityEngineService) CleanupTerminalJobs() (repository.SiteQualityJobCleanupResult, error) {
	var result repository.SiteQualityJobCleanupResult
	if s == nil || s.jobs == nil {
		return result, errors.New("SiteQuality quality engine is unavailable")
	}
	return s.jobs.DeleteTerminalJobs()
}
