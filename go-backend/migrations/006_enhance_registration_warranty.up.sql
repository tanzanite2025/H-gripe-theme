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

-- 插入示例产品注册数据（假设已有产品和用户）
-- 注意：这里使用的 user_id 和 product_id 需要根据实际数据调整
INSERT INTO product_registrations (user_id, product_id, serial_number, purchase_date, retailer, warranty_period, warranty_expires, status, notes) VALUES
(1, 1, 'SN2024001001', '2024-01-15 10:00:00', 'Official Store', 24, '2026-01-15 10:00:00', 'active', 'Purchased during new year sale'),
(1, 2, 'SN2024001002', '2024-02-20 14:30:00', 'Amazon', 12, '2025-02-20 14:30:00', 'active', 'Gift from friend'),
(1, 3, 'SN2024001003', '2023-06-10 09:15:00', 'Best Buy', 12, '2024-06-10 09:15:00', 'expired', 'Warranty expired'),
(1, 1, 'SN2024001004', '2024-03-05 16:45:00', 'Official Store', 24, '2026-03-05 16:45:00', 'active', 'Extended warranty purchased'),
(1, 2, 'SN2024001005', '2024-04-12 11:20:00', 'Walmart', 12, '2025-04-12 11:20:00', 'active', 'Regular purchase')
ON CONFLICT (serial_number) DO NOTHING;

-- 插入示例保修申请数据
INSERT INTO warranty_claims (registration_id, user_id, issue_type, description, status) VALUES
(1, 1, 'defect', 'Product has a manufacturing defect in the display', 'submitted'),
(2, 1, 'malfunction', 'Device not turning on after 3 months of use', 'reviewing'),
(3, 1, 'damage', 'Screen cracked during normal use', 'rejected'),
(4, 1, 'defect', 'Battery draining too fast', 'approved'),
(5, 1, 'malfunction', 'Buttons not responding properly', 'completed')
ON CONFLICT DO NOTHING;

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
