package service

import (
	"tanzanite/internal/domain/shipping"
	"tanzanite/internal/repository"
)

type ShippingService struct {
	shippingRepo *repository.ShippingRepository
}

type ShippingCalculationInput struct {
	TemplateID uint
	Weight     float64
	Quantity   int
	Amount     float64
	Country    string
}

type ShippingQuote struct {
	ShippingFee  float64 `json:"shipping_fee"`
	FreeShipping bool    `json:"free_shipping"`
}

func NewShippingService(shippingRepo *repository.ShippingRepository) *ShippingService {
	return &ShippingService{shippingRepo: shippingRepo}
}

func (s *ShippingService) ListTemplates() ([]shipping.ShippingTemplate, error) {
	return s.shippingRepo.FindAllTemplates()
}

func (s *ShippingService) GetTemplate(id uint) (*shipping.ShippingTemplate, error) {
	return s.shippingRepo.FindTemplateByID(id)
}

func (s *ShippingService) CalculateShipping(input ShippingCalculationInput) (*ShippingQuote, error) {
	template, err := s.GetTemplate(input.TemplateID)
	if err != nil {
		return nil, err
	}

	if template.FreeShipping && input.Amount >= template.FreeThreshold {
		return &ShippingQuote{ShippingFee: 0, FreeShipping: true}, nil
	}

	value := input.Weight
	switch template.Type {
	case "quantity":
		value = float64(input.Quantity)
	case "price":
		value = input.Amount
	}

	shippingFee := template.DefaultFee
	for _, rule := range template.Rules {
		if value >= rule.MinValue && (rule.MaxValue == 0 || value <= rule.MaxValue) {
			shippingFee = rule.Fee
			break
		}
	}

	return &ShippingQuote{ShippingFee: shippingFee, FreeShipping: false}, nil
}

func (s *ShippingService) ListCarriers(enabledOnly bool) ([]shipping.Carrier, error) {
	return s.shippingRepo.FindAllCarriers(enabledOnly)
}

func (s *ShippingService) GetCarrier(id uint) (*shipping.Carrier, error) {
	return s.shippingRepo.FindCarrierByID(id)
}

func (s *ShippingService) CreateCarrier(carrier *shipping.Carrier) error {
	return s.shippingRepo.CreateCarrier(carrier)
}

func (s *ShippingService) UpdateCarrier(carrier *shipping.Carrier) error {
	return s.shippingRepo.UpdateCarrier(carrier)
}

func (s *ShippingService) DeleteCarrier(id uint) error {
	return s.shippingRepo.DeleteCarrier(id)
}

func (s *ShippingService) ListZones() ([]shipping.ShippingZone, error) {
	return s.shippingRepo.FindAllZones()
}

func (s *ShippingService) GetZone(id uint) (*shipping.ShippingZone, error) {
	return s.shippingRepo.FindZoneByID(id)
}

func (s *ShippingService) GetTrackingEventsByTrackingNumber(trackingNumber string) ([]shipping.TrackingEvent, error) {
	return s.shippingRepo.FindTrackingEventsByTrackingNumber(trackingNumber)
}

func (s *ShippingService) GetTrackingEventsByOrderID(orderID uint) ([]shipping.TrackingEvent, error) {
	return s.shippingRepo.FindTrackingEventsByOrderID(orderID)
}

func (s *ShippingService) ListPackagingRules() ([]shipping.PackagingRule, error) {
	return s.shippingRepo.FindAllPackagingRules()
}

func (s *ShippingService) GetPackagingRule(id uint) (*shipping.PackagingRule, error) {
	return s.shippingRepo.FindPackagingRuleByID(id)
}

func (s *ShippingService) CreatePackagingRule(rule *shipping.PackagingRule) error {
	return s.shippingRepo.CreatePackagingRule(rule)
}

func (s *ShippingService) UpdatePackagingRule(rule *shipping.PackagingRule) error {
	return s.shippingRepo.UpdatePackagingRule(rule)
}

func (s *ShippingService) DeletePackagingRule(id uint) error {
	return s.shippingRepo.DeletePackagingRule(id)
}

func (s *ShippingService) CreatePackagingRuleApply(apply *shipping.PackagingRuleApply) error {
	return s.shippingRepo.CreatePackagingRuleApply(apply)
}

func (s *ShippingService) DeletePackagingRuleApply(id uint) error {
	return s.shippingRepo.DeletePackagingRuleApply(id)
}

func (s *ShippingService) GetProductPackagingRules(productID uint) ([]shipping.PackagingRule, error) {
	return s.shippingRepo.FindPackagingRulesByProductID(productID)
}
