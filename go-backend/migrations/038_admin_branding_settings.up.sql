INSERT INTO settings (key, value, type, locale, "group", is_public, description, created_at, updated_at)
VALUES
    ('site_url', '', 'string', 'en', 'site', true, 'Public site URL', NOW(), NOW()),
    ('admin_brand_name', '', 'string', 'en', 'site', true, 'Admin brand name', NOW(), NOW()),
    ('admin_brand_initial', '', 'string', 'en', 'site', true, 'Admin brand initial', NOW(), NOW()),
    ('admin_panel_label', '', 'string', 'en', 'site', true, 'Admin panel label', NOW(), NOW()),
    ('admin_login_title', '', 'string', 'en', 'site', true, 'Admin login title', NOW(), NOW()),
    ('admin_footer_text', '', 'string', 'en', 'site', true, 'Admin login footer text', NOW(), NOW()),
    ('admin_html_title', '', 'string', 'en', 'site', true, 'Admin browser title', NOW(), NOW())
ON CONFLICT (key, locale) DO NOTHING;
