-- 创建用户表
CREATE TABLE IF NOT EXISTS `users` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '用户ID',
    `uuid` varchar(36) NOT NULL UNIQUE COMMENT '用户UUID',
    `username` varchar(50) NOT NULL COMMENT '用户名',
    `email` varchar(100) NOT NULL COMMENT '邮箱',
    `phone` varchar(20) DEFAULT NULL COMMENT '手机号',
    `password_hash` varchar(255) NOT NULL COMMENT '密码哈希',
    `tenant_id` bigint(20) unsigned NOT NULL COMMENT '租户ID',
    `status` enum('pending','active','suspended','deactivated') NOT NULL DEFAULT 'pending' COMMENT '用户状态',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `last_login_at` timestamp NULL DEFAULT NULL COMMENT '最后登录时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_uuid` (`uuid`),
    UNIQUE KEY `uk_user_tenant_username` (`tenant_id`, `username`),
    UNIQUE KEY `uk_user_tenant_email` (`tenant_id`, `email`),
    KEY `idx_user_tenant` (`tenant_id`),
    KEY `idx_user_status` (`status`),
    KEY `idx_user_email` (`email`),
    KEY `idx_user_created` (`created_at`),
    CONSTRAINT `fk_user_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 创建用户角色表
CREATE TABLE IF NOT EXISTS `user_roles` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '角色ID',
    `user_id` bigint(20) unsigned NOT NULL COMMENT '用户ID',
    `role_type` varchar(50) NOT NULL COMMENT '角色类型',
    `resource_id` bigint(20) unsigned DEFAULT NULL COMMENT '资源ID(可选)',
    `tenant_id` bigint(20) unsigned NOT NULL COMMENT '租户ID',
    `permissions` json COMMENT '权限列表',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_role_resource` (`user_id`, `role_type`, `resource_id`),
    KEY `idx_user_role_tenant` (`tenant_id`),
    KEY `idx_user_role_type` (`role_type`),
    CONSTRAINT `fk_user_role_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_user_role_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色表';

-- 插入默认管理员用户（密码: admin123）
INSERT INTO `users` (`uuid`, `username`, `email`, `password_hash`, `tenant_id`, `status`) VALUES 
('550e8400-e29b-41d4-a716-446655440000', 'admin', 'admin@example.com', '$2y$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 1, 'active'),
('550e8400-e29b-41d4-a716-446655440001', 'demo', 'demo@example.com', '$2y$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 2, 'active')
ON DUPLICATE KEY UPDATE 
    `username` = VALUES(`username`),
    `email` = VALUES(`email`),
    `status` = VALUES(`status`);

-- 插入默认用户角色
INSERT INTO `user_roles` (`user_id`, `role_type`, `tenant_id`, `permissions`) VALUES 
(1, 'admin', 1, '["user.create", "user.read", "user.update", "user.delete", "tenant.manage", "system.admin"]'),
(2, 'admin', 2, '["user.create", "user.read", "user.update", "user.delete", "tenant.manage"]')
ON DUPLICATE KEY UPDATE 
    `permissions` = VALUES(`permissions`);