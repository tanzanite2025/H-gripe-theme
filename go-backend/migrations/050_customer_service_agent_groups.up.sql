-- 050: add first-class customer-service groups instead of frontend name inference.

CREATE TABLE IF NOT EXISTS customer_service_agent_groups (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_customer_service_agent_groups_status_sort
    ON customer_service_agent_groups(status, sort_order, id);

CREATE INDEX IF NOT EXISTS idx_customer_service_agent_groups_deleted_at
    ON customer_service_agent_groups(deleted_at);

CREATE TABLE IF NOT EXISTS customer_service_agent_group_members (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL REFERENCES customer_service_agent_groups(id) ON DELETE CASCADE,
    agent_profile_id BIGINT NOT NULL REFERENCES customer_service_agent_profiles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_customer_service_agent_group_member UNIQUE (group_id, agent_profile_id)
);

CREATE INDEX IF NOT EXISTS idx_customer_service_agent_group_members_group_id
    ON customer_service_agent_group_members(group_id);

CREATE INDEX IF NOT EXISTS idx_customer_service_agent_group_members_agent_profile_id
    ON customer_service_agent_group_members(agent_profile_id);

INSERT INTO customer_service_agent_groups (code, name, description, status, sort_order)
VALUES
    ('sales', 'Sales', 'Pre-sales questions, pricing and quotes.', 'active', 10),
    ('technical_support', 'Technical Support', 'Compatibility, specs, setup and troubleshooting.', 'active', 20),
    ('after_sales', 'After Sales', 'Order tracking, warranty, returns and post-purchase help.', 'active', 30)
ON CONFLICT (code) DO NOTHING;

INSERT INTO customer_service_agent_group_members (group_id, agent_profile_id)
SELECT g.id, p.id
FROM customer_service_agent_profiles p
JOIN customer_service_agent_groups g
    ON (
        (g.code = 'technical_support' AND (
            LOWER(COALESCE(p.name, '')) LIKE '%tech%'
            OR LOWER(COALESCE(p.agent_id, '')) LIKE '%tech%'
            OR LOWER(COALESCE(p.email, '')) LIKE '%tech%'
        ))
        OR (g.code = 'after_sales' AND (
            LOWER(COALESCE(p.name, '')) LIKE '%after%'
            OR LOWER(COALESCE(p.agent_id, '')) LIKE '%after%'
            OR LOWER(COALESCE(p.email, '')) LIKE '%support%'
        ))
        OR (g.code = 'sales' AND (
            LOWER(COALESCE(p.name, '')) LIKE '%sale%'
            OR LOWER(COALESCE(p.agent_id, '')) LIKE '%sale%'
            OR LOWER(COALESCE(p.email, '')) LIKE '%sale%'
        ))
    )
ON CONFLICT (group_id, agent_profile_id) DO NOTHING;

ALTER TABLE ticket_auto_replies
    ADD COLUMN IF NOT EXISTS group_id BIGINT REFERENCES customer_service_agent_groups(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_ticket_auto_replies_group_id
    ON ticket_auto_replies(group_id);
