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

-- 插入示例图片库数据
INSERT INTO galleries (name, slug, description, cover_image, status) VALUES
('Product Showcase', 'product-showcase', 'Showcase of our premium products', 'https://example.com/images/showcase-cover.jpg', 'published'),
('Customer Gallery', 'customer-gallery', 'Photos from our happy customers', 'https://example.com/images/customer-cover.jpg', 'published'),
('Behind the Scenes', 'behind-the-scenes', 'A look at how we work', 'https://example.com/images/bts-cover.jpg', 'published'),
('Event Photos', 'event-photos', 'Photos from our recent events', 'https://example.com/images/event-cover.jpg', 'published'),
('Team Gallery', 'team-gallery', 'Meet our amazing team', 'https://example.com/images/team-cover.jpg', 'published')
ON CONFLICT (slug) DO NOTHING;

-- 插入示例图片数据
INSERT INTO gallery_images (gallery_id, url, thumbnail, title, description, alt, width, height, size, tags, "order") VALUES
-- Product Showcase 图片
(1, 'https://example.com/images/product1.jpg', 'https://example.com/images/product1-thumb.jpg', 'Premium Product 1', 'Our flagship product', 'Premium Product 1', 1200, 800, 245760, 'product,premium', 1),
(1, 'https://example.com/images/product2.jpg', 'https://example.com/images/product2-thumb.jpg', 'Premium Product 2', 'Best seller', 'Premium Product 2', 1200, 800, 256000, 'product,bestseller', 2),
(1, 'https://example.com/images/product3.jpg', 'https://example.com/images/product3-thumb.jpg', 'Premium Product 3', 'New arrival', 'Premium Product 3', 1200, 800, 230400, 'product,new', 3),

-- Customer Gallery 图片
(2, 'https://example.com/images/customer1.jpg', 'https://example.com/images/customer1-thumb.jpg', 'Happy Customer 1', 'Customer testimonial photo', 'Happy Customer 1', 800, 600, 153600, 'customer,testimonial', 1),
(2, 'https://example.com/images/customer2.jpg', 'https://example.com/images/customer2-thumb.jpg', 'Happy Customer 2', 'Customer using our product', 'Happy Customer 2', 800, 600, 160000, 'customer,usage', 2),

-- Behind the Scenes 图片
(3, 'https://example.com/images/bts1.jpg', 'https://example.com/images/bts1-thumb.jpg', 'Workshop', 'Our production workshop', 'Workshop', 1600, 900, 409600, 'workshop,production', 1),
(3, 'https://example.com/images/bts2.jpg', 'https://example.com/images/bts2-thumb.jpg', 'Quality Control', 'Quality control process', 'Quality Control', 1600, 900, 420000, 'quality,process', 2),

-- Event Photos 图片
(4, 'https://example.com/images/event1.jpg', 'https://example.com/images/event1-thumb.jpg', 'Product Launch', 'Our latest product launch event', 'Product Launch', 1920, 1080, 614400, 'event,launch', 1),
(4, 'https://example.com/images/event2.jpg', 'https://example.com/images/event2-thumb.jpg', 'Trade Show', 'At the international trade show', 'Trade Show', 1920, 1080, 640000, 'event,tradeshow', 2),

-- Team Gallery 图片
(5, 'https://example.com/images/team1.jpg', 'https://example.com/images/team1-thumb.jpg', 'Team Photo', 'Our amazing team', 'Team Photo', 2000, 1333, 716800, 'team,group', 1),
(5, 'https://example.com/images/team2.jpg', 'https://example.com/images/team2-thumb.jpg', 'Office Space', 'Our modern office', 'Office Space', 1600, 1200, 512000, 'team,office', 2)
ON CONFLICT DO NOTHING;

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
