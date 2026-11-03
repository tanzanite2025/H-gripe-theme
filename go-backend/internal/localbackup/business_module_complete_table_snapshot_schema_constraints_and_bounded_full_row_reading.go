package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

// Share complete-table snapshot shapes, ordered column/constraint fingerprints and bounded full-row reads. Domain selectors and product-only row ordering remain with their modules.
const businessModuleSnapshotMaximumJSONBytes = 64 << 20

const businessModuleSnapshotMaximumRows = 100000

type businessModuleCompleteTableSnapshot struct {
	SchemaSHA256      string            `json:"schema_sha256"`
	ConstraintsSHA256 string            `json:"constraints_sha256"`
	Rows              []json.RawMessage `json:"rows"`
}

func inspectBusinessModuleCompleteTableSchema(ctx context.Context, tx *sql.Tx, table string) (string, []string, error) {
	rows, err := tx.QueryContext(ctx, `SELECT a.attname,format_type(a.atttypid,a.atttypmod),a.attnotnull,a.attidentity,a.attgenerated,COALESCE(pg_get_expr(d.adbin,d.adrelid),'') FROM pg_attribute a LEFT JOIN pg_attrdef d ON d.adrelid=a.attrelid AND d.adnum=a.attnum WHERE a.attrelid=to_regclass($1) AND a.attnum>0 AND NOT a.attisdropped ORDER BY a.attnum`, "public."+table)
	if err != nil {
		return "", nil, err
	}
	var columns []string
	var contract strings.Builder
	for rows.Next() {
		var name, typ, identity, generated, defaultValue string
		var required bool
		if err := rows.Scan(&name, &typ, &required, &identity, &generated, &defaultValue); err != nil {
			_ = rows.Close()
			return "", nil, err
		}
		if generated != "" || identity != "" {
			_ = rows.Close()
			return "", nil, errors.New("product module requires an explicit generated-column contract")
		}
		columns = append(columns, name)
		fmt.Fprintf(&contract, "%q:%q:%t:%q\n", name, typ, required, defaultValue)
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return "", nil, err
	}
	if len(columns) == 0 {
		return "", nil, fmt.Errorf("product table missing: %s", table)
	}
	return businessModuleRecoveryBytesSHA256([]byte(contract.String())), columns, nil
}

func captureBusinessModuleTableConstraintFingerprint(ctx context.Context, tx *sql.Tx, table string) (string, error) {
	var constraints string
	err := tx.QueryRowContext(ctx, `SELECT COALESCE(string_agg(pg_get_constraintdef(oid),E'\n' ORDER BY conname),'') FROM pg_constraint WHERE conrelid=to_regclass($1)`, "public."+table).Scan(&constraints)
	return businessModuleRecoveryBytesSHA256([]byte(constraints)), err
}

func readBusinessModuleCompleteTableSnapshot(ctx context.Context, tx *sql.Tx, table, predicate string, arguments ...any) (businessModuleCompleteTableSnapshot, error) {
	result := businessModuleCompleteTableSnapshot{Rows: []json.RawMessage{}}
	var err error
	result.SchemaSHA256, _, err = inspectBusinessModuleCompleteTableSchema(ctx, tx, table)
	if err != nil {
		return result, err
	}
	result.ConstraintsSHA256, err = captureBusinessModuleTableConstraintFingerprint(ctx, tx, table)
	if err != nil {
		return result, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT to_jsonb(r)::text FROM public.`+pq.QuoteIdentifier(table)+` r WHERE `+predicate+` ORDER BY to_jsonb(r)::text LIMIT 100001`, arguments...) // #nosec G202 -- Selectors are internal registry literals; data values are bound.
	if err != nil {
		return result, fmt.Errorf("无法保存模块完整记录：%s", table)
	}
	defer func(closeResource func() error) { _ = closeResource() }(rows.Close)
	totalBytes := 0
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return result, err
		}
		totalBytes += len(raw)
		if len(result.Rows) >= businessModuleSnapshotMaximumRows || totalBytes > businessModuleSnapshotMaximumJSONBytes {
			return result, errors.New("模块记录超过容量限制，未生成截断备份")
		}
		result.Rows = append(result.Rows, json.RawMessage(raw))
	}
	return result, rows.Err()
}
