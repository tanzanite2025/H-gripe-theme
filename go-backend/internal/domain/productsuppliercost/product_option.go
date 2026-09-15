package productsuppliercost

// ProductOption is the intentionally small catalog projection used by the
// supplier-cost product picker. It is not a supplier-side record.
type ProductOption struct {
	ProductName  string `json:"product_name"`
	VariantTitle string `json:"variant_title"`
	SKU          string `json:"sku"`
	Available    bool   `json:"available"`
}
