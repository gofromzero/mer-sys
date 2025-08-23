-- 创建商品定价规则表
CREATE TABLE product_pricing_rules (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT UNSIGNED NOT NULL,
    product_id BIGINT UNSIGNED NOT NULL,
    rule_type VARCHAR(50) NOT NULL COMMENT '定价规则类型：base_price, volume_discount, member_discount, time_based_discount',
    rule_config JSON NOT NULL COMMENT '规则配置JSON',
    priority INT DEFAULT 0 COMMENT '优先级，数值越大优先级越高',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    valid_from TIMESTAMP NOT NULL COMMENT '生效时间',
    valid_until TIMESTAMP NULL COMMENT '失效时间，NULL表示永久有效',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_tenant_product (tenant_id, product_id),
    INDEX idx_rule_type (rule_type),
    INDEX idx_active_valid (is_active, valid_from, valid_until),
    INDEX idx_priority (priority DESC),
    
    CONSTRAINT fk_pricing_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE,
    
    -- 确保每个商品的基础价格规则只能有一个
    UNIQUE INDEX uk_tenant_product_base_price (tenant_id, product_id, rule_type) 
        WHERE rule_type = 'base_price'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品定价规则表';

-- 创建商品权益规则表
CREATE TABLE product_rights_rules (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT UNSIGNED NOT NULL,
    product_id BIGINT UNSIGNED NOT NULL,
    rule_type VARCHAR(50) NOT NULL COMMENT '权益规则类型：fixed_rate, percentage, tiered',
    consumption_rate DECIMAL(10,4) NOT NULL COMMENT '消耗比例',
    min_rights_required DECIMAL(10,2) DEFAULT 0 COMMENT '最低权益要求',
    insufficient_rights_action VARCHAR(50) NOT NULL COMMENT '权益不足处理策略：block_purchase, partial_payment, cash_payment',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_tenant_product (tenant_id, product_id),
    INDEX idx_rule_type (rule_type),
    INDEX idx_active (is_active),
    
    CONSTRAINT fk_rights_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE,
    
    -- 确保每个商品只能有一个权益规则
    UNIQUE INDEX uk_tenant_product_rights (tenant_id, product_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品权益消耗规则表';

-- 创建促销价格表
CREATE TABLE promotional_prices (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT UNSIGNED NOT NULL,
    product_id BIGINT UNSIGNED NOT NULL,
    promotional_price JSON NOT NULL COMMENT '促销价格JSON',
    discount_percentage DECIMAL(5,2) NULL COMMENT '折扣百分比',
    valid_from TIMESTAMP NOT NULL COMMENT '促销开始时间',
    valid_until TIMESTAMP NOT NULL COMMENT '促销结束时间',
    conditions JSON NULL COMMENT '促销条件JSON',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_tenant_product (tenant_id, product_id),
    INDEX idx_valid_period (valid_from, valid_until),
    INDEX idx_active (is_active),
    INDEX idx_created_at (created_at),
    
    CONSTRAINT fk_promo_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE,
    
    -- 确保促销时间段不重叠的约束通过应用层实现
    INDEX idx_tenant_product_period (tenant_id, product_id, valid_from, valid_until)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品促销价格表';

-- 创建价格变更历史表
CREATE TABLE price_histories (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT UNSIGNED NOT NULL,
    product_id BIGINT UNSIGNED NOT NULL,
    old_price JSON NOT NULL COMMENT '旧价格JSON',
    new_price JSON NOT NULL COMMENT '新价格JSON',
    change_reason VARCHAR(255) NOT NULL COMMENT '变更原因',
    changed_by BIGINT UNSIGNED NOT NULL COMMENT '变更人ID',
    effective_date TIMESTAMP NOT NULL COMMENT '生效时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_tenant_product (tenant_id, product_id),
    INDEX idx_effective_date (effective_date),
    INDEX idx_changed_by (changed_by),
    INDEX idx_created_at (created_at),
    
    CONSTRAINT fk_price_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE,
    CONSTRAINT fk_price_changed_by FOREIGN KEY (changed_by) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='价格变更历史表';

-- 为现有商品创建默认基础定价规则的存储过程
DELIMITER //
CREATE PROCEDURE MigrateExistingProductPricing()
BEGIN
    DECLARE done INT DEFAULT FALSE;
    DECLARE v_tenant_id BIGINT UNSIGNED;
    DECLARE v_product_id BIGINT UNSIGNED;
    DECLARE v_price_amount DECIMAL(10,2);
    DECLARE v_price_currency VARCHAR(3);
    DECLARE v_created_at TIMESTAMP;
    
    -- 定义游标遍历现有商品
    DECLARE product_cursor CURSOR FOR 
        SELECT tenant_id, id, price_amount, price_currency, created_at 
        FROM products 
        WHERE price_amount > 0;
    
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = TRUE;
    
    -- 开始事务
    START TRANSACTION;
    
    OPEN product_cursor;
    
    read_loop: LOOP
        FETCH product_cursor INTO v_tenant_id, v_product_id, v_price_amount, v_price_currency, v_created_at;
        IF done THEN
            LEAVE read_loop;
        END IF;
        
        -- 为现有商品创建基础定价规则
        INSERT INTO product_pricing_rules (
            tenant_id, 
            product_id, 
            rule_type, 
            rule_config, 
            priority, 
            is_active, 
            valid_from
        ) VALUES (
            v_tenant_id,
            v_product_id,
            'base_price',
            JSON_OBJECT(
                'type', 'base_price',
                'config', JSON_OBJECT(
                    'amount', CAST(v_price_amount * 100 AS UNSIGNED),
                    'currency', v_price_currency
                )
            ),
            0,
            TRUE,
            v_created_at
        );
        
        -- 创建价格变更历史基线记录
        INSERT INTO price_histories (
            tenant_id,
            product_id,
            old_price,
            new_price,
            change_reason,
            changed_by,
            effective_date
        ) VALUES (
            v_tenant_id,
            v_product_id,
            JSON_OBJECT('amount', 0, 'currency', v_price_currency),
            JSON_OBJECT('amount', CAST(v_price_amount * 100 AS UNSIGNED), 'currency', v_price_currency),
            '系统迁移 - 初始价格记录',
            1, -- 系统用户ID，假设为1
            v_created_at
        );
        
    END LOOP;
    
    CLOSE product_cursor;
    
    -- 提交事务
    COMMIT;
    
END //
DELIMITER ;

-- 为现有商品创建默认权益规则的存储过程
DELIMITER //
CREATE PROCEDURE MigrateExistingProductRights()
BEGIN
    DECLARE done INT DEFAULT FALSE;
    DECLARE v_tenant_id BIGINT UNSIGNED;
    DECLARE v_product_id BIGINT UNSIGNED;
    DECLARE v_rights_cost DECIMAL(10,2);
    
    -- 定义游标遍历现有商品
    DECLARE rights_cursor CURSOR FOR 
        SELECT tenant_id, id, rights_cost 
        FROM products 
        WHERE rights_cost > 0;
    
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = TRUE;
    
    -- 开始事务
    START TRANSACTION;
    
    OPEN rights_cursor;
    
    read_loop: LOOP
        FETCH rights_cursor INTO v_tenant_id, v_product_id, v_rights_cost;
        IF done THEN
            LEAVE read_loop;
        END IF;
        
        -- 为现有商品创建默认权益规则
        INSERT INTO product_rights_rules (
            tenant_id,
            product_id,
            rule_type,
            consumption_rate,
            min_rights_required,
            insufficient_rights_action,
            is_active
        ) VALUES (
            v_tenant_id,
            v_product_id,
            'fixed_rate',
            v_rights_cost,
            0,
            'block_purchase',
            TRUE
        );
        
    END LOOP;
    
    CLOSE rights_cursor;
    
    -- 提交事务
    COMMIT;
    
END //
DELIMITER ;

-- 执行数据迁移（注意：生产环境中应该在维护窗口期间执行）
-- CALL MigrateExistingProductPricing();
-- CALL MigrateExistingProductRights();

-- 清理存储过程
-- DROP PROCEDURE IF EXISTS MigrateExistingProductPricing;
-- DROP PROCEDURE IF EXISTS MigrateExistingProductRights;