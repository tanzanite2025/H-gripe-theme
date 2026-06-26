package apierror

import "github.com/gin-gonic/gin"

// APIError 标准API错误结构
type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// 预定义错误代码常量
const (
	ErrCodeBadRequest   = "bad_request"
	ErrCodeUnauthorized = "unauthorized"
	ErrCodeForbidden    = "forbidden"
	ErrCodeNotFound     = "not_found"
	ErrCodeConflict     = "conflict"
	ErrCodeInternal     = "internal_error"
	ErrCodeValidation   = "validation_error"
)

// RespondError 通用错误响应
func RespondError(c *gin.Context, status int, code string, message string) {
	c.JSON(status, APIError{
		Code:    code,
		Message: message,
	})
}

// RespondErrorWithDetails 带详情的错误响应
func RespondErrorWithDetails(c *gin.Context, status int, code string, message string, details interface{}) {
	c.JSON(status, APIError{
		Code:    code,
		Message: message,
		Details: details,
	})
}

// RespondBadRequest 400 错误请求
func RespondBadRequest(c *gin.Context, message string) {
	RespondError(c, 400, ErrCodeBadRequest, message)
}

// RespondUnauthorized 401 未认证
func RespondUnauthorized(c *gin.Context) {
	RespondError(c, 401, ErrCodeUnauthorized, "Unauthorized")
}

// RespondForbidden 403 无权限
func RespondForbidden(c *gin.Context) {
	RespondError(c, 403, ErrCodeForbidden, "Forbidden")
}

// RespondNotFound 404 资源不存在
func RespondNotFound(c *gin.Context, resource string) {
	RespondError(c, 404, ErrCodeNotFound, resource+" not found")
}

// RespondConflict 409 资源冲突
func RespondConflict(c *gin.Context, message string) {
	RespondError(c, 409, ErrCodeConflict, message)
}

// RespondInternalError 500 内部错误
func RespondInternalError(c *gin.Context, err error) {
	RespondError(c, 500, ErrCodeInternal, err.Error())
}

// RespondValidationError 400 验证错误（带详情）
func RespondValidationError(c *gin.Context, details interface{}) {
	RespondErrorWithDetails(c, 400, ErrCodeValidation, "Validation failed", details)
}
