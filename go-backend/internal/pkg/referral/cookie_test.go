package referral

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignerRoundTripAndExpiry(t *testing.T) {
	now := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	signer, err := NewSigner("test-referral-cookie-secret")
	require.NoError(t, err)
	signer.now = func() time.Time { return now }

	token, encodedClaims, err := signer.Encode(" race2345 ", "link", 30*24*time.Hour)
	require.NoError(t, err)
	assert.Equal(t, "RACE2345", encodedClaims.Code)

	claims, err := signer.Decode(token)
	require.NoError(t, err)
	assert.Equal(t, encodedClaims, claims)

	signer.now = func() time.Time { return now.Add(31 * 24 * time.Hour) }
	_, err = signer.Decode(token)
	assert.ErrorIs(t, err, ErrInvalidCookie)
}

func TestSignerRejectsTampering(t *testing.T) {
	signer, err := NewSigner("test-referral-cookie-secret")
	require.NoError(t, err)
	token, _, err := signer.Encode("RACE2345", "link", time.Hour)
	require.NoError(t, err)

	_, err = signer.Decode(token + "x")
	assert.ErrorIs(t, err, ErrInvalidCookie)
}
