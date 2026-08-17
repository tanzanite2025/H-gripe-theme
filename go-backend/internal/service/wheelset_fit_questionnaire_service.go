package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	wheelsetfit "commerce-platform/internal/domain/wheelsetfit"
	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrWheelsetFitQuestionnaireInvalid         = errors.New("invalid wheelset fit questionnaire")
	ErrWheelsetFitQuestionnaireNotFound        = errors.New("wheelset fit questionnaire not found")
	ErrWheelsetFitQuestionnaireVersionNotFound = errors.New("wheelset fit questionnaire version not found")
	ErrWheelsetFitQuestionNotFound             = errors.New("wheelset fit question not found")
	ErrWheelsetFitQuestionnaireNotMutable      = errors.New("wheelset fit questionnaire version is not mutable")
)

type WheelsetFitQuestionnaireService struct {
	repo *repository.WheelsetFitQuestionnaireRepository
}

type WheelsetFitQuestionOrderInput struct {
	QuestionIDs []uint `json:"question_ids"`
}

type WheelsetFitQuestionnaireValidationResult struct {
	Valid  bool                                      `json:"valid"`
	Issues []WheelsetFitQuestionnaireValidationIssue `json:"issues"`
}

type WheelsetFitQuestionnaireValidationIssue struct {
	Severity    string `json:"severity"`
	Code        string `json:"code"`
	Message     string `json:"message"`
	QuestionID  uint   `json:"question_id,omitempty"`
	QuestionKey string `json:"question_key,omitempty"`
	OptionKey   string `json:"option_key,omitempty"`
	Locale      string `json:"locale,omitempty"`
}

func NewWheelsetFitQuestionnaireService(repo *repository.WheelsetFitQuestionnaireRepository) *WheelsetFitQuestionnaireService {
	return &WheelsetFitQuestionnaireService{repo: repo}
}

func (s *WheelsetFitQuestionnaireService) GetCurrentVersion() (*wheelsetfit.Version, error) {
	questionnaire, err := s.getQuestionnaire()
	if err != nil {
		return nil, err
	}

	version, err := s.repo.FindCurrentVersion(questionnaire.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrWheelsetFitQuestionnaireVersionNotFound
	}
	if err != nil {
		return nil, err
	}
	return version, nil
}

func (s *WheelsetFitQuestionnaireService) CreateDraft() (*wheelsetfit.Version, error) {
	return s.GetOrCreateDraft()
}

// GetOrCreateDraft returns the existing draft unchanged. Only when the
// questionnaire has no draft does the repository deep-copy its published
// version into a new draft.
func (s *WheelsetFitQuestionnaireService) GetOrCreateDraft() (*wheelsetfit.Version, error) {
	questionnaire, err := s.getQuestionnaire()
	if err != nil {
		return nil, err
	}
	return s.repo.GetOrCreateDraft(questionnaire.ID)
}

func (s *WheelsetFitQuestionnaireService) getQuestionnaire() (*wheelsetfit.Questionnaire, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("wheelset fit questionnaire service is not configured")
	}

	questionnaire, err := s.repo.FindSingleton()
	if repository.IsRecordNotFound(err) {
		return nil, ErrWheelsetFitQuestionnaireNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := validateWheelsetFitQuestionnaire(*questionnaire); err != nil {
		return nil, err
	}
	return questionnaire, nil
}

// SaveQuestion saves a question in the current draft. The draft is reused
// whenever it exists; published content is copied only by GetOrCreateDraft.
func (s *WheelsetFitQuestionnaireService) SaveQuestion(input WheelsetFitQuestionInput) (*wheelsetfit.Version, error) {
	draft, err := s.GetOrCreateDraft()
	if err != nil {
		return nil, err
	}

	saved, err := s.repo.SaveDraftQuestion(draft.ID, func(lockedDraft *wheelsetfit.Version) (*wheelsetfit.Question, error) {
		sourceLocale, err := wheelsetFitQuestionnaireSourceLocale(lockedDraft)
		if err != nil {
			return nil, err
		}
		existing := findWheelsetFitQuestion(lockedDraft.Questions, input.ID)
		if input.ID != 0 && existing == nil {
			return nil, ErrWheelsetFitQuestionNotFound
		}
		return buildWheelsetFitQuestion(input, existing, lockedDraft.ID, sourceLocale)
	})
	if errors.Is(err, repository.ErrWheelsetFitDraftVersionNotMutable) {
		return nil, ErrWheelsetFitQuestionnaireNotMutable
	}
	if err != nil {
		return nil, err
	}
	return saved, nil
}

func (s *WheelsetFitQuestionnaireService) DeleteQuestion(questionID uint) (*wheelsetfit.Version, error) {
	if questionID == 0 {
		return nil, fmt.Errorf("%w: question id is required", ErrWheelsetFitQuestionnaireInvalid)
	}

	draft, err := s.GetOrCreateDraft()
	if err != nil {
		return nil, err
	}
	deleted, err := s.repo.DeleteDraftQuestion(draft.ID, questionID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrWheelsetFitQuestionNotFound
	}
	if errors.Is(err, repository.ErrWheelsetFitDraftVersionNotMutable) {
		return nil, ErrWheelsetFitQuestionnaireNotMutable
	}
	if err != nil {
		return nil, err
	}
	return deleted, nil
}

func (s *WheelsetFitQuestionnaireService) ReorderQuestions(input WheelsetFitQuestionOrderInput) (*wheelsetfit.Version, error) {
	draft, err := s.GetOrCreateDraft()
	if err != nil {
		return nil, err
	}
	reordered, err := s.repo.ReorderDraftQuestions(draft.ID, input.QuestionIDs)
	if errors.Is(err, repository.ErrWheelsetFitQuestionOrderInvalid) {
		return nil, fmt.Errorf("%w: question_ids must include every draft question exactly once", ErrWheelsetFitQuestionnaireInvalid)
	}
	if errors.Is(err, repository.ErrWheelsetFitDraftVersionNotMutable) {
		return nil, ErrWheelsetFitQuestionnaireNotMutable
	}
	if err != nil {
		return nil, err
	}
	return reordered, nil
}

func (s *WheelsetFitQuestionnaireService) ValidateVersion(versionID uint) (*WheelsetFitQuestionnaireValidationResult, error) {
	version, err := s.repo.FindVersionByID(versionID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrWheelsetFitQuestionnaireVersionNotFound
	}
	if err != nil {
		return nil, err
	}
	result := validateWheelsetFitQuestionnaireVersion(*version)
	return &result, nil
}

func (s *WheelsetFitQuestionnaireService) PublishVersion(versionID uint, publishedBy *uint) (*wheelsetfit.Version, error) {
	version, err := s.repo.FindVersionByID(versionID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrWheelsetFitQuestionnaireVersionNotFound
	}
	if err != nil {
		return nil, err
	}
	if version.Status != wheelsetfit.VersionStatusDraft {
		return nil, ErrWheelsetFitQuestionnaireNotMutable
	}

	validation := validateWheelsetFitQuestionnaireVersion(*version)
	if !validation.Valid {
		return nil, fmt.Errorf("%w: %s", ErrWheelsetFitQuestionnaireInvalid, validation.Issues[0].Message)
	}
	if err := s.repo.PublishVersionIfValid(versionID, publishedBy, time.Now().UTC(), func(locked *wheelsetfit.Version) error {
		result := validateWheelsetFitQuestionnaireVersion(*locked)
		if result.Valid {
			return nil
		}
		return fmt.Errorf("%w: %s", ErrWheelsetFitQuestionnaireInvalid, result.Issues[0].Message)
	}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWheelsetFitQuestionnaireVersionNotFound
		}
		return nil, err
	}
	return s.repo.FindVersionByID(versionID)
}

func validateWheelsetFitQuestionnaire(questionnaire wheelsetfit.Questionnaire) error {
	if questionnaire.Slug != wheelsetfit.QuestionnaireSlug ||
		questionnaire.ProductCategorySlug != wheelsetfit.WheelsetProductCategorySlug {
		return fmt.Errorf("%w: expected singleton %q for category %q", ErrWheelsetFitQuestionnaireInvalid, wheelsetfit.QuestionnaireSlug, wheelsetfit.WheelsetProductCategorySlug)
	}
	if _, err := requireSupportedLocale(questionnaire.SourceLocale); err != nil {
		return fmt.Errorf("%w: source locale: %v", ErrWheelsetFitQuestionnaireInvalid, err)
	}
	return nil
}

func wheelsetFitQuestionnaireSourceLocale(draft *wheelsetfit.Version) (string, error) {
	sourceLocale := wheelsetfit.SourceLocale
	if draft != nil && draft.Questionnaire != nil && strings.TrimSpace(draft.Questionnaire.SourceLocale) != "" {
		sourceLocale = draft.Questionnaire.SourceLocale
	}
	sourceLocale, err := requireSupportedLocale(sourceLocale)
	if err != nil {
		return "", fmt.Errorf("%w: source locale: %v", ErrWheelsetFitQuestionnaireInvalid, err)
	}
	return sourceLocale, nil
}

func findWheelsetFitQuestion(questions []wheelsetfit.Question, id uint) *wheelsetfit.Question {
	if id == 0 {
		return nil
	}
	for index := range questions {
		if questions[index].ID == id {
			return &questions[index]
		}
	}
	return nil
}
