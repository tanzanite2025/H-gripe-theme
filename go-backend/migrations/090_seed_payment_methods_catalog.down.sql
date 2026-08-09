DELETE FROM payment_methods
WHERE code IN ('card', 'paypal', 'alipay', 'wechat');
