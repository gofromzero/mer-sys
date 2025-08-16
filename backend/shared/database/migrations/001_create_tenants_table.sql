-- 创建租户表
CREATE TABLE IF NOT EXISTS `tenants` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '租户ID',
    `name` varchar(100) NOT NULL COMMENT '租户名称',
    `code` varchar(50) NOT NULL UNIQUE COMMENT '租户代码',
    `status` enum('active','suspended','expired') NOT NULL DEFAULT 'active' COMMENT '租户状态',
    `config` json COMMENT '租户配置',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_tenant_code` (`code`),
    KEY `idx_tenant_status` (`status`),
    KEY `idx_tenant_created` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='租户表';

-- 插入默认租户数据
INSERT INTO `tenants` (`name`, `code`, `status`, `config`) VALUES 
('默认租户', 'default', 'active', '{"max_users": 100, "max_merchants": 50, "features": ["basic"], "settings": {}}'),
('演示租户', 'demo', 'active', '{"max_users": 10, "max_merchants": 5, "features": ["basic", "demo"], "settings": {"demo_mode": "true"}}')
ON DUPLICATE KEY UPDATE 
    `name` = VALUES(`name`),
    `status` = VALUES(`status`),
    `config` = VALUES(`config`);