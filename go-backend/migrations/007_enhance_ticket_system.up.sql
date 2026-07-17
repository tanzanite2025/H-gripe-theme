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
