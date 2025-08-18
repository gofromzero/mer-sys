-- 创建商品表
CREATE TABLE IF NOT EXISTS `products` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '商品ID',
    `tenant_id` bigint(20) unsigned NOT NULL COMMENT '租户ID',
    `merchant_id` bigint(20) unsigned NOT NULL COMMENT '商户ID',
    `name` varchar(200) NOT NULL COMMENT '商品名称',
    `description` text COMMENT '商品描述',
    `price_amount` decimal(10,2) NOT NULL COMMENT '价格金额',
    `price_currency` varchar(3) DEFAULT 'CNY' COMMENT '货币类型',
    `rights_cost` decimal(10,2) NOT NULL COMMENT '权益成本',
    `inventory_info` json COMMENT '库存信息',
    `status` enum('draft','active','inactive','archived') NOT NULL DEFAULT 'draft' COMMENT '商品状态',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_product_tenant` (`tenant_id`),
    KEY `idx_product_merchant` (`merchant_id`),
    KEY `idx_product_status` (`status`),
    KEY `idx_product_created` (`created_at`),
    KEY `idx_product_tenant_merchant` (`tenant_id`, `merchant_id`),
    CONSTRAINT `fk_product_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT `fk_product_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品表';

-- 插入测试商品数据
INSERT INTO `products` (`tenant_id`, `merchant_id`, `name`, `description`, `price_amount`, `price_currency`, `rights_cost`, `inventory_info`, `status`) VALUES 
(1, 1, '测试商品A', '这是租户1的测试商品A', 99.99, 'CNY', 10.00, '{"stock_quantity": 100, "reserved_quantity": 0, "track_inventory": true}', 'active'),
(1, 1, '测试商品B', '这是租户1的测试商品B', 199.99, 'CNY', 20.00, '{"stock_quantity": 50, "reserved_quantity": 0, "track_inventory": true}', 'active'),
(2, 2, '演示商品A', '这是租户2的演示商品A', 49.99, 'CNY', 5.00, '{"stock_quantity": 10, "reserved_quantity": 0, "track_inventory": true}', 'active')
ON DUPLICATE KEY UPDATE 
    `name` = VALUES(`name`),
    `description` = VALUES(`description`),
    `price_amount` = VALUES(`price_amount`),
    `rights_cost` = VALUES(`rights_cost`),
    `inventory_info` = VALUES(`inventory_info`),
    `status` = VALUES(`status`);