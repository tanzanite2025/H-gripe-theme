ALTER TABLE loyalty_transactions
    DROP COLUMN IF EXISTS debt_balance;

ALTER TABLE user_loyalty
    DROP CONSTRAINT IF EXISTS debt_points_non_negative,
    DROP COLUMN IF EXISTS debt_points;
