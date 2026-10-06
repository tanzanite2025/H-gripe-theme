package service

import (
	"testing"

	"commerce-platform/internal/domain/shipping"

	"github.com/stretchr/testify/require"
)

func TestValidateYanwenChannelCustomsDeclarationRequirementsBlocksMissingConfiguredFields(t *testing.T) {
	tests := []struct {
		name        string
		channel     shipping.YanwenPublishedChannel
		declaration YanwenCustomsDeclarationInput
		expected    string
	}{
		{
			name:     "receiver tax number",
			channel:  shipping.YanwenPublishedChannel{ProductCode: "YW-TAX", RequireReceiverTaxNumber: true},
			expected: "receiver tax number",
		},
		{
			name:     "IOSS",
			channel:  shipping.YanwenPublishedChannel{ProductCode: "YW-IOSS", RequireIOSS: true},
			expected: "IOSS number",
		},
		{
			name:     "EORI",
			channel:  shipping.YanwenPublishedChannel{ProductCode: "YW-EORI", RequireEORI: true},
			expected: "EORI number",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateYanwenChannelCustomsDeclarationRequirements(test.channel, test.declaration)
			require.ErrorContains(t, err, test.expected)
		})
	}
}

func TestValidateYanwenChannelCustomsDeclarationRequirementsAllowsConfiguredValues(t *testing.T) {
	err := validateYanwenChannelCustomsDeclarationRequirements(
		shipping.YanwenPublishedChannel{
			ProductCode:              "YW-ALL",
			RequireReceiverTaxNumber: true,
			RequireIOSS:              true,
			RequireEORI:              true,
		},
		YanwenCustomsDeclarationInput{
			ReceiverTaxNumber: " TAX-123 ",
			IOSS:              " IOSS-123 ",
			EORI:              " EORI-123 ",
		},
	)
	require.NoError(t, err)
}

func TestValidateYanwenChannelCustomsDeclarationRequirementsKeepsUnconfiguredFieldsOptional(t *testing.T) {
	err := validateYanwenChannelCustomsDeclarationRequirements(
		shipping.YanwenPublishedChannel{ProductCode: "YW-OPTIONAL"},
		YanwenCustomsDeclarationInput{},
	)
	require.NoError(t, err)
}
