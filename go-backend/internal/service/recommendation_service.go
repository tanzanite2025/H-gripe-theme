package service

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
)

const (
	RecommendationAlgorithmVersion = "contextual-v1"
	DefaultRecommendationLimit     = 6
	MaxRecommendationLimit         = 12

	recommendationCandidateMultiplier = 8
	recommendationSignalWindowDays    = 30
	recommendationPersonalWindowDays  = 14
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
	eventRepo      *repository.RecommendationEventRepository
}

type recommendationCandidateSeed struct {
	Product product.Product
	Slot    string
	Reason  string
}

type scoredRecommendationCandidate struct {
	Product           product.Product
	Score             float64
	Slot              string
	Reason            string
	SpecMatches       int
	QueryMatches      int
	GlobalSignalScore float64
	PersonalScore     float64
}

func NewRecommendationService(productService *ProductService, eventRepos ...*repository.RecommendationEventRepository) *RecommendationService {
	service := &RecommendationService{productService: productService}
	if len(eventRepos) > 0 {
		service.eventRepo = eventRepos[0]
	}
	return service
}

func (s *RecommendationService) Recommend(input RecommendationRequest) (RecommendationResult, error) {
	now := time.Now().UTC()
	result := RecommendationResult{
		RequestID:        "rec_" + uuid.NewString(),
		AlgorithmVersion: RecommendationAlgorithmVersion,
		ExpiresAt:        now.Add(5 * time.Minute),
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

	excluded := normalizeRecommendationExclusions(input)
	contextProduct, err := s.loadRecommendationContextProduct(input.ProductID)
	if err != nil {
		return result, err
	}

	categoryID := recommendationCategoryID(input.CategoryID, contextProduct)
	query := normalizeRecommendationQuery(input.Query)
	candidateLimit := recommendationCandidateLimit(limit, len(excluded))

	candidates, err := s.collectRecommendationCandidates(surface, categoryID, query, excluded, candidateLimit)
	if err != nil {
		return result, err
	}
	if len(candidates) == 0 {
		return result, nil
	}

	candidateIDs := make([]uint, 0, len(candidates))
	for _, candidate := range candidates {
		candidateIDs = append(candidateIDs, candidate.Product.ID)
	}

	globalSignals := s.loadProductSignals(repository.RecommendationSignalQuery{
		ProductIDs: candidateIDs,
		Since:      now.AddDate(0, 0, -recommendationSignalWindowDays),
	})
	personalSignals := s.loadProductSignals(repository.RecommendationSignalQuery{
		ProductIDs:  candidateIDs,
		AnonymousID: input.AnonymousID,
		SessionID:   input.SessionID,
		Since:       now.AddDate(0, 0, -recommendationPersonalWindowDays),
	})

	scored := scoreRecommendationCandidates(recommendationScoreInput{
		Surface:         surface,
		Query:           query,
		ContextProduct:  contextProduct,
		CategoryID:      categoryID,
		Candidates:      candidates,
		GlobalSignals:   globalSignals,
		PersonalSignals: personalSignals,
		Now:             now,
	})

	result.Items = make([]RecommendationProduct, 0, minInt(limit, len(scored)))
	for _, candidate := range scored {
		result.Items = append(result.Items, makeRecommendationProduct(candidate.Product, candidate.Slot, candidate.Reason))
		if len(result.Items) >= limit {
			break
		}
	}

	return result, nil
}

func (s *RecommendationService) loadRecommendationContextProduct(productID *uint) (*product.Product, error) {
	if productID == nil || *productID == 0 {
		return nil, nil
	}

	item, err := s.productService.GetRecommendationContextProduct(*productID)
	if errors.Is(err, ErrProductNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *RecommendationService) collectRecommendationCandidates(surface string, categoryID *uint, query string, excluded map[uint]struct{}, limit int) ([]recommendationCandidateSeed, error) {
	seeds := make([]recommendationCandidateSeed, 0, limit)
	seen := make(map[uint]struct{}, len(excluded)+limit)
	excludeIDs := recommendationExcludeIDs(excluded)

	for productID := range excluded {
		seen[productID] = struct{}{}
	}

	addLayer := func(productSpecificationTemplateID *uint, keyword, slot, reason string) error {
		if len(seeds) >= limit {
			return nil
		}
		products, _, err := s.productService.ListRecommendationCandidates(ProductRecommendationCandidateInput{
			ProductSpecificationTemplateID: productSpecificationTemplateID,
			Keyword:                        keyword,
			ExcludeProductIDs:              excludeIDs,
			Page:                           1,
			PageSize:                       limit,
		})
		if err != nil {
			return err
		}
		for _, item := range products {
			if len(seeds) >= limit {
				break
			}
			if _, ok := seen[item.ID]; ok {
				continue
			}
			seen[item.ID] = struct{}{}
			seeds = append(seeds, recommendationCandidateSeed{
				Product: item,
				Slot:    slot,
				Reason:  reason,
			})
		}
		return nil
	}

	if categoryID != nil && query != "" && recommendationSurfaceUsesSearch(surface) {
		if err := addLayer(categoryID, query, "category_query", "category_query_match"); err != nil {
			return nil, err
		}
	}
	if categoryID != nil {
		if err := addLayer(categoryID, "", "similar_products", "same_product_specification_template"); err != nil {
			return nil, err
		}
	}
	if query != "" {
		if err := addLayer(nil, query, "query_related", "query_match"); err != nil {
			return nil, err
		}
	}
	if err := addLayer(nil, "", "trending_available", "popular_available"); err != nil {
		return nil, err
	}

	return seeds, nil
}

func (s *RecommendationService) loadProductSignals(input repository.RecommendationSignalQuery) map[uint]repository.RecommendationProductSignal {
	signals := map[uint]repository.RecommendationProductSignal{}
	if s == nil || s.eventRepo == nil {
		return signals
	}

	loaded, err := s.eventRepo.ListProductSignals(input)
	if err != nil {
		return signals
	}
	return loaded
}

type recommendationScoreInput struct {
	Surface         string
	Query           string
	ContextProduct  *product.Product
	CategoryID      *uint
	Candidates      []recommendationCandidateSeed
	GlobalSignals   map[uint]repository.RecommendationProductSignal
	PersonalSignals map[uint]repository.RecommendationProductSignal
	Now             time.Time
}

func scoreRecommendationCandidates(input recommendationScoreInput) []scoredRecommendationCandidate {
	tokens := recommendationQueryTokens(input.Query)
	scored := make([]scoredRecommendationCandidate, 0, len(input.Candidates))
	for _, seed := range input.Candidates {
		item := seed.Product
		globalSignal := input.GlobalSignals[item.ID]
		personalSignal := input.PersonalSignals[item.ID]

		candidate := scoredRecommendationCandidate{
			Product: item,
			Score:   baseRecommendationScore(item, input.Now),
			Slot:    seed.Slot,
			Reason:  seed.Reason,
		}

		if input.CategoryID != nil && item.ProductSpecificationTemplateID != nil && *item.ProductSpecificationTemplateID == *input.CategoryID {
			candidate.Score += 90
			if candidate.Reason == "popular_available" {
				candidate.Slot = "category_followup"
				candidate.Reason = "same_product_specification_template"
			}
		}

		if input.ContextProduct != nil {
			candidate.SpecMatches = countRecommendationSpecMatches(*input.ContextProduct, item)
			if candidate.SpecMatches > 0 {
				candidate.Score += float64(candidate.SpecMatches) * 45
				candidate.Slot = "similar_products"
				candidate.Reason = "matching_specs"
			}
		}

		candidate.QueryMatches = countRecommendationQueryMatches(tokens, item)
		if candidate.QueryMatches > 0 {
			candidate.Score += float64(candidate.QueryMatches) * 35
			if candidate.Reason == "popular_available" {
				candidate.Slot = "query_related"
				candidate.Reason = "query_match"
			}
		}

		candidate.GlobalSignalScore = recommendationSignalScore(globalSignal, false)
		candidate.Score += candidate.GlobalSignalScore
		if candidate.GlobalSignalScore >= 24 && candidate.Reason == "popular_available" {
			candidate.Slot = "trending_available"
			candidate.Reason = "behavior_trending"
		}

		candidate.PersonalScore = recommendationSignalScore(personalSignal, true)
		if candidate.PersonalScore > 0 {
			candidate.Score += 160 + candidate.PersonalScore
			candidate.Slot = "personalized"
			candidate.Reason = "your_recent_activity"
		}

		scored = append(scored, candidate)
	}

	sort.SliceStable(scored, func(i, j int) bool {
		left := scored[i]
		right := scored[j]
		if left.Score != right.Score {
			return left.Score > right.Score
		}
		if left.Product.Featured != right.Product.Featured {
			return left.Product.Featured
		}
		if left.Product.ViewCount != right.Product.ViewCount {
			return left.Product.ViewCount > right.Product.ViewCount
		}
		if !left.Product.CreatedAt.Equal(right.Product.CreatedAt) {
			return left.Product.CreatedAt.After(right.Product.CreatedAt)
		}
		return left.Product.ID > right.Product.ID
	})

	return scored
}

func baseRecommendationScore(item product.Product, now time.Time) float64 {
	score := 10.0
	if item.Featured {
		score += 30
	}
	if item.ViewCount > 0 {
		score += minFloat(float64(item.ViewCount)/4, 60)
	}
	if !item.CreatedAt.IsZero() {
		ageDays := now.Sub(item.CreatedAt).Hours() / 24
		if ageDays >= 0 && ageDays <= 45 {
			score += 20 - (ageDays / 3)
		}
	}
	if primaryRecommendationImage(item) != "" {
		score += 8
	}
	return score
}

func recommendationSignalScore(signal repository.RecommendationProductSignal, personal bool) float64 {
	multiplier := 1.0
	if personal {
		multiplier = 2.5
	}

	return multiplier * (float64(signal.ProductViews)*3 +
		float64(signal.ProductDwells)*4 +
		float64(signal.RecommendationClicks)*14 +
		float64(signal.CartAdds)*24 +
		float64(signal.WishlistAdds)*18 +
		float64(signal.CheckoutStarts)*32 +
		float64(signal.Purchases)*48)
}

func countRecommendationSpecMatches(contextProduct product.Product, candidate product.Product) int {
	contextValues := recommendationComparableSpecValues(contextProduct)
	if len(contextValues) == 0 {
		return 0
	}

	matches := 0
	for _, value := range candidate.SpecValues {
		if value.SpecDefinition == nil {
			continue
		}
		if !value.SpecDefinition.IsFilterable && !value.SpecDefinition.IsVariantOption {
			continue
		}
		expected, ok := contextValues[value.SpecDefinitionID]
		if !ok {
			continue
		}
		if normalizeRecommendationComparableValue(value.Value) == expected {
			matches++
		}
	}
	return matches
}

func recommendationComparableSpecValues(item product.Product) map[uint]string {
	values := map[uint]string{}
	for _, value := range item.SpecValues {
		if value.SpecDefinition == nil {
			continue
		}
		if !value.SpecDefinition.IsFilterable && !value.SpecDefinition.IsVariantOption {
			continue
		}
		normalized := normalizeRecommendationComparableValue(value.Value)
		if normalized != "" {
			values[value.SpecDefinitionID] = normalized
		}
	}
	return values
}

func countRecommendationQueryMatches(tokens []string, item product.Product) int {
	if len(tokens) == 0 {
		return 0
	}

	text := strings.ToLower(strings.Join([]string{
		item.Name,
		item.SKU,
		item.ShortDesc,
		item.Description,
		recommendationProductSpecificationTemplateText(item),
		recommendationSpecText(item),
	}, " "))

	matches := 0
	for _, token := range tokens {
		if token != "" && strings.Contains(text, token) {
			matches++
		}
	}
	return matches
}

func makeRecommendationProduct(item product.Product, slot string, reason string) RecommendationProduct {
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
		Slot:       normalizeRecommendationLabel(slot, "trending_available"),
		Reason:     normalizeRecommendationLabel(reason, "popular_available"),
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

func normalizeRecommendationExclusions(input RecommendationRequest) map[uint]struct{} {
	excluded := make(map[uint]struct{}, len(input.ExcludeProductIDs)+1)
	for _, productID := range input.ExcludeProductIDs {
		if productID > 0 {
			excluded[productID] = struct{}{}
		}
	}
	if input.ProductID != nil && *input.ProductID > 0 {
		excluded[*input.ProductID] = struct{}{}
	}
	return excluded
}

func recommendationExcludeIDs(excluded map[uint]struct{}) []uint {
	if len(excluded) == 0 {
		return nil
	}
	ids := make([]uint, 0, len(excluded))
	for productID := range excluded {
		ids = append(ids, productID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func recommendationCategoryID(inputCategoryID *uint, contextProduct *product.Product) *uint {
	if inputCategoryID != nil && *inputCategoryID > 0 {
		return inputCategoryID
	}
	if contextProduct != nil && contextProduct.ProductSpecificationTemplateID != nil && *contextProduct.ProductSpecificationTemplateID > 0 {
		return contextProduct.ProductSpecificationTemplateID
	}
	return nil
}

func recommendationCandidateLimit(limit int, exclusions int) int {
	candidateLimit := limit*recommendationCandidateMultiplier + exclusions
	if candidateLimit < limit {
		return limit
	}
	if candidateLimit > 96 {
		return 96
	}
	return candidateLimit
}

func normalizeRecommendationQuery(query string) string {
	query = strings.TrimSpace(query)
	return truncateRecommendationRunes(query, 160)
}

func recommendationQueryTokens(query string) []string {
	query = strings.ToLower(normalizeRecommendationComparableValue(query))
	parts := strings.Fields(query)
	tokens := make([]string, 0, minInt(len(parts), 8))
	seen := map[string]struct{}{}
	for _, part := range parts {
		if len(part) < 2 {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		tokens = append(tokens, part)
		if len(tokens) >= 8 {
			break
		}
	}
	return tokens
}

func normalizeRecommendationComparableValue(value string) string {
	replacer := strings.NewReplacer(
		",", " ",
		".", " ",
		";", " ",
		":", " ",
		"/", " ",
		"\\", " ",
		"|", " ",
		"(", " ",
		")", " ",
		"[", " ",
		"]", " ",
		"{", " ",
		"}", " ",
		"_", " ",
		"-", " ",
	)
	return strings.Join(strings.Fields(strings.ToLower(replacer.Replace(strings.TrimSpace(value)))), " ")
}

func recommendationProductSpecificationTemplateText(item product.Product) string {
	if item.ProductSpecificationTemplate == nil {
		return ""
	}
	parts := []string{item.ProductSpecificationTemplate.Name, item.ProductSpecificationTemplate.Slug, item.ProductSpecificationTemplate.Description}
	for _, translation := range item.ProductSpecificationTemplate.Translations {
		parts = append(parts, translation.Name, translation.Description)
	}
	return strings.Join(parts, " ")
}

func recommendationSpecText(item product.Product) string {
	parts := make([]string, 0, len(item.SpecValues)*2)
	for _, value := range item.SpecValues {
		parts = append(parts, value.Value)
		if value.SpecDefinition != nil {
			parts = append(parts, value.SpecDefinition.Name, value.SpecDefinition.Slug)
		}
	}
	return strings.Join(parts, " ")
}

func recommendationSurfaceUsesSearch(surface string) bool {
	surface = strings.ToLower(strings.TrimSpace(surface))
	return strings.Contains(surface, "shop") || strings.Contains(surface, "search") || strings.Contains(surface, "catalog")
}

func normalizeRecommendationLabel(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return truncateRecommendationRunes(value, 64)
}

func truncateRecommendationRunes(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func minFloat(left, right float64) float64 {
	if left < right {
		return left
	}
	return right
}

func IsRecommendationValidationError(err error) bool {
	return errors.Is(err, ErrRecommendationSurfaceInvalid) ||
		errors.Is(err, ErrRecommendationLimitInvalid)
}
