-- 003_create_user_roles_table.sql
-- 创建用户角色关联表，支持多租户隔离的RBAC权限模型

-- 用户角色关联表
CREATE TABLE `user_roles` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `tenant_id` BIGINT UNSIGNED NOT NULL COMMENT '租户ID',
    `role_type` VARCHAR(50) NOT NULL COMMENT '角色类型：tenant_admin, merchant, customer',
    `resource_id` BIGINT UNSIGNED NULL COMMENT '资源ID（可选，用于细粒度权限控制）',
    `resource_type` VARCHAR(50) NULL COMMENT '资源类型（可选，如merchant, product等）',
    `granted_by` BIGINT UNSIGNED NOT NULL COMMENT '授权者用户ID',
    `expires_at` TIMESTAMP NULL COMMENT '权限过期时间（可选）',
    `status` ENUM('active', 'suspended', 'expired') NOT NULL DEFAULT 'active' COMMENT '角色状态',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    PRIMARY KEY (`id`),
    
    -- 唯一约束：同一租户下的用户对于特定资源只能有一个角色
    UNIQUE KEY `uk_user_tenant_role_resource` (`user_id`, `tenant_id`, `role_type`, `resource_id`),
    
    -- 索引优化
    KEY `idx_user_tenant` (`user_id`, `tenant_id`),
    KEY `idx_tenant_role` (`tenant_id`, `role_type`),
    KEY `idx_user_role` (`user_id`, `role_type`),
    KEY `idx_resource` (`resource_type`, `resource_id`),
    KEY `idx_granted_by` (`granted_by`),
    KEY `idx_status_expires` (`status`, `expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色关联表';

-- 用户权限缓存表（可选，用于性能优化）
CREATE TABLE `user_permissions_cache` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `tenant_id` BIGINT UNSIGNED NOT NULL COMMENT '租户ID',
    `permissions_json` JSON NOT NULL COMMENT '用户权限列表JSON',
    `roles_json` JSON NOT NULL COMMENT '用户角色列表JSON',
    `version` INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '缓存版本号',
    `expires_at` TIMESTAMP NOT NULL COMMENT '缓存过期时间',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    PRIMARY KEY (`id`),
    
    -- 唯一约束
    UNIQUE KEY `uk_user_tenant` (`user_id`, `tenant_id`),
    
    -- 索引
    KEY `idx_expires_at` (`expires_at`),
    KEY `idx_tenant_id` (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户权限缓存表';

-- 初始化默认角色数据
INSERT INTO `user_roles` (`user_id`, `tenant_id`, `role_type`, `granted_by`, `status`) VALUES
-- 为租户1创建默认管理员角色（假设用户ID 1存在）
(1, 1, 'tenant_admin', 1, 'active'),
-- 为租户1创建测试商户角色（假设用户ID 2存在）  
(2, 1, 'merchant', 1, 'active'),
-- 为租户1创建测试客户角色（假设用户ID 3存在）
(3, 1, 'customer', 1, 'active');