package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/pkg/invoice"
)

type PayPalDisputeInvoiceOptions struct {
	Seller        invoice.SellerProfile
	FontPath      string
	AutoAttachPDF bool
}

type PayPalDisputeCommercialInvoicePDF struct {
	DisputeID       uint
	PayPalDisputeID string
	DocumentNumber  string
	Filename        string
	Bytes           []byte
}

func (s *PaymentService) BuildPayPalDisputeCommercialInvoicePDF(disputeID uint) (*PayPalDisputeCommercialInvoicePDF, error) {
	pkg, err := s.BuildPayPalDisputeEvidencePackage(disputeID)
	if err != nil {
		return nil, err
	}
	if pkg == nil || pkg.Dispute == nil {
		return nil, ErrPayPalDisputeInvoiceUnavailable
	}

	document, options, err := s.paypalDisputeCommercialInvoice(pkg, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPayPalDisputeInvoiceUnavailable, err)
	}
	pdf, err := s.renderPayPalCommercialInvoicePDF(document, options)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPayPalDisputeInvoiceUnavailable, err)
	}

	return &PayPalDisputeCommercialInvoicePDF{
		DisputeID:       pkg.Dispute.ID,
		PayPalDisputeID: pkg.Dispute.PayPalDisputeID,
		DocumentNumber:  document.DocumentNumber,
		Filename:        paypalEvidenceDocumentName(document.DocumentNumber, "commercial-invoice"),
		Bytes:           pdf,
	}, nil
}

func (s *PaymentService) RenderPayPalCommercialInvoicePreview(document invoice.CommercialInvoice) (*PayPalDisputeCommercialInvoicePDF, error) {
	options := PayPalDisputeInvoiceOptions{}
	if s != nil {
		options = s.paypalDisputeInvoiceOptions
	}
	if !document.Seller.Configured() {
		configuredSeller, err := s.paypalDisputeInvoiceSellerProfile()
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrPayPalDisputeInvoiceUnavailable, err)
		}
		document.Seller = configuredSeller
	}
	if strings.TrimSpace(document.DocumentNumber) == "" {
		document.DocumentNumber = "CI-PREVIEW"
	}
	if document.DocumentDate.IsZero() {
		document.DocumentDate = time.Now().UTC()
	}
	if strings.TrimSpace(document.Currency) == "" {
		document.Currency = "USD"
	}
	if strings.TrimSpace(document.Disclaimer) == "" {
		document.Disclaimer = "Commercial invoice / order receipt prepared for payment dispute evidence. This document is not a statutory tax invoice unless the seller's applicable tax requirements are satisfied."
	}
	if err := document.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPayPalDisputeInvoiceUnavailable, err)
	}
	pdf, err := s.renderPayPalCommercialInvoicePDF(document, options)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPayPalDisputeInvoiceUnavailable, err)
	}
	return &PayPalDisputeCommercialInvoicePDF{
		DocumentNumber: document.DocumentNumber,
		Filename:       paypalEvidenceDocumentName(document.DocumentNumber, "commercial-invoice"),
		Bytes:          pdf,
	}, nil
}

func (s *PaymentService) renderPayPalCommercialInvoicePDF(document invoice.CommercialInvoice, options PayPalDisputeInvoiceOptions) ([]byte, error) {
	return invoice.RenderCommercialInvoicePDF(document, options.FontPath)
}

func (s *PaymentService) paypalDisputeInvoiceAutoAttachEnabled() bool {
	if s == nil {
		return false
	}
	return s.paypalDisputeInvoiceOptions.AutoAttachPDF
}

func (s *PaymentService) paypalDisputeCommercialInvoice(pkg *PayPalDisputeEvidencePackage, generatedAt time.Time) (invoice.CommercialInvoice, PayPalDisputeInvoiceOptions, error) {
	if pkg == nil || pkg.Order == nil {
		return invoice.CommercialInvoice{}, PayPalDisputeInvoiceOptions{}, errors.New("paypal dispute is not linked to a local order")
	}
	options := PayPalDisputeInvoiceOptions{}
	if s != nil {
		options = s.paypalDisputeInvoiceOptions
	}
	seller, err := s.paypalDisputeInvoiceSellerProfile()
	if err != nil {
		return invoice.CommercialInvoice{}, options, err
	}
	options.Seller = seller
	if !options.Seller.Configured() {
		return invoice.CommercialInvoice{}, options, errors.New("commercial invoice seller name and address are not configured")
	}

	paymentReference := ""
	if pkg.Dispute != nil {
		paymentReference = pkg.Dispute.ProviderPaymentID
	}
	document, err := invoice.BuildFromOrder(pkg.Order, options.Seller, paymentReference, generatedAt)
	return document, options, err
}

func (s *PaymentService) paypalDisputeInvoiceSellerProfile() (invoice.SellerProfile, error) {
	if s != nil && s.paypalDisputeInvoiceSellerProfileProvider != nil {
		profile, err := s.paypalDisputeInvoiceSellerProfileProvider.SellerProfile()
		if err != nil {
			return invoice.SellerProfile{}, err
		}
		if profile.Configured() {
			return profile, nil
		}
	}

	if s != nil && s.paypalDisputeInvoiceOptions.Seller.Configured() {
		return s.paypalDisputeInvoiceOptions.Seller, nil
	}
	return invoice.SellerProfileFromEnv(), nil
}
