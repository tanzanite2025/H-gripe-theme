package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestYanwenGatewayQueriesOfficialTrackingEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, http.MethodGet, request.Method)
		require.Equal(t, "TRACKING-USER", request.Header.Get("Authorization"))
		require.Equal(t, "UH123,003456", request.URL.Query().Get("nums"))
		body, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		require.Empty(t, body)
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{
      "code": 0,
      "message": "success",
      "result": [{
        "tracking_number": "UH123",
        "waybill_number": "UH123",
        "exchange_number": "003456",
        "last_mile_carrier": "DHL",
        "last_mile_carrier_website": "https://www.dhl.com/",
        "last_mile_carrier_contact_number": "123456",
        "checkpoints": [{
          "time_stamp": "2023-04-10T09:32:00",
          "time_zone": "+08",
          "tracking_status": "LH20",
          "message": "Port of departure - Departure",
          "location": "Shanghai",
          "is_last_mile_checkpoint": 0,
          "extraProperties": {"FlightNumber": "CK205"}
        }],
        "tracking_status": "LH20",
        "tracking_status_waybill": {"level1": "2", "level2": "3", "level3": "LH20"},
        "last_mile_tracking_expected": true,
        "origin_country": "CN",
        "destination_country": "DE"
      }]
    }`))
	}))
	defer server.Close()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := NewYanwenGatewayClient()
	client.httpClient = &http.Client{Transport: yanwenTestRoundTripper{target: target}}
	results, err := client.QueryYanwenTracking(context.Background(), "TRACKING-USER", []string{"UH123", "UH123", "003456"})
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "UH123", results[0].TrackingNumber)
	require.Equal(t, "LH20", results[0].TrackingStatus)
	require.Equal(t, "IN_TRANSIT_AIR", string(results[0].TrackingStage))
	require.Equal(t, 2, results[0].TrackingStageRank)
	require.False(t, results[0].TrackingHasException)
	require.Len(t, results[0].Checkpoints, 1)
	require.True(t, results[0].Checkpoints[0].ExtraProperties["FlightNumber"] == "CK205")
	require.True(t, results[0].Checkpoints[0].IsLastMileCheckpoint == false)
}

func TestYanwenTrackingParserRejectsIncompleteOfficialResponse(t *testing.T) {
	_, err := parseYanwenTrackingResponse([]byte(`{"code":0,"message":"success","result":[{"tracking_number":"UH123"}]}`))
	require.ErrorContains(t, err, "waybill_number is required")

	_, err = parseYanwenTrackingResponse([]byte(`{"code":1,"message":"invalid number","result":[]}`))
	require.ErrorContains(t, err, "invalid number")

	_, err = parseYanwenTrackingResponse([]byte(`{"code":0,"message":"success","result":null}`))
	require.ErrorContains(t, err, "missing result")
}

func TestYanwenGatewayTrackingRejectsUnrequestedRemoteIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{"code":0,"message":"success","result":[{"tracking_number":"OTHER","waybill_number":"OTHER","checkpoints":[{"time_stamp":"2023-04-10T09:32:00","time_zone":"+08","tracking_status":"OR10","message":"accepted","is_last_mile_checkpoint":0}],"tracking_status":"OR10","tracking_status_waybill":{"level1":"1","level2":"1","level3":"OR10"}}]}`))
	}))
	defer server.Close()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)
	client := NewYanwenGatewayClient()
	client.httpClient = &http.Client{Transport: yanwenTestRoundTripper{target: target}}
	_, err = client.QueryYanwenTracking(context.Background(), "TRACKING-USER", []string{"UH123"})
	require.ErrorContains(t, err, "unrequested tracking number")
}

func TestYanwenGatewayTrackingRejectsMoreThanThirtyNumbers(t *testing.T) {
	numbers := make([]string, yanwenTrackingMaximumIdentifiers+1)
	for index := range numbers {
		numbers[index] = "TRACK-" + string(rune('A'+index))
	}
	client := NewYanwenGatewayClient()
	_, err := client.QueryYanwenTracking(context.Background(), "TRACKING-USER", numbers)
	require.ErrorContains(t, err, "at most 30")
}
