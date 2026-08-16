package service

import (
	"sort"

	wheelsetfit "commerce-platform/internal/domain/wheelsetfit"

	"gorm.io/datatypes"
)

// buildWheelsetFitQuestion turns the one-question patch into a complete
// persistence aggregate. Omitted locales retain their prior values; options
// are retained only when the options field itself is omitted.
func buildWheelsetFitQuestion(input WheelsetFitQuestionInput, existing *wheelsetfit.Question, versionID uint, sourceLocale string) (*wheelsetfit.Question, error) {
	normalized, incomingTranslations, incomingOptions, err := normalizeWheelsetFitQuestionInput(input)
	if err != nil {
		return nil, err
	}

	question := &wheelsetfit.Question{
		QuestionnaireVersionID: versionID,
		QuestionKey:            normalized.QuestionKey,
		AnswerKey:              normalized.AnswerKey,
		SortOrder:              normalized.SortOrder,
		InputMode:              normalized.InputMode,
		IsRequired:             normalized.IsRequired,
		AllowUnknown:           normalized.AllowUnknown,
		IsEnabled:              normalized.IsEnabled,
		SourceRevision:         nextWheelsetFitQuestionSourceRevision(existing, incomingTranslations, sourceLocale),
	}
	if existing != nil {
		question.ID = existing.ID
	}
	question.Translations = mergeWheelsetFitQuestionTranslations(
		wheelsetFitQuestionTranslationsByLocale(existing),
		incomingTranslations,
		sourceLocale,
		question.SourceRevision,
	)
	if existing != nil && input.Options == nil {
		question.Options = completeWheelsetFitExistingOptions(existing, sourceLocale)
	} else {
		question.Options = mergeWheelsetFitQuestionOptions(existing, incomingOptions, sourceLocale)
	}
	return question, nil
}

func mergeWheelsetFitQuestionOptions(existing *wheelsetfit.Question, incoming map[string]WheelsetFitQuestionOptionInput, sourceLocale string) []wheelsetfit.Option {
	existingByKey := make(map[string]wheelsetfit.Option)
	if existing != nil {
		for _, option := range existing.Options {
			existingByKey[option.OptionKey] = option
		}
	}

	keys := make([]string, 0, len(incoming))
	for key := range incoming {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		left, right := incoming[keys[i]], incoming[keys[j]]
		if left.SortOrder == right.SortOrder {
			return left.OptionKey < right.OptionKey
		}
		return left.SortOrder < right.SortOrder
	})

	result := make([]wheelsetfit.Option, 0, len(keys))
	for _, key := range keys {
		input := incoming[key]
		previous, exists := existingByKey[key]
		effects := input.ProductFilterEffects
		if len(effects) == 0 {
			if exists && len(previous.ProductFilterEffects) > 0 {
				effects = previous.ProductFilterEffects
			} else {
				effects = datatypes.JSON([]byte("{}"))
			}
		}
		option := wheelsetfit.Option{
			ID:                   previous.ID,
			OptionKey:            input.OptionKey,
			AnswerValue:          input.AnswerValue,
			SortOrder:            input.SortOrder,
			IsUnknown:            input.IsUnknown,
			IsEnabled:            input.IsEnabled,
			ProductFilterEffects: effects,
			SourceRevision:       nextWheelsetFitOptionSourceRevision(previous, exists, input.Translations, sourceLocale),
		}
		option.Translations = mergeWheelsetFitOptionTranslations(
			wheelsetFitOptionTranslationsByLocale(&previous),
			input.Translations,
			sourceLocale,
			option.SourceRevision,
		)
		result = append(result, option)
	}
	return result
}

func completeWheelsetFitExistingOptions(existing *wheelsetfit.Question, sourceLocale string) []wheelsetfit.Option {
	if existing == nil {
		return []wheelsetfit.Option{}
	}
	options := append([]wheelsetfit.Option(nil), existing.Options...)
	for optionIndex := range options {
		option := &options[optionIndex]
		option.SourceRevision = normalizedWheelsetFitSourceRevision(option.SourceRevision)
		option.Translations = mergeWheelsetFitOptionTranslations(
			wheelsetFitOptionTranslationsByLocale(option),
			nil,
			sourceLocale,
			option.SourceRevision,
		)
	}
	sort.Slice(options, func(i, j int) bool {
		if options[i].SortOrder == options[j].SortOrder {
			return options[i].OptionKey < options[j].OptionKey
		}
		return options[i].SortOrder < options[j].SortOrder
	})
	return options
}

func wheelsetFitQuestionTranslationsByLocale(question *wheelsetfit.Question) map[string]wheelsetfit.QuestionTranslation {
	translations := make(map[string]wheelsetfit.QuestionTranslation)
	if question == nil {
		return translations
	}
	for _, translation := range question.Translations {
		translations[translation.Locale] = translation
	}
	return translations
}

func wheelsetFitOptionTranslationsByLocale(option *wheelsetfit.Option) map[string]wheelsetfit.OptionTranslation {
	translations := make(map[string]wheelsetfit.OptionTranslation)
	if option == nil {
		return translations
	}
	for _, translation := range option.Translations {
		translations[translation.Locale] = translation
	}
	return translations
}
