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

func TestConfigurationCouponRecoveryPreservesExactMoneyOriginalDatesLimitsAndProductCategoryScope(t *testing.T) {
	original := `{"id":1,"code":"ORIGINAL","type":"fixed","value_rate_decimal":7.123456789012345,"value_minor":9007199254740993,"currency":"USD","min_amount_minor":0,"max_discount_minor":0,"usage_limit":3,"usage_limit_per_user":1,"used_count":1,"start_date":"2025-01-02T03:04:05Z","end_date":"2025-02-02T03:04:05Z","enabled":false,"referral_recipient_user_id":null,"applicable_products":"[9007199254740993]","excluded_products":"[]","applicable_categories":"[7,\" Original-Slug \"]"}`
	for _, sample := range []struct {
		name, column, replacement string
		valid                     bool
	}{
		{"expired_disabled_fixed_coupon_exact_money", "", "", true},
		{"percentage_exact_rate", "type", `"percentage"`, true},
		{"zero_fixed_amount", "value_minor", "0", false},
		{"negative_minimum", "min_amount_minor", "-1", false},
		{"fractional_money", "value_minor", "1.5", false},
		{"overflow_money", "value_minor", "9223372036854775808", false},
		{"unknown_type", "type", `"other"`, false},
		{"empty_code", "code", `" "`, false},
		{"currency_alias", "currency", `"usd"`, false},
		{"unknown_currency", "currency", `"ZZZ"`, false},
		{"invalid_rate", "value_rate_decimal", `"NaN"`, false},
		{"negative_limit", "usage_limit", "-1", false},
		{"negative_per_user_limit", "usage_limit_per_user", "-1", false},
		{"count_exceeds_limit", "used_count", "4", false},
		{"missing_original_enabled", "enabled", "null", false},
		{"inverted_dates", "end_date", `"2024-01-01T00:00:00Z"`, false},
		{"invalid_recipient", "referral_recipient_user_id", "0", false},
		{"product_json_object", "applicable_products", `"{}"`, false},
		{"product_json_null", "applicable_products", `"null"`, false},
		{"product_string_identity", "applicable_products", `"[\"1\"]"`, false},
		{"product_non_positive", "excluded_products", `"[0]"`, false},
		{"empty_category_slug", "applicable_categories", `"[\" \"]"`, false},
		{"category_float_identity", "applicable_categories", `"[1.5]"`, false},
		{"category_float_reader_precision_loss", "applicable_categories", `"[9007199254740993]"`, false},
	} {
		t.Run(sample.name, func(t *testing.T) {
			var object map[string]json.RawMessage
			require.NoError(t, json.Unmarshal([]byte(original), &object))
			if sample.column != "" {
				object[sample.column] = json.RawMessage(sample.replacement)
			}
			row, err := json.Marshal(object)
			require.NoError(t, err)
			before := string(row)
			err = validateConfigurationModuleMissingCouponDomainRows(map[string][]json.RawMessage{"coupons": {row}})
			require.Equal(t, sample.valid, err == nil, "%v", err)
			require.Equal(t, before, string(row))
		})
	}
	references, err := readConfigurationCouponProductAndCategoryScopeReferences(json.RawMessage(original))
	require.NoError(t, err)
	require.Contains(t, references, configurationCouponScopeReference{Table: "products", Column: "id", Value: "9007199254740993"})
	require.Contains(t, references, configurationCouponScopeReference{Table: "product_categories", Column: "slug", Value: "original-slug"})
	for _, rate := range []string{"0", "100.000000000000001", `"12/5"`} {
		var object map[string]json.RawMessage
		require.NoError(t, json.Unmarshal([]byte(original), &object))
		object["type"], object["value_rate_decimal"] = json.RawMessage(`"percentage"`), json.RawMessage(rate)
		row, err := json.Marshal(object)
		require.NoError(t, err)
		require.Error(t, validateConfigurationModuleMissingCouponDomainRows(map[string][]json.RawMessage{"coupons": {row}}))
	}
	for _, table := range strings.Fields("coupon_usage orders order_items user_loyalty loyalty_transactions referral_reward_logs outbox_events") {
		require.False(t, configurationModuleHasMissingSiteContentWriteContract(selectedBusinessModuleCompleteSnapshot{Modules: []string{"configuration"}}, table))
	}
}

func verifyConfigurationCouponOlderPackagesWithoutOriginalReceiptOrProductScopeFingerprintsAreRejected(t *testing.T, database *sql.DB, cfg Configuration) {
	t.Helper()
	tx, err := database.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	require.NoError(t, err)
	defer tx.Rollback()
	snapshot, err := cfg.captureSelectedBusinessModuleCompleteSnapshot(context.Background(), tx, []string{"configuration"})
	require.NoError(t, err)
	observation, err := inspectModuleRecoveryDatabaseSchemaWithoutRowFingerprints(context.Background(), tx)
	require.NoError(t, err)
	actual, err := captureSelectedBusinessModuleExternalReferenceFingerprintsWithObservedSchema(context.Background(), tx, snapshot, observation)
	require.NoError(t, err)
	for _, missing := range []struct{ name, table, column, value string }{
		{"old_package_missing_original_empty_usage_set", "coupon_usage", "coupon_id", "900181"},
		{"old_package_missing_original_nonempty_usage_set", "coupon_usage", "coupon_id", "900180"},
		{"old_package_missing_original_excluded_product", "products", "id", "900180"},
	} {
		t.Run(missing.name, func(t *testing.T) {
			older := snapshot
			older.ExternalDependencies = []selectedBusinessModuleExternalReference{}
			for _, reference := range snapshot.ExternalDependencies {
				if reference.Table != missing.table || reference.Values[missing.column] != missing.value {
					older.ExternalDependencies = append(older.ExternalDependencies, reference)
				}
			}
			require.Len(t, older.ExternalDependencies, len(snapshot.ExternalDependencies)-1)
			report := &SelectedBusinessModuleIsolatedRestorePreview{}
			classifySelectedBusinessModuleExternalDependencyIssues(older, observation, actual, report)
			require.Equal(t, 1, report.DependencyConflicts)
			require.Equal(t, "unexpected_target_reference", report.DependencyIssues[0].Reason)
		})
	}
}

func TestConfigurationCouponOriginalEmptyUsageSetIsRequiredAndChangedOrMissingReceiptDigestsBlockRecovery(t *testing.T) {
	empty, err := businessModuleRecoveryCanonicalJSONSHA256([]json.RawMessage{})
	require.NoError(t, err)
	original := selectedBusinessModuleExternalReference{Table: "coupon_usage", Values: map[string]string{"coupon_id": "1"}, RowSHA256: empty}
	for _, scenario := range []string{"same_empty_set", "missing_original_set", "changed_receipts", "missing_target_set"} {
		t.Run(scenario, func(t *testing.T) {
			snapshot := selectedBusinessModuleCompleteSnapshot{Modules: []string{"configuration"}, ExternalDependencies: []selectedBusinessModuleExternalReference{original}}
			actual := []selectedBusinessModuleExternalReference{original}
			if scenario == "missing_original_set" {
				snapshot.ExternalDependencies = nil
			}
			if scenario == "changed_receipts" {
				actual[0].RowSHA256 = strings.Repeat("a", 64)
			}
			if scenario == "missing_target_set" {
				actual = nil
			}
			report := &SelectedBusinessModuleIsolatedRestorePreview{}
			classifySelectedBusinessModuleExternalDependencyIssues(snapshot, &moduleRecoveryDatabaseObservation{}, actual, report)
			require.Equal(t, scenario == "same_empty_set", len(report.DependencyIssues) == 0)
		})
	}
}

func configurationCouponRestoreIntegrationScenarios() []string {
	return []string{
		"coupon_partial_percentage_restore", "coupon_partial_expired_restore", "coupon_partial_deleted_restore", "coupon_partial_guest_restore",
		"coupon_invalid_type", "coupon_invalid_currency", "coupon_invalid_dates", "coupon_invalid_scope_json", "coupon_invalid_category_precision",
		"coupon_count_receipt_mismatch", "coupon_user_limit_exceeded", "coupon_guest_email_limit_exceeded", "coupon_guest_identity_missing", "coupon_private_recipient_mismatch",
		"coupon_changed_retained_usage", "coupon_missing_retained_usage", "coupon_changed_read_only_product", "coupon_missing_scope_product", "coupon_changed_product_category", "coupon_disabled_product_category",
		"coupon_changed_usage_order", "coupon_code_identity_conflict", "coupon_changed_existing_row", "legacy_v9_policy_requires_rehearsal",
	}
}

func createConfigurationCouponRestoreFixtureAndPreservedUsageRecords(t *testing.T, database *sql.DB, scenario string) {
	t.Helper()
	_, err := database.ExecContext(context.Background(), `
INSERT INTO product_categories(id,name,slug) VALUES(900180,'Coupon scope category','restore-coupon-scope');
UPDATE products SET product_category_id=900180 WHERE id=900001;
INSERT INTO products(id,name,slug,price_minor,currency) VALUES(900180,'Coupon excluded product','restore-coupon-excluded',123,'USD');
INSERT INTO orders(id,order_number,status,currency,payment_currency,payment_amount_minor,subtotal_amount_minor,total_amount_minor,shipping_fee_minor,tax_amount_minor,discount_amount_minor,points_value_minor)
VALUES(900180,'coupon-reversed-order','cancelled','USD','USD',1000,1000,1000,0,0,0,0),(900181,'coupon-guest-fx-order','completed','EUR','EUR',1000,1000,1000,0,0,0,0);
INSERT INTO coupons(id,code,type,value_minor,value_rate_decimal,currency,min_amount_minor,max_discount_minor,usage_limit,usage_limit_per_user,used_count,start_date,end_date,enabled,applicable_products,excluded_products,applicable_categories,referral_recipient_user_id,deleted_at)
VALUES
(900180,'RESTORE-FIXED','fixed',123,7,'USD',100,200,3,1,1,'2025-01-01','2099-01-01',true,'[900001]','[900180]','[900180," RESTORE-COUPON-SCOPE "]',null,null),
(900181,'RESTORE-PERCENT','percentage',0,12.123456789012345,'USD',0,9007199254740993,0,0,0,'2025-01-01','2099-01-01',true,'','','',null,null),
(900182,'RESTORE-EXPIRED','fixed',9007199254740993,0,'USD',0,0,0,0,0,'2024-01-01','2025-01-01',true,'','','',null,null),
(900183,'RESTORE-DISABLED','fixed',123,0,'USD',0,0,0,0,0,'2025-01-01','2099-01-01',false,'','','',null,null),
(900184,'RESTORE-DELETED','fixed',123,0,'USD',0,0,0,0,0,'2025-01-01','2099-01-01',true,'','','',null,'2025-06-01'),
(900185,'RESTORE-PRIVATE','fixed',123,0,'USD',0,0,0,0,0,'2025-01-01','2099-01-01',true,'','','',900170,null),
(900186,'RESTORE-GUEST','fixed',123,0,'USD',0,0,3,1,1,'2025-01-01','2099-01-01',true,'','','',null,null);
INSERT INTO coupon_usage(id,coupon_id,user_id,email,order_id,discount_minor,currency,status,reversed_at,reversal_reason)
VALUES(900180,900180,900001,'original-user@example.invalid',900001,123,'USD','applied',null,null),
(900181,900180,900001,'original-user@example.invalid',900180,123,'USD','reversed','2025-06-01','original reversal'),
(900182,900186,0,'guest@example.invalid',900181,111,'EUR','applied',null,null);
`)
	require.NoError(t, err)
	mutation := map[string]string{
		"coupon_invalid_type":               `UPDATE coupons SET type='other' WHERE id=900181`,
		"coupon_invalid_currency":           `UPDATE coupons SET currency='ZZZ' WHERE id=900181`,
		"coupon_invalid_dates":              `UPDATE coupons SET end_date=start_date WHERE id=900181`,
		"coupon_invalid_scope_json":         `UPDATE coupons SET applicable_products='{' WHERE id=900181`,
		"coupon_invalid_category_precision": `UPDATE coupons SET applicable_categories='[9007199254740993]' WHERE id=900181`,
		"coupon_count_receipt_mismatch":     `UPDATE coupons SET used_count=2 WHERE id=900180`,
		"coupon_user_limit_exceeded":        `UPDATE coupons SET usage_limit_per_user=0 WHERE id=900180;INSERT INTO coupon_usage(coupon_id,user_id,email,order_id,discount_minor,currency) VALUES(900180,900001,'original-user@example.invalid',900180,123,'USD');UPDATE coupons SET usage_limit_per_user=1,used_count=2 WHERE id=900180`,
		"coupon_guest_email_limit_exceeded": `UPDATE coupons SET usage_limit_per_user=0 WHERE id=900186;INSERT INTO coupon_usage(coupon_id,user_id,email,order_id,discount_minor,currency) VALUES(900186,900001,' GUEST@example.invalid ',900180,123,'USD');UPDATE coupons SET usage_limit_per_user=1,used_count=2 WHERE id=900186`,
		"coupon_guest_identity_missing":     `UPDATE coupons SET usage_limit_per_user=0 WHERE id=900186;UPDATE coupon_usage SET email='' WHERE id=900182;UPDATE coupons SET usage_limit_per_user=1 WHERE id=900186`,
		"coupon_private_recipient_mismatch": `UPDATE coupons SET referral_recipient_user_id=900170 WHERE id=900180`,
		"coupon_disabled_product_category":  `UPDATE product_categories SET is_enabled=false WHERE id=900180`,
	}[scenario]
	if mutation != "" {
		_, err = database.ExecContext(context.Background(), mutation)
		require.NoError(t, err)
	}
}

func deleteConfigurationCouponFixtureRowsAndPreserveCurrentUsageAndTargetOnlyCoupons(t *testing.T, database *sql.DB, scenario string) int {
	t.Helper()
	statement, count := `DELETE FROM coupons WHERE id BETWEEN 900180 AND 900186;`, 7
	partialID := map[string]string{"coupon_partial_percentage_restore": "900181", "coupon_partial_expired_restore": "900182", "coupon_partial_deleted_restore": "900184", "coupon_partial_guest_restore": "900186", "coupon_changed_existing_row": "900181"}[scenario]
	if partialID != "" {
		statement, count = `DELETE FROM coupons WHERE id=`+partialID+`;`, 1
	}
	_, err := database.ExecContext(context.Background(), statement+`
INSERT INTO coupons(id,code,type,value_minor,currency,start_date,end_date,enabled) VALUES(900189,'TARGET-ONLY-COUPON','fixed',321,'USD','2025-01-01','2099-01-01',false);
`)
	require.NoError(t, err)
	mutation := map[string]string{
		"coupon_changed_retained_usage":    `UPDATE coupon_usage SET discount_minor=122 WHERE id=900180`,
		"coupon_missing_retained_usage":    `DELETE FROM coupon_usage WHERE id=900181`,
		"coupon_changed_read_only_product": `UPDATE products SET name='Changed coupon product' WHERE id=900001`,
		"coupon_missing_scope_product":     `DELETE FROM products WHERE id=900180`,
		"coupon_changed_product_category":  `UPDATE product_categories SET slug='changed-coupon-scope' WHERE id=900180`,
		"coupon_changed_usage_order":       `UPDATE orders SET status='refunded' WHERE id=900180`,
		"coupon_code_identity_conflict":    `UPDATE coupons SET code='RESTORE-FIXED' WHERE id=900189`,
		"coupon_changed_existing_row":      `UPDATE coupons SET used_count=0 WHERE id=900180`,
	}[scenario]
	if mutation != "" {
		_, err = database.ExecContext(context.Background(), mutation)
		require.NoError(t, err)
	}
	return count
}

func configurationCouponIntegrationScenarioRequiresRejection(scenario string) bool {
	return strings.HasPrefix(scenario, "coupon_") && !strings.HasSuffix(scenario, "_restore")
}

func verifyRestoredCouponsThroughReadOnlyBusinessValidationWithoutConsumingUsageOrReplayingRewards(t *testing.T, database *sql.DB) {
	t.Helper()
	tx, err := database.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})
	require.NoError(t, err)
	defer tx.Rollback()
	gormDatabase, err := gorm.Open(postgres.New(postgres.Config{Conn: tx}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)
	couponRepository := repository.NewCouponRepository(gormDatabase)
	marketing := service.NewMarketingService(nil, couponRepository, nil)
	fixed, discount, err := marketing.ValidateCoupon("RESTORE-FIXED", 900170, 1000)
	require.NoError(t, err)
	require.EqualValues(t, 123, discount)
	require.Equal(t, 1, fixed.UsedCount)
	require.Equal(t, `[900001]`, fixed.ApplicableProducts)
	require.Equal(t, `[900180]`, fixed.ExcludedProducts)
	require.Equal(t, `[900180," RESTORE-COUPON-SCOPE "]`, fixed.ApplicableCategories)
	_, _, err = marketing.ValidateCoupon("RESTORE-FIXED", 900001, 1000)
	require.ErrorContains(t, err, "limit")
	percentage, discount, err := marketing.ValidateCoupon("RESTORE-PERCENT", 900170, 10000)
	require.NoError(t, err)
	require.Equal(t, "12.123456789012345", percentage.ValueRateDecimal)
	require.EqualValues(t, 1212, discount)
	_, _, err = marketing.ValidateCoupon("RESTORE-GUEST", 0, 1000, " GUEST@example.invalid ")
	require.ErrorContains(t, err, "limit")
	_, _, err = marketing.ValidateCoupon("RESTORE-GUEST", 0, 1000, "other@example.invalid")
	require.NoError(t, err)
	_, _, err = marketing.ValidateCoupon("RESTORE-PRIVATE", 900001, 1000)
	require.Error(t, err)
	_, _, err = marketing.ValidateCoupon("RESTORE-PRIVATE", 900170, 1000)
	require.NoError(t, err)
	for _, code := range []string{"RESTORE-EXPIRED", "RESTORE-DISABLED", "RESTORE-DELETED"} {
		_, _, err = marketing.ValidateCoupon(code, 900170, 1000)
		require.Error(t, err, code)
	}
	var usageCount, reversedCount, targetOnly int
	var expiredAmount int64
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT count(*),count(*) FILTER(WHERE status='reversed') FROM coupon_usage WHERE coupon_id BETWEEN 900180 AND 900186`).Scan(&usageCount, &reversedCount))
	require.Equal(t, 3, usageCount)
	require.Equal(t, 1, reversedCount)
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT value_minor FROM coupons WHERE id=900182`).Scan(&expiredAmount))
	require.EqualValues(t, 9007199254740993, expiredAmount)
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT count(*) FROM coupons WHERE id=900189 AND value_minor=321 AND NOT enabled`).Scan(&targetOnly))
	require.Equal(t, 1, targetOnly)
}
