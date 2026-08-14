ALTER TABLE ops_project_bindings
    ADD COLUMN IF NOT EXISTS quick_buy_rate_limit_policy TEXT NOT NULL DEFAULT '';

UPDATE ops_project_bindings
SET quick_buy_rate_limit_policy = '{"enabled":true,"ip_requests_per_minute":120,"ip_burst":40,"session_requests_per_minute":60,"session_burst":20,"edge_ip_requests_per_minute":240,"edge_ip_burst":80,"caddy_rate_limit_enabled":false}'
WHERE name = 'commerce-platform'
  AND environment = 'production'
  AND COALESCE(quick_buy_rate_limit_policy, '') = '';
