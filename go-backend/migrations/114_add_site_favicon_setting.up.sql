INSERT INTO settings (key, value, type, locale, "group", is_public, description, created_at, updated_at)
VALUES (
    'site_favicon',
    '',
    'string',
    'en',
    'site',
    true,
    'Browser favicon URL',
    NOW(),
    NOW()
)
ON CONFLICT (key, locale) DO NOTHING;
