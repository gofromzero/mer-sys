-- 016_create_report_tables.sql
-- 创建报表系统相关数据表

-- 报表记录表
CREATE TABLE IF NOT EXISTS reports (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    uuid CHAR(36) NOT NULL UNIQUE,
    tenant_id BIGINT UNSIGNED NOT NULL,
    report_type ENUM('financial', 'merchant_operation', 'customer_analysis') NOT NULL,
    period_type ENUM('daily', 'weekly', 'monthly', 'quarterly', 'yearly', 'custom') NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status ENUM('generating', 'completed', 'failed') NOT NULL DEFAULT 'generating',
    file_path VARCHAR(500),
    file_format ENUM('excel', 'pdf', 'json') NOT NULL DEFAULT 'excel',
    generated_by BIGINT UNSIGNED NOT NULL,
    generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NULL,
    data_summary JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_tenant_id (tenant_id),
    INDEX idx_report_type (report_type),
    INDEX idx_period (start_date, end_date),
    INDEX idx_status (status),
    INDEX idx_generated_at (generated_at),
    INDEX idx_uuid (uuid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='报表记录表';

-- 报表模板配置表
CREATE TABLE IF NOT EXISTS report_templates (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(100) NOT NULL,
    report_type ENUM('financial', 'merchant_operation', 'customer_analysis') NOT NULL,
    template_config JSON NOT NULL COMMENT '报表配置参数',
    schedule_config JSON COMMENT '调度配置',
    recipients JSON COMMENT '收件人列表',
    file_format ENUM('excel', 'pdf', 'json') NOT NULL DEFAULT 'excel',
    enabled BOOLEAN DEFAULT TRUE,
    created_by BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_tenant_id (tenant_id),
    INDEX idx_report_type (report_type),
    INDEX idx_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='报表模板配置表';

-- 报表生成任务表
CREATE TABLE IF NOT EXISTS report_jobs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT UNSIGNED NOT NULL,
    template_id BIGINT UNSIGNED NULL,
    report_id BIGINT UNSIGNED NULL,
    status ENUM('pending', 'running', 'completed', 'failed') NOT NULL DEFAULT 'pending',
    scheduled_at TIMESTAMP NOT NULL,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    error_message TEXT,
    retry_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_tenant_id (tenant_id),
    INDEX idx_status (status),
    INDEX idx_scheduled_at (scheduled_at),
    FOREIGN KEY (template_id) REFERENCES report_templates(id) ON DELETE CASCADE,
    FOREIGN KEY (report_id) REFERENCES reports(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='报表生成任务表';

-- 数据统计缓存表
CREATE TABLE IF NOT EXISTS analytics_cache (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT UNSIGNED NOT NULL,
    cache_key VARCHAR(200) NOT NULL,
    metric_type VARCHAR(50) NOT NULL,
    time_period VARCHAR(20) NOT NULL,
    data JSON NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_tenant_cache_key (tenant_id, cache_key),
    INDEX idx_expires_at (expires_at),
    UNIQUE KEY uk_cache_key (cache_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据统计缓存表';

-- 为reports表添加外键约束
ALTER TABLE reports ADD CONSTRAINT fk_reports_tenant 
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

-- 为report_templates表添加外键约束
ALTER TABLE report_templates ADD CONSTRAINT fk_templates_tenant 
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

-- 为report_jobs表添加外键约束  
ALTER TABLE report_jobs ADD CONSTRAINT fk_jobs_tenant 
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

-- 插入一些默认的报表模板
INSERT INTO report_templates (tenant_id, name, report_type, template_config, file_format, enabled, created_by) VALUES
(1, '月度财务报表', 'financial', '{"include_breakdown": true, "include_trends": true, "currency": "CNY"}', 'excel', TRUE, 1),
(1, '商户业绩排行榜', 'merchant_operation', '{"top_count": 20, "include_growth": true}', 'excel', TRUE, 1),
(1, '客户行为分析报告', 'customer_analysis', '{"include_retention": true, "include_churn": true}', 'pdf', TRUE, 1);

-- 创建索引优化查询性能
CREATE INDEX idx_orders_tenant_created_at ON orders(tenant_id, created_at);
CREATE INDEX idx_orders_tenant_merchant_created_at ON orders(tenant_id, merchant_id, created_at);
CREATE INDEX idx_orders_tenant_customer_created_at ON orders(tenant_id, customer_id, created_at);

-- 创建用于统计的视图
CREATE OR REPLACE VIEW v_daily_financial_summary AS
SELECT 
    tenant_id,
    merchant_id,
    DATE(created_at) as report_date,
    COUNT(*) as order_count,
    SUM(total_amount) as total_revenue,
    SUM(total_rights_cost) as rights_consumed,
    COUNT(DISTINCT customer_id) as customer_count,
    AVG(total_amount) as avg_order_value
FROM orders 
WHERE status IN ('completed', 'paid')
GROUP BY tenant_id, merchant_id, DATE(created_at);

CREATE OR REPLACE VIEW v_monthly_merchant_stats AS
SELECT 
    tenant_id,
    merchant_id,
    DATE_FORMAT(created_at, '%Y-%m') as report_month,
    COUNT(*) as order_count,
    SUM(total_amount) as total_revenue,
    COUNT(DISTINCT customer_id) as customer_count,
    AVG(total_amount) as avg_order_value
FROM orders 
WHERE status IN ('completed', 'paid')
GROUP BY tenant_id, merchant_id, DATE_FORMAT(created_at, '%Y-%m');

-- 添加用于清理过期数据的事件调度器
CREATE EVENT IF NOT EXISTS cleanup_expired_reports
ON SCHEDULE EVERY 1 DAY
DO
  DELETE FROM reports WHERE expires_at IS NOT NULL AND expires_at < NOW();

CREATE EVENT IF NOT EXISTS cleanup_expired_cache
ON SCHEDULE EVERY 1 HOUR  
DO
  DELETE FROM analytics_cache WHERE expires_at < NOW();