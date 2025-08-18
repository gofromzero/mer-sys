-- 创建商户表
CREATE TABLE IF NOT EXISTS `merchants` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '商户ID',
    `tenant_id` bigint(20) unsigned NOT NULL COMMENT '租户ID',
    `name` varchar(100) NOT NULL COMMENT '商户名称',
    `code` varchar(50) NOT NULL COMMENT '商户代码',
    `status` enum('pending','active','suspended','deactivated') NOT NULL DEFAULT 'pending' COMMENT '商户状态',
    `business_info` json COMMENT '商户业务信息',
    `rights_balance` json COMMENT '权益余额信息',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_merchant_tenant_code` (`tenant_id`, `code`),
    KEY `idx_merchant_tenant` (`tenant_id`),
    KEY `idx_merchant_status` (`status`),
    KEY `idx_merchant_created` (`created_at`),
    CONSTRAINT `fk_merchant_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商户表';

-- 插入测试商户数据
INSERT INTO `merchants` (`tenant_id`, `name`, `code`, `status`, `business_info`, `rights_balance`) VALUES 
(1, '默认商户', 'default-merchant', 'active', '{"type": "retail", "category": "general"}', '{"total_balance": 10000.00, "used_balance": 0.00, "frozen_balance": 0.00}'),
(2, '演示商户', 'demo-merchant', 'active', '{"type": "demo", "category": "test"}', '{"total_balance": 1000.00, "used_balance": 0.00, "frozen_balance": 0.00}')
ON DUPLICATE KEY UPDATE 
    `name` = VALUES(`name`),
    `status` = VALUES(`status`),
    `business_info` = VALUES(`business_info`),
    `rights_balance` = VALUES(`rights_balance`);