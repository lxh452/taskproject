
-- =====================================================
-- 任务清单功能 - 数据库迁移脚本
-- 用于解决多人协作时进度条修改权限模糊的问题
-- =====================================================

-- 任务清单表（用于替代任务节点进度条的模糊概念）
-- 每个用户可以在任务节点下创建自己的清单项，只能修改自己创建的清单
-- 进度通过已完成清单数/总清单数自动计算
CREATE TABLE IF NOT EXISTS `task_checklist` (
    `checklist_id` VARCHAR(32) NOT NULL COMMENT '清单ID',
    `task_node_id` VARCHAR(32) NOT NULL COMMENT '关联的任务节点ID',
    `creator_id` VARCHAR(32) NOT NULL COMMENT '创建该清单的员工ID',
    `content` TEXT NOT NULL COMMENT '清单内容',
    `is_completed` TINYINT NOT NULL DEFAULT 0 COMMENT '是否已完成：0-未完成，1-已完成',
    `complete_time` TIMESTAMP NULL COMMENT '完成时间',
    `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序顺序',
    `create_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `delete_time` TIMESTAMP NULL COMMENT '删除时间',
    
    PRIMARY KEY (`checklist_id`),
    KEY `idx_checklist_node` (`task_node_id`),
    KEY `idx_checklist_creator` (`creator_id`),
    KEY `idx_checklist_completed` (`is_completed`),
    KEY `idx_create_time` (`create_time`),
    KEY `idx_delete_time` (`delete_time`)
    
    -- 外键约束在init.sql中统一添加
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务清单表';

-- =====================================================
-- 扩展任务表：添加总节点数统计字段
-- =====================================================
-- 检查字段是否存在，不存在则添加
SET @dbname = DATABASE();
SET @tablename = 'task';
SET @columnname = 'total_node_count';
SET @preparedStatement = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
   WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = @tablename AND COLUMN_NAME = @columnname) > 0,
  'SELECT 1',
  'ALTER TABLE `task` ADD COLUMN `total_node_count` INT NOT NULL DEFAULT 0 COMMENT ''总任务节点数'' AFTER `attachment_url`'
));
PREPARE alterIfNotExists FROM @preparedStatement;
EXECUTE alterIfNotExists;
DEALLOCATE PREPARE alterIfNotExists;

SET @columnname = 'completed_node_count';
SET @preparedStatement = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
   WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = @tablename AND COLUMN_NAME = @columnname) > 0,
  'SELECT 1',
  'ALTER TABLE `task` ADD COLUMN `completed_node_count` INT NOT NULL DEFAULT 0 COMMENT ''已完成任务节点数'' AFTER `total_node_count`'
));
PREPARE alterIfNotExists FROM @preparedStatement;
EXECUTE alterIfNotExists;
DEALLOCATE PREPARE alterIfNotExists;

-- =====================================================
-- 扩展任务节点表：添加清单统计字段
-- =====================================================
SET @tablename = 'task_node';
SET @columnname = 'total_checklist_count';
SET @preparedStatement = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
   WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = @tablename AND COLUMN_NAME = @columnname) > 0,
  'SELECT 1',
  'ALTER TABLE `task_node` ADD COLUMN `total_checklist_count` INT NOT NULL DEFAULT 0 COMMENT ''总任务清单数'' AFTER `node_priority`'
));
PREPARE alterIfNotExists FROM @preparedStatement;
EXECUTE alterIfNotExists;
DEALLOCATE PREPARE alterIfNotExists;

SET @columnname = 'completed_checklist_count';
SET @preparedStatement = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
   WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = @tablename AND COLUMN_NAME = @columnname) > 0,
  'SELECT 1',
  'ALTER TABLE `task_node` ADD COLUMN `completed_checklist_count` INT NOT NULL DEFAULT 0 COMMENT ''已完成任务清单数'' AFTER `total_checklist_count`'
));
PREPARE alterIfNotExists FROM @preparedStatement;
EXECUTE alterIfNotExists;
DEALLOCATE PREPARE alterIfNotExists;

-- =====================================================
-- 简化版本的 ALTER 语句（如果上面的脚本执行失败，可以手动执行以下语句）
-- =====================================================
-- ALTER TABLE `task` ADD COLUMN `total_node_count` INT NOT NULL DEFAULT 0 COMMENT '总任务节点数' AFTER `attachment_url`;
-- ALTER TABLE `task` ADD COLUMN `completed_node_count` INT NOT NULL DEFAULT 0 COMMENT '已完成任务节点数' AFTER `total_node_count`;
-- ALTER TABLE `task_node` ADD COLUMN `total_checklist_count` INT NOT NULL DEFAULT 0 COMMENT '总任务清单数' AFTER `node_priority`;
-- ALTER TABLE `task_node` ADD COLUMN `completed_checklist_count` INT NOT NULL DEFAULT 0 COMMENT '已完成任务清单数' AFTER `total_checklist_count`;

