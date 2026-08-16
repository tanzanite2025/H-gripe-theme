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
	return &version, nil
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
