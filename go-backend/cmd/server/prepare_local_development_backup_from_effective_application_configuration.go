package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/localbackup"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/storage"
)

// Read effective application settings instead of duplicating YAML defaults or
// inventing recovery keys. This command only supports loopback development.
func prepareLocalDevelopmentBackupFromEffectiveApplicationConfiguration(cfg *config.Config) error {
	if cfg.Server.Mode != "debug" || (cfg.Database.Host != "localhost" && cfg.Database.Host != "127.0.0.1" && cfg.Database.Host != "::1") || Version == "dev" {
		return errors.New("local backup preparation requires a loopback debug database and an explicit build version")
	}
	filename := os.Getenv("LOCAL_DEVELOPMENT_BACKUP_CONFIGURATION_PATH")
	if !filepath.IsAbs(filename) {
		return errors.New("an absolute private local backup configuration path is required")
	}
	privateDirectory := filepath.Dir(filename)
	var previous localbackup.Configuration
	if data, err := os.ReadFile(filename); err == nil { // #nosec G304 G703 -- The loopback-only development command uses operator-configured absolute private paths and fixed credential filenames; no HTTP input is accepted.
		if err := json.Unmarshal(data, &previous); err != nil {
			return errors.New("existing local backup configuration is invalid; it was retained")
		}
		if previous.DatabaseName != cfg.Database.Database || previous.Environment != "development" {
			return errors.New("existing recovery materials belong to another local database or environment")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := sql.Open("postgres", cfg.Database.GetDSN())
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(db.Close)
	var migrationVersion int64
	var dirty bool
	if err := db.QueryRowContext(ctx, "SELECT version, dirty FROM schema_migrations LIMIT 1").Scan(&migrationVersion, &dirty); err != nil {
		return errors.New("cannot read local database migration identity")
	}
	if dirty {
		return errors.New("local database migration is dirty; backup preparation stopped")
	}
	mediaOrigins, err := collectLocalDevelopmentBackupExplicitHistoricalMediaOrigins(cfg.Server.BaseURL, previous.PublicOrigins, os.Getenv("LOCAL_BACKUP_REVIEWED_HISTORICAL_MEDIA_ORIGINS"))
	if err != nil {
		return err
	}
	if os.Getenv("OPS_CONNECTOR_MASTER_KEY") == "" {
		originalKey := previous.RecoverySecrets["OPS_CONNECTOR_MASTER_KEY"]
		if originalKey == "" {
			var encryptedCount int
			if err := db.QueryRowContext(ctx, "SELECT count(*) FROM ops_connectors WHERE credentials_encrypted <> ''").Scan(&encryptedCount); err != nil {
				return errors.New("cannot verify whether an original connector key is required")
			}
			if encryptedCount != 0 {
				return errors.New("existing encrypted connectors require their original OPS_CONNECTOR_MASTER_KEY; no replacement was generated")
			}
			originalKey, err = createPrivateLocalDevelopmentBackupCredential(filepath.Join(privateDirectory, "local-development-connector-master-key"))
			if err != nil {
				return err
			}
		}
		if err := os.Setenv("OPS_CONNECTOR_MASTER_KEY", originalKey); err != nil {
			return err
		}
	}
	shared := storage.LoadConfigFromEnv()
	logo := storage.LoadSiteLogoConfigFromEnv(shared)
	if shared.Type != storage.StorageTypeLocal || logo.Type != storage.StorageTypeLocal {
		return errors.New("local development backup requires local media storage")
	}
	if shared.PrivateLocalPath == "" {
		shared.PrivateLocalPath = filepath.Join(filepath.Dir(shared.LocalPath), filepath.Base(shared.LocalPath)+"-private")
	}
	roots := map[string]string{"uploads": shared.LocalPath, "uploads_private": shared.PrivateLocalPath, "site_logo_uploads": logo.LocalPath}
	for name, directory := range roots {
		absolute, err := filepath.Abs(directory)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(absolute, 0700); err != nil {
			return err
		}
		roots[name], err = filepath.EvalSymlinks(absolute)
		if err != nil {
			return err
		}
	}
	siteID := previous.SiteID
	if siteID == "" {
		var identity [12]byte
		if _, err := rand.Read(identity[:]); err != nil {
			return err
		}
		siteID = "local-development-" + hex.EncodeToString(identity[:])
	}
	maintenance := localbackup.Configuration{SiteID: siteID, Environment: "development", ApplicationVersion: Version, ExpectedMigrationVersion: migrationVersion, ApplicationDatabaseRole: cfg.Database.Username, DatabaseName: cfg.Database.Database,
		Workspace: filepath.Join(privateDirectory, "workspace"), MaintenanceDirectory: filepath.Join(privateDirectory, "maintenance-gate"), PostgreSQLCapacityDirectory: filepath.Join(privateDirectory, "database-capacity"), StorageRoots: roots, RecoverySecrets: collectRecoverySigningAndEncryptionKeysFromApplicationConfiguration(cfg),
		PassphraseFile: filepath.Join(privateDirectory, "backup-package-decryption-credential"), AuthenticationKeyFile: filepath.Join(privateDirectory, "maintenance-authentication-credential"), ProductStaticAssetsDirectory: os.Getenv("LOCAL_DEVELOPMENT_BACKUP_STATIC_ASSETS_DIRECTORY"),
		PublicOrigins: mediaOrigins, MaxPackageBytes: 100 << 30, MaxExpandedBytes: 200 << 30, MaxWorkspaceBytes: 500 << 30, MaxFiles: 1000000, WorkCopyLifetimeHours: 24}
	connection := url.URL{Scheme: "postgres", User: url.UserPassword(cfg.Database.Username, cfg.Database.Password), Host: net.JoinHostPort(cfg.Database.Host, strconv.Itoa(cfg.Database.Port)), Path: "/postgres", RawQuery: "sslmode=disable"}
	maintenance.DatabaseURL = connection.String()
	for _, directory := range []string{maintenance.Workspace, maintenance.MaintenanceDirectory, maintenance.PostgreSQLCapacityDirectory} {
		if err := os.MkdirAll(directory, 0700); err != nil { // #nosec G703 -- The loopback-only development command uses operator-configured absolute private paths and fixed credential filenames; no HTTP input is accepted.
			return err
		}
	}
	for _, credential := range []string{maintenance.PassphraseFile, maintenance.AuthenticationKeyFile} {
		if _, err := createPrivateLocalDevelopmentBackupCredential(credential); err != nil {
			return err
		}
	}
	if err := maintenance.Validate(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(maintenance, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filename+".next", data, 0600); err != nil { // #nosec G703 -- The loopback-only development command uses operator-configured absolute private paths and fixed credential filenames; no HTTP input is accepted.
		return err
	}
	return os.Rename(filename+".next", filename) // #nosec G703 -- The loopback-only development command uses operator-configured absolute private paths and fixed credential filenames; no HTTP input is accepted.
}

// Only explicit loopback origins are accepted here; neither database content
// nor arbitrary hostnames can silently expand the local backup boundary.
func collectLocalDevelopmentBackupExplicitHistoricalMediaOrigins(current string, previous []string, reviewed string) ([]string, error) {
	values := append([]string{current}, previous...)
	if reviewed != "" {
		values = append(values, strings.Split(reviewed, ",")...)
	}
	origins := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		parsed, err := url.Parse(strings.TrimSpace(value))
		if err != nil || parsed == nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") {
			return nil, errors.New("local backup reviewed media sources must be explicit loopback HTTP origins")
		}
		address := net.ParseIP(parsed.Hostname())
		if !strings.EqualFold(parsed.Hostname(), "localhost") && (address == nil || !address.IsLoopback()) {
			return nil, errors.New("local backup reviewed media sources must be explicit loopback HTTP origins")
		}
		origin := parsed.Scheme + "://" + strings.ToLower(parsed.Host)
		if !seen[origin] {
			origins = append(origins, origin)
			seen[origin] = true
		}
	}
	return origins, nil
}

func createPrivateLocalDevelopmentBackupCredential(filename string) (string, error) {
	data, err := os.ReadFile(filename) // #nosec G304 G703 -- The loopback-only development command uses operator-configured absolute private paths and fixed credential filenames; no HTTP input is accepted.
	if err == nil {
		if len(data) < 32 {
			return "", errors.New("existing local credential is invalid and was retained")
		}
		return string(data), nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	var secret [32]byte
	if _, err := rand.Read(secret[:]); err != nil {
		return "", err
	}
	value := hex.EncodeToString(secret[:])
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600) // #nosec G304 G703 -- The loopback-only development command uses operator-configured absolute private paths and fixed credential filenames; no HTTP input is accepted.
	if err != nil {
		return "", err
	}
	_, writeErr := file.WriteString(value)
	closeErr := file.Close()
	return value, errors.Join(writeErr, closeErr)
}
