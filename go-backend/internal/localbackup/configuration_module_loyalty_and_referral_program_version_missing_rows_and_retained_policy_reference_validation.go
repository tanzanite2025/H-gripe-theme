package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/lib/pq"

	"commerce-platform/internal/domain/currency"
)

// Restore immutable policy versions in dependency order. Account balances,
// registration rewards, check-ins, order earnings and coupons have other owners.
func configurationModuleMissingLoyaltyAndReferralProgramVersionWritableTableNames() []string {
	return []string{"loyalty_program_configs", "referral_program_configs"}
}

// The loyalty creator column has no FK. Preserve its original identity/content
// as a read-only dependency instead of accepting a reused numeric account ID.
func configurationModuleLoyaltyProgramCreatorReadOnlyDependencyReferences() []ModuleRecoveryDependencyReference {
	return []ModuleRecoveryDependencyReference{{SourceTable: "loyalty_program_configs", SourceColumns: []string{"created_by"}, TargetTable: "users", TargetColumns: []string{"id"}}}
}

func validateConfigurationModuleMissingLoyaltyAndReferralProgramVersionDomainRows(missing map[string][]json.RawMessage) error {
	for _, table := range configurationModuleMissingLoyaltyAndReferralProgramVersionWritableTableNames() {
		for _, row := range missing[table] {
			if err := validateConfigurationModuleRestoredLoyaltyOrReferralProgramVersionCompleteRow(table, row); err != nil {
				return fmt.Errorf("积分或推荐计划原版本字段无效（%s）：%w", table, err)
			}
		}
	}
	return nil
}

func validateConfigurationModuleRestoredLoyaltyOrReferralProgramVersionCompleteRow(table string, row json.RawMessage) error {
	var object map[string]json.RawMessage
	if json.Unmarshal(row, &object) != nil || object == nil {
		return errors.New("expected a complete policy version object")
	}
	readPolicyField := func(column string) string {
		value, _ := readBusinessModuleSnapshotRowColumnText(row, column)
		return value
	}
	readPolicyInteger := func(column string, minimum int64) (int64, error) {
		value, err := strconv.ParseInt(readPolicyField(column), 10, 64)
		if err != nil || value < minimum {
			return 0, fmt.Errorf("invalid exact policy integer: %s", column)
		}
		return value, nil
	}
	for _, column := range []string{"id", "version"} {
		if _, err := readPolicyInteger(column, 1); err != nil {
			return err
		}
	}
	if readPolicyField("status") != "active" && readPolicyField("status") != "archived" {
		return errors.New("policy status must be active or archived")
	}
	if !currency.IsCatalogCode(readPolicyField("currency")) || readPolicyField("currency") != currency.NormalizeCode(readPolicyField("currency")) {
		return errors.New("original policy currency must already be canonical")
	}
	if string(object["enabled"]) != "true" && string(object["enabled"]) != "false" {
		return errors.New("original policy enabled state is required")
	}
	if creator := object["created_by"]; len(creator) > 0 && string(creator) != "null" {
		if _, err := readPolicyInteger("created_by", 1); err != nil {
			return err
		}
	}
	if table == "loyalty_program_configs" {
		for _, column := range []string{"purchase_earn_points_per_currency_unit", "referral_referrer_points", "referral_referee_points", "checkin_base_points", "checkin_streak_bonus_points", "checkin_max_points"} {
			if _, err := readPolicyInteger(column, 0); err != nil {
				return err
			}
		}
		if _, err := readPolicyInteger("checkin_streak_interval_days", 1); err != nil {
			return err
		}
		base, _ := readPolicyInteger("checkin_base_points", 0)
		maximum, _ := readPolicyInteger("checkin_max_points", 0)
		if maximum < base {
			return errors.New("check-in maximum cannot be below the original base points")
		}
		return nil
	}
	if table != "referral_program_configs" {
		return errors.New("unregistered policy version table")
	}
	for _, column := range []string{"min_order_amount_minor", "referrer_reward_points"} {
		if _, err := readPolicyInteger(column, 0); err != nil {
			return err
		}
	}
	for _, column := range []string{"referee_benefit_value", "vesting_period_days", "undelivered_fallback_days", "attribution_ttl_days", "monthly_cap_per_referrer"} {
		if _, err := readPolicyInteger(column, 1); err != nil {
			return err
		}
	}
	vesting, _ := readPolicyInteger("vesting_period_days", 1)
	fallback, _ := readPolicyInteger("undelivered_fallback_days", 1)
	attribution, _ := readPolicyInteger("attribution_ttl_days", 1)
	if fallback < vesting || attribution > 90 || readPolicyField("referee_benefit_type") != "points" || (readPolicyField("anti_fraud_mode") != "monitor" && readPolicyField("anti_fraud_mode") != "strict") {
		return errors.New("invalid original referral benefit, policy window or anti-fraud mode")
	}
	if source := object["source_loyalty_program_config_id"]; len(source) > 0 && string(source) != "null" {
		if _, err := readPolicyInteger("source_loyalty_program_config_id", 1); err != nil {
			return err
		}
	}
	return nil
}

// Validate the final retained graph without publishing a new version, changing
// any status, invoking reward services or replaying registration/order events.
func validateConfigurationModuleRestoredLoyaltyAndReferralProgramVersionRelationships(ctx context.Context, tx *sql.Tx, missing map[string][]json.RawMessage) error {
	if err := validateConfigurationModuleMissingLoyaltyAndReferralProgramVersionDomainRows(missing); err != nil {
		return err
	}
	loyaltyIDs, referralIDs := []string{}, []string{}
	for _, row := range missing["loyalty_program_configs"] {
		id, _ := readBusinessModuleSnapshotRowColumnText(row, "id")
		loyaltyIDs = append(loyaltyIDs, id)
	}
	for _, row := range missing["referral_program_configs"] {
		id, _ := readBusinessModuleSnapshotRowColumnText(row, "id")
		referralIDs = append(referralIDs, id)
	}
	if len(loyaltyIDs)+len(referralIDs) == 0 {
		return nil
	}
	referrals, err := readProductCatalogRows(ctx, tx, "referral_program_configs", `r.id::text=ANY($1) OR r.source_loyalty_program_config_id::text=ANY($2)`, pq.Array(referralIDs), pq.Array(loyaltyIDs))
	if err != nil {
		return err
	}
	for _, row := range referrals {
		if err := validateConfigurationModuleRestoredLoyaltyOrReferralProgramVersionCompleteRow("referral_program_configs", row); err != nil {
			return fmt.Errorf("受影响的原推荐版本无效：%w", err)
		}
		source, _ := readBusinessModuleSnapshotRowColumnText(row, "source_loyalty_program_config_id")
		if source != "" {
			loyaltyIDs = append(loyaltyIDs, source)
		}
	}
	loyaltyRows, err := readProductCatalogRows(ctx, tx, "loyalty_program_configs", `r.id::text=ANY($1)`, pq.Array(loyaltyIDs))
	if err != nil {
		return err
	}
	for _, row := range loyaltyRows {
		if err := validateConfigurationModuleRestoredLoyaltyOrReferralProgramVersionCompleteRow("loyalty_program_configs", row); err != nil {
			return fmt.Errorf("受影响的原积分版本无效：%w", err)
		}
	}
	for _, table := range configurationModuleMissingLoyaltyAndReferralProgramVersionWritableTableNames() {
		if len(missing[table]) == 0 {
			continue
		}
		activeRows, err := readProductCatalogRows(ctx, tx, table, `r.status='active'`)
		if err != nil {
			return err
		}
		for _, row := range activeRows {
			if err := validateConfigurationModuleRestoredLoyaltyOrReferralProgramVersionCompleteRow(table, row); err != nil {
				return fmt.Errorf("恢复历史后当前 active 计划字段无效：%w", err)
			}
		}
		var invalid bool
		if err := tx.QueryRowContext(ctx, `SELECT count(*) FILTER (WHERE status='active')<>1 OR COALESCE(max(version) FILTER (WHERE status='active')<>max(version),true) FROM public.`+pq.QuoteIdentifier(table)).Scan(&invalid); err != nil {
			return err
		}
		if invalid {
			return errors.New("积分或推荐计划必须保留唯一且最新的 active 原版本；恢复不会归档当前策略或回退版本")
		}
	}
	var missingCreator bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM public.loyalty_program_configs p LEFT JOIN public.users u ON u.id=p.created_by WHERE p.id::text=ANY($1) AND p.created_by IS NOT NULL AND u.id IS NULL)`, pq.Array(loyaltyIDs)).Scan(&missingCreator); err != nil {
		return err
	}
	if missingCreator {
		return errors.New("积分计划缺少原创建者账号，不能按相同数字重建或改绑")
	}
	return nil
}
