package service

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"commerce-platform/internal/domain/quickbuy"
	"commerce-platform/internal/pkg/locales"
	"commerce-platform/internal/repository"

	"gorm.io/datatypes"
)

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
