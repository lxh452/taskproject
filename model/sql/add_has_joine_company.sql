-- 添加用户是否已加入公司字段
-- 执行此SQL脚本以更新user表结构

ALTER TABLE `user`
    ADD COLUMN `has_joined_company` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已加入公司 0-否 1-是'
        AFTER `status`;

-- 为已有员工记录的用户设置为已加入公司
UPDATE `user` u
SET u.has_joined_company = 1
WHERE EXISTS (
    SELECT 1 FROM employee e WHERE e.user_id = u.id AND e.delete_time IS NULL
);

-- 注意：验证查询已移除，迁移脚本中不应包含 SELECT 语句

-- 添加任务相关字段（分开执行，失败的会被跳过，成功的会继续）
ALTER TABLE `task` ADD COLUMN `task_progress` INT NOT NULL DEFAULT 0 COMMENT '任务整体进度 0-100';
ALTER TABLE `task` ADD COLUMN `total_nodes` INT NOT NULL DEFAULT 0 COMMENT '任务节点总数';
ALTER TABLE `task` ADD COLUMN `completed_nodes` INT NOT NULL DEFAULT 0 COMMENT '已完成节点数';
ALTER TABLE `task` ADD COLUMN `estimated_hours` DECIMAL(10,2) DEFAULT 0 COMMENT '预计工时（小时）';
ALTER TABLE `task` ADD COLUMN `actual_hours` DECIMAL(10,2) DEFAULT 0 COMMENT '实际工时（小时）';
ALTER TABLE `task` ADD COLUMN `leader_id` VARCHAR(32) COMMENT '任务负责人员工ID';