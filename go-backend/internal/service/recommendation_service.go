package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"tanzanite/internal/domain/product"

	"github.com/google/uuid"
)

const (
	RecommendationAlgorithmVersion = "rules-v1"
	DefaultRecommendationLimit     = 6
	MaxRecommendationLimit         = 12
)

var (
	ErrRecommendationSurfaceInvalid = errors.New("recommendation surface is invalid")
	ErrRecommendationLimitInvalid   = errors.New("recommendation limit is invalid")
)

type RecommendationRequest struct {
	Surface           string
	Locale            string
	AnonymousID       string
	SessionID         string
	ProductID         *uint
	CategoryID        *uint
	Query             string
	Route             string
	Limit             int
	ExcludeProductIDs []uint
}

type RecommendationProduct struct {
	ProductID  uint   `json:"product_id"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	Thumbnail  string `json:"thumbnail,omitempty"`
	PriceLabel string `json:"price_label,omitempty"`
	Slot       string `json:"slot"`
	Reason     string `json:"reason"`
}

type RecommendationResult struct {
	RequestID        string                  `json:"request_id"`
	AlgorithmVersion string                  `json:"algorithm_version"`
	ExpiresAt        time.Time               `json:"expires_at"`
	Items            []RecommendationProduct `json:"items"`
}

type RecommendationService struct {
	productService *ProductService
}

func NewRecommendationService(productService *ProductService) *RecommendationService {
	return &RecommendationService{productService: productService}
}

func (s *RecommendationService) Recommend(input RecommendationRequest) (RecommendationResult, error) {
	result := RecommendationResult{
		RequestID:        "rec_" + uuid.NewString(),
		AlgorithmVersion: RecommendationAlgorithmVersion,
		ExpiresAt:        time.Now().UTC().Add(5 * time.Minute),
		Items:            []RecommendationProduct{},
	}

	if s == nil || s.productService == nil {
		return result, errors.New("recommendation product service is not configured")
	}

	surface := strings.TrimSpace(input.Surface)
	if surface == "" || len(surface) > 64 {
		return result, ErrRecommendationSurfaceInvalid
	}

	limit := input.Limit
	if limit == 0 {
		limit = DefaultRecommendationLimit
	}
	if limit < 1 || limit > MaxRecommendationLimit {
		return result, ErrRecommendationLimitInvalid
	}

	excluded := make(map[uint]struct{}, len(input.ExcludeProductIDs)+1)
	for _, productID := range input.ExcludeProductIDs {
		if productID > 0 {
			excluded[productID] = struct{}{}
		}
	}
	if input.ProductID != nil && *input.ProductID > 0 {
		excluded[*input.ProductID] = struct{}{}
	}

	fetchLimit := limit + len(excluded)
	if fetchLimit < limit {
		fetchLimit = limit
	}
	if fetchLimit > MaxRecommendationLimit+50 {
		fetchLimit = MaxRecommendationLimit + 50
	}

	products, _, err := s.productService.ListPublicAvailable(
		normalizeRecommendationLocale(input.Locale),
		1,
		fetchLimit,
	)
	if err != nil {
		return result, err
	}

	result.Items = make([]RecommendationProduct, 0, minInt(limit, len(products)))
	for _, item := range products {
		if _, skip := excluded[item.ID]; skip {
			continue
		}
		result.Items = append(result.Items, makeRecommendationProduct(item))
		if len(result.Items) >= limit {
			break
		}
	}

	return result, nil
}

func makeRecommendationProduct(item product.Product) RecommendationProduct {
	price, sale := item.DisplayPrices()
	if sale != nil {
		price = *sale
	}

	return RecommendationProduct{
		ProductID:  item.ID,
		Title:      strings.TrimSpace(item.Name),
		URL:        "/shop/" + strings.TrimSpace(item.Slug),
		Thumbnail:  primaryRecommendationImage(item),
		PriceLabel: formatRecommendationPrice(price),
		Slot:       "rule_fallback",
		Reason:     "available_global",
	}
}

func primaryRecommendationImage(item product.Product) string {
	for _, media := range item.Media {
		if media.MediaType == "image" && media.IsVisible && media.IsPrimary && strings.TrimSpace(media.URL) != "" {
			return strings.TrimSpace(media.URL)
		}
	}
	for _, media := range item.Media {
		if media.MediaType == "image" && media.IsVisible && strings.TrimSpace(media.URL) != "" {
			return strings.TrimSpace(media.URL)
		}
	}
	return ""
}

func formatRecommendationPrice(price float64) string {
	if price <= 0 {
		return ""
	}
	return fmt.Sprintf("$%s", strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", price), "0"), "."))
}

func normalizeRecommendationLocale(locale string) string {
	locale = strings.TrimSpace(strings.Split(locale, ",")[0])
	locale = strings.ReplaceAll(locale, "_", "-")
	if index := strings.Index(locale, "-"); index > 0 {
		locale = locale[:index]
	}
	return strings.ToLower(locale)
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func IsRecommendationValidationError(err error) bool {
	return errors.Is(err, ErrRecommendationSurfaceInvalid) ||
		errors.Is(err, ErrRecommendationLimitInvalid)
}
