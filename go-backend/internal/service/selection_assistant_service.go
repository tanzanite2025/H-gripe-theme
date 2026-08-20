package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	selectionassistant "commerce-platform/internal/domain/selectionassistant"
	"commerce-platform/internal/repository"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	ErrSelectionAssistantInvalid      = errors.New("invalid selection assistant")
	ErrSelectionAssistantNotFound     = errors.New("selection assistant not found")
	ErrSelectionAssistantNotPublished = errors.New("selection assistant has no published version")
	ErrSelectionAssistantNotMutable   = errors.New("selection assistant version is not mutable")
	ErrSelectionAssistantVersionFound = errors.New("selection assistant version not found")
)

var selectionAssistantKeyPattern = regexp.MustCompile(`^[a-z0-9]+(?:[_-][a-z0-9]+)*$`)

const (
	selectionAssistantMaxNodes        = 200
	selectionAssistantMaxOptions      = 32
	selectionAssistantMaxFilterKeys   = 32
	selectionAssistantMaxFilterValues = 64
)

type SelectionAssistantService struct {
	repo *repository.SelectionAssistantRepository
}

type SelectionAssistantFlowInput struct {
	Slug                string                         `json:"slug"`
	Name                string                         `json:"name"`
	Description         string                         `json:"description"`
	ProductCategorySlug string                         `json:"product_category_slug"`
	IsEnabled           *bool                          `json:"is_enabled"`
	SortOrder           int                            `json:"sort_order"`
	Version             SelectionAssistantVersionInput `json:"version"`
}

type SelectionAssistantVersionInput struct {
	Config selectionassistant.Config `json:"config"`
}

type SelectionAssistantFlowSummary struct {
	ID                  uint                               `json:"id"`
	Slug                string                             `json:"slug"`
	Name                string                             `json:"name"`
	Description         string                             `json:"description"`
	ProductCategorySlug string                             `json:"product_category_slug"`
	IsEnabled           bool                               `json:"is_enabled"`
	SortOrder           int                                `json:"sort_order"`
	Versions            []SelectionAssistantVersionSummary `json:"versions,omitempty"`
	CreatedAt           time.Time                          `json:"created_at"`
	UpdatedAt           time.Time                          `json:"updated_at"`
}

type SelectionAssistantVersionSummary struct {
	ID            uint       `json:"id"`
	FlowID        uint       `json:"flow_id"`
	VersionNumber int        `json:"version_number"`
	Status        string     `json:"status"`
	PublishedAt   *time.Time `json:"published_at,omitempty"`
}

type SelectionAssistantFlowView struct {
	ID                  uint                          `json:"id"`
	Slug                string                        `json:"slug"`
	Name                string                        `json:"name"`
	Description         string                        `json:"description"`
	ProductCategorySlug string                        `json:"product_category_slug"`
	IsEnabled           bool                          `json:"is_enabled"`
	SortOrder           int                           `json:"sort_order"`
	Version             SelectionAssistantVersionView `json:"version"`
}

type SelectionAssistantVersionView struct {
	ID            uint                      `json:"id"`
	VersionNumber int                       `json:"version_number"`
	Status        string                    `json:"status"`
	Config        selectionassistant.Config `json:"config"`
	PublishedAt   *time.Time                `json:"published_at,omitempty"`
}

type SelectionAssistantValidationResult struct {
	Valid  bool                                `json:"valid"`
	Issues []SelectionAssistantValidationIssue `json:"issues"`
}

type SelectionAssistantValidationIssue struct {
	Severity  string `json:"severity"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	NodeKey   string `json:"node_key,omitempty"`
	OptionKey string `json:"option_key,omitempty"`
}

type selectionAssistantDraft struct {
	Flow    *selectionassistant.Flow
	Version *selectionassistant.Version
	Config  selectionassistant.Config
}

func NewSelectionAssistantService(repo *repository.SelectionAssistantRepository) *SelectionAssistantService {
	return &SelectionAssistantService{repo: repo}
}

func (s *SelectionAssistantService) ListFlows() ([]SelectionAssistantFlowSummary, error) {
	flows, err := s.repo.ListFlows()
	if err != nil {
		return nil, err
	}
	result := make([]SelectionAssistantFlowSummary, 0, len(flows))
	for _, flow := range flows {
		if isReservedSelectionAssistantSlug(flow.Slug) {
			continue
		}
		result = append(result, selectionAssistantFlowSummary(flow))
	}
	return result, nil
}

func (s *SelectionAssistantService) GetFlow(id uint) (*SelectionAssistantFlowView, error) {
	flow, err := s.repo.FindFlowByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSelectionAssistantNotFound
	}
	if err != nil {
		return nil, err
	}
	if isReservedSelectionAssistantSlug(flow.Slug) {
		return nil, ErrSelectionAssistantNotFound
	}
	version := latestSelectionAssistantVersion(flow.Versions)
	if version == nil {
		return nil, fmt.Errorf("%w: no version exists", ErrSelectionAssistantInvalid)
	}
	view, err := selectionAssistantVersionView(flow, version)
	if err != nil {
		return nil, err
	}
	return &view, nil
}

func (s *SelectionAssistantService) GetPublishedFlowBySlug(slug string) (*SelectionAssistantFlowView, error) {
	slug = normalizeSelectionAssistantKey(slug)
	if slug == "" {
		return nil, ErrSelectionAssistantNotFound
	}
	if isReservedSelectionAssistantSlug(slug) {
		return nil, ErrSelectionAssistantNotFound
	}

	flow, err := s.repo.FindEnabledPublishedFlowBySlug(slug)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSelectionAssistantNotFound
	}
	if err != nil {
		return nil, err
	}

	version := latestSelectionAssistantVersion(flow.Versions)
	if version == nil || version.Status != selectionassistant.FlowVersionStatusPublished {
		return nil, ErrSelectionAssistantNotPublished
	}

	view, err := selectionAssistantVersionView(flow, version)
	if err != nil {
		return nil, err
	}
	return &view, nil
}

func (s *SelectionAssistantService) CreateFlow(input SelectionAssistantFlowInput) (*SelectionAssistantFlowView, error) {
	flow, version, err := s.normalizeFlowInput(input)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateFlowWithVersion(flow, version); err != nil {
		return nil, err
	}
	return s.GetFlow(flow.ID)
}

func (s *SelectionAssistantService) SaveFlowConfiguration(id uint, input SelectionAssistantFlowInput) (*SelectionAssistantFlowView, error) {
	flow, err := s.repo.FindFlowByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSelectionAssistantNotFound
	}
	if err != nil {
		return nil, err
	}
	if isReservedSelectionAssistantSlug(flow.Slug) {
		return nil, ErrSelectionAssistantNotFound
	}
	normalizedFlow, normalizedVersion, err := s.normalizeFlowInput(input)
	if err != nil {
		return nil, err
	}
	normalizedFlow.ID = flow.ID
	normalizedFlow.CreatedAt = flow.CreatedAt

	draft, draftErr := s.repo.FindDraftVersion(flow.ID)
	if draftErr != nil && !repository.IsRecordNotFound(draftErr) {
		return nil, draftErr
	}

	if draft == nil {
		latestNumber, err := s.repo.FindLatestVersionNumber(flow.ID)
		if err != nil {
			return nil, err
		}
		normalizedVersion.FlowID = flow.ID
		normalizedVersion.VersionNumber = latestNumber + 1
		if err := s.repo.UpdateFlow(normalizedFlow); err != nil {
			return nil, err
		}
		if err := s.repo.CreateVersion(normalizedVersion); err != nil {
			return nil, err
		}
	} else {
		normalizedVersion.ID = draft.ID
		normalizedVersion.FlowID = flow.ID
		normalizedVersion.VersionNumber = draft.VersionNumber
		if err := s.repo.UpdateFlow(normalizedFlow); err != nil {
			return nil, err
		}
		if err := s.repo.UpdateDraftVersion(normalizedVersion); err != nil {
			return nil, err
		}
	}

	return s.GetFlow(flow.ID)
}

func (s *SelectionAssistantService) ValidateVersion(id uint) (*SelectionAssistantValidationResult, error) {
	version, err := s.repo.FindVersionByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSelectionAssistantVersionFound
	}
	if err != nil {
		return nil, err
	}
	config := versionConfig(*version)
	_, result := ValidateConfig(config)
	return &result, nil
}

func (s *SelectionAssistantService) PublishVersion(id uint, publishedBy *uint) (*SelectionAssistantFlowView, error) {
	version, err := s.repo.FindVersionByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSelectionAssistantVersionFound
	}
	if err != nil {
		return nil, err
	}
	if version.Flow != nil && isReservedSelectionAssistantSlug(version.Flow.Slug) {
		return nil, ErrSelectionAssistantNotFound
	}
	if version.Status != selectionassistant.FlowVersionStatusDraft {
		return nil, ErrSelectionAssistantNotMutable
	}
	config := versionConfig(*version)
	_, result := ValidateConfig(config)
	if !result.Valid {
		return nil, fmt.Errorf("%w: %s", ErrSelectionAssistantInvalid, result.Issues[0].Message)
	}
	if err := s.repo.PublishVersion(id, publishedBy, time.Now().UTC()); err != nil {
		return nil, err
	}
	return s.GetFlow(version.FlowID)
}

func (s *SelectionAssistantService) normalizeFlowInput(input SelectionAssistantFlowInput) (*selectionassistant.Flow, *selectionassistant.Version, error) {
	slug := normalizeSelectionAssistantKey(input.Slug)
	if slug == "" {
		return nil, nil, fmt.Errorf("%w: slug is required and must use lowercase letters, numbers, dashes, or underscores", ErrSelectionAssistantInvalid)
	}
	if isReservedSelectionAssistantSlug(slug) {
		return nil, nil, fmt.Errorf("%w: slug %q is managed by the wheelset fit questionnaire", ErrSelectionAssistantInvalid, slug)
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, nil, fmt.Errorf("%w: name is required", ErrSelectionAssistantInvalid)
	}
	categorySlug := strings.ToLower(strings.TrimSpace(input.ProductCategorySlug))
	if categorySlug == "" {
		categorySlug = selectionassistant.WheelsetProductCategorySlug
	}
	if categorySlug != selectionassistant.WheelsetProductCategorySlug {
		return nil, nil, fmt.Errorf("%w: product category must be %q", ErrSelectionAssistantInvalid, selectionassistant.WheelsetProductCategorySlug)
	}
	isEnabled := true
	if input.IsEnabled != nil {
		isEnabled = *input.IsEnabled
	}
	config := normalizeConfig(input.Version.Config)
	rawConfig, err := json.Marshal(config)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: encode config: %v", ErrSelectionAssistantInvalid, err)
	}
	return &selectionassistant.Flow{
			Slug:                slug,
			Name:                name,
			Description:         strings.TrimSpace(input.Description),
			ProductCategorySlug: categorySlug,
			IsEnabled:           isEnabled,
			SortOrder:           input.SortOrder,
		},
		&selectionassistant.Version{
			VersionNumber: 1,
			Status:        selectionassistant.FlowVersionStatusDraft,
			Config:        datatypes.JSON(rawConfig),
		},
		nil
}

func normalizeConfig(config selectionassistant.Config) selectionassistant.Config {
	if strings.TrimSpace(config.Kind) == "" {
		config.Kind = selectionassistant.ConfigKind
	}
	if config.SchemaVersion <= 0 {
		config.SchemaVersion = 1
	}
	if strings.TrimSpace(config.BaseProductQuery.CategorySlug) == "" {
		config.BaseProductQuery.CategorySlug = selectionassistant.WheelsetProductCategorySlug
	}
	if config.Nodes == nil {
		config.Nodes = []selectionassistant.Node{}
	}
	return config
}

func versionConfig(version selectionassistant.Version) selectionassistant.Config {
	var config selectionassistant.Config
	if err := json.Unmarshal(version.Config, &config); err != nil {
		return selectionassistant.Config{}
	}
	return normalizeConfig(config)
}

func selectionAssistantFlowSummary(flow selectionassistant.Flow) SelectionAssistantFlowSummary {
	versions := make([]SelectionAssistantVersionSummary, 0, len(flow.Versions))
	for _, version := range flow.Versions {
		versions = append(versions, SelectionAssistantVersionSummary{
			ID:            version.ID,
			FlowID:        version.FlowID,
			VersionNumber: version.VersionNumber,
			Status:        version.Status,
			PublishedAt:   version.PublishedAt,
		})
	}
	return SelectionAssistantFlowSummary{
		ID:                  flow.ID,
		Slug:                flow.Slug,
		Name:                flow.Name,
		Description:         flow.Description,
		ProductCategorySlug: flow.ProductCategorySlug,
		IsEnabled:           flow.IsEnabled,
		SortOrder:           flow.SortOrder,
		Versions:            versions,
		CreatedAt:           flow.CreatedAt,
		UpdatedAt:           flow.UpdatedAt,
	}
}

func selectionAssistantVersionView(flow *selectionassistant.Flow, version *selectionassistant.Version) (SelectionAssistantFlowView, error) {
	return SelectionAssistantFlowView{
		ID:                  flow.ID,
		Slug:                flow.Slug,
		Name:                flow.Name,
		Description:         flow.Description,
		ProductCategorySlug: flow.ProductCategorySlug,
		IsEnabled:           flow.IsEnabled,
		SortOrder:           flow.SortOrder,
		Version: SelectionAssistantVersionView{
			ID:            version.ID,
			VersionNumber: version.VersionNumber,
			Status:        version.Status,
			Config:        versionConfig(*version),
			PublishedAt:   version.PublishedAt,
		},
	}, nil
}

func latestSelectionAssistantVersion(versions []selectionassistant.Version) *selectionassistant.Version {
	if len(versions) == 0 {
		return nil
	}
	sort.SliceStable(versions, func(i, j int) bool {
		if versions[i].VersionNumber == versions[j].VersionNumber {
			return versions[i].ID > versions[j].ID
		}
		return versions[i].VersionNumber > versions[j].VersionNumber
	})
	return &versions[0]
}

func normalizeSelectionAssistantKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if !selectionAssistantKeyPattern.MatchString(value) {
		return ""
	}
	return value
}

func isReservedSelectionAssistantSlug(slug string) bool {
	return normalizeSelectionAssistantKey(slug) == wheelsetFitAssistantSlug
}

func ValidateConfig(config selectionassistant.Config) (selectionassistant.Config, SelectionAssistantValidationResult) {
	config = normalizeConfig(config)
	result := SelectionAssistantValidationResult{
		Valid:  true,
		Issues: []SelectionAssistantValidationIssue{},
	}
	addIssue := func(issue SelectionAssistantValidationIssue) {
		result.Issues = append(result.Issues, issue)
		if issue.Severity == "error" {
			result.Valid = false
		}
	}

	if config.Kind != selectionassistant.ConfigKind {
		addIssue(SelectionAssistantValidationIssue{
			Severity: "error",
			Code:     "invalid_kind",
			Message:  fmt.Sprintf("config kind must be %q", selectionassistant.ConfigKind),
		})
	}
	if config.SchemaVersion != 1 {
		addIssue(SelectionAssistantValidationIssue{
			Severity: "error",
			Code:     "unsupported_schema_version",
			Message:  "only schema version 1 is currently supported",
		})
	}
	if config.BaseProductQuery.CategorySlug != selectionassistant.WheelsetProductCategorySlug {
		addIssue(SelectionAssistantValidationIssue{
			Severity: "error",
			Code:     "invalid_base_category",
			Message:  fmt.Sprintf("base product category must be %q", selectionassistant.WheelsetProductCategorySlug),
		})
	}
	if len(config.Nodes) == 0 {
		addIssue(SelectionAssistantValidationIssue{
			Severity: "error",
			Code:     "missing_nodes",
			Message:  "at least one question or terminal node is required",
		})
		return config, result
	}
	if len(config.Nodes) > selectionAssistantMaxNodes {
		addIssue(SelectionAssistantValidationIssue{
			Severity: "error",
			Code:     "too_many_nodes",
			Message:  fmt.Sprintf("a flow can contain at most %d nodes", selectionAssistantMaxNodes),
		})
	}

	nodesByKey := make(map[string]selectionassistant.Node, len(config.Nodes))
	for _, node := range config.Nodes {
		key := normalizeSelectionAssistantKey(node.Key)
		if key == "" {
			addIssue(SelectionAssistantValidationIssue{
				Severity: "error",
				Code:     "invalid_node_key",
				Message:  "node keys must use lowercase letters, numbers, dashes, or underscores",
				NodeKey:  node.Key,
			})
			continue
		}
		if _, exists := nodesByKey[key]; exists {
			addIssue(SelectionAssistantValidationIssue{
				Severity: "error",
				Code:     "duplicate_node_key",
				Message:  fmt.Sprintf("node key %q is duplicated", key),
				NodeKey:  key,
			})
			continue
		}
		if node.Type != selectionassistant.NodeTypeQuestion &&
			node.Type != selectionassistant.NodeTypeTerminal &&
			node.Type != selectionassistant.NodeTypeSupport {
			addIssue(SelectionAssistantValidationIssue{
				Severity: "error",
				Code:     "invalid_node_type",
				Message:  fmt.Sprintf("node %q has an unsupported type %q", key, node.Type),
				NodeKey:  key,
			})
		}
		if len(node.Options) > selectionAssistantMaxOptions {
			addIssue(SelectionAssistantValidationIssue{
				Severity: "error",
				Code:     "too_many_options",
				Message:  fmt.Sprintf("node %q can contain at most %d options", key, selectionAssistantMaxOptions),
				NodeKey:  key,
			})
		}
		optionKeys := make(map[string]struct{}, len(node.Options))
		for _, option := range node.Options {
			optionKey := normalizeSelectionAssistantKey(option.Key)
			if optionKey == "" {
				addIssue(SelectionAssistantValidationIssue{
					Severity:  "error",
					Code:      "invalid_option_key",
					Message:   fmt.Sprintf("node %q contains an invalid option key", key),
					NodeKey:   key,
					OptionKey: option.Key,
				})
				continue
			}
			if _, exists := optionKeys[optionKey]; exists {
				addIssue(SelectionAssistantValidationIssue{
					Severity:  "error",
					Code:      "duplicate_option_key",
					Message:   fmt.Sprintf("option key %q is duplicated in node %q", optionKey, key),
					NodeKey:   key,
					OptionKey: optionKey,
				})
			}
			optionKeys[optionKey] = struct{}{}
			if next := strings.TrimSpace(option.NextNodeKey); next != "" && normalizeSelectionAssistantKey(next) == "" {
				addIssue(SelectionAssistantValidationIssue{
					Severity:  "error",
					Code:      "invalid_next_node_key",
					Message:   fmt.Sprintf("option %q in node %q points to an invalid node key", optionKey, key),
					NodeKey:   key,
					OptionKey: optionKey,
				})
			}
			validateQueryEffects(addIssue, key, optionKey, option.QueryEffects)
		}
		if node.Type != selectionassistant.NodeTypeQuestion && len(node.Options) > 0 {
			addIssue(SelectionAssistantValidationIssue{
				Severity: "error",
				Code:     "non_question_options",
				Message:  fmt.Sprintf("node %q is not a question and cannot contain options", key),
				NodeKey:  key,
			})
		}
		nodesByKey[key] = node
	}

	entryKey := normalizeSelectionAssistantKey(config.EntryNodeKey)
	if entryKey == "" {
		addIssue(SelectionAssistantValidationIssue{
			Severity: "error",
			Code:     "missing_entry_node",
			Message:  "entry_node_key is required",
		})
	} else if _, exists := nodesByKey[entryKey]; !exists {
		addIssue(SelectionAssistantValidationIssue{
			Severity: "error",
			Code:     "entry_node_not_found",
			Message:  fmt.Sprintf("entry node %q does not exist", entryKey),
			NodeKey:  entryKey,
		})
	}

	adjacency := make(map[string][]string, len(nodesByKey))
	for key, node := range nodesByKey {
		for _, option := range node.Options {
			nextKey := normalizeSelectionAssistantKey(option.NextNodeKey)
			if nextKey == "" {
				continue
			}
			if _, exists := nodesByKey[nextKey]; !exists {
				addIssue(SelectionAssistantValidationIssue{
					Severity:  "error",
					Code:      "next_node_not_found",
					Message:   fmt.Sprintf("option %q in node %q points to missing node %q", option.Key, key, nextKey),
					NodeKey:   key,
					OptionKey: option.Key,
				})
				continue
			}
			adjacency[key] = append(adjacency[key], nextKey)
		}
	}

	if entryKey != "" {
		visited := make(map[string]bool, len(nodesByKey))
		active := make(map[string]bool, len(nodesByKey))
		var walk func(string)
		walk = func(key string) {
			if active[key] {
				addIssue(SelectionAssistantValidationIssue{
					Severity: "error",
					Code:     "cycle_detected",
					Message:  fmt.Sprintf("cycle detected at node %q", key),
					NodeKey:  key,
				})
				return
			}
			if visited[key] {
				return
			}
			visited[key] = true
			active[key] = true
			for _, nextKey := range adjacency[key] {
				walk(nextKey)
			}
			active[key] = false
		}
		if _, exists := nodesByKey[entryKey]; exists {
			walk(entryKey)
			for key := range nodesByKey {
				if !visited[key] {
					addIssue(SelectionAssistantValidationIssue{
						Severity: "error",
						Code:     "unreachable_node",
						Message:  fmt.Sprintf("node %q cannot be reached from the entry node", key),
						NodeKey:  key,
					})
				}
			}
		}
	}

	return config, result
}

func validateQueryEffects(addIssue func(SelectionAssistantValidationIssue), nodeKey, optionKey string, effects selectionassistant.QueryEffects) {
	if len(effects.SpecFilters) > selectionAssistantMaxFilterKeys {
		addIssue(SelectionAssistantValidationIssue{
			Severity:  "error",
			Code:      "too_many_filter_keys",
			Message:   fmt.Sprintf("option %q in node %q contains too many specification filters", optionKey, nodeKey),
			NodeKey:   nodeKey,
			OptionKey: optionKey,
		})
	}
	for slug, values := range effects.SpecFilters {
		if normalizeSelectionAssistantKey(slug) == "" {
			addIssue(SelectionAssistantValidationIssue{
				Severity:  "error",
				Code:      "invalid_filter_key",
				Message:   fmt.Sprintf("option %q in node %q contains invalid filter key %q", optionKey, nodeKey, slug),
				NodeKey:   nodeKey,
				OptionKey: optionKey,
			})
		}
		if len(values) > selectionAssistantMaxFilterValues {
			addIssue(SelectionAssistantValidationIssue{
				Severity:  "error",
				Code:      "too_many_filter_values",
				Message:   fmt.Sprintf("filter %q in option %q contains too many values", slug, optionKey),
				NodeKey:   nodeKey,
				OptionKey: optionKey,
			})
		}
		for _, value := range values {
			if strings.TrimSpace(value) == "" {
				addIssue(SelectionAssistantValidationIssue{
					Severity:  "error",
					Code:      "empty_filter_value",
					Message:   fmt.Sprintf("filter %q in option %q contains an empty value", slug, optionKey),
					NodeKey:   nodeKey,
					OptionKey: optionKey,
				})
				break
			}
		}
	}
}
