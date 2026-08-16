package service

import (
	"fmt"
	"regexp"
	"strings"

	wheelsetfit "commerce-platform/internal/domain/wheelsetfit"

	"gorm.io/datatypes"
)

var wheelsetFitKeyPattern = regexp.MustCompile(`^[a-z0-9]+(?:_[a-z0-9]+)*$`)

// WheelsetFitQuestionInput is intentionally scoped to one question. Admin
// edits one locale at a time, so omitted translations retain their existing
// values instead of being interpreted as empty replacements.
type WheelsetFitQuestionInput struct {
	ID           uint                                  `json:"id"`
	QuestionKey  string                                `json:"question_key"`
	AnswerKey    string                                `json:"answer_key"`
	SortOrder    int                                   `json:"sort_order"`
	InputMode    string                                `json:"input_mode"`
	IsRequired   bool                                  `json:"is_required"`
	AllowUnknown bool                                  `json:"allow_unknown"`
	IsEnabled    bool                                  `json:"is_enabled"`
	Translations []WheelsetFitQuestionTranslationInput `json:"translations"`
	Options      []WheelsetFitQuestionOptionInput      `json:"options"`
}

type WheelsetFitQuestionTranslationInput struct {
	Locale    string `json:"locale"`
	Prompt    string `json:"prompt"`
	HelpTitle string `json:"help_title"`
	HelpBody  string `json:"help_body"`
}

type WheelsetFitQuestionOptionInput struct {
	OptionKey            string                                      `json:"option_key"`
	AnswerValue          string                                      `json:"answer_value"`
	SortOrder            int                                         `json:"sort_order"`
	IsUnknown            bool                                        `json:"is_unknown"`
	IsEnabled            bool                                        `json:"is_enabled"`
	ProductFilterEffects datatypes.JSON                              `json:"product_filter_effects"`
	Translations         []WheelsetFitQuestionOptionTranslationInput `json:"translations"`
}

type WheelsetFitQuestionOptionTranslationInput struct {
	Locale      string `json:"locale"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

func normalizeWheelsetFitQuestionInput(input WheelsetFitQuestionInput) (WheelsetFitQuestionInput, map[string]WheelsetFitQuestionTranslationInput, map[string]WheelsetFitQuestionOptionInput, error) {
	questionKey := normalizeWheelsetFitKey(input.QuestionKey)
	if questionKey == "" {
		return WheelsetFitQuestionInput{}, nil, nil, fmt.Errorf("%w: question_key is required and must use lowercase snake_case", ErrWheelsetFitQuestionnaireInvalid)
	}
	answerKey := normalizeWheelsetFitKey(input.AnswerKey)
	if answerKey == "" {
		return WheelsetFitQuestionInput{}, nil, nil, fmt.Errorf("%w: answer_key is required and must use lowercase snake_case", ErrWheelsetFitQuestionnaireInvalid)
	}
	if input.SortOrder <= 0 {
		return WheelsetFitQuestionInput{}, nil, nil, fmt.Errorf("%w: sort_order must be positive", ErrWheelsetFitQuestionnaireInvalid)
	}
	inputMode := strings.TrimSpace(input.InputMode)
	if inputMode == "" {
		inputMode = wheelsetfit.InputModeSingleChoice
	}
	if inputMode != wheelsetfit.InputModeSingleChoice {
		return WheelsetFitQuestionInput{}, nil, nil, fmt.Errorf("%w: input_mode must be %q", ErrWheelsetFitQuestionnaireInvalid, wheelsetfit.InputModeSingleChoice)
	}

	translations, err := normalizeWheelsetFitQuestionTranslations(input.Translations)
	if err != nil {
		return WheelsetFitQuestionInput{}, nil, nil, err
	}
	options, err := normalizeWheelsetFitQuestionOptions(input.Options)
	if err != nil {
		return WheelsetFitQuestionInput{}, nil, nil, err
	}

	input.QuestionKey = questionKey
	input.AnswerKey = answerKey
	input.InputMode = inputMode
	return input, translations, options, nil
}

func normalizeWheelsetFitQuestionTranslations(input []WheelsetFitQuestionTranslationInput) (map[string]WheelsetFitQuestionTranslationInput, error) {
	translations := make(map[string]WheelsetFitQuestionTranslationInput, len(input))
	for index, translation := range input {
		locale, err := requireSupportedLocale(translation.Locale)
		if err != nil {
			return nil, fmt.Errorf("%w: translation %d has invalid locale", ErrWheelsetFitQuestionnaireInvalid, index+1)
		}
		if _, duplicate := translations[locale]; duplicate {
			return nil, fmt.Errorf("%w: duplicate translation locale %q", ErrWheelsetFitQuestionnaireInvalid, locale)
		}
		translation.Locale = locale
		translation.Prompt = strings.TrimSpace(translation.Prompt)
		translation.HelpTitle = strings.TrimSpace(translation.HelpTitle)
		translation.HelpBody = strings.TrimSpace(translation.HelpBody)
		translations[locale] = translation
	}
	return translations, nil
}

func normalizeWheelsetFitQuestionOptions(input []WheelsetFitQuestionOptionInput) (map[string]WheelsetFitQuestionOptionInput, error) {
	options := make(map[string]WheelsetFitQuestionOptionInput, len(input))
	for index, option := range input {
		optionKey := normalizeWheelsetFitKey(option.OptionKey)
		if optionKey == "" {
			return nil, fmt.Errorf("%w: option %d key is required and must use lowercase snake_case", ErrWheelsetFitQuestionnaireInvalid, index+1)
		}
		if _, duplicate := options[optionKey]; duplicate {
			return nil, fmt.Errorf("%w: duplicate option key %q", ErrWheelsetFitQuestionnaireInvalid, optionKey)
		}
		option.AnswerValue = strings.TrimSpace(option.AnswerValue)
		if option.AnswerValue == "" {
			return nil, fmt.Errorf("%w: option %q answer_value is required", ErrWheelsetFitQuestionnaireInvalid, optionKey)
		}
		if option.SortOrder <= 0 {
			return nil, fmt.Errorf("%w: option %q sort_order must be positive", ErrWheelsetFitQuestionnaireInvalid, optionKey)
		}
		translations, err := normalizeWheelsetFitOptionTranslations(option.Translations)
		if err != nil {
			return nil, fmt.Errorf("%w: option %q: %v", ErrWheelsetFitQuestionnaireInvalid, optionKey, err)
		}
		option.OptionKey = optionKey
		option.Translations = translations
		options[optionKey] = option
	}
	return options, nil
}

func normalizeWheelsetFitOptionTranslations(input []WheelsetFitQuestionOptionTranslationInput) ([]WheelsetFitQuestionOptionTranslationInput, error) {
	translations := make([]WheelsetFitQuestionOptionTranslationInput, 0, len(input))
	seen := make(map[string]struct{}, len(input))
	for index, translation := range input {
		locale, err := requireSupportedLocale(translation.Locale)
		if err != nil {
			return nil, fmt.Errorf("translation %d has invalid locale", index+1)
		}
		if _, duplicate := seen[locale]; duplicate {
			return nil, fmt.Errorf("duplicate translation locale %q", locale)
		}
		seen[locale] = struct{}{}
		translation.Locale = locale
		translation.Label = strings.TrimSpace(translation.Label)
		translation.Description = strings.TrimSpace(translation.Description)
		translations = append(translations, translation)
	}
	return translations, nil
}

func normalizeWheelsetFitKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if !wheelsetFitKeyPattern.MatchString(value) {
		return ""
	}
	return value
}
