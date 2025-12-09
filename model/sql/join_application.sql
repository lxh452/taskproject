-- 加入公司申请表
CREATE TABLE `join_application` (
    `id` VARCHAR(32) NOT NULL COMMENT '申请ID',
    `user_id` VARCHAR(32) NOT NULL COMMENT '申请用户ID',
    `company_id` VARCHAR(32) NOT NULL COMMENT '目标公司ID',
    `invite_code` VARCHAR(32) COMMENT '使用的邀请码',
    `apply_reason` TEXT COMMENT '申请理由',
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0-待审批 1-已通过 2-已拒绝 3-已取消',
    `approver_id` VARCHAR(32) COMMENT '审批人员工ID',
    `approve_time` TIMESTAMP NULL COMMENT '审批时间',
    `approve_note` TEXT COMMENT '审批备注',
    `create_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '申请时间',
    `update_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    PRIMARY KEY (`id`),
    KEY `idx_join_application_user` (`user_id`),
    KEY `idx_join_application_company` (`company_id`),
    KEY `idx_join_application_status` (`status`),
    KEY `idx_join_application_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='加入公司申请表';

