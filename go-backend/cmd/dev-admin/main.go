package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/database"

	"gorm.io/gorm"
)

const (
	defaultEmail    = "admin@example.com"
	defaultUsername = "admin"
	defaultPassword = "Admin123456!"
	defaultRole     = string(auth.RoleAdmin)
)

type devAdminInput struct {
	Email           string
	Username        string
	Password        string
	Role            auth.Role
	ResetConfigured bool
	PasswordFromEnv bool
}

func main() {
	log.SetFlags(0)

	if !truthy(os.Getenv("DEV_ADMIN_BOOTSTRAP")) {
		log.Fatal("refusing to run without DEV_ADMIN_BOOTSTRAP=true")
	}

	if isProductionMode(os.Getenv("SERVER_MODE")) {
		log.Fatal("refusing to bootstrap a dev admin while SERVER_MODE is release/production")
	}

	input, err := readInput()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Init(readDatabaseConfig())
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	sqlDB, err := db.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	changed, err := ensureDevAdmin(db, input)
	if err != nil {
		log.Fatal(err)
	}

	if changed {
		fmt.Printf("DEV admin ready: %s\n", input.Email)
		if input.PasswordFromEnv {
			fmt.Println("DEV admin password: configured by DEV_ADMIN_PASSWORD")
		} else {
			fmt.Printf("DEV admin password: %s\n", input.Password)
		}
		return
	}

	fmt.Println("DEV backoffice user already exists; admin bootstrap skipped.")
	fmt.Println("Set DEV_ADMIN_RESET=true to reset/create the configured DEV admin account.")
}

func readInput() (devAdminInput, error) {
	password, passwordFromEnv := lookupTrimmedEnv("DEV_ADMIN_PASSWORD")
	if password == "" {
		password = defaultPassword
		passwordFromEnv = false
	}

	role := auth.NormalizeRole(envDefault("DEV_ADMIN_ROLE", defaultRole))
	input := devAdminInput{
		Email:           strings.ToLower(envDefault("DEV_ADMIN_EMAIL", defaultEmail)),
		Username:        envDefault("DEV_ADMIN_USERNAME", defaultUsername),
		Password:        password,
		Role:            role,
		ResetConfigured: truthy(os.Getenv("DEV_ADMIN_RESET")),
		PasswordFromEnv: passwordFromEnv,
	}

	if !strings.Contains(input.Email, "@") {
		return input, fmt.Errorf("DEV_ADMIN_EMAIL must be an email address")
	}
	if input.Username == "" {
		return input, fmt.Errorf("DEV_ADMIN_USERNAME must not be empty")
	}
	if len(input.Password) < 6 {
		return input, fmt.Errorf("DEV_ADMIN_PASSWORD must be at least 6 characters")
	}
	if !auth.IsBackofficeRole(input.Role.String()) {
		return input, fmt.Errorf("DEV_ADMIN_ROLE must be one of admin, manager, editor, support")
	}

	return input, nil
}

func ensureDevAdmin(db *gorm.DB, input devAdminInput) (bool, error) {
	if !input.ResetConfigured {
		var count int64
		if err := db.Model(&user.User{}).
			Where("role IN ? AND status = ?", []string{"admin", "manager", "editor", "support"}, "active").
			Count(&count).Error; err != nil {
			return false, fmt.Errorf("count backoffice users: %w", err)
		}
		if count > 0 {
			return false, nil
		}
	}

	var existing user.User
	err := db.Where("email = ?", input.Email).First(&existing).Error
	if err == nil {
		existing.Username = input.Username
		existing.Role = input.Role.String()
		existing.Status = "active"
		existing.FirstName = "Dev"
		existing.LastName = "Admin"
		if err := existing.HashPassword(input.Password); err != nil {
			return false, fmt.Errorf("hash password: %w", err)
		}
		if err := db.Save(&existing).Error; err != nil {
			return false, fmt.Errorf("update dev admin: %w", err)
		}
		return true, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, fmt.Errorf("find dev admin by email: %w", err)
	}

	var usernameCount int64
	if err := db.Model(&user.User{}).Where("username = ?", input.Username).Count(&usernameCount).Error; err != nil {
		return false, fmt.Errorf("check dev admin username: %w", err)
	}
	if usernameCount > 0 {
		return false, fmt.Errorf("DEV_ADMIN_USERNAME %q is already used by another account", input.Username)
	}

	newUser := user.User{
		Email:     input.Email,
		Username:  input.Username,
		FirstName: "Dev",
		LastName:  "Admin",
		Role:      input.Role.String(),
		Status:    "active",
	}
	if err := newUser.HashPassword(input.Password); err != nil {
		return false, fmt.Errorf("hash password: %w", err)
	}
	if err := db.Create(&newUser).Error; err != nil {
		return false, fmt.Errorf("create dev admin: %w", err)
	}

	return true, nil
}

func readDatabaseConfig() config.DatabaseConfig {
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
	}
}

func isProductionMode(raw string) bool {
	mode := strings.ToLower(strings.TrimSpace(raw))
	return mode == "release" || mode == "production" || mode == "prod"
}

func envDefault(key, fallback string) string {
	if value, ok := lookupTrimmedEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func envDefaultAny(keys []string, fallback string) string {
	for _, key := range keys {
		if value, ok := lookupTrimmedEnv(key); ok && value != "" {
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

func lookupTrimmedEnv(key string) (string, bool) {
	value, ok := os.LookupEnv(key)
	return strings.TrimSpace(value), ok
}

func truthy(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}
