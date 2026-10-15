package service

import (
	"context"
	"strconv"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/repository"

	"go.uber.org/zap"
)

type DashboardBusinessFollowUpBusinessIdentityReader interface {
	ReadDashboardBusinessIdentitiesByAggregateType(context.Context, string, []uint, repository.DashboardOutboxBusinessReadScope) (map[uint]repository.DashboardBusinessFollowUpBusinessIdentity, error)
}

type DashboardBusinessFollowUpReferenceResolution struct {
	References                     map[uint]repository.DashboardBusinessFollowUpBusinessIdentity
	UnavailableReferences          map[uint]bool
	DegradedBusinessAggregateTypes []string
}

type DashboardBusinessFollowUpBusinessReferenceResolver struct {
	identities DashboardBusinessFollowUpBusinessIdentityReader
}

func NewDashboardBusinessFollowUpBusinessReferenceResolver(identities DashboardBusinessFollowUpBusinessIdentityReader) *DashboardBusinessFollowUpBusinessReferenceResolver {
	return &DashboardBusinessFollowUpBusinessReferenceResolver{identities: identities}
}

// ResolveDashboardBusinessFollowUpReferences preserves successful domains when
// another lookup fails. Permission scope applies to every batch lookup.
func (r *DashboardBusinessFollowUpBusinessReferenceResolver) ResolveDashboardBusinessFollowUpReferences(ctx context.Context, items []repository.DashboardOutboxFailureSummaryItem, scope repository.DashboardOutboxBusinessReadScope) DashboardBusinessFollowUpReferenceResolution {
	result := DashboardBusinessFollowUpReferenceResolution{
		References:                     map[uint]repository.DashboardBusinessFollowUpBusinessIdentity{},
		UnavailableReferences:          map[uint]bool{},
		DegradedBusinessAggregateTypes: []string{},
	}
	idsByKind := map[string][]uint{}
	for _, item := range items {
		if id, err := strconv.ParseUint(item.AggregateID, 10, 32); err == nil && id > 0 {
			idsByKind[item.AggregateType] = append(idsByKind[item.AggregateType], uint(id))
		}
	}
	identitiesByKind := map[string]map[uint]repository.DashboardBusinessFollowUpBusinessIdentity{}
	unavailableKinds := map[string]bool{}
	for _, kind := range []string{outbox.AggregateTypeAfterSalesCase, outbox.AggregateTypePayment, outbox.AggregateTypeProduct, outbox.AggregateTypeMerchantOffer, outbox.AggregateTypeCustomerServiceConversation, outbox.AggregateTypeOrder} {
		if len(idsByKind[kind]) == 0 {
			continue
		}
		if r == nil || r.identities == nil || ctx.Err() != nil {
			unavailableKinds[kind] = true
			result.DegradedBusinessAggregateTypes = append(result.DegradedBusinessAggregateTypes, kind)
			continue
		}
		identities, err := r.identities.ReadDashboardBusinessIdentitiesByAggregateType(ctx, kind, idsByKind[kind], scope)
		if err != nil {
			logger.Error("dashboard business identity batch unavailable", zap.String("aggregate_type", kind), zap.Error(err))
			unavailableKinds[kind] = true
			result.DegradedBusinessAggregateTypes = append(result.DegradedBusinessAggregateTypes, kind)
			continue
		}
		identitiesByKind[kind] = identities
		if kind == outbox.AggregateTypeAfterSalesCase || kind == outbox.AggregateTypePayment {
			for _, identity := range identities {
				if identity.OrderID > 0 {
					idsByKind[outbox.AggregateTypeOrder] = append(idsByKind[outbox.AggregateTypeOrder], identity.OrderID)
				}
			}
		}
	}
	for _, item := range items {
		id, err := strconv.ParseUint(item.AggregateID, 10, 32)
		if err != nil || id == 0 {
			continue
		}
		if unavailableKinds[item.AggregateType] || ctx.Err() != nil {
			result.UnavailableReferences[item.ID] = true
			continue
		}
		identity, exists := identitiesByKind[item.AggregateType][uint(id)]
		if !exists {
			continue
		}
		if identity.OrderID > 0 && identity.Kind != outbox.AggregateTypeOrder {
			if unavailableKinds[outbox.AggregateTypeOrder] {
				result.UnavailableReferences[item.ID] = true
			} else {
				identity.OrderNumber = identitiesByKind[outbox.AggregateTypeOrder][identity.OrderID].OrderNumber
			}
		}
		result.References[item.ID] = identity
	}
	return result
}
