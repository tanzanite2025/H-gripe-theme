package repository

import (
	"errors"

	wheelsetfit "commerce-platform/internal/domain/wheelsetfit"

	"gorm.io/gorm"
)

type WheelsetFitQuestionnaireRepository struct {
	db *gorm.DB
}

var ErrWheelsetFitDraftVersionNotMutable = errors.New("wheelset fit questionnaire draft version is not mutable")

func NewWheelsetFitQuestionnaireRepository(db *gorm.DB) *WheelsetFitQuestionnaireRepository {
	return &WheelsetFitQuestionnaireRepository{db: db}
}

func (r *WheelsetFitQuestionnaireRepository) FindSingleton() (*wheelsetfit.Questionnaire, error) {
	var questionnaire wheelsetfit.Questionnaire
	err := r.db.
		Where("slug = ?", wheelsetfit.QuestionnaireSlug).
		First(&questionnaire).Error
	if err != nil {
		return nil, err
	}
	return &questionnaire, nil
}

func (r *WheelsetFitQuestionnaireRepository) FindVersionByID(id uint) (*wheelsetfit.Version, error) {
	var version wheelsetfit.Version
	if err := preloadWheelsetFitVersion(r.db).First(&version, id).Error; err != nil {
		return nil, err
	}
	return normalizeWheelsetFitVersion(&version), nil
}

// FindCurrentVersion prefers the editable draft. Without one, it exposes the
// latest published snapshot for read-only viewing.
func (r *WheelsetFitQuestionnaireRepository) FindCurrentVersion(questionnaireID uint) (*wheelsetfit.Version, error) {
	var version wheelsetfit.Version
	err := preloadWheelsetFitVersion(r.db).
		Where("questionnaire_id = ? AND status = ?", questionnaireID, wheelsetfit.VersionStatusDraft).
		Order("version_number DESC, id DESC").
		First(&version).Error
	if err == nil {
		return normalizeWheelsetFitVersion(&version), nil
	}
	if !IsRecordNotFound(err) {
		return nil, err
	}

	err = preloadWheelsetFitVersion(r.db).
		Where("questionnaire_id = ? AND status = ?", questionnaireID, wheelsetfit.VersionStatusPublished).
		Order("version_number DESC, id DESC").
		First(&version).Error
	if err != nil {
		return nil, err
	}
	return normalizeWheelsetFitVersion(&version), nil
}

func (r *WheelsetFitQuestionnaireRepository) FindPublishedVersion(questionnaireID uint) (*wheelsetfit.Version, error) {
	var version wheelsetfit.Version
	err := preloadWheelsetFitVersion(r.db).
		Where("questionnaire_id = ? AND status = ?", questionnaireID, wheelsetfit.VersionStatusPublished).
		Order("version_number DESC, id DESC").
		First(&version).Error
	if err != nil {
		return nil, err
	}
	return normalizeWheelsetFitVersion(&version), nil
}

func preloadWheelsetFitVersion(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Questionnaire").
		Preload("Questions", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		Preload("Questions.Translations", func(db *gorm.DB) *gorm.DB {
			return db.Order("locale ASC, id ASC")
		}).
		Preload("Questions.Options", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		Preload("Questions.Options.Translations", func(db *gorm.DB) *gorm.DB {
			return db.Order("locale ASC, id ASC")
		})
}

func normalizeWheelsetFitVersion(version *wheelsetfit.Version) *wheelsetfit.Version {
	if version == nil {
		return nil
	}
	if version.Questions == nil {
		version.Questions = make([]wheelsetfit.Question, 0)
	}
	for questionIndex := range version.Questions {
		question := &version.Questions[questionIndex]
		if question.Options == nil {
			question.Options = make([]wheelsetfit.Option, 0)
		}
		if question.Translations == nil {
			question.Translations = make([]wheelsetfit.QuestionTranslation, 0)
		}
		for optionIndex := range question.Options {
			if question.Options[optionIndex].Translations == nil {
				question.Options[optionIndex].Translations = make([]wheelsetfit.OptionTranslation, 0)
			}
		}
	}
	return version
}
