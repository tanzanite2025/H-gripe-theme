INSERT INTO settings (key, value, type, locale, "group", is_public, description, created_at, updated_at)
VALUES
    ('brand_title', '', 'string', 'en', 'site', true, 'Top gradient brand title', NOW(), NOW())
ON CONFLICT (key, locale) DO NOTHING;
