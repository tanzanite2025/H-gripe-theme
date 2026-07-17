-- 007_enhance_ticket_system.sql
-- 增强客服工单系统表

-- 确保 tickets 表存在并有所有必要字段
CREATE TABLE IF NOT EXISTS tickets (
    id SERIAL PRIMARY KEY,
    ticket_number VARCHAR(50) UNIQUE NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject VARCHAR(255) NOT NULL,
    category VARCHAR(50),
    priority VARCHAR(20) DEFAULT 'medium',
    status VARCHAR(50) DEFAULT 'open',
    assigned_to INTEGER REFERENCES users(id),
    tags VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP,
    closed_at TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 确保 ticket_messages 表存在并有所有必要字段
CREATE TABLE IF NOT EXISTS ticket_messages (
    id SERIAL PRIMARY KEY,
    ticket_id INTEGER NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_staff BOOLEAN DEFAULT FALSE,
    content TEXT NOT NULL,
    attachments TEXT,
    is_internal BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_tickets_ticket_number ON tickets(ticket_number);
CREATE INDEX IF NOT EXISTS idx_tickets_user_id ON tickets(user_id);
CREATE INDEX IF NOT EXISTS idx_tickets_category ON tickets(category);
CREATE INDEX IF NOT EXISTS idx_tickets_priority ON tickets(priority);
CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets(status);
CREATE INDEX IF NOT EXISTS idx_tickets_assigned_to ON tickets(assigned_to);
CREATE INDEX IF NOT EXISTS idx_tickets_deleted_at ON tickets(deleted_at);

CREATE INDEX IF NOT EXISTS idx_ticket_messages_ticket_id ON ticket_messages(ticket_id);
CREATE INDEX IF NOT EXISTS idx_ticket_messages_user_id ON ticket_messages(user_id);
CREATE INDEX IF NOT EXISTS idx_ticket_messages_is_staff ON ticket_messages(is_staff);

-- 插入示例工单数据
INSERT INTO tickets (ticket_number, user_id, subject, category, priority, status, tags) VALUES
('TK20240515001', 1, 'Order not received', 'order', 'high', 'open', 'shipping,urgent'),
('TK20240515002', 1, 'Product quality issue', 'product', 'medium', 'in_progress', 'quality'),
('TK20240515003', 1, 'Refund request', 'order', 'high', 'resolved', 'refund'),
('TK20240515004', 1, 'Account login problem', 'other', 'low', 'closed', 'account'),
('TK20240515005', 1, 'Shipping delay inquiry', 'shipping', 'medium', 'open', 'shipping')
ON CONFLICT (ticket_number) DO NOTHING;

-- 插入示例工单消息
INSERT INTO ticket_messages (ticket_id, user_id, is_staff, content) VALUES
(1, 1, FALSE, 'I placed an order 2 weeks ago but have not received it yet. Order number: #12345'),
(1, 2, TRUE, 'Thank you for contacting us. Let me check the shipping status for you.'),
(2, 1, FALSE, 'The product I received has a defect. Can I get a replacement?'),
(2, 2, TRUE, 'We apologize for the inconvenience. Please send us photos of the defect.'),
(3, 1, FALSE, 'I would like to request a refund for my recent purchase.'),
(3, 2, TRUE, 'Your refund request has been approved. It will be processed within 3-5 business days.'),
(4, 1, FALSE, 'I cannot log into my account. It says my password is incorrect.'),
(4, 2, TRUE, 'I have reset your password. Please check your email for the reset link.'),
(5, 1, FALSE, 'My order is showing as shipped but the tracking has not updated in 3 days.'),
(5, 2, TRUE, 'Let me contact the shipping carrier to get an update on your package.')
ON CONFLICT DO NOTHING;

-- 输出统计信息
DO $$
DECLARE
    total_tickets INTEGER;
    total_messages INTEGER;
    open_tickets INTEGER;
    resolved_tickets INTEGER;
BEGIN
    SELECT COUNT(*) INTO total_tickets FROM tickets;
    SELECT COUNT(*) INTO total_messages FROM ticket_messages;
    SELECT COUNT(*) INTO open_tickets FROM tickets WHERE status = 'open';
    SELECT COUNT(*) INTO resolved_tickets FROM tickets WHERE status = 'resolved';
    
    RAISE NOTICE 'Ticket system tables enhanced successfully';
    RAISE NOTICE 'total_tickets: %', total_tickets;
    RAISE NOTICE 'open_tickets: %', open_tickets;
    RAISE NOTICE 'resolved_tickets: %', resolved_tickets;
    RAISE NOTICE 'total_messages: %', total_messages;
END $$;
