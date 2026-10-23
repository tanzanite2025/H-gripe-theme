package main

import (
	"encoding/json"
	"io"

	"commerce-platform/internal/localbackup"
)

// Only an explicit allowlist is written. Never serialize Configuration: it
// contains the database password and the original business recovery secrets.
// This check deliberately returns before NewManager, which can recover tasks
// and dispose of interrupted candidates even without starting an HTTP server.
func checkIndependentBackupConfigurationAndWriteCredentialFreeReport(configurationPath string, output io.Writer) error {
	cfg, err := localbackup.LoadConfiguration(configurationPath)
	if err != nil {
		return err
	}
	report := struct {
		Scope                               string   `json:"scope"`
		OfflineConfigurationValid           bool     `json:"offline_configuration_valid"`
		TargetEnvironmentAcceptanceComplete bool     `json:"target_environment_acceptance_complete"`
		SiteID                              string   `json:"configured_site_id"`
		Environment                         string   `json:"configured_environment"`
		ApplicationVersion                  string   `json:"configured_application_version"`
		DatabaseName                        string   `json:"configured_database_name"`
		ExpectedMigrationVersion            int64    `json:"configured_expected_migration_version"`
		UnverifiedChecks                    []string `json:"unverified_checks"`
	}{
		Scope:                               "offline_configuration_only",
		OfflineConfigurationValid:           true,
		TargetEnvironmentAcceptanceComplete: false,
		SiteID:                              cfg.SiteID,
		Environment:                         cfg.Environment,
		ApplicationVersion:                  cfg.ApplicationVersion,
		DatabaseName:                        cfg.DatabaseName,
		ExpectedMigrationVersion:            cfg.ExpectedMigrationVersion,
		UnverifiedChecks: []string{
			"database_connectivity_server_version_privileges_and_actual_migration",
			"registered_application_build_storage_and_recovery_key_boundary_match",
			"actual_mounts_ownership_writability_and_available_capacity",
			"all_application_and_external_writers_stopped",
			"computer_held_package_identity_integrity_and_decryption",
			"business_attachments_external_facts_and_projection_reconciliation",
			"target_scale_restore_duration_failure_recovery_and_data_loss_window",
		},
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}
