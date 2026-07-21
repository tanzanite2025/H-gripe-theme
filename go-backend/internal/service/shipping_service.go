package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"tanzanite/internal/domain/shipping"
	"tanzanite/internal/repository"
)

type ShippingService struct {
	shippingRepo *repository.ShippingRepository
	productRepo  *repository.ProductRepository
}

type ShippingCalculationInput struct {
	TemplateID uint
	Weight     float64
	Quantity   int
	Amount     float64
	Country    string
}

type ShippingQuoteItemInput struct {
	ProductID     uint    `json:"product_id"`
	VariantID     *uint   `json:"variant_id,omitempty"`
	ProductTypeID *uint   `json:"product_type_id,omitempty"`
	Quantity      int     `json:"quantity"`
	UnitPrice     float64 `json:"unit_price"`
	WeightGrams   int     `json:"weight_grams"`
}

type ShippingQuoteInput struct {
	Country  string                   `json:"country"`
	Amount   float64                  `json:"amount"`
	Currency string                   `json:"currency,omitempty"`
	Items    []ShippingQuoteItemInput `json:"items"`
}

type ShippingQuoteItem struct {
	ProductID     uint    `json:"product_id"`
	VariantID     *uint   `json:"variant_id,omitempty"`
	ProductTypeID *uint   `json:"product_type_id,omitempty"`
	TemplateID    uint    `json:"template_id"`
	TemplateName  string  `json:"template_name"`
	Quantity      int     `json:"quantity"`
	UnitPrice     float64 `json:"unit_price"`
	Amount        float64 `json:"amount"`
	WeightGrams   int     `json:"weight_grams"`
	ShippingFee   float64 `json:"shipping_fee"`
	FreeShipping  bool    `json:"free_shipping"`
}

type ShippingQuote struct {
	ShippingFee  float64             `json:"shipping_fee"`
	FreeShipping bool                `json:"free_shipping"`
	Currency     string              `json:"currency,omitempty"`
	Items        []ShippingQuoteItem `json:"items,omitempty"`
}

type resolvedShippingItem struct {
	ShippingQuoteItemInput
	Amount   float64
	Template *shipping.ShippingTemplate
}

type shippingQuoteGroup struct {
	Template         *shipping.ShippingTemplate
	ItemIndexes      []int
	Amount           float64
	Quantity         int
	TotalWeightGrams int
}

func NewShippingService(shippingRepo *repository.ShippingRepository, productRepo ...*repository.ProductRepository) *ShippingService {
	service := &ShippingService{shippingRepo: shippingRepo}
	if len(productRepo) > 0 {
		service.productRepo = productRepo[0]
	}
	return service
}

func (s *ShippingService) ListTemplates() ([]shipping.ShippingTemplate, error) {
	return s.shippingRepo.FindAllTemplates()
}

func (s *ShippingService) GetTemplate(id uint) (*shipping.ShippingTemplate, error) {
	return s.shippingRepo.FindTemplateByID(id)
}

func (s *ShippingService) CreateTemplate(template *shipping.ShippingTemplate) error {
	return s.shippingRepo.CreateTemplateWithRules(template, template.Rules)
}

func (s *ShippingService) UpdateTemplate(template *shipping.ShippingTemplate) error {
	return s.shippingRepo.UpdateTemplateWithRules(template, template.Rules)
}

func (s *ShippingService) DeleteTemplate(id uint) error {
	return s.shippingRepo.DeleteTemplate(id)
}

func (s *ShippingService) ListTemplateBindings() ([]shipping.ShippingTemplateBinding, error) {
	return s.shippingRepo.FindAllTemplateBindings()
}

func (s *ShippingService) GetTemplateBinding(id uint) (*shipping.ShippingTemplateBinding, error) {
	return s.shippingRepo.FindTemplateBindingByID(id)
}

func (s *ShippingService) CreateTemplateBinding(binding *shipping.ShippingTemplateBinding) error {
	return s.shippingRepo.CreateTemplateBinding(binding)
}

func (s *ShippingService) UpdateTemplateBinding(binding *shipping.ShippingTemplateBinding) error {
	return s.shippingRepo.UpdateTemplateBinding(binding)
}

func (s *ShippingService) DeleteTemplateBinding(id uint) error {
	return s.shippingRepo.DeleteTemplateBinding(id)
}

func (s *ShippingService) CreateTemplateRule(templateID uint, rule *shipping.ShippingRule) error {
	rule.TemplateID = templateID
	return s.shippingRepo.CreateRule(rule)
}

func (s *ShippingService) UpdateTemplateRule(templateID uint, rule *shipping.ShippingRule) error {
	rule.TemplateID = templateID
	return s.shippingRepo.UpdateRuleForTemplate(rule)
}

func (s *ShippingService) DeleteTemplateRule(templateID uint, ruleID uint) error {
	return s.shippingRepo.DeleteRuleForTemplate(templateID, ruleID)
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
	case "price", "amount":
		value = input.Amount
	}

	shippingFee := template.DefaultFee
	for _, rule := range template.Rules {
		if shippingRuleMatchesCountry(rule.Region, input.Country) && value >= rule.MinValue && (rule.MaxValue == 0 || value <= rule.MaxValue) {
			shippingFee = calculateRuleFee(rule, value)
			break
		}
	}

	return &ShippingQuote{ShippingFee: roundMoney(shippingFee), FreeShipping: false}, nil
}

func (s *ShippingService) QuoteCart(input ShippingQuoteInput) (*ShippingQuote, error) {
	if s.productRepo == nil {
		return nil, errors.New("shipping quote product repository is not configured")
	}
	if len(input.Items) == 0 {
		return nil, errors.New("shipping quote requires at least one item")
	}

	items := make([]ShippingQuoteItemInput, 0, len(input.Items))
	var amount float64
	for _, item := range input.Items {
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("invalid quantity for product ID %d", item.ProductID)
		}

		product, variant, err := s.productRepo.FindPurchasableVariant(item.ProductID, item.VariantID)
		if err != nil {
			return nil, fmt.Errorf("product ID %d is not available for shipping quote: %w", item.ProductID, err)
		}
		if variant == nil {
			return nil, fmt.Errorf("product ID %d has no purchasable SKU", item.ProductID)
		}
		if variant.Weight <= 0 {
			return nil, fmt.Errorf("shipping weight is missing for SKU %s", variant.SKU)
		}

		resolvedVariantID := variant.ID
		unitPrice := variant.EffectivePrice()
		amount += unitPrice * float64(item.Quantity)
		items = append(items, ShippingQuoteItemInput{
			ProductID:     product.ID,
			VariantID:     &resolvedVariantID,
			ProductTypeID: product.ProductTypeID,
			Quantity:      item.Quantity,
			UnitPrice:     unitPrice,
			WeightGrams:   variant.Weight,
		})
	}

	input.Items = items
	input.Amount = amount
	return s.QuoteResolvedItems(input)
}

func (s *ShippingService) QuoteResolvedItems(input ShippingQuoteInput) (*ShippingQuote, error) {
	country := strings.ToUpper(strings.TrimSpace(input.Country))
	if country == "" {
		return nil, errors.New("shipping country is required")
	}
	if len(input.Items) == 0 {
		return nil, errors.New("shipping quote requires at least one item")
	}

	bindings, err := s.shippingRepo.FindEnabledTemplateBindingsWithTemplates()
	if err != nil {
		return nil, err
	}
	if len(bindings) == 0 {
		return nil, errors.New("no enabled shipping template binding is configured")
	}

	resolvedItems := make([]resolvedShippingItem, 0, len(input.Items))
	var cartAmount float64
	for _, item := range input.Items {
		if item.ProductID == 0 {
			return nil, errors.New("shipping quote item product_id is required")
		}
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("invalid quantity for product ID %d", item.ProductID)
		}
		if item.WeightGrams <= 0 {
			if item.VariantID != nil {
				return nil, fmt.Errorf("shipping weight is missing for variant ID %d", *item.VariantID)
			}
			return nil, fmt.Errorf("shipping weight is missing for product ID %d", item.ProductID)
		}

		template := selectShippingTemplateForItem(bindings, item)
		if template == nil {
			return nil, fmt.Errorf("no enabled shipping template binding matches product ID %d", item.ProductID)
		}

		amount := item.UnitPrice * float64(item.Quantity)
		cartAmount += amount
		resolvedItems = append(resolvedItems, resolvedShippingItem{
			ShippingQuoteItemInput: item,
			Amount:                 amount,
			Template:               template,
		})
	}

	if input.Amount > 0 {
		cartAmount = input.Amount
	}

	groups := make(map[uint]*shippingQuoteGroup)
	quoteItems := make([]ShippingQuoteItem, len(resolvedItems))
	for index, item := range resolvedItems {
		templateID := item.Template.ID
		group := groups[templateID]
		if group == nil {
			group = &shippingQuoteGroup{Template: item.Template}
			groups[templateID] = group
		}

		group.ItemIndexes = append(group.ItemIndexes, index)
		group.Amount += item.Amount
		group.Quantity += item.Quantity
		group.TotalWeightGrams += item.WeightGrams * item.Quantity

		quoteItems[index] = ShippingQuoteItem{
			ProductID:     item.ProductID,
			VariantID:     item.VariantID,
			ProductTypeID: item.ProductTypeID,
			TemplateID:    item.Template.ID,
			TemplateName:  item.Template.Name,
			Quantity:      item.Quantity,
			UnitPrice:     item.UnitPrice,
			Amount:        roundMoney(item.Amount),
			WeightGrams:   item.WeightGrams,
		}
	}

	var shippingFee float64
	freeShipping := false
	for _, group := range groups {
		groupFee, groupFree := calculateTemplateShippingFee(
			group.Template,
			country,
			group.TotalWeightGrams,
			group.Quantity,
			group.Amount,
			cartAmount,
		)
		if groupFree {
			freeShipping = true
		}
		shippingFee += groupFee
		distributeGroupFee(group, resolvedItems, quoteItems, groupFee, groupFree)
	}

	shippingFee = roundMoney(shippingFee)
	if shippingFee > 0 {
		freeShipping = false
	}

	currency := strings.TrimSpace(input.Currency)
	if currency == "" {
		currency = "USD"
	}

	return &ShippingQuote{
		ShippingFee:  shippingFee,
		FreeShipping: freeShipping,
		Currency:     currency,
		Items:        quoteItems,
	}, nil
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

func (s *ShippingService) CreateZone(zone *shipping.ShippingZone) error {
	return s.shippingRepo.CreateZone(zone)
}

func (s *ShippingService) UpdateZone(zone *shipping.ShippingZone) error {
	return s.shippingRepo.UpdateZone(zone)
}

func (s *ShippingService) DeleteZone(id uint) error {
	return s.shippingRepo.DeleteZone(id)
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

func selectShippingTemplateForItem(bindings []shipping.ShippingTemplateBinding, item ShippingQuoteItemInput) *shipping.ShippingTemplate {
	var selected *shipping.ShippingTemplateBinding
	selectedScore := 0

	for i := range bindings {
		binding := &bindings[i]
		if !binding.Enabled || binding.Template == nil || !binding.Template.Enabled {
			continue
		}

		score := shippingBindingMatchScore(binding, item)
		if score == 0 {
			continue
		}

		if selected == nil ||
			score > selectedScore ||
			(score == selectedScore && binding.Priority > selected.Priority) ||
			(score == selectedScore && binding.Priority == selected.Priority && binding.ID > selected.ID) {
			selected = binding
			selectedScore = score
		}
	}

	if selected == nil {
		return nil
	}
	return selected.Template
}

func shippingBindingMatchScore(binding *shipping.ShippingTemplateBinding, item ShippingQuoteItemInput) int {
	switch binding.Scope {
	case "variant":
		if item.VariantID != nil && binding.VariantID != nil && *item.VariantID == *binding.VariantID {
			return 4
		}
	case "product":
		if binding.ProductID != nil && item.ProductID == *binding.ProductID {
			return 3
		}
	case "product_type":
		if item.ProductTypeID != nil && binding.ProductTypeID != nil && *item.ProductTypeID == *binding.ProductTypeID {
			return 2
		}
	case "default":
		return 1
	}
	return 0
}

func calculateTemplateShippingFee(
	template *shipping.ShippingTemplate,
	country string,
	totalWeightGrams int,
	quantity int,
	amount float64,
	cartAmount float64,
) (float64, bool) {
	if template == nil {
		return 0, false
	}

	if template.FreeShipping && cartAmount >= template.FreeThreshold {
		return 0, true
	}

	value := float64(totalWeightGrams) / 1000
	switch template.Type {
	case "quantity", "items":
		value = float64(quantity)
	case "price", "amount":
		value = amount
	}

	shippingFee := template.DefaultFee
	for _, rule := range template.Rules {
		if shippingRuleMatchesCountry(rule.Region, country) && shippingRuleMatchesValue(rule, value) {
			shippingFee = calculateRuleFee(rule, value)
			break
		}
	}

	return roundMoney(shippingFee), false
}

func shippingRuleMatchesValue(rule shipping.ShippingRule, value float64) bool {
	return value >= rule.MinValue && (rule.MaxValue == 0 || value <= rule.MaxValue)
}

func calculateRuleFee(rule shipping.ShippingRule, value float64) float64 {
	fee := rule.Fee
	if rule.Additional > 0 && value > rule.MinValue {
		fee += math.Ceil(value-rule.MinValue) * rule.Additional
	}
	return fee
}

func shippingRuleMatchesCountry(region string, country string) bool {
	normalizedCountry := strings.ToUpper(strings.TrimSpace(country))
	if normalizedCountry == "" {
		return false
	}

	regions := normalizeShippingRegions(region)
	if len(regions) == 0 {
		return true
	}

	for _, candidate := range regions {
		switch candidate {
		case "*", "ALL", "GLOBAL", "WORLDWIDE":
			return true
		default:
			if candidate == normalizedCountry {
				return true
			}
		}
	}
	return false
}

func normalizeShippingRegions(region string) []string {
	trimmed := strings.TrimSpace(region)
	if trimmed == "" {
		return nil
	}

	var parsed []string
	if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
		return normalizeRegionList(parsed)
	}

	parts := strings.FieldsFunc(trimmed, func(r rune) bool {
		return r == ',' || r == ';' || r == '|' || r == '\n' || r == '\r' || r == '\t'
	})
	return normalizeRegionList(parts)
}

func normalizeRegionList(regions []string) []string {
	normalized := make([]string, 0, len(regions))
	for _, region := range regions {
		candidate := strings.ToUpper(strings.TrimSpace(region))
		if candidate != "" {
			normalized = append(normalized, candidate)
		}
	}
	return normalized
}

func distributeGroupFee(
	group *shippingQuoteGroup,
	resolvedItems []resolvedShippingItem,
	quoteItems []ShippingQuoteItem,
	groupFee float64,
	groupFree bool,
) {
	if group == nil || len(group.ItemIndexes) == 0 {
		return
	}

	var totalBasis float64
	for _, index := range group.ItemIndexes {
		totalBasis += shippingFeeDistributionBasis(group.Template.Type, resolvedItems[index])
	}

	remaining := roundMoney(groupFee)
	for position, index := range group.ItemIndexes {
		itemFee := 0.0
		if position == len(group.ItemIndexes)-1 {
			itemFee = remaining
		} else if totalBasis > 0 {
			itemFee = roundMoney(groupFee * shippingFeeDistributionBasis(group.Template.Type, resolvedItems[index]) / totalBasis)
			remaining = roundMoney(remaining - itemFee)
		} else {
			itemFee = roundMoney(groupFee / float64(len(group.ItemIndexes)))
			remaining = roundMoney(remaining - itemFee)
		}

		quoteItems[index].ShippingFee = itemFee
		quoteItems[index].FreeShipping = groupFree
	}
}

func shippingFeeDistributionBasis(templateType string, item resolvedShippingItem) float64 {
	switch templateType {
	case "quantity", "items":
		return float64(item.Quantity)
	case "price", "amount":
		return item.Amount
	default:
		return float64(item.WeightGrams * item.Quantity)
	}
}

func roundMoney(value float64) float64 {
	return math.Round(value*100) / 100
}
