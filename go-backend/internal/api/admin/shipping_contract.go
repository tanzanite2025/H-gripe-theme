package admin

import (
	"strings"
	shippingdomain "tanzanite/internal/domain/shipping"
)

type shippingTemplateRequest struct {
	Name          string                `json:"name" binding:"required"`
	Type          string                `json:"type" binding:"required"`
	FreeShipping  bool                  `json:"free_shipping"`
	FreeThreshold float64               `json:"free_threshold"`
	DefaultFee    float64               `json:"default_fee"`
	Description   string                `json:"description"`
	Enabled       *bool                 `json:"enabled"`
	Rules         []shippingRuleRequest `json:"rules"`
}

type shippingRuleRequest struct {
	ID         uint    `json:"id"`
	Region     string  `json:"region"`
	MinValue   float64 `json:"min_value"`
	MaxValue   float64 `json:"max_value"`
	Fee        float64 `json:"fee"`
	Additional float64 `json:"additional"`
}

type shippingZoneRequest struct {
	Name        string `json:"name" binding:"required"`
	Countries   string `json:"countries"`
	States      string `json:"states"`
	PostalCodes string `json:"postal_codes"`
	Enabled     *bool  `json:"enabled"`
}

type shippingTemplateBindingRequest struct {
	TemplateID    uint   `json:"template_id" binding:"required"`
	Scope         string `json:"scope" binding:"required"`
	ProductTypeID *uint  `json:"product_type_id"`
	ProductID     *uint  `json:"product_id"`
	VariantID     *uint  `json:"variant_id"`
	Priority      int    `json:"priority"`
	Enabled       *bool  `json:"enabled"`
}

func (r shippingTemplateRequest) toDomain() shippingdomain.ShippingTemplate {
	enabled := true
	if r.Enabled != nil {
		enabled = *r.Enabled
	}

	template := shippingdomain.ShippingTemplate{
		Name:          strings.TrimSpace(r.Name),
		Type:          strings.TrimSpace(r.Type),
		FreeShipping:  r.FreeShipping,
		FreeThreshold: r.FreeThreshold,
		DefaultFee:    r.DefaultFee,
		Description:   strings.TrimSpace(r.Description),
		Enabled:       enabled,
	}

	for _, rule := range r.Rules {
		template.Rules = append(template.Rules, rule.toDomain())
	}

	return template
}

func (r shippingRuleRequest) toDomain() shippingdomain.ShippingRule {
	return shippingdomain.ShippingRule{
		ID:         r.ID,
		Region:     strings.ToUpper(strings.TrimSpace(r.Region)),
		MinValue:   r.MinValue,
		MaxValue:   r.MaxValue,
		Fee:        r.Fee,
		Additional: r.Additional,
	}
}

func (r shippingZoneRequest) toDomain() shippingdomain.ShippingZone {
	enabled := true
	if r.Enabled != nil {
		enabled = *r.Enabled
	}

	return shippingdomain.ShippingZone{
		Name:        strings.TrimSpace(r.Name),
		Countries:   strings.TrimSpace(r.Countries),
		States:      strings.TrimSpace(r.States),
		PostalCodes: strings.TrimSpace(r.PostalCodes),
		Enabled:     enabled,
	}
}

func (r shippingTemplateBindingRequest) toDomain() shippingdomain.ShippingTemplateBinding {
	enabled := true
	if r.Enabled != nil {
		enabled = *r.Enabled
	}

	return shippingdomain.ShippingTemplateBinding{
		TemplateID:    r.TemplateID,
		Scope:         strings.ToLower(strings.TrimSpace(r.Scope)),
		ProductTypeID: r.ProductTypeID,
		ProductID:     r.ProductID,
		VariantID:     r.VariantID,
		Priority:      r.Priority,
		Enabled:       enabled,
	}
}
