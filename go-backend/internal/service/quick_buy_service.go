package service

import (
	"errors"
	"regexp"
	"time"

	productdomain "commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"

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
	ID      uint   `json:"id"`
	Slug    string `json:"slug"`
	Name    string `json:"name"`
	Primary bool   `json:"primary"`
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
