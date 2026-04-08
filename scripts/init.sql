-- mailtools 数据库初始化脚本
-- 运行方式: mysql -u root -p < scripts/init.sql
-- 或者: mariadb -u root -p < scripts/init.sql

-- 创建数据库
CREATE DATABASE IF NOT EXISTS mailtools CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 使用数据库
USE mailtools;

-- 程序会自动创建表，这里只创建数据库即可
-- GORM 会自动创建 users 表

SELECT 'Database created successfully! Run the app to auto-migrate tables.' AS status;
