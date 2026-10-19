package orderevidence

// Submission requests use strings and integer counters/IDs. UseNumber keeps
// large IDs exact while stable key ordering and whitespace survive JSONB's
// representation changes. Existing byte-exact snapshot hashes remain valid.
func CanonicalSubmissionSnapshotJSONForStorageIndependentIntegrity(payload []byte) ([]byte, error) {
	return CanonicalImmutableSnapshotJSONForStorageIndependentIntegrity(payload)
}
