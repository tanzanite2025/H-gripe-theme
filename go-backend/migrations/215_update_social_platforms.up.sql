-- Keep storefront social links aligned with the supported icon set.
-- Payment-related WeChat settings live outside the social settings group and are untouched.

DELETE FROM settings
WHERE "group" = 'social'
  AND key = 'twitter';

INSERT INTO settings (key, value, type, locale, "group", is_public, description, created_at, updated_at)
VALUES
    ('x', '', 'string', 'en', 'social', true, 'X profile URL', NOW(), NOW()),
    ('reddit', '', 'string', 'en', 'social', true, 'Reddit profile URL', NOW(), NOW())
ON CONFLICT (key, locale) DO NOTHING;

DELETE FROM settings
WHERE "group" = 'social'
  AND key IN ('linkedin', 'wechat');
