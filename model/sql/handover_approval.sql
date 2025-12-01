-- 交接审批记录表
CREATE TABLE IF NOT EXISTS `handover_approval` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `approval_id` varchar(64) NOT NULL COMMENT '审批记录ID',
  `handover_id` varchar(64) NOT NULL COMMENT '交接ID',
  `approval_step` tinyint(4) NOT NULL DEFAULT '1' COMMENT '审批步骤 1-接收人确认 2-上级审批',
  `approver_id` varchar(64) NOT NULL COMMENT '审批人ID',
  `approver_name` varchar(100) DEFAULT '' COMMENT '审批人姓名',
  `approval_type` tinyint(4) NOT NULL DEFAULT '1' COMMENT '审批类型 1-同意 2-拒绝',
  `comment` varchar(500) DEFAULT NULL COMMENT '审批意见',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_approval_id` (`approval_id`),
  KEY `idx_handover_id` (`handover_id`),
  KEY `idx_approver_id` (`approver_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='交接审批记录表';

-- 交接状态说明:
-- 0 = 待接收人确认 (发起人创建后)
-- 1 = 待上级审批 (接收人同意后)
-- 2 = 已通过 (上级审批通过)
-- 3 = 已拒绝 (接收人拒绝或上级拒绝)
-- 4 = 已完成 (交接完成)


