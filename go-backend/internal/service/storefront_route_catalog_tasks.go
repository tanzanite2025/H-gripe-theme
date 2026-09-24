package service

import (
	"context"
	"errors"
	"time"

	"commerce-platform/internal/repository"

	"github.com/google/uuid"
)

const storefrontRouteCatalogTaskTimeout = 15 * time.Minute

func (s *StorefrontRouteCatalogService) StartCheck(
	filter repository.StorefrontRouteCatalogListFilter,
	limit int,
) (StorefrontRouteCatalogCheckTask, error) {
	if s == nil || s.repository == nil {
		return StorefrontRouteCatalogCheckTask{}, errors.New("storefront route catalog service is unavailable")
	}
	if limit < 1 {
		limit = 100
	} else if limit > 200 {
		limit = 200
	}

	now := time.Now().UTC()
	task := &storefrontRouteCatalogCheckTask{data: StorefrontRouteCatalogCheckTask{
		ID:        "url-check-" + uuid.NewString(),
		Status:    StorefrontRouteCatalogCheckTaskQueued,
		StartedAt: now,
		UpdatedAt: now,
		Locale:    filter.Locale,
		Summary:   StorefrontRouteCatalogCheckSummary{},
	}}

	s.tasksMu.Lock()
	if s.checkTasks == nil {
		s.checkTasks = make(map[string]*storefrontRouteCatalogCheckTask)
	}
	s.checkTasks[task.data.ID] = task
	snapshot := task.snapshot()
	s.tasksMu.Unlock()

	go s.runCheckTask(task, filter, limit)
	return snapshot, nil
}

func (s *StorefrontRouteCatalogService) runCheckTask(
	task *storefrontRouteCatalogCheckTask,
	filter repository.StorefrontRouteCatalogListFilter,
	limit int,
) {
	task.update(func(data *StorefrontRouteCatalogCheckTask) {
		data.Status = StorefrontRouteCatalogCheckTaskRunning
		data.UpdatedAt = time.Now().UTC()
	})

	ctx, cancel := context.WithTimeout(context.Background(), storefrontRouteCatalogTaskTimeout)
	defer cancel()

	summary, err := s.checkBatch(ctx, filter, limit, func(progress StorefrontRouteCatalogCheckSummary) {
		task.update(func(data *StorefrontRouteCatalogCheckTask) {
			data.Summary = progress
			data.Checked = progress.Checked
			data.Eligible = progress.Eligible
			data.Remaining = progress.Remaining
			data.UpdatedAt = time.Now().UTC()
		})
	})
	if err != nil {
		task.update(func(data *StorefrontRouteCatalogCheckTask) {
			data.Status = StorefrontRouteCatalogCheckTaskFailed
			data.Error = err.Error()
			data.Summary = summary
			data.Checked = summary.Checked
			data.Eligible = summary.Eligible
			data.Remaining = summary.Remaining
			endedAt := time.Now().UTC()
			data.EndedAt = &endedAt
			data.UpdatedAt = endedAt
		})
		return
	}

	task.update(func(data *StorefrontRouteCatalogCheckTask) {
		data.Status = StorefrontRouteCatalogCheckTaskCompleted
		data.Summary = summary
		data.Checked = summary.Checked
		data.Eligible = summary.Eligible
		data.Remaining = summary.Remaining
		endedAt := time.Now().UTC()
		data.EndedAt = &endedAt
		data.UpdatedAt = endedAt
	})
}

func (s *StorefrontRouteCatalogService) GetCheckTask(taskID string) (StorefrontRouteCatalogCheckTask, error) {
	if s == nil {
		return StorefrontRouteCatalogCheckTask{}, errors.New("storefront route catalog service is unavailable")
	}
	s.tasksMu.RLock()
	task := s.checkTasks[taskID]
	s.tasksMu.RUnlock()
	if task == nil {
		return StorefrontRouteCatalogCheckTask{}, errors.New("URL check task not found")
	}
	return task.snapshot(), nil
}

func (task *storefrontRouteCatalogCheckTask) update(update func(*StorefrontRouteCatalogCheckTask)) {
	task.mu.Lock()
	defer task.mu.Unlock()
	update(&task.data)
}

func (task *storefrontRouteCatalogCheckTask) snapshot() StorefrontRouteCatalogCheckTask {
	task.mu.RLock()
	defer task.mu.RUnlock()
	return task.data
}
