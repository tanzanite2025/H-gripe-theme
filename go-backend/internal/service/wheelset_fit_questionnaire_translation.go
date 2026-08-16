package service

import (
	wheelsetfit "commerce-platform/internal/domain/wheelsetfit"
	"commerce-platform/internal/pkg/locales"
)

func mergeWheelsetFitQuestionTranslations(existing map[string]wheelsetfit.QuestionTranslation, incoming map[string]WheelsetFitQuestionTranslationInput, sourceLocale string, sourceRevision int) []wheelsetfit.QuestionTranslation {
	sourceContent := existing[sourceLocale]
	if sourceInput, provided := incoming[sourceLocale]; provided {
		sourceContent.Locale = sourceLocale
		sourceContent.Prompt = sourceInput.Prompt
		sourceContent.HelpTitle = sourceInput.HelpTitle
		sourceContent.HelpBody = sourceInput.HelpBody
	}

	result := make([]wheelsetfit.QuestionTranslation, 0, len(locales.EnabledLocaleCodes()))
	for _, locale := range locales.EnabledLocaleCodes() {
		translation := existing[locale]
		provided, isProvided := incoming[locale]
		if isProvided {
			translation.Locale = locale
			translation.Prompt = provided.Prompt
			translation.HelpTitle = provided.HelpTitle
			translation.HelpBody = provided.HelpBody
		}
		translation.Locale = locale
		translation.SourceRevision = sourceRevision
		if locale == sourceLocale {
			translation.TranslatedRevision = sourceRevision
		} else if isProvided {
			translation.TranslatedRevision = translatedWheelsetFitQuestionRevision(translation, sourceContent, sourceRevision)
		}
		result = append(result, translation)
	}
	return result
}

func mergeWheelsetFitOptionTranslations(existing map[string]wheelsetfit.OptionTranslation, incoming []WheelsetFitQuestionOptionTranslationInput, sourceLocale string, sourceRevision int) []wheelsetfit.OptionTranslation {
	incomingByLocale := make(map[string]WheelsetFitQuestionOptionTranslationInput, len(incoming))
	for _, translation := range incoming {
		incomingByLocale[translation.Locale] = translation
	}

	result := make([]wheelsetfit.OptionTranslation, 0, len(locales.EnabledLocaleCodes()))
	for _, locale := range locales.EnabledLocaleCodes() {
		translation := existing[locale]
		provided, isProvided := incomingByLocale[locale]
		if isProvided {
			translation.Locale = locale
			translation.Label = provided.Label
			translation.Description = provided.Description
		}
		translation.Locale = locale
		translation.SourceRevision = sourceRevision
		if locale == sourceLocale {
			translation.TranslatedRevision = sourceRevision
		} else if isProvided {
			translation.TranslatedRevision = translatedWheelsetFitOptionRevision(translation, sourceRevision)
		}
		result = append(result, translation)
	}
	return result
}

func translatedWheelsetFitQuestionRevision(translation wheelsetfit.QuestionTranslation, source wheelsetfit.QuestionTranslation, sourceRevision int) int {
	if translation.Prompt == "" {
		return 0
	}
	if source.HelpTitle != "" && translation.HelpTitle == "" {
		return 0
	}
	if source.HelpBody != "" && translation.HelpBody == "" {
		return 0
	}
	return sourceRevision
}

func translatedWheelsetFitOptionRevision(translation wheelsetfit.OptionTranslation, sourceRevision int) int {
	if translation.Label == "" {
		return 0
	}
	return sourceRevision
}

func nextWheelsetFitQuestionSourceRevision(existing *wheelsetfit.Question, incoming map[string]WheelsetFitQuestionTranslationInput, sourceLocale string) int {
	if existing == nil {
		return 1
	}
	revision := normalizedWheelsetFitSourceRevision(existing.SourceRevision)
	source, provided := incoming[sourceLocale]
	if !provided {
		return revision
	}
	previous := wheelsetFitQuestionTranslationsByLocale(existing)[sourceLocale]
	if previous.Prompt != source.Prompt || previous.HelpTitle != source.HelpTitle || previous.HelpBody != source.HelpBody {
		return revision + 1
	}
	return revision
}

func nextWheelsetFitOptionSourceRevision(existing wheelsetfit.Option, exists bool, incoming []WheelsetFitQuestionOptionTranslationInput, sourceLocale string) int {
	if !exists {
		return 1
	}
	revision := normalizedWheelsetFitSourceRevision(existing.SourceRevision)
	for _, translation := range incoming {
		if translation.Locale != sourceLocale {
			continue
		}
		previous := wheelsetFitOptionTranslationsByLocale(&existing)[sourceLocale]
		if previous.Label != translation.Label || previous.Description != translation.Description {
			return revision + 1
		}
		break
	}
	return revision
}

func normalizedWheelsetFitSourceRevision(revision int) int {
	if revision <= 0 {
		return 1
	}
	return revision
}
