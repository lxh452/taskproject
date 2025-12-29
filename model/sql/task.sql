-- 跨部门任务表/总任务表
CREATE TABLE `task` (
    `task_id` VARCHAR(32) NOT NULL COMMENT '任务id',
    `company_id` VARCHAR(32) NOT NULL COMMENT '公司id',
    `task_title` VARCHAR(200) NOT NULL COMMENT '任务标题',
    `task_detail` TEXT NOT NULL COMMENT '任务详情',
    `task_status` TINYINT NOT NULL DEFAULT 0 COMMENT '任务状态：0-未开始，1-进行中，2-已完成，3-逾期完成',
    `task_progress` INT NOT NULL DEFAULT 0 COMMENT '任务整体进度 0-100',
    `total_nodes` INT NOT NULL DEFAULT 0 COMMENT '任务节点总数',
    `completed_nodes` INT NOT NULL DEFAULT 0 COMMENT '已完成节点数',
    `estimated_hours` DECIMAL(10,2) DEFAULT 0 COMMENT '预计工时（小时）',
    `actual_hours` DECIMAL(10,2) DEFAULT 0 COMMENT '实际工时（小时）',
    `task_priority` TINYINT NOT NULL DEFAULT 0 COMMENT '任务优先级：0-不重要不紧急，1-紧急不重要，2-重要但不紧急，3-重要且紧急',
    `task_type` TINYINT NOT NULL DEFAULT 0 COMMENT '任务类型：0-单部门任务，1-跨部门任务',
    `responsible_employee_ids` TEXT COMMENT '负责人员工ID列表',
    `node_employee_ids` TEXT COMMENT '节点员工ID列表',
    `department_ids` TEXT COMMENT '涉及部门ID列表',
    `task_start_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '任务开始时间',
    `task_deadline` TIMESTAMP NOT NULL COMMENT '任务截止时间',
    `task_creator` VARCHAR(32) NOT NULL COMMENT '任务创建者员工ID',
    `leader_id` VARCHAR(32) COMMENT '任务负责人员工ID',
    `task_assigner` VARCHAR(32) COMMENT '任务分配者员工ID',
    `attachment_url` VARCHAR(500) COMMENT '附件URL',
    `create_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `delete_time` TIMESTAMP NULL COMMENT '删除时间',
    `total_node_count` INT NOT NULL DEFAULT 0 COMMENT '总任务节点数' ,
    `completed_node_count` INT NOT NULL DEFAULT 0 COMMENT '已完成任务节点数',

    
    PRIMARY KEY (`task_id`),
    KEY `idx_task_company` (`company_id`),
    KEY `idx_task_status` (`task_status`),
    KEY `idx_task_progress` (`task_progress`),
    KEY `idx_task_priority` (`task_priority`),
    KEY `idx_task_type` (`task_type`),
    KEY `idx_task_start_time` (`task_start_time`),
    KEY `idx_task_deadline` (`task_deadline`),
    KEY `idx_task_creator` (`task_creator`),
    KEY `idx_leader_id` (`leader_id`),
    KEY `idx_task_assigner` (`task_assigner`),
    KEY `idx_create_time` (`create_time`),
    KEY `idx_delete_time` (`delete_time`)
    
    -- 外键约束在init.sql中统一添加
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='跨部门任务表/总任务表';

-- 子任务表/单部门节点表
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
    `ex_node_ids` varchar(200) comment '需要优先完成的任务节点',
    `node_status` TINYINT NOT NULL DEFAULT 0 COMMENT '节点状态 0--未开始 1--进行中 2--已完成 3--已逾期',
    `node_finish_time` TIMESTAMP NULL COMMENT '节点完成时间',
    `executor_id` VARCHAR(200) NOT NULL COMMENT '节点执行人员工ID',
    `leader_id` VARCHAR(32) NOT NULL COMMENT '节点负责人员工ID',
    `progress` TINYINT NOT NULL DEFAULT 0 COMMENT '完成进度 0-100',
    `node_priority` TINYINT NOT NULL DEFAULT 0 COMMENT '节点优先级 0-低 1-中 2-高 3-紧急',
    `total_checklist_count` INT NOT NULL DEFAULT 0 COMMENT '总任务清单数',
    `completed_checklist_count` INT NOT NULL DEFAULT 0 COMMENT '已完成任务清单数',
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
    KEY `idx_delete_time` (`delete_time`)
    
    -- 外键约束在init.sql中统一添加
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='子任务表/单部门节点表';

-- 任务日志表
CREATE TABLE `task_log` (
    `log_id` VARCHAR(32) NOT NULL COMMENT '日志id',
    `task_id` VARCHAR(32) NOT NULL COMMENT '任务id',
    `task_node_id` VARCHAR(32) COMMENT '任务节点id',
    `employee_id` VARCHAR(32) NOT NULL COMMENT '操作员工id',
    `log_type` TINYINT NOT NULL COMMENT '日志类型 0-创建 1-更新 2-完成 3-交接 4-评论',
    `log_content` TEXT NOT NULL COMMENT '日志内容',
    `progress` TINYINT COMMENT '进度百分比',
    `attachment_url` VARCHAR(500) COMMENT '附件URL',
    `create_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    PRIMARY KEY (`log_id`),
    KEY `idx_task_log_task` (`task_id`),
    KEY `idx_task_log_node` (`task_node_id`),
    KEY `idx_task_log_employee` (`employee_id`),
    KEY `idx_task_log_type` (`log_type`),
    KEY `idx_create_time` (`create_time`)
    
    -- 外键约束在init.sql中统一添加
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务日志表';

-- 任务交接表
CREATE TABLE `task_handover` (
    `handover_id` VARCHAR(32) NOT NULL COMMENT '交接id',
    `task_id` VARCHAR(32) NOT NULL COMMENT '任务id',
    `from_employee_id` VARCHAR(32) NOT NULL COMMENT '原负责人员工id',
    `to_employee_id` VARCHAR(32) NOT NULL COMMENT '新负责人员工id',
    `handover_type` TINYINT NOT NULL DEFAULT 0 COMMENT '交接类型 0-提议 1-直接交接 2-系统自动',
    `handover_status` TINYINT NOT NULL DEFAULT 0 COMMENT '交接状态 0-待确认 1-已接受 2-已拒绝 3-已完成',
    `handover_reason` TEXT COMMENT '交接原因',
    `handover_note` TEXT COMMENT '交接备注',
    `approver_id` VARCHAR(32) COMMENT '审批人员工id',
    `approve_time` TIMESTAMP NULL COMMENT '审批时间',
    `create_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    PRIMARY KEY (`handover_id`),
    KEY `idx_handover_task` (`task_id`),
    KEY `idx_handover_from_employee` (`from_employee_id`),
    KEY `idx_handover_to_employee` (`to_employee_id`),
    KEY `idx_handover_type` (`handover_type`),
    KEY `idx_handover_status` (`handover_status`),
    KEY `idx_handover_approver` (`approver_id`),
    KEY `idx_create_time` (`create_time`)
    
    -- 外键约束在init.sql中统一添加
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务交接表';

