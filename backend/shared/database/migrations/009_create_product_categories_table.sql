-- 创建商品分类表，支持多级分类体系
CREATE TABLE IF NOT EXISTS `product_categories` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '分类ID',
    `tenant_id` BIGINT UNSIGNED NOT NULL COMMENT '租户ID',
    `name` VARCHAR(100) NOT NULL COMMENT '分类名称',
    `parent_id` BIGINT UNSIGNED NULL COMMENT '父分类ID',
    `level` TINYINT NOT NULL DEFAULT 1 COMMENT '分类层级',
    `path` VARCHAR(500) NOT NULL COMMENT '完整路径',
    `sort_order` INT DEFAULT 0 COMMENT '排序顺序',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态(1:启用,0:禁用)',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX `idx_tenant_parent` (`tenant_id`, `parent_id`),
    INDEX `idx_path` (`path`),
    INDEX `idx_level` (`level`),
    INDEX `idx_status` (`status`),
    
    CONSTRAINT `fk_category_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT `fk_category_parent` FOREIGN KEY (`parent_id`) REFERENCES `product_categories` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品分类表';

-- 在商品表中添加外键约束
ALTER TABLE `products` 
ADD CONSTRAINT `fk_product_category` FOREIGN KEY (`category_id`) REFERENCES `product_categories` (`id`) ON DELETE SET NULL ON UPDATE CASCADE;

-- 插入默认分类数据
INSERT INTO `product_categories` (`tenant_id`, `name`, `parent_id`, `level`, `path`, `sort_order`) VALUES 
(1, '电子产品', NULL, 1, '电子产品', 1),
(1, '服装配饰', NULL, 1, '服装配饰', 2),
(1, '生活用品', NULL, 1, '生活用品', 3),
(1, '手机数码', 1, 2, '电子产品/手机数码', 1),
(1, '电脑办公', 1, 2, '电子产品/电脑办公', 2),
(1, '男装', 2, 2, '服装配饰/男装', 1),
(1, '女装', 2, 2, '服装配饰/女装', 2),
(2, '数码产品', NULL, 1, '数码产品', 1),
(2, '日用百货', NULL, 1, '日用百货', 2)
ON DUPLICATE KEY UPDATE 
    `name` = VALUES(`name`),
    `path` = VALUES(`path`),
    `sort_order` = VALUES(`sort_order`);

-- 为现有测试商品添加分类
UPDATE `products` SET `category_id` = 4, `category_path` = '电子产品/手机数码' WHERE `tenant_id` = 1 AND `merchant_id` = 1 AND `name` LIKE '%测试商品A%';
UPDATE `products` SET `category_id` = 5, `category_path` = '电子产品/电脑办公' WHERE `tenant_id` = 1 AND `merchant_id` = 1 AND `name` LIKE '%测试商品B%';
UPDATE `products` SET `category_id` = 8, `category_path` = '数码产品' WHERE `tenant_id` = 2 AND `merchant_id` = 2;