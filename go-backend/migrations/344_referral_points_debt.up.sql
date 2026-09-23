-- Track referral clawbacks that exceed a referrer's currently available points.
-- Future positive point adjustments repay this debt before increasing the
-- spendable balance.
ALTER TABLE user_loyalty
    ADD COLUMN IF NOT EXISTS debt_points INTEGER NOT NULL DEFAULT 0;

ALTER TABLE user_loyalty
    ADD CONSTRAINT debt_points_non_negative CHECK (debt_points >= 0);

ALTER TABLE loyalty_transactions
    ADD COLUMN IF NOT EXISTS debt_balance INTEGER NOT NULL DEFAULT 0;
