-- 管理员表
CREATE TABLE `admin` (
    `id` VARCHAR(32) NOT NULL COMMENT '管理员ID',
    `username` VARCHAR(50) NOT NULL COMMENT '用户名',
    `password_hash` VARCHAR(255) NOT NULL COMMENT '密码哈希',
    `real_name` VARCHAR(50) COMMENT '真实姓名',
    `email` VARCHAR(100) COMMENT '邮箱',
    `phone` VARCHAR(20) COMMENT '手机号',
    `avatar` VARCHAR(200) COMMENT '头像URL',
    `role` VARCHAR(20) NOT NULL DEFAULT 'admin' COMMENT '角色: super_admin/admin',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 0-禁用 1-正常',
    `last_login_time` TIMESTAMP NULL COMMENT '最后登录时间',
    `last_login_ip` VARCHAR(45) COMMENT '最后登录IP',
    `create_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `delete_time` TIMESTAMP NULL COMMENT '删除时间',
    
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_admin_username` (`username`),
    UNIQUE KEY `uk_admin_email` (`email`),
    KEY `idx_admin_role` (`role`),
    KEY `idx_admin_status` (`status`),
    KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员表';

-- 登录记录表
CREATE TABLE `login_record` (
    `id` VARCHAR(32) NOT NULL COMMENT '记录ID',
    `user_id` VARCHAR(32) NOT NULL COMMENT '用户ID',
    `user_type` VARCHAR(20) NOT NULL COMMENT '用户类型: user/admin',
    `username` VARCHAR(50) COMMENT '用户名',
    `login_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
    `login_ip` VARCHAR(45) COMMENT '登录IP',
    `user_agent` VARCHAR(500) COMMENT '浏览器UA',
    `login_status` TINYINT NOT NULL COMMENT '登录状态: 0-失败 1-成功',
    `fail_reason` VARCHAR(200) COMMENT '失败原因',
    `create_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    PRIMARY KEY (`id`),
    KEY `idx_login_record_user_id` (`user_id`),
    KEY `idx_login_record_user_type` (`user_type`),
    KEY `idx_login_record_login_time` (`login_time`),
    KEY `idx_login_record_login_status` (`login_status`),
    KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='登录记录表';

-- 初始管理员账号 (密码: admin123)
-- 密码哈希使用 bcrypt 生成，cost=10
INSERT INTO `admin` (`id`, `username`, `password_hash`, `real_name`, `email`, `role`, `status`) VALUES
('admin_super_001', 'superadmin', '$2a$10$EixZaYVK1fsbw1ZfbX3OXePaWxn96p36WQoeG6Lruj3vjPGga31lW', '超级管理员', 'admin@example.com', 'super_admin', 1)
ON DUPLICATE KEY UPDATE `id` = `id`;
