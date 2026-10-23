package maintenancegate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRestoredIdentityBlocksWrongDatabaseStorageOrKeysButAllowsApplicationUpgrade(t *testing.T) {
	directory := t.TempDir()
	mediaDirectory := t.TempDir()
	otherMediaDirectory := t.TempDir()
	roots := map[string]string{"uploads": mediaDirectory}
	keys := map[string]string{"JWT_SECRET": "synthetic-original-identity-key"}
	require.NoError(t, RegisterApplicationBoundary(directory, "storefront", "build-before", roots, keys))
	boundaryBytes, err := os.ReadFile(filepath.Join(directory, "application-boundary.json"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(directory, "restored-data-identity-boundary.json"), boundaryBytes, 0600))
	require.NoError(t, RegisterApplicationBoundary(directory, "storefront", "build-after", roots, keys))
	require.Error(t, RegisterApplicationBoundary(directory, "different_store", "build-after", roots, keys))
	require.Error(t, RegisterApplicationBoundary(directory, "storefront", "build-after", map[string]string{"uploads": otherMediaDirectory}, keys))
	require.Error(t, RegisterApplicationBoundary(directory, "storefront", "build-after", roots, map[string]string{"JWT_SECRET": "synthetic-wrong-key"}))
	require.NoError(t, os.WriteFile(filepath.Join(directory, "restored-data-identity-boundary.json"), []byte(`{"protocol":99}`), 0600))
	require.Error(t, RegisterApplicationBoundary(directory, "storefront", "build-after", roots, keys))
}
