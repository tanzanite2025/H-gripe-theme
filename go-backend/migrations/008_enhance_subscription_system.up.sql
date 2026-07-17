-- Migration: 008_enhance_subscription_system
-- Description: 增强订阅系统，添加标签、退订令牌和时间戳字段
-- Date: 2024-01-15

-- 修改 subscriptions 表
ALTER TABLE subscriptions 
ADD COLUMN IF NOT EXISTS tags VARCHAR(500) DEFAULT '',
ADD COLUMN IF NOT EXISTS unsub_token VARCHAR(64) UNIQUE,
ADD COLUMN IF NOT EXISTS subscribed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN IF NOT EXISTS unsubscribed_at TIMESTAMP NULL;

-- 为现有记录生成退订令牌
UPDATE subscriptions 
SET unsub_token = MD5(email || id::text || clock_timestamp()::text || random()::text)
    || MD5(random()::text || clock_timestamp()::text || id::text)
WHERE unsub_token IS NULL OR unsub_token = '';

-- 为现有记录设置订阅时间
UPDATE subscriptions 
SET subscribed_at = created_at
WHERE subscribed_at IS NULL;

-- 为已退订的记录设置退订时间
UPDATE subscriptions 
SET unsubscribed_at = updated_at
WHERE status = 'unsubscribed' AND unsubscribed_at IS NULL;

-- 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_tags ON subscriptions(tags);
CREATE INDEX IF NOT EXISTS idx_subscriptions_unsub_token ON subscriptions(unsub_token);
CREATE INDEX IF NOT EXISTS idx_subscriptions_subscribed_at ON subscriptions(subscribed_at);

-- 添加注释
COMMENT ON COLUMN subscriptions.tags IS '订阅标签，逗号分隔';
COMMENT ON COLUMN subscriptions.unsub_token IS '退订令牌，用于一键退订';
COMMENT ON COLUMN subscriptions.subscribed_at IS '订阅时间';
COMMENT ON COLUMN subscriptions.unsubscribed_at IS '退订时间';
