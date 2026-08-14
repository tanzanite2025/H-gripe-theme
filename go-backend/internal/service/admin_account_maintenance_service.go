package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/domain/user"

	"gorm.io/gorm"
)

var (
	ErrAdminAccountStoreUnavailable = errors.New("admin account store is unavailable")
	ErrAdminAccountEmailRequired    = errors.New("admin account email is required")
	ErrAdminAccountEmailInvalid     = errors.New("admin account email is invalid")
	ErrAdminAccountUsernameInvalid  = errors.New("admin account username is invalid")
	ErrAdminAccountPasswordRequired = errors.New("admin account password is required")
	ErrAdminAccountWeakPassword     = errors.New("admin account password is too weak")
	ErrAdminAccountRoleForbidden    = errors.New("admin account role must be a backoffice role")
)

type AdminAccountMaintenanceService struct {
	db *gorm.DB
}

type AdminAccountMaintenanceInput struct {
	Email       string
	Username    string
	Password    string
	Role        string
	FirstName   string
	LastName    string
	Locale      string
	Operator    string
	AuditMethod string
	AuditPath   string
}

type AdminAccountMaintenanceResult struct {
	UserID   uint   `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Status   string `json:"status"`
	Created  bool   `json:"created"`
}

type AdminAccountMaintenanceAccount struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Role      string    `json:"role"`
	Locale    string    `json:"locale"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type normalizedAdminAccountMaintenanceInput struct {
	Email             string
	Username          string
	UsernameSpecified bool
	Password          string
	Role              auth.Role
	RoleSpecified     bool
	FirstName         string
	LastName          string
	Locale            string
	LocaleSpecified   bool
	Operator          string
	AuditMethod       string
	AuditPath         string
}

type adminAccountSnapshot struct {
	ID       uint   `json:"id,omitempty"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Status   string `json:"status"`
	Locale   string `json:"locale"`
	Deleted  bool   `json:"deleted,omitempty"`
}

func NewAdminAccountMaintenanceService(db *gorm.DB) *AdminAccountMaintenanceService {
	return &AdminAccountMaintenanceService{db: db}
}

func (s *AdminAccountMaintenanceService) EnsureBackofficeAccount(input AdminAccountMaintenanceInput) (*AdminAccountMaintenanceResult, error) {
	if s == nil || s.db == nil {
		return nil, ErrAdminAccountStoreUnavailable
	}

	normalized, err := normalizeAdminAccountMaintenanceInput(input)
	if err != nil {
		return nil, err
	}

	var result AdminAccountMaintenanceResult
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var existing user.User
		err := tx.Unscoped().Where("email = ?", normalized.Email).First(&existing).Error
		switch {
		case err == nil:
			updateInput := normalized
			if !updateInput.UsernameSpecified && strings.TrimSpace(existing.Username) != "" {
				updateInput.Username = existing.Username
			}
			if err := ensureAdminAccountUsernameAvailable(tx, updateInput.Username, existing.ID); err != nil {
				return err
			}
			oldValue := marshalAdminAccountSnapshot(existing)
			if err := applyAdminAccountMaintenanceInput(&existing, updateInput); err != nil {
				return err
			}
			if err := tx.Unscoped().Save(&existing).Error; err != nil {
				return fmt.Errorf("update admin account: %w", err)
			}
			if err := createAdminAccountMaintenanceAuditLog(tx, existing, updateInput, "reset_password", oldValue, marshalAdminAccountSnapshot(existing)); err != nil {
				return err
			}
			result = adminAccountMaintenanceResult(existing, false)
			return nil
		case errors.Is(err, gorm.ErrRecordNotFound):
			if err := ensureAdminAccountUsernameAvailable(tx, normalized.Username, 0); err != nil {
				return err
			}
			newUser := user.User{}
			if err := applyAdminAccountMaintenanceInput(&newUser, normalized); err != nil {
				return err
			}
			if err := tx.Create(&newUser).Error; err != nil {
				return fmt.Errorf("create admin account: %w", err)
			}
			if err := createAdminAccountMaintenanceAuditLog(tx, newUser, normalized, "create", "{}", marshalAdminAccountSnapshot(newUser)); err != nil {
				return err
			}
			result = adminAccountMaintenanceResult(newUser, true)
			return nil
		default:
			return fmt.Errorf("find admin account by email: %w", err)
		}
	})
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *AdminAccountMaintenanceService) ListBackofficeAccounts(search string) ([]AdminAccountMaintenanceAccount, error) {
	if s == nil || s.db == nil {
		return nil, ErrAdminAccountStoreUnavailable
	}

	query := s.db.Unscoped().Where("role IN ?", []string{
		auth.RoleAdmin.String(),
		auth.RoleManager.String(),
		auth.RoleEditor.String(),
		auth.RoleSupport.String(),
	})
	if trimmedSearch := strings.TrimSpace(search); trimmedSearch != "" {
		pattern := "%" + trimmedSearch + "%"
		query = query.Where(
			"email LIKE ? OR username LIKE ? OR first_name LIKE ? OR last_name LIKE ?",
			pattern,
			pattern,
			pattern,
			pattern,
		)
	}

	var users []user.User
	if err := query.Order("id ASC").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("list backoffice accounts: %w", err)
	}

	accounts := make([]AdminAccountMaintenanceAccount, 0, len(users))
	for _, account := range users {
		accounts = append(accounts, AdminAccountMaintenanceAccount{
			ID:        account.ID,
			Email:     account.Email,
			Username:  account.Username,
			FirstName: account.FirstName,
			LastName:  account.LastName,
			Role:      auth.NormalizeRole(account.Role).String(),
			Locale:    account.Locale,
			Status:    account.Status,
			CreatedAt: account.CreatedAt,
			UpdatedAt: account.UpdatedAt,
		})
	}
	return accounts, nil
}

func normalizeAdminAccountMaintenanceInput(input AdminAccountMaintenanceInput) (normalizedAdminAccountMaintenanceInput, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if email == "" {
		return normalizedAdminAccountMaintenanceInput{}, ErrAdminAccountEmailRequired
	}
	parsedAddress, err := mail.ParseAddress(email)
	if err != nil || parsedAddress.Address != email {
		return normalizedAdminAccountMaintenanceInput{}, ErrAdminAccountEmailInvalid
	}

	password := input.Password
	if strings.TrimSpace(password) == "" {
		return normalizedAdminAccountMaintenanceInput{}, ErrAdminAccountPasswordRequired
	}
	if err := validateAdminAccountPassword(password); err != nil {
		return normalizedAdminAccountMaintenanceInput{}, err
	}

	roleSpecified := strings.TrimSpace(input.Role) != ""
	role := auth.NormalizeRole(input.Role)
	if !roleSpecified {
		role = auth.RoleAdmin
	}
	if !auth.IsBackofficeRole(role.String()) {
		return normalizedAdminAccountMaintenanceInput{}, ErrAdminAccountRoleForbidden
	}

	username := strings.TrimSpace(input.Username)
	usernameSpecified := username != ""
	if username == "" {
		username = deriveAdminUsername(email)
	}
	if err := validateAdminAccountUsername(username); err != nil {
		return normalizedAdminAccountMaintenanceInput{}, err
	}

	locale := strings.TrimSpace(input.Locale)
	localeSpecified := locale != ""
	if locale == "" {
		locale = "en"
	}
	locale, err = requireSupportedLocale(locale)
	if err != nil {
		return normalizedAdminAccountMaintenanceInput{}, err
	}

	operator := strings.TrimSpace(input.Operator)
	if operator == "" {
		operator = "adminctl"
	}

	auditMethod := strings.ToUpper(strings.TrimSpace(input.AuditMethod))
	if auditMethod == "" {
		auditMethod = "CLI"
	}

	auditPath := strings.TrimSpace(input.AuditPath)
	if auditPath == "" {
		auditPath = "cmd/adminctl ensure-admin"
	}

	return normalizedAdminAccountMaintenanceInput{
		Email:             email,
		Username:          username,
		UsernameSpecified: usernameSpecified,
		Password:          password,
		Role:              role,
		RoleSpecified:     roleSpecified,
		FirstName:         strings.TrimSpace(input.FirstName),
		LastName:          strings.TrimSpace(input.LastName),
		Locale:            locale,
		LocaleSpecified:   localeSpecified,
		Operator:          operator,
		AuditMethod:       auditMethod,
		AuditPath:         auditPath,
	}, nil
}

func applyAdminAccountMaintenanceInput(account *user.User, input normalizedAdminAccountMaintenanceInput) error {
	creating := account.ID == 0
	account.Email = input.Email
	if creating || input.UsernameSpecified || strings.TrimSpace(account.Username) == "" {
		account.Username = input.Username
	}
	if creating || strings.TrimSpace(input.FirstName) != "" {
		account.FirstName = input.FirstName
	}
	if creating || strings.TrimSpace(input.LastName) != "" {
		account.LastName = input.LastName
	}
	if creating || input.RoleSpecified || !auth.IsBackofficeRole(account.Role) {
		account.Role = input.Role.String()
	}
	if creating || input.LocaleSpecified || strings.TrimSpace(account.Locale) == "" {
		account.Locale = input.Locale
	}
	account.Status = "active"
	account.DeletedAt = gorm.DeletedAt{}
	if err := account.HashPassword(input.Password); err != nil {
		return fmt.Errorf("hash admin account password: %w", err)
	}
	return nil
}

func validateAdminAccountUsername(username string) error {
	if len(username) < 3 || len(username) > 50 {
		return fmt.Errorf("%w: length must be between 3 and 50", ErrAdminAccountUsernameInvalid)
	}
	for _, r := range username {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return fmt.Errorf("%w: whitespace and control characters are not allowed", ErrAdminAccountUsernameInvalid)
		}
	}
	return nil
}

func validateAdminAccountPassword(password string) error {
	if len([]rune(password)) < 12 {
		return fmt.Errorf("%w: minimum length is 12", ErrAdminAccountWeakPassword)
	}

	var lower, upper, digit, symbol bool
	for _, r := range password {
		switch {
		case unicode.IsLower(r):
			lower = true
		case unicode.IsUpper(r):
			upper = true
		case unicode.IsDigit(r):
			digit = true
		default:
			symbol = true
		}
	}

	classes := 0
	for _, ok := range []bool{lower, upper, digit, symbol} {
		if ok {
			classes++
		}
	}
	if classes < 3 {
		return fmt.Errorf("%w: include at least three character classes", ErrAdminAccountWeakPassword)
	}

	normalized := strings.ToLower(password)
	if normalized == "admin123456!" || strings.Contains(normalized, "password") {
		return fmt.Errorf("%w: choose a less common secret", ErrAdminAccountWeakPassword)
	}
	return nil
}

func deriveAdminUsername(email string) string {
	localPart := strings.Split(email, "@")[0]
	var b strings.Builder
	for _, r := range strings.ToLower(localPart) {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '.', r == '_', r == '-':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}

	username := strings.Trim(b.String(), ".-_")
	if len(username) < 3 {
		username = "admin"
	}
	if len(username) > 50 {
		username = username[:50]
	}
	return username
}

func ensureAdminAccountUsernameAvailable(tx *gorm.DB, username string, currentUserID uint) error {
	query := tx.Unscoped().Where("username = ?", username)
	if currentUserID != 0 {
		query = query.Where("id <> ?", currentUserID)
	}

	var existing user.User
	err := query.First(&existing).Error
	switch {
	case err == nil:
		return ErrUsernameExists
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil
	default:
		return fmt.Errorf("check admin account username: %w", err)
	}
}

func createAdminAccountMaintenanceAuditLog(tx *gorm.DB, account user.User, input normalizedAdminAccountMaintenanceInput, action, oldValue, newValue string) error {
	entry := audit.AuditLog{
		UserID:     account.ID,
		Username:   input.Operator,
		Action:     action,
		Resource:   "user",
		ResourceID: account.ID,
		Method:     input.AuditMethod,
		Path:       input.AuditPath,
		Changes:    `{"password":"rotated","status":"active"}`,
		OldValue:   oldValue,
		NewValue:   newValue,
		Status:     "success",
		CreatedAt:  time.Now().UTC(),
	}
	if err := tx.Create(&entry).Error; err != nil {
		return fmt.Errorf("create admin account audit log: %w", err)
	}
	return nil
}

func marshalAdminAccountSnapshot(account user.User) string {
	snapshot := adminAccountSnapshot{
		ID:       account.ID,
		Email:    account.Email,
		Username: account.Username,
		Role:     account.Role,
		Status:   account.Status,
		Locale:   account.Locale,
		Deleted:  account.DeletedAt.Valid,
	}
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return "{}"
	}
	return string(payload)
}

func adminAccountMaintenanceResult(account user.User, created bool) AdminAccountMaintenanceResult {
	return AdminAccountMaintenanceResult{
		UserID:   account.ID,
		Email:    account.Email,
		Username: account.Username,
		Role:     account.Role,
		Status:   account.Status,
		Created:  created,
	}
}
