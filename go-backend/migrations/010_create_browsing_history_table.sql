-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS browsing_history (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    product_id BIGINT UNSIGNED NOT NULL,
    view_count INT NOT NULL DEFAULT 1 COMMENT '浏览次数',
    last_viewed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '最后浏览时间',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_user_id (user_id),
    INDEX idx_product_id (product_id),
    INDEX idx_last_viewed_at (last_viewed_at),
    UNIQUE KEY uk_user_product (user_id, product_id),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户浏览历史表';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS browsing_history;
-- +goose StatementEnd
