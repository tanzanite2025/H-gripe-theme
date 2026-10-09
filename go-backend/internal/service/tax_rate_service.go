package service

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/taxrate"
	"commerce-platform/internal/repository"
)

var ErrTaxRateRuleInvalid = errors.New("tax rate rule is invalid")

// TaxRateRuleInput contains only fields that determine checkout rule matching
// and calculation. Provider snapshots and legacy tax metadata are not inputs.
type TaxRateRuleInput struct {
	Name        string `json:"name"`
	Country     string `json:"country"`
	State       string `json:"state"`
	PostalCode  string `json:"postal_code"`
	RateDecimal string `json:"rate_decimal"`
	Priority    int    `json:"priority"`
	Enabled     bool   `json:"enabled"`
}

type TaxRateService struct {
	repository *repository.TaxRateRepository
}

func NewTaxRateService(taxRateRepository *repository.TaxRateRepository) *TaxRateService {
	return &TaxRateService{repository: taxRateRepository}
}

func (s *TaxRateService) ListTaxRates() ([]taxrate.TaxRate, error) {
	if s == nil || s.repository == nil {
		return nil, errors.New("tax rate service is not configured")
	}
	return s.repository.FindAllTaxRates(false)
}

func (s *TaxRateService) GetTaxRate(id uint) (*taxrate.TaxRate, error) {
	if s == nil || s.repository == nil {
		return nil, errors.New("tax rate service is not configured")
	}
	return s.repository.FindTaxRateByID(id)
}

func (s *TaxRateService) CreateTaxRateRule(input TaxRateRuleInput) (*taxrate.TaxRate, error) {
	if s == nil || s.repository == nil {
		return nil, errors.New("tax rate service is not configured")
	}
	input, err := normalizeTaxRateRuleInput(input)
	if err != nil {
		return nil, err
	}
	rate := &taxrate.TaxRate{
		Name:        input.Name,
		Country:     input.Country,
		State:       input.State,
		PostalCode:  input.PostalCode,
		RateDecimal: input.RateDecimal,
		Priority:    input.Priority,
		Enabled:     input.Enabled,
	}
	if err := s.repository.CreateTaxRate(rate); err != nil {
		return nil, fmt.Errorf("create tax rate rule: %w", err)
	}
	return rate, nil
}

func (s *TaxRateService) UpdateTaxRateRule(id uint, input TaxRateRuleInput) (*taxrate.TaxRate, error) {
	if s == nil || s.repository == nil {
		return nil, errors.New("tax rate service is not configured")
	}
	input, err := normalizeTaxRateRuleInput(input)
	if err != nil {
		return nil, err
	}
	rate, err := s.repository.FindTaxRateByID(id)
	if err != nil {
		return nil, err
	}
	rate.Name = input.Name
	rate.Country = input.Country
	rate.State = input.State
	rate.PostalCode = input.PostalCode
	rate.RateDecimal = input.RateDecimal
	rate.Priority = input.Priority
	rate.Enabled = input.Enabled
	if err := s.repository.UpdateTaxRate(rate); err != nil {
		return nil, fmt.Errorf("update tax rate rule: %w", err)
	}
	return rate, nil
}

func (s *TaxRateService) DeleteTaxRateRule(id uint) error {
	if s == nil || s.repository == nil {
		return errors.New("tax rate service is not configured")
	}
	if err := s.repository.DeleteTaxRate(id); err != nil {
		return fmt.Errorf("delete tax rate rule: %w", err)
	}
	return nil
}

func (s *TaxRateService) ListPublicTaxRates() ([]taxrate.TaxRate, error) {
	if s == nil || s.repository == nil {
		return nil, errors.New("tax rate service is not configured")
	}
	return s.repository.FindAllTaxRates(true)
}

func (s *TaxRateService) GetPublicTaxRate(id uint) (*taxrate.TaxRate, error) {
	rate, err := s.GetTaxRate(id)
	if err != nil {
		return nil, err
	}
	if !rate.Enabled {
		return nil, taxrate.ErrTaxRateNotFound
	}
	return rate, nil
}

func (s *TaxRateService) CalculateTaxMoney(
	amountMoney domainmoney.Money,
	country, state string,
	postalCodes ...string,
) (string, domainmoney.Money, error) {
	if s == nil || s.repository == nil {
		return "", domainmoney.Money{}, errors.New("tax rate service is not configured")
	}
	return s.calculateTaxMoney(s.repository, amountMoney, country, state, postalCodes...)
}

func (s *TaxRateService) CalculateTaxMoneyInTransaction(
	taxRateRepository *repository.TaxRateRepository,
	amountMoney domainmoney.Money,
	country, state string,
	postalCodes ...string,
) (string, domainmoney.Money, error) {
	if s == nil {
		return "", domainmoney.Money{}, errors.New("tax rate service is not configured")
	}
	if taxRateRepository == nil {
		return "", domainmoney.Money{}, errors.New("transactional tax rate repository is not configured")
	}
	return s.calculateTaxMoney(taxRateRepository, amountMoney, country, state, postalCodes...)
}

func (s *TaxRateService) calculateTaxMoney(
	taxRateRepository *repository.TaxRateRepository,
	amountMoney domainmoney.Money,
	country, state string,
	postalCodes ...string,
) (string, domainmoney.Money, error) {
	if err := amountMoney.Validate(); err != nil {
		return "", domainmoney.Money{}, fmt.Errorf("invalid tax amount: %w", err)
	}
	taxRate, err := taxRateRepository.FindTaxRateByLocation(country, state, postalCodes...)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			postalCode := ""
			if len(postalCodes) > 0 {
				postalCode = strings.ToUpper(strings.TrimSpace(postalCodes[0]))
			}
			return "", domainmoney.Money{}, fmt.Errorf(
				"%w: no configured tax rate for %s/%s/%s",
				ErrTaxRateUnavailable,
				strings.ToUpper(strings.TrimSpace(country)),
				strings.ToUpper(strings.TrimSpace(state)),
				postalCode,
			)
		}
		return "", domainmoney.Money{}, fmt.Errorf("failed to load tax rate for %s/%s: %w", country, state, err)
	}
	if taxRate == nil {
		return "", domainmoney.Money{}, fmt.Errorf("%w: tax rate lookup returned no result", ErrTaxRateUnavailable)
	}
	taxMoney, err := taxRate.CalculateTaxMoney(amountMoney)
	if err != nil {
		return "", domainmoney.Money{}, err
	}
	return taxRate.RateDecimal, taxMoney, nil
}

func normalizeTaxRateRuleInput(input TaxRateRuleInput) (TaxRateRuleInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return TaxRateRuleInput{}, fmt.Errorf("%w: name is required", ErrTaxRateRuleInvalid)
	}
	if utf8.RuneCountInString(input.Name) > 255 {
		return TaxRateRuleInput{}, fmt.Errorf("%w: name must not exceed 255 characters", ErrTaxRateRuleInvalid)
	}
	input.Country = strings.ToUpper(strings.TrimSpace(input.Country))
	if len(input.Country) != 2 || input.Country[0] < 'A' || input.Country[0] > 'Z' || input.Country[1] < 'A' || input.Country[1] > 'Z' {
		return TaxRateRuleInput{}, fmt.Errorf("%w: country must be a two-letter country code", ErrTaxRateRuleInvalid)
	}
	input.State = strings.ToUpper(strings.TrimSpace(input.State))
	input.PostalCode = strings.ToUpper(strings.TrimSpace(input.PostalCode))
	if len(input.State) > 100 {
		return TaxRateRuleInput{}, fmt.Errorf("%w: state must not exceed 100 characters", ErrTaxRateRuleInvalid)
	}
	if len(input.PostalCode) > 100 {
		return TaxRateRuleInput{}, fmt.Errorf("%w: postal code must not exceed 100 characters", ErrTaxRateRuleInvalid)
	}
	if input.Priority < -2147483648 || input.Priority > 2147483647 {
		return TaxRateRuleInput{}, fmt.Errorf("%w: priority must fit a 32-bit integer", ErrTaxRateRuleInvalid)
	}
	input.RateDecimal = strings.TrimSpace(input.RateDecimal)
	if err := taxrate.ValidateTaxRatePercentage(input.RateDecimal); err != nil {
		return TaxRateRuleInput{}, fmt.Errorf("%w: %v", ErrTaxRateRuleInvalid, err)
	}
	return input, nil
}
