package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"commerce-platform/internal/domain/taxrate"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestTaxRateSourceSnapshotSyncStoresVersionedReferenceDataWithoutChangingCheckoutRates(t *testing.T) {
	db, snapshotService := newTaxRateSourceSnapshotServiceTestFixture(t)
	checkoutRate := taxrate.TaxRate{
		Name:        "Existing Germany VAT",
		Country:     "DE",
		RateDecimal: "19",
		Enabled:     true,
	}
	require.NoError(t, db.Create(&checkoutRate).Error)
	_, err := snapshotService.UpdateSourceConfiguration(TaxRateSourceConfigurationUpdate{
		Enabled:              true,
		RefreshIntervalHours: 720,
	})
	require.NoError(t, err)

	responseBody := validVATComplyTaxRatesResponseBody()
	snapshotService.ConfigureHTTPClient(&http.Client{Transport: taxRateSourceSnapshotRoundTripper(func(request *http.Request) (*http.Response, error) {
		require.Equal(t, vatcomplyTaxRateAPIEndpoint, request.URL.String())
		return newTaxRateSourceSnapshotHTTPResponse(request, http.StatusOK, responseBody), nil
	})})

	firstSync, err := snapshotService.Sync(context.Background())
	require.NoError(t, err)
	require.True(t, firstSync.Changed)
	require.Equal(t, len(vatcomplyExpectedEUMemberStateCountryCodes), firstSync.Snapshot.CountryCount)
	require.Contains(t, firstSync.Snapshot.Version, "vatcomply-")
	require.Len(t, firstSync.Snapshot.ContentSHA256, 64)

	currentSnapshot, err := snapshotService.GetCurrentSnapshot()
	require.NoError(t, err)
	require.Equal(t, firstSync.Snapshot.ID, currentSnapshot.Snapshot.ID)
	var greekEntry taxrate.TaxRateSourceSnapshotEntry
	require.NoError(t, db.Where("snapshot_id = ? AND country_code = ? AND rate_type = ?", firstSync.Snapshot.ID, "GR", "standard").First(&greekEntry).Error)
	require.Equal(t, "EL", greekEntry.SourceCountryCode)
	require.Equal(t, "24", greekEntry.RateDecimal)
	var bicyclePartsCategoryEntry taxrate.TaxRateSourceSnapshotEntry
	require.NoError(t, db.Where("snapshot_id = ? AND country_code = ? AND rate_type = ? AND rate_category = ?", firstSync.Snapshot.ID, "DE", "product_category", "bicycle_parts").First(&bicyclePartsCategoryEntry).Error)
	require.Equal(t, "7", bicyclePartsCategoryEntry.RateDecimal)

	secondSync, err := snapshotService.Sync(context.Background())
	require.NoError(t, err)
	require.False(t, secondSync.Changed, "an identical normalized source response must not create another version")
	require.Equal(t, firstSync.Snapshot.ID, secondSync.Snapshot.ID)
	var snapshotCount int64
	require.NoError(t, db.Model(&taxrate.TaxRateSourceSnapshot{}).Count(&snapshotCount).Error)
	require.EqualValues(t, 1, snapshotCount)

	var unchangedCheckoutRate taxrate.TaxRate
	require.NoError(t, db.First(&unchangedCheckoutRate, checkoutRate.ID).Error)
	require.Equal(t, "19", unchangedCheckoutRate.RateDecimal)
	require.True(t, unchangedCheckoutRate.Enabled)
}

func TestTaxRateSourceSnapshotSyncRejectsIncompleteSourceResponseAndKeepsPreviousSnapshot(t *testing.T) {
	db, snapshotService := newTaxRateSourceSnapshotServiceTestFixture(t)
	_, err := snapshotService.UpdateSourceConfiguration(TaxRateSourceConfigurationUpdate{
		Enabled:              true,
		RefreshIntervalHours: 720,
	})
	require.NoError(t, err)
	responseBody := validVATComplyTaxRatesResponseBody()
	snapshotService.ConfigureHTTPClient(newStaticTaxRateSourceSnapshotHTTPClient(responseBody))
	firstSync, err := snapshotService.Sync(context.Background())
	require.NoError(t, err)

	snapshotService.ConfigureHTTPClient(newStaticTaxRateSourceSnapshotHTTPClient([]byte(`[{"country_code":"DE","country_name":"Germany","standard_rate":19,"member_state":true}]`)))
	_, err = snapshotService.Sync(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "member-state coverage mismatch")

	currentSnapshot, err := snapshotService.GetCurrentSnapshot()
	require.NoError(t, err)
	require.Equal(t, firstSync.Snapshot.ID, currentSnapshot.Snapshot.ID)
	configuration, err := snapshotService.GetSourceConfiguration()
	require.NoError(t, err)
	require.NotEmpty(t, configuration.LastSyncError)
	var snapshotCount int64
	require.NoError(t, db.Model(&taxrate.TaxRateSourceSnapshot{}).Count(&snapshotCount).Error)
	require.EqualValues(t, 1, snapshotCount, "an invalid response must not replace or partially publish a snapshot")
}

func TestTaxRateSourceSnapshotSyncRequiresEnabledConfigurationAndValidRefreshInterval(t *testing.T) {
	_, snapshotService := newTaxRateSourceSnapshotServiceTestFixture(t)
	_, err := snapshotService.Sync(context.Background())
	require.ErrorIs(t, err, ErrTaxRateSourceDisabled)

	_, err = snapshotService.UpdateSourceConfiguration(TaxRateSourceConfigurationUpdate{
		Enabled:              true,
		RefreshIntervalHours: 23,
	})
	require.ErrorIs(t, err, ErrTaxRateSourceConfigInvalid)
}

func newTaxRateSourceSnapshotServiceTestFixture(t *testing.T) (*gorm.DB, *TaxRateSourceSnapshotService) {
	t.Helper()
	database, err := gorm.Open(sqlite.Open("file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, database.AutoMigrate(
		&taxrate.TaxRate{},
		&taxrate.TaxRateSourceConfig{},
		&taxrate.TaxRateSourceSnapshot{},
		&taxrate.TaxRateSourceSnapshotEntry{},
	))
	snapshotService := NewTaxRateSourceSnapshotService(repository.NewTaxRateSourceSnapshotRepository(database))
	return database, snapshotService
}

func validVATComplyTaxRatesResponseBody() []byte {
	countryCodes := []string{"AT", "BE", "BG", "CY", "CZ", "DE", "DK", "EE", "EL", "ES", "FI", "FR", "HR", "HU", "IE", "IT", "LT", "LU", "LV", "MT", "NL", "PL", "PT", "RO", "SE", "SI", "SK"}
	responseRows := make([]string, 0, len(countryCodes))
	for _, countryCode := range countryCodes {
		countryName := countryCode + " test country"
		standardRate := "20"
		if countryCode == "EL" {
			countryName = "Greece"
			standardRate = "24"
		}
		productCategoryRates := `"rate_categories":{}`
		if countryCode == "DE" {
			productCategoryRates = `"rate_categories":{"bicycle_parts":[7]}`
		}
		responseRows = append(responseRows, `{"country_code":"`+countryCode+`","country_name":"`+countryName+`","standard_rate":`+standardRate+`,"reduced_rates":[10,13],"super_reduced_rate":null,"parking_rate":null,"currency":"EUR","member_state":true,`+productCategoryRates+`}`)
	}
	return []byte("[" + strings.Join(responseRows, ",") + "]")
}

type taxRateSourceSnapshotRoundTripper func(*http.Request) (*http.Response, error)

func (roundTripper taxRateSourceSnapshotRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return roundTripper(request)
}

func newStaticTaxRateSourceSnapshotHTTPClient(body []byte) *http.Client {
	return &http.Client{Transport: taxRateSourceSnapshotRoundTripper(func(request *http.Request) (*http.Response, error) {
		return newTaxRateSourceSnapshotHTTPResponse(request, http.StatusOK, body), nil
	})}
}

func newTaxRateSourceSnapshotHTTPResponse(request *http.Request, statusCode int, body []byte) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(string(body))),
		Request:    request,
	}
}
