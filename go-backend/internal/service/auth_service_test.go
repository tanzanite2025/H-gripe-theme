package service

import (
	"testing"
	"tanzanite/internal/domain/user"
	"tanzanite/internal/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository 模拟用户仓库
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(u *user.User) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *MockUserRepository) FindByEmail(email string) (*user.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) FindByUsername(username string) (*user.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(id uint) (*user.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) Update(u *user.User) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) List(offset, limit int) ([]user.User, int64, error) {
	args := m.Called(offset, limit)
	return args.Get(0).([]user.User), args.Get(1).(int64), args.Error(2)
}

func TestAuthService_Register(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtCfg := config.JWTConfig{
		Secret:             "test-secret",
		ExpireHours:        24,
		RefreshExpireHours: 168,
	}
	authService := NewAuthService(mockRepo, jwtCfg)

	// 测试成功注册
	mockRepo.On("FindByEmail", "test@example.com").Return(nil, assert.AnError)
	mockRepo.On("FindByUsername", "testuser").Return(nil, assert.AnError)
	mockRepo.On("Create", mock.AnythingOfType("*user.User")).Return(nil)

	u, err := authService.Register("test@example.com", "testuser", "password123")

	assert.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, "test@example.com", u.Email)
	assert.Equal(t, "testuser", u.Username)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtCfg := config.JWTConfig{
		Secret:             "test-secret",
		ExpireHours:        24,
		RefreshExpireHours: 168,
	}
	authService := NewAuthService(mockRepo, jwtCfg)

	// 创建测试用户
	testUser := &user.User{
		ID:       1,
		Email:    "test@example.com",
		Username: "testuser",
		Role:     "user",
		Status:   "active",
	}
	testUser.HashPassword("password123")

	// 测试成功登录
	mockRepo.On("FindByEmail", "test@example.com").Return(testUser, nil)

	token, u, err := authService.Login("test@example.com", "password123")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, testUser.Email, u.Email)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_ValidateToken(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtCfg := config.JWTConfig{
		Secret:             "test-secret",
		ExpireHours:        24,
		RefreshExpireHours: 168,
	}
	authService := NewAuthService(mockRepo, jwtCfg)

	// 创建测试用户
	testUser := &user.User{
		ID:       1,
		Email:    "test@example.com",
		Username: "testuser",
		Role:     "user",
	}

	// 生成token
	token, err := authService.GenerateToken(testUser)
	assert.NoError(t, err)

	// 验证token
	claims, err := authService.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, testUser.ID, claims.UserID)
	assert.Equal(t, testUser.Email, claims.Email)
}
