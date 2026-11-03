package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"

	"github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"commerce-platform/internal/domain/selectionassistant"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

// This order preserves parent-first normal FK enforcement. Runtime sessions and
// their selections are not configuration and never become writable here.
func configurationModuleMissingSelectionFlowAndQuestionnaireWritableTableNames() []string {
	return []string{
		"selection_configuration_keys", "selection_assistant_flows", "selection_assistant_flow_versions",
		"quick_buy_flows", "quick_buy_flow_versions", "quick_buy_flow_translations", "quick_buy_steps",
		"quick_buy_step_filters", "quick_buy_step_product_categories", "quick_buy_step_product_specification_templates", "quick_buy_compatibility_rules",
		"wheelset_fit_questionnaires", "wheelset_fit_questionnaire_versions", "wheelset_fit_questions",
		"wheelset_fit_question_translations", "wheelset_fit_question_options", "wheelset_fit_question_option_translations",
	}
}

var configurationModuleRestoredSelectionConfigurationKeyPattern = regexp.MustCompile(`^[a-z0-9]+(_[a-z0-9]+)*$`)

func validateConfigurationModuleMissingSelectionConfigurationKeyRows(missing map[string][]json.RawMessage) error {
	for _, row := range missing["selection_configuration_keys"] {
		kind, kindErr := readBusinessModuleSnapshotRowColumnText(row, "kind")
		code, codeErr := readBusinessModuleSnapshotRowColumnText(row, "code")
		if kindErr != nil || codeErr != nil || (kind != "question_key" && kind != "answer_key") || !configurationModuleRestoredSelectionConfigurationKeyPattern.MatchString(code) {
			return errors.New("选型配置 Key 的类别或编码无效")
		}
	}
	return nil
}

// Resolve changed children in the final SQL graph, so adding just a translation,
// option or association also validates its existing published parent. Incomplete
// drafts and archived snapshots retain their original status; nothing publishes.
func configurationModuleAffectedPublishedSelectionVersionIDs(ctx context.Context, tx *sql.Tx, missing map[string][]json.RawMessage, versionTable string, selectors map[string]struct{ column, joins, predicate string }) ([]uint, error) {
	identities := map[uint]bool{}
	for table, selector := range selectors {
		values := []string{}
		for _, row := range missing[table] {
			value, err := readBusinessModuleSnapshotRowColumnText(row, selector.column)
			if err != nil || value == "" {
				return nil, fmt.Errorf("选购流程或问卷缺少完整关联身份：%s", table)
			}
			values = append(values, value)
		}
		if len(values) == 0 {
			continue
		}
		query := `SELECT DISTINCT v.id FROM public.` + pq.QuoteIdentifier(versionTable) + ` v ` + selector.joins + ` WHERE v.status='published' AND (` + selector.predicate + `)` // #nosec G202 -- Table identifiers are quoted, join and predicate fragments come from the fixed internal selector registry, and all row values use bound parameters.
		rows, err := tx.QueryContext(ctx, query, pq.Array(values))
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var id uint
			if err := rows.Scan(&id); err != nil {
				_ = rows.Close()
				return nil, err
			}
			identities[id] = true
		}
		err = rows.Err()
		_ = rows.Close()
		if err != nil {
			return nil, err
		}
	}
	ids := make([]uint, 0, len(identities))
	for id := range identities {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids, nil
}

func validateConfigurationModuleRestoredSelectionFlowsAndQuestionnairesAgainstPublishedBusinessRules(ctx context.Context, tx *sql.Tx, missing map[string][]json.RawMessage) error {
	if err := validateConfigurationModuleMissingSelectionConfigurationKeyRows(missing); err != nil {
		return err
	}
	selectionIDs, err := configurationModuleAffectedPublishedSelectionVersionIDs(ctx, tx, missing, "selection_assistant_flow_versions", map[string]struct{ column, joins, predicate string }{
		"selection_assistant_flows":         {"id", "", "v.flow_id::text=ANY($1)"},
		"selection_assistant_flow_versions": {"id", "", "v.id::text=ANY($1)"},
	})
	if err != nil {
		return err
	}
	quickBuyIDs, err := configurationModuleAffectedPublishedSelectionVersionIDs(ctx, tx, missing, "quick_buy_flow_versions", map[string]struct{ column, joins, predicate string }{
		"quick_buy_flows":                                {"id", "", "v.flow_id::text=ANY($1)"},
		"quick_buy_flow_translations":                    {"flow_id", "", "v.flow_id::text=ANY($1)"},
		"quick_buy_flow_versions":                        {"id", "", "v.id::text=ANY($1)"},
		"quick_buy_steps":                                {"flow_version_id", "", "v.id::text=ANY($1)"},
		"quick_buy_compatibility_rules":                  {"flow_version_id", "", "v.id::text=ANY($1)"},
		"quick_buy_step_filters":                         {"step_id", "JOIN public.quick_buy_steps s ON s.flow_version_id=v.id", "s.id::text=ANY($1)"},
		"quick_buy_step_product_categories":              {"step_id", "JOIN public.quick_buy_steps s ON s.flow_version_id=v.id", "s.id::text=ANY($1)"},
		"quick_buy_step_product_specification_templates": {"step_id", "JOIN public.quick_buy_steps s ON s.flow_version_id=v.id", "s.id::text=ANY($1)"},
	})
	if err != nil {
		return err
	}
	questionnaireIDs, err := configurationModuleAffectedPublishedSelectionVersionIDs(ctx, tx, missing, "wheelset_fit_questionnaire_versions", map[string]struct{ column, joins, predicate string }{
		"wheelset_fit_questionnaires":               {"id", "", "v.questionnaire_id::text=ANY($1)"},
		"wheelset_fit_questionnaire_versions":       {"id", "", "v.id::text=ANY($1)"},
		"wheelset_fit_questions":                    {"questionnaire_version_id", "", "v.id::text=ANY($1)"},
		"wheelset_fit_question_translations":        {"question_id", "JOIN public.wheelset_fit_questions q ON q.questionnaire_version_id=v.id", "q.id::text=ANY($1)"},
		"wheelset_fit_question_options":             {"question_id", "JOIN public.wheelset_fit_questions q ON q.questionnaire_version_id=v.id", "q.id::text=ANY($1)"},
		"wheelset_fit_question_option_translations": {"option_id", "JOIN public.wheelset_fit_questions q ON q.questionnaire_version_id=v.id JOIN public.wheelset_fit_question_options o ON o.question_id=q.id", "o.id::text=ANY($1)"},
		"selection_configuration_keys":              {"id", "JOIN public.wheelset_fit_questions q ON q.questionnaire_version_id=v.id JOIN public.selection_configuration_keys k ON (k.kind='question_key' AND k.code=q.question_key) OR (k.kind='answer_key' AND k.code=q.answer_key)", "k.id::text=ANY($1)"},
	})
	if err != nil || len(selectionIDs)+len(quickBuyIDs)+len(questionnaireIDs) == 0 {
		return err
	}
	// All three validators only read through this exact transaction. No create,
	// save, get-or-create, publish, session, payment or notification method runs.
	database, err := gorm.Open(postgres.New(postgres.Config{Conn: tx}), &gorm.Config{DisableAutomaticPing: true, SkipDefaultTransaction: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return err
	}
	database = database.WithContext(ctx)
	selectionRepository := repository.NewSelectionAssistantRepository(database)
	selectionService := service.NewSelectionAssistantService(selectionRepository)
	for _, id := range selectionIDs {
		result, err := selectionService.ValidateVersion(id)
		if err != nil || !result.Valid {
			return errors.New("选购助手已发布版本的节点、跳转或筛选规则无效")
		}
		version, err := selectionRepository.FindVersionByID(id)
		if err != nil {
			return err
		}
		var config selectionassistant.Config
		if err := json.Unmarshal(version.Config, &config); err != nil {
			return errors.New("选购助手已发布版本 JSON 格式无效")
		}
		for _, node := range config.Nodes {
			for _, option := range node.Options {
				if err := validateConfigurationModulePublishedSelectionProductFilterReferences(ctx, tx, option.QueryEffects); err != nil {
					return err
				}
			}
		}
	}
	quickBuyService := service.NewQuickBuyService(repository.NewQuickBuyRepository(database), nil, nil)
	for _, id := range quickBuyIDs {
		result, err := quickBuyService.ValidateVersion(id)
		if err != nil || !result.Valid {
			return errors.New("快捷选购已发布版本的步骤、商品目录关联或规则引用无效")
		}
	}
	questionnaireRepository := repository.NewWheelsetFitQuestionnaireRepository(database)
	questionnaireService := service.NewWheelsetFitQuestionnaireService(questionnaireRepository, repository.NewSelectionConfigurationKeyRepository(database))
	for _, id := range questionnaireIDs {
		result, err := questionnaireService.ValidateVersion(id)
		if err != nil || !result.Valid {
			return errors.New("轮组选型已发布问卷的 Key、问题、选项或基础语言翻译无效")
		}
		publicFlow, err := questionnaireService.GetPublishedFlow()
		if err != nil || publicFlow.Version.ID != id {
			return errors.New("轮组选型已发布问卷无法生成前台流程")
		}
		if _, result := service.ValidateConfig(publicFlow.Version.Config); !result.Valid {
			return errors.New("轮组选型问卷生成的前台节点或筛选规则无效")
		}
		version, err := questionnaireRepository.FindVersionByID(id)
		if err != nil {
			return err
		}
		for _, question := range version.Questions {
			for _, option := range question.Options {
				var effects selectionassistant.QueryEffects
				if err := json.Unmarshal(option.ProductFilterEffects, &effects); err != nil {
					return errors.New("轮组选型问卷的商品筛选 JSON 格式无效")
				}
				if err := validateConfigurationModulePublishedSelectionProductFilterReferences(ctx, tx, effects); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// JSON filter slugs are semantic references to the current product catalog,
// not foreign keys. They must resolve to an enabled, filterable definition;
// the configuration writer cannot recreate or overwrite the product catalog.
func validateConfigurationModulePublishedSelectionProductFilterReferences(ctx context.Context, tx *sql.Tx, effects selectionassistant.QueryEffects) error {
	for slug, values := range effects.SpecFilters {
		if len(values) == 0 {
			continue
		}
		var available bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM public.product_spec_definitions d JOIN public.product_specification_templates t ON t.id=d.product_specification_template_id WHERE d.slug=$1 AND d.is_filterable AND t.is_enabled)`, slug).Scan(&available); err != nil {
			return err
		}
		if !available {
			return errors.New("选购流程或问卷引用的商品筛选规格缺失、停用或不可筛选，请先核对商品目录")
		}
	}
	return nil
}
