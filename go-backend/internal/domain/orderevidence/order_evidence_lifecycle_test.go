package orderevidence

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvaluateEvidenceCompletenessCountsWaivedAsSatisfied(t *testing.T) {
	items := []OrderEvidenceItem{
		{Status: EvidenceItemStatusComplete},
		{Status: EvidenceItemStatusWaived},
		{Status: EvidenceItemStatusDraft},
		{Status: EvidenceItemStatusMissing},
	}

	result := EvaluateEvidenceCompleteness(items)

	assert.Equal(t, 4, result.Total)
	assert.Equal(t, 1, result.Complete)
	assert.Equal(t, 1, result.Waived)
	assert.Equal(t, 2, result.Pending)
	assert.Equal(t, 2, result.Satisfied)
	assert.Equal(t, 50, result.Percent)
	assert.False(t, result.Ready)
}

func TestOrderEvidencePackageMutableAndRevisionRules(t *testing.T) {
	incomplete := OrderEvidencePackage{Status: PackageStatusIncomplete}
	assert.NoError(t, incomplete.EnsureMutable())
	assert.ErrorIs(t, incomplete.EnsureRevisionSource(), ErrOrderEvidenceRevisionSource)

	ready := OrderEvidencePackage{Status: PackageStatusReady}
	assert.NoError(t, ready.EnsureMutable())
	assert.ErrorIs(t, ready.EnsureRevisionSource(), ErrOrderEvidenceRevisionSource)

	locked := OrderEvidencePackage{Status: PackageStatusLocked}
	assert.ErrorIs(t, locked.EnsureMutable(), ErrOrderEvidencePackageLocked)
	assert.NoError(t, locked.EnsureRevisionSource())

	superseded := OrderEvidencePackage{Status: PackageStatusSuperseded}
	assert.ErrorIs(t, superseded.EnsureMutable(), ErrOrderEvidencePackageSuperseded)
}
