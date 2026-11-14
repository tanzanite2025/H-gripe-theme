package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/lib/pq"
)

type configurationCouponScopeReference struct {
	Table, Column, Value string
}

// Match checkout's product IDs and normalized product-category slugs without
// rewriting the original JSON. Category numbers must survive its float reader.
func readConfigurationCouponProductAndCategoryScopeReferences(row json.RawMessage) ([]configurationCouponScopeReference, error) {
	references := []configurationCouponScopeReference{}
	for _, column := range []string{"applicable_products", "excluded_products", "applicable_categories"} {
		text, err := readBusinessModuleSnapshotRowColumnText(row, column)
		if err != nil {
			return nil, err
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		var values []json.RawMessage
		if !strings.HasPrefix(text, "[") || json.Unmarshal([]byte(text), &values) != nil {
			return nil, errors.New("coupon scope must be a JSON array")
		}
		for _, value := range values {
			reference := configurationCouponScopeReference{Table: "products", Column: "id"}
			if column == "applicable_categories" {
				reference.Table = "product_categories"
				if len(value) > 0 && value[0] == '"' {
					var slug string
					if json.Unmarshal(value, &slug) != nil || strings.TrimSpace(slug) == "" {
						return nil, errors.New("coupon category slug is empty")
					}
					reference.Column, reference.Value = "slug", strings.ToLower(strings.TrimSpace(slug))
					references = append(references, reference)
					continue
				}
			}
			id, err := strconv.ParseInt(string(value), 10, 64)
			if err != nil || id <= 0 || (column == "applicable_categories" && id > 9007199254740991) {
				return nil, errors.New("coupon scope requires exact positive integer identities")
			}
			reference.Value = strconv.FormatInt(id, 10)
			references = append(references, reference)
		}
	}
	return references, nil
}

// Include the empty receipt set: used_count=0 alone cannot prove that an old
// package had no usage. One bounded scan covers all selected coupons.
func readConfigurationCouponRetainedUsageReceiptSets(ctx context.Context, tx *sql.Tx, coupons []json.RawMessage) (map[string][]json.RawMessage, error) {
	sets, ids := map[string][]json.RawMessage{}, []string{}
	for _, row := range coupons {
		id, err := readBusinessModuleSnapshotRowColumnText(row, "id")
		if err != nil || !canonicalModuleRecoveryDurableEventPositiveID(id) {
			return nil, errors.New("coupon original identity is invalid")
		}
		if _, exists := sets[id]; !exists {
			sets[id] = []json.RawMessage{}
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return sets, nil
	}
	rows, err := readProductCatalogRows(ctx, tx, "coupon_usage", "coupon_id::text=ANY($1)", pq.Array(ids))
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		id, _ := readBusinessModuleSnapshotRowColumnText(row, "coupon_id")
		sets[id] = append(sets[id], row)
	}
	return sets, nil
}

func captureConfigurationCouponOriginalReceiptSetsAndScopeDependencies(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, dependencies map[string]selectedBusinessModuleExternalReference) error {
	coupons := snapshot.Tables["coupons"].Rows
	if len(coupons) == 0 {
		return nil
	}
	sets, err := readConfigurationCouponRetainedUsageReceiptSets(ctx, tx, coupons)
	if err != nil {
		return err
	}
	selectedRows := indexOrderModuleHistoricalRowsByTableAndID(snapshot)
	captureDependency := func(reference configurationCouponScopeReference) error {
		values := map[string]string{reference.Column: reference.Value}
		identity := selectedBusinessModuleExternalReferenceIdentity(selectedBusinessModuleExternalReference{Table: reference.Table, Values: values})
		if _, captured := dependencies[identity]; captured {
			return nil
		}
		predicate := pq.QuoteIdentifier(reference.Column) + "::text=$1"
		if reference.Column == "slug" {
			predicate = "lower(btrim(slug))=$1"
		}
		rows, err := readProductCatalogRows(ctx, tx, reference.Table, predicate, reference.Value)
		if err != nil {
			return err
		}
		digest := "missing"
		if len(rows) == 1 {
			id, _ := readBusinessModuleSnapshotRowColumnText(rows[0], "id")
			if len(selectedRows[reference.Table][id]) > 0 {
				return nil
			}
			digest = businessModuleRecoveryBytesSHA256(rows[0])
		}
		dependencies[identity] = selectedBusinessModuleExternalReference{Table: reference.Table, Values: values, RowSHA256: digest}
		return nil
	}
	for id, receipts := range sets {
		if _, included := snapshot.Tables["coupon_usage"]; !included {
			digest, err := businessModuleRecoveryCanonicalJSONSHA256(receipts)
			if err != nil {
				return err
			}
			reference := selectedBusinessModuleExternalReference{Table: "coupon_usage", Values: map[string]string{"coupon_id": id}, RowSHA256: digest}
			dependencies[selectedBusinessModuleExternalReferenceIdentity(reference)] = reference
		}
		for _, receipt := range receipts {
			for _, target := range []struct{ column, table string }{{"user_id", "users"}, {"order_id", "orders"}} {
				value, _ := readBusinessModuleSnapshotRowColumnText(receipt, target.column)
				if canonicalModuleRecoveryDurableEventPositiveID(value) {
					if err := captureDependency(configurationCouponScopeReference{Table: target.table, Column: "id", Value: value}); err != nil {
						return err
					}
				}
			}
		}
	}
	for _, row := range coupons {
		references, err := readConfigurationCouponProductAndCategoryScopeReferences(row)
		if err != nil {
			continue // Export malformed originals; the restore domain check refuses them.
		}
		for _, reference := range references {
			if err := captureDependency(reference); err != nil {
				return err
			}
		}
	}
	return nil
}
