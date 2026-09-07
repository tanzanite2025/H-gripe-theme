package admin

import "commerce-platform/internal/service"

type productQualityRequirementRequest struct {
	VariantID              *uint  `json:"variant_id"`
	SpokeTensionQCRequired bool   `json:"spoke_tension_qc_required"`
	Status                 string `json:"status"`
	RuleVersion            string `json:"rule_version" binding:"required"`
	Reason                 string `json:"reason"`
}

func (r productQualityRequirementRequest) toServiceInput(productID, actorID uint) service.ProductQualityRequirementInput {
	return service.ProductQualityRequirementInput{
		ProductID:              productID,
		VariantID:              r.VariantID,
		SpokeTensionQCRequired: r.SpokeTensionQCRequired,
		Status:                 r.Status,
		RuleVersion:            r.RuleVersion,
		Reason:                 r.Reason,
		CreatedBy:              actorID,
	}
}
