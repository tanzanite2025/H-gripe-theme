-- Route reconciliation is a forward data repair. Rolling back only removes
-- the bookkeeping columns/indexes; it intentionally keeps repaired paths and
-- FAQ pages so a rollback cannot resurrect the known-broken route snapshot.

DROP INDEX IF EXISTS idx_faq_pages_route_key_locale_live;
DROP INDEX IF EXISTS idx_faq_pages_last_seen_at;
DROP INDEX IF EXISTS idx_faq_pages_route_status;
DROP INDEX IF EXISTS idx_faq_pages_manifest_version;
DROP INDEX IF EXISTS idx_faq_pages_route_key;

ALTER TABLE faq_pages DROP COLUMN IF EXISTS last_seen_at;
ALTER TABLE faq_pages DROP COLUMN IF EXISTS route_status;
ALTER TABLE faq_pages DROP COLUMN IF EXISTS manifest_version;
ALTER TABLE faq_pages DROP COLUMN IF EXISTS route_key;
