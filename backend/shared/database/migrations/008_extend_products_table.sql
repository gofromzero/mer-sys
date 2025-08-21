-- 扩展商品表以支持分类、标签、状态、版本等字段
-- 修改价格存储格式以支持 Money 类型
-- 添加图片管理和版本控制字段

-- 首先添加新的字段
ALTER TABLE `products` 
ADD COLUMN `category_id` BIGINT UNSIGNED NULL COMMENT '商品分类ID' AFTER `description`,
ADD COLUMN `category_path` VARCHAR(500) NULL COMMENT '分类层级路径' AFTER `category_id`,
ADD COLUMN `tags` JSON NULL COMMENT '商品标签数组' AFTER `category_path`,
ADD COLUMN `price` JSON NOT NULL COMMENT '价格信息JSON(amount, currency)' AFTER `tags`,
ADD COLUMN `images` JSON NULL COMMENT '商品图片信息' AFTER `inventory_info`,
ADD COLUMN `version` INT NOT NULL DEFAULT 1 COMMENT '商品版本号' AFTER `images`,
ADD INDEX `idx_category` (`category_id`),
ADD INDEX `idx_status_merchant` (`status`, `merchant_id`),
ADD INDEX `idx_version` (`version`);

-- 更新状态枚举值，添加 'deleted' 状态，移除 'archived'
ALTER TABLE `products` 
MODIFY COLUMN `status` ENUM('draft', 'active', 'inactive', 'deleted') NOT NULL DEFAULT 'draft' COMMENT '商品状态';

-- 迁移现有数据到新格式
-- 将 price_amount 和 price_currency 合并到 price JSON 字段
UPDATE `products` 
SET `price` = JSON_OBJECT('amount', CAST(`price_amount` * 100 AS SIGNED), 'currency', IFNULL(`price_currency`, 'CNY'))
WHERE `price` IS NULL OR JSON_VALID(`price`) = 0;

-- 删除旧的价格字段
ALTER TABLE `products` 
DROP COLUMN `price_amount`,
DROP COLUMN `price_currency`;

-- 更新 rights_cost 为分为单位 (乘以100)
UPDATE `products` 
SET `rights_cost` = `rights_cost` * 100
WHERE `rights_cost` IS NOT NULL;

-- 修改 rights_cost 字段类型为 BIGINT
ALTER TABLE `products` 
MODIFY COLUMN `rights_cost` BIGINT NOT NULL DEFAULT 0 COMMENT '权益成本(分为单位)';