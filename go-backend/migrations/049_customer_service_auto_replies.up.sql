-- 049: version the customer-service automatic-reply rules and add the
-- structured message fields required by the admin-managed reply system.

CREATE TABLE IF NOT EXISTS ticket_auto_replies (
    id BIGSERIAL PRIMARY KEY,
    type VARCHAR(20) NOT NULL,
    trigger_keyword VARCHAR(255) NOT NULL DEFAULT '',
    reply_message TEXT NOT NULL,
    agent_id VARCHAR(100) NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    priority INTEGER NOT NULL DEFAULT 0,
    match_type VARCHAR(20) NOT NULL DEFAULT 'exact',
    locale VARCHAR(20) NOT NULL DEFAULT '*',
    message_type VARCHAR(40) NOT NULL DEFAULT 'text',
    metadata TEXT NOT NULL DEFAULT '',
    attachments TEXT NOT NULL DEFAULT '',
    cooldown_seconds INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE ticket_auto_replies
    ADD COLUMN IF NOT EXISTS trigger_keyword VARCHAR(255) NOT NULL DEFAULT '';

ALTER TABLE ticket_auto_replies
    ADD COLUMN IF NOT EXISTS reply_message TEXT NOT NULL DEFAULT '';

ALTER TABLE ticket_auto_replies
    ADD COLUMN IF NOT EXISTS agent_id VARCHAR(100) NOT NULL DEFAULT '';

ALTER TABLE ticket_auto_replies
    ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;

ALTER TABLE ticket_auto_replies
    ADD COLUMN IF NOT EXISTS priority INTEGER NOT NULL DEFAULT 0;

ALTER TABLE ticket_auto_replies
    ADD COLUMN IF NOT EXISTS match_type VARCHAR(20) NOT NULL DEFAULT 'exact';

ALTER TABLE ticket_auto_replies
    ADD COLUMN IF NOT EXISTS locale VARCHAR(20) NOT NULL DEFAULT '*';

ALTER TABLE ticket_auto_replies
    ADD COLUMN IF NOT EXISTS message_type VARCHAR(40) NOT NULL DEFAULT 'text';

ALTER TABLE ticket_auto_replies
    ADD COLUMN IF NOT EXISTS metadata TEXT NOT NULL DEFAULT '';

ALTER TABLE ticket_auto_replies
    ADD COLUMN IF NOT EXISTS attachments TEXT NOT NULL DEFAULT '';

ALTER TABLE ticket_auto_replies
    ADD COLUMN IF NOT EXISTS cooldown_seconds INTEGER NOT NULL DEFAULT 0;

ALTER TABLE ticket_auto_replies
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

ALTER TABLE ticket_auto_replies
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

UPDATE ticket_auto_replies
SET locale = '*'
WHERE locale IS NULL OR BTRIM(locale) = '';

UPDATE ticket_auto_replies
SET message_type = 'text'
WHERE message_type IS NULL OR BTRIM(message_type) = '';

UPDATE ticket_auto_replies
SET metadata = ''
WHERE metadata IS NULL;

UPDATE ticket_auto_replies
SET attachments = ''
WHERE attachments IS NULL;

UPDATE ticket_auto_replies
SET cooldown_seconds = 86400
WHERE type = 'welcome' AND cooldown_seconds = 0;

CREATE INDEX IF NOT EXISTS idx_ticket_auto_replies_active_type
    ON ticket_auto_replies(type, is_active, priority DESC, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_ticket_auto_replies_locale_agent
    ON ticket_auto_replies(locale, agent_id);
