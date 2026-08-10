-- QUICK is configured by quick_buy_flows and quick_buy_flow_versions.
-- Remove the obsolete generic settings group so it cannot be edited as a
-- second source of truth.
DELETE FROM settings
WHERE "group" = 'quick-buy';
