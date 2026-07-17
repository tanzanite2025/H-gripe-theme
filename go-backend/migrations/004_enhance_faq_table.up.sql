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

-- 验证查询
SELECT 
    'FAQ table enhanced successfully' as message,
    COUNT(*) as total_faqs,
    COUNT(DISTINCT category) as total_categories,
    COUNT(DISTINCT locale) as total_locales
FROM faqs;
