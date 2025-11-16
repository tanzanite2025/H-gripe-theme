-- 更新货币从 CNY 改为 USD
-- 执行此脚本前请备份数据库！

-- 1. 更新订单表的默认货币
ALTER TABLE `wp_tanz_orders` 
MODIFY COLUMN `currency` CHAR(3) NOT NULL DEFAULT 'USD';

-- 2. 更新现有订单的货币（如果需要）
-- 注意：这会将所有 CNY 订单改为 USD，请根据实际情况决定是否执行
-- UPDATE `wp_tanz_orders` SET `currency` = 'USD' WHERE `currency` = 'CNY';

-- 3. 更新礼品卡表的默认货币
ALTER TABLE `wp_tanz_giftcards` 
MODIFY COLUMN `currency` VARCHAR(16) NOT NULL DEFAULT 'USD';

-- 4. 更新现有礼品卡的货币（如果需要）
-- 注意：这会将所有 CNY 礼品卡改为 USD，请根据实际情况决定是否执行
-- UPDATE `wp_tanz_giftcards` SET `currency` = 'USD' WHERE `currency` = 'CNY';

-- 完成！
-- 注意：优惠券表 (wp_tanz_coupons) 没有货币字段，因为优惠券金额是应用到订单上的
