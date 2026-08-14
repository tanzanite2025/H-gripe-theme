CREATE TABLE IF NOT EXISTS quick_buy_flow_translations (
    id BIGSERIAL PRIMARY KEY,
    flow_id BIGINT NOT NULL REFERENCES quick_buy_flows(id) ON DELETE CASCADE,
    locale VARCHAR(32) NOT NULL,
    selection_help_text TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_quick_buy_flow_translations_locale
        CHECK (locale <> '')
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_quick_buy_flow_translations_flow_locale
    ON quick_buy_flow_translations(flow_id, locale);

CREATE INDEX IF NOT EXISTS idx_quick_buy_flow_translations_locale
    ON quick_buy_flow_translations(locale);

ALTER TABLE quick_buy_flows
    ADD COLUMN IF NOT EXISTS selection_help_text TEXT NOT NULL DEFAULT '';
