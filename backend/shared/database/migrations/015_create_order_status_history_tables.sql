-- 订单状态历史表
CREATE TABLE order_status_history (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT UNSIGNED NOT NULL,
    order_id BIGINT UNSIGNED NOT NULL,
    from_status TINYINT NOT NULL,
    to_status TINYINT NOT NULL,
    reason VARCHAR(255) NOT NULL,
    operator_id BIGINT UNSIGNED,
    operator_type ENUM('customer', 'merchant', 'system', 'admin') NOT NULL DEFAULT 'system',
    metadata JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_tenant_id (tenant_id),
    INDEX idx_order_id (order_id),
    INDEX idx_to_status (to_status),
    INDEX idx_created_at (created_at),
    INDEX idx_operator (operator_id, operator_type),
    CONSTRAINT fk_status_history_order FOREIGN KEY (order_id) REFERENCES orders (id) ON DELETE CASCADE
);

-- 订单超时配置表
CREATE TABLE order_timeout_configs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT UNSIGNED NOT NULL,
    merchant_id BIGINT UNSIGNED,
    payment_timeout_minutes INT NOT NULL DEFAULT 30,
    processing_timeout_hours INT NOT NULL DEFAULT 24,
    auto_complete_enabled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_tenant_id (tenant_id),
    INDEX idx_merchant_id (merchant_id),
    UNIQUE KEY uk_tenant_merchant (tenant_id, merchant_id)
);

-- 扩展orders表（添加字段）
ALTER TABLE orders 
ADD COLUMN status_updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
ADD INDEX idx_status_updated_at (status_updated_at);