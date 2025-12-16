-- 任务节点完成审批表
CREATE TABLE IF NOT EXISTS `task_node_completion_approval` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `approval_id` varchar(64) NOT NULL COMMENT '审批记录ID',
  `task_node_id` varchar(64) NOT NULL COMMENT '任务节点ID',
  `approver_id` varchar(64) NOT NULL COMMENT '审批人ID（项目负责人）',
  `approver_name` varchar(100) DEFAULT '' COMMENT '审批人姓名',
  `approval_type` tinyint(4) NOT NULL DEFAULT '0' COMMENT '审批类型 0-待审批 1-同意 2-拒绝',
  `comment` varchar(500) DEFAULT NULL COMMENT '审批意见',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_approval_id` (`approval_id`),
  KEY `idx_task_node_id` (`task_node_id`),
  KEY `idx_approver_id` (`approver_id`),
  KEY `idx_approval_type` (`approval_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务节点完成审批表';

-- 节点状态说明:
-- 0 = 待处理 (未开始)
-- 1 = 进行中 (已启动，在流程设计器中流转后自动变为进行中)
-- 2 = 已完成 (审批通过后变为已完成)
-- 3 = 已逾期
-- 4 = 待审批 (员工提交审批后，等待负责人审批)

