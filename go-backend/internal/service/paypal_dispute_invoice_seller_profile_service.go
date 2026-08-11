package service

import (
	"errors"
	"fmt"
	"strings"

	settingdomain "tanzanite/internal/domain/setting"
	"tanzanite/internal/pkg/invoice"
	"tanzanite/internal/repository"
)

var ErrPayPalDisputeInvoiceSellerProfileIncomplete = errors.New("PayPal commercial invoice seller name and address are required")

type PayPalDisputeInvoiceSellerProfileService struct {
	settings *SettingService
}

func NewPayPalDisputeInvoiceSellerProfileService(settings *SettingService) *PayPalDisputeInvoiceSellerProfileService {
	return &PayPalDisputeInvoiceSellerProfileService{settings: settings}
}

func (s *PayPalDisputeInvoiceSellerProfileService) Get() (*settingdomain.PayPalDisputeInvoiceSellerProfileSettings, error) {
	if s == nil {
		return nil, errors.New("PayPal commercial invoice seller profile service is unavailable")
	}

	profile := sellerProfileSettingsFromEnv()
	if s.settings == nil {
		return &profile, nil
	}

	fields := []struct {
		key    string
		target *string
	}{
		{settingdomain.PayPalDisputeInvoiceSellerProfileKeyName, &profile.Name},
		{settingdomain.PayPalDisputeInvoiceSellerProfileKeyAddress, &profile.Address},
		{settingdomain.PayPalDisputeInvoiceSellerProfileKeyEmail, &profile.Email},
		{settingdomain.PayPalDisputeInvoiceSellerProfileKeyPhone, &profile.Phone},
		{settingdomain.PayPalDisputeInvoiceSellerProfileKeyWebsite, &profile.Website},
		{settingdomain.PayPalDisputeInvoiceSellerProfileKeyTaxID, &profile.TaxID},
	}
	for _, field := range fields {
		record, err := s.settings.Get(field.key, "global")
		if err != nil {
			if repository.IsRecordNotFound(err) {
				continue
			}
			return nil, err
		}
		if strings.TrimSpace(record.Value) != "" {
			*field.target = strings.TrimSpace(record.Value)
		}
	}

	return &profile, nil
}

func (s *PayPalDisputeInvoiceSellerProfileService) Update(request settingdomain.PayPalDisputeInvoiceSellerProfileUpdateRequest) (*settingdomain.PayPalDisputeInvoiceSellerProfileSettings, error) {
	if s == nil || s.settings == nil {
		return nil, errors.New("PayPal commercial invoice seller profile service is unavailable")
	}

	profile := request.Settings()
	if !profile.Configured() {
		return nil, ErrPayPalDisputeInvoiceSellerProfileIncomplete
	}

	records := []settingdomain.Setting{
		payPalInvoiceSellerProfileRecord(settingdomain.PayPalDisputeInvoiceSellerProfileKeyName, profile.Name, "Seller legal/business name"),
		payPalInvoiceSellerProfileRecord(settingdomain.PayPalDisputeInvoiceSellerProfileKeyAddress, profile.Address, "Seller commercial address"),
		payPalInvoiceSellerProfileRecord(settingdomain.PayPalDisputeInvoiceSellerProfileKeyEmail, profile.Email, "Seller contact email"),
		payPalInvoiceSellerProfileRecord(settingdomain.PayPalDisputeInvoiceSellerProfileKeyPhone, profile.Phone, "Seller contact phone"),
		payPalInvoiceSellerProfileRecord(settingdomain.PayPalDisputeInvoiceSellerProfileKeyWebsite, profile.Website, "Seller website"),
		payPalInvoiceSellerProfileRecord(settingdomain.PayPalDisputeInvoiceSellerProfileKeyTaxID, profile.TaxID, "Seller tax identifier"),
	}
	if err := s.settings.BatchSet(records); err != nil {
		return nil, err
	}

	return s.Get()
}

func (s *PayPalDisputeInvoiceSellerProfileService) SellerProfile() (invoice.SellerProfile, error) {
	profile, err := s.Get()
	if err != nil {
		return invoice.SellerProfile{}, err
	}
	return invoice.SellerProfile{
		Name:    profile.Name,
		Address: profile.Address,
		Email:   profile.Email,
		Phone:   profile.Phone,
		Website: profile.Website,
		TaxID:   profile.TaxID,
	}, nil
}

func sellerProfileSettingsFromEnv() settingdomain.PayPalDisputeInvoiceSellerProfileSettings {
	profile := invoice.SellerProfileFromEnv()
	return settingdomain.PayPalDisputeInvoiceSellerProfileSettings{
		Name:    profile.Name,
		Address: profile.Address,
		Email:   profile.Email,
		Phone:   profile.Phone,
		Website: profile.Website,
		TaxID:   profile.TaxID,
	}
}

func payPalInvoiceSellerProfileRecord(key, value, description string) settingdomain.Setting {
	return settingdomain.Setting{
		Key:         key,
		Value:       strings.TrimSpace(value),
		Type:        "string",
		Group:       settingdomain.PayPalDisputeInvoiceSellerProfileGroup,
		Locale:      "global",
		IsPublic:    false,
		Description: fmt.Sprintf("PayPal commercial invoice seller profile: %s", description),
	}
}
