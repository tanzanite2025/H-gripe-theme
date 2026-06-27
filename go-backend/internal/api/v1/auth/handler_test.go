package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"tanzanite/internal/domain/user"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/service"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository for testing
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

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestRegisterHandler(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtCfg := config.JWTConfig{
		Secret:             "test-secret",
		ExpireHours:        24,
		RefreshExpireHours: 168,
	}
	authService := service.NewAuthService(mockRepo, jwtCfg)
	handler := NewHandler(authService)

	router := setupTestRouter()
	router.POST("/register", handler.Register)

	// 准备请求
	reqBody := RegisterRequest{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	mockRepo.On("FindByEmail", "test@example.com").Return(nil, assert.AnError)
	mockRepo.On("FindByUsername", "testuser").Return(nil, assert.AnError)
	mockRepo.On("Create", mock.AnythingOfType("*user.User")).Return(nil)

	// 发送请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Created successfully", response["message"])
	assert.NotNil(t, response["data"])
	mockRepo.AssertExpectations(t)
}

func TestLoginHandler(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtCfg := config.JWTConfig{
		Secret:             "test-secret",
		ExpireHours:        24,
		RefreshExpireHours: 168,
	}
	authService := service.NewAuthService(mockRepo, jwtCfg)
	handler := NewHandler(authService)

	router := setupTestRouter()
	router.POST("/login", handler.Login)

	// 创建测试用户
	testUser := &user.User{
		ID:       1,
		Email:    "test@example.com",
		Username: "testuser",
		Role:     "user",
		Status:   "active",
	}
	testUser.HashPassword("password123")

	// 准备请求
	reqBody := LoginRequest{
		EmailOrUsername: "test@example.com",
		Password:        "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	mockRepo.On("FindByEmail", "test@example.com").Return(testUser, nil)

	// 发送请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotNil(t, response["data"])
	assert.NotEmpty(t, w.Result().Cookies())
	mockRepo.AssertExpectations(t)
}
