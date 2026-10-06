package service

import (
	"context"
	"errors"
	"time"

	"commerce-platform/internal/repository"

	"github.com/google/uuid"
)

const storefrontRouteCatalogTaskMinimumTimeout = 15 * time.Minute
const storefrontRouteCatalogPersistenceAllowance = 250 * time.Millisecond

func (s *StorefrontRouteCatalogService) StartCheck(
	filter repository.StorefrontRouteCatalogListFilter,
	batchSize int,
) (StorefrontRouteCatalogCheckTask, error) {
	if s == nil || s.repository == nil {
		return StorefrontRouteCatalogCheckTask{}, errors.New("storefront route catalog service is unavailable")
	}
	releaseOperation, err := s.beginCatalogOperation()
	if err != nil {
		return StorefrontRouteCatalogCheckTask{}, err
	}
	if batchSize < 1 || batchSize > 200 {
		batchSize = 200
	}
	filter.CheckableOnly = true
	filter.Page = 1
	filter.PageSize = 1
	_, eligible, err := s.repository.List(filter)
	if err != nil {
		releaseOperation()
		return StorefrontRouteCatalogCheckTask{}, err
	}
	filter.PageSize = batchSize

	taskTimeout := storefrontRouteCatalogCheckTaskTimeout(eligible)
	now := time.Now().UTC()
	eligibleCount := int(eligible)
	totalBatches := 0
	if eligibleCount > 0 {
		totalBatches = (eligibleCount + batchSize - 1) / batchSize
	}
	task := &storefrontRouteCatalogCheckTask{data: StorefrontRouteCatalogCheckTask{
		ID:        "url-check-" + uuid.NewString(),
		Status:    StorefrontRouteCatalogCheckTaskQueued,
		StartedAt: now,
		UpdatedAt: now,
		Locale:    filter.Locale,
		Eligible:  eligibleCount,
		Remaining: eligibleCount,
		BatchSize: batchSize,
		Summary: StorefrontRouteCatalogCheckSummary{
			Eligible:     eligibleCount,
			Remaining:    eligibleCount,
			BatchSize:    batchSize,
			TotalBatches: totalBatches,
		},
	}}

	s.tasksMu.Lock()
	if s.checkTasks == nil {
		s.checkTasks = make(map[string]*storefrontRouteCatalogCheckTask)
	}
	s.checkTasks[task.data.ID] = task
	snapshot := task.snapshot()
	s.tasksMu.Unlock()

	go func() {
		defer releaseOperation()
		s.runCheckTask(task, filter, batchSize, taskTimeout)
	}()
	return snapshot, nil
}

func (s *StorefrontRouteCatalogService) runCheckTask(
	task *storefrontRouteCatalogCheckTask,
	filter repository.StorefrontRouteCatalogListFilter,
	batchSize int,
	taskTimeout time.Duration,
) {
	task.update(func(data *StorefrontRouteCatalogCheckTask) {
		data.Status = StorefrontRouteCatalogCheckTaskRunning
		data.UpdatedAt = time.Now().UTC()
	})

	ctx, cancel := context.WithTimeout(context.Background(), taskTimeout)
	defer cancel()

	summary, err := s.checkBatch(ctx, filter, batchSize, func(progress StorefrontRouteCatalogCheckSummary) {
		task.update(func(data *StorefrontRouteCatalogCheckTask) {
			data.Summary = progress
			data.Checked = progress.Checked
			data.Eligible = progress.Eligible
			data.Remaining = progress.Remaining
			data.BatchSize = progress.BatchSize
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
			data.BatchSize = summary.BatchSize
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
		data.BatchSize = summary.BatchSize
		endedAt := time.Now().UTC()
		data.EndedAt = &endedAt
		data.UpdatedAt = endedAt
	})
}

func storefrontRouteCatalogCheckTaskTimeout(eligible int64) time.Duration {
	if eligible < 1 {
		return storefrontRouteCatalogTaskMinimumTimeout
	}

	workerCount := int64(routeCheckConcurrency)
	requestWaves := (eligible + workerCount - 1) / workerCount
	requestBudget := time.Duration(requestWaves) * storefrontRouteCatalogRequestTimeout
	writeBudget := time.Duration(eligible) * storefrontRouteCatalogPersistenceAllowance
	estimated := requestBudget + writeBudget + time.Minute
	if estimated < storefrontRouteCatalogTaskMinimumTimeout {
		return storefrontRouteCatalogTaskMinimumTimeout
	}
	return estimated + estimated/2
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
