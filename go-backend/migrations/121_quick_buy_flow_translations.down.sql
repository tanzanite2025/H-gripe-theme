ALTER TABLE quick_buy_flows
    DROP COLUMN IF EXISTS selection_help_text;

DROP INDEX IF EXISTS idx_quick_buy_flow_translations_locale;
DROP INDEX IF EXISTS idx_quick_buy_flow_translations_flow_locale;

DROP TABLE IF EXISTS quick_buy_flow_translations;
