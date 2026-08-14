ALTER TABLE quick_buy_steps
    ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS help_text TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS quick_buy_step_translations (
    id BIGSERIAL PRIMARY KEY,
    step_id BIGINT NOT NULL REFERENCES quick_buy_steps(id) ON DELETE CASCADE,
    locale VARCHAR(32) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    help_text TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_quick_buy_step_translations_locale
        CHECK (locale <> '')
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_quick_buy_step_translations_step_locale
    ON quick_buy_step_translations(step_id, locale);

CREATE INDEX IF NOT EXISTS idx_quick_buy_step_translations_locale
    ON quick_buy_step_translations(locale);

ALTER TABLE quick_buy_flow_translations
    RENAME COLUMN help_text TO selection_help_text;

ALTER TABLE quick_buy_flows
    RENAME COLUMN help_text TO selection_help_text;
