-- 柳屋公司测试数据插入脚本
-- 根据各个 SQL 文件结构生成
-- 密码: 123456 (bcrypt 加密)

SET NAMES utf8mb4;

-- 1. 插入公司数据 (company.sql)
INSERT INTO `company` (
    `id`, `name`, `company_attributes`, `company_business`, `owner`,
    `description`, `address`, `phone`, `email`, `status`
) VALUES (
             'company_liuwu_001', '柳屋', 0, 0, 'user_admin_001',
             '专注于企业任务管理和团队协作的科技公司',
             '北京市朝阳区科技园区', '010-12345678', 'contact@liuwu.com', 1
         );

-- 2. 插入部门数据 (company.sql)
INSERT INTO `department` (
    `id`, `company_id`, `department_name`, `department_code`,
    `department_priority`, `description`, `status`
) VALUES
      ('dept_product_001', 'company_liuwu_001', '产品部门', 'PROD', 5, '负责产品规划、设计和需求分析', 1),
      ('dept_dev_001', 'company_liuwu_001', '开发部门', 'DEV', 4, '负责产品开发和维护', 1),
      ('dept_ops_001', 'company_liuwu_001', '运维部门', 'OPS', 3, '负责系统运维和部署', 1),
      ('dept_data_001', 'company_liuwu_001', '数据处理部门', 'DATA', 2, '负责数据分析和处理', 1);

-- 3. 插入职位数据 (company.sql)
INSERT INTO `position` (
    `id`, `department_id`, `position_name`, `position_code`, `job_type`,
    `position_level`, `required_skills`, `job_description`, `responsibilities`,
    `requirements`, `salary_range_min`, `salary_range_max`, `is_management`,
    `max_employees`, `current_employees`, `status`
) VALUES
-- 产品部门职位
('pos_pm_001', 'dept_product_001', '产品经理', 'PM', 2, 3, '产品规划,需求分析,项目管理', '负责产品全生命周期管理', '制定产品策略,需求分析,项目协调', '3年以上产品经验,良好的沟通能力', 15000.00, 25000.00, 1, 2, 1, 1),
('pos_ux_001', 'dept_product_001', 'UI/UX设计师', 'UX', 1, 2, 'UI设计,UX设计,原型设计', '负责产品界面设计和用户体验优化', '界面设计,交互设计,用户研究', '2年以上设计经验,熟练使用设计工具', 12000.00, 20000.00, 0, 3, 1, 1),

-- 开发部门职位
('pos_dev_lead_001', 'dept_dev_001', '开发主管', 'DEV_LEAD', 0, 4, 'Go,Java,Python,团队管理', '负责开发团队管理和技术架构', '技术架构设计,团队管理,代码审查', '5年以上开发经验,2年以上管理经验', 20000.00, 35000.00, 1, 1, 1, 1),
('pos_backend_001', 'dept_dev_001', '后端开发工程师', 'BACKEND', 0, 3, 'Go,MySQL,Redis,微服务', '负责后端服务开发和维护', 'API开发,数据库设计,系统优化', '3年以上后端开发经验', 15000.00, 25000.00, 0, 5, 1, 1),
('pos_frontend_001', 'dept_dev_001', '前端开发工程师', 'FRONTEND', 0, 2, 'React,Vue,JavaScript,TypeScript', '负责前端界面开发和维护', '页面开发,组件设计,性能优化', '2年以上前端开发经验', 12000.00, 20000.00, 0, 3, 1, 1),

-- 运维部门职位
('pos_devops_001', 'dept_ops_001', 'DevOps工程师', 'DEVOPS', 0, 3, 'Docker,Kubernetes,CI/CD,Linux', '负责系统部署和运维', '系统部署,监控,自动化运维', '3年以上运维经验', 15000.00, 25000.00, 0, 2, 1, 1),

-- 数据处理部门职位
('pos_data_analyst_001', 'dept_data_001', '数据分析师', 'DATA_ANALYST', 1, 2, 'Python,SQL,数据分析,机器学习', '负责数据分析和挖掘', '数据分析,报表制作,业务洞察', '2年以上数据分析经验', 13000.00, 22000.00, 0, 3, 0, 1);

-- 4. 插入用户数据 (user.sql)
INSERT INTO `user` (
    `id`, `username`, `password_hash`, `email`, `phone`,
    `real_name`, `gender`, `status`
) VALUES
      ('user_admin_001', 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'admin@liuwu.com', '13800138000', '系统管理员', 1, 1),
      ('user_zhang_001', 'zhangsan', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '1844554263@qq.com', '13800138001', '张三', 1, 1),
      ('user_li_001', 'lisi', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '3068105274@qq.com', '13800138002', '李四', 1, 1),
      ('user_wang_001', 'wangwu', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'xhliu@gizwits.com', '13800138003', '王五', 1, 1),
      ('user_zhao_001', 'zhaoliu', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '2035213408@qq.com', '13800138004', '赵六', 1, 1),
      ('user_chen_001', 'chenqi', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'lxh17306642597@gmail.com', '13800138005', '陈七', 2, 1),
      ('user_liu_001', 'liuba', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '779909861@qq.com', '13800138006', '刘八', 1, 1),
      ('user_sun_001', 'sunjiu', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'sunjiu@liuwu.com', '13800138007', '孙九', 1, 1),
      ('user_wu_001', 'wushi', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'wushi@liuwu.com', '13800138008', '吴十', 2, 1),
      ('user_zhou_001', 'zhoushiyi', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'zhoushiyi@liuwu.com', '13800138009', '周十一', 1, 1);

-- 5. 插入员工数据 (user.sql)
INSERT INTO `employee` (
    `id`, `user_id`, `company_id`, `department_id`, `position_id`,
    `employee_id`, `real_name`, `email`, `phone`, `skills`, `role_tags`,
    `hire_date`, `status`
) VALUES
      ('emp_admin_001', 'user_admin_001', 'company_liuwu_001', NULL, NULL, 'EMP001', '系统管理员', 'admin@liuwu.com', '13800138000', '系统管理,权限管理', '管理员', '2024-01-01', 1),
      ('emp_zhang_001', 'user_zhang_001', 'company_liuwu_001', 'dept_product_001', 'pos_pm_001', 'EMP002', '张三', '1844554263@qq.com', '13800138001', '产品规划,需求分析', '产品经理', '2024-02-01', 1),
      ('emp_li_001', 'user_li_001', 'company_liuwu_001', 'dept_product_001', 'pos_ux_001', 'EMP003', '李四', '3068105274@qq.com', '13800138002', 'UI设计,UX设计', 'UI设计师', '2024-02-15', 1),
      ('emp_wang_001', 'user_wang_001', 'company_liuwu_001', 'dept_dev_001', 'pos_dev_lead_001', 'EMP004', '王五', 'xhliu@gizwits.com', '13800138003', 'Go,微服务,团队管理', '开发主管', '2024-03-01', 1),
      ('emp_zhao_001', 'user_zhao_001', 'company_liuwu_001', 'dept_dev_001', 'pos_backend_001', 'EMP005', '赵六', '2035213408@qq.com', '13800138004', 'Go,MySQL,Redis', '后端工程师', '2024-03-15', 1),
      ('emp_chen_001', 'user_chen_001', 'company_liuwu_001', 'dept_dev_001', 'pos_frontend_001', 'EMP006', '陈七', 'lxh17306642597@gmail.com', '13800138005', 'React,JavaScript,TypeScript', '前端工程师', '2024-04-01', 1),
      ('emp_liu_001', 'user_liu_001', 'company_liuwu_001', 'dept_ops_001', 'pos_devops_001', 'EMP007', '刘八', '779909861@qq.com', '13800138006', 'Docker,Kubernetes,CI/CD', 'DevOps工程师', '2024-04-15', 1),
      ('emp_sun_001', 'user_sun_001', 'company_liuwu_001', 'dept_data_001', 'pos_data_analyst_001', 'EMP008', '孙九', 'sunjiu@liuwu.com', '13800138007', 'Python,SQL,数据分析', '数据分析师', '2024-05-01', 1),
      ('emp_wu_001', 'user_wu_001', 'company_liuwu_001', 'dept_dev_001', 'pos_backend_001', 'EMP009', '吴十', 'wushi@liuwu.com', '13800138008', 'Java,Spring Boot,微服务', '后端工程师', '2024-05-15', 1),
      ('emp_zhou_001', 'user_zhou_001', 'company_liuwu_001', 'dept_dev_001', 'pos_frontend_001', 'EMP010', '周十一', 'zhoushiyi@liuwu.com', '13800138009', 'Vue,TypeScript,Element UI', '前端工程师', '2024-06-01', 1);

-- 6. 设置部门经理
UPDATE `department` SET `manager_id` = 'emp_zhang_001' WHERE `id` = 'dept_product_001';
UPDATE `department` SET `manager_id` = 'emp_wang_001' WHERE `id` = 'dept_dev_001';
UPDATE `department` SET `manager_id` = 'emp_liu_001' WHERE `id` = 'dept_ops_001';

-- DDL: 角色与职位角色关联表（如不存在则创建）
CREATE TABLE IF NOT EXISTS `role` (
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

CREATE TABLE IF NOT EXISTS `position_role` (
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='职位角色关联表';

-- 6.1 角色与职位角色关联
-- 角色表
-- 清空旧数据（不保留旧版权限）
DELETE FROM `position_role` WHERE `position_id` IN ('pos_dev_lead_001','pos_backend_001','pos_frontend_001');
DELETE FROM `role` WHERE `id` IN ('role_admin_001','role_manager_001','role_dev_001','role_org_001');

-- 新版权限：使用数字码 JSON 数组
-- admin：拥有全部权限码
INSERT INTO `role` (
    `id`, `company_id`, `role_name`, `role_code`, `role_description`,
    `is_system`, `permissions`, `status`, `delete_time`
) VALUES
    (
      'role_admin_001', 'company_liuwu_001', '系统管理员', 'ADMIN', '系统级最高权限', 1,
      '[1,2,3,4,5,10,11,12,13,20,21,22,23,30,31,32,40,41,42,43,45,46,47,48,50,51,52,53,60,61,62,63,64,65,70,71,72,73,74]',
      1, NULL
    ),
    -- 项目经理：任务全权限 + 通知全权限 + 交接审批
    (
      'role_manager_001', 'company_liuwu_001', '项目经理', 'PM', '项目管理与审批', 0,
      '[1,2,3,4,5,30,31,32,22]',
      1, NULL
    ),
    -- 开发工程师：任务只读 + 节点更新 + 通知只读
    (
      'role_dev_001', 'company_liuwu_001', '开发工程师', 'DEV', '开发相关权限', 0,
      '[1,11,30]',
      1, NULL
    ),
    -- 组织管理员：公司/部门/职位读写删
    (
      'role_org_001', 'company_liuwu_001', '组织管理员', 'ORG', '组织结构管理权限', 0,
      '[40,41,42,43,45,46,47,48,50,51,52,53]',
      1, NULL
    );

-- 职位角色关联（通过职位分配角色，员工通过职位获得权限）
-- 注意：系统管理员没有职位，需要通过其他方式分配权限
-- 开发主管职位 -> 项目经理角色
-- 后端工程师职位 -> 开发工程师角色 + 组织管理员角色
-- 前端工程师职位 -> 开发工程师角色
-- 数据分析师职位 -> 开发工程师角色
INSERT INTO `position_role` (
    `id`, `position_id`, `role_id`, `grant_by`, `grant_time`, `expire_time`, `status`
) VALUES
    ('pr_pm_001', 'pos_dev_lead_001', 'role_manager_001', NULL, NOW(), NULL, 1),
    ('pr_dev_001', 'pos_backend_001', 'role_dev_001', NULL, NOW(), NULL, 1),
    ('pr_org_001', 'pos_backend_001', 'role_org_001', NULL, NOW(), NULL, 1),
    ('pr_dev_002', 'pos_frontend_001', 'role_dev_001', NULL, NOW(), NULL, 1),
    ('pr_data_001', 'pos_data_analyst_001', 'role_dev_001', NULL, NOW(), NULL, 1);

-- 验证数据查询
SELECT '=== 公司信息 ===' as info;
SELECT * FROM `company` WHERE `id` = 'company_liuwu_001';

SELECT '=== 部门信息 ===' as info;
SELECT d.*, c.name as company_name FROM `department` d
                                            JOIN `company` c ON d.company_id = c.id
WHERE d.company_id = 'company_liuwu_001';

SELECT '=== 职位信息 ===' as info;
SELECT p.*, d.department_name FROM `position` p
                                       JOIN `department` d ON p.department_id = d.id
WHERE d.company_id = 'company_liuwu_001';

SELECT '=== 用户信息 ===' as info;
SELECT u.* FROM `user` u
WHERE u.id IN ('user_admin_001', 'user_zhang_001', 'user_li_001', 'user_wang_001', 'user_zhao_001', 'user_chen_001', 'user_liu_001', 'user_sun_001', 'user_wu_001', 'user_zhou_001');

SELECT '=== 员工信息 ===' as info;
SELECT e.*, d.department_name, p.position_name FROM `employee` e
                                                        LEFT JOIN `department` d ON e.department_id = d.id
                                                        LEFT JOIN `position` p ON e.position_id = p.id
WHERE e.company_id = 'company_liuwu_001';

-- 7. 任务与任务节点（用于前后端联调）
-- 7.1 任务（严格按 model/sql/task.sql 字段）
INSERT INTO `task` (
    `task_id`, `company_id`, `task_title`, `task_detail`,
    `task_status`, `task_priority`, `task_type`,
    `responsible_employee_ids`, `node_employee_ids`, `department_ids`,
    `task_start_time`, `task_deadline`, `task_creator`, `task_assigner`, `attachment_url`, `delete_time`
) VALUES
    ('task_liuwu_001', 'company_liuwu_001', '版本迭代-后端', '后端接口与数据库表结构调整',
     1, 1, 1,
     'emp_wang_001', 'emp_zhao_001', 'dept_dev_001',
     '2025-10-29 10:00:00', '2025-11-02 18:00:00', 'emp_wang_001', NULL, NULL, NULL),
    ('task_liuwu_002', 'company_liuwu_001', '版本迭代-前端', '前端页面与甘特图/时间轴联调',
     1, 2, 1,
     'emp_wang_001', 'emp_chen_001', 'dept_dev_001',
     '2025-10-30 10:00:00', '2025-11-03 18:00:00', 'emp_wang_001', NULL, NULL, NULL),
    ('task_liuwu_003', 'company_liuwu_001', '联调与测试', '接口联调与冒烟测试',
     0, 3, 0,
     'emp_liu_001', 'emp_liu_001', 'dept_ops_001',
     '2025-11-03 09:00:00', '2025-11-05 18:00:00', 'emp_liu_001', NULL, NULL, NULL),
    ('task_liuwu_004', 'company_liuwu_001', '用户权限系统优化', '优化用户权限管理模块，支持细粒度权限控制',
     1, 2, 1,
     'emp_wang_001', 'emp_wu_001', 'dept_dev_001',
     '2025-11-01 09:00:00', '2025-11-08 18:00:00', 'emp_wang_001', NULL, NULL, NULL),
    ('task_liuwu_005', 'company_liuwu_001', '数据报表开发', '开发数据分析和报表展示功能',
     1, 1, 1,
     'emp_zhang_001', 'emp_sun_001', 'dept_data_001',
     '2025-11-02 09:00:00', '2025-11-10 18:00:00', 'emp_zhang_001', NULL, NULL, NULL),
    ('task_liuwu_006', 'company_liuwu_001', '移动端适配', '完成移动端页面适配和响应式设计',
     1, 2, 1,
     'emp_wang_001', 'emp_zhou_001', 'dept_dev_001',
     '2025-11-03 09:00:00', '2025-11-12 18:00:00', 'emp_wang_001', NULL, NULL, NULL),
    ('task_liuwu_007', 'company_liuwu_001', '性能优化', '系统性能优化和数据库查询优化',
     1, 3, 1,
     'emp_wang_001', 'emp_zhao_001,emp_wu_001', 'dept_dev_001',
     '2025-11-04 09:00:00', '2025-11-15 18:00:00', 'emp_wang_001', NULL, NULL, NULL),
    ('task_liuwu_008', 'company_liuwu_001', '安全加固', '系统安全加固和漏洞修复',
     0, 3, 0,
     'emp_liu_001', 'emp_liu_001', 'dept_ops_001',
     '2025-11-05 09:00:00', '2025-11-18 18:00:00', 'emp_liu_001', NULL, NULL, NULL),
    ('task_liuwu_009', 'company_liuwu_001', 'UI设计规范', '制定UI设计规范和组件库',
     1, 1, 0,
     'emp_zhang_001', 'emp_li_001', 'dept_product_001',
     '2025-11-06 09:00:00', '2025-11-20 18:00:00', 'emp_zhang_001', NULL, NULL, NULL),
    ('task_liuwu_010', 'company_liuwu_001', 'API文档编写', '编写完整的API接口文档',
     1, 2, 1,
     'emp_wang_001', 'emp_zhao_001', 'dept_dev_001',
     '2025-11-07 09:00:00', '2025-11-22 18:00:00', 'emp_wang_001', NULL, NULL, NULL);

-- 7.2 任务节点
INSERT INTO `task_node` (
    `task_node_id`, `task_id`, `department_id`, `node_name`, `node_detail`, `ex_node_ids`,
    `node_deadline`, `node_start_time`, `estimated_days`, `actual_days`,
    `node_status`, `node_finish_time`, `executor_id`, `leader_id`, `progress`, `node_priority`, `delete_time`
) VALUES
    ('node_liuwu_001', 'task_liuwu_001', 'dept_dev_001', '接口开发', '用户/任务/节点接口', NULL,
     '2025-11-02 18:00:00', '2025-10-29 10:00:00', 5, NULL,
     1, NULL, 'emp_zhao_001', 'emp_wang_001', 40, 3, NULL),
    ('node_liuwu_002', 'task_liuwu_001', 'dept_dev_001', '数据库设计', '数据库表结构设计和优化', NULL,
     '2025-11-01 18:00:00', '2025-10-29 10:00:00', 3, NULL,
     2, '2025-11-01 17:00:00', 'emp_zhao_001', 'emp_wang_001', 100, 3, NULL),
    ('node_liuwu_003', 'task_liuwu_002', 'dept_dev_001', '页面实现', '首页甘特图/时间轴/列表页', NULL,
     '2025-11-03 18:00:00', '2025-10-30 10:00:00', 5, NULL,
     1, NULL, 'emp_chen_001', 'emp_wang_001', 20, 2, NULL),
    ('node_liuwu_004', 'task_liuwu_002', 'dept_dev_001', '组件开发', '开发通用组件和业务组件', NULL,
     '2025-11-02 18:00:00', '2025-10-30 10:00:00', 3, NULL,
     1, NULL, 'emp_zhou_001', 'emp_wang_001', 30, 2, NULL),
    ('node_liuwu_005', 'task_liuwu_003', 'dept_ops_001', '联调测试', '接口联调、bug修复', NULL,
     '2025-11-05 18:00:00', '2025-11-03 09:00:00', 3, NULL,
     0, NULL, 'emp_liu_001', 'emp_liu_001', 0, 1, NULL),
    ('node_liuwu_006', 'task_liuwu_003', 'dept_ops_001', '环境部署', '测试环境部署和配置', NULL,
     '2025-11-04 18:00:00', '2025-11-03 09:00:00', 2, NULL,
     1, NULL, 'emp_liu_001', 'emp_liu_001', 50, 2, NULL),
    ('node_liuwu_007', 'task_liuwu_004', 'dept_dev_001', '权限模块设计', '设计权限管理模块架构', NULL,
     '2025-11-05 18:00:00', '2025-11-01 09:00:00', 4, NULL,
     1, NULL, 'emp_wu_001', 'emp_wang_001', 25, 2, NULL),
    ('node_liuwu_008', 'task_liuwu_004', 'dept_dev_001', '权限接口开发', '开发权限管理相关接口', NULL,
     '2025-11-08 18:00:00', '2025-11-05 09:00:00', 3, NULL,
     0, NULL, 'emp_wu_001', 'emp_wang_001', 0, 2, NULL),
    ('node_liuwu_009', 'task_liuwu_005', 'dept_data_001', '数据分析', '进行数据分析和挖掘', NULL,
     '2025-11-08 18:00:00', '2025-11-02 09:00:00', 6, NULL,
     1, NULL, 'emp_sun_001', 'emp_zhang_001', 15, 1, NULL),
    ('node_liuwu_010', 'task_liuwu_005', 'dept_data_001', '报表开发', '开发数据报表和可视化', NULL,
     '2025-11-10 18:00:00', '2025-11-08 09:00:00', 2, NULL,
     0, NULL, 'emp_sun_001', 'emp_zhang_001', 0, 1, NULL),
    ('node_liuwu_011', 'task_liuwu_006', 'dept_dev_001', '响应式布局', '实现响应式布局设计', NULL,
     '2025-11-10 18:00:00', '2025-11-03 09:00:00', 7, NULL,
     1, NULL, 'emp_zhou_001', 'emp_wang_001', 10, 2, NULL),
    ('node_liuwu_012', 'task_liuwu_006', 'dept_dev_001', '移动端适配', '完成移动端页面适配', NULL,
     '2025-11-12 18:00:00', '2025-11-10 09:00:00', 2, NULL,
     0, NULL, 'emp_zhou_001', 'emp_wang_001', 0, 2, NULL),
    ('node_liuwu_013', 'task_liuwu_007', 'dept_dev_001', '代码优化', '代码重构和性能优化', NULL,
     '2025-11-12 18:00:00', '2025-11-04 09:00:00', 8, NULL,
     1, NULL, 'emp_zhao_001', 'emp_wang_001', 5, 3, NULL),
    ('node_liuwu_014', 'task_liuwu_007', 'dept_dev_001', '数据库优化', '数据库查询优化和索引优化', NULL,
     '2025-11-15 18:00:00', '2025-11-12 09:00:00', 3, NULL,
     0, NULL, 'emp_wu_001', 'emp_wang_001', 0, 3, NULL),
    ('node_liuwu_015', 'task_liuwu_008', 'dept_ops_001', '安全扫描', '进行安全漏洞扫描', NULL,
     '2025-11-15 18:00:00', '2025-11-05 09:00:00', 10, NULL,
     0, NULL, 'emp_liu_001', 'emp_liu_001', 0, 3, NULL),
    ('node_liuwu_016', 'task_liuwu_008', 'dept_ops_001', '漏洞修复', '修复发现的安全漏洞', NULL,
     '2025-11-18 18:00:00', '2025-11-15 09:00:00', 3, NULL,
     0, NULL, 'emp_liu_001', 'emp_liu_001', 0, 3, NULL),
    ('node_liuwu_017', 'task_liuwu_009', 'dept_product_001', '设计规范制定', '制定UI设计规范和标准', NULL,
     '2025-11-15 18:00:00', '2025-11-06 09:00:00', 9, NULL,
     1, NULL, 'emp_li_001', 'emp_zhang_001', 20, 1, NULL),
    ('node_liuwu_018', 'task_liuwu_009', 'dept_product_001', '组件库搭建', '搭建UI组件库', NULL,
     '2025-11-20 18:00:00', '2025-11-15 09:00:00', 5, NULL,
     0, NULL, 'emp_li_001', 'emp_zhang_001', 0, 1, NULL),
    ('node_liuwu_019', 'task_liuwu_010', 'dept_dev_001', '接口文档编写', '编写API接口文档', NULL,
     '2025-11-20 18:00:00', '2025-11-07 09:00:00', 13, NULL,
     1, NULL, 'emp_zhao_001', 'emp_wang_001', 8, 2, NULL),
    ('node_liuwu_020', 'task_liuwu_010', 'dept_dev_001', '文档审核', '审核和完善API文档', NULL,
     '2025-11-22 18:00:00', '2025-11-20 09:00:00', 2, NULL,
     0, NULL, 'emp_wang_001', 'emp_wang_001', 0, 2, NULL);

-- 8. 交接（演示：把前端页面实现从陈七交接给赵六）
INSERT INTO `task_handover` (
    `handover_id`, `task_id`, `from_employee_id`, `to_employee_id`,
    `handover_type`, `handover_status`,
    `handover_reason`, `handover_note`, `approver_id`, `approve_time`
) VALUES (
    'handover_liuwu_001', 'task_liuwu_002', 'emp_chen_001', 'emp_zhao_001',
    1, 1,
    '人员调整', '请赵六协助处理前端接口对接', 'emp_wang_001', NOW()
);

-- 9. 通知（分配/提醒）
INSERT INTO `notification` (
    `id`, `employee_id`, `title`, `content`, `type`, `priority`, `is_read`
) VALUES
    ('notify_liuwu_001', 'emp_zhao_001', '任务分配：接口开发', '你被指派到任务节点【接口开发】', 1, 2, 0),
    ('notify_liuwu_002', 'emp_chen_001', '截止提醒：页面实现', '任务节点【页面实现】将于3日18:00到期', 2, 1, 0),
    ('notify_liuwu_003', 'emp_wu_001', '任务分配：权限模块设计', '你被指派到任务节点【权限模块设计】', 1, 2, 0),
    ('notify_liuwu_004', 'emp_sun_001', '任务分配：数据分析', '你被指派到任务节点【数据分析】', 1, 1, 0),
    ('notify_liuwu_005', 'emp_zhou_001', '任务分配：响应式布局', '你被指派到任务节点【响应式布局】', 1, 2, 0),
    ('notify_liuwu_006', 'emp_zhao_001', '截止提醒：代码优化', '任务节点【代码优化】将于12日18:00到期', 2, 3, 0),
    ('notify_liuwu_007', 'emp_liu_001', '任务分配：安全扫描', '你被指派到任务节点【安全扫描】', 1, 3, 0),
    ('notify_liuwu_008', 'emp_li_001', '任务分配：设计规范制定', '你被指派到任务节点【设计规范制定】', 1, 1, 0),
    ('notify_liuwu_009', 'emp_zhao_001', '任务分配：接口文档编写', '你被指派到任务节点【接口文档编写】', 1, 2, 0),
    ('notify_liuwu_010', 'emp_wang_001', '系统通知：新版本发布', '系统新版本v2.0已发布，请及时查看更新内容', 0, 2, 0);

-- 10. 校验新增数据
SELECT '=== 任务（共10个）===' as info; SELECT task_id, task_title, task_priority, task_status, task_deadline FROM `task` WHERE company_id='company_liuwu_001' ORDER BY task_id;
SELECT '=== 任务节点（共20个）===' as info; SELECT task_node_id, task_id, node_name, node_deadline, node_status, progress FROM `task_node` WHERE task_id LIKE 'task_liuwu_%' ORDER BY task_node_id;
SELECT '=== 员工（共10个）===' as info; SELECT e.id, e.employee_id, e.real_name, d.department_name, p.position_name FROM `employee` e LEFT JOIN `department` d ON e.department_id = d.id LEFT JOIN `position` p ON e.position_id = p.id WHERE e.company_id='company_liuwu_001' ORDER BY e.employee_id;
SELECT '=== 交接 ===' as info; SELECT handover_id, task_id, from_employee_id, to_employee_id, handover_status FROM `task_handover` WHERE handover_id='handover_liuwu_001';
SELECT '=== 通知（共10条）===' as info; SELECT id, employee_id, title, type, priority, is_read FROM `notification` WHERE id LIKE 'notify_liuwu_%' ORDER BY id;
