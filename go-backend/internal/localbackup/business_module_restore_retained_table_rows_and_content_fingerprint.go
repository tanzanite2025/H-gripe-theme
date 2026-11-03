package localbackup

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/lib/pq"
)

// Preserve the existing retained-row count and aggregate comparison format. This fingerprint compares retained contents; it does not authenticate individual records.
type BusinessModuleRestoreTablePreservation struct {
	Table                      string `json:"table"`
	RetainedRecords            int64  `json:"retained_records"`
	ProposedReplacementRecords int    `json:"proposed_replacement_records"`
	ContentFingerprint         string `json:"content_fingerprint"`
}

func fingerprintBusinessModuleRetainedTargetTable(ctx context.Context, transaction *sql.Tx, table, shape string, excludedPrimaryKeys []json.RawMessage) (BusinessModuleRestoreTablePreservation, error) {
	result := BusinessModuleRestoreTablePreservation{Table: table, ProposedReplacementRecords: len(excludedPrimaryKeys)}
	if excludedPrimaryKeys == nil {
		excludedPrimaryKeys = []json.RawMessage{}
	}
	encoded, err := json.Marshal(excludedPrimaryKeys)
	if err != nil {
		return result, err
	}
	query := `SELECT count(*),coalesce(sum(('x'||substr(md5(to_jsonb(record)::text),1,15))::bit(60)::bigint)::text,'0'),coalesce(sum(('x'||substr(md5(to_jsonb(record)::text),17,15))::bit(60)::bigint)::text,'0') FROM public.` + pq.QuoteIdentifier(table) + ` record WHERE NOT EXISTS(SELECT 1 FROM jsonb_array_elements($1::jsonb) excluded(key) WHERE to_jsonb(record) @> excluded.key)`
	var first, second string
	if err := transaction.QueryRowContext(ctx, query, string(encoded)).Scan(&result.RetainedRecords, &first, &second); err != nil {
		return result, err
	}
	digest := sha256.Sum256([]byte(fmt.Sprintf("%s:%d:%s:%s", shape, result.RetainedRecords, first, second)))
	result.ContentFingerprint = hex.EncodeToString(digest[:])
	return result, nil
}
