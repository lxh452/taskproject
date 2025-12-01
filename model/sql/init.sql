-- 企业任务交接与派发系统数据库初始化脚本
-- 创建顺序：company -> department -> position -> user -> employee -> role -> position_role -> task -> task_node -> task_log -> task_handover -> notification -> system_config -> operation_log

-- 设置字符集
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- 1. 公司表
CREATE TABLE `company` (
    `id` VARCHAR(32) NOT NULL COMMENT '公司id',
    `name` VARCHAR(100) NOT NULL COMMENT '公司名字',
    `company_attributes` TINYINT NOT NULL DEFAULT 0 COMMENT '企业属性 0--民营企业 1--外资企业 2--国有企业 3--合资企业',
    `company_business` TINYINT NOT NULL DEFAULT 0 COMMENT '公司业务 0--科技类 1--文化传媒类 2--咨询类 3--管理类',
    `owner` VARCHAR(32) NOT NULL COMMENT '公司拥有者，与user表的id关联',
    `description` TEXT COMMENT '公司描述',
    `address` VARCHAR(200) COMMENT '公司地址',
    `phone` VARCHAR(20) COMMENT '联系电话',
    `email` VARCHAR(100) COMMENT '联系邮箱',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 0-禁用 1-正常',
    `create_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `delete_time` TIMESTAMP NULL COMMENT '删除时间',
    
    PRIMARY KEY (`id`),
    KEY `idx_company_name` (`name`),
    KEY `idx_company_owner` (`owner`),
    KEY `idx_company_status` (`status`),
    KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='公司表';

-- 2. 部门表
CREATE TABLE `department` (
    `id` VARCHAR(32) NOT NULL COMMENT '部门id',
    `company_id` VARCHAR(32) NOT NULL COMMENT '公司id',
    `parent_id` VARCHAR(32) NULL COMMENT '父部门id',
    `department_name` VARCHAR(50) NOT NULL COMMENT '部门名',
    `department_code` VARCHAR(20) COMMENT '部门编码',
    `department_priority` TINYINT NOT NULL DEFAULT 0 COMMENT '部门优先级 0--6 根据拥有者进行权限的分配，后续分配任务或者修改权限根据这个优先级来 6权限最大',
    `manager_id` VARCHAR(32) COMMENT '部门经理id',
    `description` TEXT COMMENT '部门描述',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 0-禁用 1-正常',
    `create_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `delete_time` TIMESTAMP NULL COMMENT '删除时间',
    
    PRIMARY KEY (`id`),
    KEY `idx_department_company` (`company_id`),
    KEY `idx_department_parent` (`parent_id`),
    KEY `idx_department_manager` (`manager_id`),
    KEY `idx_department_priority` (`department_priority`),
    KEY `idx_department_status` (`status`),
    KEY `idx_create_time` (`create_time`),
    
    CONSTRAINT `fk_department_company` FOREIGN KEY (`company_id`) REFERENCES `company`(`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_department_parent` FOREIGN KEY (`parent_id`) REFERENCES `department`(`id`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='部门表';

-- 3. 职位表
CREATE TABLE `position` (
    `id` VARCHAR(32) NOT NULL COMMENT '职位id',
    `department_id` VARCHAR(32) NOT NULL COMMENT '部门id',
    `position_name` VARCHAR(50) NOT NULL COMMENT '职位名称',
    `position_code` VARCHAR(20) COMMENT '职位编码',
    `job_type` TINYINT NOT NULL DEFAULT 0 COMMENT '岗位类型 0--专业技术类 1--专业支持类 2--管理类 3--营销类 4--操作类',
    `position_level` TINYINT NOT NULL DEFAULT 1 COMMENT '职位级别 1-初级 2-中级 3-高级 4-专家 5-资深专家',
    `required_skills` TEXT COMMENT '所需技能标签',
    `job_description` TEXT COMMENT '职位描述',
    `responsibilities` TEXT COMMENT '工作职责',
    `requirements` TEXT COMMENT '任职要求',
    `salary_range_min` DECIMAL(10,2) COMMENT '薪资范围最小值',
    `salary_range_max` DECIMAL(10,2) COMMENT '薪资范围最大值',
    `is_management` TINYINT NOT NULL DEFAULT 0 COMMENT '是否管理岗位 0-否 1-是',
    `max_employees` INT DEFAULT 0 COMMENT '最大员工数量 0-不限制',
    `current_employees` INT DEFAULT 0 COMMENT '当前员工数量',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 0-禁用 1-正常',
    `create_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `delete_time` TIMESTAMP NULL COMMENT '删除时间',
    
    PRIMARY KEY (`id`),
    KEY `idx_position_department` (`department_id`),
    KEY `idx_position_name` (`position_name`),
    KEY `idx_position_code` (`position_code`),
    KEY `idx_position_type` (`job_type`),
    KEY `idx_position_level` (`position_level`),
    KEY `idx_position_management` (`is_management`),
    KEY `idx_position_status` (`status`),
    KEY `idx_create_time` (`create_time`),
    
    CONSTRAINT `fk_position_department` FOREIGN KEY (`department_id`) REFERENCES `department`(`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='职位表';

-- 4. 用户表
CREATE TABLE `user` (
    `id` VARCHAR(32) NOT NULL COMMENT '用户id',
    `company_id` VARCHAR(32) NOT NULL COMMENT '公司id',
    `department_id` VARCHAR(32) COMMENT '部门id',
    `position_id` VARCHAR(32) COMMENT '职位id',
    `username` VARCHAR(50) NOT NULL COMMENT '用户名',
    `real_name` VARCHAR(50) NOT NULL COMMENT '真实姓名',
    `email` VARCHAR(100) COMMENT '邮箱',
    `phone` VARCHAR(20) COMMENT '手机号',
    `avatar` VARCHAR(200) COMMENT '头像URL',
    `gender` TINYINT COMMENT '性别 0-未知 1-男 2-女',
    `birthday` DATE COMMENT '生日',
    `employee_id` VARCHAR(20) COMMENT '工号',
    `hire_date` DATE COMMENT '入职日期',
    `leave_date` DATE COMMENT '离职日期',
    `skills` TEXT COMMENT '技能标签',
    `role_tags` TEXT COMMENT '角色标签',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 0-离职 1-在职 2-请假',
    `create_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `delete_time` TIMESTAMP NULL COMMENT '删除时间',
    
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_username` (`username`),
    UNIQUE KEY `uk_user_email` (`email`),
    UNIQUE KEY `uk_user_phone` (`phone`),
    UNIQUE KEY `uk_user_employee_id` (`employee_id`),
    KEY `idx_user_company` (`company_id`),
    KEY `idx_user_department` (`department_id`),
    KEY `idx_user_position` (`position_id`),
    KEY `idx_user_status` (`status`),
    KEY `idx_user_hire_date` (`hire_date`),
    KEY `idx_create_time` (`create_time`),
    
    CONSTRAINT `fk_user_company` FOREIGN KEY (`company_id`) REFERENCES `company`(`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_user_department` FOREIGN KEY (`department_id`) REFERENCES `department`(`id`) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT `fk_user_position` FOREIGN KEY (`position_id`) REFERENCES `position`(`id`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 5. 用户认证表
CREATE TABLE `user_auth` (
    `id` VARCHAR(32) NOT NULL COMMENT '认证id',
    `user_id` VARCHAR(32) NOT NULL COMMENT '用户id',
    `auth_type` TINYINT NOT NULL DEFAULT 0 COMMENT '认证类型 0-密码 1-手机 2-邮箱 3-第三方',
    `auth_key` VARCHAR(100) NOT NULL COMMENT '认证标识（用户名/手机/邮箱等）',
    `auth_value` VARCHAR(255) NOT NULL COMMENT '认证值（密码hash/验证码等）',
    `is_primary` TINYINT NOT NULL DEFAULT 0 COMMENT '是否主认证方式 0-否 1-是',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 0-禁用 1-正常',
    `last_login_time` TIMESTAMP NULL COMMENT '最后登录时间',
    `last_login_ip` VARCHAR(45) COMMENT '最后登录IP',
    `login_failed_count` INT NOT NULL DEFAULT 0 COMMENT '登录失败次数',
    `locked_until` TIMESTAMP NULL COMMENT '锁定到期时间',
    `create_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_auth_key` (`auth_key`),
    KEY `idx_user_auth_user` (`user_id`),
    KEY `idx_user_auth_type` (`auth_type`),
    KEY `idx_user_auth_status` (`status`),
    KEY `idx_last_login_time` (`last_login_time`),
    KEY `idx_create_time` (`create_time`),
    
    CONSTRAINT `fk_user_auth_user` FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户认证表';

-- 6. 任务表
CREATE TABLE `task` (
    `task_id` VARCHAR(32) NOT NULL COMMENT '任务id',
    `company_id` VARCHAR(32) NOT NULL COMMENT '公司id',
    `task_title` VARCHAR(200) NOT NULL COMMENT '任务标题',
    `task_detail` TEXT NOT NULL COMMENT '任务详情',
    `task_status` TINYINT NOT NULL DEFAULT 0 COMMENT '任务状态：0-未开始，1-进行中，2-已完成，3-逾期完成',
    `task_priority` TINYINT NOT NULL DEFAULT 0 COMMENT '任务优先级：0-不重要不紧急，1-紧急不重要，2-重要但不紧急，3-重要且紧急',
    `task_type` TINYINT NOT NULL DEFAULT 0 COMMENT '任务类型：0-单部门任务，1-跨部门任务',
    `responsible_employee_ids` TEXT COMMENT '负责人员工ID列表',
    `node_employee_ids` TEXT COMMENT '节点员工ID列表',
    `department_ids` TEXT COMMENT '涉及部门ID列表',
    `task_start_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '任务开始时间',
    `task_deadline` TIMESTAMP NOT NULL COMMENT '任务截止时间',
    `task_creator` VARCHAR(32) NOT NULL COMMENT '任务创建者ID',
    `task_assigner` VARCHAR(32) COMMENT '任务分配者ID',
    `attachment_url` VARCHAR(500) COMMENT '附件URL',
    `create_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `delete_time` TIMESTAMP NULL COMMENT '删除时间',
    
    PRIMARY KEY (`task_id`),
    KEY `idx_task_company` (`company_id`),
    KEY `idx_task_status` (`task_status`),
    KEY `idx_task_priority` (`task_priority`),
    KEY `idx_task_type` (`task_type`),
    KEY `idx_task_start_time` (`task_start_time`),
    KEY `idx_task_deadline` (`task_deadline`),
    KEY `idx_task_creator` (`task_creator`),
    KEY `idx_task_assigner` (`task_assigner`),
    KEY `idx_create_time` (`create_time`),
    KEY `idx_delete_time` (`delete_time`),
    
    CONSTRAINT `fk_task_company` FOREIGN KEY (`company_id`) REFERENCES `company`(`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_task_creator` FOREIGN KEY (`task_creator`) REFERENCES `user`(`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT `fk_task_assigner` FOREIGN KEY (`task_assigner`) REFERENCES `user`(`id`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务表';

-- 7. 任务节点表
CREATE TABLE `task_node` (
    `task_node_id` VARCHAR(32) NOT NULL COMMENT '任务节点id',
    `task_id` VARCHAR(32) NOT NULL COMMENT '任务id',
    `department_id` VARCHAR(32) NOT NULL COMMENT '部门id',
    `node_name` VARCHAR(200) NOT NULL COMMENT '节点名称',
    `node_detail` TEXT COMMENT '节点详情',
    `node_deadline` TIMESTAMP NOT NULL COMMENT '节点截止时间',
    `node_start_time` TIMESTAMP NOT NULL COMMENT '节点开始时间',
    `estimated_days` INT NOT NULL COMMENT '预计完成天数',
    `actual_days` INT COMMENT '实际完成天数',
    `node_status` TINYINT NOT NULL DEFAULT 0 COMMENT '节点状态 0--未开始 1--进行中 2--已完成 3--已逾期',
    `node_finish_time` TIMESTAMP NULL COMMENT '节点完成时间',
    `executor_id` VARCHAR(32) NOT NULL COMMENT '节点执行人ID',
    `leader_id` VARCHAR(32) NOT NULL COMMENT '节点负责人ID',
    `progress` TINYINT NOT NULL DEFAULT 0 COMMENT '完成进度 0-100',
    `node_priority` TINYINT NOT NULL DEFAULT 0 COMMENT '节点优先级 0-低 1-中 2-高 3-紧急',
    `create_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `delete_time` TIMESTAMP NULL COMMENT '删除时间',
    
    PRIMARY KEY (`task_node_id`),
    KEY `idx_task_node_task_id` (`task_id`),
    KEY `idx_task_node_department` (`department_id`),
    KEY `idx_task_node_deadline` (`node_deadline`),
    KEY `idx_task_node_start_time` (`node_start_time`),
    KEY `idx_task_node_status` (`node_status`),
    KEY `idx_task_node_executor` (`executor_id`),
    KEY `idx_task_node_leader` (`leader_id`),
    KEY `idx_task_node_priority` (`node_priority`),
    KEY `idx_create_time` (`create_time`),
    KEY `idx_delete_time` (`delete_time`),
    
    CONSTRAINT `fk_task_node_task` FOREIGN KEY (`task_id`) REFERENCES `task`(`task_id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_task_node_department` FOREIGN KEY (`department_id`) REFERENCES `department`(`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_task_node_executor` FOREIGN KEY (`executor_id`) REFERENCES `user`(`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT `fk_task_node_leader` FOREIGN KEY (`leader_id`) REFERENCES `user`(`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务节点表';

-- 8. 任务日志表
CREATE TABLE `task_log` (
    `log_id` VARCHAR(32) NOT NULL COMMENT '日志id',
    `task_id` VARCHAR(32) NOT NULL COMMENT '任务id',
    `task_node_id` VARCHAR(32) COMMENT '任务节点id',
    `user_id` VARCHAR(32) NOT NULL COMMENT '操作用户id',
    `log_type` TINYINT NOT NULL COMMENT '日志类型 0-创建 1-更新 2-完成 3-交接 4-评论',
    `log_content` TEXT NOT NULL COMMENT '日志内容',
    `progress` TINYINT COMMENT '进度百分比',
    `attachment_url` VARCHAR(500) COMMENT '附件URL',
    `create_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    PRIMARY KEY (`log_id`),
    KEY `idx_task_log_task` (`task_id`),
    KEY `idx_task_log_node` (`task_node_id`),
    KEY `idx_task_log_user` (`user_id`),
    KEY `idx_task_log_type` (`log_type`),
    KEY `idx_create_time` (`create_time`),
    
    CONSTRAINT `fk_task_log_task` FOREIGN KEY (`task_id`) REFERENCES `task`(`task_id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_task_log_node` FOREIGN KEY (`task_node_id`) REFERENCES `task_node`(`task_node_id`) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT `fk_task_log_user` FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务日志表';

-- 9. 任务交接表
CREATE TABLE `task_handover` (
    `handover_id` VARCHAR(32) NOT NULL COMMENT '交接id',
    `task_id` VARCHAR(32) NOT NULL COMMENT '任务id',
    `from_user_id` VARCHAR(32) NOT NULL COMMENT '原负责人id',
    `to_user_id` VARCHAR(32) NOT NULL COMMENT '新负责人id',
    `handover_type` TINYINT NOT NULL DEFAULT 0 COMMENT '交接类型 0-提议 1-直接交接 2-系统自动',
    `handover_status` TINYINT NOT NULL DEFAULT 0 COMMENT '交接状态 0-待确认 1-已接受 2-已拒绝 3-已完成',
    `handover_reason` TEXT COMMENT '交接原因',
    `handover_note` TEXT COMMENT '交接备注',
    `approver_id` VARCHAR(32) COMMENT '审批人id',
    `approve_time` TIMESTAMP NULL COMMENT '审批时间',
    `create_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    PRIMARY KEY (`handover_id`),
    KEY `idx_handover_task` (`task_id`),
    KEY `idx_handover_from_user` (`from_user_id`),
    KEY `idx_handover_to_user` (`to_user_id`),
    KEY `idx_handover_type` (`handover_type`),
    KEY `idx_handover_status` (`handover_status`),
    KEY `idx_handover_approver` (`approver_id`),
    KEY `idx_create_time` (`create_time`),
    
    CONSTRAINT `fk_handover_task` FOREIGN KEY (`task_id`) REFERENCES `task`(`task_id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_handover_from_user` FOREIGN KEY (`from_user_id`) REFERENCES `user`(`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT `fk_handover_to_user` FOREIGN KEY (`to_user_id`) REFERENCES `user`(`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT `fk_handover_approver` FOREIGN KEY (`approver_id`) REFERENCES `user`(`id`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务交接表';

-- 10. 通知表
CREATE TABLE `notification` (
    `id` VARCHAR(32) NOT NULL COMMENT '通知id',
    `user_id` VARCHAR(32) NOT NULL COMMENT '接收用户id',
    `title` VARCHAR(200) NOT NULL COMMENT '通知标题',
    `content` TEXT NOT NULL COMMENT '通知内容',
    `type` TINYINT NOT NULL DEFAULT 0 COMMENT '通知类型 0-系统 1-任务 2-交接 3-提醒',
    `category` VARCHAR(50) COMMENT '通知分类',
    `is_read` TINYINT NOT NULL DEFAULT 0 COMMENT '是否已读 0-未读 1-已读',
    `read_time` TIMESTAMP NULL COMMENT '阅读时间',
    `priority` TINYINT NOT NULL DEFAULT 0 COMMENT '优先级 0-低 1-中 2-高 3-紧急',
    `related_id` VARCHAR(32) COMMENT '关联对象id（任务id等）',
    `related_type` VARCHAR(50) COMMENT '关联对象类型',
    `create_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    PRIMARY KEY (`id`),
    KEY `idx_notification_user` (`user_id`),
    KEY `idx_notification_type` (`type`),
    KEY `idx_notification_category` (`category`),
    KEY `idx_notification_read` (`is_read`),
    KEY `idx_notification_priority` (`priority`),
    KEY `idx_notification_related` (`related_id`, `related_type`),
    KEY `idx_create_time` (`create_time`),
    
    CONSTRAINT `fk_notification_user` FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知表';

-- 11. 用户权限表
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
    KEY `idx_create_time` (`create_time`),
    
    CONSTRAINT `fk_permission_user` FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_permission_grant_by` FOREIGN KEY (`grant_by`) REFERENCES `user`(`id`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户权限表';

-- 添加外键约束
-- 公司表外键
ALTER TABLE `company` ADD CONSTRAINT `fk_company_owner` FOREIGN KEY (`owner`) REFERENCES `user`(`id`) ON DELETE RESTRICT ON UPDATE CASCADE;

-- 部门表外键
ALTER TABLE `department` ADD CONSTRAINT `fk_department_company` FOREIGN KEY (`company_id`) REFERENCES `company`(`id`) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE `department` ADD CONSTRAINT `fk_department_parent` FOREIGN KEY (`parent_id`) REFERENCES `department`(`id`) ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE `department` ADD CONSTRAINT `fk_department_manager` FOREIGN KEY (`manager_id`) REFERENCES `employee`(`id`) ON DELETE SET NULL ON UPDATE CASCADE;

-- 职位表外键
ALTER TABLE `position` ADD CONSTRAINT `fk_position_department` FOREIGN KEY (`department_id`) REFERENCES `department`(`id`) ON DELETE CASCADE ON UPDATE CASCADE;

-- 员工表外键
ALTER TABLE `employee` ADD CONSTRAINT `fk_employee_user` FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE `employee` ADD CONSTRAINT `fk_employee_company` FOREIGN KEY (`company_id`) REFERENCES `company`(`id`) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE `employee` ADD CONSTRAINT `fk_employee_department` FOREIGN KEY (`department_id`) REFERENCES `department`(`id`) ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE `employee` ADD CONSTRAINT `fk_employee_position` FOREIGN KEY (`position_id`) REFERENCES `position`(`id`) ON DELETE SET NULL ON UPDATE CASCADE;

-- 角色表外键
ALTER TABLE `role` ADD CONSTRAINT `fk_role_company` FOREIGN KEY (`company_id`) REFERENCES `company`(`id`) ON DELETE CASCADE ON UPDATE CASCADE;

-- 职位角色关联表外键
ALTER TABLE `position_role` ADD CONSTRAINT `fk_position_role_position` FOREIGN KEY (`position_id`) REFERENCES `position`(`id`) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE `position_role` ADD CONSTRAINT `fk_position_role_role` FOREIGN KEY (`role_id`) REFERENCES `role`(`id`) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE `position_role` ADD CONSTRAINT `fk_position_role_grant_by` FOREIGN KEY (`grant_by`) REFERENCES `employee`(`id`) ON DELETE SET NULL ON UPDATE CASCADE;

-- 任务表外键
ALTER TABLE `task` ADD CONSTRAINT `fk_task_company` FOREIGN KEY (`company_id`) REFERENCES `company`(`id`) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE `task` ADD CONSTRAINT `fk_task_creator` FOREIGN KEY (`task_creator`) REFERENCES `employee`(`id`) ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE `task` ADD CONSTRAINT `fk_task_assigner` FOREIGN KEY (`task_assigner`) REFERENCES `employee`(`id`) ON DELETE SET NULL ON UPDATE CASCADE;

-- 任务节点表外键
ALTER TABLE `task_node` ADD CONSTRAINT `fk_task_node_task` FOREIGN KEY (`task_id`) REFERENCES `task`(`task_id`) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE `task_node` ADD CONSTRAINT `fk_task_node_department` FOREIGN KEY (`department_id`) REFERENCES `department`(`id`) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE `task_node` ADD CONSTRAINT `fk_task_node_executor` FOREIGN KEY (`executor_id`) REFERENCES `employee`(`id`) ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE `task_node` ADD CONSTRAINT `fk_task_node_leader` FOREIGN KEY (`leader_id`) REFERENCES `employee`(`id`) ON DELETE RESTRICT ON UPDATE CASCADE;

-- 任务日志表外键
ALTER TABLE `task_log` ADD CONSTRAINT `fk_task_log_task` FOREIGN KEY (`task_id`) REFERENCES `task`(`task_id`) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE `task_log` ADD CONSTRAINT `fk_task_log_node` FOREIGN KEY (`task_node_id`) REFERENCES `task_node`(`task_node_id`) ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE `task_log` ADD CONSTRAINT `fk_task_log_employee` FOREIGN KEY (`employee_id`) REFERENCES `employee`(`id`) ON DELETE RESTRICT ON UPDATE CASCADE;

-- 任务交接表外键
ALTER TABLE `task_handover` ADD CONSTRAINT `fk_handover_task` FOREIGN KEY (`task_id`) REFERENCES `task`(`task_id`) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE `task_handover` ADD CONSTRAINT `fk_handover_from_employee` FOREIGN KEY (`from_employee_id`) REFERENCES `employee`(`id`) ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE `task_handover` ADD CONSTRAINT `fk_handover_to_employee` FOREIGN KEY (`to_employee_id`) REFERENCES `employee`(`id`) ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE `task_handover` ADD CONSTRAINT `fk_handover_approver` FOREIGN KEY (`approver_id`) REFERENCES `employee`(`id`) ON DELETE SET NULL ON UPDATE CASCADE;

-- 用户权限表外键
ALTER TABLE `user_permission` ADD CONSTRAINT `fk_permission_user` FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE `user_permission` ADD CONSTRAINT `fk_permission_grant_by` FOREIGN KEY (`grant_by`) REFERENCES `user`(`id`) ON DELETE SET NULL ON UPDATE CASCADE;

-- 通知表外键
ALTER TABLE `notification` ADD CONSTRAINT `fk_notification_employee` FOREIGN KEY (`employee_id`) REFERENCES `employee`(`id`) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE `notification` ADD CONSTRAINT `fk_notification_sender` FOREIGN KEY (`sender_id`) REFERENCES `employee`(`id`) ON DELETE SET NULL ON UPDATE CASCADE;

-- 操作日志表外键
ALTER TABLE `operation_log` ADD CONSTRAINT `fk_log_user` FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE `operation_log` ADD CONSTRAINT `fk_log_employee` FOREIGN KEY (`employee_id`) REFERENCES `employee`(`id`) ON DELETE SET NULL ON UPDATE CASCADE;

-- 恢复外键检查
SET FOREIGN_KEY_CHECKS = 1;
