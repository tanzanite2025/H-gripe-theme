package admin

import "commerce-platform/internal/service"

type storefrontMarketRequest struct {
	Code                string   `json:"code" binding:"required"`
	Name                string   `json:"name"`
	Countries           []string `json:"countries"`
	DefaultLocale       string   `json:"default_locale"`
	SupportedLocales    []string `json:"supported_locales"`
	DefaultCurrency     string   `json:"default_currency"`
	DisplayCurrencies   []string `json:"display_currencies"`
	PaymentMethodPolicy string   `json:"payment_method_policy"`
	LogisticsPolicy     string   `json:"logistics_policy"`
	TaxPolicy           string   `json:"tax_policy"`
	Enabled             *bool    `json:"enabled"`
	Priority            int      `json:"priority"`
}

func (r storefrontMarketRequest) toServiceInput() service.StorefrontMarketInput {
	return service.StorefrontMarketInput{
		Code:                r.Code,
		Name:                r.Name,
		Countries:           r.Countries,
		DefaultLocale:       r.DefaultLocale,
		SupportedLocales:    r.SupportedLocales,
		DefaultCurrency:     r.DefaultCurrency,
		DisplayCurrencies:   r.DisplayCurrencies,
		PaymentMethodPolicy: r.PaymentMethodPolicy,
		LogisticsPolicy:     r.LogisticsPolicy,
		TaxPolicy:           r.TaxPolicy,
		Enabled:             r.Enabled,
		Priority:            r.Priority,
	}
}
