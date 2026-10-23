package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/localbackup"
)

// Run the real main function in a subprocess so a regression that starts the
// executor, exits fatally or listens for HTTP cannot hide behind a helper test.
func TestIndependentBackupConfigurationOnlyCommandSubprocess(t *testing.T) {
	if os.Getenv("LOCAL_BACKUP_CONFIGURATION_ONLY_TEST_SUBPROCESS") != "1" {
		return
	}
	flag.CommandLine = flag.NewFlagSet("local-backup-maintenance", flag.ExitOnError)
	os.Args = []string{"local-backup-maintenance", "-config", os.Getenv("LOCAL_BACKUP_CONFIGURATION_ONLY_TEST_FILE"), "-check-configuration-only", "-listen", os.Getenv("LOCAL_BACKUP_CONFIGURATION_ONLY_TEST_OCCUPIED_ADDRESS")}
	main()
	os.Exit(0)
}

func TestIndependentBackupConfigurationOnlyCommandDoesNotConnectRecoverLockOrExposeCredentials(t *testing.T) {
	for _, scenario := range []string{"valid_with_interrupted_and_malformed_tasks", "identical_credentials", "nested_media_directory", "unknown_configuration_field"} {
		t.Run(scenario, func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { listener.Close() })
			connectionObserved := make(chan bool, 1)
			go func() {
				connection, err := listener.Accept()
				if connection != nil {
					connection.Close()
				}
				connectionObserved <- err == nil
			}()
			root := t.TempDir()
			cfg := createConfigurationOnlyCommandSyntheticFixture(t, root, listener.Addr().String())
			if scenario == "identical_credentials" {
				data, err := os.ReadFile(cfg.PassphraseFile)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(cfg.AuthenticationKeyFile, data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			if scenario == "nested_media_directory" {
				cfg.StorageRoots["uploads_private"] = cfg.Workspace
			}
			configurationBytes, err := json.Marshal(cfg)
			if err != nil {
				t.Fatal(err)
			}
			if scenario == "unknown_configuration_field" {
				configurationBytes = append(configurationBytes[:len(configurationBytes)-1], []byte(`,"unexpected_setting":"synthetic-unknown-value-must-stay-private"}`)...)
			}
			configurationPath := filepath.Join(root, "synthetic-independent-backup-configuration.json")
			if err := os.WriteFile(configurationPath, configurationBytes, 0600); err != nil {
				t.Fatal(err)
			}
			before := fingerprintConfigurationOnlyCommandFixtureFiles(t, root)
			executable, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, executable, "-test.run=^TestIndependentBackupConfigurationOnlyCommandSubprocess$")
			command.Env = append(os.Environ(), "LOCAL_BACKUP_CONFIGURATION_ONLY_TEST_SUBPROCESS=1", "LOCAL_BACKUP_CONFIGURATION_ONLY_TEST_FILE="+configurationPath, "LOCAL_BACKUP_CONFIGURATION_ONLY_TEST_OCCUPIED_ADDRESS="+listener.Addr().String())
			output, commandErr := command.CombinedOutput()
			listener.Close()
			if <-connectionObserved {
				t.Error("configuration-only command opened a database connection")
			}
			if ctx.Err() != nil {
				t.Fatal("configuration-only command did not exit without starting a server")
			}
			if after := fingerprintConfigurationOnlyCommandFixtureFiles(t, root); !reflect.DeepEqual(before, after) {
				t.Error("configuration-only command changed recovery files or created executor locks")
			}
			secrets := []string{cfg.DatabaseURL, "synthetic-database-password-must-stay-private", "synthetic-package-passphrase-must-stay-private", "synthetic-maintenance-authentication-must-stay-private", "synthetic-unknown-value-must-stay-private"}
			for _, secret := range cfg.RecoverySecrets {
				secrets = append(secrets, secret)
			}
			for _, secret := range secrets {
				if strings.Contains(string(output), secret) {
					t.Fatal("configuration-only output exposed a credential")
				}
			}
			if scenario != "valid_with_interrupted_and_malformed_tasks" {
				if commandErr == nil {
					t.Fatal("invalid configuration was accepted")
				}
				if strings.Contains(string(output), `"offline_configuration_valid": true`) {
					t.Fatal("failed validation emitted a success report")
				}
				return
			}
			if commandErr != nil {
				t.Fatalf("offline check failed: %v", commandErr)
			}
			var report map[string]any
			if err := json.Unmarshal(output, &report); err != nil {
				t.Fatalf("command did not emit a JSON report: %v", err)
			}
			if report["scope"] != "offline_configuration_only" || report["offline_configuration_valid"] != true || report["target_environment_acceptance_complete"] != false {
				t.Error("report did not distinguish offline configuration validity from target acceptance")
			}
			if report["configured_site_id"] != cfg.SiteID || report["configured_database_name"] != cfg.DatabaseName || report["configured_expected_migration_version"] != float64(cfg.ExpectedMigrationVersion) {
				t.Error("report did not identify the configured recovery target")
			}
			if checks, ok := report["unverified_checks"].([]any); !ok || len(checks) == 0 {
				t.Error("report omitted the remaining target acceptance checks")
			}
		})
	}
}

func createConfigurationOnlyCommandSyntheticFixture(t *testing.T, root, databaseAddress string) localbackup.Configuration {
	t.Helper()
	cfg := localbackup.Configuration{
		SiteID: "synthetic-offline-store", Environment: "test", ApplicationVersion: "synthetic-build-configuration-only",
		ExpectedMigrationVersion: 434, ApplicationDatabaseRole: "synthetic_application_role", DatabaseName: "synthetic_store",
		DatabaseURL: "postgres://synthetic_maintenance:synthetic-database-password-must-stay-private@" + databaseAddress + "/postgres?sslmode=disable",
		Workspace:   filepath.Join(root, "workspace"), MaintenanceDirectory: filepath.Join(root, "gate"),
		PostgreSQLCapacityDirectory: filepath.Join(root, "postgresql-capacity"),
		StorageRoots:                map[string]string{},
		RecoverySecrets:             map[string]string{"JWT_SECRET": "synthetic-jwt-secret-must-stay-private", "PAYMENT_CONFIG_MASTER_KEY": "synthetic-payment-key-must-stay-private", "OPS_CONNECTOR_MASTER_KEY": "synthetic-connector-key-must-stay-private"},
		PassphraseFile:              filepath.Join(root, "package-passphrase"), AuthenticationKeyFile: filepath.Join(root, "maintenance-authentication"),
	}
	for _, name := range []string{"uploads", "uploads_private", "site_logo_uploads"} {
		cfg.StorageRoots[name] = filepath.Join(root, name)
	}
	for _, directory := range []string{cfg.Workspace, cfg.MaintenanceDirectory, cfg.PostgreSQLCapacityDirectory, cfg.StorageRoots["uploads"], cfg.StorageRoots["uploads_private"], cfg.StorageRoots["site_logo_uploads"]} {
		if err := os.Mkdir(directory, 0700); err != nil {
			t.Fatal(err)
		}
	}
	for filename, contents := range map[string]string{
		cfg.PassphraseFile:        "synthetic-package-passphrase-must-stay-private",
		cfg.AuthenticationKeyFile: "synthetic-maintenance-authentication-must-stay-private",
		filepath.Join(cfg.MaintenanceDirectory, "application-maintenance.json"): "existing-persistent-maintenance-fence",
	} {
		if err := os.WriteFile(filename, []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, taskID := range []string{strings.Repeat("a", 32), strings.Repeat("b", 32)} {
		directory := filepath.Join(cfg.Workspace, taskID)
		if err := os.Mkdir(directory, 0700); err != nil {
			t.Fatal(err)
		}
		contents := `{"id":"` + taskID + `","state":"running"}`
		if strings.HasPrefix(taskID, "b") {
			contents = "deliberately malformed interrupted task journal"
		}
		if err := os.WriteFile(filepath.Join(directory, "task.json"), []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return cfg
}

func fingerprintConfigurationOnlyCommandFixtureFiles(t *testing.T, root string) map[string]string {
	t.Helper()
	fingerprints := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		identity := info.Mode().String()
		if !entry.IsDir() {
			contents, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			identity += fmt.Sprintf(":%x", sha256.Sum256(contents))
		}
		fingerprints[path] = identity
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return fingerprints
}
