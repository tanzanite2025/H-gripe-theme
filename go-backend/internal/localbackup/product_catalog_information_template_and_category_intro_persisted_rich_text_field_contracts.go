package localbackup

import (
	"encoding/json"
	"fmt"
)

const productCatalogTemplateAndCategoryRichTextGuardVersion = 1

// These are HTML fields in the actual editor/service contract. Category
// description and SEO metadata remain text and must not imply file ownership.
func productCatalogPersistedRichTextMediaColumnsForTable(table string, row json.RawMessage) ([]string, error) {
	if table == "products" {
		return productCatalogPersistedRichTextMediaColumns(row)
	}
	column := ""
	switch table {
	case "product_information_templates":
		column = "content"
	case "product_categories", "product_category_translations":
		column = "seo_intro"
	default:
		return nil, nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(row, &fields); err != nil {
		return nil, err
	}
	if _, exists := fields[column]; !exists {
		return nil, fmt.Errorf("product rich text %s is missing its %s column", table, column)
	}
	return []string{column}, nil
}

func productCatalogRichTextFieldDisplayIdentity(table, column string) string {
	if table == "products" {
		return column
	}
	return table + "." + column
}

func productCatalogKnownPlainTextFieldsExcludedFromMediaDiscovery(table string) []string {
	switch table {
	case "product_information_templates":
		return []string{"kind", "name", "slug", "locale"}
	case "product_categories", "product_category_translations":
		return []string{"name", "slug", "locale", "description", "meta_title", "meta_description"}
	default:
		return nil
	}
}
