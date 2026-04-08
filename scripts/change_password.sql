-- 修改用户密码的脚本
-- 运行方式: mysql -u root -p < scripts/change_password.sql
-- 然后输入新密码

-- 使用数据库
USE mailtools;

-- 更新 admin 密码 (替换下面的 'your_new_password' 为实际密码)
-- 密码使用 bcrypt hash
UPDATE users SET password_hash = '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy' WHERE username = 'admin';

SELECT 'Password updated for admin user' AS status;
