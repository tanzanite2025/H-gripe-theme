package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/emailtoken"

	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

const emailChallengeDeliveryTestSecret = "test-email-secret"

type failingEmailChallengeSender struct{}

func (failingEmailChallengeSender) SendEmail([]string, string, string) error {
	return errors.New("smtp unavailable")
}

type recordingEmailDeliveryResult struct {
	results []bool
}

func (r *recordingEmailDeliveryResult) RecordDeliveryResult(_ string, success bool) {
	r.results = append(r.results, success)
}

func TestEmailChallengeDeliveryOutboxHandlerSendsValidEvent(t *testing.T) {
	sender := &recordingEmailSender{}
	resultRecorder := &recordingEmailDeliveryResult{}
	event := newEmailChallengeDeliveryTestEvent(t, time.Now().UTC().Add(time.Hour))

	require.NoError(t, NewEmailChallengeDeliveryOutboxHandler(sender, resultRecorder, emailChallengeDeliveryTestSecret).Handle(context.Background(), event))
	require.Equal(t, []string{"Open https://example.test/confirm/" + rebuildEmailChallengeTestToken(t, event)}, sender.bodies)
	require.Equal(t, []bool{true}, resultRecorder.results)
}

func TestEmailChallengeDeliveryOutboxHandlerReturnsSenderErrorForRetry(t *testing.T) {
	event := newEmailChallengeDeliveryTestEvent(t, time.Now().UTC().Add(time.Hour))
	resultRecorder := &recordingEmailDeliveryResult{}
	err := NewEmailChallengeDeliveryOutboxHandler(failingEmailChallengeSender{}, resultRecorder, emailChallengeDeliveryTestSecret).Handle(context.Background(), event)
	require.EqualError(t, err, "smtp unavailable")
	require.Equal(t, []bool{false}, resultRecorder.results)
}

func TestEmailChallengeDeliveryOutboxHandlerRejectsInvalidPayload(t *testing.T) {
	err := NewEmailChallengeDeliveryOutboxHandler(&recordingEmailSender{}, nil, emailChallengeDeliveryTestSecret).Handle(context.Background(), outbox.Event{
		EventType: outbox.EventTypeEmailChallengeDelivery,
		Payload:   datatypes.JSON([]byte(`{"recipient_email":"not-an-email"}`)),
	})
	require.Error(t, err)
}

func TestEmailChallengeDeliveryOutboxHandlerSkipsExpiredToken(t *testing.T) {
	sender := &recordingEmailSender{}
	event := newEmailChallengeDeliveryTestEvent(t, time.Now().UTC().Add(-time.Minute))

	require.NoError(t, NewEmailChallengeDeliveryOutboxHandler(sender, nil, emailChallengeDeliveryTestSecret).Handle(context.Background(), event))
	require.Empty(t, sender.bodies)
}

func newEmailChallengeDeliveryTestEvent(t *testing.T, expiresAt time.Time) outbox.Event {
	t.Helper()
	claims := emailtoken.Claims{
		Purpose:   "subscription:confirm",
		Email:     "rider@example.test",
		Subject:   "rider@example.test",
		Nonce:     "fixed-test-nonce",
		ExpiresAt: expiresAt.Unix(),
	}
	token, err := emailtoken.Sign(emailChallengeDeliveryTestSecret, claims)
	require.NoError(t, err)
	payload, err := json.Marshal(outbox.EmailChallengeDeliveryPayload{
		RecipientEmail:   claims.Email,
		DeliverySubject:  "Confirm subscription",
		BodyTemplate:     "Open https://example.test/confirm/" + emailChallengeTokenPlaceholder,
		Purpose:          claims.Purpose,
		ChallengeSubject: claims.Subject,
		Nonce:            claims.Nonce,
		ExpiresAt:        expiresAt,
		RequestedAt:      time.Now().UTC(),
	})
	require.NoError(t, err)
	return outbox.Event{
		EventType:   outbox.EventTypeEmailChallengeDelivery,
		AggregateID: emailtoken.Hash(token),
		Payload:     datatypes.JSON(payload),
	}
}

func rebuildEmailChallengeTestToken(t *testing.T, event outbox.Event) string {
	t.Helper()
	var payload outbox.EmailChallengeDeliveryPayload
	require.NoError(t, json.Unmarshal(event.Payload, &payload))
	token, err := emailtoken.Sign(emailChallengeDeliveryTestSecret, emailtoken.Claims{
		Purpose:   payload.Purpose,
		Email:     payload.RecipientEmail,
		Subject:   payload.ChallengeSubject,
		Nonce:     payload.Nonce,
		ExpiresAt: payload.ExpiresAt.Unix(),
	})
	require.NoError(t, err)
	return token
}
