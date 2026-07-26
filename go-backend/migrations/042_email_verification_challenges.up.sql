CREATE TABLE IF NOT EXISTS email_verification_challenges (
    id BIGSERIAL PRIMARY KEY,
    purpose VARCHAR(80) NOT NULL,
    email VARCHAR(255) NOT NULL,
    subject VARCHAR(500) NOT NULL,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_email_verification_challenges_purpose
    ON email_verification_challenges(purpose);
CREATE INDEX IF NOT EXISTS idx_email_verification_challenges_email
    ON email_verification_challenges(email);
CREATE INDEX IF NOT EXISTS idx_email_verification_challenges_expires_at
    ON email_verification_challenges(expires_at);
