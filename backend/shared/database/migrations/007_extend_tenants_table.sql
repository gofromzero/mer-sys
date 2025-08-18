-- 扩展租户表，添加业务相关字段
ALTER TABLE `tenants` 
ADD COLUMN `business_type` varchar(50) COMMENT '业务类型' AFTER `config`,
ADD COLUMN `contact_person` varchar(100) COMMENT '联系人' AFTER `business_type`,
ADD COLUMN `contact_email` varchar(100) COMMENT '联系邮箱' AFTER `contact_person`,
ADD COLUMN `contact_phone` varchar(20) COMMENT '联系电话' AFTER `contact_email`,
ADD COLUMN `address` text COMMENT '地址信息' AFTER `contact_phone`,
ADD COLUMN `registration_time` timestamp NULL COMMENT '注册时间' AFTER `address`,
ADD COLUMN `activation_time` timestamp NULL COMMENT '激活时间' AFTER `registration_time`;

-- 添加相应的索引
ALTER TABLE `tenants`
ADD INDEX `idx_tenant_business_type` (`business_type`),
ADD INDEX `idx_tenant_contact_email` (`contact_email`),
ADD INDEX `idx_tenant_registration_time` (`registration_time`);