package localbackup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
)

// Read the original whole-storefront target fingerprint, including absent-database and company-static rules.
func (cfg Configuration) targetFingerprint(ctx context.Context) (string, map[string]int64, error) {
	control, err := cfg.connectDatabase(ctx, "postgres")
	if err != nil {
		return "", nil, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(control.Close)
	var exists bool
	if err := control.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)`, cfg.DatabaseName).Scan(&exists); err != nil {
		return "", nil, err
	}
	fingerprint := "absent_database"
	tables := map[string]int64{}
	if exists {
		db, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
		if err != nil {
			return "", nil, err
		}
		tables, fingerprint, _, err = captureDatabaseInventory(ctx, db)
		_ = db.Close()
		if err != nil {
			return "", nil, err
		}
	}
	media, err := cfg.mediaFingerprint(ctx)
	if err != nil {
		return "", nil, err
	}
	identity := fingerprint + ":" + media
	if cfg.ProductStaticAssetsDirectory != "" {
		staticFingerprint, err := cfg.companyStaticReleaseFileFingerprint(ctx)
		if err != nil {
			return "", nil, err
		}
		identity += ":company-static:" + staticFingerprint
	}
	digest := sha256.Sum256([]byte(identity))
	return hex.EncodeToString(digest[:]), tables, nil
}
