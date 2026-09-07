package orderevidence

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrOrderEvidencePackageLocked     = errors.New("order evidence package is locked")
	ErrOrderEvidencePackageIncomplete = errors.New("order evidence package is incomplete")
	ErrOrderEvidencePackageSuperseded = errors.New("order evidence package is superseded")
	ErrOrderEvidenceRevisionSource    = errors.New("order evidence revision requires a locked package")
)

// EvidenceCompleteness is the package-level readiness calculation. A waived
// item satisfies the package requirement but remains distinguishable from
// captured evidence for audit and UI purposes.
type EvidenceCompleteness struct {
	Total     int  `json:"total"`
	Complete  int  `json:"complete"`
	Waived    int  `json:"waived"`
	Pending   int  `json:"pending"`
	Satisfied int  `json:"satisfied"`
	Percent   int  `json:"percent"`
	Ready     bool `json:"ready"`
}

func EvaluateEvidenceCompleteness(items []OrderEvidenceItem) EvidenceCompleteness {
	result := EvidenceCompleteness{Total: len(items)}
	for _, item := range items {
		switch strings.ToLower(strings.TrimSpace(item.Status)) {
		case EvidenceItemStatusComplete:
			result.Complete++
		case EvidenceItemStatusWaived:
			result.Waived++
		default:
			result.Pending++
		}
	}
	result.Satisfied = result.Complete + result.Waived
	if result.Total > 0 {
		result.Percent = result.Satisfied * 100 / result.Total
		result.Ready = result.Satisfied == result.Total
	}
	return result
}

func (p OrderEvidencePackage) IsLocked() bool {
	return strings.EqualFold(strings.TrimSpace(p.Status), PackageStatusLocked)
}

func (p OrderEvidencePackage) IsSuperseded() bool {
	return strings.EqualFold(strings.TrimSpace(p.Status), PackageStatusSuperseded)
}

func (p OrderEvidencePackage) EnsureMutable() error {
	switch strings.ToLower(strings.TrimSpace(p.Status)) {
	case PackageStatusLocked:
		return ErrOrderEvidencePackageLocked
	case PackageStatusSuperseded:
		return ErrOrderEvidencePackageSuperseded
	case PackageStatusIncomplete, PackageStatusReady:
		return nil
	default:
		return fmt.Errorf("unsupported order evidence package status %q", p.Status)
	}
}

func (p OrderEvidencePackage) EnsureRevisionSource() error {
	if p.IsLocked() {
		return nil
	}
	return ErrOrderEvidenceRevisionSource
}
