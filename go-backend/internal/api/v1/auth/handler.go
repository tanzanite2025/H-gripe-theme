package auth

import (
	"errors"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/pkg/securecookie"
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

type GoogleLoginRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

// Register 用户注册
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	user, err := h.authService.Register(req.Email, req.Username, req.Password)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Created(c, gin.H{
		"message": "User registered successfully",
		"user":    user.ToResponse(),
	})
}

// Login 用户登录
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	token, user, err := h.authService.Login(req.EmailOrUsername, req.Password)
	if err != nil {
		apierror.RespondUnauthorized(c)
		return
	}

	securecookie.SetAuthToken(c, token, 3600*24*7)
	if _, err := securecookie.SetCSRFToken(c, 3600*24*7); err != nil {
		apierror.RespondInternalError(c, errors.New("failed to generate CSRF token"))
		return
	}

	response.Success(c, gin.H{
		"user": user.ToResponse(),
	})
}

// GetProfile 获取当前用户信息
func (h *Handler) GoogleLogin(c *gin.Context) {
	var req GoogleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	token, user, err := h.authService.LoginWithGoogle(c.Request.Context(), req.IDToken)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	securecookie.SetAuthToken(c, token, 3600*24*7)
	if _, err := securecookie.SetCSRFToken(c, 3600*24*7); err != nil {
		apierror.RespondInternalError(c, errors.New("failed to generate CSRF token"))
		return
	}

	response.Success(c, gin.H{
		"user": user.ToResponse(),
	})
}

func (h *Handler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	user, err := h.authService.GetUserByID(userID.(uint))
	if err != nil {
		apierror.RespondNotFound(c, "User")
		return
	}

	response.Success(c, user.ToResponse())
}

// Logout 用户登出
func (h *Handler) Logout(c *gin.Context) {
	// JWT是无状态的，客户端删除token即可
	securecookie.ClearAuthToken(c)
	securecookie.ClearCSRFToken(c)
	response.SuccessWithMessage(c, "Logged out successfully", nil)
}
