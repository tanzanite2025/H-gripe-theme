DROP TRIGGER IF EXISTS trigger_enforce_coupon_per_user_usage_limit ON coupon_usage;

DROP FUNCTION IF EXISTS enforce_coupon_per_user_usage_limit();

DROP INDEX IF EXISTS idx_coupon_usage_coupon_user;
