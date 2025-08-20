-- 扩展商户表以支持审批流程
ALTER TABLE `merchants` 
ADD COLUMN `registration_time` timestamp NULL COMMENT '注册申请时间',
ADD COLUMN `approval_time` timestamp NULL COMMENT '审批时间',
ADD COLUMN `approved_by` bigint(20) unsigned NULL COMMENT '审批人ID',
ADD KEY `idx_merchant_registration` (`registration_time`),
ADD KEY `idx_merchant_approval` (`approval_time`),
ADD KEY `idx_merchant_approved_by` (`approved_by`);

-- 更新现有数据，设置注册时间为创建时间
UPDATE `merchants` SET `registration_time` = `created_at` WHERE `registration_time` IS NULL;

-- 对于已激活的商户，设置审批时间为创建时间（假设立即审批）
UPDATE `merchants` SET `approval_time` = `created_at` WHERE `status` = 'active' AND `approval_time` IS NULL;