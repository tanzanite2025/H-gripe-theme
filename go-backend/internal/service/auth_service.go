package service

import (
	"errors"
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
}

func NewAuthService(userRepo UserRepository, jwtCfg config.JWTConfig) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		jwtCfg:   jwtCfg,
	}
}

// Claims JWT声明
type Claims struct {
	UserID   uint   `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
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
func (s *AuthService) GenerateToken(u *user.User) (string, error) {
	claims := Claims{
		UserID:   u.ID,
		Email:    u.Email,
		Username: u.Username,
		Role:     string(auth.NormalizeRole(u.Role)),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtCfg.GetJWTExpireDuration())),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtCfg.Secret))
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
