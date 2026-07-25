package product

// ProductListResponse 产品列表响应
type ProductListResponse struct {
	ID           uint     `json:"id"`
	SKU          string   `json:"sku"`
	Name         string   `json:"name"`
	Slug         string   `json:"slug"`
	ShortDesc    string   `json:"short_description"`
	Price        float64  `json:"price"`
	SalePrice    *float64 `json:"sale_price"`
	Stock        int      `json:"stock"`
	Status       string   `json:"status"`
	Locale       string   `json:"locale"`
	Featured     bool     `json:"featured"`
	FeaturedImg  string   `json:"featured_image"`
	VariantCount int      `json:"variant_count"`
}

// ToListResponse 转换为列表响应
func (p *Product) ToListResponse() *ProductListResponse {
	price, salePrice := p.DisplayPrices()
	resp := &ProductListResponse{
		ID:           p.ID,
		SKU:          p.DisplaySKU(),
		Name:         p.Name,
		Slug:         p.Slug,
		ShortDesc:    p.ShortDesc,
		Price:        price,
		SalePrice:    salePrice,
		Stock:        p.TotalVariantStock(),
		Status:       p.Status,
		Locale:       p.Locale,
		Featured:     p.Featured,
		VariantCount: len(p.Variants),
	}

	if len(p.Media) > 0 {
		for _, item := range p.Media {
			if item.MediaType == "image" && item.IsVisible && item.URL != "" {
				resp.FeaturedImg = item.URL
				break
			}
		}
	}

	return resp
}
