package service

import (
	"fmt"
	"net/url"

	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"
)

func buildDashboardBusinessFollowUpRelatedRecordNavigation(reference repository.DashboardBusinessFollowUpBusinessIdentity, definition DashboardBusinessFollowUpTaskDefinition, viewerRole auth.Role) DashboardBusinessFollowUpRelatedRecord {
	related := DashboardBusinessFollowUpRelatedRecord{Label: formatDashboardBusinessFollowUpBusinessIdentityLabel(reference)}
	switch {
	case reference.OrderID > 0 && reference.OrderNumber != "" && viewerRole.HasPermission(auth.PermOrderView):
		related.RequiredPermission = string(auth.PermOrderView)
		related.Path = fmt.Sprintf("/orders/list?order_id=%d", reference.OrderID)
	case definition.ReadPermission == auth.PermMerchantView && viewerRole.HasPermission(auth.PermMerchantView):
		related.RequiredPermission = string(auth.PermMerchantView)
		related.Path = "/google-merchant"
	case reference.Kind == outbox.AggregateTypeProduct && viewerRole.HasPermission(auth.PermProductView):
		related.RequiredPermission = string(auth.PermProductView)
		related.Path = "/catalog/products"
	case reference.Kind == outbox.AggregateTypeMerchantOffer && viewerRole.HasPermission(auth.PermMerchantView):
		related.RequiredPermission = string(auth.PermMerchantView)
		related.Path = "/google-merchant"
	case reference.Kind == outbox.AggregateTypeCustomerServiceConversation && viewerRole.HasPermission(auth.PermTicketView):
		related.RequiredPermission = string(auth.PermTicketView)
		related.Path = fmt.Sprintf("/support/conversations?conversation_id=%d&search=%s", reference.ID, url.QueryEscape(reference.TicketNumber))
	}
	return related
}
