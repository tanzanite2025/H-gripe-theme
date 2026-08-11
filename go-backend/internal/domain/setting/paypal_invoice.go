package setting

import "strings"

const PayPalDisputeInvoiceSellerProfileGroup = "paypal_dispute_invoice_seller"

const (
	PayPalDisputeInvoiceSellerProfileKeyName    = "paypal_dispute_invoice_seller_name"
	PayPalDisputeInvoiceSellerProfileKeyAddress = "paypal_dispute_invoice_seller_address"
	PayPalDisputeInvoiceSellerProfileKeyEmail   = "paypal_dispute_invoice_seller_email"
	PayPalDisputeInvoiceSellerProfileKeyPhone   = "paypal_dispute_invoice_seller_phone"
	PayPalDisputeInvoiceSellerProfileKeyWebsite = "paypal_dispute_invoice_seller_website"
	PayPalDisputeInvoiceSellerProfileKeyTaxID   = "paypal_dispute_invoice_seller_tax_id"
)

type PayPalDisputeInvoiceSellerProfileSettings struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Website string `json:"website"`
	TaxID   string `json:"tax_id"`
}

func (profile PayPalDisputeInvoiceSellerProfileSettings) Configured() bool {
	return strings.TrimSpace(profile.Name) != "" && strings.TrimSpace(profile.Address) != ""
}

type PayPalDisputeInvoiceSellerProfileUpdateRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Website string `json:"website"`
	TaxID   string `json:"tax_id"`
}

func (request PayPalDisputeInvoiceSellerProfileUpdateRequest) Settings() PayPalDisputeInvoiceSellerProfileSettings {
	return PayPalDisputeInvoiceSellerProfileSettings{
		Name:    strings.TrimSpace(request.Name),
		Address: strings.TrimSpace(request.Address),
		Email:   strings.TrimSpace(request.Email),
		Phone:   strings.TrimSpace(request.Phone),
		Website: strings.TrimSpace(request.Website),
		TaxID:   strings.TrimSpace(request.TaxID),
	}
}
