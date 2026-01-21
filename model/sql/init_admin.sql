-- 初始化/重置管理员账号脚本
-- 用法: 在MySQL中执行此脚本
-- 数据库: task_project

-- 删除已存在的管理员账号（可选）
DELETE FROM `admin` WHERE `username` = 'superadmin';

-- 插入管理员账号
-- 用户名: superadmin
-- 密码: admin123
-- 密码哈希: $2a$10$EixZaYVK1fsbw1ZfbX3OXePaWxn96p36WQoeG6Lruj3vjPGga31lW (使用bcrypt生成, cost=10)
INSERT INTO `admin` (`id`, `username`, `password_hash`, `real_name`, `email`, `role`, `status`) VALUES
('admin_super_001', 'superadmin', '$2a$10$EixZaYVK1fsbw1ZfbX3OXePaWxn96p36WQoeG6Lruj3vjPGga31lW', '超级管理员', 'admin@example.com', 'super_admin', 1);

-- 验证插入
SELECT `id`, `username`, `real_name`, `email`, `role`, `status` FROM `admin` WHERE `username` = 'superadmin';
