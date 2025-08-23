-- 库存变更记录表
CREATE TABLE inventory_records (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT UNSIGNED NOT NULL,
    product_id BIGINT UNSIGNED NOT NULL,
    change_type VARCHAR(50) NOT NULL,
    quantity_before INT NOT NULL,
    quantity_after INT NOT NULL,
    quantity_changed INT NOT NULL,
    reason VARCHAR(255) NOT NULL,
    reference_id VARCHAR(100) NULL,
    operated_by BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_tenant_product (tenant_id, product_id),
    INDEX idx_change_type (change_type),
    INDEX idx_created_at (created_at),
    INDEX idx_reference_id (reference_id),
    CONSTRAINT fk_inventory_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE
);

-- 库存预留锁定表
CREATE TABLE inventory_reservations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT UNSIGNED NOT NULL,
    product_id BIGINT UNSIGNED NOT NULL,
    reserved_quantity INT NOT NULL,
    reference_type VARCHAR(50) NOT NULL,
    reference_id VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    expires_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_tenant_product (tenant_id, product_id),
    INDEX idx_reference (reference_type, reference_id),
    INDEX idx_status_expires (status, expires_at),
    CONSTRAINT fk_reservation_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE
);

-- 库存预警规则表
CREATE TABLE inventory_alerts (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT UNSIGNED NOT NULL,
    product_id BIGINT UNSIGNED NOT NULL,
    alert_type VARCHAR(50) NOT NULL,
    threshold_value INT NOT NULL,
    notification_channels JSON NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    last_triggered_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_tenant_product (tenant_id, product_id),
    INDEX idx_alert_type (alert_type),
    INDEX idx_active (is_active),
    CONSTRAINT fk_alert_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE
);