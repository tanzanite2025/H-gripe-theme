package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/lib/pq"
)

const customerProfileModuleFormalRestorePolicy = "restore_missing_customer_profiles_and_closed_new_account_facts_preserve_current_identities_v1"

type CustomerProfileModuleRestoreEligibility struct {
	Policy   string   `json:"policy"`
	Eligible bool     `json:"eligible"`
	Blockers []string `json:"blockers"`
}

// This dependency order governs bounded customer inserts, not arbitrary overwrite.
// Export selectors have their own row filters; cross-scope checks compare the sets.
func customerProfileModuleOwnedTableNames() []string {
	return strings.Fields("users visitor_profiles subscriptions wishlist_items user_loyalty loyalty_transactions check_ins user_referral_identities referral_records referral_reward_logs referral_transitions referrals")
}

func customerProfileModuleHasDedicatedMissingRowContract(snapshot selectedBusinessModuleCompleteSnapshot, table string) bool {
	if !reflect.DeepEqual(snapshot.Modules, []string{"customers"}) {
		return false
	}
	for _, name := range customerProfileModuleOwnedTableNames() {
		if name == table {
			return true
		}
	}
	return false
}

func validateCustomerProfileModuleInsertOnlyWriteContract(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, table string, observation *moduleRecoveryDatabaseObservation) error {
	if !customerProfileModuleHasDedicatedMissingRowContract(snapshot, table) {
		return validateSelectedBusinessModuleIsolatedInsertWriteContract(ctx, tx, table, observation)
	}
	return validateBusinessModuleInsertOnlyTableContractWithUpdateOnlyGuardsPreserved(ctx, tx, table, observation)
}

func customerProfileModuleFinancialOwnerColumns(table string) []string {
	switch table {
	case "user_loyalty", "loyalty_transactions", "check_ins", "user_referral_identities":
		return []string{"user_id"}
	case "referral_records":
		return []string{"referrer_id", "referee_id"}
	case "referral_reward_logs":
		return []string{"recipient_user_id"}
	case "referrals":
		return []string{"referrer_id", "referred_id"}
	}
	return nil
}

// Existing credentials, restrictions, balances and consent remain current.
// A reused primary key or a changed owner is never treated as a newer profile.
func canPreserveCurrentCustomerProfileModuleRow(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, table string, row json.RawMessage) (bool, error) {
	if !customerProfileModuleHasDedicatedMissingRowContract(snapshot, table) {
		return false, nil
	}
	columns := customerProfileModuleFinancialOwnerColumns(table)
	switch table {
	case "users":
		columns = strings.Fields("email username role created_at")
	case "visitor_profiles":
		columns = strings.Fields("user_id customer_service_visitor_hash cart_session_id created_at")
	case "subscriptions":
		columns = strings.Fields("email created_at")
	case "wishlist_items":
		columns = strings.Fields("user_id product_id created_at")
	case "loyalty_transactions":
		columns = strings.Fields("user_id type points balance source source_id created_at")
	case "check_ins":
		columns = strings.Fields("user_id check_in_date points_earned created_at")
	case "user_referral_identities":
		columns = strings.Fields("user_id referral_code created_at")
	case "referral_records":
		columns = strings.Fields("referral_identity_id program_config_id referrer_id referee_id created_at")
	case "referral_reward_logs":
		columns = strings.Fields("referral_record_id program_config_id recipient_user_id recipient_role reward_type points_amount idempotency_key loyalty_transaction_id created_at")
	case "referral_transitions":
		return false, nil
	case "referrals":
		columns = strings.Fields("referrer_id referred_id referral_code created_at")
	}
	if len(columns) == 0 {
		return false, nil
	}
	var predicates []string
	for _, column := range columns {
		predicates = append(predicates, "to_jsonb(r)->"+pq.QuoteLiteral(column)+" IS NOT DISTINCT FROM $1::jsonb->"+pq.QuoteLiteral(column))
	}
	var sameIdentity bool
	err := tx.QueryRowContext(ctx, `SELECT `+strings.Join(predicates, " AND ")+` FROM public.`+pq.QuoteIdentifier(table)+` r WHERE r.id=(($1::jsonb)->>'id')::bigint`, string(row)).Scan(&sameIdentity)
	return sameIdentity, err
}

func inspectCustomerProfileModuleRestoreEligibility(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, missing map[string][]json.RawMessage) (*CustomerProfileModuleRestoreEligibility, error) {
	result := &CustomerProfileModuleRestoreEligibility{Policy: customerProfileModuleFormalRestorePolicy, Eligible: true, Blockers: []string{}}
	record := func(reason string) {
		for _, existing := range result.Blockers {
			if existing == reason {
				return
			}
		}
		result.Eligible = false
		result.Blockers = append(result.Blockers, reason)
	}
	if !reflect.DeepEqual(snapshot.Modules, []string{"customers"}) {
		record("standalone_customer_package_required")
		return result, nil
	}
	missingAccounts := map[string]bool{}
	for _, row := range missing["users"] {
		id, _ := readBusinessModuleSnapshotRowColumnText(row, "id")
		role, _ := readBusinessModuleSnapshotRowColumnText(row, "role")
		if id == "" || (role != "user" && role != "test_user") {
			record("only_storefront_customer_accounts_can_be_restored")
			continue
		}
		missingAccounts[id] = true
		var identityTaken bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM public.users WHERE lower(email)=lower($1::jsonb->>'email') OR lower(username)=lower($1::jsonb->>'username'))`, string(row)).Scan(&identityTaken); err != nil {
			return nil, err
		}
		if identityTaken {
			record("customer_email_or_username_already_has_another_identity")
		}
	}
	index := indexOrderModuleHistoricalRowsByTableAndID(snapshot)
	for _, table := range []string{"subscriptions", "visitor_profiles"} {
		for _, row := range missing[table] {
			predicate := `lower(email)=lower($1::jsonb->>'email')`
			if table == "visitor_profiles" {
				predicate = `(coalesce(customer_service_visitor_hash,'')<>'' AND customer_service_visitor_hash=$1::jsonb->>'customer_service_visitor_hash') OR (coalesce(cart_session_id,'')<>'' AND cart_session_id=$1::jsonb->>'cart_session_id')`
			}
			var identityTaken bool
			if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM public.`+pq.QuoteIdentifier(table)+` WHERE `+predicate+`)`, string(row)).Scan(&identityTaken); err != nil {
				return nil, err
			}
			if identityTaken {
				record("customer_contact_or_visitor_identity_already_exists")
			}
		}
	}
	financialAccounts := map[string]bool{}
	for table, rows := range missing {
		if len(rows) == 0 {
			continue
		}
		if !customerProfileModuleHasDedicatedMissingRowContract(snapshot, table) {
			record("shared_media_or_execution_facts_must_already_exist:" + table)
			continue
		}
		for _, row := range rows {
			owners := customerProfileModuleFinancialOwnerColumns(table)
			if table == "referral_transitions" {
				recordID, _ := readBusinessModuleSnapshotRowColumnText(row, "referral_record_id")
				row = index["referral_records"][recordID]
				owners = customerProfileModuleFinancialOwnerColumns("referral_records")
				if len(row) == 0 {
					record("referral_transition_requires_original_record")
					continue
				}
			}
			for _, column := range owners {
				owner, _ := readBusinessModuleSnapshotRowColumnText(row, column)
				if owner == "" || owner == "0" || owner == "null" {
					continue
				}
				financialAccounts[owner] = true
				if !missingAccounts[owner] {
					record("missing_points_or_referral_facts_require_entirely_missing_account_cohort")
				}
			}
			if table == "referral_records" {
				status, _ := readBusinessModuleSnapshotRowColumnText(row, "status")
				if status != "settled" && status != "expired" && status != "revoked" && status != "reversed" {
					record("active_referral_execution_cannot_be_restored")
				}
			}
			if table == "referral_reward_logs" {
				status, _ := readBusinessModuleSnapshotRowColumnText(row, "status")
				if status != "released" && status != "forfeited" && status != "reversed" {
					record("unfinished_referral_reward_cannot_be_restored")
				}
			}
			if table == "referrals" {
				status, _ := readBusinessModuleSnapshotRowColumnText(row, "status")
				if status != "completed" && status != "expired" {
					record("active_legacy_referral_cannot_be_restored")
				}
			}
		}
	}
	for owner := range financialAccounts {
		originalBalances := 0
		for _, row := range snapshot.Tables["user_loyalty"].Rows {
			userID, _ := readBusinessModuleSnapshotRowColumnText(row, "user_id")
			if userID == owner {
				originalBalances++
			}
		}
		if originalBalances != 1 {
			record("missing_account_cohort_requires_original_balance_summary")
		}
		// An old balance is not evidence of today's balance. Only an absent
		// account cohort with no current financial footprint can be copied whole.
		for _, table := range customerProfileModuleOwnedTableNames() {
			columns := customerProfileModuleFinancialOwnerColumns(table)
			if len(columns) == 0 {
				continue
			}
			var predicates []string
			for _, column := range columns {
				predicates = append(predicates, pq.QuoteIdentifier(column)+"::text=$1")
			}
			var exists bool
			if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM public.`+pq.QuoteIdentifier(table)+` WHERE `+strings.Join(predicates, " OR ")+`)`, owner).Scan(&exists); err != nil {
				return nil, err
			}
			if exists {
				record("missing_account_has_current_points_or_referral_facts")
			}
		}
		var externalFinancialFacts bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM public.orders WHERE user_id::text=$1)`, owner).Scan(&externalFinancialFacts); err != nil {
			return nil, err
		}
		if externalFinancialFacts {
			record("missing_account_balance_requires_current_order_reconciliation")
		}
	}
	// These tables have business user references without database foreign keys.
	for _, table := range []string{"visitor_profiles", "user_loyalty", "loyalty_transactions", "check_ins", "wishlist_items", "referrals"} {
		for _, row := range missing[table] {
			columns := []string{"user_id"}
			if table == "referrals" {
				columns = []string{"referrer_id", "referred_id"}
			}
			for _, column := range columns {
				owner, _ := readBusinessModuleSnapshotRowColumnText(row, column)
				if owner == "" || owner == "0" || owner == "null" || missingAccounts[owner] {
					continue
				}
				var exists bool
				if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM public.users WHERE id::text=$1 AND role IN ('user','test_user'))`, owner).Scan(&exists); err != nil {
					return nil, err
				}
				if !exists {
					record("customer_business_reference_requires_original_storefront_account")
				}
			}
		}
	}
	sort.Strings(result.Blockers)
	return result, nil
}

// INSERT only, in dependency order, with ordinary FK/check/unique constraints
// and triggers enabled. No registration, reward or notification service runs.
func restoreCustomerProfileModuleMissingRowsWithNormalConstraintsAndVerifiedPreservation(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, observation *moduleRecoveryDatabaseObservation, missing map[string][]json.RawMessage, transactionalSequences bool) (int, error) {
	sequences, err := readBusinessModuleRestoreSequenceStates(ctx, tx)
	if err != nil {
		return 0, err
	}
	before := map[string]BusinessModuleRestoreTablePreservation{}
	for table := range observation.Tables {
		preserved, err := fingerprintBusinessModuleRetainedTargetTable(ctx, tx, table, observation.Shapes[table], missing[table])
		if err != nil {
			return 0, err
		}
		before[table] = preserved
	}
	restored := 0
	for _, table := range customerProfileModuleOwnedTableNames() {
		if len(missing[table]) == 0 {
			continue
		}
		if err := validateCustomerProfileModuleInsertOnlyWriteContract(ctx, tx, snapshot, table, observation); err != nil {
			return 0, err
		}
		var columns []string
		for column := range observation.Columns[table] {
			columns = append(columns, pq.QuoteIdentifier(column))
		}
		sort.Strings(columns)
		list := strings.Join(columns, ",")
		for _, row := range missing[table] {
			if _, err := tx.ExecContext(ctx, `INSERT INTO public.`+pq.QuoteIdentifier(table)+` (`+list+`) SELECT `+list+` FROM jsonb_populate_record(NULL::public.`+pq.QuoteIdentifier(table)+`,$1::jsonb)`, string(row)); err != nil { // #nosec G202 -- Table and column identifiers use pq.QuoteIdentifier from the validated schema; complete row values use bound JSON parameters.
				return 0, fmt.Errorf("客户完整行恢复违反身份或业务约束：%s", table)
			}
			var same bool
			if err := tx.QueryRowContext(ctx, `SELECT to_jsonb(record)=$1::jsonb FROM public.`+pq.QuoteIdentifier(table)+` record WHERE `+businessModuleCompleteRowTypedPrimaryKeyPredicate(table, observation.PrimaryKeys[table]), string(row)).Scan(&same); err != nil || !same {
				return 0, fmt.Errorf("客户完整行写入后无法与原包核对一致：%s", table)
			}
			restored++
		}
	}
	if err := validateProductCatalogAffectedForeignKeys(ctx, tx, productCatalogModuleSnapshot{Tables: snapshot.Tables}); err != nil {
		return 0, err
	}
	for table, expected := range before {
		actual, err := fingerprintBusinessModuleRetainedTargetTable(ctx, tx, table, observation.Shapes[table], missing[table])
		if err != nil {
			return 0, err
		}
		if actual.ContentFingerprint != expected.ContentFingerprint || actual.RetainedRecords != expected.RetainedRecords {
			return 0, fmt.Errorf("客户恢复改变了当前保留记录：%s", table)
		}
	}
	if transactionalSequences {
		err = advanceBusinessModuleRestoreOwnedSequencesTransactionallyAndVerifyPreservation(ctx, tx, sequences, missing)
	} else {
		_, err = advanceCustomerServiceCandidateOwnedSequencesWithoutRewinding(ctx, tx, sequences, missing)
	}
	return restored, err
}
