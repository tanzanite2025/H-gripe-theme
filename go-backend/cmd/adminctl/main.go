package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/database"
	"commerce-platform/internal/service"
)

const productionConfirmationValue = "reset-production-admin"

func main() {
	log.SetFlags(0)
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		log.Fatal(err)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		printUsage(stderr)
		return errors.New("command is required")
	}

	switch args[0] {
	case "ensure-admin":
		return runEnsureAdmin(args[1:], stdout, stderr)
	case "help", "-h", "--help":
		printUsage(stdout)
		return nil
	default:
		printUsage(stderr)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runEnsureAdmin(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("ensure-admin", flag.ContinueOnError)
	fs.SetOutput(stderr)
	configPath := fs.String("config", "", "optional app config file")
	operator := fs.String("operator", envDefault("ADMIN_OPERATOR", "adminctl"), "operator label written to audit_logs")
	if err := fs.Parse(args); err != nil {
		return err
	}

	dbCfg, serverMode, err := readRuntimeConfig(*configPath)
	if err != nil {
		return err
	}
	if err := requireProductionConfirmation(serverMode); err != nil {
		return err
	}

	password, passwordSource, err := readAdminPassword()
	if err != nil {
		return err
	}

	db, err := database.Init(dbCfg)
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	result, err := service.NewAdminAccountMaintenanceService(db).EnsureBackofficeAccount(service.AdminAccountMaintenanceInput{
		Email:     envTrim("ADMIN_EMAIL"),
		Username:  envTrim("ADMIN_USERNAME"),
		Password:  password,
		Role:      envTrim("ADMIN_ROLE"),
		FirstName: envTrim("ADMIN_FIRST_NAME"),
		LastName:  envTrim("ADMIN_LAST_NAME"),
		Locale:    envTrim("ADMIN_LOCALE"),
		Operator:  *operator,
	})
	if err != nil {
		return err
	}

	action := "reset"
	if result.Created {
		action = "created"
	}
	fmt.Fprintf(stdout, "Admin account %s: id=%d email=%s username=%s role=%s status=%s\n", action, result.UserID, result.Email, result.Username, result.Role, result.Status)
	fmt.Fprintf(stdout, "Password accepted from %s and was not printed.\n", passwordSource)
	return nil
}

func readRuntimeConfig(configPath string) (config.DatabaseConfig, string, error) {
	if strings.TrimSpace(configPath) != "" {
		cfg, err := config.Load(configPath)
		if err != nil {
			return config.DatabaseConfig{}, "", fmt.Errorf("load config: %w", err)
		}
		return cfg.Database, cfg.Server.Mode, nil
	}

	return config.DatabaseConfig{
		Driver:          envDefaultAny([]string{"DB_DRIVER", "DATABASE_DRIVER"}, "postgres"),
		Host:            envDefaultAny([]string{"DB_HOST", "DATABASE_HOST"}, "localhost"),
		Port:            envIntDefaultAny([]string{"DB_PORT", "DATABASE_PORT"}, 9400),
		Username:        envDefaultAny([]string{"DB_USERNAME", "DATABASE_USERNAME"}, "commerce_platform"),
		Password:        envDefaultAny([]string{"DB_PASSWORD", "DATABASE_PASSWORD"}, "commerce_platform_password"),
		Database:        envDefaultAny([]string{"DB_NAME", "DATABASE_NAME"}, "commerce_platform"),
		MaxIdleConns:    2,
		MaxOpenConns:    5,
		ConnMaxLifetime: 3600,
		LogLevel:        envDefaultAny([]string{"DB_LOG_LEVEL", "DATABASE_LOG_LEVEL"}, "silent"),
	}, envDefault("SERVER_MODE", "debug"), nil
}

func readAdminPassword() (string, string, error) {
	passwordFile := envTrim("ADMIN_PASSWORD_FILE")
	password := envTrim("ADMIN_PASSWORD")
	if passwordFile != "" && password != "" {
		return "", "", errors.New("set only one of ADMIN_PASSWORD or ADMIN_PASSWORD_FILE")
	}
	if passwordFile != "" {
		payload, err := os.ReadFile(passwordFile)
		if err != nil {
			return "", "", fmt.Errorf("read ADMIN_PASSWORD_FILE: %w", err)
		}
		return strings.TrimRight(string(payload), "\r\n"), "ADMIN_PASSWORD_FILE", nil
	}
	if password != "" {
		return password, "ADMIN_PASSWORD", nil
	}
	return "", "", errors.New("ADMIN_PASSWORD or ADMIN_PASSWORD_FILE is required")
}

func requireProductionConfirmation(serverMode string) error {
	if !isProductionMode(serverMode) {
		return nil
	}
	if envTrim("ADMINCTL_CONFIRM") != productionConfirmationValue {
		return fmt.Errorf("SERVER_MODE=%s requires ADMINCTL_CONFIRM=%s", serverMode, productionConfirmationValue)
	}
	return nil
}

func isProductionMode(raw string) bool {
	mode := strings.ToLower(strings.TrimSpace(raw))
	return mode == "release" || mode == "production" || mode == "prod"
}

func envTrim(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func envDefault(key, fallback string) string {
	if value := envTrim(key); value != "" {
		return value
	}
	return fallback
}

func envDefaultAny(keys []string, fallback string) string {
	for _, key := range keys {
		if value := envTrim(key); value != "" {
			return value
		}
	}
	return fallback
}

func envIntDefaultAny(keys []string, fallback int) int {
	value := envDefaultAny(keys, "")
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: adminctl ensure-admin [-config path] [-operator label]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Required environment:")
	fmt.Fprintln(w, "  ADMIN_EMAIL")
	fmt.Fprintln(w, "  ADMIN_PASSWORD or ADMIN_PASSWORD_FILE")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Optional environment:")
	fmt.Fprintln(w, "  ADMIN_USERNAME, ADMIN_ROLE, ADMIN_FIRST_NAME, ADMIN_LAST_NAME, ADMIN_LOCALE")
	fmt.Fprintln(w, "  ADMINCTL_CONFIRM=reset-production-admin when SERVER_MODE is release/production/prod")
}
