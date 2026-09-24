package migrations

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test328OrderEvidenceAmountMinorMigrationIsCanonical(t *testing.T) {
	root := filepath.Join("328_order_evidence_amount_minor.up.sql")
	contents, err := os.ReadFile(root)
	require.NoError(t, err)
	sql := strings.ToLower(string(contents))
	for _, fragment := range []string{
		"add column if not exists order_total_amount_minor bigint",
		"add column if not exists order_total_usd_minor bigint",
		"add column if not exists order_total_usd_snapshot_minor bigint",
		"drop column if exists order_total_amount",
		"drop column if exists order_total_usd",
		"drop column if exists order_total_usd_snapshot",
	} {
		require.Contains(t, sql, fragment)
	}
}
