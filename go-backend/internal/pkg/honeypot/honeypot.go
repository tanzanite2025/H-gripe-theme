package honeypot

import (
	"strings"

	"commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/pkg/metrics"

	"go.uber.org/zap"
)

const (
	ModeOff     = "off"
	ModeShadow  = "shadow"
	ModeEnforce = "enforce"
)

// Policy controls whether honeypot hits are ignored, observed, or silently
// dropped. An empty or unknown mode fails safe to enforcement; configuration
// validation rejects unknown values before a production server starts.
type Policy struct {
	enabled bool
	enforce bool
}

func NewPolicy(mode string) Policy {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case ModeOff:
		return Policy{}
	case ModeShadow:
		return Policy{enabled: true}
	default:
		return Policy{enabled: true, enforce: true}
	}
}

// Filled reports whether a decoy field contains anything other than
// whitespace. Decoy fields are intentionally endpoint-specific; callers must
// decide the response shape and status code for the endpoint they protect.
func Filled(value string) bool {
	return strings.TrimSpace(value) != ""
}

// ShouldDrop observes a decoy value and returns whether the caller must stop
// before executing business side effects.
func (p Policy) ShouldDrop(value, form, field, route string) bool {
	if !p.enabled || !Filled(value) {
		return false
	}
	RecordBlocked(form, field, route)
	return p.enforce
}

// RecordBlocked records a honeypot hit without retaining request data that may
// contain personal information. The form and field values are static,
// reviewed identifiers rather than user-controlled labels.
func RecordBlocked(form, field, route string) {
	metrics.HoneypotBlocked.WithLabelValues(form, field).Inc()
	logger.Warn("[HONEYPOT_BLOCKED]",
		zap.String("form", form),
		zap.String("field", field),
		zap.String("route", route),
	)
}
