-- Initial Database Schema for Storefront Go Backend
-- Applied by golang-migrate during server startup.

-- Users Table
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role VARCHAR(50) DEFAULT 'user',
    locale VARCHAR(10) DEFAULT 'en',
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- Posts Table
CREATE TABLE IF NOT EXISTS posts (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    content TEXT,
    excerpt TEXT,
    status VARCHAR(20) DEFAULT 'draft',
    author_id BIGINT NOT NULL,
    locale VARCHAR(10) DEFAULT 'en',
    parent_id BIGINT,
    featured_image VARCHAR(500),
    view_count INT DEFAULT 0,
    meta_title VARCHAR(255),
    meta_description TEXT,
    tags VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (author_id) REFERENCES users(id),
    FOREIGN KEY (parent_id) REFERENCES posts(id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_posts_slug_locale ON posts(slug, locale);
CREATE INDEX IF NOT EXISTS idx_posts_status ON posts(status);
CREATE INDEX IF NOT EXISTS idx_posts_locale ON posts(locale);
CREATE INDEX IF NOT EXISTS idx_posts_author_id ON posts(author_id);
CREATE INDEX IF NOT EXISTS idx_posts_deleted_at ON posts(deleted_at);

-- Categories Table
CREATE TABLE IF NOT EXISTS categories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    description TEXT,
    locale VARCHAR(10) DEFAULT 'en',
    parent_id BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES categories(id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_category_slug_locale ON categories(slug, locale);
CREATE INDEX IF NOT EXISTS idx_categories_deleted_at ON categories(deleted_at);

-- Post Categories (Many-to-Many)
CREATE TABLE IF NOT EXISTS post_categories (
    id BIGSERIAL PRIMARY KEY,
    post_id BIGINT NOT NULL,
    category_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_post_categories_post_id ON post_categories(post_id);
CREATE INDEX IF NOT EXISTS idx_post_categories_category_id ON post_categories(category_id);

-- Products Table
CREATE TABLE IF NOT EXISTS products (
    id BIGSERIAL PRIMARY KEY,
    sku VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(500) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    description TEXT,
    short_description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    sale_price DECIMAL(10, 2),
    stock INT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    locale VARCHAR(10) DEFAULT 'en',
    parent_id BIGINT,
    featured BOOLEAN DEFAULT FALSE,
    view_count INT DEFAULT 0,
    meta_title VARCHAR(255),
    meta_description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES products(id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_product_slug_locale ON products(slug, locale);
CREATE INDEX IF NOT EXISTS idx_products_sku ON products(sku);
CREATE INDEX IF NOT EXISTS idx_products_status ON products(status);
CREATE INDEX IF NOT EXISTS idx_products_locale ON products(locale);
CREATE INDEX IF NOT EXISTS idx_products_deleted_at ON products(deleted_at);

-- Carts Table
CREATE TABLE IF NOT EXISTS carts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    session_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_carts_user_id ON carts(user_id);
CREATE INDEX IF NOT EXISTS idx_carts_session_id ON carts(session_id);
CREATE INDEX IF NOT EXISTS idx_carts_deleted_at ON carts(deleted_at);

-- Cart Items Table
CREATE TABLE IF NOT EXISTS cart_items (
    id BIGSERIAL PRIMARY KEY,
    cart_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    quantity INT NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (cart_id) REFERENCES carts(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id)
);

CREATE INDEX IF NOT EXISTS idx_cart_items_cart_id ON cart_items(cart_id);
CREATE INDEX IF NOT EXISTS idx_cart_items_product_id ON cart_items(product_id);

-- Settings Table
CREATE TABLE IF NOT EXISTS settings (
    id BIGSERIAL PRIMARY KEY,
    key VARCHAR(255) NOT NULL,
    value TEXT,
    type VARCHAR(50) DEFAULT 'string',
    locale VARCHAR(10) DEFAULT 'en',
    "group" VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_setting_key_locale ON settings(key, locale);
CREATE INDEX IF NOT EXISTS idx_settings_group ON settings("group");
CREATE INDEX IF NOT EXISTS idx_settings_deleted_at ON settings(deleted_at);

-- Media Table
CREATE TABLE IF NOT EXISTS media (
    id BIGSERIAL PRIMARY KEY,
    filename VARCHAR(255) NOT NULL,
    url VARCHAR(500) NOT NULL,
    mime_type VARCHAR(100),
    size BIGINT,
    width INT,
    height INT,
    alt VARCHAR(255),
    caption TEXT,
    uploader_id BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (uploader_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_media_uploader_id ON media(uploader_id);
CREATE INDEX IF NOT EXISTS idx_media_deleted_at ON media(deleted_at);

-- FAQs Table
CREATE TABLE IF NOT EXISTS faqs (
    id BIGSERIAL PRIMARY KEY,
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    category VARCHAR(100),
    locale VARCHAR(10) DEFAULT 'en',
    parent_id BIGINT,
    "order" INT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'published',
    view_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES faqs(id)
);

CREATE INDEX IF NOT EXISTS idx_faqs_category ON faqs(category);
CREATE INDEX IF NOT EXISTS idx_faqs_locale ON faqs(locale);
CREATE INDEX IF NOT EXISTS idx_faqs_status ON faqs(status);
CREATE INDEX IF NOT EXISTS idx_faqs_deleted_at ON faqs(deleted_at);

-- Subscriptions Table
CREATE TABLE IF NOT EXISTS subscriptions (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    locale VARCHAR(10) DEFAULT 'en',
    source VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_email ON subscriptions(email);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_deleted_at ON subscriptions(deleted_at);

-- Historical application baseline tables.
--
-- These tables were previously created only by GORM AutoMigrate. They are
-- defined here so a new database can replay the versioned SQL migrations
-- without depending on application startup order.

-- Orders are intentionally created in their pre-currency/pre-FX shape.
-- Migrations 022, 059, and 109 evolve these columns later.
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    order_number VARCHAR(255) NOT NULL,
    user_id BIGINT,
    status VARCHAR(50) DEFAULT 'pending',
    payment_method VARCHAR(100),
    payment_status VARCHAR(50) DEFAULT 'unpaid',
    shipping_method VARCHAR(100),
    shipping_status VARCHAR(50) DEFAULT 'pending',
    tracking_number VARCHAR(255),
    carrier_code VARCHAR(100),
    subtotal_amount NUMERIC(12, 2) NOT NULL,
    shipping_fee NUMERIC(12, 2) DEFAULT 0,
    tax_amount NUMERIC(12, 2) DEFAULT 0,
    discount_amount NUMERIC(12, 2) DEFAULT 0,
    total_amount NUMERIC(12, 2) NOT NULL,
    coupon_code VARCHAR(100),
    points_used INTEGER DEFAULT 0,
    points_value NUMERIC(12, 2) DEFAULT 0,
    shipping_first_name TEXT,
    shipping_last_name TEXT,
    shipping_company TEXT,
    shipping_address1 TEXT,
    shipping_address2 TEXT,
    shipping_city TEXT,
    shipping_state TEXT,
    shipping_postal_code TEXT,
    shipping_country TEXT,
    shipping_phone TEXT,
    shipping_email TEXT,
    billing_first_name TEXT,
    billing_last_name TEXT,
    billing_company TEXT,
    billing_address1 TEXT,
    billing_address2 TEXT,
    billing_city TEXT,
    billing_state TEXT,
    billing_postal_code TEXT,
    billing_country TEXT,
    billing_phone TEXT,
    billing_email TEXT,
    customer_note TEXT,
    admin_note TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    paid_at TIMESTAMPTZ,
    shipped_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_order_number ON orders(order_number);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_payment_status ON orders(payment_status);
CREATE INDEX IF NOT EXISTS idx_orders_shipping_status ON orders(shipping_status);
CREATE INDEX IF NOT EXISTS idx_orders_deleted_at ON orders(deleted_at);

-- Variant identity is added and made mandatory by migration 013.
CREATE TABLE IF NOT EXISTS order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    product_name TEXT NOT NULL,
    sku TEXT,
    quantity INTEGER NOT NULL,
    price NUMERIC(12, 2) NOT NULL,
    subtotal NUMERIC(12, 2) NOT NULL,
    tax_amount NUMERIC(12, 2) DEFAULT 0,
    discount NUMERIC(12, 2) DEFAULT 0,
    total NUMERIC(12, 2) NOT NULL,
    attributes TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_orders_items FOREIGN KEY (order_id) REFERENCES orders(id)
);

CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items(product_id);

CREATE TABLE IF NOT EXISTS transactions (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    transaction_id VARCHAR(255),
    payment_method VARCHAR(100) NOT NULL,
    amount NUMERIC(12, 2) NOT NULL,
    currency VARCHAR(8) NOT NULL,
    status VARCHAR(50),
    gateway_response TEXT,
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_transaction_id ON transactions(transaction_id);
CREATE INDEX IF NOT EXISTS idx_transactions_order_id ON transactions(order_id);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
CREATE INDEX IF NOT EXISTS idx_transactions_deleted_at ON transactions(deleted_at);

-- Refund monetary snapshots are added by migrations 061 and 110.
CREATE TABLE IF NOT EXISTS refunds (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    transaction_id BIGINT,
    refund_id VARCHAR(255),
    amount NUMERIC(12, 2) NOT NULL,
    reason TEXT,
    status VARCHAR(50),
    refunded_by BIGINT,
    gateway_response TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_refunds_refund_id ON refunds(refund_id);
CREATE INDEX IF NOT EXISTS idx_refunds_order_id ON refunds(order_id);
CREATE INDEX IF NOT EXISTS idx_refunds_transaction_id ON refunds(transaction_id);
CREATE INDEX IF NOT EXISTS idx_refunds_status ON refunds(status);
CREATE INDEX IF NOT EXISTS idx_refunds_deleted_at ON refunds(deleted_at);

CREATE TABLE IF NOT EXISTS payment_methods (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(100) NOT NULL,
    icon VARCHAR(255),
    description TEXT,
    fee_type VARCHAR(50) DEFAULT 'fixed',
    fee_value NUMERIC(12, 2) DEFAULT 0,
    min_amount NUMERIC(12, 2) DEFAULT 0,
    max_amount NUMERIC(12, 2) DEFAULT 0,
    enabled BOOLEAN DEFAULT TRUE,
    sort_order INTEGER DEFAULT 0,
    settings TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_payment_methods_code ON payment_methods(code);
CREATE INDEX IF NOT EXISTS idx_payment_methods_deleted_at ON payment_methods(deleted_at);

CREATE TABLE IF NOT EXISTS tax_rates (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    country VARCHAR(100),
    state VARCHAR(100),
    city VARCHAR(100),
    postal_code VARCHAR(100),
    rate NUMERIC(8, 4) NOT NULL,
    priority INTEGER DEFAULT 0,
    compound BOOLEAN DEFAULT FALSE,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_tax_rates_country ON tax_rates(country);
CREATE INDEX IF NOT EXISTS idx_tax_rates_state ON tax_rates(state);
CREATE INDEX IF NOT EXISTS idx_tax_rates_deleted_at ON tax_rates(deleted_at);

CREATE TABLE IF NOT EXISTS coupons (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL,
    value NUMERIC(12, 2) NOT NULL,
    description TEXT,
    min_amount NUMERIC(12, 2) DEFAULT 0,
    max_discount NUMERIC(12, 2) DEFAULT 0,
    usage_limit INTEGER DEFAULT 0,
    usage_limit_per_user INTEGER DEFAULT 0,
    used_count INTEGER DEFAULT 0,
    start_date TIMESTAMPTZ,
    end_date TIMESTAMPTZ,
    applicable_products TEXT,
    excluded_products TEXT,
    applicable_categories TEXT,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT coupon_usage_limit_non_negative CHECK (usage_limit >= 0),
    CONSTRAINT coupon_used_count_valid CHECK (used_count >= 0 AND (usage_limit = 0 OR used_count <= usage_limit))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_coupons_code ON coupons(code);
CREATE INDEX IF NOT EXISTS idx_coupons_deleted_at ON coupons(deleted_at);

CREATE TABLE IF NOT EXISTS coupon_usage (
    id BIGSERIAL PRIMARY KEY,
    coupon_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    order_id BIGINT NOT NULL,
    discount NUMERIC(12, 2) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_coupon_usage_coupon_id ON coupon_usage(coupon_id);
CREATE INDEX IF NOT EXISTS idx_coupon_usage_user_id ON coupon_usage(user_id);
CREATE INDEX IF NOT EXISTS idx_coupon_usage_order_id ON coupon_usage(order_id);

-- Gift-card cents conversion is performed by migration 026.
CREATE TABLE IF NOT EXISTS gift_cards (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(100) NOT NULL,
    initial_value NUMERIC(12, 2) NOT NULL,
    balance NUMERIC(12, 2) NOT NULL,
    currency VARCHAR(8) NOT NULL,
    status VARCHAR(50),
    recipient_email VARCHAR(255),
    recipient_name VARCHAR(255),
    sender_name VARCHAR(255),
    message TEXT,
    cover_image TEXT,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_gift_cards_code ON gift_cards(code);
CREATE INDEX IF NOT EXISTS idx_gift_cards_status ON gift_cards(status);
CREATE INDEX IF NOT EXISTS idx_gift_cards_deleted_at ON gift_cards(deleted_at);

CREATE TABLE IF NOT EXISTS gift_card_transactions (
    id BIGSERIAL PRIMARY KEY,
    gift_card_id BIGINT NOT NULL,
    order_id BIGINT,
    type VARCHAR(50) NOT NULL,
    amount NUMERIC(12, 2) NOT NULL,
    balance NUMERIC(12, 2) NOT NULL,
    note TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_gift_card_transactions_gift_card_id ON gift_card_transactions(gift_card_id);
CREATE INDEX IF NOT EXISTS idx_gift_card_transactions_order_id ON gift_card_transactions(order_id);

CREATE TABLE IF NOT EXISTS loyalty_transactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    type VARCHAR(50) NOT NULL,
    points INTEGER NOT NULL,
    balance INTEGER NOT NULL,
    source VARCHAR(100),
    source_id BIGINT,
    description TEXT,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_loyalty_transactions_user_id ON loyalty_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_loyalty_transactions_type ON loyalty_transactions(type);
CREATE INDEX IF NOT EXISTS idx_loyalty_transactions_source ON loyalty_transactions(source);
CREATE INDEX IF NOT EXISTS idx_loyalty_transactions_deleted_at ON loyalty_transactions(deleted_at);

CREATE TABLE IF NOT EXISTS user_loyalty (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    total_points INTEGER DEFAULT 0,
    available_points INTEGER DEFAULT 0,
    used_points INTEGER DEFAULT 0,
    expired_points INTEGER DEFAULT 0,
    member_level_id BIGINT,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT total_points_non_negative CHECK (total_points >= 0),
    CONSTRAINT available_points_non_negative CHECK (available_points >= 0),
    CONSTRAINT used_points_non_negative CHECK (used_points >= 0),
    CONSTRAINT expired_points_non_negative CHECK (expired_points >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_loyalty_user_id ON user_loyalty(user_id);

CREATE TABLE IF NOT EXISTS check_ins (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    check_in_date VARCHAR(32) NOT NULL,
    points_earned INTEGER NOT NULL,
    consecutive_days INTEGER DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_check_ins_user_id ON check_ins(user_id);
CREATE INDEX IF NOT EXISTS idx_check_ins_check_in_date ON check_ins(check_in_date);
CREATE INDEX IF NOT EXISTS idx_check_ins_user_date ON check_ins(user_id, check_in_date);

CREATE TABLE IF NOT EXISTS referrals (
    id BIGSERIAL PRIMARY KEY,
    referrer_id BIGINT NOT NULL,
    referred_id BIGINT,
    referral_code VARCHAR(255),
    status VARCHAR(50),
    points_earned INTEGER DEFAULT 0,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_referrals_referral_code ON referrals(referral_code);
CREATE INDEX IF NOT EXISTS idx_referrals_referrer_id ON referrals(referrer_id);
CREATE INDEX IF NOT EXISTS idx_referrals_referred_id ON referrals(referred_id);
CREATE INDEX IF NOT EXISTS idx_referrals_status ON referrals(status);
CREATE INDEX IF NOT EXISTS idx_referrals_deleted_at ON referrals(deleted_at);

CREATE TABLE IF NOT EXISTS product_attributes (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(120) NOT NULL,
    slug VARCHAR(120) NOT NULL,
    type VARCHAR(32) NOT NULL DEFAULT 'select',
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_filterable BOOLEAN NOT NULL DEFAULT TRUE,
    affects_sku BOOLEAN NOT NULL DEFAULT TRUE,
    affects_stock BOOLEAN NOT NULL DEFAULT FALSE,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    meta TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_product_attributes_slug ON product_attributes(slug);

CREATE TABLE IF NOT EXISTS product_attribute_values (
    id BIGSERIAL PRIMARY KEY,
    attribute_id BIGINT NOT NULL,
    name VARCHAR(120) NOT NULL,
    slug VARCHAR(120) NOT NULL,
    value VARCHAR(255),
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    meta TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_product_attributes_values FOREIGN KEY (attribute_id) REFERENCES product_attributes(id)
);

CREATE INDEX IF NOT EXISTS idx_product_attribute_values_attribute_id ON product_attribute_values(attribute_id);

CREATE TABLE IF NOT EXISTS reviews (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    order_id BIGINT,
    rating INTEGER NOT NULL,
    title TEXT,
    content TEXT,
    images TEXT,
    pros TEXT,
    cons TEXT,
    status VARCHAR(50) DEFAULT 'pending',
    featured BOOLEAN DEFAULT FALSE,
    verified BOOLEAN DEFAULT FALSE,
    helpful_count INTEGER DEFAULT 0,
    reply_content TEXT,
    replied_at TIMESTAMPTZ,
    replied_by BIGINT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_reviews_product FOREIGN KEY (product_id) REFERENCES products(id),
    CONSTRAINT fk_reviews_user FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_reviews_product_id ON reviews(product_id);
CREATE INDEX IF NOT EXISTS idx_reviews_user_id ON reviews(user_id);
CREATE INDEX IF NOT EXISTS idx_reviews_order_id ON reviews(order_id);
CREATE INDEX IF NOT EXISTS idx_reviews_status ON reviews(status);
CREATE INDEX IF NOT EXISTS idx_reviews_deleted_at ON reviews(deleted_at);

CREATE TABLE IF NOT EXISTS review_helpful (
    id BIGSERIAL PRIMARY KEY,
    review_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    helpful BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_review_helpful_review_id ON review_helpful(review_id);
CREATE INDEX IF NOT EXISTS idx_review_helpful_user_id ON review_helpful(user_id);

CREATE TABLE IF NOT EXISTS customer_service_agent_profiles (
    id BIGSERIAL PRIMARY KEY,
    agent_id VARCHAR(50) NOT NULL,
    user_id BIGINT,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100),
    avatar VARCHAR(500),
    whats_app VARCHAR(50),
    status VARCHAR(20) DEFAULT 'active',
    online_status VARCHAR(20) DEFAULT 'offline',
    last_active_at TIMESTAMPTZ,
    last_login TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_customer_service_agent_profiles_user FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_customer_service_agent_profiles_agent_id
    ON customer_service_agent_profiles(agent_id);
CREATE INDEX IF NOT EXISTS idx_customer_service_agent_profiles_user_id
    ON customer_service_agent_profiles(user_id);
CREATE INDEX IF NOT EXISTS idx_customer_service_agent_profiles_email
    ON customer_service_agent_profiles(email);
CREATE INDEX IF NOT EXISTS idx_customer_service_agent_profiles_status
    ON customer_service_agent_profiles(status);
CREATE INDEX IF NOT EXISTS idx_customer_service_agent_profiles_online_status
    ON customer_service_agent_profiles(online_status);
CREATE INDEX IF NOT EXISTS idx_customer_service_agent_profiles_deleted_at
    ON customer_service_agent_profiles(deleted_at);

CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    username TEXT,
    action TEXT NOT NULL,
    resource TEXT NOT NULL,
    resource_id BIGINT,
    method TEXT,
    path TEXT,
    ip_address TEXT,
    user_agent TEXT,
    changes TEXT,
    old_value TEXT,
    new_value TEXT,
    status TEXT,
    error_message TEXT,
    duration BIGINT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_id ON audit_logs(resource_id);

CREATE TABLE IF NOT EXISTS feedback (
    id BIGSERIAL PRIMARY KEY,
    thread_key TEXT NOT NULL,
    user_id BIGINT NOT NULL,
    name TEXT,
    email TEXT,
    content TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    locale VARCHAR(20),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_feedback_thread_key ON feedback(thread_key);
CREATE INDEX IF NOT EXISTS idx_feedback_user_id ON feedback(user_id);
CREATE INDEX IF NOT EXISTS idx_feedback_status ON feedback(status);
CREATE INDEX IF NOT EXISTS idx_feedback_deleted_at ON feedback(deleted_at);

CREATE TABLE IF NOT EXISTS suggestion_feedback (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    full_name VARCHAR(120),
    email VARCHAR(190),
    country VARCHAR(80),
    order_number VARCHAR(80),
    product_category VARCHAR(60),
    request_type VARCHAR(60),
    message TEXT NOT NULL,
    attachments JSONB,
    meta JSONB,
    status VARCHAR(25) NOT NULL DEFAULT 'new',
    member_level_required VARCHAR(60),
    member_level_met BOOLEAN DEFAULT FALSE,
    eligibility_hash VARCHAR(190),
    reviewed_by BIGINT,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_suggestion_feedback_user_id ON suggestion_feedback(user_id);
CREATE INDEX IF NOT EXISTS idx_suggestion_feedback_status ON suggestion_feedback(status);
CREATE INDEX IF NOT EXISTS idx_suggestion_feedback_deleted_at ON suggestion_feedback(deleted_at);

CREATE TABLE IF NOT EXISTS spoke_histories (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    wheel_type VARCHAR(40),
    source_type VARCHAR(40),
    rim_brand VARCHAR(120),
    rim_model VARCHAR(160),
    hub_brand VARCHAR(120),
    hub_model VARCHAR(160),
    erdmm NUMERIC,
    left_flange_pcdmm NUMERIC,
    right_flange_pcdmm NUMERIC,
    left_flange_to_center_mm NUMERIC,
    right_flange_to_center_mm NUMERIC,
    spoke_count INTEGER,
    lacing_pattern VARCHAR(60),
    nipple_type VARCHAR(60),
    left_length_mm NUMERIC,
    right_length_mm NUMERIC,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_spoke_histories_user_id ON spoke_histories(user_id);
CREATE INDEX IF NOT EXISTS idx_spoke_histories_wheel_type ON spoke_histories(wheel_type);
CREATE INDEX IF NOT EXISTS idx_spoke_histories_source_type ON spoke_histories(source_type);
CREATE INDEX IF NOT EXISTS idx_spoke_histories_rim_brand ON spoke_histories(rim_brand);
CREATE INDEX IF NOT EXISTS idx_spoke_histories_rim_model ON spoke_histories(rim_model);
CREATE INDEX IF NOT EXISTS idx_spoke_histories_hub_brand ON spoke_histories(hub_brand);
CREATE INDEX IF NOT EXISTS idx_spoke_histories_hub_model ON spoke_histories(hub_model);
CREATE INDEX IF NOT EXISTS idx_spoke_histories_deleted_at ON spoke_histories(deleted_at);

CREATE TABLE IF NOT EXISTS wishlist_items (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_wishlist_items_product FOREIGN KEY (product_id) REFERENCES products(id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_wishlist_user_product ON wishlist_items(user_id, product_id);
CREATE INDEX IF NOT EXISTS idx_wishlist_items_user_id ON wishlist_items(user_id);
CREATE INDEX IF NOT EXISTS idx_wishlist_items_product_id ON wishlist_items(product_id);
CREATE INDEX IF NOT EXISTS idx_wishlist_items_deleted_at ON wishlist_items(deleted_at);
