package repository

import (
	wheelsetfit "commerce-platform/internal/domain/wheelsetfit"

	"gorm.io/gorm"
)

func cloneWheelsetFitQuestions(source []wheelsetfit.Question) []wheelsetfit.Question {
	questions := make([]wheelsetfit.Question, len(source))
	for questionIndex := range source {
		questions[questionIndex] = source[questionIndex]
		questions[questionIndex].Translations = append([]wheelsetfit.QuestionTranslation(nil), source[questionIndex].Translations...)
		questions[questionIndex].Options = make([]wheelsetfit.Option, len(source[questionIndex].Options))
		for optionIndex := range source[questionIndex].Options {
			questions[questionIndex].Options[optionIndex] = source[questionIndex].Options[optionIndex]
			questions[questionIndex].Options[optionIndex].Translations = append(
				[]wheelsetfit.OptionTranslation(nil),
				source[questionIndex].Options[optionIndex].Translations...,
			)
		}
	}
	return questions
}

func createWheelsetFitQuestions(tx *gorm.DB, versionID uint, questions []wheelsetfit.Question) error {
	for questionIndex := range questions {
		question := &questions[questionIndex]
		question.ID = 0
		question.QuestionnaireVersionID = versionID
		translations := question.Translations
		options := question.Options
		question.Translations = nil
		question.Options = nil
		if err := tx.Create(question).Error; err != nil {
			return err
		}
		question.Translations = translations
		question.Options = options
		if err := createWheelsetFitQuestionChildren(tx, question); err != nil {
			return err
		}
	}
	return nil
}

func createWheelsetFitQuestionChildren(tx *gorm.DB, question *wheelsetfit.Question) error {
	translations := question.Translations
	for translationIndex := range translations {
		translations[translationIndex].ID = 0
		translations[translationIndex].QuestionID = question.ID
	}
	if len(translations) > 0 {
		if err := tx.Create(&translations).Error; err != nil {
			return err
		}
	}

	options := question.Options
	for optionIndex := range options {
		option := &options[optionIndex]
		option.ID = 0
		option.QuestionID = question.ID
		optionTranslations := option.Translations
		option.Translations = nil
		if err := tx.Create(option).Error; err != nil {
			return err
		}
		for translationIndex := range optionTranslations {
			optionTranslations[translationIndex].ID = 0
			optionTranslations[translationIndex].OptionID = option.ID
		}
		if len(optionTranslations) > 0 {
			if err := tx.Create(&optionTranslations).Error; err != nil {
				return err
			}
		}
		option.Translations = optionTranslations
	}
	question.Translations = translations
	question.Options = options
	return nil
}

// upsertWheelsetFitQuestionChildren receives the completed aggregate built by
// the service. Translation rows are upserted by locale and option identities
// are kept stable through option_key.
func upsertWheelsetFitQuestionChildren(tx *gorm.DB, question *wheelsetfit.Question) error {
	if err := upsertWheelsetFitQuestionTranslations(tx, question.ID, question.Translations); err != nil {
		return err
	}

	var existingOptions []wheelsetfit.Option
	if err := tx.Where("question_id = ?", question.ID).Find(&existingOptions).Error; err != nil {
		return err
	}
	existingByKey := make(map[string]wheelsetfit.Option, len(existingOptions))
	for _, option := range existingOptions {
		existingByKey[option.OptionKey] = option
	}

	providedKeys := make(map[string]struct{}, len(question.Options))
	for optionIndex := range question.Options {
		option := &question.Options[optionIndex]
		providedKeys[option.OptionKey] = struct{}{}
		existing, exists := existingByKey[option.OptionKey]
		if !exists {
			option.ID = 0
			option.QuestionID = question.ID
			translations := option.Translations
			option.Translations = nil
			if err := tx.Create(option).Error; err != nil {
				return err
			}
			option.Translations = translations
			if err := upsertWheelsetFitOptionTranslations(tx, option.ID, translations); err != nil {
				return err
			}
			continue
		}

		option.ID = existing.ID
		option.QuestionID = question.ID
		result := tx.Model(&wheelsetfit.Option{}).
			Where("id = ? AND question_id = ?", option.ID, question.ID).
			Updates(map[string]interface{}{
				"answer_value":           option.AnswerValue,
				"sort_order":             option.SortOrder,
				"is_unknown":             option.IsUnknown,
				"is_enabled":             option.IsEnabled,
				"product_filter_effects": option.ProductFilterEffects,
				"source_revision":        option.SourceRevision,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		if err := upsertWheelsetFitOptionTranslations(tx, option.ID, option.Translations); err != nil {
			return err
		}
	}

	for _, option := range existingOptions {
		if _, provided := providedKeys[option.OptionKey]; provided {
			continue
		}
		if err := tx.Delete(&wheelsetfit.Option{}, option.ID).Error; err != nil {
			return err
		}
	}
	return nil
}

func upsertWheelsetFitQuestionTranslations(tx *gorm.DB, questionID uint, translations []wheelsetfit.QuestionTranslation) error {
	for translationIndex := range translations {
		translation := &translations[translationIndex]
		var existing wheelsetfit.QuestionTranslation
		err := tx.Where("question_id = ? AND locale = ?", questionID, translation.Locale).First(&existing).Error
		if IsRecordNotFound(err) {
			translation.ID = 0
			translation.QuestionID = questionID
			if err := tx.Create(translation).Error; err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		translation.ID = existing.ID
		translation.QuestionID = questionID
		if err := tx.Model(&wheelsetfit.QuestionTranslation{}).Where("id = ?", existing.ID).Updates(map[string]interface{}{
			"prompt":              translation.Prompt,
			"help_title":          translation.HelpTitle,
			"help_body":           translation.HelpBody,
			"source_revision":     translation.SourceRevision,
			"translated_revision": translation.TranslatedRevision,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func upsertWheelsetFitOptionTranslations(tx *gorm.DB, optionID uint, translations []wheelsetfit.OptionTranslation) error {
	for translationIndex := range translations {
		translation := &translations[translationIndex]
		var existing wheelsetfit.OptionTranslation
		err := tx.Where("option_id = ? AND locale = ?", optionID, translation.Locale).First(&existing).Error
		if IsRecordNotFound(err) {
			translation.ID = 0
			translation.OptionID = optionID
			if err := tx.Create(translation).Error; err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		translation.ID = existing.ID
		translation.OptionID = optionID
		if err := tx.Model(&wheelsetfit.OptionTranslation{}).Where("id = ?", existing.ID).Updates(map[string]interface{}{
			"label":               translation.Label,
			"description":         translation.Description,
			"source_revision":     translation.SourceRevision,
			"translated_revision": translation.TranslatedRevision,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}
