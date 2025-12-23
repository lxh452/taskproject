-- 交接审批记录表（同时支持交接审批和任务节点完成审批）
CREATE TABLE IF NOT EXISTS `handover_approval` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `approval_id` varchar(64) NOT NULL COMMENT '审批记录ID',
  `handover_id` varchar(64) DEFAULT NULL COMMENT '交接ID（交接审批时使用）',
  `task_node_id` varchar(64) DEFAULT NULL COMMENT '任务节点ID（任务节点完成审批时使用）',
  `approval_step` tinyint(4) NOT NULL DEFAULT '1' COMMENT '审批步骤 1-接收人确认 2-上级审批 3-任务节点完成审批',
  `approver_id` varchar(64) NOT NULL COMMENT '审批人ID',
  `approver_name` varchar(100) DEFAULT '' COMMENT '审批人姓名',
  `approval_type` tinyint(4) NOT NULL DEFAULT '0' COMMENT '审批类型 0-待审批 1-同意 2-拒绝',
  `comment` varchar(500) DEFAULT NULL COMMENT '审批意见',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_approval_id` (`approval_id`),
  KEY `idx_handover_id` (`handover_id`),
  KEY `idx_task_node_id` (`task_node_id`),
  KEY `idx_approver_id` (`approver_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='审批记录表（交接审批和任务节点完成审批）';

-- 交接状态说明:
-- 0 = 待接收人确认 (发起人创建后)
-- 1 = 待上级审批 (接收人同意后)
-- 2 = 已通过 (上级审批通过)
-- 3 = 已拒绝 (接收人拒绝或上级拒绝)
-- 4 = 已完成 (交接完成)

-- 审批步骤说明:
-- 1 = 接收人确认（交接审批）
-- 2 = 上级审批（交接审批）
-- 3 = 任务节点完成审批

-- 审批类型说明:
-- 0 = 待审批
-- 1 = 同意
-- 2 = 拒绝

-- ============================================
-- 如果表已存在，执行以下迁移语句添加新字段：
-- ============================================
ALTER TABLE `handover_approval` ADD COLUMN `task_node_id` VARCHAR(64) NULL COMMENT '任务节点ID（任务节点完成审批时使用）' AFTER `handover_id`;
ALTER TABLE `handover_approval` ADD COLUMN `update_time` DATETIME NULL COMMENT '更新时间' AFTER `create_time`;
ALTER TABLE `handover_approval` MODIFY COLUMN `handover_id` VARCHAR(64) NULL COMMENT '交接ID（交接审批时使用）';
ALTER TABLE `handover_approval` MODIFY COLUMN `approval_step` TINYINT(4) NOT NULL DEFAULT '1' COMMENT '审批步骤 1-接收人确认 2-上级审批 3-任务节点完成审批';
ALTER TABLE `handover_approval` MODIFY COLUMN `approval_type` TINYINT(4) NOT NULL DEFAULT '0' COMMENT '审批类型 0-待审批 1-同意 2-拒绝';
CREATE INDEX `idx_task_node_id` ON `handover_approval` (`task_node_id`);
