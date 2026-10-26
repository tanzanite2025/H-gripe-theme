// Package localbackup implements the application's independent, offline recovery
// protocol. It deliberately has no dependency on application login or job tables.
package localbackup

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// Version 2 binds sequence state into database fingerprints and records the
// source database size required for isolated restore capacity checks.
const PackageVersion = 2
const ChunkBytes int64 = 8 << 20

var safeDatabaseName = regexp.MustCompile(`^[a-z][a-z0-9_]{0,39}$`)
var safeTaskID = regexp.MustCompile(`^[a-f0-9]{32}$`)
var mediaRoots = []string{"uploads", "uploads_private", "site_logo_uploads"}

type Configuration struct {
	SiteID                               string                                             `json:"site_id"`
	Environment                          string                                             `json:"environment"`
	ApplicationVersion                   string                                             `json:"application_version"`
	ExpectedMigrationVersion             int64                                              `json:"expected_migration_version"`
	ApplicationDatabaseRole              string                                             `json:"application_database_role"`
	DatabaseURL                          string                                             `json:"database_url"`
	DatabaseName                         string                                             `json:"database_name"`
	Workspace                            string                                             `json:"workspace"`
	MaintenanceDirectory                 string                                             `json:"maintenance_directory"`
	PostgreSQLCapacityDirectory          string                                             `json:"postgresql_capacity_directory"`
	StorageRoots                         map[string]string                                  `json:"storage_roots"`
	RecoverySecrets                      map[string]string                                  `json:"recovery_secrets"`
	PublicOrigins                        []string                                           `json:"public_origins"`
	ProductStaticAssetsDirectory         string                                             `json:"product_static_assets_directory,omitempty"`
	PassphraseFile                       string                                             `json:"passphrase_file"`
	AuthenticationKeyFile                string                                             `json:"authentication_key_file"`
	MaxPackageBytes                      int64                                              `json:"max_package_bytes"`
	MaxExpandedBytes                     int64                                              `json:"max_expanded_bytes"`
	MaxWorkspaceBytes                    int64                                              `json:"max_workspace_bytes"`
	MaxFiles                             int                                                `json:"max_files"`
	WorkCopyLifetimeHours                int                                                `json:"work_copy_lifetime_hours"`
	ProductViewRecoveryRedis             *ProductViewRecoveryRedisConnectionConfiguration   `json:"product_view_recovery_redis,omitempty"`
	ProductViewRecoveryWriterInventory   []string                                           `json:"product_view_recovery_writer_inventory,omitempty"`
	StoppedProductLocalMediaVerification *StoppedProductLocalMediaVerificationConfiguration `json:"stopped_product_local_media_verification,omitempty"`
}

type FileIdentity struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

type Manifest struct {
	FormatVersion                         int              `json:"format_version"`
	PackageID                             string           `json:"package_id"`
	SiteID                                string           `json:"site_id"`
	Environment                           string           `json:"environment"`
	ApplicationVersion                    string           `json:"application_version"`
	MigrationVersion                      int64            `json:"migration_version"`
	SnapshotID                            string           `json:"snapshot_id"`
	DatabaseFingerprint                   string           `json:"database_fingerprint"`
	DatabaseSizeBytes                     int64            `json:"database_size_bytes"`
	CapturedAt                            time.Time        `json:"captured_at"`
	Scope                                 string           `json:"scope"`
	SelectedModules                       []string         `json:"selected_modules,omitempty"`
	Tables                                map[string]int64 `json:"tables"`
	Files                                 []FileIdentity   `json:"files"`
	MediaFileCount                        int              `json:"media_file_count"`
	MediaFileBytes                        int64            `json:"media_file_bytes"`
	Gaps                                  []string         `json:"gaps"`
	StorefrontCompanyStaticReleaseVersion string           `json:"storefront_company_static_release_version,omitempty"`
}

func LoadConfiguration(filename string) (Configuration, error) {
	var cfg Configuration
	data, err := os.ReadFile(filename) // #nosec G304 -- This filename is the operator-supplied maintenance configuration path, not an HTTP parameter.
	if err != nil {
		return cfg, err
	}
	if err := decodeStrictJSON(data, &cfg); err != nil {
		return cfg, err
	}
	if err := cfg.inheritMediaOriginsFromExistingServerAndStorefrontDeploymentURLs(); err != nil {
		return cfg, err
	}
	if cfg.MaxPackageBytes == 0 {
		cfg.MaxPackageBytes = 100 << 30
	}
	if cfg.MaxExpandedBytes == 0 {
		cfg.MaxExpandedBytes = 200 << 30
	}
	if cfg.MaxWorkspaceBytes == 0 {
		cfg.MaxWorkspaceBytes = 500 << 30
	}
	if cfg.MaxFiles == 0 {
		cfg.MaxFiles = 1000000
	}
	if cfg.WorkCopyLifetimeHours == 0 {
		cfg.WorkCopyLifetimeHours = 24
	}
	return cfg, cfg.Validate()
}

func (cfg Configuration) Validate() error {
	if cfg.ProductViewRecoveryRedis != nil {
		if err := cfg.ProductViewRecoveryRedis.validateExplicitProductViewRecoveryRedisConnection(); err != nil {
			return err
		}
		if len(cfg.ProductViewRecoveryWriterInventory) == 0 {
			return errors.New("product view recovery requires the complete single-host writer inventory")
		}
		for _, writer := range cfg.ProductViewRecoveryWriterInventory {
			if strings.TrimSpace(writer) == "" {
				return errors.New("product view writer inventory must not contain empty entries")
			}
		}
	} else if len(cfg.ProductViewRecoveryWriterInventory) > 0 {
		return errors.New("product view writer inventory requires an explicit private Redis connection")
	}
	if cfg.SiteID == "" || cfg.Environment == "" || cfg.ApplicationVersion == "" || cfg.ExpectedMigrationVersion <= 0 || cfg.ApplicationDatabaseRole == "" || !safeDatabaseName.MatchString(cfg.DatabaseName) || cfg.DatabaseName == "postgres" || strings.HasPrefix(cfg.DatabaseName, "template") {
		return errors.New("explicit site, environment, application version and dedicated application database are required")
	}
	if cfg.MaxPackageBytes < ChunkBytes || cfg.MaxExpandedBytes < cfg.MaxPackageBytes || cfg.MaxWorkspaceBytes < cfg.MaxExpandedBytes || cfg.MaxFiles < 1 || cfg.MaxFiles > 1000000 || cfg.WorkCopyLifetimeHours < 1 {
		return errors.New("invalid package, expanded, workspace or lifetime limits")
	}
	if len(cfg.StorageRoots) != len(mediaRoots) {
		return errors.New("all three application-owned storage roots are required")
	}
	paths := []string{cfg.Workspace, cfg.MaintenanceDirectory, cfg.PostgreSQLCapacityDirectory}
	if cfg.ProductStaticAssetsDirectory != "" {
		paths = append(paths, cfg.ProductStaticAssetsDirectory)
	}
	for _, name := range mediaRoots {
		paths = append(paths, cfg.StorageRoots[name])
	}
	for i, path := range paths {
		if !filepath.IsAbs(path) || filepath.Dir(filepath.Clean(path)) == filepath.Clean(path) {
			return errors.New("workspace, gate, PostgreSQL capacity and storage paths must be absolute dedicated directories")
		}
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			return fmt.Errorf("configured directory must already exist: %s", path)
		}
		paths[i] = resolved
		info, err := os.Stat(resolved)
		if err != nil || !info.IsDir() {
			return errors.New("configured path is not a directory")
		}
	}
	for i := range paths {
		for j := i + 1; j < len(paths); j++ {
			if pathContains(paths[i], paths[j]) || pathContains(paths[j], paths[i]) {
				return errors.New("workspace, gate, PostgreSQL capacity and media roots must be disjoint")
			}
		}
	}
	if cfg.RecoverySecrets["JWT_SECRET"] == "" || cfg.RecoverySecrets["PAYMENT_CONFIG_MASTER_KEY"] == "" || cfg.RecoverySecrets["OPS_CONNECTOR_MASTER_KEY"] == "" {
		return errors.New("JWT, payment and connector recovery keys are required")
	}
	for _, path := range []string{cfg.PassphraseFile, cfg.AuthenticationKeyFile} {
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil || !filepath.IsAbs(resolved) {
			return errors.New("credential files must use explicit existing absolute paths")
		}
		for _, root := range paths {
			if pathContains(root, resolved) {
				return errors.New("recovery authentication and encryption credentials must be stored outside media and workspace")
			}
		}
	}
	passphrasePath, _ := filepath.EvalSymlinks(cfg.PassphraseFile)
	authenticationPath, _ := filepath.EvalSymlinks(cfg.AuthenticationKeyFile)
	if passphrasePath == authenticationPath {
		return errors.New("package encryption and maintenance authentication require separate credential files")
	}
	passphrase, err := readProtectedCredential(cfg.PassphraseFile)
	if err != nil {
		return err
	}
	authentication, err := readProtectedCredential(cfg.AuthenticationKeyFile)
	if err == nil && authentication == passphrase {
		return errors.New("maintenance authentication must differ from the package decryption credential")
	}
	return err
}

func pathContains(parent, child string) bool {
	relative, err := filepath.Rel(parent, child)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func readProtectedCredential(filename string) (string, error) {
	info, err := os.Stat(filename)
	if err != nil || !info.Mode().IsRegular() || info.Size() > 4096 {
		return "", errors.New("credential file is missing or invalid")
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0077 != 0 {
		return "", errors.New("credential files must be private (mode 0600 or stricter)")
	}
	data, err := os.ReadFile(filename) // #nosec G304 -- This operator-configured credential file has already passed regular-file, size and private-permission checks.
	if err != nil {
		return "", err
	}
	value := strings.TrimRight(string(data), "\r\n")
	if len(value) < 32 {
		return "", errors.New("credential must contain at least 32 characters")
	}
	return value, nil
}

func newTaskID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes[:])
}

func decodeStrictJSON(data []byte, value any) error {
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return errors.New("unexpected trailing JSON content")
	}
	return nil
}

func writeDurableJSON(filename string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return writeDurableFile(filename, data)
}

func writeDurableFile(filename string, data []byte) error {
	temporary, err := os.OpenFile(filename+".next", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600) // #nosec G304,G703 -- Callers supply operator-configured or manager-owned private workspace paths, never raw HTTP filenames.
	if err != nil {
		return err
	}
	_, writeErr := temporary.Write(data)
	syncErr := temporary.Sync()
	closeErr := temporary.Close()
	if err := errors.Join(writeErr, syncErr, closeErr); err != nil {
		return err
	}
	if err := os.Rename(filename+".next", filename); err != nil { // #nosec G703 -- Callers supply operator-configured or manager-owned private workspace paths, never raw HTTP filenames.
		return err
	}
	return syncDirectory(filepath.Dir(filename))
}

func syncDirectory(directory string) error {
	file, err := os.Open(directory) // #nosec G304,G703 -- Callers supply operator-configured or manager-owned private workspace paths, never raw HTTP filenames.
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(file.Close)
	return syncDirectoryFile(file)
}

func fileIdentity(filename, packagePath string) (FileIdentity, error) {
	file, err := os.Open(filename) // #nosec G304,G703 -- Callers supply operator-configured or manager-owned private workspace paths, never raw HTTP filenames.
	if err != nil {
		return FileIdentity{}, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(file.Close)
	hash := sha256.New()
	size, err := io.Copy(hash, file)
	return FileIdentity{Path: packagePath, Size: size, SHA256: hex.EncodeToString(hash.Sum(nil))}, err
}
