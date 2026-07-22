package shipping

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifyTrackingWebhookAcceptsTrack17SignHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	payload := []byte(`{"event":"TRACKING_UPDATED","data":{"number":"TRACK123"}}`)
	secret := "track17-security-key"
	hash := sha256.New()
	hash.Write(payload)
	hash.Write([]byte("/"))
	hash.Write([]byte(secret))

	request, err := http.NewRequest(http.MethodPost, "/api/v1/shipping/webhook/17track", nil)
	require.NoError(t, err)
	request.Header.Set("sign", hex.EncodeToString(hash.Sum(nil)))
	context := &gin.Context{Request: request}

	assert.True(t, verifyTrackingWebhook(context, secret, payload))
}

func TestVerifyTrackingWebhookAcceptsInternalHMACSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)
	payload := []byte(`{"tracking_number":"TRACK123"}`)
	secret := "internal-webhook-secret"
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/shipping/webhook/mock", nil)
	require.NoError(t, err)
	request.Header.Set("X-Tanzanite-Signature", "sha256="+hex.EncodeToString(mac.Sum(nil)))
	context := &gin.Context{Request: request}

	assert.True(t, verifyTrackingWebhook(context, secret, payload))
}

func TestParseTrackingWebhookInputSupportsTrack17TrackingUpdatedPayload(t *testing.T) {
	payload := []byte(`{
		"event": "TRACKING_UPDATED",
		"data": {
			"number": "YT123456789CN",
			"carrier": "190271",
			"track_info": {
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
								"key": "190271",
								"name": "YunExpress"
							},
							"events": [
								{
									"time_iso": "2026-07-22T10:20:30Z",
									"description": "Delivered to recipient",
									"location": "Los Angeles, CA",
									"stage": "Delivered",
									"sub_status": "Delivered"
								},
								{
									"time_utc": "2026-07-21 08:10:00",
									"description": "Arrived at destination facility",
									"location": "Los Angeles, CA",
									"stage": "InTransit",
									"sub_status": "InTransit"
								}
							]
						}
					]
				}
			}
		}
	}`)

	input, err := parseTrackingWebhookInput(payload)

	require.NoError(t, err)
	assert.Equal(t, "YT123456789CN", input.TrackingNumber)
	assert.Equal(t, "190271", input.ProviderCarrierCode)
	assert.Equal(t, "Delivered", input.Status)
	require.Len(t, input.Events, 2)
	assert.Equal(t, "Delivered", input.Events[0].Status)
	assert.Equal(t, "Delivered to recipient", input.Events[0].Description)
	assert.Equal(t, "Los Angeles, CA", input.Events[0].Location)
	assert.Equal(t, time.Date(2026, 7, 22, 10, 20, 30, 0, time.UTC), input.Events[0].EventTime)
	assert.Equal(t, "InTransit", input.Events[1].Status)
	assert.Equal(t, time.Date(2026, 7, 21, 8, 10, 0, 0, time.UTC), input.Events[1].EventTime)
}

func TestParseTrackingWebhookInputFallsBackToTrack17LatestEvent(t *testing.T) {
	payload := []byte(`{
		"event": "TRACKING_UPDATED",
		"data": {
			"number": "YT123456789CN",
			"track_info": {
				"latest_status": {
					"status": "InTransit",
					"sub_status": "Departed"
				},
				"latest_event": {
					"time_iso": "2026-07-22T10:20:30Z",
					"description": "Departed from facility",
					"location": "Shenzhen, CN",
					"stage": "InTransit"
				},
				"tracking": {
					"providers": [
						{
							"provider": {
								"key": "190271",
								"name": "YunExpress"
							},
							"events": []
						}
					]
				}
			}
		}
	}`)

	input, err := parseTrackingWebhookInput(payload)

	require.NoError(t, err)
	assert.Equal(t, "YT123456789CN", input.TrackingNumber)
	assert.Equal(t, "190271", input.ProviderCarrierCode)
	assert.Equal(t, "Departed", input.Status)
	require.Len(t, input.Events, 1)
	assert.Equal(t, "Departed", input.Events[0].Status)
	assert.Equal(t, "Departed from facility", input.Events[0].Description)
	assert.Equal(t, "Shenzhen, CN", input.Events[0].Location)
}
