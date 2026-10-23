package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/lib/pq"

	"commerce-platform/internal/pkg/safehtml"
	"commerce-platform/internal/pkg/storage"
)

const productCatalogRichTextMediaGuardVersion = 1

// Migration 435 aligns runtime writers with short_description. Older snapshots
// may still contain short_desc; keep their independent placement checks.
func productCatalogPersistedRichTextMediaColumns(row json.RawMessage) ([]string, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(row, &fields); err != nil {
		return nil, err
	}
	if _, exists := fields["description"]; !exists {
		return nil, errors.New("product row is missing its description column")
	}
	columns := []string{"description"}
	for _, column := range []string{"short_description", "short_desc"} {
		if _, exists := fields[column]; exists {
			columns = append(columns, column)
		}
	}
	if len(columns) == 1 {
		return nil, errors.New("product row is missing its persisted short description column")
	}
	return columns, nil
}

// Ordinary hyperlinks do not depend on a local file. Media sources and links
// into this site's managed file namespaces do; use the business HTML parser so
// text, alt labels, removed elements and coincidental substrings are excluded.
func (cfg Configuration) productCatalogRichTextAttributeRequiresLocalFile(reference safehtml.RichTextMediaAttributeReference) bool {
	if reference.ExpectedMediaType != "" {
		return true
	}
	parsed, err := url.Parse(strings.TrimSpace(*reference.URL))
	if err != nil {
		return false
	}
	if parsed.Host != "" {
		recognized := false
		for _, origin := range cfg.PublicOrigins {
			configured, err := url.Parse(origin)
			if err == nil && configured.Scheme == parsed.Scheme && strings.EqualFold(configured.Host, parsed.Host) {
				recognized = true
			}
		}
		if !recognized {
			return false
		}
	}
	for _, prefix := range []string{"/uploads/", "/uploads-private/", "/site-logo-uploads/", "/company/"} {
		if strings.HasPrefix(parsed.Path, prefix) {
			return true
		}
	}
	return false
}

func (cfg Configuration) resolveProductCatalogRichTextDurableMediaSource(reference string) (string, string, error) {
	parsed, err := url.Parse(strings.TrimSpace(reference))
	if err != nil || parsed.Scheme != "" && !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") || parsed.Host != "" && parsed.Scheme == "" {
		return "", "", errors.New("product rich text media requires a durable local or trusted HTTP URL")
	}
	if _, _, err := cfg.resolveProductCatalogImageSource(reference); err != nil {
		return "", "", err
	}
	root, key := cfg.resolveLocalMediaReference(reference)
	return root, key, nil
}

func (cfg Configuration) validateProductCatalogCompleteRichTextRegisteredMediaTypes(snapshot productCatalogModuleSnapshot) error {
	identities, assets, err := cfg.indexProductCatalogRegisteredMediaReferenceIdentities(snapshot)
	if err != nil {
		return err
	}
	for _, table := range []string{"products", "product_information_templates", "product_categories", "product_category_translations"} {
		for _, row := range snapshot.Tables[table].Rows {
			columns, err := productCatalogPersistedRichTextMediaColumnsForTable(table, row)
			if err != nil {
				return err
			}
			for _, column := range columns {
				fieldIdentity := productCatalogRichTextFieldDisplayIdentity(table, column)
				value, err := productCatalogRowString(row, column)
				if err != nil {
					return err
				}
				document, err := safehtml.ParseRichTextMediaSourceAndFileLinkAttributes(value)
				if err != nil {
					return err
				}
				for _, reference := range document.References {
					if !cfg.productCatalogRichTextAttributeRequiresLocalFile(reference) {
						continue
					}
					root, key, err := cfg.resolveProductCatalogRichTextDurableMediaSource(*reference.URL)
					if err != nil {
						return fmt.Errorf("product rich text %s %s: %w", fieldIdentity, reference.ElementAndAttribute, err)
					}
					if root != "uploads" {
						continue // Bundled company images have no shared registration.
					}
					matches := identities[key]
					if len(matches) > 1 {
						return errors.New("product rich text media has multiple shared object registrations")
					}
					if len(matches) == 1 {
						identity := matches[0]
						if _, exists := assets[identity.AssetID]; !exists || reference.ExpectedMediaType != "" && identity.MediaType != reference.ExpectedMediaType {
							return fmt.Errorf("product rich text %s %s media type does not match its registered asset", fieldIdentity, reference.ElementAndAttribute)
						}
					}
				}
			}
		}
	}
	return nil
}

func (cfg Configuration) validateProductCatalogIncomingRichTextLegacyMediaOccurrencesBeforeMerge(ctx context.Context, tx *sql.Tx, snapshot productCatalogModuleSnapshot) error {
	identities, _, err := cfg.indexProductCatalogRegisteredMediaReferenceIdentities(snapshot)
	if err != nil {
		return err
	}
	dedicatedAssets := map[string]bool{}
	for _, row := range snapshot.Tables["media_assets"].Rows {
		id, _ := productCatalogRowString(row, "id")
		key, _ := productCatalogRowString(row, "storage_key")
		address, _ := productCatalogRowString(row, "url")
		_, urlKey := cfg.resolveLocalMediaReference(address)
		dedicatedAssets[id] = storage.DedicatedPublicObjectNamespaceForStorageKey(key) != "" || storage.DedicatedPublicObjectNamespaceForStorageKey(urlKey) != ""
	}
	for _, table := range []string{"products", "product_information_templates", "product_categories", "product_category_translations"} {
		for _, row := range snapshot.Tables[table].Rows {
			columns, err := productCatalogPersistedRichTextMediaColumnsForTable(table, row)
			if err != nil {
				return err
			}
			id, err := productCatalogRowString(row, "id")
			if err != nil {
				return err
			}
			var previous string
			err = tx.QueryRowContext(ctx, `SELECT to_jsonb(p)::text FROM public.`+pq.QuoteIdentifier(table)+` p WHERE id::text=$1`, id).Scan(&previous)
			if err != nil && err != sql.ErrNoRows {
				return err
			}
			for _, column := range columns {
				value, err := productCatalogRowString(row, column)
				if err != nil {
					return err
				}
				previousValue := ""
				if previous != "" {
					previousValue, err = productCatalogRowString(json.RawMessage(previous), column)
					if err != nil {
						return err
					}
				}
				document, err := safehtml.ParseRichTextMediaSourceAndFileLinkAttributes(value)
				if err != nil {
					return err
				}
				oldDocument, err := safehtml.ParseRichTextMediaSourceAndFileLinkAttributes(previousValue)
				if err != nil {
					return err
				}
				previousOccurrences := map[string]int{}
				for _, reference := range oldDocument.References {
					previousOccurrences[reference.ElementAndAttribute+"\x00"+*reference.URL]++
				}
				for _, reference := range document.References {
					occurrenceKey := reference.ElementAndAttribute + "\x00" + *reference.URL
					retained := previousOccurrences[occurrenceKey] > 0
					if retained {
						previousOccurrences[occurrenceKey]--
					}
					if !cfg.productCatalogRichTextAttributeRequiresLocalFile(reference) {
						continue
					}
					root, key, err := cfg.resolveProductCatalogRichTextDurableMediaSource(*reference.URL)
					if err != nil {
						return err
					}
					if root != "uploads" {
						continue
					}
					registered := len(identities[key]) > 0
					dedicated := storage.DedicatedPublicObjectNamespaceForStorageKey(key) != ""
					for _, identity := range identities[key] {
						dedicated = dedicated || dedicatedAssets[identity.AssetID]
					}
					if !registered {
						rows, err := tx.QueryContext(ctx, `SELECT COALESCE(storage_key,''),COALESCE(url,'') FROM public.media_assets WHERE storage_key=$1 OR url=$2 OR id IN (SELECT media_asset_id FROM public.media_asset_derivatives WHERE storage_key=$1 OR url=$2)`, key, *reference.URL)
						if err != nil {
							return err
						}
						for rows.Next() {
							var assetKey, assetURL string
							if err := rows.Scan(&assetKey, &assetURL); err != nil {
								_ = rows.Close()
								return err
							}
							_, urlKey := cfg.resolveLocalMediaReference(assetURL)
							registered = true
							dedicated = dedicated || storage.DedicatedPublicObjectNamespaceForStorageKey(assetKey) != "" || storage.DedicatedPublicObjectNamespaceForStorageKey(urlKey) != ""
						}
						err = rows.Err()
						_ = rows.Close()
						if err != nil {
							return err
						}
					}
					if !retained && dedicated {
						return errors.New("product rich text dedicated lifecycle media can only retain its existing field, attribute and occurrence count")
					}
					if !registered && !retained {
						return errors.New("unregistered product rich text media can only retain its existing product, field, attribute, URL and occurrence count")
					}
				}
			}
		}
	}
	return nil
}
