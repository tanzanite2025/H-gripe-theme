ALTER TABLE coupon_usage
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'applied',
    ADD COLUMN IF NOT EXISTS reversed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS reversal_reason TEXT;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
          FROM pg_constraint
         WHERE conname = 'coupon_usage_status_valid'
    ) THEN
        ALTER TABLE coupon_usage
            ADD CONSTRAINT coupon_usage_status_valid
            CHECK (status IN ('applied', 'reversed'));
    END IF;
END
$$;

CREATE INDEX IF NOT EXISTS idx_coupon_usage_coupon_user_status
    ON coupon_usage(coupon_id, user_id, status);

DROP TRIGGER IF EXISTS trigger_enforce_coupon_per_user_usage_limit ON coupon_usage;

CREATE OR REPLACE FUNCTION enforce_coupon_per_user_usage_limit()
RETURNS TRIGGER AS $$
DECLARE
    per_user_limit INTEGER;
    existing_usage_count INTEGER;
BEGIN
    SELECT usage_limit_per_user
      INTO per_user_limit
      FROM coupons
      WHERE id = NEW.coupon_id
      FOR UPDATE;

    IF per_user_limit IS NULL THEN
        RAISE EXCEPTION 'coupon % does not exist', NEW.coupon_id
            USING ERRCODE = '23503';
    END IF;

    IF per_user_limit > 0 AND NEW.user_id > 0 AND NEW.status = 'applied' THEN
        SELECT COUNT(*)
          INTO existing_usage_count
          FROM coupon_usage
          WHERE coupon_id = NEW.coupon_id
            AND user_id = NEW.user_id
            AND status = 'applied'
            AND (TG_OP = 'INSERT' OR id <> NEW.id);

        IF existing_usage_count >= per_user_limit THEN
            RAISE EXCEPTION 'coupon per-user usage limit reached'
                USING ERRCODE = '23514';
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_enforce_coupon_per_user_usage_limit
    BEFORE INSERT OR UPDATE OF coupon_id, user_id, status ON coupon_usage
    FOR EACH ROW
    EXECUTE FUNCTION enforce_coupon_per_user_usage_limit();
