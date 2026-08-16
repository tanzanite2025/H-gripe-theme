CREATE TABLE IF NOT EXISTS wheelset_fit_questionnaires (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(120) NOT NULL UNIQUE,
    product_category_slug VARCHAR(120) NOT NULL DEFAULT 'wheelset',
    source_locale VARCHAR(32) NOT NULL DEFAULT 'zh_cn',
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_wheelset_fit_questionnaires_slug
        CHECK (slug = 'wheelset-fit'),
    CONSTRAINT ck_wheelset_fit_questionnaires_product_category
        CHECK (product_category_slug = 'wheelset'),
    CONSTRAINT ck_wheelset_fit_questionnaires_source_locale
        CHECK (source_locale = 'zh_cn')
);

CREATE TABLE IF NOT EXISTS wheelset_fit_questionnaire_versions (
    id BIGSERIAL PRIMARY KEY,
    questionnaire_id BIGINT NOT NULL REFERENCES wheelset_fit_questionnaires(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    status VARCHAR(24) NOT NULL DEFAULT 'draft',
    published_at TIMESTAMPTZ,
    published_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_wheelset_fit_questionnaire_versions_status
        CHECK (status IN ('draft', 'published', 'archived')),
    CONSTRAINT ck_wheelset_fit_questionnaire_versions_number
        CHECK (version_number > 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_wheelset_fit_questionnaire_versions_number
    ON wheelset_fit_questionnaire_versions(questionnaire_id, version_number);

CREATE UNIQUE INDEX IF NOT EXISTS idx_wheelset_fit_questionnaire_versions_one_draft
    ON wheelset_fit_questionnaire_versions(questionnaire_id)
    WHERE status = 'draft';

CREATE UNIQUE INDEX IF NOT EXISTS idx_wheelset_fit_questionnaire_versions_one_published
    ON wheelset_fit_questionnaire_versions(questionnaire_id)
    WHERE status = 'published';

CREATE INDEX IF NOT EXISTS idx_wheelset_fit_questionnaire_versions_status
    ON wheelset_fit_questionnaire_versions(questionnaire_id, status, version_number DESC);

CREATE TABLE IF NOT EXISTS wheelset_fit_questions (
    id BIGSERIAL PRIMARY KEY,
    questionnaire_version_id BIGINT NOT NULL REFERENCES wheelset_fit_questionnaire_versions(id) ON DELETE CASCADE,
    question_key VARCHAR(120) NOT NULL,
    answer_key VARCHAR(120) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 10,
    input_mode VARCHAR(32) NOT NULL DEFAULT 'single_choice',
    is_required BOOLEAN NOT NULL DEFAULT TRUE,
    allow_unknown BOOLEAN NOT NULL DEFAULT TRUE,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    source_revision INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_wheelset_fit_questions_key
        CHECK (question_key ~ '^[a-z0-9]+(_[a-z0-9]+)*$'),
    CONSTRAINT ck_wheelset_fit_questions_answer_key
        CHECK (answer_key ~ '^[a-z0-9]+(_[a-z0-9]+)*$'),
    CONSTRAINT ck_wheelset_fit_questions_input_mode
        CHECK (input_mode = 'single_choice'),
    CONSTRAINT ck_wheelset_fit_questions_sort_order
        CHECK (sort_order > 0),
    CONSTRAINT ck_wheelset_fit_questions_source_revision
        CHECK (source_revision > 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_wheelset_fit_questions_version_key
    ON wheelset_fit_questions(questionnaire_version_id, question_key);

CREATE UNIQUE INDEX IF NOT EXISTS idx_wheelset_fit_questions_version_sort_order
    ON wheelset_fit_questions(questionnaire_version_id, sort_order);

CREATE INDEX IF NOT EXISTS idx_wheelset_fit_questions_version_order
    ON wheelset_fit_questions(questionnaire_version_id, sort_order, id);

CREATE TABLE IF NOT EXISTS wheelset_fit_question_options (
    id BIGSERIAL PRIMARY KEY,
    question_id BIGINT NOT NULL REFERENCES wheelset_fit_questions(id) ON DELETE CASCADE,
    option_key VARCHAR(120) NOT NULL,
    answer_value VARCHAR(160) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 10,
    is_unknown BOOLEAN NOT NULL DEFAULT FALSE,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    product_filter_effects JSONB NOT NULL DEFAULT '{}'::jsonb,
    source_revision INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_wheelset_fit_question_options_key
        CHECK (option_key ~ '^[a-z0-9]+(_[a-z0-9]+)*$'),
    CONSTRAINT ck_wheelset_fit_question_options_sort_order
        CHECK (sort_order > 0),
    CONSTRAINT ck_wheelset_fit_question_options_source_revision
        CHECK (source_revision > 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_wheelset_fit_question_options_question_key
    ON wheelset_fit_question_options(question_id, option_key);

CREATE UNIQUE INDEX IF NOT EXISTS idx_wheelset_fit_question_options_question_sort_order
    ON wheelset_fit_question_options(question_id, sort_order);

CREATE INDEX IF NOT EXISTS idx_wheelset_fit_question_options_question_order
    ON wheelset_fit_question_options(question_id, sort_order, id);

CREATE TABLE IF NOT EXISTS wheelset_fit_question_translations (
    id BIGSERIAL PRIMARY KEY,
    question_id BIGINT NOT NULL REFERENCES wheelset_fit_questions(id) ON DELETE CASCADE,
    locale VARCHAR(32) NOT NULL,
    prompt TEXT NOT NULL DEFAULT '',
    help_title TEXT NOT NULL DEFAULT '',
    help_body TEXT NOT NULL DEFAULT '',
    source_revision INTEGER NOT NULL DEFAULT 1,
    translated_revision INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_wheelset_fit_question_translations_source_revision
        CHECK (source_revision > 0),
    CONSTRAINT ck_wheelset_fit_question_translations_translated_revision
        CHECK (translated_revision >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_wheelset_fit_question_translations_question_locale
    ON wheelset_fit_question_translations(question_id, locale);

CREATE INDEX IF NOT EXISTS idx_wheelset_fit_question_translations_locale
    ON wheelset_fit_question_translations(locale, question_id);

CREATE TABLE IF NOT EXISTS wheelset_fit_question_option_translations (
    id BIGSERIAL PRIMARY KEY,
    option_id BIGINT NOT NULL REFERENCES wheelset_fit_question_options(id) ON DELETE CASCADE,
    locale VARCHAR(32) NOT NULL,
    label TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    source_revision INTEGER NOT NULL DEFAULT 1,
    translated_revision INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_wheelset_fit_question_option_translations_source_revision
        CHECK (source_revision > 0),
    CONSTRAINT ck_wheelset_fit_question_option_translations_translated_revision
        CHECK (translated_revision >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_wheelset_fit_question_option_translations_option_locale
    ON wheelset_fit_question_option_translations(option_id, locale);

CREATE INDEX IF NOT EXISTS idx_wheelset_fit_question_option_translations_locale
    ON wheelset_fit_question_option_translations(locale, option_id);

INSERT INTO wheelset_fit_questionnaires (
    slug,
    product_category_slug,
    source_locale,
    is_enabled
)
VALUES (
    'wheelset-fit',
    'wheelset',
    'zh_cn',
    TRUE
)
ON CONFLICT (slug) DO UPDATE SET
    product_category_slug = 'wheelset',
    source_locale = 'zh_cn',
    updated_at = NOW();

INSERT INTO wheelset_fit_questionnaire_versions (
    questionnaire_id,
    version_number,
    status
)
SELECT
    questionnaire.id,
    1,
    'draft'
FROM wheelset_fit_questionnaires AS questionnaire
WHERE questionnaire.slug = 'wheelset-fit'
  AND NOT EXISTS (
      SELECT 1
      FROM wheelset_fit_questionnaire_versions AS version
      WHERE version.questionnaire_id = questionnaire.id
  );
