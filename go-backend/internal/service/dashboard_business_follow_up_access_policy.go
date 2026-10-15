package service

import (
	"sort"

	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"
)

func GetDashboardBusinessFollowUpAllowedEventTypes(viewerRole auth.Role) []string {
	allowed := []string{}
	for eventType, definition := range dashboardBusinessFollowUpTaskDefinitions {
		if viewerRole.HasPermission(definition.ReadPermission) {
			allowed = append(allowed, eventType)
		}
	}
	sort.Strings(allowed)
	return allowed
}

func dashboardBusinessFollowUpAggregateReadPermission(aggregateType string) auth.Permission {
	switch aggregateType {
	case outbox.AggregateTypeOrder, outbox.AggregateTypeAfterSalesCase, outbox.AggregateTypePayment:
		return auth.PermOrderView
	case outbox.AggregateTypeProduct:
		return auth.PermProductView
	case outbox.AggregateTypeMerchantOffer:
		return auth.PermMerchantView
	case outbox.AggregateTypeCustomerServiceConversation:
		return auth.PermTicketView
	default:
		return ""
	}
}

func buildDashboardBusinessFollowUpReadScope(viewerRole auth.Role, viewerUserID uint) repository.DashboardOutboxBusinessReadScope {
	scope := repository.DashboardOutboxBusinessReadScope{CanViewAllEventTypes: viewerRole.HasPermission(auth.PermOpsView), AllowedEventTypes: GetDashboardBusinessFollowUpAllowedEventTypes(viewerRole)}
	for _, kind := range []string{outbox.AggregateTypeOrder, outbox.AggregateTypeAfterSalesCase, outbox.AggregateTypePayment, outbox.AggregateTypeProduct, outbox.AggregateTypeMerchantOffer, outbox.AggregateTypeCustomerServiceConversation} {
		if viewerRole.HasPermission(dashboardBusinessFollowUpAggregateReadPermission(kind)) {
			scope.AllowedBusinessAggregateTypes = append(scope.AllowedBusinessAggregateTypes, kind)
		}
	}
	if !scope.CanViewAllEventTypes && viewerRole.HasPermission(auth.PermTicketView) {
		scope.AssignedAgentUserID = &viewerUserID
	}
	return scope
}
