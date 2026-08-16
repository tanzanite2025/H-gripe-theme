package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

const (
	CustomsLookupProviderUSHTS         = "us_hts"
	CustomsLookupProviderUKTradeTariff = "uk_trade_tariff"
	customsLookupDefaultLimit          = 10
	customsLookupMaxLimit              = 25
	customsLookupSettingsGroup         = "api"
	customsLookupSettingsLocale        = "en"
	customsLookupDefaultUSHTSBaseURL   = "https://hts.usitc.gov/reststop/search"
	customsLookupDefaultUKTradeBaseURL = "https://www.trade-tariff.service.gov.uk/api/v2/commodities"
	customsLookupDefaultAPIKeyHeader   = "X-API-Key"
)

var (
	ErrCustomsClassificationNotFound   = errors.New("customs classification profile not found")
	ErrCustomsClassificationInvalid    = errors.New("customs classification profile invalid")
	ErrCustomsClassificationSlugExists = errors.New("customs classification profile slug already exists")
	ErrCustomsLookupInvalid            = errors.New("customs classification lookup invalid")
	ErrCustomsLookupUnavailable        = errors.New("customs classification lookup unavailable")
)

type CustomsClassificationInput struct {
	ProductSpecificationTemplateID *uint
	Name                           string
	Slug                           string
	ComponentKind                  string
	Material                       string
	HSCode                         string
	CNCode                         string
	CountryOfOrigin                string
	CustomsDescription             string
	Source                         string
	SourceCode                     string
	SourceURL                      string
	Notes                          string
	Status                         string
}

type CustomsClassificationListInput struct {
	ProductSpecificationTemplateID uint
	ComponentKind                  string
	Material                       string
	Status                         string
	Search                         string
	IncludePaused                  bool
}

type CustomsClassificationLookupInput struct {
	Provider string
	Query    string
	Limit    int
}

type CustomsClassificationLookupCandidate struct {
	Provider           string `json:"provider"`
	SourceCode         string `json:"source_code"`
	HSCode             string `json:"hs_code"`
	CNCode             string `json:"cn_code,omitempty"`
	Description        string `json:"description"`
	CustomsDescription string `json:"customs_description"`
	Duty               string `json:"duty,omitempty"`
	SourceURL          string `json:"source_url"`
}

type customsLookupHTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type customsLookupProviderConfig struct {
	Enabled      bool
	Endpoint     string
	APIKey       string
	APIKeyHeader string
}

type CustomsClassificationService struct {
	repo           *repository.CustomsClassificationRepository
	settings       *SettingService
	httpClient     customsLookupHTTPClient
	usHTSBaseURL   string
	ukTradeBaseURL string
}

func NewCustomsClassificationService(repo *repository.CustomsClassificationRepository, settings ...*SettingService) *CustomsClassificationService {
	var settingService *SettingService
	if len(settings) > 0 {
		settingService = settings[0]
	}
	return &CustomsClassificationService{
		repo:           repo,
		settings:       settingService,
		httpClient:     &http.Client{Timeout: 8 * time.Second},
		usHTSBaseURL:   customsLookupDefaultUSHTSBaseURL,
		ukTradeBaseURL: customsLookupDefaultUKTradeBaseURL,
	}
}

func (s *CustomsClassificationService) ConfigureSettings(settings *SettingService) {
	if s == nil {
		return
	}
	s.settings = settings
}

func (s *CustomsClassificationService) ConfigureLookupHTTPClient(client customsLookupHTTPClient) {
	if s == nil || client == nil {
		return
	}
	s.httpClient = client
}

func (s *CustomsClassificationService) ConfigureLookupBaseURLs(usHTSBaseURL, ukTradeBaseURL string) {
	if s == nil {
		return
	}
	if strings.TrimSpace(usHTSBaseURL) != "" {
		s.usHTSBaseURL = strings.TrimRight(strings.TrimSpace(usHTSBaseURL), "/")
	}
	if strings.TrimSpace(ukTradeBaseURL) != "" {
		s.ukTradeBaseURL = strings.TrimRight(strings.TrimSpace(ukTradeBaseURL), "/")
	}
}

func (s *CustomsClassificationService) List(input CustomsClassificationListInput) ([]product.CustomsClassificationProfile, error) {
	return s.repo.List(repository.CustomsClassificationListFilter{
		ProductSpecificationTemplateID: input.ProductSpecificationTemplateID,
		ComponentKind:                  strings.TrimSpace(input.ComponentKind),
		Material:                       strings.TrimSpace(input.Material),
		Status:                         strings.TrimSpace(input.Status),
		Search:                         strings.TrimSpace(input.Search),
		IncludePaused:                  input.IncludePaused,
	})
}

func (s *CustomsClassificationService) Get(id uint) (*product.CustomsClassificationProfile, error) {
	profile, err := s.repo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCustomsClassificationNotFound
	}
	return profile, err
}

func (s *CustomsClassificationService) Create(input CustomsClassificationInput) (*product.CustomsClassificationProfile, error) {
	profile, err := normalizeCustomsClassificationInput(input)
	if err != nil {
		return nil, err
	}
	exists, err := s.repo.SlugExists(profile.Slug, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrCustomsClassificationSlugExists
	}
	if err := s.repo.Create(profile); err != nil {
		return nil, err
	}
	return s.Get(profile.ID)
}

func (s *CustomsClassificationService) Update(id uint, input CustomsClassificationInput) (*product.CustomsClassificationProfile, error) {
	if _, err := s.Get(id); err != nil {
		return nil, err
	}
	profile, err := normalizeCustomsClassificationInput(input)
	if err != nil {
		return nil, err
	}
	profile.ID = id
	exists, err := s.repo.SlugExists(profile.Slug, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrCustomsClassificationSlugExists
	}
	if err := s.repo.Update(profile); err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *CustomsClassificationService) Delete(id uint) error {
	if _, err := s.Get(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}

func (s *CustomsClassificationService) Lookup(input CustomsClassificationLookupInput) ([]CustomsClassificationLookupCandidate, error) {
	provider := strings.ToLower(strings.TrimSpace(input.Provider))
	query := strings.TrimSpace(input.Query)
	limit := input.Limit
	if limit <= 0 {
		limit = customsLookupDefaultLimit
	}
	if limit > customsLookupMaxLimit {
		limit = customsLookupMaxLimit
	}
	if query == "" {
		return nil, fmt.Errorf("%w: query is required", ErrCustomsLookupInvalid)
	}

	switch provider {
	case CustomsLookupProviderUSHTS:
		return s.lookupUSHTS(query, limit)
	case CustomsLookupProviderUKTradeTariff:
		return s.lookupUKTradeTariff(query, limit)
	default:
		return nil, fmt.Errorf("%w: unsupported provider", ErrCustomsLookupInvalid)
	}
}

func normalizeCustomsClassificationInput(input CustomsClassificationInput) (*product.CustomsClassificationProfile, error) {
	name := strings.TrimSpace(input.Name)
	slug := normalizeCustomsClassificationSlug(input.Slug)
	hsCode := onlyDigits(input.HSCode)
	cnCode := onlyDigits(input.CNCode)
	countryOfOrigin := strings.ToUpper(strings.TrimSpace(input.CountryOfOrigin))
	status := strings.ToLower(strings.TrimSpace(input.Status))
	if status == "" {
		status = product.CustomsClassificationStatusActive
	}

	if name == "" || slug == "" {
		return nil, fmt.Errorf("%w: name and slug are required", ErrCustomsClassificationInvalid)
	}
	if hsCode == "" || len(hsCode) != 6 {
		return nil, fmt.Errorf("%w: HS Code must contain exactly 6 digits", ErrCustomsClassificationInvalid)
	}
	if cnCode != "" && len(cnCode) != 8 {
		return nil, fmt.Errorf("%w: EU CN Code must contain exactly 8 digits", ErrCustomsClassificationInvalid)
	}
	if countryOfOrigin != "" && (len(countryOfOrigin) != 2 || !isUppercaseLettersOnly(countryOfOrigin)) {
		return nil, fmt.Errorf("%w: country of origin must be a 2-letter ISO code", ErrCustomsClassificationInvalid)
	}
	customsDescription := strings.TrimSpace(input.CustomsDescription)
	if len(customsDescription) > 255 {
		return nil, fmt.Errorf("%w: customs description is too long", ErrCustomsClassificationInvalid)
	}
	if !product.IsCustomsClassificationStatus(status) {
		return nil, fmt.Errorf("%w: unsupported status", ErrCustomsClassificationInvalid)
	}

	return &product.CustomsClassificationProfile{
		ProductSpecificationTemplateID: input.ProductSpecificationTemplateID,
		Name:                           name,
		Slug:                           slug,
		ComponentKind:                  strings.TrimSpace(input.ComponentKind),
		Material:                       strings.TrimSpace(input.Material),
		HSCode:                         hsCode,
		CNCode:                         cnCode,
		CountryOfOrigin:                countryOfOrigin,
		CustomsDescription:             customsDescription,
		Source:                         strings.ToLower(strings.TrimSpace(input.Source)),
		SourceCode:                     strings.TrimSpace(input.SourceCode),
		SourceURL:                      strings.TrimSpace(input.SourceURL),
		Notes:                          strings.TrimSpace(input.Notes),
		Status:                         status,
	}, nil
}

func normalizeCustomsClassificationSlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "_", "-")
	spacePattern := regexp.MustCompile(`\s+`)
	value = spacePattern.ReplaceAllString(value, "-")
	unsupportedPattern := regexp.MustCompile(`[^a-z0-9-]+`)
	value = unsupportedPattern.ReplaceAllString(value, "")
	repeatedPattern := regexp.MustCompile(`-+`)
	value = repeatedPattern.ReplaceAllString(value, "-")
	return strings.Trim(value, "-")
}

func onlyDigits(value string) string {
	var builder strings.Builder
	for _, char := range strings.TrimSpace(value) {
		if char >= '0' && char <= '9' {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}

type usHTSLookupRow struct {
	HTSNo       string `json:"htsno"`
	Description string `json:"description"`
	General     string `json:"general"`
}

func (s *CustomsClassificationService) lookupUSHTS(query string, limit int) ([]CustomsClassificationLookupCandidate, error) {
	config, err := s.lookupProviderConfig(CustomsLookupProviderUSHTS, s.usHTSBaseURL)
	if err != nil {
		return nil, err
	}
	if !config.Enabled {
		return nil, fmt.Errorf("%w: US HTS lookup is disabled", ErrCustomsLookupUnavailable)
	}

	endpoint, err := url.Parse(applyCustomsLookupAPIKey(config.Endpoint, config.APIKey))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid US HTS endpoint", ErrCustomsLookupUnavailable)
	}
	params := endpoint.Query()
	params.Set("keyword", query)
	endpoint.RawQuery = params.Encode()

	var rows []usHTSLookupRow
	if err := s.getJSON(endpoint.String(), &rows, config); err != nil {
		return nil, err
	}

	candidates := make([]CustomsClassificationLookupCandidate, 0, limit)
	seen := map[string]struct{}{}
	for _, row := range rows {
		code := onlyDigits(row.HTSNo)
		if len(code) < 6 || strings.TrimSpace(row.Description) == "" {
			continue
		}
		if _, exists := seen[row.HTSNo]; exists {
			continue
		}
		seen[row.HTSNo] = struct{}{}
		candidates = append(candidates, CustomsClassificationLookupCandidate{
			Provider:           CustomsLookupProviderUSHTS,
			SourceCode:         strings.TrimSpace(row.HTSNo),
			HSCode:             code[:6],
			Description:        strings.TrimSpace(row.Description),
			CustomsDescription: titleLikeCustomsDescription(row.Description),
			Duty:               strings.TrimSpace(stripHTMLTags(row.General)),
			SourceURL:          "https://hts.usitc.gov/search?query=" + url.QueryEscape(query),
		})
		if len(candidates) >= limit {
			break
		}
	}
	return candidates, nil
}

type ukTradeTariffResponse struct {
	Data struct {
		Attributes struct {
			GoodsNomenclatureItemID string `json:"goods_nomenclature_item_id"`
			Description             string `json:"description"`
			DescriptionPlain        string `json:"description_plain"`
			FormattedDescription    string `json:"formatted_description"`
			Declarable              bool   `json:"declarable"`
		} `json:"attributes"`
	} `json:"data"`
}

func (s *CustomsClassificationService) lookupUKTradeTariff(query string, limit int) ([]CustomsClassificationLookupCandidate, error) {
	config, err := s.lookupProviderConfig(CustomsLookupProviderUKTradeTariff, s.ukTradeBaseURL)
	if err != nil {
		return nil, err
	}
	if !config.Enabled {
		return nil, fmt.Errorf("%w: UK Trade Tariff lookup is disabled", ErrCustomsLookupUnavailable)
	}

	code := onlyDigits(query)
	if len(code) != 8 && len(code) != 10 {
		return nil, fmt.Errorf("%w: UK Trade Tariff lookup expects an 8 or 10 digit commodity code", ErrCustomsLookupInvalid)
	}
	if len(code) == 8 {
		code += "00"
	}

	endpoint := strings.TrimRight(applyCustomsLookupAPIKey(config.Endpoint, config.APIKey), "/") + "/" + code
	var response ukTradeTariffResponse
	if err := s.getJSON(endpoint, &response, config); err != nil {
		return nil, err
	}
	itemCode := onlyDigits(response.Data.Attributes.GoodsNomenclatureItemID)
	if len(itemCode) < 8 {
		return nil, fmt.Errorf("%w: empty commodity response", ErrCustomsLookupUnavailable)
	}
	description := strings.TrimSpace(response.Data.Attributes.DescriptionPlain)
	if description == "" {
		description = strings.TrimSpace(response.Data.Attributes.Description)
	}
	if description == "" {
		description = strings.TrimSpace(response.Data.Attributes.FormattedDescription)
	}
	if description == "" {
		description = itemCode
	}

	candidate := CustomsClassificationLookupCandidate{
		Provider:           CustomsLookupProviderUKTradeTariff,
		SourceCode:         itemCode,
		HSCode:             itemCode[:6],
		CNCode:             itemCode[:8],
		Description:        description,
		CustomsDescription: titleLikeCustomsDescription(description),
		SourceURL:          "https://www.trade-tariff.service.gov.uk/commodities/" + itemCode,
	}
	return []CustomsClassificationLookupCandidate{candidate}, nil
}

func (s *CustomsClassificationService) getJSON(endpoint string, target interface{}, config customsLookupProviderConfig) error {
	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("%w: invalid lookup request", ErrCustomsLookupUnavailable)
	}
	request.Header.Set("Accept", "application/json")
	if strings.TrimSpace(config.APIKey) != "" && strings.TrimSpace(config.APIKeyHeader) != "" {
		request.Header.Set(strings.TrimSpace(config.APIKeyHeader), strings.TrimSpace(config.APIKey))
	}
	response, err := s.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCustomsLookupUnavailable, err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%w: provider returned %d", ErrCustomsLookupUnavailable, response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if err != nil {
		return fmt.Errorf("%w: failed to read provider response", ErrCustomsLookupUnavailable)
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("%w: invalid provider response", ErrCustomsLookupUnavailable)
	}
	return nil
}

func (s *CustomsClassificationService) lookupProviderConfig(provider, defaultEndpoint string) (customsLookupProviderConfig, error) {
	config := customsLookupProviderConfig{
		Enabled:      true,
		Endpoint:     defaultEndpoint,
		APIKeyHeader: customsLookupDefaultAPIKeyHeader,
	}
	if s == nil || s.settings == nil {
		return config, nil
	}

	settings, err := s.settings.GetByGroup(customsLookupSettingsGroup, customsLookupSettingsLocale)
	if err != nil {
		return config, fmt.Errorf("%w: failed to load lookup settings", ErrCustomsLookupUnavailable)
	}

	values := make(map[string]string, len(settings))
	for _, item := range settings {
		values[item.Key] = strings.TrimSpace(item.Value)
	}

	prefix := "customs_lookup_" + provider
	if value, ok := values[prefix+"_enabled"]; ok {
		config.Enabled = parseCustomsLookupBoolean(value, config.Enabled)
	}
	if value, ok := values[prefix+"_endpoint"]; ok && value != "" {
		config.Endpoint = value
	}
	if value, ok := values[prefix+"_api_key"]; ok {
		config.APIKey = value
	}
	if value, ok := values[prefix+"_api_key_header"]; ok {
		config.APIKeyHeader = value
	}
	return config, nil
}

func applyCustomsLookupAPIKey(endpoint, apiKey string) string {
	if strings.TrimSpace(apiKey) == "" {
		return endpoint
	}
	return strings.ReplaceAll(endpoint, "{apiKey}", url.PathEscape(strings.TrimSpace(apiKey)))
}

func parseCustomsLookupBoolean(value string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on", "enabled":
		return true
	case "0", "false", "no", "n", "off", "disabled":
		return false
	default:
		return fallback
	}
}

var htmlTagPattern = regexp.MustCompile(`<[^>]+>`)

func stripHTMLTags(value string) string {
	return strings.TrimSpace(htmlTagPattern.ReplaceAllString(value, ""))
}

func titleLikeCustomsDescription(value string) string {
	value = stripHTMLTags(value)
	value = strings.Trim(value, " :.;\t\n\r")
	if len(value) > 255 {
		value = value[:255]
	}
	return value
}
