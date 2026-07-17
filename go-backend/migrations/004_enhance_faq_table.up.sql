-- 增强 FAQ 表
-- 确保所有必要的字段和索引存在

-- 检查并添加缺失的字段
DO $$
BEGIN
    -- 添加 parent_id 字段（如果不存在）
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'faqs' AND column_name = 'parent_id'
    ) THEN
        ALTER TABLE faqs ADD COLUMN parent_id INTEGER;
    END IF;

    -- 添加 order 字段（如果不存在）
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'faqs' AND column_name = 'order'
    ) THEN
        ALTER TABLE faqs ADD COLUMN "order" INTEGER DEFAULT 0;
    END IF;

    -- 添加 status 字段（如果不存在）
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'faqs' AND column_name = 'status'
    ) THEN
        ALTER TABLE faqs ADD COLUMN status VARCHAR(20) DEFAULT 'published';
    END IF;

    -- 添加 view_count 字段（如果不存在）
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'faqs' AND column_name = 'view_count'
    ) THEN
        ALTER TABLE faqs ADD COLUMN view_count INTEGER DEFAULT 0;
    END IF;
END $$;

-- 创建索引（如果不存在）
CREATE INDEX IF NOT EXISTS idx_faqs_category ON faqs(category);
CREATE INDEX IF NOT EXISTS idx_faqs_locale ON faqs(locale);
CREATE INDEX IF NOT EXISTS idx_faqs_parent_id ON faqs(parent_id);
CREATE INDEX IF NOT EXISTS idx_faqs_status ON faqs(status);
CREATE INDEX IF NOT EXISTS idx_faqs_order ON faqs("order");
CREATE INDEX IF NOT EXISTS idx_faqs_view_count ON faqs(view_count);

-- 插入示例 FAQ 数据（如果表为空）
INSERT INTO faqs (question, answer, category, locale, "order", status, created_at, updated_at)
SELECT * FROM (VALUES
    ('What is your return policy?', 'We offer a 30-day return policy for all unused items in original packaging.', 'General', 'en', 1, 'published', NOW(), NOW()),
    ('How long does shipping take?', 'Standard shipping takes 5-7 business days. Express shipping is available for 2-3 business days.', 'Shipping', 'en', 2, 'published', NOW(), NOW()),
    ('Do you ship internationally?', 'Yes, we ship to most countries worldwide. Shipping costs and times vary by location.', 'Shipping', 'en', 3, 'published', NOW(), NOW()),
    ('How can I track my order?', 'Once your order ships, you will receive a tracking number via email.', 'Orders', 'en', 4, 'published', NOW(), NOW()),
    ('What payment methods do you accept?', 'We accept credit cards (Visa, MasterCard, Amex), PayPal, and bank transfers.', 'Payment', 'en', 5, 'published', NOW(), NOW()),
    ('Can I cancel my order?', 'Orders can be cancelled within 24 hours of placement. After that, please contact customer service.', 'Orders', 'en', 6, 'published', NOW(), NOW()),
    ('Do you offer warranty?', 'Yes, all products come with a 1-year manufacturer warranty.', 'Products', 'en', 7, 'published', NOW(), NOW()),
    ('How do I contact customer support?', 'You can reach us via email at support@tanzanite.site or through our contact form.', 'Support', 'en', 8, 'published', NOW(), NOW()),
    ('Are your products authentic?', 'Yes, we only sell 100% authentic products directly from manufacturers.', 'Products', 'en', 9, 'published', NOW(), NOW()),
    ('Can I change my shipping address?', 'Yes, if your order has not shipped yet. Please contact us immediately.', 'Shipping', 'en', 10, 'published', NOW(), NOW())
) AS v(question, answer, category, locale, "order", status, created_at, updated_at)
WHERE NOT EXISTS (SELECT 1 FROM faqs LIMIT 1);

-- 验证查询
SELECT 
    'FAQ table enhanced successfully' as message,
    COUNT(*) as total_faqs,
    COUNT(DISTINCT category) as total_categories,
    COUNT(DISTINCT locale) as total_locales
FROM faqs;
