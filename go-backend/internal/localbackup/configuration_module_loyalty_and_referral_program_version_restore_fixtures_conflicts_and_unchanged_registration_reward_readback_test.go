package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func TestConfigurationLoyaltyAndReferralProgramVersionsPreserveOriginalPolicyFieldsWithoutCreditingPoints(t *testing.T) {
	loyalty := `{"id":1,"version":201,"status":"active","enabled":false,"currency":"USD","purchase_earn_points_per_currency_unit":2,"referral_referrer_points":17,"referral_referee_points":23,"checkin_base_points":5,"checkin_streak_interval_days":7,"checkin_streak_bonus_points":11,"checkin_max_points":29,"created_by":null}`
	referral := `{"id":2,"version":201,"status":"archived","enabled":false,"currency":"USD","min_order_amount_minor":9007199254740993,"referrer_reward_points":17,"referee_benefit_type":"points","referee_benefit_value":23,"vesting_period_days":30,"undelivered_fallback_days":45,"attribution_ttl_days":90,"monthly_cap_per_referrer":10,"anti_fraud_mode":"strict","source_loyalty_program_config_id":1,"created_by":null}`
	for _, example := range []struct {
		name, table, original, column, replacement string
		valid                                      bool
	}{
		{"disabled_loyalty_original_values", "loyalty_program_configs", loyalty, "", "", true},
		{"archived_referral_original_values", "referral_program_configs", referral, "", "", true},
		{"active_loyalty_version", "loyalty_program_configs", loyalty, "enabled", "true", true},
		{"invalid_version", "loyalty_program_configs", loyalty, "version", "0", false},
		{"unknown_status", "loyalty_program_configs", loyalty, "status", `"draft"`, false},
		{"currency_alias", "loyalty_program_configs", loyalty, "currency", `"usd"`, false},
		{"unknown_currency", "loyalty_program_configs", loyalty, "currency", `"ZZZ"`, false},
		{"negative_purchase_points", "loyalty_program_configs", loyalty, "purchase_earn_points_per_currency_unit", "-1", false},
		{"fractional_registration_points", "loyalty_program_configs", loyalty, "referral_referrer_points", "1.5", false},
		{"negative_referee_points", "loyalty_program_configs", loyalty, "referral_referee_points", "-1", false},
		{"zero_checkin_interval", "loyalty_program_configs", loyalty, "checkin_streak_interval_days", "0", false},
		{"checkin_maximum_below_base", "loyalty_program_configs", loyalty, "checkin_max_points", "4", false},
		{"invalid_creator", "loyalty_program_configs", loyalty, "created_by", "0", false},
		{"missing_enabled_state", "loyalty_program_configs", loyalty, "enabled", "null", false},
		{"negative_qualification_amount", "referral_program_configs", referral, "min_order_amount_minor", "-1", false},
		{"zero_registration_benefit", "referral_program_configs", referral, "referee_benefit_value", "0", false},
		{"retired_coupon_benefit", "referral_program_configs", referral, "referee_benefit_type", `"coupon"`, false},
		{"zero_vesting_window", "referral_program_configs", referral, "vesting_period_days", "0", false},
		{"fallback_before_vesting", "referral_program_configs", referral, "undelivered_fallback_days", "29", false},
		{"attribution_beyond_contract", "referral_program_configs", referral, "attribution_ttl_days", "91", false},
		{"zero_monthly_cap", "referral_program_configs", referral, "monthly_cap_per_referrer", "0", false},
		{"unknown_fraud_mode", "referral_program_configs", referral, "anti_fraud_mode", `"disabled"`, false},
		{"invalid_source_identity", "referral_program_configs", referral, "source_loyalty_program_config_id", "0", false},
	} {
		t.Run(example.name, func(t *testing.T) {
			row := json.RawMessage(example.original)
			if example.column != "" {
				var object map[string]json.RawMessage
				require.NoError(t, json.Unmarshal(row, &object))
				object[example.column] = json.RawMessage(example.replacement)
				var err error
				row, err = json.Marshal(object)
				require.NoError(t, err)
			}
			original := string(row)
			err := validateConfigurationModuleMissingLoyaltyAndReferralProgramVersionDomainRows(map[string][]json.RawMessage{example.table: {row}})
			require.Equal(t, example.valid, err == nil, "%v", err)
			require.Equal(t, original, string(row))
		})
	}
	for _, table := range strings.Fields("user_loyalty loyalty_transactions check_ins referral_records referral_reward_logs referral_transitions user_referral_identities referrals coupon_usage outbox_events") {
		require.False(t, configurationModuleHasMissingSiteContentWriteContract(selectedBusinessModuleCompleteSnapshot{Modules: []string{"configuration"}}, table), table)
	}
}

func configurationLoyaltyAndReferralProgramVersionRestoreIntegrationScenarios() []string {
	return []string{
		"program_partial_archived_loyalty_restore", "program_partial_archived_referral_restore", "program_partial_active_referral_restore", "program_partial_active_loyalty_and_referral_restore",
		"program_invalid_loyalty_currency", "program_invalid_loyalty_status", "program_invalid_loyalty_checkin_interval", "program_invalid_loyalty_checkin_maximum", "program_invalid_loyalty_version", "program_invalid_referral_currency",
		"program_loyalty_active_policy_conflict", "program_referral_active_policy_conflict", "program_loyalty_restored_active_not_latest", "program_referral_restored_active_not_latest", "program_referral_version_identity_conflict",
		"program_changed_existing_loyalty", "program_changed_original_creator", "program_missing_original_creator", "program_partial_referral_invalid_existing_loyalty",
		"program_partial_loyalty_invalid_current_active", "program_partial_referral_invalid_current_active", "legacy_v8_policy_requires_rehearsal",
	}
}

func createConfigurationLoyaltyAndReferralProgramVersionRestoreFixtureAndPreservedRegistrationRewards(t *testing.T, database *sql.DB, scenario string) {
	t.Helper()
	_, err := database.ExecContext(context.Background(), `
UPDATE loyalty_program_configs SET status='archived' WHERE status='active';
UPDATE referral_program_configs SET status='archived' WHERE status='active';
INSERT INTO users(id,email,username,password,role) VALUES(900170,'preserved-referee@example.invalid','preserved-referee','synthetic-password','user'),(900172,'original-policy-creator@example.invalid','original-policy-creator','synthetic-password','admin');
INSERT INTO loyalty_program_configs(id,version,status,enabled,currency,exchange_rate_points,purchase_earn_points_per_currency_unit,referral_referrer_points,referral_referee_points,checkin_base_points,checkin_streak_interval_days,checkin_streak_bonus_points,checkin_max_points,created_by,created_at,updated_at)
VALUES(900170,200,'archived',false,'USD',100,2,17,23,5,7,11,29,900172,'2025-01-02 03:04:05','2025-01-02 03:04:05'),(900171,201,'active',true,'USD',100,3,31,37,7,9,13,41,900172,'2025-01-03 03:04:05','2025-01-03 03:04:05');
INSERT INTO referral_program_configs(id,version,status,enabled,currency,min_order_amount_minor,referrer_reward_points,referee_benefit_type,referee_benefit_value,vesting_period_days,undelivered_fallback_days,attribution_ttl_days,monthly_cap_per_referrer,anti_fraud_mode,source_loyalty_program_config_id,created_by,created_at)
VALUES(900170,200,'archived',true,'USD',9007199254740993,17,'points',23,30,45,90,10,'strict',(SELECT min(id) FROM loyalty_program_configs),900001,'2025-01-02T03:04:05Z'),
(900171,201,'active',false,'USD',9007199254740993,31,'points',37,31,46,89,11,'monitor',900171,900001,'2025-01-03T03:04:05Z');
INSERT INTO user_referral_identities(id,user_id,referral_code) VALUES(900170,900001,'BACKUPKEEP');
INSERT INTO referral_records(id,referral_identity_id,program_config_id,referrer_id,referee_id,referral_code_snapshot,attribution_source,status,expires_at,settled_at)
VALUES(900170,900170,(SELECT min(id) FROM referral_program_configs),900001,900170,'BACKUPKEEP','manual_input','settled','2025-02-01T00:00:00Z','2025-01-02T00:00:00Z');
INSERT INTO referral_reward_logs(id,referral_record_id,program_config_id,recipient_user_id,recipient_role,reward_type,points_amount,loyalty_transaction_id,idempotency_key,status,released_at)
VALUES(900170,900170,(SELECT min(id) FROM referral_program_configs),900001,'referrer','points',17,900110,'original-registration-reward-must-not-replay','released','2025-01-02T00:00:00Z');
`)
	require.NoError(t, err)
	mutation := map[string]string{
		"program_invalid_loyalty_currency":                  `UPDATE loyalty_program_configs SET currency='ZZZ' WHERE id=900170`,
		"program_invalid_loyalty_status":                    `UPDATE loyalty_program_configs SET status='pending' WHERE id=900170`,
		"program_invalid_loyalty_checkin_interval":          `UPDATE loyalty_program_configs SET checkin_streak_interval_days=0 WHERE id=900170`,
		"program_invalid_loyalty_checkin_maximum":           `UPDATE loyalty_program_configs SET checkin_max_points=4 WHERE id=900170`,
		"program_invalid_loyalty_version":                   `UPDATE loyalty_program_configs SET version=0 WHERE id=900170`,
		"program_invalid_referral_currency":                 `UPDATE referral_program_configs SET currency='ZZZ' WHERE id=900170`,
		"program_partial_referral_invalid_existing_loyalty": `UPDATE loyalty_program_configs SET checkin_streak_interval_days=0 WHERE id=(SELECT min(id) FROM loyalty_program_configs)`,
		"program_partial_loyalty_invalid_current_active":    `UPDATE loyalty_program_configs SET checkin_streak_interval_days=0 WHERE id=900171`,
		"program_partial_referral_invalid_current_active":   `UPDATE referral_program_configs SET currency='ZZZ' WHERE id=900171`,
	}[scenario]
	if mutation != "" {
		_, err = database.ExecContext(context.Background(), mutation)
		require.NoError(t, err)
	}
}

func deleteConfigurationLoyaltyAndReferralProgramVersionFixtureRowsAndPreserveTargetOnlyHistory(t *testing.T, database *sql.DB, scenario string) int {
	t.Helper()
	statement, count := `DELETE FROM referral_program_configs WHERE id IN(900170,900171);DELETE FROM loyalty_program_configs WHERE id IN(900170,900171);`, 4
	switch scenario {
	case "program_partial_archived_loyalty_restore", "program_partial_loyalty_invalid_current_active":
		statement, count = `DELETE FROM loyalty_program_configs WHERE id=900170;`, 1
	case "program_partial_archived_referral_restore", "program_partial_referral_invalid_existing_loyalty", "program_changed_existing_loyalty", "program_partial_referral_invalid_current_active":
		statement, count = `DELETE FROM referral_program_configs WHERE id=900170;`, 1
	case "program_partial_active_referral_restore":
		statement, count = `DELETE FROM referral_program_configs WHERE id=900171;`, 1
	case "program_partial_active_loyalty_and_referral_restore":
		statement, count = `DELETE FROM referral_program_configs WHERE id=900171;DELETE FROM loyalty_program_configs WHERE id=900171;`, 2
	}
	_, err := database.ExecContext(context.Background(), statement+`
INSERT INTO loyalty_program_configs(id,version,status,enabled,currency,exchange_rate_points,purchase_earn_points_per_currency_unit,referral_referrer_points,referral_referee_points,checkin_base_points,checkin_streak_interval_days,checkin_streak_bonus_points,checkin_max_points)
VALUES(900179,199,'archived',false,'USD',100,0,0,0,1,7,0,1);
INSERT INTO referral_program_configs(id,version,status,enabled,currency,referee_benefit_value) VALUES(900179,199,'archived',false,'USD',5);
`)
	require.NoError(t, err)
	mutation := map[string]string{
		"program_loyalty_active_policy_conflict":      `UPDATE loyalty_program_configs SET status='active',version=202 WHERE id=900179`,
		"program_referral_active_policy_conflict":     `UPDATE referral_program_configs SET status='active',version=202 WHERE id=900179`,
		"program_loyalty_restored_active_not_latest":  `UPDATE loyalty_program_configs SET version=202 WHERE id=900179`,
		"program_referral_restored_active_not_latest": `UPDATE referral_program_configs SET version=202 WHERE id=900179`,
		"program_referral_version_identity_conflict":  `UPDATE referral_program_configs SET version=200 WHERE id=900179`,
		"program_changed_existing_loyalty":            `UPDATE loyalty_program_configs SET enabled=NOT enabled WHERE id=900170`,
		"program_changed_original_creator":            `UPDATE users SET username='changed-policy-creator' WHERE id=900172`,
		"program_missing_original_creator":            `DELETE FROM users WHERE id=900172`,
	}[scenario]
	if mutation != "" {
		_, err = database.ExecContext(context.Background(), mutation)
		require.NoError(t, err)
	}
	return count
}

func configurationLoyaltyAndReferralProgramVersionIntegrationScenarioRequiresRejection(scenario string) bool {
	return strings.HasPrefix(scenario, "program_") && !strings.HasSuffix(scenario, "_restore")
}

func verifyRestoredLoyaltyAndReferralProgramVersionsThroughReadOnlyBusinessServicesAndUnchangedRegistrationRewardFacts(t *testing.T, database *sql.DB) {
	t.Helper()
	tx, err := database.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})
	require.NoError(t, err)
	defer tx.Rollback()
	gormDatabase, err := gorm.Open(postgres.New(postgres.Config{Conn: tx}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)
	loyaltyProgram := service.NewLoyaltyProgramService(repository.NewLoyaltyProgramRepository(gormDatabase))
	active, err := loyaltyProgram.GetPublicConfig()
	require.NoError(t, err)
	require.EqualValues(t, 900171, active.ID)
	require.Equal(t, 201, active.Version)
	require.True(t, active.Enabled)
	require.Equal(t, 3, active.PurchaseEarnPointsPerUnit)
	require.Equal(t, 31, active.ReferralReferrerPoints)
	require.Equal(t, 37, active.ReferralRefereePoints)
	require.Equal(t, 9, active.CheckInStreakIntervalDays)
	referralProgram := service.NewReferralService(nil, nil, repository.NewReferralProgramRepository(gormDatabase), nil, "https://synthetic.invalid", "synthetic-program-readback-only-secret")
	referral, err := referralProgram.GetAdminProgramConfig()
	require.NoError(t, err)
	require.EqualValues(t, 900171, referral.ID)
	require.Equal(t, 201, referral.Version)
	require.False(t, referral.Enabled)
	require.Equal(t, "points", referral.RefereeBenefitType)
	require.EqualValues(t, 37, referral.RefereeBenefitValue)
	require.EqualValues(t, 9007199254740993, referral.MinOrderAmountMinor)
	require.NotNil(t, referral.SourceLoyaltyProgramConfigID)
	require.EqualValues(t, 900171, *referral.SourceLoyaltyProgramConfigID)
	var historicalStatus, historicalCreatedAt string
	var historicalEnabled bool
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT status,enabled,to_char(created_at,'YYYY-MM-DD HH24:MI:SS') FROM loyalty_program_configs WHERE id=900170`).Scan(&historicalStatus, &historicalEnabled, &historicalCreatedAt))
	require.Equal(t, "archived", historicalStatus)
	require.False(t, historicalEnabled)
	require.Equal(t, "2025-01-02 03:04:05", historicalCreatedAt)
	var balance, rewardCount, pointsTransactions, extraHistory int
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT available_points FROM user_loyalty WHERE user_id=900001`).Scan(&balance))
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT count(*) FROM referral_reward_logs WHERE id=900170 AND status='released' AND points_amount=17 AND loyalty_transaction_id=900110`).Scan(&rewardCount))
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT count(*) FROM loyalty_transactions WHERE user_id=900001`).Scan(&pointsTransactions))
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT (SELECT count(*) FROM loyalty_program_configs WHERE id=900179 AND version=199 AND status='archived')+(SELECT count(*) FROM referral_program_configs WHERE id=900179 AND version=199 AND status='archived')`).Scan(&extraHistory))
	require.Equal(t, 17, balance)
	require.Equal(t, 1, rewardCount)
	require.Equal(t, 1, pointsTransactions)
	require.Equal(t, 2, extraHistory)
}
