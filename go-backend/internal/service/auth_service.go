package service

import (
	"errors"

	"tanzanite/internal/domain/user"
	"tanzanite/internal/pkg/config"
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

// GetUserByID 根据ID获取用户
func (s *AuthService) GetUserByID(id uint) (*user.User, error) {
	return s.userRepo.FindByID(id)
}
