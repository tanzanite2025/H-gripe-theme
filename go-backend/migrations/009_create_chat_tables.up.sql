-- 009: 创建聊天消息相关表

-- 聊天消息表
CREATE TABLE IF NOT EXISTS chat_messages (
    id BIGSERIAL PRIMARY KEY,
    session_id VARCHAR(255) NOT NULL,
    message_id VARCHAR(255) NOT NULL UNIQUE,
    content TEXT,
    role VARCHAR(50) NOT NULL,  -- user, agent, system
    timestamp BIGINT NOT NULL,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    agent_id VARCHAR(100),
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_chat_messages_session_id ON chat_messages(session_id);
CREATE INDEX idx_chat_messages_user_id ON chat_messages(user_id);
CREATE INDEX idx_chat_messages_timestamp ON chat_messages(timestamp);
CREATE INDEX idx_chat_messages_created_at ON chat_messages(created_at);

-- 聊天会话表（可选，用于会话管理）
CREATE TABLE IF NOT EXISTS chat_sessions (
    id BIGSERIAL PRIMARY KEY,
    session_id VARCHAR(255) NOT NULL UNIQUE,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    agent_id VARCHAR(100),
    status VARCHAR(50) DEFAULT 'active',  -- active, closed
    last_message TEXT,
    message_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_chat_sessions_user_id ON chat_sessions(user_id);
CREATE INDEX idx_chat_sessions_status ON chat_sessions(status);
CREATE INDEX idx_chat_sessions_updated_at ON chat_sessions(updated_at);

-- 触发器：自动更新 updated_at
CREATE OR REPLACE FUNCTION update_chat_tables_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_chat_messages_updated_at
    BEFORE UPDATE ON chat_messages
    FOR EACH ROW
    EXECUTE FUNCTION update_chat_tables_updated_at();

CREATE TRIGGER trigger_update_chat_sessions_updated_at
    BEFORE UPDATE ON chat_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_chat_tables_updated_at();

-- 注释
COMMENT ON TABLE chat_messages IS '聊天消息表';
COMMENT ON TABLE chat_sessions IS '聊天会话表';
COMMENT ON COLUMN chat_messages.session_id IS '会话ID，用于关联同一次对话';
COMMENT ON COLUMN chat_messages.message_id IS '消息唯一ID，前端生成，用于幂等性';
COMMENT ON COLUMN chat_messages.role IS '消息角色: user(用户), agent(客服), system(系统)';
COMMENT ON COLUMN chat_messages.timestamp IS '消息时间戳（毫秒）';
COMMENT ON COLUMN chat_messages.metadata IS '额外元数据，JSON格式';
