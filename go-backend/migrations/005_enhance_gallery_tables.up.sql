-- 005_enhance_gallery_tables.sql
-- 增强图片库表结构

-- 确保 galleries 表存在并有所有必要字段
CREATE TABLE IF NOT EXISTS galleries (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    cover_image VARCHAR(500),
    view_count INTEGER DEFAULT 0,
    status VARCHAR(50) DEFAULT 'published',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 确保 gallery_images 表存在并有所有必要字段
CREATE TABLE IF NOT EXISTS gallery_images (
    id SERIAL PRIMARY KEY,
    gallery_id INTEGER NOT NULL REFERENCES galleries(id) ON DELETE CASCADE,
    url VARCHAR(500) NOT NULL,
    thumbnail VARCHAR(500),
    title VARCHAR(255),
    description TEXT,
    alt VARCHAR(255),
    width INTEGER,
    height INTEGER,
    size BIGINT,
    tags VARCHAR(500),
    "order" INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_galleries_slug ON galleries(slug);
CREATE INDEX IF NOT EXISTS idx_galleries_status ON galleries(status);
CREATE INDEX IF NOT EXISTS idx_galleries_deleted_at ON galleries(deleted_at);

CREATE INDEX IF NOT EXISTS idx_gallery_images_gallery_id ON gallery_images(gallery_id);
CREATE INDEX IF NOT EXISTS idx_gallery_images_order ON gallery_images("order");
CREATE INDEX IF NOT EXISTS idx_gallery_images_deleted_at ON gallery_images(deleted_at);

-- 输出统计信息
DO $$
DECLARE
    total_galleries INTEGER;
    total_images INTEGER;
BEGIN
    SELECT COUNT(*) INTO total_galleries FROM galleries;
    SELECT COUNT(*) INTO total_images FROM gallery_images;
    
    RAISE NOTICE 'Gallery tables enhanced successfully';
    RAISE NOTICE 'total_galleries: %', total_galleries;
    RAISE NOTICE 'total_images: %', total_images;
END $$;
