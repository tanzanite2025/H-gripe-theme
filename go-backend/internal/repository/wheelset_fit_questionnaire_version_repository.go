package repository

import (
	"time"

	wheelsetfit "commerce-platform/internal/domain/wheelsetfit"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetOrCreateDraft returns the editable draft when one already exists. When
// there is no draft, it deep-copies the current published snapshot instead of
// creating an empty version.
func (r *WheelsetFitQuestionnaireRepository) GetOrCreateDraft(questionnaireID uint) (*wheelsetfit.Version, error) {
	var draftID uint
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var questionnaire wheelsetfit.Questionnaire
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&questionnaire, questionnaireID).Error; err != nil {
			return err
		}

		var draft wheelsetfit.Version
		err := preloadWheelsetFitVersion(tx.Clauses(clause.Locking{Strength: "UPDATE"})).
			Where("questionnaire_id = ? AND status = ?", questionnaireID, wheelsetfit.VersionStatusDraft).
			Order("version_number DESC, id DESC").
			First(&draft).Error
		if err == nil {
			draftID = draft.ID
			return nil
		}
		if !IsRecordNotFound(err) {
			return err
		}

		latestNumber, err := nextWheelsetFitVersionNumber(tx, questionnaireID)
		if err != nil {
			return err
		}
		created := &wheelsetfit.Version{
			QuestionnaireID: questionnaireID,
			VersionNumber:   latestNumber,
			Status:          wheelsetfit.VersionStatusDraft,
		}
		if err := tx.Create(created).Error; err != nil {
			return err
		}

		var published wheelsetfit.Version
		err = preloadWheelsetFitVersion(tx).
			Where("questionnaire_id = ? AND status = ?", questionnaireID, wheelsetfit.VersionStatusPublished).
			Order("version_number DESC, id DESC").
			First(&published).Error
		if err != nil && !IsRecordNotFound(err) {
			return err
		}
		if err == nil {
			if err := createWheelsetFitQuestions(tx, created.ID, cloneWheelsetFitQuestions(published.Questions)); err != nil {
				return err
			}
		}

		draftID = created.ID
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.FindVersionByID(draftID)
}

// PublishVersionIfValid validates the locked draft immediately before status
// changes, so the validator sees exactly the snapshot that will be published.
func (r *WheelsetFitQuestionnaireRepository) PublishVersionIfValid(versionID uint, publishedBy *uint, publishedAt time.Time, validate func(*wheelsetfit.Version) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var version wheelsetfit.Version
		if err := preloadWheelsetFitVersion(tx.Clauses(clause.Locking{Strength: "UPDATE"})).First(&version, versionID).Error; err != nil {
			return err
		}
		if validate != nil {
			if err := validate(&version); err != nil {
				return err
			}
		}
		if err := tx.Model(&wheelsetfit.Version{}).
			Where("questionnaire_id = ? AND id <> ? AND status = ?", version.QuestionnaireID, version.ID, wheelsetfit.VersionStatusPublished).
			Update("status", wheelsetfit.VersionStatusArchived).Error; err != nil {
			return err
		}
		result := tx.Model(&wheelsetfit.Version{}).
			Where("id = ? AND status = ?", version.ID, wheelsetfit.VersionStatusDraft).
			Updates(map[string]interface{}{
				"status":       wheelsetfit.VersionStatusPublished,
				"published_at": publishedAt,
				"published_by": publishedBy,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func nextWheelsetFitVersionNumber(tx *gorm.DB, questionnaireID uint) (int, error) {
	var latest wheelsetfit.Version
	err := tx.Where("questionnaire_id = ?", questionnaireID).
		Order("version_number DESC, id DESC").
		First(&latest).Error
	if IsRecordNotFound(err) {
		return 1, nil
	}
	if err != nil {
		return 0, err
	}
	return latest.VersionNumber + 1, nil
}
