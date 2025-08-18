-- 创建订单表
CREATE TABLE IF NOT EXISTS `orders` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '订单ID',
    `tenant_id` bigint(20) unsigned NOT NULL COMMENT '租户ID',
    `merchant_id` bigint(20) unsigned NOT NULL COMMENT '商户ID',
    `customer_id` bigint(20) unsigned NOT NULL COMMENT '客户ID(用户ID)',
    `order_number` varchar(50) NOT NULL COMMENT '订单号',
    `status` enum('pending','paid','processing','completed','cancelled','refunded') NOT NULL DEFAULT 'pending' COMMENT '订单状态',
    `items` json NOT NULL COMMENT '订单项目',
    `payment_info` json COMMENT '支付信息',
    `verification_info` json COMMENT '核销信息',
    `total_amount` decimal(10,2) NOT NULL COMMENT '总金额',
    `total_rights_cost` decimal(10,2) NOT NULL COMMENT '总权益成本',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_order_number` (`order_number`),
    KEY `idx_order_tenant` (`tenant_id`),
    KEY `idx_order_merchant` (`merchant_id`),
    KEY `idx_order_customer` (`customer_id`),
    KEY `idx_order_status` (`status`),
    KEY `idx_order_created` (`created_at`),
    KEY `idx_order_tenant_merchant` (`tenant_id`, `merchant_id`),
    KEY `idx_order_tenant_customer` (`tenant_id`, `customer_id`),
    CONSTRAINT `fk_order_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT `fk_order_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT `fk_order_customer` FOREIGN KEY (`customer_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单表';

-- 插入测试订单数据
INSERT INTO `orders` (`tenant_id`, `merchant_id`, `customer_id`, `order_number`, `status`, `items`, `payment_info`, `verification_info`, `total_amount`, `total_rights_cost`) VALUES 
(1, 1, 1, 'ORD-T1-001', 'completed', '[{"product_id": 1, "quantity": 2, "price": 99.99}]', '{"method": "wechat", "transaction_id": "wx123456"}', '{"verification_code": "ABC123", "verified_at": "2024-01-01 12:00:00"}', 199.98, 20.00),
(2, 2, 2, 'ORD-T2-001', 'pending', '[{"product_id": 3, "quantity": 1, "price": 49.99}]', '{"method": "alipay", "transaction_id": ""}', null, 49.99, 5.00)
ON DUPLICATE KEY UPDATE 
    `status` = VALUES(`status`),
    `items` = VALUES(`items`),
    `payment_info` = VALUES(`payment_info`),
    `verification_info` = VALUES(`verification_info`),
    `total_amount` = VALUES(`total_amount`),
    `total_rights_cost` = VALUES(`total_rights_cost`);