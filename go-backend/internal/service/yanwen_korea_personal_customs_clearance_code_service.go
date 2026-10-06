package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// YanwenKoreaPersonalCustomsClearanceCodeService owns the Yanwen-only Korean
// PCCC preflight boundary. It does not write generic customs, order, tracking,
// or evidence records.
type YanwenKoreaPersonalCustomsClearanceCodeService struct {
	credentialResolver YanwenGatewayCredentialResolver
	gateway            *YanwenGatewayClient
}

type YanwenKoreaPersonalCustomsClearanceCodeVerificationInput struct {
	Environment   string `json:"environment"`
	RecipientName string `json:"recipient_name"`
	Phone         string `json:"phone"`
	TaxNumber     string `json:"tax_number"`
	PostalCode    string `json:"zip_code"`
}

type YanwenKoreaPersonalCustomsClearanceCodeVerificationResult struct {
	Environment    string `json:"environment"`
	OfficialPassed bool   `json:"official_passed"`
	Code           string `json:"code"`
	Message        string `json:"message"`
}

func NewYanwenKoreaPersonalCustomsClearanceCodeService(credentialResolver YanwenGatewayCredentialResolver) *YanwenKoreaPersonalCustomsClearanceCodeService {
	gateway := NewYanwenGatewayClient()
	if provider, ok := credentialResolver.(yanwenGatewayClientProvider); ok {
		if configuredGateway := provider.yanwenGatewayClient(); configuredGateway != nil {
			gateway = configuredGateway
		}
	}
	return &YanwenKoreaPersonalCustomsClearanceCodeService{credentialResolver: credentialResolver, gateway: gateway}
}

func (s *YanwenKoreaPersonalCustomsClearanceCodeService) VerifyKoreaPersonalCustomsClearanceCode(
	ctx context.Context,
	input YanwenKoreaPersonalCustomsClearanceCodeVerificationInput,
) (YanwenKoreaPersonalCustomsClearanceCodeVerificationResult, error) {
	if s == nil || s.credentialResolver == nil {
		return YanwenKoreaPersonalCustomsClearanceCodeVerificationResult{}, errors.New("Yanwen Korea PCCC service is not configured")
	}
	input.Environment = normalizeYanwenEnvironment(input.Environment)
	if !isSupportedYanwenEnvironment(input.Environment) {
		return YanwenKoreaPersonalCustomsClearanceCodeVerificationResult{}, errors.New("Yanwen environment must be fat or production")
	}
	input.RecipientName = strings.TrimSpace(input.RecipientName)
	input.Phone = strings.TrimSpace(input.Phone)
	input.TaxNumber = strings.TrimSpace(input.TaxNumber)
	input.PostalCode = strings.TrimSpace(input.PostalCode)
	if input.RecipientName == "" {
		return YanwenKoreaPersonalCustomsClearanceCodeVerificationResult{}, errors.New("Korea PCCC recipient name is required")
	}
	if input.Phone == "" {
		return YanwenKoreaPersonalCustomsClearanceCodeVerificationResult{}, errors.New("Korea PCCC recipient phone is required")
	}
	if input.TaxNumber == "" {
		return YanwenKoreaPersonalCustomsClearanceCodeVerificationResult{}, errors.New("Korea PCCC tax number is required")
	}
	if !isFiveDigitYanwenPostalCode(input.PostalCode) {
		return YanwenKoreaPersonalCustomsClearanceCodeVerificationResult{}, errors.New("Korea PCCC postal code must contain exactly 5 digits")
	}

	credentials, err := s.credentialResolver.resolveYanwenGatewayCredentials(YanwenAPIConfigInput{Environment: input.Environment})
	if err != nil {
		return YanwenKoreaPersonalCustomsClearanceCodeVerificationResult{}, fmt.Errorf("resolve Yanwen %s credentials: %w", input.Environment, err)
	}
	if s.gateway == nil {
		return YanwenKoreaPersonalCustomsClearanceCodeVerificationResult{}, errors.New("Yanwen Korea PCCC gateway is not configured")
	}
	official, err := s.gateway.VerifyYanwenKoreaPersonalCustomsClearanceCode(ctx, credentials, YanwenKoreaPersonalCustomsClearanceCodeVerificationRequest{
		ReceiverInfo: YanwenKoreaPersonalCustomsClearanceCodeReceiverInfo{
			Name:      input.RecipientName,
			Phone:     input.Phone,
			TaxNumber: input.TaxNumber,
			ZipCode:   input.PostalCode,
		},
	})
	if err != nil {
		return YanwenKoreaPersonalCustomsClearanceCodeVerificationResult{}, err
	}
	return YanwenKoreaPersonalCustomsClearanceCodeVerificationResult{
		Environment:    input.Environment,
		OfficialPassed: official.OfficialPassed,
		Code:           official.Code,
		Message:        official.Message,
	}, nil
}

func isFiveDigitYanwenPostalCode(value string) bool {
	if len(value) != 5 {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}
