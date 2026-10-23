package localbackup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/lib/pq"
)

const productCatalogStandardWriteGuardVersion = 1

// These exact bodies come from migration 432. Only surrounding whitespace and
// CRLF line endings are normalized; names alone never approve a function body.
const productCatalogSystemCategoryIdentityAndTreeProtectionFunctionBody = `BEGIN
    IF TG_OP <> 'INSERT' AND OLD.slug IN ('wheel-components', 'carbon-rim') THEN
        IF TG_OP = 'DELETE' THEN
            RAISE EXCEPTION 'System category % cannot be deleted', OLD.slug USING ERRCODE = '23514';
        END IF;
        IF NEW.id IS DISTINCT FROM OLD.id OR NEW.slug IS DISTINCT FROM OLD.slug THEN
            RAISE EXCEPTION 'System category % identity cannot be changed', OLD.slug USING ERRCODE = '23514';
        END IF;
    END IF;
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;
    IF NEW.slug = 'wheel-components' AND (NEW.parent_id IS NOT NULL OR NEW.depth <> 1 OR NOT NEW.is_enabled) THEN
        RAISE EXCEPTION 'wheel-components must remain an enabled root category' USING ERRCODE = '23514';
    END IF;
    IF NEW.slug = 'carbon-rim' AND (
        NEW.depth <> 2 OR NOT NEW.is_enabled OR NOT EXISTS (
            SELECT 1 FROM product_categories parent WHERE parent.id = NEW.parent_id
            AND parent.slug = 'wheel-components' AND parent.parent_id IS NULL AND parent.depth = 1 AND parent.is_enabled
        )
    ) THEN
        RAISE EXCEPTION 'carbon-rim must remain enabled directly under wheel-components' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END`

const productCatalogCarbonRimProductTypeAndCategoryBindingFunctionBody = `DECLARE
    category_slug TEXT;
BEGIN
    IF TG_OP = 'UPDATE' AND NEW.product_type IS DISTINCT FROM OLD.product_type
       AND (NEW.product_type = 'carbon_rim' OR OLD.product_type = 'carbon_rim') THEN
        RAISE EXCEPTION 'Dedicated carbon_rim product type is immutable' USING ERRCODE = '23514';
    END IF;
    SELECT slug INTO category_slug FROM product_categories WHERE id = NEW.product_category_id;
    IF NEW.product_type = 'carbon_rim' AND category_slug IS DISTINCT FROM 'carbon-rim' THEN
        RAISE EXCEPTION 'carbon_rim products must belong to the carbon-rim system category' USING ERRCODE = '23514';
    END IF;
    IF NEW.product_type <> 'carbon_rim' AND category_slug = 'carbon-rim' THEN
        RAISE EXCEPTION 'carbon-rim is reserved for dedicated carbon_rim products' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END`

type productCatalogApprovedStandardTriggerContract struct {
	Table, Name, UpdatedColumns, FunctionBody string
	TriggerType                               int
}

func productCatalogApprovedStandardTriggerContracts() []productCatalogApprovedStandardTriggerContract {
	return []productCatalogApprovedStandardTriggerContract{
		{"product_categories", "protect_carbon_rim_system_category_identity_and_tree_position", "", productCatalogSystemCategoryIdentityAndTreeProtectionFunctionBody, 31},
		{"products", "require_carbon_rim_product_type_and_system_category_binding", "product_category_id,product_type", productCatalogCarbonRimProductTypeAndCategoryBindingFunctionBody, 23},
	}
}

func validateProductCatalogStandardTriggersAndRejectUnregisteredRewriteRules(ctx context.Context, tx *sql.Tx) error {
	allowedTables := productCatalogAllowedTableNames()
	definitions, err := inspectModuleRecoveryDatabaseWriteSideEffectDefinitions(ctx, tx)
	if err != nil {
		return err
	}
	approved := map[string]productCatalogApprovedStandardTriggerContract{}
	for _, contract := range productCatalogApprovedStandardTriggerContracts() {
		approved[contract.Table+"/"+contract.Name] = contract
	}
	seen := map[string]bool{}
	for table, effects := range definitions {
		if !allowedTables[table] {
			continue
		}
		for _, effect := range effects {
			identity := table + "/" + effect.Name
			contract, exists := approved[identity]
			if !exists || effect.Kind != "trigger" || effect.Enabled != "O" || seen[identity] {
				return fmt.Errorf("product restore has an unregistered or altered write trigger/rewrite rule on %s", table)
			}
			var triggerType int
			var functionSchema, functionName, language, resultType, updatedColumns, body string
			var metadataAllowed bool
			err := tx.QueryRowContext(ctx, `SELECT t.tgtype::int, namespace.nspname, p.proname, language.lanname, p.prorettype::regtype::text,
COALESCE((SELECT string_agg(a.attname,',' ORDER BY a.attname) FROM unnest(t.tgattr::smallint[]) column_number JOIN pg_attribute a ON a.attrelid=t.tgrelid AND a.attnum=column_number),''), p.prosrc,
t.tgnargs=0 AND octet_length(t.tgargs)=0 AND t.tgqual IS NULL AND t.tgconstraint=0 AND NOT t.tgdeferrable AND NOT t.tginitdeferred AND t.tgoldtable IS NULL AND t.tgnewtable IS NULL
AND p.pronargs=0 AND p.prokind='f' AND NOT p.proretset AND NOT p.prosecdef AND NOT p.proleakproof AND p.provolatile='v' AND p.proparallel='u' AND p.proconfig IS NULL AND p.prosupport=0
FROM pg_trigger t JOIN pg_proc p ON p.oid=t.tgfoid JOIN pg_namespace namespace ON namespace.oid=p.pronamespace JOIN pg_language language ON language.oid=p.prolang
WHERE t.tgrelid=to_regclass($1) AND t.tgname=$2 AND NOT t.tgisinternal`, "public."+table, effect.Name).Scan(&triggerType, &functionSchema, &functionName, &language, &resultType, &updatedColumns, &body, &metadataAllowed)
			if err != nil {
				return err
			}
			body = strings.TrimSpace(strings.ReplaceAll(body, "\r\n", "\n"))
			if !metadataAllowed || triggerType != contract.TriggerType || functionSchema != "public" || functionName != contract.Name || language != "plpgsql" || resultType != "trigger" || updatedColumns != contract.UpdatedColumns || body != contract.FunctionBody {
				return fmt.Errorf("product standard trigger function or execution contract differs on %s", table)
			}
			seen[identity] = true
		}
	}
	if len(seen) != len(approved) {
		return errors.New("product standard trigger contract is missing")
	}
	return nil
}

type productCatalogProtectedSystemCategoryIdentities struct {
	WheelComponentsID int64
	CarbonRimID       int64
}

func inspectProductCatalogRequiredSystemCategoriesAndCarbonRimBindings(ctx context.Context, tx *sql.Tx) (productCatalogProtectedSystemCategoryIdentities, error) {
	var identities productCatalogProtectedSystemCategoryIdentities
	if err := tx.QueryRowContext(ctx, `SELECT root.id, carbon.id FROM public.product_categories root JOIN public.product_categories carbon ON carbon.parent_id=root.id WHERE root.slug='wheel-components' AND root.parent_id IS NULL AND root.depth=1 AND root.is_enabled AND carbon.slug='carbon-rim' AND carbon.depth=2 AND carbon.is_enabled`).Scan(&identities.WheelComponentsID, &identities.CarbonRimID); err != nil {
		if err == sql.ErrNoRows {
			return identities, errors.New("product restore requires the enabled wheel-components root and its carbon-rim system category")
		}
		return identities, err
	}
	var invalid bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM public.products WHERE (product_type='carbon_rim' AND product_category_id IS DISTINCT FROM $1) OR (product_type IS DISTINCT FROM 'carbon_rim' AND product_category_id=$1))`, identities.CarbonRimID).Scan(&invalid); err != nil {
		return identities, err
	}
	if invalid {
		return identities, errors.New("product restore would break carbon_rim product type and system category bindings")
	}
	return identities, nil
}

func lockProductCatalogWriteTablesAndValidateStandardBusinessContractsBeforeMerge(ctx context.Context, tx *sql.Tx, snapshot productCatalogModuleSnapshot) (productCatalogProtectedSystemCategoryIdentities, error) {
	var identities productCatalogProtectedSystemCategoryIdentities
	names := []string{}
	for table := range productCatalogAllowedTableNames() {
		names = append(names, "public."+pq.QuoteIdentifier(table))
	}
	sort.Strings(names)
	// Prevent an isolated-stage ALTER from installing an ALWAYS/REPLICA trigger
	// between validation and writes. This never locks the live target for merging.
	if _, err := tx.ExecContext(ctx, `LOCK TABLE `+strings.Join(names, ",")+` IN SHARE ROW EXCLUSIVE MODE`); err != nil {
		return identities, err
	}
	if err := validateProductCatalogStandardTriggersAndRejectUnregisteredRewriteRules(ctx, tx); err != nil {
		return identities, err
	}
	identities, err := inspectProductCatalogRequiredSystemCategoriesAndCarbonRimBindings(ctx, tx)
	if err != nil {
		return identities, err
	}
	for _, row := range snapshot.Tables["products"].Rows {
		id, err := productCatalogRowString(row, "id")
		if err != nil {
			return identities, err
		}
		incomingType, err := productCatalogRowString(row, "product_type")
		if err != nil {
			return identities, err
		}
		var currentType string
		err = tx.QueryRowContext(ctx, `SELECT product_type FROM public.products WHERE id::text=$1`, id).Scan(&currentType)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return identities, err
		}
		if currentType != incomingType && (currentType == "carbon_rim" || incomingType == "carbon_rim") {
			return identities, errors.New("product restore cannot change an existing dedicated carbon_rim product type")
		}
	}
	return identities, nil
}
