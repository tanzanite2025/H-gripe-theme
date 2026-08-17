package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	selectionassistant "commerce-platform/internal/domain/selectionassistant"
	wheelsetfit "commerce-platform/internal/domain/wheelsetfit"

	"gorm.io/gorm"
)

const (
	wheelsetFitAssistantSlug        = "wheelset-fit-helper"
	wheelsetFitAssistantName        = "Wheelset fit helper"
	wheelsetFitAssistantDescription = "Fixed ordered wheelset fit questionnaire"
	wheelsetFitSummaryNodeKey       = "summary"
)

type WheelsetFitQuestionnaireFlowView struct {
	ID                  uint                                    `json:"id"`
	Slug                string                                  `json:"slug"`
	Name                string                                  `json:"name"`
	Description         string                                  `json:"description"`
	ProductCategorySlug string                                  `json:"product_category_slug"`
	IsEnabled           bool                                    `json:"is_enabled"`
	SortOrder           int                                     `json:"sort_order"`
	Version             WheelsetFitQuestionnaireFlowVersionView `json:"version"`
}

type WheelsetFitQuestionnaireFlowVersionView struct {
	ID            uint                      `json:"id"`
	VersionNumber int                       `json:"version_number"`
	Status        string                    `json:"status"`
	Config        selectionassistant.Config `json:"config"`
	PublishedAt   *time.Time                `json:"published_at,omitempty"`
}

func (s *WheelsetFitQuestionnaireService) GetPublishedFlow() (*WheelsetFitQuestionnaireFlowView, error) {
	questionnaire, err := s.getQuestionnaire()
	if err != nil {
		return nil, err
	}
	version, err := s.repo.FindPublishedVersion(questionnaire.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrWheelsetFitQuestionnaireVersionNotFound
	}
	if err != nil {
		return nil, err
	}
	flow, err := buildWheelsetFitQuestionnaireFlow(questionnaire, version)
	if err != nil {
		return nil, err
	}
	return flow, nil
}

func buildWheelsetFitQuestionnaireFlow(questionnaire *wheelsetfit.Questionnaire, version *wheelsetfit.Version) (*WheelsetFitQuestionnaireFlowView, error) {
	if questionnaire == nil || version == nil {
		return nil, ErrWheelsetFitQuestionnaireInvalid
	}

	config, err := buildWheelsetFitQuestionnaireConfig(questionnaire, version)
	if err != nil {
		return nil, err
	}

	return &WheelsetFitQuestionnaireFlowView{
		ID:                  questionnaire.ID,
		Slug:                wheelsetFitAssistantSlug,
		Name:                wheelsetFitAssistantName,
		Description:         wheelsetFitAssistantDescription,
		ProductCategorySlug: questionnaire.ProductCategorySlug,
		IsEnabled:           questionnaire.IsEnabled,
		SortOrder:           100,
		Version: WheelsetFitQuestionnaireFlowVersionView{
			ID:            version.ID,
			VersionNumber: version.VersionNumber,
			Status:        version.Status,
			Config:        config,
			PublishedAt:   version.PublishedAt,
		},
	}, nil
}

func buildWheelsetFitQuestionnaireConfig(questionnaire *wheelsetfit.Questionnaire, version *wheelsetfit.Version) (selectionassistant.Config, error) {
	sourceLocale := wheelsetfit.SourceLocale
	if questionnaire != nil && strings.TrimSpace(questionnaire.SourceLocale) != "" {
		sourceLocale = questionnaire.SourceLocale
	}
	sourceLocale = strings.TrimSpace(sourceLocale)
	questions := make([]wheelsetfit.Question, 0, len(version.Questions))
	for _, question := range version.Questions {
		if !question.IsEnabled {
			continue
		}
		questions = append(questions, question)
	}
	if len(questions) == 0 {
		return selectionassistant.Config{}, fmt.Errorf("%w: no enabled questions available", ErrWheelsetFitQuestionnaireInvalid)
	}

	nodes := make([]selectionassistant.Node, 0, len(questions)+1)
	for index, question := range questions {
		nextNodeKey := wheelsetFitSummaryNodeKey
		if index+1 < len(questions) {
			nextNodeKey = questions[index+1].QuestionKey
		}
		nodes = append(nodes, selectionassistant.Node{
			Key:     question.QuestionKey,
			Type:    selectionassistant.NodeTypeQuestion,
			Prompt:  localizedQuestionPromptMap(question, sourceLocale),
			Helper:  localizedQuestionHelperMap(question, sourceLocale),
			Options: localizedQuestionOptions(question, sourceLocale, nextNodeKey),
			Editor: selectionassistant.EditorPosition{
				X: index * 320,
				Y: 0,
			},
		})
	}

	nodes = append(nodes, selectionassistant.Node{
		Key:    wheelsetFitSummaryNodeKey,
		Type:   selectionassistant.NodeTypeTerminal,
		Prompt: localizedSummaryPromptMap(sourceLocale),
		Helper: localizedSummaryHelperMap(sourceLocale),
		Editor: selectionassistant.EditorPosition{
			X: len(questions) * 320,
			Y: 0,
		},
	})

	return selectionassistant.Config{
		Kind:          selectionassistant.ConfigKind,
		SchemaVersion: 1,
		EntryNodeKey:  questions[0].QuestionKey,
		BaseProductQuery: selectionassistant.BaseProductQuery{
			CategorySlug: questionnaire.ProductCategorySlug,
		},
		Nodes: nodes,
	}, nil
}

func localizedQuestionPromptMap(question wheelsetfit.Question, sourceLocale string) map[string]string {
	translations := wheelsetFitQuestionTranslationsByLocale(&question)
	result := make(map[string]string, len(translations)+2)
	for _, translation := range translations {
		if text := strings.TrimSpace(translation.Prompt); text != "" {
			result[translation.Locale] = text
		}
	}
	seed := strings.TrimSpace(translations[sourceLocale].Prompt)
	if seed == "" {
		seed = strings.TrimSpace(translations[wheelsetfit.SourceLocale].Prompt)
	}
	if seed != "" {
		if result["en"] == "" {
			result["en"] = seed
		}
		if result["zh_cn"] == "" {
			result["zh_cn"] = seed
		}
	}
	return result
}

func localizedQuestionHelperMap(question wheelsetfit.Question, sourceLocale string) map[string]string {
	translations := wheelsetFitQuestionTranslationsByLocale(&question)
	result := make(map[string]string, len(translations)+2)
	for _, translation := range translations {
		helper := strings.TrimSpace(strings.TrimSpace(translation.HelpTitle + " " + translation.HelpBody))
		if helper != "" {
			result[translation.Locale] = helper
		}
	}
	seed := localizedQuestionHelperText(translations[sourceLocale])
	if seed == "" {
		seed = localizedQuestionHelperText(translations[wheelsetfit.SourceLocale])
	}
	if seed != "" {
		if result["en"] == "" {
			result["en"] = seed
		}
		if result["zh_cn"] == "" {
			result["zh_cn"] = seed
		}
	}
	return result
}

func localizedQuestionHelperText(translation wheelsetfit.QuestionTranslation) string {
	return strings.TrimSpace(strings.TrimSpace(translation.HelpTitle + " " + translation.HelpBody))
}

func localizedQuestionOptions(question wheelsetfit.Question, sourceLocale, nextNodeKey string) []selectionassistant.Option {
	result := make([]selectionassistant.Option, 0, len(question.Options))
	for _, option := range question.Options {
		if !option.IsEnabled {
			continue
		}
		optionTranslations := wheelsetFitOptionTranslationsByLocale(&option)
		result = append(result, selectionassistant.Option{
			Key:           option.OptionKey,
			Label:         localizedOptionLabelMap(optionTranslations, sourceLocale),
			Description:   localizedOptionDescriptionMap(optionTranslations, sourceLocale),
			AnswerEffects: map[string]string{question.AnswerKey: option.AnswerValue},
			QueryEffects:  localizedOptionQueryEffects(option.ProductFilterEffects),
			NextNodeKey:   nextNodeKey,
		})
	}
	return result
}

func localizedSummaryPromptMap(sourceLocale string) map[string]string {
	result := map[string]string{
		"en":    "Fit profile complete",
		"zh_cn": "选型结果已生成",
	}
	if strings.EqualFold(sourceLocale, "zh_cn") {
		result["zh_cn"] = "选型结果已生成"
	}
	return result
}

func localizedSummaryHelperMap(sourceLocale string) map[string]string {
	result := map[string]string{
		"en":    "Send this fit profile to support for review.",
		"zh_cn": "可把这份选型结果发给客服继续处理。",
	}
	if strings.EqualFold(sourceLocale, "zh_cn") {
		result["zh_cn"] = "可把这份选型结果发给客服继续处理。"
	}
	return result
}

func localizedOptionLabelMap(translations map[string]wheelsetfit.OptionTranslation, sourceLocale string) map[string]string {
	result := make(map[string]string, len(translations)+2)
	for _, translation := range translations {
		if text := strings.TrimSpace(translation.Label); text != "" {
			result[translation.Locale] = text
		}
	}
	seed := strings.TrimSpace(translations[sourceLocale].Label)
	if seed == "" {
		seed = strings.TrimSpace(translations[wheelsetfit.SourceLocale].Label)
	}
	if seed != "" {
		if result["en"] == "" {
			result["en"] = seed
		}
		if result["zh_cn"] == "" {
			result["zh_cn"] = seed
		}
	}
	return result
}

func localizedOptionDescriptionMap(translations map[string]wheelsetfit.OptionTranslation, sourceLocale string) map[string]string {
	result := make(map[string]string, len(translations)+2)
	for _, translation := range translations {
		if text := strings.TrimSpace(translation.Description); text != "" {
			result[translation.Locale] = text
		}
	}
	seed := strings.TrimSpace(translations[sourceLocale].Description)
	if seed == "" {
		seed = strings.TrimSpace(translations[wheelsetfit.SourceLocale].Description)
	}
	if seed != "" {
		if result["en"] == "" {
			result["en"] = seed
		}
		if result["zh_cn"] == "" {
			result["zh_cn"] = seed
		}
	}
	return result
}

func localizedOptionQueryEffects(raw []byte) selectionassistant.QueryEffects {
	raw = []byte(strings.TrimSpace(string(raw)))
	if len(raw) == 0 {
		return selectionassistant.QueryEffects{}
	}
	var payload struct {
		Keyword     string              `json:"keyword"`
		SpecFilters map[string][]string `json:"spec_filters"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return selectionassistant.QueryEffects{}
	}
	effects := selectionassistant.QueryEffects{
		Keyword: strings.TrimSpace(payload.Keyword),
	}
	if len(payload.SpecFilters) > 0 {
		effects.SpecFilters = make(map[string][]string, len(payload.SpecFilters))
		for key, values := range payload.SpecFilters {
			filtered := make([]string, 0, len(values))
			for _, value := range values {
				if text := strings.TrimSpace(value); text != "" {
					filtered = append(filtered, text)
				}
			}
			if len(filtered) > 0 {
				effects.SpecFilters[key] = filtered
			}
		}
	}
	return effects
}

func localizedSummaryPrompt(locale string) string {
	if strings.EqualFold(locale, "zh_cn") {
		return "继续由客服帮你确认细节"
	}
	return "Review the fit profile with support"
}

func localizedSummaryHelper(locale string) string {
	if strings.EqualFold(locale, "zh_cn") {
		return "把这份问卷结果发给客服，继续处理轮组匹配、商品建议和特殊需求。"
	}
	return "Send this questionnaire result to support for wheelset matching, product suggestions, and special requests."
}
