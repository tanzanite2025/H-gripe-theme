ALTER TABLE shipping_fpx_api_configs
    DROP COLUMN IF EXISTS last_sync_preserved_enabled,
    DROP COLUMN IF EXISTS last_sync_updated,
    DROP COLUMN IF EXISTS last_sync_added,
    DROP COLUMN IF EXISTS last_sync_scanned;
