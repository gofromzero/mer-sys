-- 017_create_scheduled_tasks_table.sql
-- 创建定时任务表

DROP TABLE IF EXISTS `scheduled_tasks`;

CREATE TABLE `scheduled_tasks` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '任务ID',
  `tenant_id` bigint NOT NULL COMMENT '租户ID',
  `task_name` varchar(100) NOT NULL COMMENT '任务名称',
  `task_description` text COMMENT '任务描述',
  `report_type` varchar(50) NOT NULL COMMENT '报表类型',
  `cron_expression` varchar(100) NOT NULL COMMENT 'Cron表达式',
  `report_config` json NOT NULL COMMENT '报表配置(JSON格式)',
  `recipients` json NOT NULL COMMENT '接收人列表(JSON数组)',
  `is_enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用(1:启用 0:禁用)',
  `last_run_time` datetime DEFAULT NULL COMMENT '最后执行时间',
  `last_run_status` varchar(20) DEFAULT 'pending' COMMENT '最后执行状态(pending/running/completed/failed)',
  `last_run_message` text COMMENT '最后执行结果信息',
  `next_run_time` datetime DEFAULT NULL COMMENT '下次运行时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_report_type` (`report_type`),
  KEY `idx_is_enabled` (`is_enabled`),
  KEY `idx_next_run_time` (`next_run_time`),
  KEY `idx_last_run_status` (`last_run_status`),
  KEY `idx_created_at` (`created_at`),
  CONSTRAINT `fk_scheduled_tasks_tenant_id` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='定时任务表';

-- 插入一些示例定时任务
INSERT INTO `scheduled_tasks` (
  `tenant_id`, 
  `task_name`, 
  `task_description`, 
  `report_type`, 
  `cron_expression`, 
  `report_config`, 
  `recipients`, 
  `is_enabled`,
  `next_run_time`
) VALUES 
(
  1, 
  '每日财务报表', 
  '每天上午9点生成前一天的财务报表并发送给财务团队', 
  'financial', 
  '0 9 * * *',
  '{"start_date": "yesterday", "end_date": "yesterday", "format": "excel", "filters": {}}',
  '["finance@company.com", "manager@company.com"]',
  1,
  DATE_ADD(CURRENT_TIMESTAMP, INTERVAL 1 DAY)
),
(
  1, 
  '周度商户分析报告', 
  '每周一上午9点生成上周的商户运营分析报告', 
  'merchant_operation', 
  '0 9 * * 1',
  '{"start_date": "last_week_start", "end_date": "last_week_end", "format": "pdf", "filters": {}}',
  '["operations@company.com", "sales@company.com"]',
  1,
  DATE_ADD(CURRENT_TIMESTAMP, INTERVAL 7 DAY)
),
(
  1, 
  '月度客户行为报表', 
  '每月1号上午9点生成上月的客户行为分析报告', 
  'customer_behavior', 
  '0 9 1 * *',
  '{"start_date": "last_month_start", "end_date": "last_month_end", "format": "excel", "filters": {"include_retention_analysis": true}}',
  '["marketing@company.com", "product@company.com"]',
  1,
  DATE_ADD(CURRENT_TIMESTAMP, INTERVAL 1 MONTH)
),
(
  2, 
  '每小时系统监控报告', 
  '每小时生成系统运行状态报告（仅测试租户）', 
  'financial', 
  '0 * * * *',
  '{"start_date": "1_hour_ago", "end_date": "now", "format": "json", "filters": {"quick_summary": true}}',
  '["admin@test.com"]',
  0,
  DATE_ADD(CURRENT_TIMESTAMP, INTERVAL 1 HOUR)
);