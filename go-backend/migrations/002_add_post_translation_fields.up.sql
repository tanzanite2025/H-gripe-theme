-- 添加博客多语言支持字段
-- 迁移版本: 002
-- 创建时间: 2026-05-25

-- ============================================
-- 1. 添加翻译组ID字段
-- ============================================

-- 检查字段是否存在，如果不存在则添加
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'posts' AND column_name = 'translation_group_id'
    ) THEN
        ALTER TABLE posts ADD COLUMN translation_group_id INTEGER;
        RAISE NOTICE '已添加 translation_group_id 字段';
    ELSE
        RAISE NOTICE 'translation_group_id 字段已存在';
    END IF;
END $$;

-- 添加索引
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes 
        WHERE tablename = 'posts' AND indexname = 'idx_posts_translation_group_id'
    ) THEN
        CREATE INDEX idx_posts_translation_group_id ON posts(translation_group_id);
        RAISE NOTICE '已创建 translation_group_id 索引';
    ELSE
        RAISE NOTICE 'translation_group_id 索引已存在';
    END IF;
END $$;

-- ============================================
-- 2. 添加 SEO 元数据字段
-- ============================================

-- meta_keywords 字段
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'posts' AND column_name = 'meta_keywords'
    ) THEN
        ALTER TABLE posts ADD COLUMN meta_keywords VARCHAR(255);
        RAISE NOTICE '已添加 meta_keywords 字段';
    ELSE
        RAISE NOTICE 'meta_keywords 字段已存在';
    END IF;
END $$;

-- canonical_url 字段
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'posts' AND column_name = 'canonical_url'
    ) THEN
        ALTER TABLE posts ADD COLUMN canonical_url VARCHAR(500);
        RAISE NOTICE '已添加 canonical_url 字段';
    ELSE
        RAISE NOTICE 'canonical_url 字段已存在';
    END IF;
END $$;

-- ============================================
-- 3. 数据迁移（可选）
-- ============================================

-- 如果有 parent_id 字段，可以将其迁移到 translation_group_id
-- 注意：这个逻辑需要根据实际情况调整

-- 示例：将 parent_id 复制到 translation_group_id
-- UPDATE posts 
-- SET translation_group_id = parent_id
-- WHERE parent_id IS NOT NULL AND translation_group_id IS NULL;

-- 示例：为没有翻译组的文章设置自己的ID作为翻译组ID
-- UPDATE posts 
-- SET translation_group_id = id
-- WHERE translation_group_id IS NULL;

-- ============================================
-- 4. 验证
-- ============================================

-- 显示 posts 表的结构
SELECT 
    column_name, 
    data_type, 
    character_maximum_length,
    is_nullable
FROM information_schema.columns
WHERE table_name = 'posts'
ORDER BY ordinal_position;

-- 显示 posts 表的索引
SELECT 
    indexname,
    indexdef
FROM pg_indexes
WHERE tablename = 'posts'
ORDER BY indexname;

-- 统计翻译组数据
SELECT 
    COUNT(*) as total_posts,
    COUNT(DISTINCT translation_group_id) as total_groups,
    COUNT(CASE WHEN translation_group_id IS NOT NULL THEN 1 END) as posts_with_group
FROM posts;

-- 显示翻译组示例（前5个有多个语言版本的组）
SELECT 
    translation_group_id,
    COUNT(*) as language_count,
    STRING_AGG(locale, ', ') as locales,
    STRING_AGG(title, ' | ') as titles
FROM posts
WHERE translation_group_id IS NOT NULL
GROUP BY translation_group_id
HAVING COUNT(*) > 1
ORDER BY language_count DESC
LIMIT 5;
