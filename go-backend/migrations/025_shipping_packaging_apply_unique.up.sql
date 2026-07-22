CREATE UNIQUE INDEX IF NOT EXISTS idx_shipping_packaging_rule_apply_rule_product
ON shipping_packaging_rule_applies (rule_id, product_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_shipping_packaging_rule_apply_product
ON shipping_packaging_rule_applies (product_id);
