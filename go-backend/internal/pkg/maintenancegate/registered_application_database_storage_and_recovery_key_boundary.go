package maintenancegate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
)

type ApplicationBoundary struct {
	Protocol             int                         `json:"protocol"`
	Database             string                      `json:"database"`
	StorageRoots         map[string]string           `json:"storage_roots"`
	ApplicationVersion   string                      `json:"application_version"`
	RecoverySecretSHA256 map[string]string           `json:"recovery_secret_sha256"`
	ExternalProjections  *ExternalProjectionBoundary `json:"external_projections,omitempty"`
}

// Protocol 3 prevents older protocol-2 binaries, including ones that ignore the
// optional product configuration, from restarting with the old unsafe flusher.
func (boundary ApplicationBoundary) ValidateRegisteredRecoveryProjectionProtocol() error {
	if boundary.Protocol == 1 && boundary.ExternalProjections == nil {
		return nil
	}
	if (boundary.Protocol != 2 && boundary.Protocol != 3) || boundary.ExternalProjections == nil || boundary.ExternalProjections.Validate() != nil {
		return errors.New("registered recovery projection protocol is invalid")
	}
	expected := 2
	if boundary.ExternalProjections.ProductCatalog != nil && boundary.ExternalProjections.ProductCatalog.Version == 2 {
		expected = 3
	}
	if boundary.Protocol != expected {
		return errors.New("registered product flush contract requires its matching recovery protocol")
	}
	return nil
}

func RegisterApplicationBoundary(directory, database, version string, storageRoots, recoverySecrets map[string]string, externalProjections ...ExternalProjectionBoundary) error {
	if directory == "" {
		return nil
	}
	boundary := ApplicationBoundary{Protocol: 1, Database: database, ApplicationVersion: version, StorageRoots: storageRoots, RecoverySecretSHA256: map[string]string{}}
	if len(externalProjections) > 1 {
		return errors.New("exactly one external projection boundary may be registered")
	}
	if len(externalProjections) == 1 {
		if err := externalProjections[0].Validate(); err != nil {
			return err
		}
		boundary.ExternalProjections = &externalProjections[0]
		// Protocol 2 makes older business binaries reject the restored guard
		// instead of silently ignoring its new external-projection fields.
		boundary.Protocol = 2
		if boundary.ExternalProjections.ProductCatalog != nil && boundary.ExternalProjections.ProductCatalog.Version == 2 {
			boundary.Protocol = 3
		}
	}
	for name, value := range storageRoots {
		absolute, err := filepath.Abs(value)
		if err != nil {
			return err
		}
		resolved, err := filepath.EvalSymlinks(absolute)
		if err != nil {
			return err
		}
		boundary.StorageRoots[name] = resolved
	}
	for name, value := range recoverySecrets {
		if value == "" {
			return errors.New("backup boundary requires explicit recovery signing and encryption keys")
		}
		digest := sha256.Sum256([]byte(value))
		boundary.RecoverySecretSHA256[name] = hex.EncodeToString(digest[:])
	}
	if guardBytes, err := os.ReadFile(filepath.Join(directory, "restored-data-identity-boundary.json")); err == nil {
		var guard ApplicationBoundary
		if json.Unmarshal(guardBytes, &guard) != nil || guard.ValidateRegisteredRecoveryProjectionProtocol() != nil || (guard.Protocol >= 2 && boundary.Protocol != guard.Protocol) || guard.Database != boundary.Database || !reflect.DeepEqual(guard.StorageRoots, boundary.StorageRoots) || !reflect.DeepEqual(guard.RecoverySecretSHA256, boundary.RecoverySecretSHA256) {
			return errors.New("restored data requires the recorded database, storage mapping and recovery signing/encryption keys before application startup")
		}
		if guard.ExternalProjections != nil && !reflect.DeepEqual(guard.ExternalProjections, boundary.ExternalProjections) {
			return errors.New("restored data requires the reconciled external search, Redis, realtime and media delivery configuration before application startup")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	data, err := json.Marshal(boundary)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(directory, "application-boundary-*.next")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if err := file.Chmod(0600); err != nil {
		file.Close()
		return err
	}
	_, writeErr := file.Write(data)
	syncErr := file.Sync()
	closeErr := file.Close()
	if err := errors.Join(writeErr, syncErr, closeErr); err != nil {
		return err
	}
	return os.Rename(file.Name(), filepath.Join(directory, "application-boundary.json"))
}
