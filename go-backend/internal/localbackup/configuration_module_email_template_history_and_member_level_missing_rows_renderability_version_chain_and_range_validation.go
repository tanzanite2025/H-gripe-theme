package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/lib/pq"

	"commerce-platform/internal/domain/notification"
	"commerce-platform/internal/domain/taxrate"
	"commerce-platform/internal/pkg/locales"
	"commerce-platform/internal/service"
)

func configurationModuleMissingEmailTemplateAndMemberLevelWritableTableNames() []string {
	return []string{"email_templates", "email_template_versions", "member_levels"}
}

// Validate original values through the provider-free renderer. Do not call Save,
// Rollback, notification delivery, account registration or point/reward methods.
func validateConfigurationModuleMissingEmailTemplateAndMemberLevelDomainRows(missing map[string][]json.RawMessage) error {
	for _, table := range configurationModuleMissingEmailTemplateAndMemberLevelWritableTableNames() {
		for _, row := range missing[table] {
			if err := validateConfigurationModuleRestoredEmailTemplateOrMemberLevelCompleteRow(table, row); err != nil {
				return fmt.Errorf("邮件模板或会员等级备份字段不符合恢复规则：%s", table)
			}
		}
	}
	return nil
}

func validateConfigurationModuleRestoredEmailTemplateOrMemberLevelCompleteRow(table string, row json.RawMessage) error {
	field := func(column string) string {
		value, _ := readBusinessModuleSnapshotRowColumnText(row, column)
		return value
	}
	if table == "member_levels" {
		minimum, minErr := strconv.ParseInt(field("min_points"), 10, 64)
		maximum, maxErr := strconv.ParseInt(field("max_points"), 10, 64)
		if strings.TrimSpace(field("name")) == "" || minErr != nil || maxErr != nil || minimum < 0 || maximum < minimum {
			return errors.New("invalid member-level name or point range")
		}
		// NUMERIC text is read without decoding it to a binary floating point.
		return taxrate.ValidateTaxRatePercentage(field("discount_rate_decimal"))
	}
	var template notification.EmailTemplate
	if err := json.Unmarshal(row, &template); err != nil {
		return err
	}
	definition, err := service.LookupTransactionalNotificationTemplate(template.Code)
	if err != nil || definition.Code != template.Code || !locales.IsSupportedCode(template.Locale) || locales.ResolveSupported(template.Locale) != template.Locale {
		return errors.New("unknown template identity or noncanonical locale")
	}
	if table == "email_template_versions" {
		// Version snapshots intentionally have no category column.
		template.Category = definition.Category
	}
	if template.Category != definition.Category {
		return errors.New("template category differs from the registered contract")
	}
	if err := template.Validate(); err != nil {
		return err
	}
	for _, variables := range []struct {
		actual   json.RawMessage
		expected []string
	}{{json.RawMessage(template.AllowedVariables), definition.AllowedVariables}, {json.RawMessage(template.RequiredVariables), definition.RequiredVariables}} {
		var values []string
		if err := json.Unmarshal(variables.actual, &values); err != nil || len(values) != len(variables.expected) {
			return errors.New("invalid template variable JSON or contract")
		}
		seen := map[string]bool{}
		for _, value := range values {
			if value == "" || value != strings.TrimSpace(value) || seen[value] {
				return errors.New("noncanonical or duplicate template variable")
			}
			seen[value] = true
		}
		for _, required := range variables.expected {
			if !seen[required] {
				return errors.New("template variable definition differs from the registered contract")
			}
		}
	}
	values := map[string]string{}
	for _, variable := range definition.AllowedVariables {
		values[variable] = "configuration-restore-render-validation"
	}
	_, err = service.RenderTransactionalNotificationTemplate(&template, values)
	return err
}

// Existing parents and target-only histories participate in the final graph.
// A current template must have its original complete immutable version chain;
// restoring history never advances or rolls back the current version pointer.
func validateConfigurationModuleRestoredEmailTemplateHistoriesAndMemberLevelRanges(ctx context.Context, tx *sql.Tx, missing map[string][]json.RawMessage) error {
	if err := validateConfigurationModuleMissingEmailTemplateAndMemberLevelDomainRows(missing); err != nil {
		return err
	}
	ids := []string{}
	for _, pair := range []struct{ table, column string }{{"email_templates", "id"}, {"email_template_versions", "template_id"}} {
		for _, row := range missing[pair.table] {
			id, _ := readBusinessModuleSnapshotRowColumnText(row, pair.column)
			ids = append(ids, id)
		}
	}
	if len(ids) > 0 {
		for _, pair := range []struct{ table, predicate string }{{"email_templates", `r.id::text=ANY($1)`}, {"email_template_versions", `r.template_id::text=ANY($1)`}} {
			rows, err := readProductCatalogRows(ctx, tx, pair.table, pair.predicate, pq.Array(ids))
			if err != nil {
				return err
			}
			for _, row := range rows {
				if err := validateConfigurationModuleRestoredEmailTemplateOrMemberLevelCompleteRow(pair.table, row); err != nil {
					return errors.New("恢复后的邮件模板父记录或完整历史版本无效")
				}
			}
		}
		var invalid bool
		err := tx.QueryRowContext(ctx, `SELECT EXISTS (
SELECT 1 FROM public.email_templates t
LEFT JOIN LATERAL (SELECT count(*) versions,min(version) first_version,max(version) last_version,
 bool_or(code<>t.code OR locale<>t.locale) wrong_identity
 FROM public.email_template_versions v WHERE v.template_id=t.id) h ON true
LEFT JOIN public.email_template_versions current_version ON current_version.template_id=t.id AND current_version.version=t.version
WHERE t.id::text=ANY($1) AND (h.versions<>t.version OR h.first_version<>1 OR h.last_version<>t.version OR h.wrong_identity
 OR current_version.id IS NULL
 OR ROW(current_version.code,current_version.locale,current_version.name,current_version.subject_template,current_version.body_html,current_version.body_text,current_version.allowed_variables,current_version.required_variables)
 IS DISTINCT FROM ROW(t.code,t.locale,t.name,t.subject_template,t.body_html,t.body_text,t.allowed_variables,t.required_variables)))`, pq.Array(ids)).Scan(&invalid)
		if err != nil {
			return err
		}
		if invalid {
			return errors.New("邮件模板身份、完整历史版本链或当前版本快照不一致")
		}
	}
	levelIDs := []string{}
	for _, row := range missing["member_levels"] {
		id, _ := readBusinessModuleSnapshotRowColumnText(row, "id")
		levelIDs = append(levelIDs, id)
	}
	if len(levelIDs) > 0 {
		var overlap bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM public.member_levels restored JOIN public.member_levels other ON other.id<>restored.id AND other.deleted_at IS NULL AND restored.min_points<=other.max_points AND restored.max_points>=other.min_points WHERE restored.id::text=ANY($1) AND restored.deleted_at IS NULL)`, pq.Array(levelIDs)).Scan(&overlap); err != nil {
			return err
		}
		if overlap {
			return errors.New("补回会员等级会与现有或同包等级的积分区间重叠，请先核对规则")
		}
	}
	return nil
}
