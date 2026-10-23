package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"commerce-platform/internal/domain/product"
	"github.com/lib/pq"
)

const productCatalogMediaPlacementGuardVersion = 1

type productCatalogRegisteredMediaReferenceIdentity struct {
	AssetID, MediaType string
}

func (cfg Configuration) indexProductCatalogRegisteredMediaReferenceIdentities(snapshot productCatalogModuleSnapshot) (map[string][]productCatalogRegisteredMediaReferenceIdentity, map[string]string, error) {
	assets := map[string]string{}
	for _, row := range snapshot.Tables["media_assets"].Rows {
		id, err := productCatalogRowString(row, "id")
		if err != nil {
			return nil, nil, err
		}
		mediaType, err := productCatalogRowString(row, "media_type")
		if err != nil {
			return nil, nil, err
		}
		assets[id] = mediaType
	}
	identities := map[string][]productCatalogRegisteredMediaReferenceIdentity{}
	for _, table := range []string{"media_assets", "media_asset_derivatives"} {
		for _, row := range snapshot.Tables[table].Rows {
			idColumn := "id"
			if table == "media_asset_derivatives" {
				idColumn = "media_asset_id"
			}
			assetID, err := productCatalogRowString(row, idColumn)
			if err != nil {
				return nil, nil, err
			}
			key, err := productCatalogRowString(row, "storage_key")
			if err != nil {
				return nil, nil, err
			}
			if key == "" {
				address, err := productCatalogRowString(row, "url")
				if err != nil {
					return nil, nil, err
				}
				_, key = cfg.resolveLocalMediaReference(address)
			}
			if key != "" {
				identities[key] = append(identities[key], productCatalogRegisteredMediaReferenceIdentity{assetID, assets[assetID]})
			}
		}
	}
	return identities, assets, nil
}

func (cfg Configuration) validateProductCatalogMediaReferenceAssetIdentityAndType(reference, requestedAssetID, expectedType string, required bool, identities map[string][]productCatalogRegisteredMediaReferenceIdentity, assets map[string]string) (string, error) {
	if strings.TrimSpace(reference) == "" {
		if required || requestedAssetID != "" {
			return "", errors.New("product media binding requires a persisted URL")
		}
		return "", nil
	}
	root, key := cfg.resolveLocalMediaReference(reference)
	if root != "uploads" || key == "" {
		if _, _, err := cfg.resolveProductCatalogImageSource(reference); err != nil {
			return "", err
		}
		if requestedAssetID != "" {
			return "", errors.New("product media asset ID and URL identify different files")
		}
		return "", nil // Bundled /company/ images have no shared media ID.
	}
	matches := identities[key]
	if len(matches) > 1 {
		return "", errors.New("product required image has multiple shared object registrations")
	}
	if len(matches) == 0 {
		if requestedAssetID != "" {
			return "", errors.New("product media asset ID and URL identify different files")
		}
		return "", nil // Exact legacy placements are checked before the merge.
	}
	identity := matches[0]
	if requestedAssetID != "" && requestedAssetID != identity.AssetID {
		return "", errors.New("product media asset ID and URL identify different files")
	}
	if _, exists := assets[identity.AssetID]; !exists || identity.MediaType != expectedType {
		return "", errors.New("product media type does not match its registered asset")
	}
	return identity.AssetID, nil
}

// Validate the complete final scope, including retained rows and tombstones.
// Independent FKs do not prove that a variant and its media share a product.
func (cfg Configuration) validateProductCatalogCompleteMediaPlacementOwnershipAndAssetBindings(snapshot productCatalogModuleSnapshot) error {
	identities, assets, err := cfg.indexProductCatalogRegisteredMediaReferenceIdentities(snapshot)
	if err != nil {
		return err
	}
	owners := map[string]map[string]string{}
	for _, table := range []string{"product_variants", "product_variant_option_values"} {
		owners[table] = map[string]string{}
		for _, row := range snapshot.Tables[table].Rows {
			id, err := productCatalogRowString(row, "id")
			if err != nil {
				return err
			}
			owner, err := productCatalogRowString(row, "product_id")
			if err != nil {
				return err
			}
			owners[table][id] = owner
		}
	}
	for _, row := range snapshot.Tables["product_media"].Rows {
		field := func(column string) string { value, _ := productCatalogRowString(row, column); return value }
		for _, relation := range []struct{ column, table string }{{"variant_id", "product_variants"}, {"variant_option_value_id", "product_variant_option_values"}} {
			if id := field(relation.column); id != "" && owners[relation.table][id] != field("product_id") {
				return fmt.Errorf("product media %s has a variant or option value belonging to another product", field("id"))
			}
		}
		mainAssetID, err := cfg.validateProductCatalogMediaReferenceAssetIdentityAndType(field("url"), field("media_asset_id"), field("media_type"), true, identities, assets)
		if err != nil {
			return fmt.Errorf("product media %s: %w", field("id"), err)
		}
		for _, column := range []string{"thumbnail_url", "poster_url"} {
			if _, err := cfg.validateProductCatalogMediaReferenceAssetIdentityAndType(field(column), "", "image", false, identities, assets); err != nil {
				return fmt.Errorf("product media %s %s: %w", field("id"), column, err)
			}
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(row, &fields); err != nil {
			return err
		}
		var variants map[string]product.ProductMediaImageVariant
		if err := decodeStrictJSON(fields["image_variants"], &variants); err != nil || variants == nil {
			return errors.New("product image variants require the registered object format")
		}
		for preset, variant := range variants {
			if preset == "" || strings.TrimSpace(preset) != preset || variant.Width < 0 || variant.Height < 0 || mainAssetID == "" || field("media_type") != "image" {
				return errors.New("product image variants require a registered parent image")
			}
			assetID, err := cfg.validateProductCatalogMediaReferenceAssetIdentityAndType(variant.URL, mainAssetID, "image", true, identities, assets)
			if err != nil || assetID != mainAssetID {
				return errors.New("product image variant URL does not belong to its parent media asset")
			}
		}
	}
	for _, table := range []string{"product_variant_option_values", "product_spec_option_items"} {
		for _, row := range snapshot.Tables[table].Rows {
			address, err := productCatalogRowString(row, "swatch_url")
			if err != nil {
				return err
			}
			assetID, err := productCatalogRowString(row, "swatch_media_asset_id")
			if err != nil {
				return err
			}
			if _, err := cfg.validateProductCatalogMediaReferenceAssetIdentityAndType(address, assetID, "image", false, identities, assets); err != nil {
				return fmt.Errorf("product swatch in %s: %w", table, err)
			}
		}
	}
	return nil
}

func (cfg Configuration) validateProductCatalogIncomingMediaOwnersAndUnregisteredLegacyURLsBeforeMerge(ctx context.Context, tx *sql.Tx, snapshot productCatalogModuleSnapshot) error {
	identities, _, err := cfg.indexProductCatalogRegisteredMediaReferenceIdentities(snapshot)
	if err != nil {
		return err
	}
	for _, placement := range []struct {
		table, owner string
		columns      []string
	}{
		{"product_media", "product_id", []string{"url", "thumbnail_url", "poster_url"}},
		{"product_variant_option_values", "product_id", []string{"swatch_url"}},
		{"product_spec_option_items", "spec_definition_id", []string{"swatch_url"}},
	} {
		for _, row := range snapshot.Tables[placement.table].Rows {
			id, err := productCatalogRowString(row, "id")
			if err != nil {
				return err
			}
			owner, err := productCatalogRowString(row, placement.owner)
			if err != nil {
				return err
			}
			var previousRow string
			err = tx.QueryRowContext(ctx, `SELECT to_jsonb(m)::text FROM public.`+pq.QuoteIdentifier(placement.table)+` m WHERE id::text=$1`, id).Scan(&previousRow)
			if err != nil && err != sql.ErrNoRows {
				return err
			}
			if err == nil {
				previousOwner, err := productCatalogRowString(json.RawMessage(previousRow), placement.owner)
				if err != nil {
					return err
				}
				if owner != previousOwner {
					return errors.New("existing product media identity cannot move to another product")
				}
			}
			legacyPlacementRetained := previousRow != ""
			if placement.table == "product_variant_option_values" && legacyPlacementRetained {
				previousDefinition, _ := productCatalogRowString(json.RawMessage(previousRow), "spec_definition_id")
				incomingDefinition, err := productCatalogRowString(row, "spec_definition_id")
				if err != nil {
					return err
				}
				legacyPlacementRetained = previousDefinition == incomingDefinition
			}
			for _, column := range placement.columns {
				address, err := productCatalogRowString(row, column)
				if err != nil {
					return err
				}
				root, key := cfg.resolveLocalMediaReference(address)
				if root != "uploads" || key == "" || len(identities[key]) > 0 {
					continue
				}
				var registered bool
				if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM public.media_assets WHERE storage_key=$1 OR url=$2 UNION ALL SELECT 1 FROM public.media_asset_derivatives WHERE storage_key=$1 OR url=$2)`, key, address).Scan(&registered); err != nil {
					return err
				}
				previousAddress, _ := productCatalogRowString(json.RawMessage(previousRow), column)
				if !registered && (!legacyPlacementRetained || previousAddress != address) {
					return errors.New("unregistered product media URL can only remain on its exact existing placement and field")
				}
			}
		}
	}
	return nil
}
