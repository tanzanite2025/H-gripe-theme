package localbackup

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/lib/pq"
)

// This policy applies to product-only recovery. Whole-storefront recovery retains
// its complete database snapshot, including the inventory belonging to that point.
const productCatalogCurrentInventoryPreservationPolicy = "preserve_existing_target_inventory_and_restore_new_variant_snapshot_stock"

type ProductCatalogInventoryPreservationResult struct {
	Policy                                    string `json:"policy"`
	ExistingVariantStocksPreserved            int    `json:"existing_variant_stocks_preserved"`
	DifferentSnapshotStockValuesNotApplied    int    `json:"different_snapshot_stock_values_not_applied"`
	NewVariantStocksRestoredFromSnapshot      int    `json:"new_variant_stocks_restored_from_snapshot"`
	ExistingProductViewCountsPreserved        int    `json:"existing_product_view_counts_preserved"`
	ExistingInventoryAndIdentityFactsVerified bool   `json:"existing_inventory_and_identity_facts_verified"`
}

type productCatalogExistingInventoryAndIdentityFacts struct {
	ProductIDs           []string
	VariantIDs           []string
	PreservedFingerprint string
}

// Prepare all changes before the first INSERT. Identifiers and inventory owners
// cannot be reassigned to another SKU while orders/refunds still reference them.
// Only the in-memory merge changes: the authenticated package remains intact.
func prepareProductCatalogSnapshotWithCurrentInventoryAndStableBusinessIdentities(ctx context.Context, tx *sql.Tx, snapshot productCatalogModuleSnapshot) (productCatalogModuleSnapshot, ProductCatalogInventoryPreservationResult, productCatalogExistingInventoryAndIdentityFacts, error) {
	prepared := snapshot
	prepared.Tables = make(map[string]productCatalogTableSnapshot, len(snapshot.Tables))
	for name, table := range snapshot.Tables {
		prepared.Tables[name] = table
	}
	result := ProductCatalogInventoryPreservationResult{Policy: productCatalogCurrentInventoryPreservationPolicy}
	facts := productCatalogExistingInventoryAndIdentityFacts{}
	for _, tableName := range []string{"products", "product_variants"} {
		table := snapshot.Tables[tableName]
		preparedRows := make([]json.RawMessage, 0, len(table.Rows))
		for _, sourceRow := range table.Rows {
			var sourceFields map[string]json.RawMessage
			if err := json.Unmarshal(sourceRow, &sourceFields); err != nil {
				return snapshot, result, facts, err
			}
			identityColumns := []string{"id", "slug", "locale", "parent_id", "created_at"}
			preservedColumn := "view_count"
			if tableName == "product_variants" {
				identityColumns = []string{"id", "product_id", "sku", "master_variant_id", "option_values", "created_at"}
				preservedColumn = "stock"
			}
			comparisons := make([]string, 0, len(identityColumns))
			for _, column := range append(append([]string{}, identityColumns...), preservedColumn) {
				if _, exists := sourceFields[column]; !exists {
					return snapshot, result, facts, errors.New("product recovery row lacks a required inventory or identity field")
				}
			}
			for _, column := range identityColumns {
				quoted := pq.QuoteIdentifier(column)
				comparisons = append(comparisons, "current."+quoted+" IS NOT DISTINCT FROM source."+quoted)
			}
			if tableName == "product_variants" {
				stock, err := productCatalogRowString(sourceRow, "stock")
				value, parseErr := strconv.ParseInt(stock, 10, 32)
				if err != nil || parseErr != nil || value < 0 {
					return snapshot, result, facts, errors.New("product snapshot contains an invalid inventory quantity")
				}
			}
			var currentRow string
			var sameIdentity, samePreservedValue bool
			query := `SELECT to_jsonb(current)::text, (` + strings.Join(comparisons, " AND ") + `), current.` + pq.QuoteIdentifier(preservedColumn) + ` IS NOT DISTINCT FROM source.` + pq.QuoteIdentifier(preservedColumn) +
				` FROM jsonb_populate_record(NULL::public.` + pq.QuoteIdentifier(tableName) + `,$1::jsonb) source JOIN public.` + pq.QuoteIdentifier(tableName) + ` current ON current.id=source.id`
			err := tx.QueryRowContext(ctx, query, string(sourceRow)).Scan(&currentRow, &sameIdentity, &samePreservedValue)
			if errors.Is(err, sql.ErrNoRows) {
				preparedRows = append(preparedRows, sourceRow)
				if tableName == "product_variants" {
					result.NewVariantStocksRestoredFromSnapshot++
				}
				continue
			}
			if err != nil {
				return snapshot, result, facts, fmt.Errorf("could not compare persisted product inventory and identity fields in %s", tableName)
			}
			if !sameIdentity {
				return snapshot, result, facts, fmt.Errorf("existing %s identity or inventory owner differs; product recovery requires explicit identity reconciliation", tableName)
			}
			var currentFields map[string]json.RawMessage
			if err := json.Unmarshal([]byte(currentRow), &currentFields); err != nil {
				return snapshot, result, facts, err
			}
			sourceFields[preservedColumn] = currentFields[preservedColumn]
			row, err := json.Marshal(sourceFields)
			if err != nil {
				return snapshot, result, facts, err
			}
			preparedRows = append(preparedRows, row)
			id, err := productCatalogRowString(json.RawMessage(currentRow), "id")
			if err != nil {
				return snapshot, result, facts, err
			}
			if tableName == "product_variants" {
				facts.VariantIDs = append(facts.VariantIDs, id)
				result.ExistingVariantStocksPreserved++
				if !samePreservedValue {
					result.DifferentSnapshotStockValuesNotApplied++
				}
			} else {
				facts.ProductIDs = append(facts.ProductIDs, id)
				result.ExistingProductViewCountsPreserved++
			}
		}
		table.Rows = preparedRows
		prepared.Tables[tableName] = table
	}
	var err error
	facts.PreservedFingerprint, err = captureProductCatalogExistingInventoryAndIdentityFingerprint(ctx, tx, facts)
	return prepared, result, facts, err
}

func captureProductCatalogExistingInventoryAndIdentityFingerprint(ctx context.Context, tx *sql.Tx, facts productCatalogExistingInventoryAndIdentityFacts) (string, error) {
	digest := sha256.New()
	for _, selected := range []struct {
		table, columns string
		ids            []string
	}{
		{"products", "id,slug,locale,parent_id,created_at,view_count", facts.ProductIDs},
		{"product_variants", "id,product_id,sku,master_variant_id,option_values,created_at,stock", facts.VariantIDs},
	} {
		rows, err := tx.QueryContext(ctx, `SELECT to_jsonb(facts)::text FROM (SELECT `+selected.columns+` FROM public.`+pq.QuoteIdentifier(selected.table)+` WHERE id::text=ANY($1) ORDER BY id) facts`, pq.Array(selected.ids)) // #nosec G202 -- Column lists are fixed literals; table identifiers are quoted and selected IDs are bound parameters.
		if err != nil {
			return "", err
		}
		_, _ = fmt.Fprintf(digest, "%q:%d\n", selected.table, len(selected.ids))
		count := 0
		for rows.Next() {
			var raw string
			if err := rows.Scan(&raw); err != nil {
				_ = rows.Close()
				return "", err
			}
			_, _ = fmt.Fprintf(digest, "%d:%s\n", len(raw), raw)
			count++
		}
		err = rows.Err()
		_ = rows.Close()
		if err != nil {
			return "", err
		}
		if count != len(selected.ids) {
			return "", errors.New("a preserved product inventory or identity row is missing")
		}
	}
	return fmt.Sprintf("%x", digest.Sum(nil)), nil
}

func validateProductCatalogSelectedAndRetainedSharedInventoryOwners(ctx context.Context, tx *sql.Tx, snapshot productCatalogModuleSnapshot) error {
	ids := make([]string, 0, len(snapshot.Tables["product_variants"].Rows))
	for _, row := range snapshot.Tables["product_variants"].Rows {
		id, err := productCatalogRowString(row, "id")
		if err != nil {
			return err
		}
		ids = append(ids, id)
	}
	var invalid bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(
SELECT 1 FROM public.product_variants child
LEFT JOIN public.product_variants owner ON owner.id=child.master_variant_id
LEFT JOIN public.products child_product ON child_product.id=child.product_id
WHERE (child.id::text=ANY($1) OR child.master_variant_id::text=ANY($1))
AND child.master_variant_id IS NOT NULL
AND (owner.id IS NULL OR owner.id=child.id OR owner.master_variant_id IS NOT NULL
 OR child.stock<>0 OR child_product.parent_id IS DISTINCT FROM owner.product_id
 OR child.option_values IS DISTINCT FROM owner.option_values))`, pq.Array(ids)).Scan(&invalid)
	if err != nil {
		return err
	}
	if invalid {
		return errors.New("selected or retained translated variants have an invalid shared inventory owner")
	}
	return nil
}
