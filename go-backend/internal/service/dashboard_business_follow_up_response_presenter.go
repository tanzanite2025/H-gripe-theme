package service

import (
	"fmt"
	"strings"

	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"
)

func formatDashboardBusinessFollowUpBusinessIdentityLabel(reference repository.DashboardBusinessFollowUpBusinessIdentity) string {
	var label string
	switch reference.Kind {
	case outbox.AggregateTypeOrder:
		return "订单 " + reference.OrderNumber
	case outbox.AggregateTypeAfterSalesCase:
		label = fmt.Sprintf("售后案件 #%d", reference.ID)
	case outbox.AggregateTypePayment:
		label = fmt.Sprintf("退款记录 #%d", reference.ID)
	case outbox.AggregateTypeProduct:
		label = fmt.Sprintf("商品 #%d", reference.ID)
	case outbox.AggregateTypeMerchantOffer:
		label = fmt.Sprintf("渠道商品 #%d", reference.ID)
	case outbox.AggregateTypeCustomerServiceConversation:
		return "客服会话 " + reference.TicketNumber
	}
	if reference.Name != "" {
		label += " · " + reference.Name
	}
	if reference.OrderNumber != "" {
		label += " · 订单 " + reference.OrderNumber
	}
	return label
}

func describeDashboardBusinessFollowUpFailureReason(recordedError string) string {
	normalized := strings.ToLower(recordedError)
	switch {
	case strings.Contains(normalized, "no handler"), strings.Contains(normalized, "not configured"):
		return "任务所需的处理配置不完整，需要对应负责人核查。"
	case strings.Contains(normalized, "timeout"), strings.Contains(normalized, "deadline exceeded"):
		return "执行时等待服务响应超时，需要核实服务恢复情况及实际执行结果。"
	case strings.Contains(normalized, "unauthorized"), strings.Contains(normalized, "401"), strings.Contains(normalized, "403"), strings.Contains(normalized, "authentication"):
		return "服务返回了授权或访问错误，需要对应负责人核查账号授权与配置。"
	case strings.Contains(normalized, "connection refused"), strings.Contains(normalized, "network"), strings.Contains(normalized, "connection reset"):
		return "执行时未能正常连接服务，需要对应负责人检查连接。"
	case strings.Contains(normalized, "invalid"), strings.Contains(normalized, "incomplete"), strings.Contains(normalized, "missing"):
		return "执行所需的数据未通过检查，需要对应负责人核对资料。"
	default:
		return "系统记录了执行错误，具体原因需要对应负责人进一步排查。"
	}
}

func buildDashboardBusinessFollowUpItem(item repository.DashboardOutboxFailureSummaryItem, reference *repository.DashboardBusinessFollowUpBusinessIdentity, viewerRole auth.Role) DashboardBusinessFollowUpItem {
	definition, classified := dashboardBusinessFollowUpTaskDefinitions[item.EventType]
	if !classified {
		definition = DashboardBusinessFollowUpTaskDefinition{Title: "系统后台任务执行未完成", Category: "系统维护", Impact: "这项任务需要系统负责人判断具体影响。", ResponsibleTeam: "系统负责人", SalesNextSteps: []string{"将跟进说明交给系统负责人核查；业务是否完成需在对应记录中确认。"}}
	}
	result := DashboardBusinessFollowUpItem{ID: item.ID, Title: definition.Title, Category: definition.Category, BusinessImpact: definition.Impact, ResponsibleTeam: definition.ResponsibleTeam, SalesNextSteps: definition.SalesNextSteps, ReasonSummary: describeDashboardBusinessFollowUpFailureReason(item.LastError), ProcessingState: "manual_attention", Attempts: item.Attempts, MaxAttempts: item.MaxAttempts, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt, LastAttemptAt: item.LastAttemptAt, RelatedRecordNotice: "当前未能确认关联业务记录，请将事项编号交给对应负责人定位。"}
	if item.Status == outbox.EventStatusFailed && item.Attempts < item.MaxAttempts {
		result.ProcessingState = "automatic_retry_scheduled"
		result.NextRetryAt = &item.AvailableAt
	}
	if reference == nil {
		return result
	}
	related := buildDashboardBusinessFollowUpRelatedRecordNavigation(*reference, definition, viewerRole)
	result.RelatedRecord = &related
	result.RelatedRecordNotice = ""
	if related.Path == "" {
		result.RelatedRecordNotice = "已确认业务标识，但当前未能定位可打开的记录。请把事项编号与业务标识交给对应负责人核查。"
	}
	return result
}
