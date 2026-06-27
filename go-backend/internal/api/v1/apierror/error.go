package apierror

import (
	"errors"
	"fmt"
	"net/http"
	appLogger "tanzanite/internal/pkg/logger"
	"tanzanite/internal/pkg/requestctx"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AppError represents a unified application error
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"` // Not serialized to JSON
}

func (e *AppError) Error() string {
	return e.Message
}

// New creates a new AppError
func New(code string, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// Common errors
var (
	ErrInternal = &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    "An internal server error occurred",
		StatusCode: http.StatusInternalServerError,
	}
)

// Send handles error responses in a unified way.
// If err is an AppError, it returns the specific status and message.
// Otherwise, it logs the error and returns a generic 500.
func Send(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.StatusCode, gin.H{
			"error": appErr.Message,
			"code":  appErr.Code,
		})
		return
	}

	// For non-AppErrors, log it and return a generic 500 to hide internal details.
	logUnhandledError(c, err)
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": ErrInternal.Message,
		"code":  ErrInternal.Code,
	})
}

func logUnhandledError(c *gin.Context, err error) {
	fields := []zap.Field{
		zap.String("error_type", fmt.Sprintf("%T", err)),
		zap.Int("status", http.StatusInternalServerError),
	}

	if c != nil && c.Request != nil {
		fields = append(fields,
			zap.String("method", c.Request.Method),
			zap.String("path", safeRoutePath(c)),
		)
		if traceID, ok := requestctx.TraceID(c.Request.Context()); ok {
			fields = append(fields, zap.String("trace_id", traceID))
		}
	}

	appLogger.Error("unhandled API error", fields...)
}

func safeRoutePath(c *gin.Context) string {
	if route := c.FullPath(); route != "" {
		return route
	}
	if c.Request != nil && c.Request.URL != nil {
		return c.Request.URL.Path
	}
	return ""
}
