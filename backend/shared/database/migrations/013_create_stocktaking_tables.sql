-- 库存盘点表迁移
-- Migration: 013_create_stocktaking_tables.sql
-- Description: 创建库存盘点相关表

-- 库存盘点任务表
CREATE TABLE IF NOT EXISTS inventory_stocktaking (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT UNSIGNED NOT NULL COMMENT '租户ID',
    name VARCHAR(255) NOT NULL COMMENT '盘点名称',
    description TEXT COMMENT '盘点说明',
    status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '盘点状态: pending, in_progress, completed, cancelled',
    product_ids JSON COMMENT '盘点商品ID列表，空则全量盘点',
    started_by BIGINT UNSIGNED NOT NULL COMMENT '发起人ID',
    completed_by BIGINT UNSIGNED NULL COMMENT '完成人ID',
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '开始时间',
    completed_at TIMESTAMP NULL COMMENT '完成时间',
    summary TEXT COMMENT '盘点总结',
    notes TEXT COMMENT '备注信息',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX idx_tenant_status (tenant_id, status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='库存盘点任务表';

-- 库存盘点记录表
CREATE TABLE IF NOT EXISTS inventory_stocktaking_records (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT UNSIGNED NOT NULL COMMENT '租户ID',
    stocktaking_id BIGINT UNSIGNED NOT NULL COMMENT '盘点任务ID',
    product_id BIGINT UNSIGNED NOT NULL COMMENT '商品ID',
    system_count INT NOT NULL COMMENT '系统库存数量',
    actual_count INT NOT NULL COMMENT '实际盘点数量',
    difference INT NOT NULL COMMENT '差异数量 (actual - system)',
    reason VARCHAR(500) COMMENT '差异原因',
    checked_by BIGINT UNSIGNED NOT NULL COMMENT '盘点人ID',
    checked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '盘点时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX idx_tenant_stocktaking (tenant_id, stocktaking_id),
    INDEX idx_product (product_id),
    INDEX idx_checked_at (checked_at),
    
    FOREIGN KEY (stocktaking_id) REFERENCES inventory_stocktaking(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='库存盘点记录表';

-- 添加审计日志表的索引优化（如果表已存在）
ALTER TABLE inventory_audit_logs 
ADD INDEX IF NOT EXISTS idx_audit_type_level (audit_type, level),
ADD INDEX IF NOT EXISTS idx_operation_time (operation_type, created_at);

-- 添加库存记录表的复合索引优化（如果表已存在）
ALTER TABLE inventory_records 
ADD INDEX IF NOT EXISTS idx_tenant_product_time (tenant_id, product_id, created_at),
ADD INDEX IF NOT EXISTS idx_change_type_time (change_type, created_at);

-- 库存预警表索引优化（如果表已存在）
ALTER TABLE inventory_alerts
ADD INDEX IF NOT EXISTS idx_tenant_active (tenant_id, is_active),
ADD INDEX IF NOT EXISTS idx_alert_type_active (alert_type, is_active);

-- 插入示例盘点任务数据（可选，用于测试）
INSERT INTO inventory_stocktaking (tenant_id, name, description, status, started_by) VALUES
(1, '2024年第三季度全面盘点', '针对所有商品进行全面库存盘点', 'completed', 1),
(1, '高价值商品专项盘点', '针对单价超过1000元的商品进行专项盘点', 'in_progress', 1),
(1, '季末库存清查', '季末例行库存清查和盘点', 'pending', 1);

-- 创建库存统计视图（可选，用于快速查询统计信息）
CREATE OR REPLACE VIEW inventory_statistics_view AS
SELECT 
    p.tenant_id,
    COUNT(*) as total_products,
    SUM(CASE WHEN p.inventory_info->>'$.track_inventory' = 'true' 
             AND CAST(p.inventory_info->>'$.stock_quantity' AS SIGNED) <= COALESCE(CAST(p.inventory_info->>'$.low_stock_threshold' AS SIGNED), 10) 
        THEN 1 ELSE 0 END) as low_stock_products,
    SUM(CASE WHEN p.inventory_info->>'$.track_inventory' = 'true' 
             AND CAST(p.inventory_info->>'$.stock_quantity' AS SIGNED) <= 0 
        THEN 1 ELSE 0 END) as out_of_stock_products,
    SUM(CASE WHEN p.inventory_info->>'$.track_inventory' = 'true'
        THEN CAST(p.inventory_info->>'$.stock_quantity' AS SIGNED) * p.price_amount / 100.0 
        ELSE 0 END) as total_inventory_value,
    NOW() as last_updated
FROM products p
WHERE p.status = 'active'
GROUP BY p.tenant_id;

-- 创建活跃预警统计视图
CREATE OR REPLACE VIEW active_alerts_view AS
SELECT 
    ia.tenant_id,
    COUNT(*) as total_active_alerts,
    SUM(CASE WHEN ia.alert_type = 'low_stock' THEN 1 ELSE 0 END) as low_stock_alerts,
    SUM(CASE WHEN ia.alert_type = 'out_of_stock' THEN 1 ELSE 0 END) as out_of_stock_alerts,
    SUM(CASE WHEN ia.alert_type = 'overstock' THEN 1 ELSE 0 END) as overstock_alerts
FROM inventory_alerts ia
WHERE ia.is_active = 1
GROUP BY ia.tenant_id;

-- 添加注释说明
ALTER TABLE inventory_stocktaking COMMENT = '库存盘点任务表 - 管理库存盘点的生命周期和基本信息';
ALTER TABLE inventory_stocktaking_records COMMENT = '库存盘点记录表 - 记录每个商品的具体盘点结果和差异信息';

-- 完成提示
SELECT 'Stocktaking tables created successfully!' as status;