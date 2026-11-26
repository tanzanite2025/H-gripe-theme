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

// Version 5 replaces the temporary spokes-category rejection with the dedicated spoke product binding.
const productCatalogStandardWriteGuardVersion = 5

// These exact bodies come from migrations 432, 439, 440, 442, 444 and 445. Only surrounding
// whitespace and CRLF line endings are normalized; names alone never approve a function body.
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

const productCatalogBicycleTireSystemCategoryIdentityAndTreeProtectionFunctionBody = `BEGIN
    IF TG_OP <> 'INSERT' AND OLD.slug = 'tire' THEN
        IF TG_OP = 'DELETE' THEN
            RAISE EXCEPTION 'The tire system category cannot be deleted' USING ERRCODE = '23514';
        END IF;
        IF NEW.id IS DISTINCT FROM OLD.id OR NEW.slug IS DISTINCT FROM OLD.slug THEN
            RAISE EXCEPTION 'The tire system category identity cannot be changed' USING ERRCODE = '23514';
        END IF;
    END IF;

    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;

    IF NEW.slug = 'tire' AND (
        NEW.depth <> 2 OR NOT NEW.is_enabled OR NOT EXISTS (
            SELECT 1 FROM product_categories parent
            WHERE parent.id = NEW.parent_id
              AND parent.slug = 'wheel-components'
              AND parent.parent_id IS NULL
              AND parent.depth = 1
              AND parent.is_enabled
        )
    ) THEN
        RAISE EXCEPTION 'tire must remain enabled directly under wheel-components' USING ERRCODE = '23514';
    END IF;

    RETURN NEW;
END`

const productCatalogBicycleTireProductTypeAndCategoryBindingFunctionBody = `DECLARE
    category_slug TEXT;
BEGIN
    IF TG_OP = 'UPDATE'
       AND NEW.product_type IS DISTINCT FROM OLD.product_type
       AND (NEW.product_type = 'bicycle_tire' OR OLD.product_type = 'bicycle_tire') THEN
        RAISE EXCEPTION 'Dedicated bicycle_tire product type is immutable' USING ERRCODE = '23514';
    END IF;

    IF NEW.product_type = 'bicycle_tire' THEN
        SELECT slug INTO category_slug
        FROM product_categories
        WHERE id = NEW.product_category_id;

        IF category_slug IS DISTINCT FROM 'tire' THEN
            RAISE EXCEPTION 'bicycle_tire products must belong to the tire category' USING ERRCODE = '23514';
        END IF;
    END IF;

    RETURN NEW;
END`

const productCatalogWheelsetProductTypeAndCategoryBindingFunctionBody = `DECLARE
    category_slug TEXT;
    category_enabled BOOLEAN;
BEGIN
    IF TG_OP = 'UPDATE'
       AND NEW.product_type IS DISTINCT FROM OLD.product_type
       AND (NEW.product_type = 'wheelset' OR OLD.product_type = 'wheelset') THEN
        RAISE EXCEPTION 'Dedicated wheelset product type is immutable' USING ERRCODE = '23514';
    END IF;

    IF NEW.product_type = 'wheelset' THEN
        SELECT slug, is_enabled INTO category_slug, category_enabled
        FROM product_categories
        WHERE id = NEW.product_category_id;

        IF category_slug IS DISTINCT FROM 'wheelset' OR category_enabled IS DISTINCT FROM TRUE THEN
            RAISE EXCEPTION 'wheelset products must belong to the enabled wheelset system category' USING ERRCODE = '23514';
        END IF;
    END IF;

    RETURN NEW;
END`

const productCatalogWheelsetSystemCategoryIdentityProtectionFunctionBody = `BEGIN
    IF TG_OP <> 'INSERT' AND OLD.slug = 'wheelset' THEN
        IF TG_OP = 'DELETE' THEN
            RAISE EXCEPTION 'The wheelset system category cannot be deleted' USING ERRCODE = '23514';
        END IF;
        IF NEW.id IS DISTINCT FROM OLD.id OR NEW.slug IS DISTINCT FROM OLD.slug THEN
            RAISE EXCEPTION 'The wheelset system category identity cannot be changed' USING ERRCODE = '23514';
        END IF;
    END IF;

    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;

    IF NEW.slug = 'wheelset' AND NOT NEW.is_enabled THEN
        RAISE EXCEPTION 'The wheelset system category must remain enabled' USING ERRCODE = '23514';
    END IF;

    RETURN NEW;
END`

const productCatalogSpokesSystemCategoryIdentityAndTreeProtectionFunctionBody = `BEGIN
    IF TG_OP <> 'INSERT' AND OLD.slug = 'spokes' THEN
        IF TG_OP = 'DELETE' THEN
            RAISE EXCEPTION 'The spokes system category cannot be deleted' USING ERRCODE = '23514';
        END IF;
        IF NEW.id IS DISTINCT FROM OLD.id OR NEW.slug IS DISTINCT FROM OLD.slug THEN
            RAISE EXCEPTION 'The spokes system category identity cannot be changed' USING ERRCODE = '23514';
        END IF;
    END IF;

    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;

    IF NEW.slug = 'spokes' AND (
        NEW.depth <> 2 OR NOT NEW.is_enabled OR NOT EXISTS (
            SELECT 1 FROM product_categories parent
            WHERE parent.id = NEW.parent_id
              AND parent.slug = 'wheel-components'
              AND parent.parent_id IS NULL
              AND parent.depth = 1
              AND parent.is_enabled
        )
    ) THEN
        RAISE EXCEPTION 'spokes must remain enabled directly under wheel-components' USING ERRCODE = '23514';
    END IF;

    RETURN NEW;
END`

const productCatalogRejectProductBindingToUnimplementedSpokesCategoryFunctionBody = `DECLARE
    category_slug TEXT;
BEGIN
    SELECT slug INTO category_slug
    FROM product_categories
    WHERE id = NEW.product_category_id;

    IF category_slug = 'spokes' THEN
        RAISE EXCEPTION 'spokes products require the dedicated spoke product type, which is not implemented yet' USING ERRCODE = '23514';
    END IF;

    RETURN NEW;
END`

const productCatalogSpokeProductTypeAndCategoryBindingFunctionBody = `DECLARE
    category_slug TEXT;
    category_enabled BOOLEAN;
BEGIN
    IF TG_OP = 'UPDATE'
       AND NEW.product_type IS DISTINCT FROM OLD.product_type
       AND (NEW.product_type = 'spoke' OR OLD.product_type = 'spoke') THEN
        RAISE EXCEPTION 'Dedicated spoke product type is immutable' USING ERRCODE = '23514';
    END IF;

    SELECT slug, is_enabled INTO category_slug, category_enabled
    FROM product_categories
    WHERE id = NEW.product_category_id;

    IF NEW.product_type = 'spoke' THEN
        IF category_slug IS DISTINCT FROM 'spokes' OR category_enabled IS DISTINCT FROM TRUE THEN
            RAISE EXCEPTION 'spoke products must belong to the enabled spokes system category' USING ERRCODE = '23514';
        END IF;
    ELSIF category_slug = 'spokes' THEN
        RAISE EXCEPTION 'spokes is reserved for dedicated spoke products' USING ERRCODE = '23514';
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
		{"product_categories", "protect_bicycle_tire_system_category_identity_and_tree_position", "", productCatalogBicycleTireSystemCategoryIdentityAndTreeProtectionFunctionBody, 31},
		{"products", "require_bicycle_tire_product_type_and_tire_category", "product_category_id,product_type", productCatalogBicycleTireProductTypeAndCategoryBindingFunctionBody, 23},
		{"product_categories", "protect_wheelset_system_category_identity", "", productCatalogWheelsetSystemCategoryIdentityProtectionFunctionBody, 31},
		{"products", "require_wheelset_product_type_and_system_category_binding", "product_category_id,product_type", productCatalogWheelsetProductTypeAndCategoryBindingFunctionBody, 23},
		{"product_categories", "protect_spokes_system_category_identity_and_tree_position", "", productCatalogSpokesSystemCategoryIdentityAndTreeProtectionFunctionBody, 31},
		{"products", "require_spoke_product_type_and_system_category_binding", "product_category_id,product_type", productCatalogSpokeProductTypeAndCategoryBindingFunctionBody, 23},
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

type productCatalogProtectedDedicatedSystemCategoryIdentities struct {
	WheelComponentsID int64
	CarbonRimID       int64
	BicycleTireID     int64
	WheelsetID        int64
	SpokesID          int64
}

func inspectProductCatalogRequiredDedicatedSystemCategoriesAndBindings(ctx context.Context, tx *sql.Tx) (productCatalogProtectedDedicatedSystemCategoryIdentities, error) {
	var identities productCatalogProtectedDedicatedSystemCategoryIdentities
	if err := tx.QueryRowContext(ctx, `
SELECT root.id, carbon.id, tire.id, wheelset.id, spokes.id
FROM public.product_categories root
JOIN public.product_categories carbon ON carbon.parent_id=root.id
JOIN public.product_categories tire ON tire.parent_id=root.id
JOIN public.product_categories spokes ON spokes.parent_id=root.id
JOIN public.product_categories wheelset ON wheelset.slug='wheelset' AND wheelset.is_enabled
WHERE root.slug='wheel-components' AND root.parent_id IS NULL AND root.depth=1 AND root.is_enabled
  AND carbon.slug='carbon-rim' AND carbon.depth=2 AND carbon.is_enabled
	AND tire.slug='tire' AND tire.depth=2 AND tire.is_enabled
	AND spokes.slug='spokes' AND spokes.depth=2 AND spokes.is_enabled`).Scan(&identities.WheelComponentsID, &identities.CarbonRimID, &identities.BicycleTireID, &identities.WheelsetID, &identities.SpokesID); err != nil {
		if err == sql.ErrNoRows {
			return identities, errors.New("product restore requires the enabled wheel-components root, its carbon-rim, tire and spokes categories, and the enabled wheelset category")
		}
		return identities, err
	}
	var invalid bool
	if err := tx.QueryRowContext(ctx, `
SELECT EXISTS(
    SELECT 1 FROM public.products
    WHERE (product_type='carbon_rim' AND product_category_id IS DISTINCT FROM $1)
       OR (product_type<>'carbon_rim' AND product_category_id=$1)
       OR (product_type='bicycle_tire' AND product_category_id IS DISTINCT FROM $2)
       OR (product_type='wheelset' AND product_category_id IS DISTINCT FROM $3)
       OR (product_type='spoke' AND product_category_id IS DISTINCT FROM $4)
       OR (product_type<>'spoke' AND product_category_id=$4)
)`, identities.CarbonRimID, identities.BicycleTireID, identities.WheelsetID, identities.SpokesID).Scan(&invalid); err != nil {
		return identities, err
	}
	if invalid {
		return identities, errors.New("product restore would break dedicated product type and system category bindings")
	}
	return identities, nil
}

func lockProductCatalogWriteTablesAndValidateStandardBusinessContractsBeforeMerge(ctx context.Context, tx *sql.Tx, snapshot productCatalogModuleSnapshot) (productCatalogProtectedDedicatedSystemCategoryIdentities, error) {
	var identities productCatalogProtectedDedicatedSystemCategoryIdentities
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
	identities, err := inspectProductCatalogRequiredDedicatedSystemCategoriesAndBindings(ctx, tx)
	if err != nil {
		return identities, err
	}
	for _, row := range snapshot.Tables["products"].Rows {
		id, err := readBusinessModuleSnapshotRowColumnText(row, "id")
		if err != nil {
			return identities, err
		}
		incomingType, err := readBusinessModuleSnapshotRowColumnText(row, "product_type")
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
		if currentType != incomingType && (currentType == "carbon_rim" || incomingType == "carbon_rim" || currentType == "bicycle_tire" || incomingType == "bicycle_tire" || currentType == "wheelset" || incomingType == "wheelset" || currentType == "spoke" || incomingType == "spoke") {
			return identities, errors.New("product restore cannot change an existing dedicated product type")
		}
	}
	return identities, nil
}
