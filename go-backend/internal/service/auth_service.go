package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"tanzanite/internal/domain/auth"
	"tanzanite/internal/domain/user"
	"tanzanite/internal/pkg/config"

	"github.com/golang-jwt/jwt/v5"
)

type UserRepository interface {
	Create(u *user.User) error
	FindByEmail(email string) (*user.User, error)
	FindByUsername(username string) (*user.User, error)
	FindByID(id uint) (*user.User, error)
	Update(u *user.User) error
	Delete(id uint) error
	List(offset, limit int) ([]user.User, int64, error)
}

type AuthService struct {
	userRepo UserRepository
	jwtCfg   config.JWTConfig
	oauthCfg config.OAuthConfig
}

func NewAuthService(userRepo UserRepository, jwtCfg config.JWTConfig, oauthCfg ...config.OAuthConfig) *AuthService {
	oauthConfig := config.OAuthConfig{}
	if len(oauthCfg) > 0 {
		oauthConfig = oauthCfg[0]
	}
	return &AuthService{
		userRepo: userRepo,
		jwtCfg:   jwtCfg,
		oauthCfg: oauthConfig,
	}
}

// Claims JWT声明
type Claims struct {
	UserID   uint   `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
	TokenUse string `json:"token_use,omitempty"`
	jwt.RegisteredClaims
}

// Register 用户注册
func (s *AuthService) Register(email, username, password string) (*user.User, error) {
	// 检查邮箱是否已存在
	if _, err := s.userRepo.FindByEmail(email); err == nil {
		return nil, errors.New("email already exists")
	}

	// 检查用户名是否已存在
	if _, err := s.userRepo.FindByUsername(username); err == nil {
		return nil, errors.New("username already exists")
	}

	// 创建用户
	u := &user.User{
		Email:    email,
		Username: username,
		Role:     "user",
		Status:   "active",
	}

	if err := u.HashPassword(password); err != nil {
		return nil, err
	}

	if err := s.userRepo.Create(u); err != nil {
		return nil, err
	}

	return u, nil
}

// Login 用户登录
func (s *AuthService) Login(emailOrUsername, password string) (string, *user.User, error) {
	// 尝试通过邮箱查找
	u, err := s.userRepo.FindByEmail(emailOrUsername)
	if err != nil {
		// 尝试通过用户名查找
		u, err = s.userRepo.FindByUsername(emailOrUsername)
		if err != nil {
			return "", nil, errors.New("invalid credentials")
		}
	}

	// 验证密码
	if !u.CheckPassword(password) {
		return "", nil, errors.New("invalid credentials")
	}

	// 检查用户状态
	if u.Status != "active" {
		return "", nil, errors.New("user account is not active")
	}

	// 生成JWT
	token, err := s.GenerateToken(u)
	if err != nil {
		return "", nil, err
	}

	return token, u, nil
}

// GenerateToken 生成JWT令牌
type googleTokenInfo struct {
	Issuer        string `json:"iss"`
	Subject       string `json:"sub"`
	Audience      string `json:"aud"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
}

func (s *AuthService) LoginWithGoogle(ctx context.Context, idToken string) (string, *user.User, error) {
	idToken = strings.TrimSpace(idToken)
	if idToken == "" {
		return "", nil, errors.New("google id token is required")
	}
	if strings.TrimSpace(s.oauthCfg.GoogleClientID) == "" {
		return "", nil, errors.New("google login is not configured")
	}

	tokenInfo, err := s.verifyGoogleIDToken(ctx, idToken)
	if err != nil {
		return "", nil, err
	}

	existingUser, err := s.userRepo.FindByEmail(tokenInfo.Email)
	if err == nil {
		if existingUser.Status != "active" {
			return "", nil, errors.New("user account is not active")
		}
		token, err := s.GenerateToken(existingUser)
		if err != nil {
			return "", nil, err
		}
		return token, existingUser, nil
	}

	createdUser, err := s.createGoogleUser(tokenInfo)
	if err != nil {
		return "", nil, err
	}
	token, err := s.GenerateToken(createdUser)
	if err != nil {
		return "", nil, err
	}
	return token, createdUser, nil
}

func (s *AuthService) verifyGoogleIDToken(ctx context.Context, idToken string) (*googleTokenInfo, error) {
	endpoint := "https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(idToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to verify google id token: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid google id token")
	}

	var tokenInfo googleTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return nil, fmt.Errorf("failed to parse google token response: %w", err)
	}

	tokenInfo.Email = strings.ToLower(strings.TrimSpace(tokenInfo.Email))
	tokenInfo.Subject = strings.TrimSpace(tokenInfo.Subject)
	if tokenInfo.Audience != strings.TrimSpace(s.oauthCfg.GoogleClientID) {
		return nil, errors.New("google token audience mismatch")
	}
	if tokenInfo.Issuer != "accounts.google.com" && tokenInfo.Issuer != "https://accounts.google.com" {
		return nil, errors.New("google token issuer mismatch")
	}
	if tokenInfo.Email == "" || !strings.EqualFold(tokenInfo.EmailVerified, "true") {
		return nil, errors.New("google email is not verified")
	}
	if tokenInfo.Subject == "" {
		return nil, errors.New("google subject is missing")
	}
	return &tokenInfo, nil
}

func (s *AuthService) createGoogleUser(tokenInfo *googleTokenInfo) (*user.User, error) {
	password, err := randomPassword()
	if err != nil {
		return nil, err
	}

	createdUser := &user.User{
		Email:     tokenInfo.Email,
		Username:  s.googleUsername(tokenInfo.Email, tokenInfo.Subject),
		FirstName: strings.TrimSpace(tokenInfo.GivenName),
		LastName:  strings.TrimSpace(tokenInfo.FamilyName),
		Role:      string(auth.RoleUser),
		Status:    "active",
	}
	if createdUser.FirstName == "" && createdUser.LastName == "" {
		createdUser.FirstName, createdUser.LastName = splitGoogleName(tokenInfo.Name)
	}

	if err := createdUser.HashPassword(password); err != nil {
		return nil, err
	}
	if err := s.userRepo.Create(createdUser); err != nil {
		return nil, err
	}
	return createdUser, nil
}

func (s *AuthService) googleUsername(email string, subject string) string {
	emailPrefix := strings.Split(strings.ToLower(strings.TrimSpace(email)), "@")[0]
	base := sanitizeUsername(emailPrefix)
	if len(base) < 3 {
		base = "google_user"
	}

	shortSubject := subject
	if len(shortSubject) > 12 {
		shortSubject = shortSubject[:12]
	}
	candidates := []string{
		base,
		base + "_" + shortSubject,
		"google_" + shortSubject,
	}
	for _, candidate := range candidates {
		if _, err := s.userRepo.FindByUsername(candidate); err != nil {
			return candidate
		}
	}
	return fmt.Sprintf("google_%s_%d", shortSubject, time.Now().Unix())
}

func sanitizeUsername(value string) string {
	re := regexp.MustCompile(`[^a-z0-9_]+`)
	value = re.ReplaceAllString(strings.ToLower(strings.TrimSpace(value)), "_")
	value = strings.Trim(value, "_")
	if len(value) > 50 {
		value = value[:50]
	}
	return value
}

func splitGoogleName(name string) (string, string) {
	parts := strings.Fields(strings.TrimSpace(name))
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

func randomPassword() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func (s *AuthService) GenerateToken(u *user.User) (string, error) {
	claims := Claims{
		UserID:   u.ID,
		Email:    u.Email,
		Username: u.Username,
		Role:     string(auth.NormalizeRole(u.Role)),
		TokenUse: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtCfg.GetJWTExpireDuration())),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtCfg.Secret))
}

func (s *AuthService) GenerateRefreshToken(u *user.User) (string, error) {
	claims := Claims{
		UserID:   u.ID,
		Email:    u.Email,
		Username: u.Username,
		Role:     string(auth.NormalizeRole(u.Role)),
		TokenUse: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtCfg.GetRefreshExpireDuration())),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtCfg.Secret))
}

func (s *AuthService) AccessTokenMaxAgeSeconds() int {
	return int(s.jwtCfg.GetJWTExpireDuration().Seconds())
}

func (s *AuthService) RefreshTokenMaxAgeSeconds() int {
	return int(s.jwtCfg.GetRefreshExpireDuration().Seconds())
}

// ValidateToken 验证JWT令牌
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtCfg.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateActiveToken verifies the token and refreshes user state from storage.
func (s *AuthService) ValidateActiveToken(tokenString string) (*Claims, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenUse != "" && claims.TokenUse != "access" {
		return nil, errors.New("invalid access token")
	}

	currentUser, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, err
	}
	if currentUser.Status != "active" {
		return nil, errors.New("user account is not active")
	}

	claims.Email = currentUser.Email
	claims.Username = currentUser.Username
	claims.Role = string(auth.NormalizeRole(currentUser.Role))

	return claims, nil
}

func (s *AuthService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenUse != "refresh" {
		return nil, errors.New("invalid refresh token")
	}

	currentUser, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, err
	}
	if currentUser.Status != "active" {
		return nil, errors.New("user account is not active")
	}

	claims.Email = currentUser.Email
	claims.Username = currentUser.Username
	claims.Role = string(auth.NormalizeRole(currentUser.Role))

	return claims, nil
}

// GetUserByID 根据ID获取用户
func (s *AuthService) GetUserByID(id uint) (*user.User, error) {
	return s.userRepo.FindByID(id)
}
