package procurement

// ProductOption is the intentionally small catalog projection used by the
// procurement product picker. It is not a procurement record.
type ProductOption struct {
	ProductName  string `json:"product_name"`
	VariantTitle string `json:"variant_title"`
	SKU          string `json:"sku"`
	Available    bool   `json:"available"`
}
