
-- 用户权限表
CREATE TABLE `user_permission` (
    `id` VARCHAR(32) NOT NULL COMMENT '权限id',
    `user_id` VARCHAR(32) NOT NULL COMMENT '用户id',
    `permission_code` VARCHAR(50) NOT NULL COMMENT '权限编码',
    `permission_name` VARCHAR(100) NOT NULL COMMENT '权限名称',
    `resource_type` TINYINT NOT NULL COMMENT '资源类型 0-菜单 1-按钮 2-接口 3-数据',
    `resource_id` VARCHAR(32) COMMENT '资源id',
    `grant_type` TINYINT NOT NULL DEFAULT 0 COMMENT '授权类型 0-直接授权 1-角色授权 2-部门授权',
    `grant_by` VARCHAR(32) COMMENT '授权人id',
    `expire_time` TIMESTAMP NULL COMMENT '过期时间',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 0-禁用 1-正常',
    `create_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    PRIMARY KEY (`id`),
    KEY `idx_permission_user` (`user_id`),
    KEY `idx_permission_code` (`permission_code`),
    KEY `idx_permission_type` (`resource_type`),
    KEY `idx_permission_grant_type` (`grant_type`),
    KEY `idx_permission_status` (`status`),
    KEY `idx_create_time` (`create_time`)

    -- 外键约束在init.sql中统一添加
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户权限表';

-- 通知表
CREATE TABLE `notification` (
    `id` VARCHAR(32) NOT NULL COMMENT '通知id',
    `employee_id` VARCHAR(32) NOT NULL COMMENT '接收员工id',
    `title` VARCHAR(200) NOT NULL COMMENT '通知标题',
    `content` TEXT NOT NULL COMMENT '通知内容',
    `type` TINYINT NOT NULL DEFAULT 0 COMMENT '通知类型 0-系统 1-任务 2-交接 3-提醒',
    `category` VARCHAR(50) COMMENT '通知分类',
    `is_read` TINYINT NOT NULL DEFAULT 0 COMMENT '是否已读 0-未读 1-已读',
    `read_time` TIMESTAMP NULL COMMENT '阅读时间',
    `priority` TINYINT NOT NULL DEFAULT 0 COMMENT '优先级 0-低 1-中 2-高 3-紧急',
    `related_id` VARCHAR(32) COMMENT '关联对象id（任务id等）',
    `related_type` VARCHAR(50) COMMENT '关联对象类型',
    `sender_id` VARCHAR(32) COMMENT '发送者员工id',
    `create_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    PRIMARY KEY (`id`),
    KEY `idx_notification_employee` (`employee_id`),
    KEY `idx_notification_type` (`type`),
    KEY `idx_notification_category` (`category`),
    KEY `idx_notification_read` (`is_read`),
    KEY `idx_notification_priority` (`priority`),
    index `idx_notification_related` (`related_id`, `related_type`),
    KEY `idx_notification_sender` (`sender_id`),
    KEY `idx_create_time` (`create_time`)

    -- 外键约束在init.sql中统一添加
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知表';
