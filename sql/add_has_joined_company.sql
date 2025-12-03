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

-- 验证更新结果
SELECT id, username, has_joined_company FROM `user` WHERE delete_time IS NULL;


