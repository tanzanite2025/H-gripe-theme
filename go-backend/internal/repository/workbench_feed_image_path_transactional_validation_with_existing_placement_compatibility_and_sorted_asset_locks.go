package repository

import "gorm.io/gorm"

// The caller must hold the active feed entry row lock before reading previousPaths,
// and keep this transaction open until its media rows have been written.
// Paths are matched as occurrences within that entry because the legacy input
// contract has no placement IDs. Existing occurrences can move, but cannot expand.
func ValidateAndCanonicalizeWorkbenchFeedImagePathsWithinTransaction(tx *gorm.DB, resolver PublicBusinessMediaReferenceObjectKeyResolver, paths, previousPaths []string) error {
	remaining := make(map[string]int, len(previousPaths))
	for _, path := range previousPaths {
		remaining[path]++
	}
	var bindings []publicBusinessMediaReferenceBinding
	var references []string
	for index, path := range paths {
		if remaining[path] > 0 {
			remaining[path]--
			continue // Preserve the exact historical address without inventing a new binding.
		}
		bindings = append(bindings, publicBusinessMediaReferenceBinding{URL: &paths[index], ExpectedMediaType: "image"})
		references = append(references, path)
	}
	resolved, err := lockAndResolvePublicBusinessMediaReferenceIdentities(tx, resolver, references, nil)
	if err != nil {
		return err
	}
	for _, binding := range bindings {
		if err := resolved.validateAndCanonicalizePublicBusinessMediaReferenceBinding(binding); err != nil {
			return err
		}
	}
	return nil
}
