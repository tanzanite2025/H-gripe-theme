package service

import (
	"errors"
	"fmt"
	"strings"

	wheelsetfit "commerce-platform/internal/domain/wheelsetfit"
	"commerce-platform/internal/repository"
)

var (
	ErrWheelsetFitQuestionnaireInvalid    = errors.New("invalid wheelset fit questionnaire")
	ErrWheelsetFitQuestionnaireNotFound   = errors.New("wheelset fit questionnaire not found")
	ErrWheelsetFitQuestionNotFound        = errors.New("wheelset fit question not found")
	ErrWheelsetFitQuestionnaireNotMutable = errors.New("wheelset fit questionnaire version is not mutable")
)

type WheelsetFitQuestionnaireService struct {
	repo *repository.WheelsetFitQuestionnaireRepository
}

func NewWheelsetFitQuestionnaireService(repo *repository.WheelsetFitQuestionnaireRepository) *WheelsetFitQuestionnaireService {
	return &WheelsetFitQuestionnaireService{repo: repo}
}

// GetOrCreateDraft returns the existing draft unchanged. Only when the
// questionnaire has no draft does the repository deep-copy its published
// version into a new draft.
func (s *WheelsetFitQuestionnaireService) GetOrCreateDraft() (*wheelsetfit.Version, error) {
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
	return s.repo.GetOrCreateDraft(questionnaire.ID)
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
