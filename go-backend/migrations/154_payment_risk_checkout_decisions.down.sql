ALTER TABLE payment_risk_snapshots
    DROP COLUMN IF EXISTS three_ds_upgrade_rate,
    DROP COLUMN IF EXISTS three_ds_exemption_count,
    DROP COLUMN IF EXISTS three_ds_challenge_count,
    DROP COLUMN IF EXISTS three_ds_upgrade_count,
    DROP COLUMN IF EXISTS checkout_attempt_count;

DROP TABLE IF EXISTS payment_risk_checkout_decisions;
