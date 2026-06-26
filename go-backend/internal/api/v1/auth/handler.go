package auth

import (
	"net/http"
	"tanzanite/internal/api/v1/apierror"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	authService *service.AuthService
}

func NewHandler(authService *service.AuthService) *Handler {
	return &Handler{
		authService: authService,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	EmailOrUsername string `json:"email_or_username" binding:"required"`
	Password        string `json:"password" binding:"required"`
}

// Register 用户注册
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.Send(c, apierror.New("BAD_REQUEST", "Invalid request payload", http.StatusBadRequest))
		return
	}

	user, err := h.authService.Register(req.Email, req.Username, req.Password)
	if err != nil {
		apierror.Send(c, apierror.New("BAD_REQUEST", err.Error(), http.StatusBadRequest))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "user registered successfully",
		"user":    user.ToResponse(),
	})
}

// Login 用户登录
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.Send(c, apierror.New("BAD_REQUEST", "Invalid request payload", http.StatusBadRequest))
		return
	}

	token, user, err := h.authService.Login(req.EmailOrUsername, req.Password)
	if err != nil {
		apierror.Send(c, apierror.New("UNAUTHORIZED", err.Error(), http.StatusUnauthorized))
		return
	}

	c.SetCookie("auth_token", token, 3600*24*7, "/", "", true, true)

	c.JSON(http.StatusOK, gin.H{
		"user":  user.ToResponse(),
	})
}

// GetProfile 获取当前用户信息
func (h *Handler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.Send(c, apierror.New("UNAUTHORIZED", "unauthorized", http.StatusUnauthorized))
		return
	}

	user, err := h.authService.GetUserByID(userID.(uint))
	if err != nil {
		apierror.Send(c, apierror.New("NOT_FOUND", "user not found", http.StatusNotFound))
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// Logout 用户登出
func (h *Handler) Logout(c *gin.Context) {
	// JWT是无状态的，客户端删除token即可
	c.SetCookie("auth_token", "", -1, "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}
