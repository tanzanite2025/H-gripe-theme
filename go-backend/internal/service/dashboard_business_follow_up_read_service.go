package service

import (
	"context"
	"errors"
	"time"

	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/repository"

	"go.uber.org/zap"
)

// These interfaces expose no mutations or worker lifecycle operations.
type DashboardBusinessFollowUpFailureSnapshotReader interface {
	GetDashboardOutboxFailureSnapshot(context.Context, int, repository.DashboardOutboxBusinessReadScope) (*repository.DashboardOutboxFailureSnapshot, error)
	FindDashboardBusinessFollowUpFailureByID(context.Context, uint, repository.DashboardOutboxBusinessReadScope) (*repository.DashboardOutboxFailureSummaryItem, error)
}

type DashboardBusinessFollowUpBusinessReferenceReader interface {
	ResolveDashboardBusinessFollowUpReferences(context.Context, []repository.DashboardOutboxFailureSummaryItem, repository.DashboardOutboxBusinessReadScope) DashboardBusinessFollowUpReferenceResolution
}

type DashboardBusinessFollowUpReadService struct {
	failures   DashboardBusinessFollowUpFailureSnapshotReader
	references DashboardBusinessFollowUpBusinessReferenceReader
}

func NewDashboardBusinessFollowUpReadService(failures DashboardBusinessFollowUpFailureSnapshotReader, references DashboardBusinessFollowUpBusinessReferenceReader) *DashboardBusinessFollowUpReadService {
	return &DashboardBusinessFollowUpReadService{failures: failures, references: references}
}

func (s *DashboardBusinessFollowUpReadService) GetCommandCenterOutboxFailures(ctx context.Context, viewerRole auth.Role, viewerUserID uint) DashboardCommandCenterResponse {
	return s.getDashboardBusinessFollowUpSnapshot(ctx, viewerRole, viewerUserID, 5)
}

func (s *DashboardBusinessFollowUpReadService) GetDashboardBusinessFollowUpList(ctx context.Context, viewerRole auth.Role, viewerUserID uint) DashboardCommandCenterResponse {
	return s.getDashboardBusinessFollowUpSnapshot(ctx, viewerRole, viewerUserID, 100)
}

func (s *DashboardBusinessFollowUpReadService) getDashboardBusinessFollowUpSnapshot(ctx context.Context, viewerRole auth.Role, viewerUserID uint, previewLimit int) DashboardCommandCenterResponse {
	response := DashboardCommandCenterResponse{
		Timestamp: time.Now().UTC(),
		ActionCenter: DashboardCommandCenterActionItems{OutboxFailures: DashboardCommandCenterOutboxFailures{
			Status: "unavailable", Items: []DashboardBusinessFollowUpItem{}, PreviewLimit: previewLimit,
		}},
		SystemDegradedNodes: []string{},
	}
	if s == nil || s.failures == nil {
		response.SystemDegradedNodes = append(response.SystemDegradedNodes, "outbox_failures")
		return response
	}
	scope := buildDashboardBusinessFollowUpReadScope(viewerRole, viewerUserID)
	snapshot, err := s.failures.GetDashboardOutboxFailureSnapshot(ctx, previewLimit, scope)
	if err != nil {
		logger.Error("dashboard command center outbox query failed", zap.Error(err))
		response.SystemDegradedNodes = append(response.SystemDegradedNodes, "outbox_failures")
		return response
	}
	items, contextUnavailable := s.describeDashboardBusinessFollowUpItems(ctx, snapshot.Items, viewerRole, scope)
	if contextUnavailable {
		response.SystemDegradedNodes = append(response.SystemDegradedNodes, "outbox_business_references")
	}
	automaticRetryCount := snapshot.TotalCount - snapshot.ManualAttentionCount
	response.ActionCenter.OutboxFailures = DashboardCommandCenterOutboxFailures{
		Status: "ready", TotalCount: &snapshot.TotalCount, FailedCount: &snapshot.FailedCount,
		DeadLetterCount: &snapshot.DeadLetterCount, ManualAttentionCount: &snapshot.ManualAttentionCount,
		AutomaticRetryCount: &automaticRetryCount, PreviewLimit: previewLimit, Items: items,
	}
	return response
}

func (s *DashboardBusinessFollowUpReadService) describeDashboardBusinessFollowUpItems(ctx context.Context, summaries []repository.DashboardOutboxFailureSummaryItem, viewerRole auth.Role, scope repository.DashboardOutboxBusinessReadScope) ([]DashboardBusinessFollowUpItem, bool) {
	permittedReferences := make([]repository.DashboardOutboxFailureSummaryItem, 0, len(summaries))
	for _, summary := range summaries {
		permission := dashboardBusinessFollowUpAggregateReadPermission(summary.AggregateType)
		if permission != "" && viewerRole.HasPermission(permission) {
			permittedReferences = append(permittedReferences, summary)
		}
	}
	resolution := DashboardBusinessFollowUpReferenceResolution{}
	if s.references != nil {
		resolution = s.references.ResolveDashboardBusinessFollowUpReferences(ctx, permittedReferences, scope)
	} else if len(permittedReferences) > 0 {
		resolution.UnavailableReferences = map[uint]bool{}
		for _, summary := range permittedReferences {
			resolution.UnavailableReferences[summary.ID] = true
		}
	}
	items := make([]DashboardBusinessFollowUpItem, 0, len(summaries))
	for _, summary := range summaries {
		var reference *repository.DashboardBusinessFollowUpBusinessIdentity
		if related, exists := resolution.References[summary.ID]; exists {
			reference = &related
		}
		item := buildDashboardBusinessFollowUpItem(summary, reference, viewerRole)
		if resolution.UnavailableReferences[summary.ID] {
			item.RelatedRecordNotice = "关联业务资料暂时无法完整读取，待办数量仍有效。请刷新重试，或将事项编号交给对应负责人定位。"
		}
		items = append(items, item)
	}
	return items, len(resolution.DegradedBusinessAggregateTypes) > 0 || len(resolution.UnavailableReferences) > 0
}

func (s *DashboardBusinessFollowUpReadService) GetDashboardBusinessFollowUpDetail(ctx context.Context, viewerRole auth.Role, viewerUserID, eventID uint) (*DashboardBusinessFollowUpItem, error) {
	if s == nil || s.failures == nil {
		return nil, errors.New("dashboard business follow-up reader is unavailable")
	}
	scope := buildDashboardBusinessFollowUpReadScope(viewerRole, viewerUserID)
	summary, err := s.failures.FindDashboardBusinessFollowUpFailureByID(ctx, eventID, scope)
	if err != nil {
		return nil, err
	}
	items, _ := s.describeDashboardBusinessFollowUpItems(ctx, []repository.DashboardOutboxFailureSummaryItem{*summary}, viewerRole, scope)
	return &items[0], nil
}
