
-- 公司表
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
    
    -- 外键约束在init.sql中统一添加
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='公司表';

-- 部门表
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
    KEY `idx_create_time` (`create_time`)
    
    -- 外键约束在init.sql中统一添加
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='部门表';

-- 职位表
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
    KEY `idx_create_time` (`create_time`)
    
    -- 外键约束在init.sql中统一添加
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='职位表';

