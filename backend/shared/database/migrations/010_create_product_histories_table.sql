-- 创建商品变更历史表
CREATE TABLE IF NOT EXISTS `product_histories` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '历史记录ID',
    `tenant_id` BIGINT UNSIGNED NOT NULL COMMENT '租户ID',
    `product_id` BIGINT UNSIGNED NOT NULL COMMENT '商品ID',
    `version` INT NOT NULL COMMENT '版本号',
    `field_name` VARCHAR(100) NOT NULL COMMENT '变更字段名',
    `old_value` TEXT NULL COMMENT '变更前值',
    `new_value` TEXT NULL COMMENT '变更后值',
    `operation` VARCHAR(20) NOT NULL COMMENT '操作类型',
    `changed_by` BIGINT UNSIGNED NOT NULL COMMENT '变更人用户ID',
    `changed_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '变更时间',
    
    INDEX `idx_tenant_product` (`tenant_id`, `product_id`),
    INDEX `idx_version` (`version`),
    INDEX `idx_changed_at` (`changed_at`),
    INDEX `idx_operation` (`operation`),
    INDEX `idx_changed_by` (`changed_by`),
    
    CONSTRAINT `fk_history_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT `fk_history_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT `fk_history_user` FOREIGN KEY (`changed_by`) REFERENCES `users` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品变更历史表';

-- 为现有商品创建初始历史记录
INSERT INTO `product_histories` (`tenant_id`, `product_id`, `version`, `field_name`, `old_value`, `new_value`, `operation`, `changed_by`)
SELECT 
    p.`tenant_id`,
    p.`id`,
    p.`version`,
    'initial_creation',
    NULL,
    CONCAT('商品"', p.`name`, '"创建'),
    'create',
    1  -- 假设系统用户ID为1，实际应该根据需求调整
FROM `products` p
WHERE NOT EXISTS (
    SELECT 1 FROM `product_histories` ph 
    WHERE ph.`product_id` = p.`id` AND ph.`operation` = 'create'
);