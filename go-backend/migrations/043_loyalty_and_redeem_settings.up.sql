-- Expose point redemption and loyalty rules for existing databases.
-- Keep administrator overrides intact when the migration is rerun.

INSERT INTO settings (key, value, type, locale, "group", is_public, description, created_at, updated_at)
VALUES
    ('tz_redeem_enabled', '1', 'boolean', 'en', 'redeem', true, 'Whether point redemption is enabled', NOW(), NOW()),
    ('tz_redeem_enabled', '1', 'boolean', 'zh', 'redeem', true, '是否启用积分兑换', NOW(), NOW()),
    ('tz_redeem_exchange_rate', '100', 'number', 'en', 'redeem', true, 'Redemption exchange rate (e.g. 100 points = 1 unit)', NOW(), NOW()),
    ('tz_redeem_exchange_rate', '100', 'number', 'zh', 'redeem', true, '积分兑换比例（如100积分=1元）', NOW(), NOW()),
    ('tz_redeem_min_points', '1000', 'number', 'en', 'redeem', true, 'Minimum points required to redeem', NOW(), NOW()),
    ('tz_redeem_min_points', '1000', 'number', 'zh', 'redeem', true, '兑换所需最小积分', NOW(), NOW()),
    ('tz_redeem_max_value_per_day', '500', 'number', 'en', 'redeem', true, 'Maximum value redeemable per day', NOW(), NOW()),
    ('tz_redeem_max_value_per_day', '500', 'number', 'zh', 'redeem', true, '每日最大可兑换价值', NOW(), NOW()),
    ('tz_redeem_card_expiry_days', '365', 'number', 'en', 'redeem', true, 'Redeemed gift card expiry days', NOW(), NOW()),
    ('tz_redeem_card_expiry_days', '365', 'number', 'zh', 'redeem', true, '兑换出的礼品卡有效期天数', NOW(), NOW()),
    ('tz_redeem_preset_values', '10,50,100,200,500', 'string', 'en', 'redeem', true, 'Preset gift card values for redemption', NOW(), NOW()),
    ('tz_redeem_preset_values', '10,50,100,200,500', 'string', 'zh', 'redeem', true, '预设的可兑换礼品卡面额', NOW(), NOW()),
    ('tz_loyalty_referral_referrer_points', '100', 'number', 'en', 'loyalty', true, 'Points awarded to the referrer after the first purchase', NOW(), NOW()),
    ('tz_loyalty_referral_referrer_points', '100', 'number', 'zh', 'loyalty', true, '被推荐用户首次购买后，推荐人获得的积分', NOW(), NOW()),
    ('tz_loyalty_referral_referee_points', '50', 'number', 'en', 'loyalty', true, 'Points awarded to the referred user after the first purchase', NOW(), NOW()),
    ('tz_loyalty_referral_referee_points', '50', 'number', 'zh', 'loyalty', true, '被推荐用户首次购买后，被推荐人获得的积分', NOW(), NOW()),
    ('tz_loyalty_checkin_base_points', '10', 'number', 'en', 'loyalty', true, 'Base points awarded for a daily check-in', NOW(), NOW()),
    ('tz_loyalty_checkin_base_points', '10', 'number', 'zh', 'loyalty', true, '每日签到基础积分', NOW(), NOW()),
    ('tz_loyalty_checkin_streak_interval_days', '7', 'number', 'en', 'loyalty', true, 'Consecutive check-in days required for a bonus', NOW(), NOW()),
    ('tz_loyalty_checkin_streak_interval_days', '7', 'number', 'zh', 'loyalty', true, '连续签到多少天增加一次奖励', NOW(), NOW()),
    ('tz_loyalty_checkin_streak_bonus_points', '5', 'number', 'en', 'loyalty', true, 'Additional points for each completed check-in streak interval', NOW(), NOW()),
    ('tz_loyalty_checkin_streak_bonus_points', '5', 'number', 'zh', 'loyalty', true, '每完成一个连续签到周期增加的积分', NOW(), NOW()),
    ('tz_loyalty_checkin_max_points', '50', 'number', 'en', 'loyalty', true, 'Maximum points awarded for one daily check-in', NOW(), NOW()),
    ('tz_loyalty_checkin_max_points', '50', 'number', 'zh', 'loyalty', true, '单次每日签到最多获得的积分', NOW(), NOW())
ON CONFLICT (key, locale) DO NOTHING;
