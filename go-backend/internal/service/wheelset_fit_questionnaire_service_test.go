package service

import (
	"encoding/json"
	"errors"
	"testing"

	wheelsetfit "commerce-platform/internal/domain/wheelsetfit"
	"commerce-platform/internal/pkg/locales"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestWheelsetFitQuestionnaireReusesExistingDraft(t *testing.T) {
	db, service, _ := newWheelsetFitQuestionnaireTestService(t)
	questionnaire := seedWheelsetFitQuestionnaire(t, db)
	published := wheelsetfit.Version{
		QuestionnaireID: questionnaire.ID,
		VersionNumber:   1,
		Status:          wheelsetfit.VersionStatusPublished,
	}
	require.NoError(t, db.Create(&published).Error)
	draft := wheelsetfit.Version{
		QuestionnaireID: questionnaire.ID,
		VersionNumber:   2,
		Status:          wheelsetfit.VersionStatusDraft,
	}
	require.NoError(t, db.Create(&draft).Error)

	loaded, err := service.GetOrCreateDraft()

	require.NoError(t, err)
	require.Equal(t, draft.ID, loaded.ID)
	assert.Equal(t, 2, loaded.VersionNumber)
	var versionCount int64
	require.NoError(t, db.Model(&wheelsetfit.Version{}).Where("questionnaire_id = ?", questionnaire.ID).Count(&versionCount).Error)
	assert.Equal(t, int64(2), versionCount)
}

func TestWheelsetFitQuestionnaireReturnsEmptyQuestionsArrayForEmptyDraft(t *testing.T) {
	db, service, _ := newWheelsetFitQuestionnaireTestService(t)
	seedWheelsetFitQuestionnaire(t, db)

	draft, err := service.GetOrCreateDraft()
	require.NoError(t, err)
	require.NotNil(t, draft.Questions)
	assert.Empty(t, draft.Questions)

	payload, err := json.Marshal(draft)
	require.NoError(t, err)
	assert.Contains(t, string(payload), `"questions":[]`)
}

func TestWheelsetFitQuestionnaireReturnsEmptyOptionsArrayForQuestionWithoutOptions(t *testing.T) {
	db, service, _ := newWheelsetFitQuestionnaireTestService(t)
	seedWheelsetFitQuestionnaire(t, db)

	version, err := service.SaveQuestion(WheelsetFitQuestionInput{
		QuestionKey:  "rear_axle",
		AnswerKey:    "rear_axle",
		SortOrder:    10,
		InputMode:    wheelsetfit.InputModeSingleChoice,
		IsRequired:   true,
		AllowUnknown: true,
		IsEnabled:    true,
		Translations: []WheelsetFitQuestionTranslationInput{
			{Locale: "zh_cn", Prompt: "请选择后轴规格"},
		},
		Options: []WheelsetFitQuestionOptionInput{},
	})
	require.NoError(t, err)
	require.Len(t, version.Questions, 1)
	require.NotNil(t, version.Questions[0].Options)
	assert.Empty(t, version.Questions[0].Options)

	payload, err := json.Marshal(version)
	require.NoError(t, err)
	assert.Contains(t, string(payload), `"options":[]`)
}

func TestWheelsetFitQuestionnaireCreatesDraftByDeepCopyingPublishedVersion(t *testing.T) {
	db, service, repo := newWheelsetFitQuestionnaireTestService(t)
	questionnaire := seedWheelsetFitQuestionnaire(t, db)
	published := wheelsetfit.Version{
		QuestionnaireID: questionnaire.ID,
		VersionNumber:   1,
		Status:          wheelsetfit.VersionStatusPublished,
	}
	require.NoError(t, db.Create(&published).Error)

	publishedQuestion := wheelsetfit.Question{
		QuestionnaireVersionID: published.ID,
		QuestionKey:            "rear_axle",
		AnswerKey:              "rear_axle",
		SortOrder:              10,
		InputMode:              wheelsetfit.InputModeSingleChoice,
		IsRequired:             true,
		AllowUnknown:           true,
		IsEnabled:              true,
		SourceRevision:         4,
		Translations: []wheelsetfit.QuestionTranslation{
			{Locale: "zh_cn", Prompt: "后轴规格", SourceRevision: 4, TranslatedRevision: 4},
			{Locale: "en", Prompt: "Rear axle", SourceRevision: 4, TranslatedRevision: 4},
		},
		Options: []wheelsetfit.Option{
			{
				OptionKey:            "boost_148",
				AnswerValue:          "12x148",
				SortOrder:            10,
				IsEnabled:            true,
				ProductFilterEffects: datatypes.JSON([]byte(`{"spec_filters":{"rear_axle":["12x148"]}}`)),
				SourceRevision:       3,
				Translations: []wheelsetfit.OptionTranslation{
					{Locale: "zh_cn", Label: "148 Boost", SourceRevision: 3, TranslatedRevision: 3},
					{Locale: "en", Label: "148 Boost", SourceRevision: 3, TranslatedRevision: 3},
				},
			},
		},
	}
	// Historical published content is seeded directly; repository mutations
	// intentionally accept draft versions only.
	require.NoError(t, db.Create(&publishedQuestion).Error)

	draft, err := service.GetOrCreateDraft()

	require.NoError(t, err)
	assert.Equal(t, wheelsetfit.VersionStatusDraft, draft.Status)
	assert.Equal(t, 2, draft.VersionNumber)
	require.Len(t, draft.Questions, 1)
	draftQuestion := draft.Questions[0]
	assert.NotEqual(t, publishedQuestion.ID, draftQuestion.ID)
	require.Len(t, draftQuestion.Translations, 2)
	assert.NotEqual(t, publishedQuestion.Translations[0].ID, draftQuestion.Translations[0].ID)
	require.Len(t, draftQuestion.Options, 1)
	assert.NotEqual(t, publishedQuestion.Options[0].ID, draftQuestion.Options[0].ID)
	assert.NotEqual(t, publishedQuestion.Options[0].Translations[0].ID, draftQuestion.Options[0].Translations[0].ID)
	assert.Equal(t, "后轴规格", wheelsetFitQuestionTranslation(t, draftQuestion, "zh_cn").Prompt)
	assert.Equal(t, "148 Boost", wheelsetFitOptionTranslation(t, draftQuestion.Options[0], "en").Label)

	updated, err := service.SaveQuestion(wheelsetFitQuestionInput(
		draftQuestion.ID,
		[]WheelsetFitQuestionTranslationInput{{Locale: "zh_cn", Prompt: "更新后的后轴规格"}},
		[]WheelsetFitQuestionOptionTranslationInput{{Locale: "zh_cn", Label: "148 Boost"}},
	))
	require.NoError(t, err)
	assert.Equal(t, "更新后的后轴规格", wheelsetFitQuestionTranslation(t, *wheelsetFitQuestionByID(t, updated, draftQuestion.ID), "zh_cn").Prompt)

	reloadedPublished, err := repo.FindVersionByID(published.ID)
	require.NoError(t, err)
	assert.Equal(t, "后轴规格", wheelsetFitQuestionTranslation(t, reloadedPublished.Questions[0], "zh_cn").Prompt)
}

func TestWheelsetFitQuestionnaireSaveQuestionPreservesLocalesAndTracksSourceRevisions(t *testing.T) {
	db, service, _ := newWheelsetFitQuestionnaireTestService(t)
	seedWheelsetFitQuestionnaire(t, db)

	created, err := service.SaveQuestion(wheelsetFitQuestionInput(
		0,
		[]WheelsetFitQuestionTranslationInput{
			{Locale: "zh_cn", Prompt: "请选择后轴规格", HelpTitle: "为什么需要这个？", HelpBody: "后轴必须匹配车架。"},
			{Locale: "en", Prompt: "Choose your rear axle", HelpTitle: "Why do we need this?", HelpBody: "The rear axle must match the frame."},
		},
		[]WheelsetFitQuestionOptionTranslationInput{
			{Locale: "zh_cn", Label: "12x148 Boost", Description: "常见山地后轴标准"},
			{Locale: "en", Label: "12x148 Boost", Description: "Common mountain-bike rear axle standard"},
		},
	))
	require.NoError(t, err)
	require.Len(t, created.Questions, 1)
	question := created.Questions[0]
	option := question.Options[0]
	assert.Equal(t, 1, question.SourceRevision)
	assert.Equal(t, 1, option.SourceRevision)
	assert.Len(t, question.Translations, len(locales.EnabledLocaleCodes()))
	assert.Len(t, option.Translations, len(locales.EnabledLocaleCodes()))
	assert.Equal(t, "Choose your rear axle", wheelsetFitQuestionTranslation(t, question, "en").Prompt)
	assert.Equal(t, "12x148 Boost", wheelsetFitOptionTranslation(t, option, "en").Label)
	assert.Equal(t, 1, wheelsetFitQuestionTranslation(t, question, "en").TranslatedRevision)
	assert.Equal(t, 1, wheelsetFitOptionTranslation(t, option, "en").TranslatedRevision)

	afterQuestionSourceChange, err := service.SaveQuestion(wheelsetFitQuestionInput(
		question.ID,
		[]WheelsetFitQuestionTranslationInput{{Locale: "zh_cn", Prompt: "请选择您的后轴规格", HelpTitle: "为什么需要这个？", HelpBody: "后轴必须匹配车架。"}},
		[]WheelsetFitQuestionOptionTranslationInput{{Locale: "zh_cn", Label: "12x148 Boost", Description: "常见山地后轴标准"}},
	))
	require.NoError(t, err)
	question = *wheelsetFitQuestionByID(t, afterQuestionSourceChange, question.ID)
	option = question.Options[0]
	assert.Equal(t, 2, question.SourceRevision)
	assert.Equal(t, 1, option.SourceRevision)
	assert.Equal(t, "Choose your rear axle", wheelsetFitQuestionTranslation(t, question, "en").Prompt)
	assert.Equal(t, "The rear axle must match the frame.", wheelsetFitQuestionTranslation(t, question, "en").HelpBody)
	assert.Equal(t, 2, wheelsetFitQuestionTranslation(t, question, "en").SourceRevision)
	assert.Equal(t, 1, wheelsetFitQuestionTranslation(t, question, "en").TranslatedRevision)
	assert.Len(t, question.Translations, len(locales.EnabledLocaleCodes()))

	afterOptionSourceChange, err := service.SaveQuestion(wheelsetFitQuestionInput(
		question.ID,
		[]WheelsetFitQuestionTranslationInput{{Locale: "zh_cn", Prompt: "请选择您的后轴规格", HelpTitle: "为什么需要这个？", HelpBody: "后轴必须匹配车架。"}},
		[]WheelsetFitQuestionOptionTranslationInput{{Locale: "zh_cn", Label: "12x148 Boost 后轴", Description: "常见山地后轴标准"}},
	))
	require.NoError(t, err)
	question = *wheelsetFitQuestionByID(t, afterOptionSourceChange, question.ID)
	option = question.Options[0]
	assert.Equal(t, 2, question.SourceRevision)
	assert.Equal(t, 2, option.SourceRevision)
	assert.Equal(t, "12x148 Boost", wheelsetFitOptionTranslation(t, option, "en").Label)
	assert.Equal(t, 2, wheelsetFitOptionTranslation(t, option, "en").SourceRevision)
	assert.Equal(t, 1, wheelsetFitOptionTranslation(t, option, "en").TranslatedRevision)
	assert.Len(t, option.Translations, len(locales.EnabledLocaleCodes()))

	afterEnglishSave, err := service.SaveQuestion(wheelsetFitQuestionInput(
		question.ID,
		[]WheelsetFitQuestionTranslationInput{{Locale: "en", Prompt: "Choose your rear axle standard", HelpTitle: "Why do we need this?", HelpBody: "The rear axle must match the frame."}},
		[]WheelsetFitQuestionOptionTranslationInput{{Locale: "en", Label: "12x148 Boost axle", Description: "Common mountain-bike rear axle standard"}},
	))
	require.NoError(t, err)
	question = *wheelsetFitQuestionByID(t, afterEnglishSave, question.ID)
	option = question.Options[0]
	assert.Equal(t, 2, question.SourceRevision)
	assert.Equal(t, 2, option.SourceRevision)
	assert.Equal(t, "请选择您的后轴规格", wheelsetFitQuestionTranslation(t, question, "zh_cn").Prompt)
	assert.Equal(t, "12x148 Boost 后轴", wheelsetFitOptionTranslation(t, option, "zh_cn").Label)
	assert.Equal(t, 2, wheelsetFitQuestionTranslation(t, question, "en").TranslatedRevision)
	assert.Equal(t, 2, wheelsetFitOptionTranslation(t, option, "en").TranslatedRevision)
}

func TestWheelsetFitQuestionnaireSaveQuestionPreservesOptionsWhenOmitted(t *testing.T) {
	db, service, _ := newWheelsetFitQuestionnaireTestService(t)
	seedWheelsetFitQuestionnaire(t, db)

	created, err := service.SaveQuestion(wheelsetFitQuestionInput(
		0,
		[]WheelsetFitQuestionTranslationInput{{Locale: "zh_cn", Prompt: "请选择后轴规格"}},
		[]WheelsetFitQuestionOptionTranslationInput{{Locale: "zh_cn", Label: "12x148 Boost"}},
	))
	require.NoError(t, err)
	question := created.Questions[0]
	optionID := question.Options[0].ID

	updated, err := service.SaveQuestion(WheelsetFitQuestionInput{
		ID:           question.ID,
		QuestionKey:  question.QuestionKey,
		AnswerKey:    question.AnswerKey,
		SortOrder:    question.SortOrder,
		InputMode:    question.InputMode,
		IsRequired:   question.IsRequired,
		AllowUnknown: question.AllowUnknown,
		IsEnabled:    question.IsEnabled,
		Translations: []WheelsetFitQuestionTranslationInput{{Locale: "en", Prompt: "Choose your rear axle"}},
	})
	require.NoError(t, err)
	question = *wheelsetFitQuestionByID(t, updated, question.ID)
	require.Len(t, question.Options, 1)
	assert.Equal(t, optionID, question.Options[0].ID)
	assert.Equal(t, "12x148", question.Options[0].AnswerValue)
	assert.Len(t, question.Options[0].Translations, len(locales.EnabledLocaleCodes()))
}

func TestWheelsetFitQuestionnaireSaveQuestionClearsOptionsWhenExplicitlyEmpty(t *testing.T) {
	db, service, _ := newWheelsetFitQuestionnaireTestService(t)
	seedWheelsetFitQuestionnaire(t, db)

	created, err := service.SaveQuestion(wheelsetFitQuestionInput(
		0,
		[]WheelsetFitQuestionTranslationInput{{Locale: "zh_cn", Prompt: "请选择后轴规格"}},
		[]WheelsetFitQuestionOptionTranslationInput{{Locale: "zh_cn", Label: "12x148 Boost"}},
	))
	require.NoError(t, err)
	question := created.Questions[0]

	updated, err := service.SaveQuestion(WheelsetFitQuestionInput{
		ID:           question.ID,
		QuestionKey:  question.QuestionKey,
		AnswerKey:    question.AnswerKey,
		SortOrder:    question.SortOrder,
		InputMode:    question.InputMode,
		IsRequired:   question.IsRequired,
		AllowUnknown: question.AllowUnknown,
		IsEnabled:    question.IsEnabled,
		Options:      []WheelsetFitQuestionOptionInput{},
	})
	require.NoError(t, err)
	question = *wheelsetFitQuestionByID(t, updated, question.ID)
	assert.Empty(t, question.Options)

	var optionCount int64
	require.NoError(t, db.Model(&wheelsetfit.Option{}).Where("question_id = ?", question.ID).Count(&optionCount).Error)
	assert.Zero(t, optionCount)
}

func TestWheelsetFitQuestionnaireReordersAndDeletesDraftQuestions(t *testing.T) {
	db, service, _ := newWheelsetFitQuestionnaireTestService(t)
	seedWheelsetFitQuestionnaire(t, db)

	first, err := service.SaveQuestion(wheelsetFitQuestionInput(
		0,
		[]WheelsetFitQuestionTranslationInput{{Locale: "zh_cn", Prompt: "请选择后轴规格"}},
		[]WheelsetFitQuestionOptionTranslationInput{{Locale: "zh_cn", Label: "148 Boost"}},
	))
	require.NoError(t, err)
	firstQuestion := first.Questions[0]

	secondInput := wheelsetFitQuestionInput(
		0,
		[]WheelsetFitQuestionTranslationInput{{Locale: "zh_cn", Prompt: "请选择前轴规格"}},
		[]WheelsetFitQuestionOptionTranslationInput{{Locale: "zh_cn", Label: "110 Boost"}},
	)
	secondInput.QuestionKey = "front_axle"
	secondInput.AnswerKey = "front_axle"
	secondInput.SortOrder = 20
	secondInput.Options[0].OptionKey = "boost_110"
	secondInput.Options[0].AnswerValue = "15x110"
	secondInput.Options[0].SortOrder = 10
	second, err := service.SaveQuestion(secondInput)
	require.NoError(t, err)
	secondQuestion := wheelsetFitQuestionByID(t, second, second.Questions[1].ID)

	reordered, err := service.ReorderQuestions(WheelsetFitQuestionOrderInput{
		QuestionIDs: []uint{secondQuestion.ID, firstQuestion.ID},
	})
	require.NoError(t, err)
	require.Len(t, reordered.Questions, 2)
	assert.Equal(t, secondQuestion.ID, reordered.Questions[0].ID)
	assert.Equal(t, 10, reordered.Questions[0].SortOrder)
	assert.Equal(t, firstQuestion.ID, reordered.Questions[1].ID)
	assert.Equal(t, 20, reordered.Questions[1].SortOrder)

	deleted, err := service.DeleteQuestion(secondQuestion.ID)
	require.NoError(t, err)
	require.Len(t, deleted.Questions, 1)
	assert.Equal(t, firstQuestion.ID, deleted.Questions[0].ID)
}

func TestWheelsetFitQuestionnaireValidatesBeforePublishing(t *testing.T) {
	db, service, _ := newWheelsetFitQuestionnaireTestService(t)
	seedWheelsetFitQuestionnaire(t, db)

	emptyDraft, err := service.GetOrCreateDraft()
	require.NoError(t, err)
	invalid, err := service.ValidateVersion(emptyDraft.ID)
	require.NoError(t, err)
	assert.False(t, invalid.Valid)
	assert.Equal(t, "missing_questions", invalid.Issues[0].Code)

	_, err = service.PublishVersion(emptyDraft.ID, nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrWheelsetFitQuestionnaireInvalid))

	saved, err := service.SaveQuestion(wheelsetFitQuestionInput(
		0,
		[]WheelsetFitQuestionTranslationInput{{Locale: "zh_cn", Prompt: "请选择后轴规格"}},
		[]WheelsetFitQuestionOptionTranslationInput{{Locale: "zh_cn", Label: "148 Boost"}},
	))
	require.NoError(t, err)

	valid, err := service.ValidateVersion(saved.ID)
	require.NoError(t, err)
	assert.True(t, valid.Valid)

	published, err := service.PublishVersion(saved.ID, nil)
	require.NoError(t, err)
	assert.Equal(t, wheelsetfit.VersionStatusPublished, published.Status)

	current, err := service.GetCurrentVersion()
	require.NoError(t, err)
	assert.Equal(t, published.ID, current.ID)
	assert.Equal(t, wheelsetfit.VersionStatusPublished, current.Status)
}

func newWheelsetFitQuestionnaireTestService(t *testing.T) (*gorm.DB, *WheelsetFitQuestionnaireService, *repository.WheelsetFitQuestionnaireRepository) {
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

	require.NoError(t, db.AutoMigrate(
		&wheelsetfit.Questionnaire{},
		&wheelsetfit.Version{},
		&wheelsetfit.Question{},
		&wheelsetfit.Option{},
		&wheelsetfit.QuestionTranslation{},
		&wheelsetfit.OptionTranslation{},
	))
	repo := repository.NewWheelsetFitQuestionnaireRepository(db)
	return db, NewWheelsetFitQuestionnaireService(repo), repo
}

func seedWheelsetFitQuestionnaire(t *testing.T, db *gorm.DB) wheelsetfit.Questionnaire {
	t.Helper()
	questionnaire := wheelsetfit.Questionnaire{
		Slug:                wheelsetfit.QuestionnaireSlug,
		ProductCategorySlug: wheelsetfit.WheelsetProductCategorySlug,
		SourceLocale:        wheelsetfit.SourceLocale,
		IsEnabled:           true,
	}
	require.NoError(t, db.Create(&questionnaire).Error)
	return questionnaire
}

func wheelsetFitQuestionInput(id uint, translations []WheelsetFitQuestionTranslationInput, optionTranslations []WheelsetFitQuestionOptionTranslationInput) WheelsetFitQuestionInput {
	return WheelsetFitQuestionInput{
		ID:           id,
		QuestionKey:  "rear_axle",
		AnswerKey:    "rear_axle",
		SortOrder:    10,
		InputMode:    wheelsetfit.InputModeSingleChoice,
		IsRequired:   true,
		AllowUnknown: true,
		IsEnabled:    true,
		Translations: translations,
		Options: []WheelsetFitQuestionOptionInput{
			{
				OptionKey:            "boost_148",
				AnswerValue:          "12x148",
				SortOrder:            10,
				IsEnabled:            true,
				ProductFilterEffects: datatypes.JSON([]byte(`{"spec_filters":{"rear_axle":["12x148"]}}`)),
				Translations:         optionTranslations,
			},
		},
	}
}

func wheelsetFitQuestionByID(t *testing.T, version *wheelsetfit.Version, id uint) *wheelsetfit.Question {
	t.Helper()
	for index := range version.Questions {
		if version.Questions[index].ID == id {
			return &version.Questions[index]
		}
	}
	t.Fatalf("question %d not found in version %d", id, version.ID)
	return nil
}

func wheelsetFitQuestionTranslation(t *testing.T, question wheelsetfit.Question, locale string) wheelsetfit.QuestionTranslation {
	t.Helper()
	for _, translation := range question.Translations {
		if translation.Locale == locale {
			return translation
		}
	}
	t.Fatalf("question translation %q not found", locale)
	return wheelsetfit.QuestionTranslation{}
}

func wheelsetFitOptionTranslation(t *testing.T, option wheelsetfit.Option, locale string) wheelsetfit.OptionTranslation {
	t.Helper()
	for _, translation := range option.Translations {
		if translation.Locale == locale {
			return translation
		}
	}
	t.Fatalf("option translation %q not found", locale)
	return wheelsetfit.OptionTranslation{}
}
