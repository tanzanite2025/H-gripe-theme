CREATE TABLE IF NOT EXISTS selection_configuration_keys (
    id BIGSERIAL PRIMARY KEY,
    kind VARCHAR(32) NOT NULL,
    code VARCHAR(120) NOT NULL,
    display_label TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 10,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_selection_configuration_keys_kind_code
    ON selection_configuration_keys(kind, code);

INSERT INTO selection_configuration_keys (kind, code, display_label, description, is_enabled, sort_order, created_at, updated_at)
SELECT DISTINCT
    'question_key',
    q.question_key,
    q.question_key,
    '',
    TRUE,
    10,
    NOW(),
    NOW()
FROM wheelset_fit_questions q
WHERE BTRIM(q.question_key) <> ''
ON CONFLICT (kind, code) DO NOTHING;

INSERT INTO selection_configuration_keys (kind, code, display_label, description, is_enabled, sort_order, created_at, updated_at)
SELECT DISTINCT
    'answer_key',
    q.answer_key,
    q.answer_key,
    '',
    TRUE,
    10,
    NOW(),
    NOW()
FROM wheelset_fit_questions q
WHERE BTRIM(q.answer_key) <> ''
ON CONFLICT (kind, code) DO NOTHING;
