package product

import productdomain "commerce-platform/internal/domain/product"

func publicProductTranslationRoutesFromDomain(items []productdomain.ProductTranslationRoute) []PublicProductTranslationRoute {
	if len(items) == 0 {
		return nil
	}

	routes := make([]PublicProductTranslationRoute, 0, len(items))
	for _, item := range items {
		if item.Locale == "" || item.Slug == "" {
			continue
		}
		routes = append(routes, PublicProductTranslationRoute{
			Locale: item.Locale,
			Slug:   item.Slug,
		})
	}
	if len(routes) == 0 {
		return nil
	}
	return routes
}
