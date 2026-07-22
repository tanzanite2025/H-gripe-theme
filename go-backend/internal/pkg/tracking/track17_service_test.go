package tracking

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrack17BatchTrackUsesV24GetTrackInfoAndParsesEvents(t *testing.T) {
	var requestPath string
	var token string
	var payload []track17RequestItem
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath = r.URL.Path
		token = r.Header.Get("17token")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"code": 0,
			"data": {
				"accepted": [
					{
						"number": "YT123456789CN",
						"carrier": 190271,
						"track_info": {
							"shipping_info": {
								"shipper_address": {
									"country": "CN",
									"state": "GD",
									"city": "SHENZHEN",
									"postal_code": "518000"
								},
								"recipient_address": {
									"country": "US",
									"state": "CA",
									"city": "Los Angeles"
								}
							},
							"latest_status": {
								"status": "InfoReceived",
								"sub_status": "Delivered"
							},
							"latest_event": {
								"time_iso": "2026-07-22T10:20:30Z",
								"description": "Delivered to recipient",
								"location": "Los Angeles, CA",
								"stage": "Delivered",
								"sub_status": "Delivered"
							},
							"tracking": {
								"providers": [
									{
										"provider": {
											"key": 190271,
											"name": "YunExpress"
										},
										"events": [
											{
												"time_iso": "2026-07-22T10:20:30Z",
												"description": "Delivered to recipient",
												"location": "Los Angeles, CA",
												"stage": "Delivered",
												"sub_status": "Delivered"
											}
										]
									}
								]
							}
						}
					}
				],
				"rejected": []
			}
		}`))
	}))
	defer server.Close()

	service := NewTrackingService(&Config{
		APIKey:  "test-token",
		BaseURL: server.URL,
		Timeout: time.Second,
	})

	results, err := service.BatchTrack(t.Context(), []TrackingRequest{
		{TrackingNumber: "YT123456789CN", Carrier: "190271"},
	})

	require.NoError(t, err)
	assert.Equal(t, "/track/v2.4/gettrackinfo", requestPath)
	assert.Equal(t, "test-token", token)
	require.Len(t, payload, 1)
	assert.Equal(t, "YT123456789CN", payload[0].Number)
	assert.JSONEq(t, "190271", string(payload[0].Carrier))
	require.Len(t, results, 1)
	assert.Equal(t, "YT123456789CN", results[0].TrackingNumber)
	assert.Equal(t, "190271", results[0].Carrier)
	assert.Equal(t, "Delivered", results[0].Status)
	assert.Equal(t, 4, results[0].StatusCode)
	require.Len(t, results[0].Events, 1)
	assert.Equal(t, "Delivered to recipient", results[0].Events[0].Description)
	assert.Equal(t, "Los Angeles, CA", results[0].Events[0].Location.City)
	assert.Equal(t, time.Date(2026, 7, 22, 10, 20, 30, 0, time.UTC), results[0].Events[0].Time)
	require.NotNil(t, results[0].Origin)
	assert.Equal(t, "SHENZHEN", results[0].Origin.City)
	require.NotNil(t, results[0].Destination)
	assert.Equal(t, "Los Angeles", results[0].Destination.City)
}

func TestTrack17RegisterTrackingsUsesV24Register(t *testing.T) {
	var requestPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"code": 0,
			"data": {
				"accepted": [
					{
						"number": "YT123456789CN",
						"carrier": 190271
					}
				],
				"rejected": []
			}
		}`))
	}))
	defer server.Close()

	service := NewTrackingService(&Config{
		APIKey:  "test-token",
		BaseURL: server.URL,
		Timeout: time.Second,
	})

	err := service.(TrackingRegistrar).RegisterTrackings(t.Context(), []TrackingRequest{
		{TrackingNumber: "YT123456789CN", Carrier: "190271"},
	})

	require.NoError(t, err)
	assert.Equal(t, "/track/v2.4/register", requestPath)
}
