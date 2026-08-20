package service

import (
	"errors"
	"fmt"

	selectionconfiguration "commerce-platform/internal/domain/selectionconfiguration"
	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

func (s *WheelsetFitQuestionnaireService) validateWheelsetFitQuestionKeyAndAnswerKeyAreRegisteredAndEnabled(
	questionKey string,
	answerKey string,
) error {
	if s.selectionConfigurationKeyRepository == nil {
		return nil
	}

	normalizedQuestionKey := normalizeWheelsetFitKey(questionKey)
	normalizedAnswerKey := normalizeWheelsetFitKey(answerKey)
	if normalizedQuestionKey == "" || normalizedAnswerKey == "" {
		return nil
	}

	if err := validateWheelsetFitQuestionnaireKeyIsRegisteredAndEnabled(
		s.selectionConfigurationKeyRepository,
		selectionconfiguration.SelectionConfigurationKeyKindQuestionKey,
		normalizedQuestionKey,
		ErrWheelsetFitQuestionKeyNotRegistered,
		ErrWheelsetFitQuestionKeyDisabled,
	); err != nil {
		return err
	}

	return validateWheelsetFitQuestionnaireKeyIsRegisteredAndEnabled(
		s.selectionConfigurationKeyRepository,
		selectionconfiguration.SelectionConfigurationKeyKindAnswerKey,
		normalizedAnswerKey,
		ErrWheelsetFitAnswerKeyNotRegistered,
		ErrWheelsetFitAnswerKeyDisabled,
	)
}

func validateWheelsetFitQuestionnaireKeyIsRegisteredAndEnabled(
	selectionConfigurationKeyRepository *repository.SelectionConfigurationKeyRepository,
	keyKind string,
	keyCode string,
	notRegisteredError error,
	disabledError error,
) error {
	selectionConfigurationKey, err := selectionConfigurationKeyRepository.FindSelectionConfigurationKeyByKindAndCode(keyKind, keyCode)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("%w: %s", notRegisteredError, keyCode)
	}
	if err != nil {
		return fmt.Errorf("load selection configuration key %q: %w", keyCode, err)
	}
	if !selectionConfigurationKey.IsEnabled {
		return fmt.Errorf("%w: %s", disabledError, keyCode)
	}
	return nil
}
