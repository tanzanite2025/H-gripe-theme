package orderevidence

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubmissionSnapshotIntegritySurvivesJSONBWhitespaceAndKeyOrderWithoutLosingLargeIntegerIdentity(t *testing.T) {
	original := []byte(`{"z":9007199254740993,"a":{"url":"https://cdn.example.test/one.pdf","count":2}}`)
	canonical, err := CanonicalSubmissionSnapshotJSONForStorageIndependentIntegrity(original)
	require.NoError(t, err)
	require.Contains(t, string(canonical), "9007199254740993")
	snapshot := OrderEvidenceSubmissionSnapshot{SnapshotData: []byte(`{"a": {"count": 2, "url": "https://cdn.example.test/one.pdf"}, "z": 9007199254740993}`), SnapshotSHA256: fmt.Sprintf("%x", sha256.Sum256(canonical))}
	require.NoError(t, snapshot.VerifyIntegrity())
	snapshot.SnapshotData = []byte(`{"a": {"count": 2, "url": "https://cdn.example.test/two.pdf"}, "z": 9007199254740993}`)
	require.Error(t, snapshot.VerifyIntegrity())
	snapshot.SnapshotData = original
	snapshot.SnapshotSHA256 = fmt.Sprintf("%x", sha256.Sum256(original))
	require.NoError(t, snapshot.VerifyIntegrity())
}
