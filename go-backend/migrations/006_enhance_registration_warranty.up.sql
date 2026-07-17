-- 006_enhance_registration_warranty.sql
-- 增强产品注册和保修表

-- 确保 product_registrations 表存在并有所有必要字段
CREATE TABLE IF NOT EXISTS product_registrations (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    serial_number VARCHAR(255) UNIQUE NOT NULL,
    purchase_date TIMESTAMP NOT NULL,
    purchase_proof VARCHAR(500),
    retailer VARCHAR(255),
    warranty_period INTEGER NOT NULL DEFAULT 12,
    warranty_expires TIMESTAMP NOT NULL,
    status VARCHAR(50) DEFAULT 'active',
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 确保 warranty_claims 表存在并有所有必要字段
CREATE TABLE IF NOT EXISTS warranty_claims (
    id SERIAL PRIMARY KEY,
    registration_id INTEGER NOT NULL REFERENCES product_registrations(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    issue_type VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    images TEXT,
    status VARCHAR(50) DEFAULT 'submitted',
    resolution TEXT,
    processed_by INTEGER REFERENCES users(id),
    processed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_product_registrations_user_id ON product_registrations(user_id);
CREATE INDEX IF NOT EXISTS idx_product_registrations_product_id ON product_registrations(product_id);
CREATE INDEX IF NOT EXISTS idx_product_registrations_serial_number ON product_registrations(serial_number);
CREATE INDEX IF NOT EXISTS idx_product_registrations_status ON product_registrations(status);
CREATE INDEX IF NOT EXISTS idx_product_registrations_warranty_expires ON product_registrations(warranty_expires);
CREATE INDEX IF NOT EXISTS idx_product_registrations_deleted_at ON product_registrations(deleted_at);

CREATE INDEX IF NOT EXISTS idx_warranty_claims_registration_id ON warranty_claims(registration_id);
CREATE INDEX IF NOT EXISTS idx_warranty_claims_user_id ON warranty_claims(user_id);
CREATE INDEX IF NOT EXISTS idx_warranty_claims_status ON warranty_claims(status);
CREATE INDEX IF NOT EXISTS idx_warranty_claims_deleted_at ON warranty_claims(deleted_at);

-- 输出统计信息
DO $$
DECLARE
    total_registrations INTEGER;
    total_claims INTEGER;
    active_registrations INTEGER;
    expired_registrations INTEGER;
BEGIN
    SELECT COUNT(*) INTO total_registrations FROM product_registrations;
    SELECT COUNT(*) INTO total_claims FROM warranty_claims;
    SELECT COUNT(*) INTO active_registrations FROM product_registrations WHERE status = 'active';
    SELECT COUNT(*) INTO expired_registrations FROM product_registrations WHERE status = 'expired';
    
    RAISE NOTICE 'Product registration and warranty tables enhanced successfully';
    RAISE NOTICE 'total_registrations: %', total_registrations;
    RAISE NOTICE 'active_registrations: %', active_registrations;
    RAISE NOTICE 'expired_registrations: %', expired_registrations;
    RAISE NOTICE 'total_warranty_claims: %', total_claims;
END $$;
