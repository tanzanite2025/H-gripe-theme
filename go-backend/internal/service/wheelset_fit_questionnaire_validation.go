package service

import (
	"encoding/json"
	"fmt"
	"strings"

	wheelsetfit "commerce-platform/internal/domain/wheelsetfit"
	"commerce-platform/internal/pkg/locales"
)

func validateWheelsetFitQuestionnaireVersion(version wheelsetfit.Version) WheelsetFitQuestionnaireValidationResult {
	result := WheelsetFitQuestionnaireValidationResult{
		Valid:  true,
		Issues: []WheelsetFitQuestionnaireValidationIssue{},
	}
	addIssue := func(issue WheelsetFitQuestionnaireValidationIssue) {
		result.Issues = append(result.Issues, issue)
		if issue.Severity == "error" {
			result.Valid = false
		}
	}

	if version.Questionnaire == nil {
		addIssue(WheelsetFitQuestionnaireValidationIssue{
			Severity: "error",
			Code:     "missing_questionnaire",
			Message:  "问卷版本缺少所属问卷",
		})
		return result
	}
	if err := validateWheelsetFitQuestionnaire(*version.Questionnaire); err != nil {
		addIssue(WheelsetFitQuestionnaireValidationIssue{
			Severity: "error",
			Code:     "invalid_questionnaire",
			Message:  err.Error(),
		})
		return result
	}

	sourceLocale := version.Questionnaire.SourceLocale
	if len(version.Questions) == 0 {
		addIssue(WheelsetFitQuestionnaireValidationIssue{
			Severity: "error",
			Code:     "missing_questions",
			Message:  "至少需要一个问题才能发布问卷",
		})
		return result
	}

	questionKeys := make(map[string]struct{}, len(version.Questions))
	answerKeys := make(map[string]struct{}, len(version.Questions))
	sortOrders := make(map[int]struct{}, len(version.Questions))
	for _, question := range version.Questions {
		context := WheelsetFitQuestionnaireValidationIssue{
			QuestionID:  question.ID,
			QuestionKey: question.QuestionKey,
		}
		if normalizeWheelsetFitKey(question.QuestionKey) == "" {
			addIssue(withWheelsetFitQuestionnaireIssue(context, "error", "invalid_question_key", "问题 key 必须使用小写 snake_case"))
		} else if _, duplicated := questionKeys[question.QuestionKey]; duplicated {
			addIssue(withWheelsetFitQuestionnaireIssue(context, "error", "duplicate_question_key", fmt.Sprintf("问题 key %q 重复", question.QuestionKey)))
		} else {
			questionKeys[question.QuestionKey] = struct{}{}
		}
		if normalizeWheelsetFitKey(question.AnswerKey) == "" {
			addIssue(withWheelsetFitQuestionnaireIssue(context, "error", "invalid_answer_key", "回答 key 必须使用小写 snake_case"))
		} else if _, duplicated := answerKeys[question.AnswerKey]; duplicated {
			addIssue(withWheelsetFitQuestionnaireIssue(context, "error", "duplicate_answer_key", fmt.Sprintf("回答 key %q 重复", question.AnswerKey)))
		} else {
			answerKeys[question.AnswerKey] = struct{}{}
		}
		if question.SortOrder <= 0 {
			addIssue(withWheelsetFitQuestionnaireIssue(context, "error", "invalid_question_order", "问题排序必须为正数"))
		} else if _, duplicated := sortOrders[question.SortOrder]; duplicated {
			addIssue(withWheelsetFitQuestionnaireIssue(context, "error", "duplicate_question_order", fmt.Sprintf("问题排序 %d 重复", question.SortOrder)))
		} else {
			sortOrders[question.SortOrder] = struct{}{}
		}
		if question.InputMode != wheelsetfit.InputModeSingleChoice {
			addIssue(withWheelsetFitQuestionnaireIssue(context, "error", "invalid_input_mode", "当前问卷只支持单选问题"))
		}

		translations := wheelsetFitQuestionTranslationsByLocale(&question)
		source := translations[sourceLocale]
		if strings.TrimSpace(source.Prompt) == "" {
			addIssue(withWheelsetFitQuestionnaireIssue(context, "error", "missing_source_prompt", "基础语言的问题标题不能为空"))
		}
		if (strings.TrimSpace(source.HelpTitle) == "") != (strings.TrimSpace(source.HelpBody) == "") {
			addIssue(withWheelsetFitQuestionnaireIssue(context, "warning", "incomplete_source_help", "HELP 标题和说明建议同时填写"))
		}
		addWheelsetFitQuestionTranslationWarnings(addIssue, context, translations, sourceLocale, question.SourceRevision)

		enabledOptions := 0
		unknownOptions := 0
		optionKeys := make(map[string]struct{}, len(question.Options))
		optionOrders := make(map[int]struct{}, len(question.Options))
		for _, option := range question.Options {
			optionContext := context
			optionContext.OptionKey = option.OptionKey
			if normalizeWheelsetFitKey(option.OptionKey) == "" {
				addIssue(withWheelsetFitQuestionnaireIssue(optionContext, "error", "invalid_option_key", "选项 key 必须使用小写 snake_case"))
			} else if _, duplicated := optionKeys[option.OptionKey]; duplicated {
				addIssue(withWheelsetFitQuestionnaireIssue(optionContext, "error", "duplicate_option_key", fmt.Sprintf("选项 key %q 重复", option.OptionKey)))
			} else {
				optionKeys[option.OptionKey] = struct{}{}
			}
			if strings.TrimSpace(option.AnswerValue) == "" {
				addIssue(withWheelsetFitQuestionnaireIssue(optionContext, "error", "missing_answer_value", "选项回答值不能为空"))
			}
			if option.SortOrder <= 0 {
				addIssue(withWheelsetFitQuestionnaireIssue(optionContext, "error", "invalid_option_order", "选项排序必须为正数"))
			} else if _, duplicated := optionOrders[option.SortOrder]; duplicated {
				addIssue(withWheelsetFitQuestionnaireIssue(optionContext, "error", "duplicate_option_order", fmt.Sprintf("选项排序 %d 重复", option.SortOrder)))
			} else {
				optionOrders[option.SortOrder] = struct{}{}
			}
			if option.IsEnabled {
				enabledOptions++
			}
			if option.IsUnknown {
				unknownOptions++
				if !question.AllowUnknown {
					addIssue(withWheelsetFitQuestionnaireIssue(optionContext, "error", "unknown_option_not_allowed", "当前问题未启用“不确定”选项"))
				}
			}
			if err := validateWheelsetFitProductFilterEffects(option.ProductFilterEffects); err != nil {
				addIssue(withWheelsetFitQuestionnaireIssue(optionContext, "error", "invalid_product_filter_effects", err.Error()))
			}

			optionTranslations := wheelsetFitOptionTranslationsByLocale(&option)
			if option.IsEnabled && strings.TrimSpace(optionTranslations[sourceLocale].Label) == "" {
				addIssue(withWheelsetFitQuestionnaireIssue(optionContext, "error", "missing_source_option_label", "基础语言的选项名称不能为空"))
			}
			addWheelsetFitOptionTranslationWarnings(addIssue, optionContext, optionTranslations, sourceLocale, option.SourceRevision)
		}
		if question.IsEnabled && enabledOptions == 0 {
			addIssue(withWheelsetFitQuestionnaireIssue(context, "error", "missing_enabled_options", "启用的问题至少需要一个启用的选项"))
		}
		if unknownOptions > 1 {
			addIssue(withWheelsetFitQuestionnaireIssue(context, "error", "multiple_unknown_options", "每个问题只能有一个“不确定”选项"))
		}
	}

	return result
}

func withWheelsetFitQuestionnaireIssue(
	issue WheelsetFitQuestionnaireValidationIssue,
	severity string,
	code string,
	message string,
) WheelsetFitQuestionnaireValidationIssue {
	issue.Severity = severity
	issue.Code = code
	issue.Message = message
	return issue
}

func addWheelsetFitQuestionTranslationWarnings(
	addIssue func(WheelsetFitQuestionnaireValidationIssue),
	context WheelsetFitQuestionnaireValidationIssue,
	translations map[string]wheelsetfit.QuestionTranslation,
	sourceLocale string,
	sourceRevision int,
) {
	missing := 0
	outdated := 0
	for _, locale := range locales.EnabledLocaleCodes() {
		if locale == sourceLocale {
			continue
		}
		translation := translations[locale]
		if strings.TrimSpace(translation.Prompt) == "" {
			missing++
			continue
		}
		if translation.TranslatedRevision < sourceRevision {
			outdated++
		}
	}
	if missing > 0 {
		addIssue(withWheelsetFitQuestionnaireIssue(
			context,
			"warning",
			"missing_question_translations",
			fmt.Sprintf("问题有 %d 个语言尚未翻译", missing),
		))
	}
	if outdated > 0 {
		addIssue(withWheelsetFitQuestionnaireIssue(
			context,
			"warning",
			"outdated_question_translations",
			fmt.Sprintf("问题有 %d 个语言需要根据基础文案更新", outdated),
		))
	}
}

func addWheelsetFitOptionTranslationWarnings(
	addIssue func(WheelsetFitQuestionnaireValidationIssue),
	context WheelsetFitQuestionnaireValidationIssue,
	translations map[string]wheelsetfit.OptionTranslation,
	sourceLocale string,
	sourceRevision int,
) {
	missing := 0
	outdated := 0
	for _, locale := range locales.EnabledLocaleCodes() {
		if locale == sourceLocale {
			continue
		}
		translation := translations[locale]
		if strings.TrimSpace(translation.Label) == "" {
			missing++
			continue
		}
		if translation.TranslatedRevision < sourceRevision {
			outdated++
		}
	}
	if missing > 0 {
		addIssue(withWheelsetFitQuestionnaireIssue(
			context,
			"warning",
			"missing_option_translations",
			fmt.Sprintf("选项有 %d 个语言尚未翻译", missing),
		))
	}
	if outdated > 0 {
		addIssue(withWheelsetFitQuestionnaireIssue(
			context,
			"warning",
			"outdated_option_translations",
			fmt.Sprintf("选项有 %d 个语言需要根据基础文案更新", outdated),
		))
	}
}

func validateWheelsetFitProductFilterEffects(value []byte) error {
	raw := strings.TrimSpace(string(value))
	if raw == "" {
		return nil
	}
	var decoded any
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		return fmt.Errorf("商品筛选规则必须是有效 JSON")
	}
	if _, ok := decoded.(map[string]any); !ok {
		return fmt.Errorf("商品筛选规则必须是 JSON 对象")
	}
	return nil
}
