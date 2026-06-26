package apierror

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
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
	log.Printf("[CRITICAL] Unhandled error: %v", err)
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": ErrInternal.Message,
		"code":  ErrInternal.Code,
	})
}
