package repository

import (
	wheelsetfit "commerce-platform/internal/domain/wheelsetfit"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SaveDraftQuestion locks the draft before exposing its aggregate to build.
// Callers therefore merge against current data and write their replacement in
// one transaction, rather than using a stale draft snapshot.
func (r *WheelsetFitQuestionnaireRepository) SaveDraftQuestion(versionID uint, build func(*wheelsetfit.Version) (*wheelsetfit.Question, error)) (*wheelsetfit.Version, error) {
	if build == nil {
		return nil, gorm.ErrInvalidData
	}

	var savedVersionID uint
	err := r.db.Transaction(func(tx *gorm.DB) error {
		draft, err := lockWheelsetFitDraft(tx, versionID)
		if err != nil {
			return err
		}
		question, err := build(draft)
		if err != nil {
			return err
		}
		if question == nil || question.QuestionnaireVersionID != draft.ID {
			return gorm.ErrInvalidData
		}

		if question.ID == 0 {
			if err := createWheelsetFitQuestion(tx, question); err != nil {
				return err
			}
		} else if err := replaceWheelsetFitQuestion(tx, question); err != nil {
			return err
		}
		savedVersionID = draft.ID
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.FindVersionByID(savedVersionID)
}

func lockWheelsetFitDraft(tx *gorm.DB, versionID uint) (*wheelsetfit.Version, error) {
	var draft wheelsetfit.Version
	if err := preloadWheelsetFitVersion(tx.Clauses(clause.Locking{Strength: "UPDATE"})).First(&draft, versionID).Error; err != nil {
		return nil, err
	}
	if draft.Status != wheelsetfit.VersionStatusDraft {
		return nil, ErrWheelsetFitDraftVersionNotMutable
	}
	return &draft, nil
}

func createWheelsetFitQuestion(tx *gorm.DB, question *wheelsetfit.Question) error {
	translations := question.Translations
	options := question.Options
	question.Translations = nil
	question.Options = nil
	if err := tx.Create(question).Error; err != nil {
		return err
	}
	question.Translations = translations
	question.Options = options
	return createWheelsetFitQuestionChildren(tx, question)
}

func replaceWheelsetFitQuestion(tx *gorm.DB, question *wheelsetfit.Question) error {
	result := tx.Model(&wheelsetfit.Question{}).
		Where("id = ? AND questionnaire_version_id = ?", question.ID, question.QuestionnaireVersionID).
		Updates(map[string]interface{}{
			"question_key":    question.QuestionKey,
			"answer_key":      question.AnswerKey,
			"sort_order":      question.SortOrder,
			"input_mode":      question.InputMode,
			"is_required":     question.IsRequired,
			"allow_unknown":   question.AllowUnknown,
			"is_enabled":      question.IsEnabled,
			"source_revision": question.SourceRevision,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return upsertWheelsetFitQuestionChildren(tx, question)
}
