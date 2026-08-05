package attribution

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSignerRoundTripsNormalizedAttributionContext(t *testing.T) {
	signer, err := NewSigner("attribution-test-secret")
	require.NoError(t, err)
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	signer.now = func() time.Time { return now }

	token, err := signer.Encode(Context{
		Source:      " newsletter ",
		Campaign:    strings.Repeat("c", 200),
		ClickIDKind: "gclid",
		ClickID:     " click-123 ",
	})
	require.NoError(t, err)

	value, err := signer.Decode(token)
	require.NoError(t, err)
	require.Equal(t, "newsletter", value.Source)
	require.Len(t, value.Campaign, 160)
	require.Equal(t, "gclid", value.ClickIDKind)
	require.Equal(t, "click-123", value.ClickID)
	require.Equal(t, now, value.CapturedAt)
	require.Equal(t, now.Add(CookieMaxAge*time.Second), value.ExpiresAt)
}

func TestSignerRejectsTamperedOrExpiredContext(t *testing.T) {
	signer, err := NewSigner("attribution-test-secret")
	require.NoError(t, err)
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	signer.now = func() time.Time { return now }

	token, err := signer.Encode(Context{Source: "search"})
	require.NoError(t, err)

	_, err = signer.Decode(token + "x")
	require.ErrorIs(t, err, ErrInvalidContext)

	signer.now = func() time.Time { return now.Add(CookieMaxAge*time.Second + time.Second) }
	_, err = signer.Decode(token)
	require.ErrorIs(t, err, ErrInvalidContext)
}

func TestFromMetadataDropsUnrelatedFieldsAndDerivesClickSource(t *testing.T) {
	value, ok := FromMetadata(map[string]any{
		"gclid":        "abc123",
		"unexpected":   "discarded",
		"utm_campaign": "summer-sale",
	})

	require.True(t, ok)
	require.Equal(t, "google", value.Source)
	require.Equal(t, "summer-sale", value.Campaign)
	require.Equal(t, "gclid", value.ClickIDKind)
	require.Equal(t, "abc123", value.ClickID)
	require.NotContains(t, value.Metadata(), "unexpected")
}
