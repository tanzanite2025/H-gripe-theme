-- Historical migration 202. Keep stable after it has been applied.
-- This migration creates stable empty setting rows; runtime default text comes from generated Go/Nuxt fallbacks.
-- Do not put editable default prose in a historical migration.

INSERT INTO settings (key, value, type, locale, "group", is_public, description, created_at, updated_at)
SELECT seed.key, '', 'string', seed.locale, 'website_name', true, seed.description, NOW(), NOW()
FROM (
    VALUES
        ('website_name_status', 'status', 'en', 'Why this name page: status'),
        ('website_name_intro', 'intro', 'en', 'Why this name page: intro'),
        ('website_name_eyebrow', 'eyebrow', 'en', 'Why this name page: eyebrow'),
        ('website_name_title', 'title', 'en', 'Why this name page: title'),
        ('website_name_body', 'body', 'en', 'Why this name page: body'),
        ('website_name_note', 'note', 'en', 'Why this name page: note'),
        ('website_name_status', 'status', 'zh_cn', 'Why this name page: status'),
        ('website_name_intro', 'intro', 'zh_cn', 'Why this name page: intro'),
        ('website_name_eyebrow', 'eyebrow', 'zh_cn', 'Why this name page: eyebrow'),
        ('website_name_title', 'title', 'zh_cn', 'Why this name page: title'),
        ('website_name_body', 'body', 'zh_cn', 'Why this name page: body'),
        ('website_name_note', 'note', 'zh_cn', 'Why this name page: note')
) AS seed(key, legacy_key, locale, description)
WHERE NOT EXISTS (
    SELECT 1
    FROM settings AS legacy
    WHERE legacy."group" = 'website_name'
      AND legacy.key = seed.legacy_key
      AND legacy.locale = seed.locale
)
ON CONFLICT (key, locale) DO NOTHING;
