-- 增强 settings 表
-- 添加 is_public 和 description 字段

-- 添加 is_public 字段（如果不存在）
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'settings' AND column_name = 'is_public'
    ) THEN
        ALTER TABLE settings ADD COLUMN is_public BOOLEAN DEFAULT true;
    END IF;
END $$;

-- 添加 description 字段（如果不存在）
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'settings' AND column_name = 'description'
    ) THEN
        ALTER TABLE settings ADD COLUMN description TEXT;
    END IF;
END $$;

-- 创建索引（如果不存在）
CREATE INDEX IF NOT EXISTS idx_settings_group ON settings("group");
CREATE INDEX IF NOT EXISTS idx_settings_is_public ON settings(is_public);

-- 插入默认设置（如果不存在）

-- 站点设置
INSERT INTO settings (key, value, type, locale, "group", is_public, description, created_at, updated_at)
VALUES 
    ('site_name', 'Tanzanite', 'string', 'en', 'site', true, 'Site name', NOW(), NOW()),
    ('site_description', 'Premium E-commerce Platform', 'string', 'en', 'site', true, 'Site description', NOW(), NOW()),
    ('site_logo', '/images/logo.png', 'string', 'en', 'site', true, 'Site logo URL', NOW(), NOW()),
    ('contact_email', 'contact@tanzanite.site', 'string', 'en', 'site', true, 'Contact email', NOW(), NOW()),
    ('contact_phone', '+1-234-567-8900', 'string', 'en', 'site', true, 'Contact phone', NOW(), NOW())
ON CONFLICT (key, locale) DO NOTHING;

-- 快速购买设置
INSERT INTO settings (key, value, type, locale, "group", is_public, description, created_at, updated_at)
VALUES 
    ('enabled', 'true', 'boolean', 'en', 'quick-buy', true, 'Enable quick buy feature', NOW(), NOW()),
    ('button_text', 'Quick Buy', 'string', 'en', 'quick-buy', true, 'Quick buy button text', NOW(), NOW()),
    ('success_message', 'Added to cart successfully!', 'string', 'en', 'quick-buy', true, 'Success message', NOW(), NOW()),
    ('require_login', 'false', 'boolean', 'en', 'quick-buy', true, 'Require login for quick buy', NOW(), NOW())
ON CONFLICT (key, locale) DO NOTHING;

-- SEO 设置
INSERT INTO settings (key, value, type, locale, "group", is_public, description, created_at, updated_at)
VALUES 
    ('meta_title', 'Tanzanite - Premium E-commerce', 'string', 'en', 'seo', true, 'Default meta title', NOW(), NOW()),
    ('meta_description', 'Shop premium products at Tanzanite', 'string', 'en', 'seo', true, 'Default meta description', NOW(), NOW()),
    ('meta_keywords', 'ecommerce, shop, products', 'string', 'en', 'seo', true, 'Default meta keywords', NOW(), NOW()),
    ('google_analytics', '', 'string', 'en', 'seo', true, 'Google Analytics ID', NOW(), NOW()),
    ('google_tag_manager', '', 'string', 'en', 'seo', true, 'Google Tag Manager ID', NOW(), NOW())
ON CONFLICT (key, locale) DO NOTHING;

-- 社交媒体设置
INSERT INTO settings (key, value, type, locale, "group", is_public, description, created_at, updated_at)
VALUES 
    ('facebook', '', 'string', 'en', 'social', true, 'Facebook page URL', NOW(), NOW()),
    ('twitter', '', 'string', 'en', 'social', true, 'Twitter profile URL', NOW(), NOW()),
    ('instagram', '', 'string', 'en', 'social', true, 'Instagram profile URL', NOW(), NOW()),
    ('linkedin', '', 'string', 'en', 'social', true, 'LinkedIn profile URL', NOW(), NOW()),
    ('youtube', '', 'string', 'en', 'social', true, 'YouTube channel URL', NOW(), NOW()),
    ('wechat', '', 'string', 'en', 'social', true, 'WeChat ID', NOW(), NOW())
ON CONFLICT (key, locale) DO NOTHING;

-- 邮件设置（敏感信息，不公开）
INSERT INTO settings (key, value, type, locale, "group", is_public, description, created_at, updated_at)
VALUES 
    ('smtp_host', 'smtp.example.com', 'string', 'en', 'email', false, 'SMTP server host', NOW(), NOW()),
    ('smtp_port', '587', 'number', 'en', 'email', false, 'SMTP server port', NOW(), NOW()),
    ('smtp_username', '', 'string', 'en', 'email', false, 'SMTP username', NOW(), NOW()),
    ('smtp_password', '', 'string', 'en', 'email', false, 'SMTP password (encrypted)', NOW(), NOW()),
    ('from_email', 'noreply@tanzanite.site', 'string', 'en', 'email', false, 'From email address', NOW(), NOW()),
    ('from_name', 'Tanzanite', 'string', 'en', 'email', false, 'From name', NOW(), NOW())
ON CONFLICT (key, locale) DO NOTHING;

-- 验证查询
SELECT 
    'Settings table enhanced successfully' as message,
    COUNT(*) as total_settings,
    COUNT(DISTINCT "group") as total_groups
FROM settings;
