INSERT INTO settings (key, value, type, locale, "group", is_public, description, created_at, updated_at)
VALUES
    ('paypal_dispute_invoice_seller_name', '', 'string', 'global', 'paypal_dispute_invoice_seller', false, 'PayPal commercial invoice seller profile: Seller legal/business name', NOW(), NOW()),
    ('paypal_dispute_invoice_seller_address', '', 'string', 'global', 'paypal_dispute_invoice_seller', false, 'PayPal commercial invoice seller profile: Seller commercial address', NOW(), NOW()),
    ('paypal_dispute_invoice_seller_email', '', 'string', 'global', 'paypal_dispute_invoice_seller', false, 'PayPal commercial invoice seller profile: Seller contact email', NOW(), NOW()),
    ('paypal_dispute_invoice_seller_phone', '', 'string', 'global', 'paypal_dispute_invoice_seller', false, 'PayPal commercial invoice seller profile: Seller contact phone', NOW(), NOW()),
    ('paypal_dispute_invoice_seller_website', '', 'string', 'global', 'paypal_dispute_invoice_seller', false, 'PayPal commercial invoice seller profile: Seller website', NOW(), NOW()),
    ('paypal_dispute_invoice_seller_tax_id', '', 'string', 'global', 'paypal_dispute_invoice_seller', false, 'PayPal commercial invoice seller profile: Seller tax identifier', NOW(), NOW())
ON CONFLICT (key, locale) DO NOTHING;
