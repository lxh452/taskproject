-- 角色表
CREATE TABLE `role` (
    `id` VARCHAR(32) NOT NULL COMMENT '角色id',
    `company_id` VARCHAR(32) NOT NULL COMMENT '公司id',
    `role_name` VARCHAR(50) NOT NULL COMMENT '角色名称',
    `role_code` VARCHAR(30) NOT NULL COMMENT '角色编码',
    `role_description` TEXT COMMENT '角色描述',
    `is_system` TINYINT NOT NULL DEFAULT 0 COMMENT '是否系统角色 0-否 1-是',
    `permissions` TEXT COMMENT '权限列表',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 0-禁用 1-正常',
    `create_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `delete_time` TIMESTAMP NULL COMMENT '删除时间',
    
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_role_company_code` (`company_id`, `role_code`),
    KEY `idx_role_company` (`company_id`),
    KEY `idx_role_name` (`role_name`),
    KEY `idx_role_system` (`is_system`),
    KEY `idx_role_status` (`status`),
    KEY `idx_create_time` (`create_time`)
    
    -- 外键约束在init.sql中统一添加
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

-- 职位角色关联表（职位通过角色获得权限，员工通过职位获得权限）
CREATE TABLE `position_role` (
    `id` VARCHAR(32) NOT NULL COMMENT '关联id',
    `position_id` VARCHAR(32) NOT NULL COMMENT '职位id',
    `role_id` VARCHAR(32) NOT NULL COMMENT '角色id',
    `grant_by` VARCHAR(32) COMMENT '授权人id',
    `grant_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '授权时间',
    `expire_time` TIMESTAMP NULL COMMENT '过期时间',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 0-禁用 1-正常',
    `create_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_position_role` (`position_id`, `role_id`),
    KEY `idx_position_role_position` (`position_id`),
    KEY `idx_position_role_role` (`role_id`),
    KEY `idx_position_role_grant_by` (`grant_by`),
    KEY `idx_position_role_status` (`status`),
    KEY `idx_grant_time` (`grant_time`)
    
    -- 外键约束在init.sql中统一添加
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='职位角色关联表';

-- 系统配置表
CREATE TABLE `system_config` (
    `id` VARCHAR(32) NOT NULL COMMENT '配置id',
    `config_key` VARCHAR(100) NOT NULL COMMENT '配置键',
    `config_value` TEXT COMMENT '配置值',
    `config_type` TINYINT NOT NULL DEFAULT 0 COMMENT '配置类型 0-字符串 1-数字 2-布尔 3-JSON',
    `config_group` VARCHAR(50) COMMENT '配置分组',
    `description` VARCHAR(200) COMMENT '配置描述',
    `is_system` TINYINT NOT NULL DEFAULT 0 COMMENT '是否系统配置 0-否 1-是',
    `is_encrypted` TINYINT NOT NULL DEFAULT 0 COMMENT '是否加密 0-否 1-是',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 0-禁用 1-正常',
    `create_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_config_key` (`config_key`),
    KEY `idx_config_group` (`config_group`),
    KEY `idx_config_type` (`config_type`),
    KEY `idx_config_system` (`is_system`),
    KEY `idx_config_status` (`status`),
    KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';

-- 操作日志表
CREATE TABLE `operation_log` (
    `id` VARCHAR(32) NOT NULL COMMENT '日志id',
    `user_id` VARCHAR(32) COMMENT '操作用户id',
    `employee_id` VARCHAR(32) COMMENT '操作员工id',
    `operation_type` VARCHAR(50) NOT NULL COMMENT '操作类型',
    `operation_name` VARCHAR(100) NOT NULL COMMENT '操作名称',
    `operation_desc` TEXT COMMENT '操作描述',
    `request_method` VARCHAR(10) COMMENT '请求方法',
    `request_url` VARCHAR(500) COMMENT '请求URL',
    `request_params` TEXT COMMENT '请求参数',
    `response_data` TEXT COMMENT '响应数据',
    `ip_address` VARCHAR(45) COMMENT 'IP地址',
    `user_agent` VARCHAR(500) COMMENT '用户代理',
    `execution_time` INT COMMENT '执行时间(毫秒)',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 0-失败 1-成功',
    `error_message` TEXT COMMENT '错误信息',
    `create_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    PRIMARY KEY (`id`),
    KEY `idx_log_user` (`user_id`),
    KEY `idx_log_employee` (`employee_id`),
    KEY `idx_log_type` (`operation_type`),
    KEY `idx_log_status` (`status`),
    KEY `idx_log_ip` (`ip_address`),
    KEY `idx_create_time` (`create_time`)
    
    -- 外键约束在init.sql中统一添加
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作日志表';
