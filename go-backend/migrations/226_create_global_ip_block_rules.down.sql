DROP INDEX IF EXISTS idx_global_ip_block_rules_created_at_live;
DROP INDEX IF EXISTS idx_global_ip_block_rules_source_live;
DROP INDEX IF EXISTS idx_global_ip_block_rules_active_live;
DROP INDEX IF EXISTS idx_global_ip_block_rules_cidr_live;
DROP TABLE IF EXISTS global_ip_block_rules;
DROP INDEX IF EXISTS idx_visitor_profiles_ip_address_live;
ALTER TABLE visitor_profiles DROP COLUMN IF EXISTS ip_address;
