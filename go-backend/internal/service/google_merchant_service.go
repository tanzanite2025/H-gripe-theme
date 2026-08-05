package service

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"tanzanite/internal/domain/currency"
	"tanzanite/internal/domain/merchant"
	"tanzanite/internal/domain/product"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrGoogleMerchantOfferNotFound = errors.New("google merchant offer not found")
	ErrGoogleMerchantOfferInvalid  = errors.New("google merchant offer invalid")
)

type GoogleMerchantService struct {
	offers         *repository.GoogleMerchantRepository
	products       *repository.ProductRepository
	currencyPolicy *CurrencyPolicyService
	googleConfig   config.GoogleMerchantConfig
	storefrontURL  string
}

type GoogleMerchantOfferInput struct {
	ProductID             uint     `json:"product_id"`
	VariantID             uint     `json:"variant_id"`
	OfferID               string   `json:"offer_id"`
	Title                 string   `json:"title"`
	Description           string   `json:"description"`
	Brand                 string   `json:"brand"`
	Condition             string   `json:"condition"`
	GoogleProductCategory string   `json:"google_product_category"`
	GTIN                  string   `json:"gtin"`
	MPN                   string   `json:"mpn"`
	IdentifierExists      *bool    `json:"identifier_exists"`
	TargetCountry         string   `json:"target_country"`
	ContentLanguage       string   `json:"content_language"`
	CurrencyCode          string   `json:"currency_code"`
	FeedLabel             string   `json:"feed_label"`
	PriceOverride         *float64 `json:"price_override"`
	SalePriceOverride     *float64 `json:"sale_price_override"`
	PublicationStatus     string   `json:"publication_status"`
}

func NewGoogleMerchantService(
	offers *repository.GoogleMerchantRepository,
	products *repository.ProductRepository,
	currencyPolicy *CurrencyPolicyService,
	googleConfig config.GoogleMerchantConfig,
	storefrontURL string,
) *GoogleMerchantService {
	return &GoogleMerchantService{
		offers:         offers,
		products:       products,
		currencyPolicy: currencyPolicy,
		googleConfig:   googleConfig,
		storefrontURL:  strings.TrimRight(strings.TrimSpace(storefrontURL), "/"),
	}
}

func (s *GoogleMerchantService) ListOffers() ([]merchant.GoogleMerchantOffer, error) {
	return s.offers.ListOffers()
}

func (s *GoogleMerchantService) CreateOffer(input GoogleMerchantOfferInput) (*merchant.GoogleMerchantOffer, error) {
	offer, err := s.buildOffer(input)
	if err != nil {
		return nil, err
	}
	if existing, err := s.offers.FindOfferByVariantID(offer.VariantID); err == nil && existing != nil {
		return nil, fmt.Errorf("%w: this SKU has already been selected", ErrGoogleMerchantOfferInvalid)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err := s.offers.CreateOffer(offer); err != nil {
		return nil, err
	}
	return s.offers.FindOfferByID(offer.ID)
}

func (s *GoogleMerchantService) UpdateOffer(id uint, input GoogleMerchantOfferInput) (*merchant.GoogleMerchantOffer, error) {
	existing, err := s.offers.FindOfferByID(id)
	if err != nil {
		return nil, normalizeGoogleMerchantOfferError(err)
	}
	offer, err := s.buildOffer(input)
	if err != nil {
		return nil, err
	}
	if err := validateGoogleMerchantRemoteIdentityChange(existing, offer); err != nil {
		return nil, err
	}
	if other, err := s.offers.FindOfferByVariantID(offer.VariantID); err == nil && other != nil && other.ID != existing.ID {
		return nil, fmt.Errorf("%w: this SKU has already been selected", ErrGoogleMerchantOfferInvalid)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	offer.ID = existing.ID
	offer.CreatedAt = existing.CreatedAt
	offer.SyncStatus = existing.SyncStatus
	offer.LastValidatedAt = existing.LastValidatedAt
	offer.LastSyncAt = existing.LastSyncAt
	offer.LastError = existing.LastError
	if existing.SyncStatus == "synced" && !googleMerchantOfferSyncPayloadEqual(existing, offer) {
		offer.SyncStatus = "ready"
		offer.LastError = "local Google Merchant fields changed after the last sync"
	}
	if err := s.offers.UpdateOffer(offer); err != nil {
		return nil, err
	}
	return s.offers.FindOfferByID(id)
}

func (s *GoogleMerchantService) DeleteOffer(id uint) error {
	offer, err := s.offers.FindOfferByID(id)
	if err != nil {
		return normalizeGoogleMerchantOfferError(err)
	}
	if googleMerchantOfferHasRemoteSubmission(offer) {
		return fmt.Errorf("%w: remove the offer from Google before deleting local sync config", ErrGoogleMerchantOfferInvalid)
	}
	return s.offers.DeleteOffer(id)
}

func googleMerchantOfferHasRemoteSubmission(offer *merchant.GoogleMerchantOffer) bool {
	return offer != nil && offer.LastSyncAt != nil && offer.SyncStatus != "removed"
}

func validateGoogleMerchantRemoteIdentityChange(existing, next *merchant.GoogleMerchantOffer) error {
	if existing == nil || next == nil || !googleMerchantOfferHasRemoteSubmission(existing) {
		return nil
	}
	if existing.OfferID != next.OfferID || existing.ContentLanguage != next.ContentLanguage || existing.FeedLabel != next.FeedLabel {
		return fmt.Errorf("%w: synced offer identity cannot be changed; remove it from Google first", ErrGoogleMerchantOfferInvalid)
	}
	return nil
}

func googleMerchantOfferSyncPayloadEqual(existing, next *merchant.GoogleMerchantOffer) bool {
	if existing == nil || next == nil {
		return existing == next
	}
	return existing.ProductID == next.ProductID &&
		existing.VariantID == next.VariantID &&
		existing.OfferID == next.OfferID &&
		existing.Title == next.Title &&
		existing.Description == next.Description &&
		existing.Brand == next.Brand &&
		existing.Condition == next.Condition &&
		existing.GoogleProductCategory == next.GoogleProductCategory &&
		existing.GTIN == next.GTIN &&
		existing.MPN == next.MPN &&
		googleMerchantBoolPtrEqual(existing.IdentifierExists, next.IdentifierExists) &&
		existing.TargetCountry == next.TargetCountry &&
		existing.ContentLanguage == next.ContentLanguage &&
		existing.CurrencyCode == next.CurrencyCode &&
		existing.FeedLabel == next.FeedLabel &&
		googleMerchantFloatPtrEqual(existing.PriceOverride, next.PriceOverride) &&
		googleMerchantFloatPtrEqual(existing.SalePriceOverride, next.SalePriceOverride) &&
		existing.PublicationStatus == next.PublicationStatus
}

func googleMerchantBoolPtrEqual(left, right *bool) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func googleMerchantFloatPtrEqual(left, right *float64) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func (s *GoogleMerchantService) ValidateOffer(id uint) (*merchant.GoogleMerchantOffer, error) {
	offer, err := s.offers.FindOfferByID(id)
	if err != nil {
		return nil, normalizeGoogleMerchantOfferError(err)
	}
	storefrontBaseURL := s.storefrontURLForOfferValidation()
	if err := s.validateReadyOfferWithStorefrontURL(offer, storefrontBaseURL); err != nil {
		offer.SyncStatus = "validation_failed"
		offer.LastError = err.Error()
		_ = s.offers.UpdateOffer(offer)
		return nil, err
	}
	now := time.Now().UTC()
	offer.LastValidatedAt = &now
	if offer.SyncStatus == "removed" {
		offer.LastSyncAt = nil
	}
	offer.LastError = ""
	offer.SyncStatus = "ready"
	if offer.PublicationStatus == "draft" {
		offer.PublicationStatus = "ready"
	}
	if err := s.offers.UpdateOffer(offer); err != nil {
		return nil, err
	}
	return s.offers.FindOfferByID(id)
}

func (s *GoogleMerchantService) storefrontURLForOfferValidation() string {
	connection, err := s.offers.FindConnection()
	if err != nil {
		return s.effectiveStorefrontBaseURL(nil)
	}
	return s.effectiveStorefrontBaseURL(connection)
}

func (s *GoogleMerchantService) buildOffer(input GoogleMerchantOfferInput) (*merchant.GoogleMerchantOffer, error) {
	if input.ProductID == 0 || input.VariantID == 0 {
		return nil, fmt.Errorf("%w: product_id and variant_id are required", ErrGoogleMerchantOfferInvalid)
	}
	item, err := s.products.FindByID(input.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: product not found", ErrGoogleMerchantOfferInvalid)
		}
		return nil, err
	}
	variant := findGoogleMerchantVariant(item, input.VariantID)
	if variant == nil {
		return nil, fmt.Errorf("%w: selected SKU does not belong to this product", ErrGoogleMerchantOfferInvalid)
	}
	offerID := strings.TrimSpace(input.OfferID)
	if offerID == "" {
		offerID = "tz-" + strings.ToLower(strings.ReplaceAll(variant.SKU, " ", "-"))
	}
	if len(offerID) > 160 {
		return nil, fmt.Errorf("%w: offer_id must not exceed 160 characters", ErrGoogleMerchantOfferInvalid)
	}
	status := strings.ToLower(strings.TrimSpace(input.PublicationStatus))
	if status == "" {
		status = "draft"
	}
	if status != "draft" && status != "ready" && status != "paused" {
		return nil, fmt.Errorf("%w: publication_status must be draft, ready, or paused", ErrGoogleMerchantOfferInvalid)
	}
	condition := strings.ToLower(strings.TrimSpace(input.Condition))
	if condition == "" {
		condition = "new"
	}
	if condition != "new" && condition != "used" && condition != "refurbished" {
		return nil, fmt.Errorf("%w: condition must be new, used, or refurbished", ErrGoogleMerchantOfferInvalid)
	}
	gtin := strings.NewReplacer(" ", "", "-", "").Replace(strings.TrimSpace(input.GTIN))
	if gtin != "" && (!isDigitLength(gtin, 8) && !isDigitLength(gtin, 12) && !isDigitLength(gtin, 13) && !isDigitLength(gtin, 14)) {
		return nil, fmt.Errorf("%w: gtin must contain 8, 12, 13, or 14 digits", ErrGoogleMerchantOfferInvalid)
	}
	mpn := strings.TrimSpace(input.MPN)
	if len([]rune(mpn)) > 70 {
		return nil, fmt.Errorf("%w: mpn must not exceed 70 characters", ErrGoogleMerchantOfferInvalid)
	}
	if input.IdentifierExists != nil && !*input.IdentifierExists && (gtin != "" || mpn != "") {
		return nil, fmt.Errorf("%w: GTIN and MPN must be empty when identifier_exists is false", ErrGoogleMerchantOfferInvalid)
	}
	if input.PriceOverride != nil && *input.PriceOverride <= 0 {
		return nil, fmt.Errorf("%w: price_override must be greater than zero", ErrGoogleMerchantOfferInvalid)
	}
	if input.SalePriceOverride != nil && *input.SalePriceOverride <= 0 {
		return nil, fmt.Errorf("%w: sale_price_override must be greater than zero", ErrGoogleMerchantOfferInvalid)
	}

	targetCountry := strings.ToUpper(strings.TrimSpace(input.TargetCountry))
	feedLabel := strings.ToUpper(strings.TrimSpace(input.FeedLabel))
	if feedLabel == "" {
		feedLabel = targetCountry
	}

	return &merchant.GoogleMerchantOffer{
		ProductID:             item.ID,
		VariantID:             variant.ID,
		OfferID:               offerID,
		Title:                 strings.TrimSpace(input.Title),
		Description:           strings.TrimSpace(input.Description),
		Brand:                 strings.TrimSpace(input.Brand),
		Condition:             condition,
		GoogleProductCategory: strings.TrimSpace(input.GoogleProductCategory),
		GTIN:                  gtin,
		MPN:                   mpn,
		IdentifierExists:      input.IdentifierExists,
		TargetCountry:         targetCountry,
		ContentLanguage:       strings.ToLower(strings.TrimSpace(input.ContentLanguage)),
		CurrencyCode:          currency.NormalizeCode(input.CurrencyCode),
		FeedLabel:             feedLabel,
		PriceOverride:         input.PriceOverride,
		SalePriceOverride:     input.SalePriceOverride,
		PublicationStatus:     status,
		SyncStatus:            "not_synced",
	}, nil
}

func (s *GoogleMerchantService) validateReadyOffer(offer *merchant.GoogleMerchantOffer) error {
	return s.validateReadyOfferWithStorefrontURL(offer, s.effectiveStorefrontBaseURL(nil))
}

func (s *GoogleMerchantService) validateReadyOfferWithStorefrontURL(offer *merchant.GoogleMerchantOffer, storefrontBaseURL string) error {
	if offer == nil || offer.Product == nil || offer.Variant == nil {
		return fmt.Errorf("%w: offer source SKU is unavailable", ErrGoogleMerchantOfferInvalid)
	}
	if offer.Product.Status != "active" || !offer.Variant.IsActive {
		return fmt.Errorf("%w: product and SKU must both be active", ErrGoogleMerchantOfferInvalid)
	}
	if strings.TrimSpace(offer.Product.Description) == "" || !hasGoogleMerchantImage(offer.Product.Media) {
		return fmt.Errorf("%w: source product requires a description and visible image", ErrGoogleMerchantOfferInvalid)
	}
	if _, err := firstGoogleMerchantImage(storefrontBaseURL, offer.Product.Media); err != nil {
		return fmt.Errorf("%w: %v", ErrGoogleMerchantOfferInvalid, err)
	}
	if offer.Brand == "" || offer.GoogleProductCategory == "" || offer.IdentifierExists == nil {
		return fmt.Errorf("%w: brand, Google category, and identifier status are required", ErrGoogleMerchantOfferInvalid)
	}
	if *offer.IdentifierExists && offer.GTIN == "" && offer.MPN == "" {
		return fmt.Errorf("%w: GTIN or MPN is required when identifiers exist", ErrGoogleMerchantOfferInvalid)
	}
	if !isUpperAlpha(offer.TargetCountry, 2) || !isLowerAlpha(offer.ContentLanguage, 2, 2) {
		return fmt.Errorf("%w: target country and content language are required", ErrGoogleMerchantOfferInvalid)
	}
	if !isGoogleMerchantFeedLabel(offer.FeedLabel) {
		return fmt.Errorf("%w: feed label is required and may contain only A-Z, 0-9, hyphen, or underscore", ErrGoogleMerchantOfferInvalid)
	}
	if !currency.IsCatalogCode(offer.CurrencyCode) {
		return fmt.Errorf("%w: supported currency is required", ErrGoogleMerchantOfferInvalid)
	}
	if s.currencyPolicy != nil {
		if _, err := s.currencyPolicy.ValidateAcceptedCurrency(offer.CurrencyCode); err != nil {
			return fmt.Errorf("%w: currency must be accepted for order payment collection", ErrGoogleMerchantOfferInvalid)
		}
	}
	price := offer.Variant.Price
	if offer.PriceOverride != nil {
		price = *offer.PriceOverride
	}
	sale := offer.Variant.SalePrice
	if offer.SalePriceOverride != nil {
		sale = offer.SalePriceOverride
	}
	if price <= 0 || (sale != nil && (*sale <= 0 || *sale >= price)) {
		return fmt.Errorf("%w: effective sale price must be below effective price", ErrGoogleMerchantOfferInvalid)
	}
	if _, err := googleMerchantProductURL(storefrontBaseURL, offer.Product, offer.Variant); err != nil {
		return fmt.Errorf("%w: %v", ErrGoogleMerchantOfferInvalid, err)
	}
	return nil
}

func normalizeGoogleMerchantOfferError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrGoogleMerchantOfferNotFound
	}
	return err
}

func findGoogleMerchantVariant(item *product.Product, variantID uint) *product.ProductVariant {
	for index := range item.Variants {
		if item.Variants[index].ID == variantID {
			return &item.Variants[index]
		}
	}
	return nil
}

func hasGoogleMerchantImage(media []product.ProductMedia) bool {
	for _, item := range media {
		if item.IsVisible && item.MediaType == "image" && strings.TrimSpace(item.URL) != "" {
			return true
		}
	}
	return false
}

func isDigitLength(value string, length int) bool {
	if len(value) != length {
		return false
	}
	for _, character := range value {
		if !unicode.IsDigit(character) {
			return false
		}
	}
	return true
}

func isUpperAlpha(value string, length int) bool {
	if len(value) != length {
		return false
	}
	for _, character := range value {
		if !unicode.IsUpper(character) {
			return false
		}
	}
	return true
}

func isLowerAlpha(value string, min, max int) bool {
	if len(value) < min || len(value) > max {
		return false
	}
	for _, character := range value {
		if !unicode.IsLower(character) {
			return false
		}
	}
	return true
}

func isGoogleMerchantFeedLabel(value string) bool {
	if len(value) == 0 || len(value) > 20 {
		return false
	}
	for _, character := range value {
		if !((character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') ||
			character == '-' ||
			character == '_') {
			return false
		}
	}
	return true
}
