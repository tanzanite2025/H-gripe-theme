package service

import (
	"context"
	"encoding/json"
	"testing"

	"commerce-platform/internal/domain/taxrate"

	"github.com/stretchr/testify/require"
)

func TestTaxRateSourceConfigurationExplicitlyClearsPreviousErrorAfterSuccessfulSync(t *testing.T) {
	db, snapshotService := newTaxRateSourceSnapshotServiceTestFixture(t)
	_, err := snapshotService.UpdateSourceConfiguration(TaxRateSourceConfigurationUpdate{Enabled: true, RefreshIntervalHours: 720})
	require.NoError(t, err)
	require.NoError(t, db.Model(&taxrate.TaxRateSourceConfig{}).Where("id = ?", 1).Update("last_sync_error", "column country_rate_comments does not exist").Error)
	configuration, err := snapshotService.GetSourceConfiguration()
	require.NoError(t, err)
	require.NotEmpty(t, configuration.LastSyncError)
	snapshotService.ConfigureHTTPClient(newStaticTaxRateSourceSnapshotHTTPClient(validVATComplyTaxRatesResponseBody()))
	_, err = snapshotService.Sync(context.Background())
	require.NoError(t, err)
	configuration, err = snapshotService.GetSourceConfiguration()
	require.NoError(t, err)
	require.Empty(t, configuration.LastSyncError)
	encoded, err := json.Marshal(configuration)
	require.NoError(t, err)
	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(encoded, &response))
	require.Contains(t, response, "last_sync_error", "clients must receive an explicit reset after recovery")
	require.Equal(t, "", response["last_sync_error"])
}

func TestTaxRateSnapshotPreservesCountryRateCommentsAndVersionsAnnotationOnlyChanges(t *testing.T) {
	db, snapshotService := newTaxRateSourceSnapshotServiceTestFixture(t)
	_, err := snapshotService.UpdateSourceConfiguration(TaxRateSourceConfigurationUpdate{Enabled: true, RefreshIntervalHours: 720})
	require.NoError(t, err)

	var countries []map[string]interface{}
	require.NoError(t, json.Unmarshal(validVATComplyTaxRatesResponseBody(), &countries))
	for _, country := range countries {
		if country["country_code"] == "EL" {
			country["rate_comments"] = map[string][]string{
				"13.0":  {"For tariff heading 6506: bicycle helmets only.\nExcludes other headgear.", "Continuation of scope."},
				"8.500": {"Comment-only rate without a category or reduced-rate entry."},
			}
		}
	}
	responseBody, err := json.Marshal(countries)
	require.NoError(t, err)
	snapshotService.ConfigureHTTPClient(newStaticTaxRateSourceSnapshotHTTPClient(responseBody))
	firstSync, err := snapshotService.Sync(context.Background())
	require.NoError(t, err)
	current, err := snapshotService.GetCurrentSnapshot()
	require.NoError(t, err)
	require.Equal(t, []string{"For tariff heading 6506: bicycle helmets only.\nExcludes other headgear.", "Continuation of scope."}, current.Snapshot.CountryRateComments["GR"]["13"])
	require.Equal(t, []string{"Comment-only rate without a category or reduced-rate entry."}, current.Snapshot.CountryRateComments["GR"]["8.5"])
	require.NotContains(t, current.Snapshot.CountryRateComments, "EL")
	require.NotNil(t, current.Snapshot.CountryRateComments["DE"], "a collected country without notes must differ from an old snapshot")
	encoded, err := json.Marshal(current)
	require.NoError(t, err)
	var publicSnapshot struct {
		Snapshot struct {
			CountryRateComments taxrate.TaxRateSourceCountryRateComments `json:"country_rate_comments"`
		} `json:"snapshot"`
	}
	require.NoError(t, json.Unmarshal(encoded, &publicSnapshot))
	require.Equal(t, current.Snapshot.CountryRateComments, publicSnapshot.Snapshot.CountryRateComments)

	unchanged, err := snapshotService.Sync(context.Background())
	require.NoError(t, err)
	require.False(t, unchanged.Changed)
	for _, country := range countries {
		if country["country_code"] == "EL" {
			country["rate_comments"].(map[string][]string)["13.0"][0] = "Revised scope for bicycle helmets."
		}
	}
	responseBody, err = json.Marshal(countries)
	require.NoError(t, err)
	snapshotService.ConfigureHTTPClient(newStaticTaxRateSourceSnapshotHTTPClient(responseBody))
	updated, err := snapshotService.Sync(context.Background())
	require.NoError(t, err)
	require.True(t, updated.Changed)
	require.NotEqual(t, firstSync.Snapshot.ContentSHA256, updated.Snapshot.ContentSHA256)
	require.Equal(t, firstSync.Snapshot.RateCount, updated.Snapshot.RateCount)
	var historical taxrate.TaxRateSourceSnapshot
	require.NoError(t, db.First(&historical, firstSync.Snapshot.ID).Error)
	require.Equal(t, current.Snapshot.CountryRateComments, historical.CountryRateComments, "new annotations must not rewrite history")

	snapshotService.ConfigureHTTPClient(newStaticTaxRateSourceSnapshotHTTPClient(validVATComplyTaxRatesResponseBody()))
	removed, err := snapshotService.Sync(context.Background())
	require.NoError(t, err)
	require.True(t, removed.Changed, "removing annotations is also a source change")
	require.Empty(t, removed.Snapshot.CountryRateComments["GR"])
}

func TestTaxRateSnapshotInvalidAnnotationRatePreservesLastValidSnapshot(t *testing.T) {
	_, snapshotService := newTaxRateSourceSnapshotServiceTestFixture(t)
	_, err := snapshotService.UpdateSourceConfiguration(TaxRateSourceConfigurationUpdate{Enabled: true, RefreshIntervalHours: 720})
	require.NoError(t, err)
	snapshotService.ConfigureHTTPClient(newStaticTaxRateSourceSnapshotHTTPClient(validVATComplyTaxRatesResponseBody()))
	firstSync, err := snapshotService.Sync(context.Background())
	require.NoError(t, err)
	var countries []map[string]interface{}
	require.NoError(t, json.Unmarshal(validVATComplyTaxRatesResponseBody(), &countries))
	for _, invalidRate := range []string{"unknown", "101", "-1"} {
		countries[0]["rate_comments"] = map[string][]string{invalidRate: {"Invalid rate annotation"}}
		responseBody, err := json.Marshal(countries)
		require.NoError(t, err)
		snapshotService.ConfigureHTTPClient(newStaticTaxRateSourceSnapshotHTTPClient(responseBody))
		_, err = snapshotService.Sync(context.Background())
		require.ErrorContains(t, err, "rate comment")
		current, err := snapshotService.GetCurrentSnapshot()
		require.NoError(t, err)
		require.Equal(t, firstSync.Snapshot.ID, current.Snapshot.ID)
	}
}

func TestTaxRateSnapshotHistoricalMissingCommentsRemainUnknownUntilNewCollection(t *testing.T) {
	db, snapshotService := newTaxRateSourceSnapshotServiceTestFixture(t)
	_, err := snapshotService.UpdateSourceConfiguration(TaxRateSourceConfigurationUpdate{Enabled: true, RefreshIntervalHours: 720})
	require.NoError(t, err)
	snapshotService.ConfigureHTTPClient(newStaticTaxRateSourceSnapshotHTTPClient(validVATComplyTaxRatesResponseBody()))
	firstSync, err := snapshotService.Sync(context.Background())
	require.NoError(t, err)
	// Model an existing pre-migration record: SQL NULL and its old hash format.
	require.NoError(t, db.Model(&taxrate.TaxRateSourceSnapshot{}).Where("id = ?", firstSync.Snapshot.ID).Updates(map[string]interface{}{
		"country_rate_comments": nil, "content_sha256": "historical-hash-before-comment-collection",
	}).Error)
	current, err := snapshotService.GetCurrentSnapshot()
	require.NoError(t, err)
	require.Nil(t, current.Snapshot.CountryRateComments)
	encoded, err := json.Marshal(current)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"country_rate_comments":null`)
	collected, err := snapshotService.Sync(context.Background())
	require.NoError(t, err)
	require.True(t, collected.Changed)
	require.Len(t, collected.Snapshot.CountryRateComments, 27)
	var historical taxrate.TaxRateSourceSnapshot
	require.NoError(t, db.First(&historical, firstSync.Snapshot.ID).Error)
	require.Nil(t, historical.CountryRateComments)
}
