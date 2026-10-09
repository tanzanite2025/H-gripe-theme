package service

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestVATComplyTaxRateSourceAdapterRejectsRedirectsOutsideConfiguredHost(t *testing.T) {
	transportCalls := 0
	adapter := newVATComplyTaxRateSourceAdapter()
	adapter.httpClient.Transport = taxRateSourceSnapshotRoundTripper(func(request *http.Request) (*http.Response, error) {
		transportCalls++
		response := newTaxRateSourceSnapshotHTTPResponse(request, http.StatusFound, nil)
		response.Header.Set("Location", "http://127.0.0.1/internal-metadata")
		return response, nil
	})

	_, _, err := adapter.FetchVATComplyTaxRateEntries(context.Background(), time.Now().UTC())
	require.Error(t, err)
	require.Contains(t, strings.ToLower(err.Error()), "redirected outside")
	require.Equal(t, 1, transportCalls, "the adapter must not send a request to the redirected host")
}

func TestVATComplyMemberStateCoverageRejectsMissingAndUnexpectedCountries(t *testing.T) {
	actualCountryCodes := make(map[string]struct{}, len(vatcomplyExpectedEUMemberStateCountryCodes))
	for countryCode := range vatcomplyExpectedEUMemberStateCountryCodes {
		actualCountryCodes[countryCode] = struct{}{}
	}
	delete(actualCountryCodes, "DE")
	actualCountryCodes["XX"] = struct{}{}

	err := validateVATComplyMemberStateCoverage(actualCountryCodes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing: DE")
	require.Contains(t, err.Error(), "unexpected: XX")
}
