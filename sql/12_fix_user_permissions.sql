-- 修复用户权限路径
USE iam_db;

-- 更新用户权限路径,移除尾部斜杠,使用前缀匹配
UPDATE permissions SET path = '/api/v1/users*' WHERE code = 'user:view';
UPDATE permissions SET path = '/api/v1/users' WHERE code = 'user:create';
UPDATE permissions SET path = '/api/v1/users*' WHERE code = 'user:edit';
UPDATE permissions SET path = '/api/v1/users*' WHERE code = 'user:delete';
UPDATE permissions SET path = '/api/v1/users*' WHERE code = 'user:assign_role';
UPDATE permissions SET path = '/api/v1/users*' WHERE code = 'user:change_status';
