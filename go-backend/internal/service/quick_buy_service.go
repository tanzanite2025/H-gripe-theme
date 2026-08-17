package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	currencydomain "commerce-platform/internal/domain/currency"
	productdomain "commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/quickbuy"
	"commerce-platform/internal/pkg/locales"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

var (
	ErrQuickBuyInvalid         = errors.New("invalid quick buy flow")
	ErrQuickBuyNotFound        = errors.New("quick buy flow not found")
	ErrQuickBuyNotMutable      = errors.New("quick buy version is not mutable")
	ErrQuickBuySessionNotFound = errors.New("quick buy session not found")

	quickBuyKeyPattern = regexp.MustCompile(`^[a-z0-9]+(?:[_-][a-z0-9]+)*$`)
)

const (
	quickBuyCandidateDefaultPageSize = 12
	quickBuyCandidateMaxPageSize     = 24
	quickBuyDefaultFlowSlug          = "quick-build"
	quickBuyDefaultFlowName          = "QUICK Build"
	quickBuyDefaultFlowDescription   = "Default QUICK build flow"
	quickBuyDefaultFlowEntrySurface  = "dock"
	quickBuyDefaultFlowSortOrder     = 100
)

var quickBuyDefaultStepKeys = [...]string{
	"product-search",
	"specifications",
	"quantity",
}

type QuickBuyService struct {
	repo                *repository.QuickBuyRepository
	productRepo         *repository.ProductRepository
	productCategoryRepo *repository.ProductCategoryRepository
	mediaURLResolver    PublicMediaURLResolver
}

type QuickBuyFlowInput struct {
	Slug         string                         `json:"slug"`
	Name         string                         `json:"name"`
	Description  string                         `json:"description"`
	HelpText     string                         `json:"help_text"`
	Translations []QuickBuyFlowTranslationInput `json:"translations"`
	EntrySurface string                         `json:"entry_surface"`
	IsEnabled    *bool                          `json:"is_enabled"`
	SortOrder    int                            `json:"sort_order"`
	Version      QuickBuyVersionInput           `json:"version"`
}

type QuickBuyFlowTranslationInput struct {
	ID       uint   `json:"id"`
	Locale   string `json:"locale"`
	HelpText string `json:"help_text"`
}

type QuickBuyVersionInput struct {
	StartsAt *time.Time          `json:"starts_at"`
	EndsAt   *time.Time          `json:"ends_at"`
	Steps    []QuickBuyStepInput `json:"steps"`
}

type QuickBuyStepInput struct {
	StepKey                         string `json:"step_key"`
	Name                            string `json:"name"`
	ProductCategoryIDs              []uint `json:"product_category_ids"`
	ProductSpecificationTemplateIDs []uint `json:"product_specification_template_ids"`
}

type QuickBuyFlowSummary struct {
	ID           uint                     `json:"id"`
	Slug         string                   `json:"slug"`
	Name         string                   `json:"name"`
	Description  string                   `json:"description"`
	HelpText     string                   `json:"help_text"`
	EntrySurface string                   `json:"entry_surface"`
	IsEnabled    bool                     `json:"is_enabled"`
	SortOrder    int                      `json:"sort_order"`
	Versions     []QuickBuyVersionSummary `json:"versions,omitempty"`
	CreatedAt    time.Time                `json:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at"`
}

type QuickBuyVersionSummary struct {
	ID            uint       `json:"id"`
	FlowID        uint       `json:"flow_id"`
	VersionNumber int        `json:"version_number"`
	Status        string     `json:"status"`
	PublishedAt   *time.Time `json:"published_at,omitempty"`
	StartsAt      *time.Time `json:"starts_at,omitempty"`
	EndsAt        *time.Time `json:"ends_at,omitempty"`
}

type QuickBuyFlowView struct {
	ID           uint                          `json:"id"`
	Slug         string                        `json:"slug"`
	Name         string                        `json:"name"`
	Description  string                        `json:"description"`
	HelpText     string                        `json:"help_text"`
	Translations []QuickBuyFlowTranslationView `json:"translations,omitempty"`
	EntrySurface string                        `json:"entry_surface"`
	IsEnabled    bool                          `json:"is_enabled"`
	SortOrder    int                           `json:"sort_order"`
	Version      QuickBuyVersionView           `json:"version"`
	Steps        []QuickBuyStepView            `json:"steps"`
}

// QuickBuyPublicFlowView is the runtime projection exposed to the storefront.
// It contains only the resolved locale content needed to render QUICK.
type QuickBuyPublicFlowView struct {
	ID           uint                `json:"id"`
	Slug         string              `json:"slug"`
	Name         string              `json:"name"`
	Description  string              `json:"description"`
	HelpText     string              `json:"help_text"`
	EntrySurface string              `json:"entry_surface"`
	IsEnabled    bool                `json:"is_enabled"`
	Version      QuickBuyVersionView `json:"version"`
	Steps        []QuickBuyStepView  `json:"steps"`
}

type QuickBuyFlowTranslationView struct {
	ID       uint   `json:"id"`
	Locale   string `json:"locale"`
	HelpText string `json:"help_text"`
}

type QuickBuyVersionView struct {
	ID            uint       `json:"id"`
	VersionNumber int        `json:"version_number"`
	Status        string     `json:"status"`
	PublishedAt   *time.Time `json:"published_at,omitempty"`
	StartsAt      *time.Time `json:"starts_at,omitempty"`
	EndsAt        *time.Time `json:"ends_at,omitempty"`
}

type QuickBuyStepView struct {
	ID                            uint                                       `json:"id"`
	StepKey                       string                                     `json:"step_key"`
	Slug                          string                                     `json:"slug"`
	Name                          string                                     `json:"name"`
	SortOrder                     int                                        `json:"sort_order"`
	ProductCategories             []QuickBuyProductCategoryView              `json:"product_categories"`
	ProductSpecificationTemplates []QuickBuyProductSpecificationTemplateView `json:"product_specification_templates"`
	Filters                       []QuickBuySpecFilterView                   `json:"filters"`
}

type QuickBuySpecFilterView struct {
	ID              uint     `json:"id"`
	Name            string   `json:"name"`
	Slug            string   `json:"slug"`
	Unit            string   `json:"unit,omitempty"`
	FieldType       string   `json:"field_type"`
	Presentation    string   `json:"presentation"`
	IsVariantOption bool     `json:"is_variant_option"`
	Multiple        bool     `json:"multiple"`
	Values          []string `json:"values"`
}

type QuickBuyProductSpecificationTemplateView struct {
	ID       uint   `json:"id"`
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	ImageURL string `json:"image_url,omitempty"`
	Primary  bool   `json:"primary"`
}

type QuickBuyProductCategoryView struct {
	ID       uint   `json:"id"`
	ParentID *uint  `json:"parent_id,omitempty"`
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Depth    int    `json:"depth"`
	ImageURL string `json:"image_url,omitempty"`
	Primary  bool   `json:"primary"`
}

type QuickBuyValidationResult struct {
	Valid  bool                      `json:"valid"`
	Issues []QuickBuyValidationIssue `json:"issues"`
}

type QuickBuyValidationIssue struct {
	Severity                       string `json:"severity"`
	Code                           string `json:"code"`
	Message                        string `json:"message"`
	StepKey                        string `json:"step_key,omitempty"`
	RuleKey                        string `json:"rule_key,omitempty"`
	ProductCategoryID              uint   `json:"product_category_id,omitempty"`
	ProductSpecificationTemplateID uint   `json:"product_specification_template_id,omitempty"`
}

type QuickBuySessionInput struct {
	FlowID        uint   `json:"flow_id"`
	FlowVersionID uint   `json:"flow_version_id"`
	Surface       string `json:"surface"`
	Locale        string `json:"locale"`
	MarketCountry string `json:"market_country"`
	Currency      string `json:"currency"`
	AnonymousID   string `json:"anonymous_id"`
	UserID        *uint  `json:"-"`
}

type QuickBuySelectionUpdateInput struct {
	Selections []QuickBuySelectionInput `json:"selections"`
}

type QuickBuySelectionInput struct {
	StepKey   string `json:"step_key"`
	ProductID uint   `json:"product_id"`
	VariantID *uint  `json:"variant_id"`
	Quantity  int    `json:"quantity"`
}

type QuickBuySessionValidationResult struct {
	Valid  bool                             `json:"valid"`
	Issues []QuickBuySessionValidationIssue `json:"issues"`
}

type QuickBuySessionValidationIssue struct {
	Severity  string `json:"severity"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	StepKey   string `json:"step_key,omitempty"`
	ProductID uint   `json:"product_id,omitempty"`
	VariantID *uint  `json:"variant_id,omitempty"`
}

type QuickBuySessionView struct {
	SessionToken     string                           `json:"session_token"`
	FlowID           uint                             `json:"flow_id"`
	FlowVersionID    uint                             `json:"flow_version_id"`
	Locale           string                           `json:"locale"`
	MarketCountry    string                           `json:"market_country"`
	Currency         string                           `json:"currency"`
	Status           string                           `json:"status"`
	ValidationStatus string                           `json:"validation_status"`
	SubtotalSnapshot float64                          `json:"subtotal_snapshot"`
	WeightSnapshotG  int                              `json:"weight_snapshot_g"`
	ExpiresAt        *time.Time                       `json:"expires_at,omitempty"`
	Flow             *QuickBuyPublicFlowView          `json:"flow,omitempty"`
	Items            []QuickBuySessionItemView        `json:"items"`
	Validation       *QuickBuySessionValidationResult `json:"validation,omitempty"`
	CreatedAt        time.Time                        `json:"created_at"`
	UpdatedAt        time.Time                        `json:"updated_at"`
}

type QuickBuySessionItemView struct {
	ID                uint           `json:"id"`
	StepKey           string         `json:"step_key"`
	ProductID         uint           `json:"product_id"`
	VariantID         *uint          `json:"variant_id,omitempty"`
	Quantity          int            `json:"quantity"`
	UnitPriceSnapshot float64        `json:"unit_price_snapshot"`
	CurrencySnapshot  string         `json:"currency_snapshot"`
	WeightSnapshotG   int            `json:"weight_snapshot_g"`
	ProductSnapshot   datatypes.JSON `json:"product_snapshot"`
	VariantSnapshot   datatypes.JSON `json:"variant_snapshot"`
	SortOrder         int            `json:"sort_order"`
}

type QuickBuyCandidateInput struct {
	StepKey     string              `json:"step_key"`
	Keyword     string              `json:"keyword"`
	Locale      string              `json:"locale"`
	Currency    string              `json:"currency"`
	SpecFilters map[string][]string `json:"spec_filters"`
	Page        int                 `json:"page"`
	PageSize    int                 `json:"page_size"`
}

type QuickBuyCandidateResult struct {
	FlowID        uint                    `json:"flow_id"`
	FlowVersionID uint                    `json:"flow_version_id"`
	Locale        string                  `json:"locale"`
	Currency      string                  `json:"currency"`
	Step          QuickBuyStepView        `json:"step"`
	Products      []productdomain.Product `json:"-"`
	Page          int                     `json:"page"`
	PageSize      int                     `json:"page_size"`
	Total         int64                   `json:"total"`
	HasMore       bool                    `json:"has_more"`
}

func NewQuickBuyService(repo *repository.QuickBuyRepository, productRepo *repository.ProductRepository, productCategoryRepo *repository.ProductCategoryRepository) *QuickBuyService {
	return &QuickBuyService{repo: repo, productRepo: productRepo, productCategoryRepo: productCategoryRepo}
}

func (s *QuickBuyService) ConfigureMediaService(mediaService *MediaService) {
	if s == nil {
		return
	}
	s.mediaURLResolver = mediaService
}

func (s *QuickBuyService) ListFlows() ([]QuickBuyFlowSummary, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("quick buy service is not configured")
	}
	flows, err := s.repo.ListFlows()
	if err != nil {
		return nil, err
	}
	result := make([]QuickBuyFlowSummary, 0, len(flows))
	for _, flow := range flows {
		result = append(result, quickBuyFlowSummary(flow))
	}
	return result, nil
}

func (s *QuickBuyService) GetFlow(id uint, locale string) (*QuickBuyFlowView, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("quick buy service is not configured")
	}
	flow, err := s.repo.FindFlowByID(id)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrQuickBuyNotFound
		}
		return nil, err
	}
	if len(flow.Versions) == 0 {
		requestLocale := locales.ResolveSupported(locale)
		return &QuickBuyFlowView{
			ID:           flow.ID,
			Slug:         flow.Slug,
			Name:         flow.Name,
			Description:  flow.Description,
			HelpText:     quickBuyFlowHelpText(*flow, requestLocale),
			Translations: quickBuyFlowTranslationViews(flow.Translations),
			EntrySurface: flow.EntrySurface,
			IsEnabled:    flow.IsEnabled,
			SortOrder:    flow.SortOrder,
			Version:      QuickBuyVersionView{},
			Steps:        []QuickBuyStepView{},
		}, nil
	}
	version, err := s.repo.FindVersionByID(flow.Versions[0].ID)
	if err != nil {
		return nil, err
	}
	return quickBuyFlowView(*version, locale), nil
}

func (s *QuickBuyService) CurrentFlow(surface, locale string) (*QuickBuyPublicFlowView, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("quick buy service is not configured")
	}
	version, err := s.findCurrentPublishedVersion(surface)
	if err != nil {
		return nil, err
	}
	if version == nil {
		return nil, nil
	}
	return quickBuyPublicFlowView(*version, locale, s.mediaURLResolver), nil
}

func (s *QuickBuyService) CreateSession(input QuickBuySessionInput) (*QuickBuySessionView, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("quick buy service is not configured")
	}
	version, locale, country, err := s.resolveSessionVersion(input)
	if err != nil {
		return nil, err
	}
	if version == nil {
		return nil, ErrQuickBuyNotFound
	}

	currency := normalizeQuickBuyCurrency(input.Currency)
	validation := s.validateQuickBuySession(*version, nil)
	expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour)
	session := &quickbuy.Session{
		SessionToken:     generateQuickBuySessionToken(),
		FlowID:           version.FlowID,
		FlowVersionID:    version.ID,
		Locale:           locale,
		MarketCountry:    country,
		Currency:         currency,
		AnonymousID:      strings.TrimSpace(input.AnonymousID),
		UserID:           input.UserID,
		Status:           quickbuy.SessionStatusActive,
		ValidationStatus: quickBuySessionValidationStatus(validation),
		Metadata:         datatypes.JSON([]byte("{}")),
		ExpiresAt:        &expiresAt,
	}
	if err := s.repo.CreateSession(session); err != nil {
		return nil, err
	}
	loaded, err := s.repo.FindSessionByToken(session.SessionToken)
	if err != nil {
		return nil, err
	}
	return quickBuySessionView(*loaded, &validation, s.mediaURLResolver), nil
}

func (s *QuickBuyService) GetSession(token string) (*QuickBuySessionView, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("quick buy service is not configured")
	}
	session, err := s.findActiveSession(token)
	if err != nil {
		return nil, err
	}
	return quickBuySessionView(*session, nil, s.mediaURLResolver), nil
}

func (s *QuickBuyService) UpdateSessionSelections(token string, input QuickBuySelectionUpdateInput) (*QuickBuySessionView, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("quick buy service is not configured")
	}
	session, err := s.findActiveSession(token)
	if err != nil {
		return nil, err
	}
	version, err := s.sessionVersion(session)
	if err != nil {
		return nil, err
	}

	stepKeys := make(map[string]struct{}, len(input.Selections))
	nextItems := make([]quickbuy.SessionItem, 0, len(session.Items)+len(input.Selections))
	for _, selection := range input.Selections {
		stepKey := normalizeQuickBuyKey(selection.StepKey)
		if stepKey == "" {
			return nil, fmt.Errorf("%w: selection step_key is required", ErrQuickBuyInvalid)
		}
		stepKeys[stepKey] = struct{}{}
	}
	for _, item := range session.Items {
		if _, replaced := stepKeys[item.StepKey]; replaced {
			continue
		}
		nextItems = append(nextItems, item)
	}

	for index, selection := range input.Selections {
		item, clearStep, err := s.sessionItemFromSelection(*session, *version, selection, index)
		if err != nil {
			return nil, err
		}
		if clearStep {
			continue
		}
		nextItems = append(nextItems, *item)
	}
	sort.SliceStable(nextItems, func(i, j int) bool {
		if nextItems[i].SortOrder == nextItems[j].SortOrder {
			return nextItems[i].ID < nextItems[j].ID
		}
		return nextItems[i].SortOrder < nextItems[j].SortOrder
	})
	if err := validateQuickBuySelectionBounds(*version, nextItems); err != nil {
		return nil, err
	}

	validation := s.validateQuickBuySession(*version, nextItems)
	subtotal, weightG := quickBuySessionTotals(nextItems)
	if err := s.repo.ReplaceSessionItems(session.ID, nextItems, session.Status, quickBuySessionValidationStatus(validation), subtotal, weightG); err != nil {
		return nil, err
	}
	loaded, err := s.repo.FindSessionByToken(session.SessionToken)
	if err != nil {
		return nil, err
	}
	return quickBuySessionView(*loaded, &validation, s.mediaURLResolver), nil
}

func (s *QuickBuyService) ValidateSession(token string) (*QuickBuySessionView, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("quick buy service is not configured")
	}
	session, err := s.findActiveSession(token)
	if err != nil {
		return nil, err
	}
	version, err := s.sessionVersion(session)
	if err != nil {
		return nil, err
	}
	validation := s.validateQuickBuySession(*version, session.Items)
	return quickBuySessionView(*session, &validation, s.mediaURLResolver), nil
}

func (s *QuickBuyService) ListSessionStepCandidates(token string, input QuickBuyCandidateInput) (*QuickBuyCandidateResult, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("quick buy service is not configured")
	}
	session, err := s.findActiveSession(token)
	if err != nil {
		return nil, err
	}
	version, err := s.sessionVersion(session)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Locale) == "" {
		input.Locale = session.Locale
	}
	if strings.TrimSpace(input.Currency) == "" {
		input.Currency = session.Currency
	}
	return s.listVersionStepCandidates(*version, input)
}

func (s *QuickBuyService) PreviewVersionStepCandidates(versionID uint, input QuickBuyCandidateInput) (*QuickBuyCandidateResult, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("quick buy service is not configured")
	}
	version, err := s.repo.FindVersionByID(versionID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrQuickBuyNotFound
		}
		return nil, err
	}
	return s.listVersionStepCandidates(*version, input)
}

func (s *QuickBuyService) CreateFlow(input QuickBuyFlowInput) (*QuickBuyFlowView, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("quick buy service is not configured")
	}
	flow, version, err := s.normalizeFlowAndVersion(input)
	if err != nil {
		return nil, err
	}
	if existing, err := s.repo.FindFlowBySlug(flow.Slug); err == nil && existing != nil {
		return nil, fmt.Errorf("%w: flow slug already exists", ErrQuickBuyInvalid)
	} else if err != nil && !repository.IsRecordNotFound(err) {
		return nil, err
	}
	if err := s.repo.CreateFlowWithVersion(flow, version); err != nil {
		return nil, err
	}
	loaded, err := s.repo.FindVersionByID(version.ID)
	if err != nil {
		return nil, err
	}
	return quickBuyFlowView(*loaded, ""), nil
}

func (s *QuickBuyService) UpdateFlow(id uint, input QuickBuyFlowInput) (*QuickBuyFlowSummary, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("quick buy service is not configured")
	}
	flow, err := normalizeFlowInput(input)
	if err != nil {
		return nil, err
	}
	existingFlow, err := s.repo.FindFlowByID(id)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrQuickBuyNotFound
		}
		return nil, err
	}
	if isDefaultQuickBuyFlowSlug(existingFlow.Slug) && !isDefaultQuickBuyFlowSlug(flow.Slug) {
		return nil, fmt.Errorf("%w: the default quick-build flow slug cannot be changed", ErrQuickBuyInvalid)
	}
	normalizeDefaultQuickBuyFlow(flow)
	flow.ID = id
	if existing, err := s.repo.FindFlowBySlug(flow.Slug); err == nil && existing != nil && existing.ID != id {
		return nil, fmt.Errorf("%w: flow slug already exists", ErrQuickBuyInvalid)
	} else if err != nil && !repository.IsRecordNotFound(err) {
		return nil, err
	}
	if err := s.repo.UpdateFlow(flow); err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrQuickBuyNotFound
		}
		return nil, err
	}
	loaded, err := s.repo.FindFlowByID(id)
	if err != nil {
		return nil, err
	}
	summary := quickBuyFlowSummary(*loaded)
	return &summary, nil
}

func (s *QuickBuyService) CreateDraftVersion(flowID uint, input QuickBuyVersionInput) (*QuickBuyFlowView, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("quick buy service is not configured")
	}
	flow, err := s.repo.FindFlowByID(flowID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrQuickBuyNotFound
		}
		return nil, err
	}
	version, err := s.normalizeVersionInputForFlow(flow.Slug, input)
	if err != nil {
		return nil, err
	}
	if err := validateQuickBuyDefaultSteps(flow.Slug, version.Steps); err != nil {
		return nil, err
	}
	latest, err := s.repo.FindLatestVersionNumber(flowID)
	if err != nil {
		return nil, err
	}
	version.FlowID = flowID
	version.VersionNumber = latest + 1
	version.Status = quickbuy.FlowVersionStatusDraft
	if err := s.repo.CreateVersion(version); err != nil {
		return nil, err
	}
	loaded, err := s.repo.FindVersionByID(version.ID)
	if err != nil {
		return nil, err
	}
	return quickBuyFlowView(*loaded, ""), nil
}

func (s *QuickBuyService) UpdateDraftVersion(versionID uint, input QuickBuyVersionInput) (*QuickBuyFlowView, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("quick buy service is not configured")
	}
	existing, err := s.repo.FindVersionByID(versionID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrQuickBuyNotFound
		}
		return nil, err
	}
	if existing.Status != quickbuy.FlowVersionStatusDraft {
		return nil, ErrQuickBuyNotMutable
	}
	flowSlug := ""
	if existing.Flow != nil {
		flowSlug = existing.Flow.Slug
	}
	version, err := s.normalizeVersionInputForFlow(flowSlug, input)
	if err != nil {
		return nil, err
	}
	if existing.Flow != nil {
		if err := validateQuickBuyDefaultSteps(existing.Flow.Slug, version.Steps); err != nil {
			return nil, err
		}
	}
	version.ID = existing.ID
	version.FlowID = existing.FlowID
	version.VersionNumber = existing.VersionNumber
	version.Status = existing.Status
	if err := s.repo.ReplaceVersion(version); err != nil {
		return nil, err
	}
	loaded, err := s.repo.FindVersionByID(version.ID)
	if err != nil {
		return nil, err
	}
	return quickBuyFlowView(*loaded, ""), nil
}

func (s *QuickBuyService) SaveFlowConfiguration(flowID uint, input QuickBuyFlowInput) (*QuickBuyFlowView, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("quick buy service is not configured")
	}
	flow, err := normalizeFlowInput(input)
	if err != nil {
		return nil, err
	}
	existingFlow, err := s.repo.FindFlowByID(flowID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrQuickBuyNotFound
		}
		return nil, err
	}
	if isDefaultQuickBuyFlowSlug(existingFlow.Slug) && !isDefaultQuickBuyFlowSlug(flow.Slug) {
		return nil, fmt.Errorf("%w: the default quick-build flow slug cannot be changed", ErrQuickBuyInvalid)
	}
	normalizeDefaultQuickBuyFlow(flow)
	flow.ID = flowID
	if existing, err := s.repo.FindFlowBySlug(flow.Slug); err == nil && existing != nil && existing.ID != flowID {
		return nil, fmt.Errorf("%w: flow slug already exists", ErrQuickBuyInvalid)
	} else if err != nil && !repository.IsRecordNotFound(err) {
		return nil, err
	}

	version, err := s.normalizeVersionInputForFlow(flow.Slug, input.Version)
	if err != nil {
		return nil, err
	}
	if err := validateQuickBuyDefaultSteps(flow.Slug, version.Steps); err != nil {
		return nil, err
	}
	versionID := uint(0)
	for _, candidate := range existingFlow.Versions {
		if candidate.Status == quickbuy.FlowVersionStatusDraft {
			versionID = candidate.ID
			break
		}
	}
	if versionID > 0 {
		existingVersion, err := s.repo.FindVersionByID(versionID)
		if err != nil {
			if repository.IsRecordNotFound(err) {
				return nil, ErrQuickBuyNotFound
			}
			return nil, err
		}
		if existingVersion.Status != quickbuy.FlowVersionStatusDraft {
			return nil, ErrQuickBuyNotMutable
		}
		version.ID = existingVersion.ID
		version.FlowID = existingVersion.FlowID
		version.VersionNumber = existingVersion.VersionNumber
		version.Status = existingVersion.Status
	} else {
		latest, err := s.repo.FindLatestVersionNumber(flowID)
		if err != nil {
			return nil, err
		}
		version.FlowID = flowID
		version.VersionNumber = latest + 1
		version.Status = quickbuy.FlowVersionStatusDraft
	}

	if err := s.repo.SaveFlowConfiguration(flow, version, versionID > 0); err != nil {
		return nil, err
	}
	loaded, err := s.repo.FindVersionByID(version.ID)
	if err != nil {
		return nil, err
	}
	return quickBuyFlowView(*loaded, ""), nil
}

func (s *QuickBuyService) PublishVersion(versionID uint, publishedBy *uint) (*QuickBuyFlowView, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("quick buy service is not configured")
	}
	version, err := s.repo.FindVersionByID(versionID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrQuickBuyNotFound
		}
		return nil, err
	}
	if err := validateQuickBuyVersionForPublish(*version); err != nil {
		return nil, err
	}
	if err := s.repo.PublishVersion(versionID, publishedBy, time.Now().UTC()); err != nil {
		return nil, err
	}
	loaded, err := s.repo.FindVersionByID(versionID)
	if err != nil {
		return nil, err
	}
	return quickBuyFlowView(*loaded, ""), nil
}

func (s *QuickBuyService) ValidateVersion(versionID uint) (*QuickBuyValidationResult, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("quick buy service is not configured")
	}
	version, err := s.repo.FindVersionByID(versionID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrQuickBuyNotFound
		}
		return nil, err
	}
	result := validateQuickBuyVersion(*version)
	return &result, nil
}

func (s *QuickBuyService) findCurrentPublishedVersion(surface string) (*quickbuy.Version, error) {
	surface = normalizeQuickBuySurface(surface)

	versions, err := s.repo.ListPublishedVersions(surface, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	if len(versions) == 0 {
		return nil, nil
	}
	version := versions[0]
	return &version, nil
}

func (s *QuickBuyService) resolveSessionVersion(input QuickBuySessionInput) (*quickbuy.Version, string, string, error) {
	locale := locales.ResolveSupported(input.Locale)
	country := normalizeQuickBuyCountry(input.MarketCountry)
	if input.FlowVersionID == 0 {
		version, err := s.findCurrentPublishedVersion(input.Surface)
		return version, locale, country, err
	}

	version, err := s.repo.FindVersionByID(input.FlowVersionID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, locale, country, ErrQuickBuyNotFound
		}
		return nil, locale, country, err
	}
	if version.Status != quickbuy.FlowVersionStatusPublished || version.Flow == nil || !version.Flow.IsEnabled {
		return nil, locale, country, ErrQuickBuyNotFound
	}
	if !quickBuyVersionIsActive(*version, time.Now().UTC()) {
		return nil, locale, country, ErrQuickBuyNotFound
	}
	if input.FlowID > 0 && version.FlowID != input.FlowID {
		return nil, locale, country, fmt.Errorf("%w: flow_id does not match flow_version_id", ErrQuickBuyInvalid)
	}
	return version, locale, country, nil
}

func (s *QuickBuyService) findActiveSession(token string) (*quickbuy.Session, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrQuickBuySessionNotFound
	}
	session, err := s.repo.FindSessionByToken(token)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrQuickBuySessionNotFound
		}
		return nil, err
	}
	if session.ExpiresAt != nil && session.ExpiresAt.Before(time.Now().UTC()) {
		return nil, ErrQuickBuySessionNotFound
	}
	return session, nil
}

func (s *QuickBuyService) sessionVersion(session *quickbuy.Session) (*quickbuy.Version, error) {
	if session.Version != nil {
		return session.Version, nil
	}
	version, err := s.repo.FindVersionByID(session.FlowVersionID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrQuickBuyNotFound
		}
		return nil, err
	}
	return version, nil
}

func (s *QuickBuyService) listVersionStepCandidates(version quickbuy.Version, input QuickBuyCandidateInput) (*QuickBuyCandidateResult, error) {
	if s.productRepo == nil {
		return nil, errors.New("product repository is not configured")
	}

	stepKey := normalizeQuickBuyKey(input.StepKey)
	if stepKey == "" {
		return nil, fmt.Errorf("%w: step_key is required", ErrQuickBuyInvalid)
	}
	step := quickBuyStepByKey(version, stepKey)
	if step == nil {
		return nil, fmt.Errorf("%w: step %q does not exist in this QUICK version", ErrQuickBuyInvalid, stepKey)
	}

	locale := locales.ResolveSupported(input.Locale)
	currency := normalizeQuickBuyCurrency(input.Currency)
	page, pageSize := normalizeQuickBuyCandidatePaging(input.Page, input.PageSize)
	exposeProductSpecificationTemplates := !isDefaultQuickBuyFlow(version) && len(step.ProductSpecificationTemplates) > 0
	if !exposeProductSpecificationTemplates {
		input.SpecFilters = nil
	}
	specFilters, err := normalizeQuickBuySpecFilters(*step, input.SpecFilters)
	if err != nil {
		return nil, err
	}
	result := &QuickBuyCandidateResult{
		FlowID:        version.FlowID,
		FlowVersionID: version.ID,
		Locale:        locale,
		Currency:      currency,
		Step:          quickBuyStepView(*step, locale, exposeProductSpecificationTemplates, s.mediaURLResolver),
		Products:      []productdomain.Product{},
		Page:          page,
		PageSize:      pageSize,
	}
	if step.SelectionMode == quickbuy.SelectionModeAuto {
		return result, nil
	}

	products, total, err := s.productRepo.ListQuickBuyCandidates(repository.ProductQuickBuyCandidateQuery{
		Locale:                          locale,
		ProductSpecificationTemplateIDs: quickBuyStepProductSpecificationTemplateIDsForVersion(version, *step),
		ProductCategoryIDs:              quickBuyStepProductCategoryIDs(*step),
		Keyword:                         strings.TrimSpace(input.Keyword),
		SpecFilters:                     specFilters,
		Offset:                          (page - 1) * pageSize,
		Limit:                           pageSize,
	})
	if err != nil {
		return nil, err
	}
	result.Products = products
	result.Total = total
	result.HasMore = int64(page*pageSize) < total
	if !exposeProductSpecificationTemplates {
		return result, nil
	}
	filterValues, err := s.productRepo.ListQuickBuyFilterValues(repository.ProductQuickBuyCandidateQuery{
		Locale:                          locale,
		ProductSpecificationTemplateIDs: quickBuyStepProductSpecificationTemplateIDsForVersion(version, *step),
		ProductCategoryIDs:              quickBuyStepProductCategoryIDs(*step),
		Keyword:                         strings.TrimSpace(input.Keyword),
		SpecFilters:                     specFilters,
	}, quickBuyFilterableSpecSlugs(*step))
	if err != nil {
		return nil, err
	}
	result.Step.Filters = quickBuyStepFiltersForScope(*step, filterValues, exposeProductSpecificationTemplates)
	return result, nil
}

func (s *QuickBuyService) sessionItemFromSelection(session quickbuy.Session, version quickbuy.Version, selection QuickBuySelectionInput, index int) (*quickbuy.SessionItem, bool, error) {
	stepKey := normalizeQuickBuyKey(selection.StepKey)
	if stepKey == "" {
		return nil, false, fmt.Errorf("%w: selection step_key is required", ErrQuickBuyInvalid)
	}
	step := quickBuyStepByKey(version, stepKey)
	if step == nil {
		return nil, false, fmt.Errorf("%w: step %q does not exist in this QUICK version", ErrQuickBuyInvalid, stepKey)
	}
	if selection.ProductID == 0 {
		return nil, true, nil
	}
	if step.SelectionMode == quickbuy.SelectionModeAuto {
		return nil, false, fmt.Errorf("%w: step %q does not accept manual selections", ErrQuickBuyInvalid, stepKey)
	}
	quantity := selection.Quantity
	if quantity <= 0 {
		quantity = step.DefaultQuantity
	}
	if quantity <= 0 {
		quantity = 1
	}
	if quantity > 999 {
		return nil, false, fmt.Errorf("%w: step %q quantity is too large", ErrQuickBuyInvalid, stepKey)
	}
	if s.productRepo == nil {
		return nil, false, errors.New("product repository is not configured")
	}
	productItem, variant, err := s.productRepo.FindPurchasableVariant(selection.ProductID, selection.VariantID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, false, fmt.Errorf("%w: product %d is not available", ErrQuickBuyInvalid, selection.ProductID)
		}
		return nil, false, err
	}
	if err := s.validateQuickBuyProductAllowedForStep(*step, *productItem, !isDefaultQuickBuyFlow(version)); err != nil {
		return nil, false, err
	}
	if variant.Stock < quantity {
		return nil, false, fmt.Errorf("%w: product %d does not have enough stock for step %q", ErrQuickBuyInvalid, selection.ProductID, stepKey)
	}

	variantID := variant.ID
	price := variant.EffectivePrice()
	currency := normalizeQuickBuyCurrency(variant.Currency)
	if currency == "" {
		currency = normalizeQuickBuyCurrency(productItem.DisplayPriceCurrency())
	}
	if currency == "" {
		currency = session.Currency
	}
	return &quickbuy.SessionItem{
		StepID:            step.ID,
		StepKey:           step.StepKey,
		ProductID:         productItem.ID,
		VariantID:         &variantID,
		Quantity:          quantity,
		UnitPriceSnapshot: price,
		CurrencySnapshot:  currency,
		WeightSnapshotG:   variant.Weight,
		ProductSnapshot:   quickBuyProductSnapshot(*productItem, s.mediaURLResolver),
		VariantSnapshot:   quickBuyVariantSnapshot(*variant),
		SortOrder:         step.SortOrder*100 + index + 1,
	}, false, nil
}

func (s *QuickBuyService) validateQuickBuySession(version quickbuy.Version, items []quickbuy.SessionItem) QuickBuySessionValidationResult {
	result := QuickBuySessionValidationResult{Valid: true, Issues: []QuickBuySessionValidationIssue{}}
	itemsByStep := make(map[string][]quickbuy.SessionItem, len(items))
	for _, item := range items {
		itemsByStep[item.StepKey] = append(itemsByStep[item.StepKey], item)
		step := quickBuyStepByKey(version, item.StepKey)
		if step == nil {
			result.addIssue("step_missing", fmt.Sprintf("selection references missing step %q", item.StepKey), item.StepKey, item.ProductID, item.VariantID)
			continue
		}
		if item.Quantity <= 0 {
			result.addIssue("invalid_quantity", fmt.Sprintf("step %q has a non-positive quantity", item.StepKey), item.StepKey, item.ProductID, item.VariantID)
		}
		if s.productRepo == nil {
			continue
		}
		productItem, variant, err := s.productRepo.FindPurchasableVariant(item.ProductID, item.VariantID)
		if err != nil {
			result.addIssue("product_unavailable", fmt.Sprintf("product %d is no longer available", item.ProductID), item.StepKey, item.ProductID, item.VariantID)
			continue
		}
		if err := s.validateQuickBuyProductAllowedForStep(*step, *productItem, !isDefaultQuickBuyFlow(version)); err != nil {
			result.addIssue("product_not_allowed", err.Error(), item.StepKey, item.ProductID, item.VariantID)
		}
		if variant.Stock < item.Quantity {
			result.addIssue("stock_unavailable", fmt.Sprintf("product %d no longer has enough stock", item.ProductID), item.StepKey, item.ProductID, item.VariantID)
		}
	}

	for _, step := range version.Steps {
		stepItems := itemsByStep[step.StepKey]
		if step.IsRequired && step.SelectionMode != quickbuy.SelectionModeAuto && len(stepItems) == 0 {
			result.addIssue("required_step_missing", fmt.Sprintf("required step %q has no selection", step.StepKey), step.StepKey, 0, nil)
		}
		if step.SelectionMode == quickbuy.SelectionModeSingle && len(stepItems) > 1 {
			result.addIssue("single_step_multiple_items", fmt.Sprintf("single-select step %q has more than one selection", step.StepKey), step.StepKey, 0, nil)
		}
		if step.MaxSelect > 0 && len(stepItems) > step.MaxSelect {
			result.addIssue("max_select_exceeded", fmt.Sprintf("step %q exceeds max_select", step.StepKey), step.StepKey, 0, nil)
		}
		if step.MinSelect > 0 && len(stepItems) < step.MinSelect {
			result.addIssue("min_select_missing", fmt.Sprintf("step %q has fewer selections than min_select", step.StepKey), step.StepKey, 0, nil)
		}
	}
	return result
}

func (s *QuickBuyService) normalizeFlowAndVersion(input QuickBuyFlowInput) (*quickbuy.Flow, *quickbuy.Version, error) {
	flow, err := normalizeFlowInput(input)
	if err != nil {
		return nil, nil, err
	}
	normalizeDefaultQuickBuyFlow(flow)
	version, err := s.normalizeVersionInputForFlow(flow.Slug, input.Version)
	if err != nil {
		return nil, nil, err
	}
	if err := validateQuickBuyDefaultSteps(flow.Slug, version.Steps); err != nil {
		return nil, nil, err
	}
	version.VersionNumber = 1
	version.Status = quickbuy.FlowVersionStatusDraft
	return flow, version, nil
}

func normalizeFlowInput(input QuickBuyFlowInput) (*quickbuy.Flow, error) {
	slug := normalizeQuickBuyKey(input.Slug)
	if slug == "" {
		return nil, fmt.Errorf("%w: slug is required", ErrQuickBuyInvalid)
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrQuickBuyInvalid)
	}
	enabled := true
	if input.IsEnabled != nil {
		enabled = *input.IsEnabled
	}
	sortOrder := input.SortOrder
	if sortOrder <= 0 {
		sortOrder = 100
	}
	translations, err := normalizeQuickBuyFlowTranslations(input.Translations)
	if err != nil {
		return nil, err
	}
	flow := &quickbuy.Flow{
		Slug:         slug,
		Name:         name,
		Description:  strings.TrimSpace(input.Description),
		HelpText:     strings.TrimSpace(input.HelpText),
		Translations: translations,
		EntrySurface: normalizeQuickBuySurface(input.EntrySurface),
		IsEnabled:    enabled,
		SortOrder:    sortOrder,
	}
	return flow, nil
}

func normalizeQuickBuyFlowTranslations(input []QuickBuyFlowTranslationInput) ([]quickbuy.FlowTranslation, error) {
	if len(input) == 0 {
		return nil, nil
	}

	result := make([]quickbuy.FlowTranslation, 0, len(input))
	seenLocales := make(map[string]struct{}, len(input))
	for index, item := range input {
		localeValue := strings.TrimSpace(item.Locale)
		if localeValue == "" {
			continue
		}
		locale, err := requireSupportedLocale(localeValue)
		if err != nil {
			return nil, fmt.Errorf("%w: flow translation %d has invalid locale", ErrQuickBuyInvalid, index+1)
		}
		helpText := strings.TrimSpace(item.HelpText)
		if helpText == "" {
			continue
		}
		if _, exists := seenLocales[locale]; exists {
			return nil, fmt.Errorf("%w: flow has duplicate translation locale %q", ErrQuickBuyInvalid, locale)
		}
		seenLocales[locale] = struct{}{}
		result = append(result, quickbuy.FlowTranslation{
			ID:       item.ID,
			Locale:   locale,
			HelpText: helpText,
		})
	}
	return result, nil
}

func (s *QuickBuyService) normalizeVersionInput(input QuickBuyVersionInput) (*quickbuy.Version, error) {
	if input.EndsAt != nil && input.StartsAt != nil && !input.EndsAt.After(*input.StartsAt) {
		return nil, fmt.Errorf("%w: ends_at must be after starts_at", ErrQuickBuyInvalid)
	}
	steps, err := s.normalizeStepInputs(input.Steps)
	if err != nil {
		return nil, err
	}
	return &quickbuy.Version{
		Status:   quickbuy.FlowVersionStatusDraft,
		StartsAt: input.StartsAt,
		EndsAt:   input.EndsAt,
		Steps:    steps,
	}, nil
}

func (s *QuickBuyService) normalizeVersionInputForFlow(flowSlug string, input QuickBuyVersionInput) (*quickbuy.Version, error) {
	version, err := s.normalizeVersionInput(input)
	if err != nil {
		return nil, err
	}
	if isDefaultQuickBuyFlowSlug(flowSlug) {
		if err := normalizeDefaultQuickBuySteps(version.Steps); err != nil {
			return nil, err
		}
	}
	return version, nil
}

func (s *QuickBuyService) normalizeStepInputs(input []QuickBuyStepInput) ([]quickbuy.Step, error) {
	steps := make([]quickbuy.Step, 0, len(input))
	seen := make(map[string]struct{}, len(input))
	for index, item := range input {
		stepKey := normalizeQuickBuyKey(item.StepKey)
		if stepKey == "" {
			return nil, fmt.Errorf("%w: step %d key is required", ErrQuickBuyInvalid, index+1)
		}
		if _, exists := seen[stepKey]; exists {
			return nil, fmt.Errorf("%w: duplicate step key %q", ErrQuickBuyInvalid, stepKey)
		}
		seen[stepKey] = struct{}{}
		name := strings.TrimSpace(item.Name)
		if name == "" {
			return nil, fmt.Errorf("%w: step %q name is required", ErrQuickBuyInvalid, stepKey)
		}
		productCategories, err := s.normalizeStepProductCategories(item.ProductCategoryIDs)
		if err != nil {
			return nil, fmt.Errorf("%w: step %q product categories: %v", ErrQuickBuyInvalid, stepKey, err)
		}
		productSpecificationTemplates, err := s.normalizeStepProductSpecificationTemplates(item.ProductSpecificationTemplateIDs)
		if err != nil {
			return nil, fmt.Errorf("%w: step %q product specification templates: %v", ErrQuickBuyInvalid, stepKey, err)
		}
		steps = append(steps, quickbuy.Step{
			StepKey:                       stepKey,
			Name:                          name,
			SortOrder:                     (index + 1) * 10,
			SelectionMode:                 quickbuy.SelectionModeSingle,
			IsRequired:                    true,
			MinSelect:                     0,
			MaxSelect:                     1,
			DefaultQuantity:               1,
			AllowSkip:                     false,
			ProductCategories:             productCategories,
			ProductSpecificationTemplates: productSpecificationTemplates,
		})
	}
	sort.SliceStable(steps, func(i, j int) bool {
		if steps[i].SortOrder == steps[j].SortOrder {
			return steps[i].StepKey < steps[j].StepKey
		}
		return steps[i].SortOrder < steps[j].SortOrder
	})
	return steps, nil
}

func (s *QuickBuyService) normalizeStepProductCategories(ids []uint) ([]quickbuy.StepProductCategory, error) {
	seen := make(map[uint]struct{}, len(ids))
	result := make([]quickbuy.StepProductCategory, 0, len(ids))
	for index, id := range ids {
		if id == 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		if s.productCategoryRepo != nil {
			if _, err := s.productCategoryRepo.FindByID(id); err != nil {
				if repository.IsRecordNotFound(err) {
					return nil, fmt.Errorf("product category %d does not exist", id)
				}
				return nil, err
			}
		}
		result = append(result, quickbuy.StepProductCategory{
			ProductCategoryID: id,
			IsPrimary:         len(result) == 0,
			SortOrder:         (index + 1) * 10,
		})
	}
	return result, nil
}

func (s *QuickBuyService) normalizeStepProductSpecificationTemplates(ids []uint) ([]quickbuy.StepProductSpecificationTemplate, error) {
	seen := make(map[uint]struct{}, len(ids))
	result := make([]quickbuy.StepProductSpecificationTemplate, 0, len(ids))
	for index, id := range ids {
		if id == 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		if s.productRepo != nil {
			if _, err := s.productRepo.FindProductSpecificationTemplateByID(id); err != nil {
				if repository.IsRecordNotFound(err) {
					return nil, fmt.Errorf("product specification template %d does not exist", id)
				}
				return nil, err
			}
		}
		result = append(result, quickbuy.StepProductSpecificationTemplate{
			ProductSpecificationTemplateID: id,
			IsPrimary:                      len(result) == 0,
			SortOrder:                      (index + 1) * 10,
		})
	}
	return result, nil
}

func validateQuickBuyVersionForPublish(version quickbuy.Version) error {
	result := validateQuickBuyVersion(version)
	if result.Valid {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrQuickBuyInvalid, result.errorSummary())
}

func validateQuickBuyVersion(version quickbuy.Version) QuickBuyValidationResult {
	result := QuickBuyValidationResult{Valid: true, Issues: []QuickBuyValidationIssue{}}
	if version.Flow == nil {
		result.addIssue("error", "flow_missing", "version is not linked to a QUICK flow", "", "", 0)
	}
	if version.EndsAt != nil && version.StartsAt != nil && !version.EndsAt.After(*version.StartsAt) {
		result.addIssue("error", "invalid_time_window", "ends_at must be after starts_at", "", "", 0)
	}
	if len(version.Steps) == 0 {
		result.addIssue("error", "steps_required", "at least one step is required before publishing", "", "", 0)
		return result
	}

	stepKeys := make(map[string]struct{}, len(version.Steps))
	for _, step := range version.Steps {
		stepKey := normalizeQuickBuyKey(step.StepKey)
		if stepKey == "" || stepKey != step.StepKey {
			result.addIssue("error", "invalid_step_key", fmt.Sprintf("step key %q is invalid", step.StepKey), step.StepKey, "", 0)
		}
		if _, exists := stepKeys[step.StepKey]; exists {
			result.addIssue("error", "duplicate_step_key", fmt.Sprintf("step key %q is duplicated", step.StepKey), step.StepKey, "", 0)
		}
		stepKeys[step.StepKey] = struct{}{}
		if strings.TrimSpace(step.Name) == "" {
			result.addIssue("error", "step_name_required", "step name is required", step.StepKey, "", 0)
		}
		if !quickBuySelectionModeIsValid(step.SelectionMode) {
			result.addIssue("error", "invalid_selection_mode", fmt.Sprintf("step %q has invalid selection mode %q", step.StepKey, step.SelectionMode), step.StepKey, "", 0)
		}
		if step.MinSelect < 0 || step.MaxSelect < 0 || step.DefaultQuantity <= 0 {
			result.addIssue("error", "invalid_selection_bounds", fmt.Sprintf("step %q has invalid quantity or selection bounds", step.StepKey), step.StepKey, "", 0)
		}
		if step.SelectionMode != quickbuy.SelectionModeAuto && step.MaxSelect > 0 && step.MinSelect > step.MaxSelect {
			result.addIssue("error", "invalid_selection_bounds", fmt.Sprintf("step %q min_select cannot exceed max_select", step.StepKey), step.StepKey, "", 0)
		}
		if step.SelectionMode == quickbuy.SelectionModeSingle && step.MaxSelect > 1 {
			result.addIssue("error", "single_step_max_select", fmt.Sprintf("single-select step %q cannot allow more than one selection", step.StepKey), step.StepKey, "", 0)
		}
		if step.IsRequired && len(step.ProductCategories) == 0 && len(step.ProductSpecificationTemplates) == 0 && step.SelectionMode != quickbuy.SelectionModeAuto && !isDefaultQuickBuyFlow(version) {
			result.addIssue("error", "required_step_product_categories", fmt.Sprintf("required step %q needs at least one product category", step.StepKey), step.StepKey, "", 0)
		}
		if step.IsRequired && step.AllowSkip {
			result.addIssue("warning", "required_step_allows_skip", fmt.Sprintf("required step %q is also marked as skippable", step.StepKey), step.StepKey, "", 0)
		}
		if !step.IsRequired && step.MinSelect > 0 {
			result.addIssue("warning", "optional_step_min_select", fmt.Sprintf("optional step %q has min_select greater than zero", step.StepKey), step.StepKey, "", 0)
		}

		productCategories := make(map[uint]struct{}, len(step.ProductCategories))
		for _, item := range step.ProductCategories {
			if item.ProductCategoryID == 0 {
				result.addProductCategoryIssue("error", "invalid_product_category", fmt.Sprintf("step %q contains an empty product category reference", step.StepKey), step.StepKey, item.ProductCategoryID)
				continue
			}
			if _, exists := productCategories[item.ProductCategoryID]; exists {
				result.addProductCategoryIssue("error", "duplicate_product_category", fmt.Sprintf("step %q references product category %d more than once", step.StepKey, item.ProductCategoryID), step.StepKey, item.ProductCategoryID)
				continue
			}
			productCategories[item.ProductCategoryID] = struct{}{}
			if item.ProductCategory == nil {
				result.addProductCategoryIssue("error", "missing_product_category", fmt.Sprintf("step %q references missing product category %d", step.StepKey, item.ProductCategoryID), step.StepKey, item.ProductCategoryID)
				continue
			}
			if !item.ProductCategory.IsEnabled {
				result.addProductCategoryIssue("error", "disabled_product_category", fmt.Sprintf("step %q references disabled product category %q", step.StepKey, item.ProductCategory.Slug), step.StepKey, item.ProductCategoryID)
			}
		}

		if !isDefaultQuickBuyFlow(version) {
			productSpecificationTemplates := make(map[uint]struct{}, len(step.ProductSpecificationTemplates))
			for _, item := range step.ProductSpecificationTemplates {
				if item.ProductSpecificationTemplateID == 0 {
					result.addIssue("error", "invalid_product_specification_template", fmt.Sprintf("step %q contains an empty product specification template reference", step.StepKey), step.StepKey, "", 0)
					continue
				}
				if _, exists := productSpecificationTemplates[item.ProductSpecificationTemplateID]; exists {
					result.addIssue("error", "duplicate_product_specification_template", fmt.Sprintf("step %q references product specification template %d more than once", step.StepKey, item.ProductSpecificationTemplateID), step.StepKey, "", item.ProductSpecificationTemplateID)
					continue
				}
				productSpecificationTemplates[item.ProductSpecificationTemplateID] = struct{}{}
				if item.ProductSpecificationTemplate == nil {
					result.addIssue("error", "missing_product_specification_template", fmt.Sprintf("step %q references missing product specification template %d", step.StepKey, item.ProductSpecificationTemplateID), step.StepKey, "", item.ProductSpecificationTemplateID)
					continue
				}
				if !item.ProductSpecificationTemplate.IsEnabled {
					result.addIssue("error", "disabled_product_specification_template", fmt.Sprintf("step %q references disabled product specification template %q", step.StepKey, item.ProductSpecificationTemplate.Slug), step.StepKey, "", item.ProductSpecificationTemplateID)
				}
			}
		}
	}
	if version.Flow != nil && isDefaultQuickBuyFlowSlug(version.Flow.Slug) {
		for _, requiredStepKey := range quickBuyDefaultStepKeys {
			if _, exists := stepKeys[requiredStepKey]; !exists {
				result.addIssue(
					"error",
					"default_step_required",
					fmt.Sprintf("default quick-build step %q cannot be removed", requiredStepKey),
					requiredStepKey,
					"",
					0,
				)
			}
		}
	}

	ruleKeys := make(map[string]struct{}, len(version.Rules))
	for _, rule := range version.Rules {
		if !rule.IsEnabled {
			continue
		}
		ruleKey := normalizeQuickBuyKey(rule.RuleKey)
		if ruleKey == "" || ruleKey != rule.RuleKey {
			result.addIssue("error", "invalid_rule_key", fmt.Sprintf("rule key %q is invalid", rule.RuleKey), "", rule.RuleKey, 0)
		}
		if _, exists := ruleKeys[rule.RuleKey]; exists {
			result.addIssue("error", "duplicate_rule_key", fmt.Sprintf("rule key %q is duplicated", rule.RuleKey), "", rule.RuleKey, 0)
		}
		ruleKeys[rule.RuleKey] = struct{}{}
		if rule.SourceStepKey != "" {
			if _, exists := stepKeys[rule.SourceStepKey]; !exists {
				result.addIssue("error", "rule_source_step_missing", fmt.Sprintf("rule %q references missing source step %q", rule.RuleKey, rule.SourceStepKey), rule.SourceStepKey, rule.RuleKey, 0)
			}
		}
		if rule.TargetStepKey != "" {
			if _, exists := stepKeys[rule.TargetStepKey]; !exists {
				result.addIssue("error", "rule_target_step_missing", fmt.Sprintf("rule %q references missing target step %q", rule.RuleKey, rule.TargetStepKey), rule.TargetStepKey, rule.RuleKey, 0)
			}
		}
		if rule.Severity != "error" && rule.Severity != "warning" && rule.Severity != "info" {
			result.addIssue("error", "invalid_rule_severity", fmt.Sprintf("rule %q has invalid severity %q", rule.RuleKey, rule.Severity), "", rule.RuleKey, 0)
		}
	}

	return result
}

func (result *QuickBuyValidationResult) addIssue(severity, code, message, stepKey, ruleKey string, productSpecificationTemplateID uint) {
	if severity == "error" {
		result.Valid = false
	}
	result.Issues = append(result.Issues, QuickBuyValidationIssue{
		Severity:                       severity,
		Code:                           code,
		Message:                        message,
		StepKey:                        stepKey,
		RuleKey:                        ruleKey,
		ProductSpecificationTemplateID: productSpecificationTemplateID,
	})
}

func (result *QuickBuyValidationResult) addProductCategoryIssue(severity, code, message, stepKey string, productCategoryID uint) {
	if severity == "error" {
		result.Valid = false
	}
	result.Issues = append(result.Issues, QuickBuyValidationIssue{
		Severity:          severity,
		Code:              code,
		Message:           message,
		StepKey:           stepKey,
		ProductCategoryID: productCategoryID,
	})
}

func (result QuickBuyValidationResult) errorSummary() string {
	for _, issue := range result.Issues {
		if issue.Severity == "error" {
			return issue.Message
		}
	}
	return "validation failed"
}

func (result *QuickBuySessionValidationResult) addIssue(code, message, stepKey string, productID uint, variantID *uint) {
	result.Valid = false
	result.Issues = append(result.Issues, QuickBuySessionValidationIssue{
		Severity:  "error",
		Code:      code,
		Message:   message,
		StepKey:   stepKey,
		ProductID: productID,
		VariantID: variantID,
	})
}

func quickBuySessionValidationStatus(result QuickBuySessionValidationResult) string {
	hasWarning := false
	for _, issue := range result.Issues {
		if issue.Severity == "error" {
			return quickbuy.ValidationStatusInvalid
		}
		if issue.Severity == "warning" {
			hasWarning = true
		}
	}
	if hasWarning {
		return quickbuy.ValidationStatusWarning
	}
	return quickbuy.ValidationStatusValid
}

func validateQuickBuySelectionBounds(version quickbuy.Version, items []quickbuy.SessionItem) error {
	itemsByStep := make(map[string][]quickbuy.SessionItem, len(items))
	for _, item := range items {
		itemsByStep[item.StepKey] = append(itemsByStep[item.StepKey], item)
	}
	for _, step := range version.Steps {
		stepItems := itemsByStep[step.StepKey]
		if step.SelectionMode == quickbuy.SelectionModeSingle && len(stepItems) > 1 {
			return fmt.Errorf("%w: step %q accepts only one selection", ErrQuickBuyInvalid, step.StepKey)
		}
		if step.MaxSelect > 0 && len(stepItems) > step.MaxSelect {
			return fmt.Errorf("%w: step %q exceeds max_select", ErrQuickBuyInvalid, step.StepKey)
		}
	}
	return nil
}

func (s *QuickBuyService) validateQuickBuyProductAllowedForStep(step quickbuy.Step, item productdomain.Product, enforceProductSpecificationTemplates bool) error {
	if err := s.validateQuickBuyProductCategoryAllowedForStep(step, item); err != nil {
		return err
	}
	if !enforceProductSpecificationTemplates {
		return nil
	}
	return validateQuickBuyProductSpecificationTemplateAllowedForStep(step, item)
}

func (s *QuickBuyService) validateQuickBuyProductCategoryAllowedForStep(step quickbuy.Step, item productdomain.Product) error {
	categoryIDs := quickBuyStepProductCategoryIDs(step)
	if len(categoryIDs) == 0 {
		return nil
	}
	if item.ProductCategoryID == nil {
		return fmt.Errorf("%w: product %d has no product category for step %q", ErrQuickBuyInvalid, item.ID, step.StepKey)
	}
	if s.productRepo == nil {
		return nil
	}
	allowed, err := s.productRepo.ProductCategoryInQuickBuyScope(item.ProductCategoryID, categoryIDs)
	if err != nil {
		return err
	}
	if allowed {
		return nil
	}
	return fmt.Errorf("%w: product %d is not in an allowed product category for step %q", ErrQuickBuyInvalid, item.ID, step.StepKey)
}

func validateQuickBuyProductSpecificationTemplateAllowedForStep(step quickbuy.Step, item productdomain.Product) error {
	if len(step.ProductSpecificationTemplates) == 0 {
		return nil
	}
	if item.ProductSpecificationTemplateID == nil {
		return fmt.Errorf("%w: product %d has no product specification template for step %q", ErrQuickBuyInvalid, item.ID, step.StepKey)
	}
	for _, productSpecificationTemplate := range step.ProductSpecificationTemplates {
		if productSpecificationTemplate.ProductSpecificationTemplateID == *item.ProductSpecificationTemplateID {
			return nil
		}
	}
	return fmt.Errorf("%w: product %d is not allowed for step %q", ErrQuickBuyInvalid, item.ID, step.StepKey)
}

func quickBuyStepByKey(version quickbuy.Version, stepKey string) *quickbuy.Step {
	for index := range version.Steps {
		if version.Steps[index].StepKey == stepKey {
			return &version.Steps[index]
		}
	}
	return nil
}

func quickBuyStepProductCategoryIDs(step quickbuy.Step) []uint {
	ids := make([]uint, 0, len(step.ProductCategories))
	for _, item := range step.ProductCategories {
		if item.ProductCategoryID == 0 {
			continue
		}
		ids = append(ids, item.ProductCategoryID)
	}
	return ids
}

func quickBuyStepProductSpecificationTemplateIDs(step quickbuy.Step) []uint {
	ids := make([]uint, 0, len(step.ProductSpecificationTemplates))
	for _, item := range step.ProductSpecificationTemplates {
		if item.ProductSpecificationTemplateID == 0 {
			continue
		}
		ids = append(ids, item.ProductSpecificationTemplateID)
	}
	return ids
}

func quickBuyStepProductSpecificationTemplateIDsForVersion(version quickbuy.Version, step quickbuy.Step) []uint {
	if isDefaultQuickBuyFlow(version) {
		return nil
	}
	return quickBuyStepProductSpecificationTemplateIDs(step)
}

func normalizeQuickBuyCandidatePaging(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = quickBuyCandidateDefaultPageSize
	}
	if pageSize > quickBuyCandidateMaxPageSize {
		pageSize = quickBuyCandidateMaxPageSize
	}
	return page, pageSize
}

func quickBuySessionTotals(items []quickbuy.SessionItem) (float64, int) {
	var subtotal float64
	var weightG int
	for _, item := range items {
		quantity := item.Quantity
		if quantity <= 0 {
			quantity = 1
		}
		subtotal += item.UnitPriceSnapshot * float64(quantity)
		weightG += item.WeightSnapshotG * quantity
	}
	return subtotal, weightG
}

func quickBuySessionView(session quickbuy.Session, validation *QuickBuySessionValidationResult, resolvers ...PublicMediaURLResolver) *QuickBuySessionView {
	resolver := quickBuyMediaResolver(resolvers)
	items := make([]QuickBuySessionItemView, 0, len(session.Items))
	for _, item := range session.Items {
		items = append(items, QuickBuySessionItemView{
			ID:                item.ID,
			StepKey:           item.StepKey,
			ProductID:         item.ProductID,
			VariantID:         item.VariantID,
			Quantity:          item.Quantity,
			UnitPriceSnapshot: item.UnitPriceSnapshot,
			CurrencySnapshot:  item.CurrencySnapshot,
			WeightSnapshotG:   item.WeightSnapshotG,
			ProductSnapshot:   quickBuyPublicProductSnapshot(item.ProductSnapshot, resolver),
			VariantSnapshot:   item.VariantSnapshot,
			SortOrder:         item.SortOrder,
		})
	}

	var flow *QuickBuyPublicFlowView
	if session.Version != nil && session.Version.Flow != nil {
		flow = quickBuyPublicFlowView(*session.Version, session.Locale, resolver)
	}
	return &QuickBuySessionView{
		SessionToken:     session.SessionToken,
		FlowID:           session.FlowID,
		FlowVersionID:    session.FlowVersionID,
		Locale:           session.Locale,
		MarketCountry:    session.MarketCountry,
		Currency:         session.Currency,
		Status:           session.Status,
		ValidationStatus: session.ValidationStatus,
		SubtotalSnapshot: session.SubtotalSnapshot,
		WeightSnapshotG:  session.WeightSnapshotG,
		ExpiresAt:        session.ExpiresAt,
		Flow:             flow,
		Items:            items,
		Validation:       validation,
		CreatedAt:        session.CreatedAt,
		UpdatedAt:        session.UpdatedAt,
	}
}

func quickBuyPublicProductSnapshot(raw datatypes.JSON, resolver PublicMediaURLResolver) datatypes.JSON {
	if resolver == nil || len(raw) == 0 {
		return raw
	}

	var snapshot map[string]interface{}
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return raw
	}
	for _, key := range []string{"thumbnail", "featured_image"} {
		if value, ok := snapshot[key].(string); ok {
			snapshot[key] = canonicalPublicMediaURL(resolver, value)
		}
	}
	return quickBuyJSON(snapshot)
}

func quickBuyProductSnapshot(item productdomain.Product, resolvers ...PublicMediaURLResolver) datatypes.JSON {
	thumbnail := quickBuyProductThumbnail(item, resolvers...)
	return quickBuyJSON(map[string]interface{}{
		"id":                                item.ID,
		"product_specification_template_id": item.ProductSpecificationTemplateID,
		"sku":                               item.SKU,
		"name":                              item.Name,
		"slug":                              item.Slug,
		"thumbnail":                         thumbnail,
		"featured_image":                    thumbnail,
		"currency":                          item.Currency,
		"price":                             item.Price,
		"sale_price":                        item.SalePrice,
		"status":                            item.Status,
	})
}

func quickBuyProductThumbnail(item productdomain.Product, resolvers ...PublicMediaURLResolver) string {
	resolver := quickBuyMediaResolver(resolvers)
	for _, media := range item.Media {
		if media.MediaType == "image" && media.IsVisible && media.IsPrimary {
			if strings.TrimSpace(media.ThumbnailURL) != "" {
				return canonicalPublicMediaURL(resolver, media.ThumbnailURL)
			}
			return canonicalPublicMediaURL(resolver, media.URL)
		}
	}
	for _, media := range item.Media {
		if media.MediaType == "image" && media.IsVisible {
			if strings.TrimSpace(media.ThumbnailURL) != "" {
				return canonicalPublicMediaURL(resolver, media.ThumbnailURL)
			}
			return canonicalPublicMediaURL(resolver, media.URL)
		}
	}
	return ""
}

func quickBuyVariantSnapshot(item productdomain.ProductVariant) datatypes.JSON {
	return quickBuyJSON(map[string]interface{}{
		"id":            item.ID,
		"product_id":    item.ProductID,
		"sku":           item.SKU,
		"title":         item.Title,
		"currency":      item.Currency,
		"price":         item.Price,
		"sale_price":    item.SalePrice,
		"stock":         item.Stock,
		"weight_grams":  item.Weight,
		"is_default":    item.IsDefault,
		"option_values": item.OptionValues,
	})
}

func quickBuyJSON(value interface{}) datatypes.JSON {
	encoded, err := json.Marshal(value)
	if err != nil {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON(encoded)
}

func quickBuyFlowSummary(flow quickbuy.Flow) QuickBuyFlowSummary {
	versions := make([]QuickBuyVersionSummary, 0, len(flow.Versions))
	for _, version := range flow.Versions {
		versions = append(versions, QuickBuyVersionSummary{
			ID:            version.ID,
			FlowID:        version.FlowID,
			VersionNumber: version.VersionNumber,
			Status:        version.Status,
			PublishedAt:   version.PublishedAt,
			StartsAt:      version.StartsAt,
			EndsAt:        version.EndsAt,
		})
	}
	return QuickBuyFlowSummary{
		ID:           flow.ID,
		Slug:         flow.Slug,
		Name:         flow.Name,
		Description:  flow.Description,
		HelpText:     flow.HelpText,
		EntrySurface: flow.EntrySurface,
		IsEnabled:    flow.IsEnabled,
		SortOrder:    flow.SortOrder,
		Versions:     versions,
		CreatedAt:    flow.CreatedAt,
		UpdatedAt:    flow.UpdatedAt,
	}
}

func quickBuyFlowView(version quickbuy.Version, locale string) *QuickBuyFlowView {
	if version.Flow == nil {
		return nil
	}
	requestLocale := locales.ResolveSupported(locale)
	steps := quickBuyStepViews(version.Steps, requestLocale, !isDefaultQuickBuyFlow(version))
	return &QuickBuyFlowView{
		ID:           version.Flow.ID,
		Slug:         version.Flow.Slug,
		Name:         version.Flow.Name,
		Description:  version.Flow.Description,
		HelpText:     quickBuyFlowHelpText(*version.Flow, requestLocale),
		Translations: quickBuyFlowTranslationViews(version.Flow.Translations),
		EntrySurface: version.Flow.EntrySurface,
		IsEnabled:    version.Flow.IsEnabled,
		SortOrder:    version.Flow.SortOrder,
		Version: QuickBuyVersionView{
			ID:            version.ID,
			VersionNumber: version.VersionNumber,
			Status:        version.Status,
			PublishedAt:   version.PublishedAt,
			StartsAt:      version.StartsAt,
			EndsAt:        version.EndsAt,
		},
		Steps: steps,
	}
}

func quickBuyPublicFlowView(version quickbuy.Version, locale string, resolvers ...PublicMediaURLResolver) *QuickBuyPublicFlowView {
	if version.Flow == nil {
		return nil
	}
	resolver := quickBuyMediaResolver(resolvers)
	requestLocale := locales.ResolveSupported(locale)
	return &QuickBuyPublicFlowView{
		ID:           version.Flow.ID,
		Slug:         version.Flow.Slug,
		Name:         version.Flow.Name,
		Description:  version.Flow.Description,
		HelpText:     quickBuyFlowHelpText(*version.Flow, requestLocale),
		EntrySurface: version.Flow.EntrySurface,
		IsEnabled:    version.Flow.IsEnabled,
		Version: QuickBuyVersionView{
			ID:            version.ID,
			VersionNumber: version.VersionNumber,
			Status:        version.Status,
			PublishedAt:   version.PublishedAt,
			StartsAt:      version.StartsAt,
			EndsAt:        version.EndsAt,
		},
		Steps: quickBuyStepViews(version.Steps, requestLocale, !isDefaultQuickBuyFlow(version), resolver),
	}
}

func quickBuyStepViews(steps []quickbuy.Step, locale string, exposeProductSpecificationTemplates bool, resolvers ...PublicMediaURLResolver) []QuickBuyStepView {
	resolver := quickBuyMediaResolver(resolvers)
	result := make([]QuickBuyStepView, 0, len(steps))
	for _, step := range steps {
		result = append(result, quickBuyStepView(step, locale, exposeProductSpecificationTemplates, resolver))
	}
	return result
}

func quickBuyFlowHelpText(flow quickbuy.Flow, locale string) string {
	helpText := strings.TrimSpace(flow.HelpText)
	if translation := quickBuyFlowTranslationForLocale(flow.Translations, locale); translation != nil && strings.TrimSpace(translation.HelpText) != "" {
		helpText = strings.TrimSpace(translation.HelpText)
	}
	return helpText
}

func quickBuyFlowTranslationForLocale(translations []quickbuy.FlowTranslation, locale string) *quickbuy.FlowTranslation {
	requestedLocale := locales.ResolveSupported(locale)
	if requestedLocale == "" {
		return nil
	}
	for index := range translations {
		if translations[index].Locale == requestedLocale {
			return &translations[index]
		}
	}
	return nil
}

func quickBuyFlowTranslationViews(translations []quickbuy.FlowTranslation) []QuickBuyFlowTranslationView {
	if len(translations) == 0 {
		return []QuickBuyFlowTranslationView{}
	}
	result := make([]QuickBuyFlowTranslationView, 0, len(translations))
	for _, translation := range translations {
		result = append(result, QuickBuyFlowTranslationView{
			ID:       translation.ID,
			Locale:   translation.Locale,
			HelpText: translation.HelpText,
		})
	}
	return result
}

func quickBuyStepView(step quickbuy.Step, locale string, exposeProductSpecificationTemplates bool, resolvers ...PublicMediaURLResolver) QuickBuyStepView {
	resolver := quickBuyMediaResolver(resolvers)
	productCategories := make([]QuickBuyProductCategoryView, 0, len(step.ProductCategories))
	productSpecificationTemplates := make([]QuickBuyProductSpecificationTemplateView, 0, len(step.ProductSpecificationTemplates))
	stepSlug := step.StepKey
	for index, item := range step.ProductCategories {
		if item.ProductCategory == nil {
			continue
		}
		if stepSlug == step.StepKey && index == 0 && !isDefaultQuickBuyStepKey(step.StepKey) {
			stepSlug = item.ProductCategory.Slug
		}
		productCategories = append(productCategories, quickBuyProductCategoryView(*item.ProductCategory, locale, item.IsPrimary, resolver))
	}
	if exposeProductSpecificationTemplates {
		for index, item := range step.ProductSpecificationTemplates {
			if item.ProductSpecificationTemplate == nil {
				continue
			}
			if index == 0 {
				stepSlug = item.ProductSpecificationTemplate.Slug
			}
			productSpecificationTemplates = append(productSpecificationTemplates, quickBuyProductSpecificationTemplateView(*item.ProductSpecificationTemplate, locale, item.IsPrimary, resolver))
		}
	}
	return QuickBuyStepView{
		ID:                            step.ID,
		StepKey:                       step.StepKey,
		Slug:                          stepSlug,
		Name:                          step.Name,
		SortOrder:                     step.SortOrder,
		ProductCategories:             productCategories,
		ProductSpecificationTemplates: productSpecificationTemplates,
		Filters:                       quickBuyStepFiltersForScope(step, nil, exposeProductSpecificationTemplates),
	}
}

func quickBuyStepFiltersForScope(step quickbuy.Step, valuesBySlug map[string][]string, exposeProductSpecificationTemplates bool) []QuickBuySpecFilterView {
	if !exposeProductSpecificationTemplates {
		return []QuickBuySpecFilterView{}
	}
	return quickBuyStepFilters(step, valuesBySlug)
}

func quickBuyFilterableSpecDefinitions(step quickbuy.Step) []productdomain.SpecDefinition {
	definitionsBySlug := make(map[string]productdomain.SpecDefinition)
	for _, item := range step.ProductSpecificationTemplates {
		if item.ProductSpecificationTemplate == nil {
			continue
		}
		for _, definition := range item.ProductSpecificationTemplate.SpecDefinitions {
			slug := strings.TrimSpace(definition.Slug)
			if slug == "" || !definition.IsVisible || !definition.IsFilterable {
				continue
			}
			if _, exists := definitionsBySlug[slug]; !exists {
				definitionsBySlug[slug] = definition
			}
		}
	}

	definitions := make([]productdomain.SpecDefinition, 0, len(definitionsBySlug))
	for _, definition := range definitionsBySlug {
		definitions = append(definitions, definition)
	}
	sort.SliceStable(definitions, func(i, j int) bool {
		if definitions[i].SortOrder == definitions[j].SortOrder {
			return definitions[i].Slug < definitions[j].Slug
		}
		return definitions[i].SortOrder < definitions[j].SortOrder
	})
	return definitions
}

func quickBuyFilterableSpecSlugs(step quickbuy.Step) []string {
	definitions := quickBuyFilterableSpecDefinitions(step)
	slugs := make([]string, 0, len(definitions))
	for _, definition := range definitions {
		slugs = append(slugs, definition.Slug)
	}
	return slugs
}

func normalizeQuickBuySpecFilters(step quickbuy.Step, input map[string][]string) (map[string][]string, error) {
	if len(input) == 0 {
		return nil, nil
	}

	allowed := make(map[string]struct{})
	for _, slug := range quickBuyFilterableSpecSlugs(step) {
		allowed[slug] = struct{}{}
	}

	result := make(map[string][]string)
	for rawSlug, rawValues := range input {
		slug := strings.TrimSpace(rawSlug)
		if slug == "" {
			continue
		}
		if _, exists := allowed[slug]; !exists {
			return nil, fmt.Errorf("%w: step %q does not expose filterable specification %q", ErrQuickBuyInvalid, step.StepKey, slug)
		}
		values := make([]string, 0, len(rawValues))
		seen := make(map[string]struct{}, len(rawValues))
		for _, rawValue := range rawValues {
			value := strings.TrimSpace(rawValue)
			if value == "" {
				continue
			}
			if _, exists := seen[value]; exists {
				continue
			}
			seen[value] = struct{}{}
			values = append(values, value)
			if len(values) >= 64 {
				break
			}
		}
		if len(values) > 0 {
			result[slug] = values
		}
	}
	if len(result) == 0 {
		return nil, nil
	}
	return result, nil
}

func quickBuyStepFilters(step quickbuy.Step, valuesBySlug map[string][]string) []QuickBuySpecFilterView {
	definitions := quickBuyFilterableSpecDefinitions(step)
	result := make([]QuickBuySpecFilterView, 0, len(definitions))
	for _, definition := range definitions {
		values := []string{}
		if valuesBySlug != nil && valuesBySlug[definition.Slug] != nil {
			values = append(values, valuesBySlug[definition.Slug]...)
		}
		result = append(result, QuickBuySpecFilterView{
			ID:              definition.ID,
			Name:            definition.Name,
			Slug:            definition.Slug,
			Unit:            definition.Unit,
			FieldType:       definition.FieldType,
			Presentation:    definition.Presentation,
			IsVariantOption: definition.IsVariantOption,
			Multiple:        true,
			Values:          values,
		})
	}
	return result
}

func quickBuyProductSpecificationTemplateView(item productdomain.ProductSpecificationTemplate, locale string, primary bool, resolvers ...PublicMediaURLResolver) QuickBuyProductSpecificationTemplateView {
	resolver := quickBuyMediaResolver(resolvers)
	return QuickBuyProductSpecificationTemplateView{
		ID:       item.ID,
		Slug:     item.Slug,
		Name:     item.NameForLocale(locale),
		ImageURL: canonicalPublicMediaURL(resolver, item.ImageURL),
		Primary:  primary,
	}
}

func quickBuyProductCategoryView(item productdomain.ProductCategory, locale string, primary bool, resolvers ...PublicMediaURLResolver) QuickBuyProductCategoryView {
	resolver := quickBuyMediaResolver(resolvers)
	name := strings.TrimSpace(item.Name)
	for _, translation := range item.Translations {
		if locales.Normalize(translation.Locale) != locale {
			continue
		}
		if translatedName := strings.TrimSpace(translation.Name); translatedName != "" {
			name = translatedName
		}
		break
	}
	return QuickBuyProductCategoryView{
		ID:       item.ID,
		ParentID: item.ParentID,
		Slug:     item.Slug,
		Name:     name,
		Depth:    item.Depth,
		ImageURL: canonicalPublicMediaURL(resolver, item.ImageURL),
		Primary:  primary,
	}
}

func quickBuyMediaResolver(resolvers []PublicMediaURLResolver) PublicMediaURLResolver {
	if len(resolvers) == 0 {
		return nil
	}
	return resolvers[0]
}

func normalizeQuickBuyKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "-")
	if !quickBuyKeyPattern.MatchString(value) {
		return ""
	}
	return value
}

func isDefaultQuickBuyFlowSlug(slug string) bool {
	return normalizeQuickBuyKey(slug) == quickBuyDefaultFlowSlug
}

func isDefaultQuickBuyStepKey(stepKey string) bool {
	normalized := normalizeQuickBuyKey(stepKey)
	for _, defaultStepKey := range quickBuyDefaultStepKeys {
		if normalized == defaultStepKey {
			return true
		}
	}
	return false
}

func isDefaultQuickBuyFlow(version quickbuy.Version) bool {
	return version.Flow != nil && isDefaultQuickBuyFlowSlug(version.Flow.Slug)
}

func normalizeDefaultQuickBuyFlow(flow *quickbuy.Flow) {
	if flow == nil || !isDefaultQuickBuyFlowSlug(flow.Slug) {
		return
	}
	flow.Slug = quickBuyDefaultFlowSlug
	flow.Name = quickBuyDefaultFlowName
	flow.Description = quickBuyDefaultFlowDescription
	flow.EntrySurface = quickBuyDefaultFlowEntrySurface
	flow.IsEnabled = true
	flow.SortOrder = quickBuyDefaultFlowSortOrder
}

func normalizeDefaultQuickBuySteps(steps []quickbuy.Step) error {
	stepByKey := make(map[string]quickbuy.Step, len(steps))
	for _, step := range steps {
		stepByKey[normalizeQuickBuyKey(step.StepKey)] = step
	}

	ordered := make([]quickbuy.Step, 0, len(steps))
	for _, stepKey := range quickBuyDefaultStepKeys {
		step, exists := stepByKey[stepKey]
		if !exists {
			return fmt.Errorf("%w: default quick-build step %q cannot be removed", ErrQuickBuyInvalid, stepKey)
		}
		ordered = append(ordered, step)
		delete(stepByKey, stepKey)
	}

	extraSteps := make([]quickbuy.Step, 0, len(stepByKey))
	for _, step := range stepByKey {
		extraSteps = append(extraSteps, step)
	}
	sort.SliceStable(extraSteps, func(i, j int) bool {
		if extraSteps[i].SortOrder == extraSteps[j].SortOrder {
			return extraSteps[i].StepKey < extraSteps[j].StepKey
		}
		return extraSteps[i].SortOrder < extraSteps[j].SortOrder
	})
	ordered = append(ordered, extraSteps...)

	for index := range ordered {
		step := &ordered[index]
		step.SortOrder = (index + 1) * 10
		step.SelectionMode = quickbuy.SelectionModeSingle
		step.IsRequired = true
		step.MinSelect = 0
		step.MaxSelect = 1
		step.DefaultQuantity = 1
		step.AllowSkip = false
		step.ProductSpecificationTemplates = nil
		if index < len(quickBuyDefaultStepKeys) {
			step.StepKey = quickBuyDefaultStepKeys[index]
		}
	}

	copy(steps, ordered)
	return nil
}

func validateQuickBuyDefaultSteps(flowSlug string, steps []quickbuy.Step) error {
	if !isDefaultQuickBuyFlowSlug(flowSlug) {
		return nil
	}
	stepKeys := make(map[string]struct{}, len(steps))
	for _, step := range steps {
		stepKeys[normalizeQuickBuyKey(step.StepKey)] = struct{}{}
	}
	for _, requiredStepKey := range quickBuyDefaultStepKeys {
		if _, exists := stepKeys[requiredStepKey]; !exists {
			return fmt.Errorf("%w: default quick-build step %q cannot be removed", ErrQuickBuyInvalid, requiredStepKey)
		}
	}
	return nil
}

func normalizeQuickBuySurface(value string) string {
	value = normalizeQuickBuyKey(value)
	if value == "" {
		return "dock"
	}
	return value
}

func normalizeQuickBuySelectionMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case quickbuy.SelectionModeMultiple:
		return quickbuy.SelectionModeMultiple
	case quickbuy.SelectionModeQuantity:
		return quickbuy.SelectionModeQuantity
	case quickbuy.SelectionModeAuto:
		return quickbuy.SelectionModeAuto
	default:
		return quickbuy.SelectionModeSingle
	}
}

func quickBuySelectionModeIsValid(value string) bool {
	switch value {
	case quickbuy.SelectionModeSingle, quickbuy.SelectionModeMultiple, quickbuy.SelectionModeQuantity, quickbuy.SelectionModeAuto:
		return true
	default:
		return false
	}
}

func normalizeQuickBuyCountry(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	var b strings.Builder
	for _, r := range value {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func quickBuyVersionIsActive(version quickbuy.Version, now time.Time) bool {
	if version.Status != quickbuy.FlowVersionStatusPublished || version.Flow == nil || !version.Flow.IsEnabled {
		return false
	}
	if version.StartsAt != nil && version.StartsAt.After(now) {
		return false
	}
	if version.EndsAt != nil && !version.EndsAt.After(now) {
		return false
	}
	return true
}

func normalizeQuickBuyCurrency(value string) string {
	currency := currencydomain.NormalizeCode(value)
	if currency == "" || !currencydomain.IsCatalogCode(currency) {
		return productdomain.DefaultPriceCurrency
	}
	return currency
}

func generateQuickBuySessionToken() string {
	return "qb_" + strings.ReplaceAll(uuid.NewString(), "-", "")
}
