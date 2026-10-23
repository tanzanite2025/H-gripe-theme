package localbackup

import (
	"crypto/subtle"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const PermissionView = "ops:backup:view"
const PermissionExport = "ops:backup:export"
const PermissionUpload = "ops:backup:upload"
const PermissionRestore = "ops:backup:restore"

type MaintenanceSession struct {
	Permissions  []string `json:"permissions"`
	DownloadTask string   `json:"download_task,omitempty"`
	jwt.RegisteredClaims
}

func IssueMaintenanceSession(key, site, environment, actor string, permissions []string) (string, error) {
	if len(key) < 32 || site == "" || environment == "" || actor == "" {
		return "", errors.New("maintenance authentication is not configured")
	}
	now := time.Now()
	claims := MaintenanceSession{Permissions: permissions, RegisteredClaims: jwt.RegisteredClaims{Subject: actor, Issuer: "independent-local-backup", Audience: jwt.ClaimStrings{site + "/" + environment}, IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(30 * time.Minute))}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(key))
}

func (cfg Configuration) AuthenticateMaintenanceSession(value string) (*MaintenanceSession, error) {
	key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
	if err != nil {
		return nil, err
	}
	if len(value) == len(key) && subtle.ConstantTimeCompare([]byte(value), []byte(key)) == 1 {
		return &MaintenanceSession{Permissions: []string{PermissionView, PermissionExport, PermissionUpload, PermissionRestore}, RegisteredClaims: jwt.RegisteredClaims{Subject: "independent-recovery-operator"}}, nil
	}
	claims := &MaintenanceSession{}
	token, err := jwt.ParseWithClaims(value, claims, func(token *jwt.Token) (any, error) { return []byte(key), nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired(), jwt.WithIssuer("independent-local-backup"), jwt.WithAudience(cfg.SiteID+"/"+cfg.Environment))
	if err != nil || !token.Valid || claims.Subject == "" {
		return nil, errors.New("maintenance session is invalid or expired")
	}
	return claims, nil
}

func (claims *MaintenanceSession) HasPermission(permission string) bool {
	for _, allowed := range claims.Permissions {
		if allowed == permission {
			return true
		}
	}
	return false
}

func (cfg Configuration) issueDownloadSession(actor, id string) (string, error) {
	key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
	if err != nil {
		return "", err
	}
	claims := MaintenanceSession{DownloadTask: id, Permissions: []string{PermissionExport}, RegisteredClaims: jwt.RegisteredClaims{Subject: actor, Issuer: "independent-local-backup", Audience: jwt.ClaimStrings{cfg.SiteID + "/" + cfg.Environment}, ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute))}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(key))
}
