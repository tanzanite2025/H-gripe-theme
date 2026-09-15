ALTER TABLE coupon_usage
    ADD COLUMN IF NOT EXISTS email VARCHAR(255) NOT NULL DEFAULT '';

UPDATE coupon_usage AS cu
   SET email = lower(btrim(COALESCE(orders.shipping_email, '')))
  FROM orders
 WHERE cu.order_id = orders.id
   AND btrim(COALESCE(cu.email, '')) = ''
   AND btrim(COALESCE(orders.shipping_email, '')) <> '';

CREATE INDEX IF NOT EXISTS idx_coupon_usage_coupon_email_status
    ON coupon_usage(coupon_id, email, status);

DROP TRIGGER IF EXISTS trigger_enforce_coupon_per_user_usage_limit ON coupon_usage;

CREATE OR REPLACE FUNCTION enforce_coupon_per_user_usage_limit()
RETURNS TRIGGER AS $$
DECLARE
    per_user_limit INTEGER;
    existing_usage_count BIGINT;
    normalized_email VARCHAR(255);
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

    normalized_email := lower(btrim(COALESCE(NEW.email, '')));
    NEW.email := normalized_email;

    IF per_user_limit > 0 AND NEW.status = 'applied' THEN
        IF NEW.user_id > 0 THEN
            SELECT COUNT(*)
              INTO existing_usage_count
              FROM coupon_usage
              WHERE coupon_id = NEW.coupon_id
                AND user_id = NEW.user_id
                AND status = 'applied'
                AND (TG_OP = 'INSERT' OR id <> NEW.id);
        ELSIF normalized_email <> '' THEN
            SELECT COUNT(*)
              INTO existing_usage_count
              FROM coupon_usage
              WHERE coupon_id = NEW.coupon_id
                AND lower(btrim(email)) = normalized_email
                AND status = 'applied'
                AND (TG_OP = 'INSERT' OR id <> NEW.id);
        ELSE
            RAISE EXCEPTION 'coupon usage identity required'
                USING ERRCODE = '23514';
        END IF;

        IF existing_usage_count >= per_user_limit THEN
            RAISE EXCEPTION 'coupon per-user usage limit reached'
                USING ERRCODE = '23514';
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_enforce_coupon_per_user_usage_limit
    BEFORE INSERT OR UPDATE OF coupon_id, user_id, email, status ON coupon_usage
    FOR EACH ROW
    EXECUTE FUNCTION enforce_coupon_per_user_usage_limit();
