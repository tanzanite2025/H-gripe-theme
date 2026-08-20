package service

import (
	"errors"
	"testing"

	selectionconfiguration "commerce-platform/internal/domain/selectionconfiguration"
	wheelsetfit "commerce-platform/internal/domain/wheelsetfit"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSelectionConfigurationKeyServiceCreatesListsAndKeepsCodeImmutable(t *testing.T) {
	db := newSelectionConfigurationKeyServiceTestDatabase(t)
	selectionConfigurationKeyRepository := repository.NewSelectionConfigurationKeyRepository(db)
	selectionConfigurationKeyService := NewSelectionConfigurationKeyService(selectionConfigurationKeyRepository)

	created, err := selectionConfigurationKeyService.CreateSelectionConfigurationKey(SelectionConfigurationKeyInput{
		Kind:         SelectionConfigurationKeyKindQuestionKey,
		Code:         "rear_axle",
		DisplayLabel: "后轴规格",
		Description:  "轮组选型问卷的后轴问题",
		IsEnabled:    true,
		SortOrder:    10,
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, "rear_axle", created.Code)
	loadedCreated, err := selectionConfigurationKeyRepository.FindSelectionConfigurationKeyByKindAndCode(SelectionConfigurationKeyKindQuestionKey, "rear_axle")
	require.NoError(t, err)
	assert.True(t, loadedCreated.IsEnabled)

	updated, err := selectionConfigurationKeyService.UpdateSelectionConfigurationKey(created.ID, SelectionConfigurationKeyInput{
		Kind:         SelectionConfigurationKeyKindQuestionKey,
		Code:         "front_axle",
		DisplayLabel: "前轴规格",
		Description:  "不能修改已有 Code",
		IsEnabled:    true,
		SortOrder:    20,
	})
	require.Error(t, err)
	assert.Nil(t, updated)
	assert.True(t, errors.Is(err, ErrSelectionConfigurationKeyCodeImmutable))

	updated, err = selectionConfigurationKeyService.UpdateSelectionConfigurationKey(created.ID, SelectionConfigurationKeyInput{
		Kind:         SelectionConfigurationKeyKindQuestionKey,
		Code:         "rear_axle",
		DisplayLabel: "后轴标准",
		Description:  "只允许修改展示信息",
		IsEnabled:    false,
		SortOrder:    20,
	})
	require.NoError(t, err)
	assert.Equal(t, "后轴标准", updated.DisplayLabel)
	assert.False(t, updated.IsEnabled)

	options, err := selectionConfigurationKeyService.ListEnabledSelectionConfigurationKeyOptions(SelectionConfigurationKeyKindQuestionKey)
	require.NoError(t, err)
	assert.Empty(t, options)
}

func TestWheelsetFitQuestionnaireRejectsUnregisteredAndDisabledSelectionConfigurationKeys(t *testing.T) {
	db := newSelectionConfigurationKeyServiceTestDatabase(t)
	require.NoError(t, db.AutoMigrate(
		&wheelsetfit.Questionnaire{},
		&wheelsetfit.Version{},
		&wheelsetfit.Question{},
		&wheelsetfit.Option{},
		&wheelsetfit.QuestionTranslation{},
		&wheelsetfit.OptionTranslation{},
	))
	questionnaire := wheelsetfit.Questionnaire{
		Slug:                wheelsetfit.QuestionnaireSlug,
		ProductCategorySlug: wheelsetfit.WheelsetProductCategorySlug,
		SourceLocale:        wheelsetfit.SourceLocale,
		IsEnabled:           true,
	}
	require.NoError(t, db.Create(&questionnaire).Error)

	selectionConfigurationKeyRepository := repository.NewSelectionConfigurationKeyRepository(db)
	selectionConfigurationKeyService := NewSelectionConfigurationKeyService(selectionConfigurationKeyRepository)
	wheelsetFitQuestionnaireService := NewWheelsetFitQuestionnaireService(
		repository.NewWheelsetFitQuestionnaireRepository(db),
		selectionConfigurationKeyRepository,
	)
	require.NotNil(t, wheelsetFitQuestionnaireService.selectionConfigurationKeyRepository)

	input := wheelsetFitQuestionInput(
		0,
		[]WheelsetFitQuestionTranslationInput{{Locale: "zh_cn", Prompt: "请选择后轴规格"}},
		[]WheelsetFitQuestionOptionTranslationInput{{Locale: "zh_cn", Label: "148 Boost"}},
	)
	_, err := wheelsetFitQuestionnaireService.SaveQuestion(input)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrWheelsetFitQuestionKeyNotRegistered))

	_, err = selectionConfigurationKeyService.CreateSelectionConfigurationKey(SelectionConfigurationKeyInput{
		Kind:         SelectionConfigurationKeyKindQuestionKey,
		Code:         "rear_axle",
		DisplayLabel: "后轴规格",
		IsEnabled:    false,
	})
	require.NoError(t, err)
	loadedDisabledQuestionKey, err := selectionConfigurationKeyRepository.FindSelectionConfigurationKeyByKindAndCode(SelectionConfigurationKeyKindQuestionKey, "rear_axle")
	require.NoError(t, err)
	assert.False(t, loadedDisabledQuestionKey.IsEnabled)
	err = wheelsetFitQuestionnaireService.validateWheelsetFitQuestionKeyAndAnswerKeyAreRegisteredAndEnabled("rear_axle", "rear_axle")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrWheelsetFitQuestionKeyDisabled))
	_, err = selectionConfigurationKeyService.CreateSelectionConfigurationKey(SelectionConfigurationKeyInput{
		Kind:         SelectionConfigurationKeyKindAnswerKey,
		Code:         "rear_axle",
		DisplayLabel: "后轴回答",
		IsEnabled:    true,
	})
	require.NoError(t, err)

	_, err = wheelsetFitQuestionnaireService.SaveQuestion(input)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrWheelsetFitQuestionKeyDisabled))

	questionKey, err := selectionConfigurationKeyRepository.FindSelectionConfigurationKeyByKindAndCode(
		SelectionConfigurationKeyKindQuestionKey,
		"rear_axle",
	)
	require.NoError(t, err)
	questionKey.IsEnabled = true
	require.NoError(t, selectionConfigurationKeyRepository.SaveSelectionConfigurationKey(questionKey))

	answerKey, err := selectionConfigurationKeyRepository.FindSelectionConfigurationKeyByKindAndCode(
		SelectionConfigurationKeyKindAnswerKey,
		"rear_axle",
	)
	require.NoError(t, err)
	answerKey.IsEnabled = false
	require.NoError(t, selectionConfigurationKeyRepository.SaveSelectionConfigurationKey(answerKey))

	_, err = wheelsetFitQuestionnaireService.SaveQuestion(input)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrWheelsetFitAnswerKeyDisabled))
}

func newSelectionConfigurationKeyServiceTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	require.NoError(t, db.AutoMigrate(&selectionconfiguration.SelectionConfigurationKey{}))
	return db
}
