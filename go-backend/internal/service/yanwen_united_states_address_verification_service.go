package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// YanwenUnitedStatesAddressVerificationService owns the Yanwen-only US
// address preflight boundary. It does not write generic customs, order,
// tracking, or evidence records.
type YanwenUnitedStatesAddressVerificationService struct {
	credentialResolver YanwenGatewayCredentialResolver
	gateway            *YanwenGatewayClient
}

type YanwenUnitedStatesAddressVerificationInput struct {
	Environment string `json:"environment"`
	Address     string `json:"address"`
	ZipCode     string `json:"zip_code"`
	City        string `json:"city"`
	State       string `json:"state"`
}

type YanwenUnitedStatesAddressVerificationResult struct {
	Environment        string `json:"environment"`
	OfficialPassed     bool   `json:"official_passed"`
	Code               string `json:"code"`
	Message            string `json:"message"`
	NormalizedAddress  string `json:"normalized_address"`
	NormalizedCity     string `json:"normalized_city"`
	NormalizedState    string `json:"normalized_state"`
	NormalizedZipCode4 string `json:"normalized_zip_code4"`
	NormalizedZipCode5 string `json:"normalized_zip_code5"`
}

func NewYanwenUnitedStatesAddressVerificationService(credentialResolver YanwenGatewayCredentialResolver) *YanwenUnitedStatesAddressVerificationService {
	gateway := NewYanwenGatewayClient()
	if provider, ok := credentialResolver.(yanwenGatewayClientProvider); ok {
		if configuredGateway := provider.yanwenGatewayClient(); configuredGateway != nil {
			gateway = configuredGateway
		}
	}
	return &YanwenUnitedStatesAddressVerificationService{credentialResolver: credentialResolver, gateway: gateway}
}

func (s *YanwenUnitedStatesAddressVerificationService) VerifyUnitedStatesAddress(
	ctx context.Context,
	input YanwenUnitedStatesAddressVerificationInput,
) (YanwenUnitedStatesAddressVerificationResult, error) {
	if s == nil || s.credentialResolver == nil {
		return YanwenUnitedStatesAddressVerificationResult{}, errors.New("Yanwen US address verification service is not configured")
	}
	input.Environment = normalizeYanwenEnvironment(input.Environment)
	if !isSupportedYanwenEnvironment(input.Environment) {
		return YanwenUnitedStatesAddressVerificationResult{}, errors.New("Yanwen environment must be fat or production")
	}
	input.Address = strings.TrimSpace(input.Address)
	input.ZipCode = strings.TrimSpace(input.ZipCode)
	input.City = strings.TrimSpace(input.City)
	input.State = strings.TrimSpace(input.State)
	if input.Address == "" {
		return YanwenUnitedStatesAddressVerificationResult{}, errors.New("US address is required")
	}
	if input.ZipCode == "" {
		return YanwenUnitedStatesAddressVerificationResult{}, errors.New("US address postal code is required")
	}
	if input.City == "" {
		return YanwenUnitedStatesAddressVerificationResult{}, errors.New("US address city is required")
	}
	if input.State == "" {
		return YanwenUnitedStatesAddressVerificationResult{}, errors.New("US address state is required")
	}

	credentials, err := s.credentialResolver.resolveYanwenGatewayCredentials(YanwenAPIConfigInput{Environment: input.Environment})
	if err != nil {
		return YanwenUnitedStatesAddressVerificationResult{}, fmt.Errorf("resolve Yanwen %s credentials: %w", input.Environment, err)
	}
	if s.gateway == nil {
		return YanwenUnitedStatesAddressVerificationResult{}, errors.New("Yanwen US address verification gateway is not configured")
	}
	official, err := s.gateway.VerifyYanwenUnitedStatesAddress(ctx, credentials, YanwenUnitedStatesAddressVerificationRequest{
		ReceiverInfo: YanwenUnitedStatesAddressReceiverInfo{
			Address: input.Address,
			ZipCode: input.ZipCode,
			City:    input.City,
			State:   input.State,
		},
	})
	if err != nil {
		return YanwenUnitedStatesAddressVerificationResult{}, err
	}
	return YanwenUnitedStatesAddressVerificationResult{
		Environment:        input.Environment,
		OfficialPassed:     official.OfficialPassed,
		Code:               official.Code,
		Message:            official.Message,
		NormalizedAddress:  official.NormalizedAddress,
		NormalizedCity:     official.NormalizedCity,
		NormalizedState:    official.NormalizedState,
		NormalizedZipCode4: official.NormalizedZipCode4,
		NormalizedZipCode5: official.NormalizedZipCode5,
	}, nil
}
