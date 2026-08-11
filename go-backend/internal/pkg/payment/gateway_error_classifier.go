package payment

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"syscall"
)

// IsTransientGatewayNetworkOrServerError identifies failures that indicate a
// provider/network incident and are safe to include in gateway health metrics.
// Card declines, validation errors, 3DS decisions, and order mismatches do not
// match this classifier and therefore do not open the circuit.
func IsTransientGatewayNetworkOrServerError(err error) bool {
	if err == nil || errors.Is(err, context.Canceled) {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, io.EOF) ||
		errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.ENETUNREACH) ||
		errors.Is(err, syscall.EHOSTUNREACH) {
		return true
	}

	var networkError net.Error
	if errors.As(err, &networkError) && (networkError.Timeout() || networkError.Temporary()) {
		return true
	}

	message := strings.ToLower(strings.TrimSpace(err.Error()))
	if message == "" ||
		strings.Contains(message, "invalid payment request") ||
		strings.Contains(message, "invalid currency") ||
		strings.Contains(message, "order mismatch") ||
		strings.Contains(message, "insufficient funds") ||
		strings.Contains(message, "declined") ||
		strings.Contains(message, "3ds") {
		return false
	}

	for _, fragment := range []string{
		" 500",
		" 502",
		" 503",
		" 504",
		"status code: 500",
		"status code: 502",
		"status code: 503",
		"status code: 504",
		"bad gateway",
		"service unavailable",
		"gateway timeout",
		"connection refused",
		"connection reset",
		"network is unreachable",
		"host is unreachable",
		"no such host",
		"tls handshake timeout",
		"i/o timeout",
		"timed out",
		"temporary failure",
		"upstream error",
	} {
		if strings.Contains(message, fragment) {
			return true
		}
	}
	return false
}
