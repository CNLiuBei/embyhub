-- 修复embyhub用户的数据库权限

-- 授予embyhub用户访问embyhub数据库的权限
GRANT ALL PRIVILEGES ON DATABASE embyhub TO embyhub;

-- 授予embyhub用户访问public schema的权限
GRANT ALL PRIVILEGES ON SCHEMA public TO embyhub;

-- 授予embyhub用户访问所有表的权限
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO embyhub;

-- 授予embyhub用户访问所有序列的权限（用于自增ID）
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO embyhub;

-- 设置默认权限，确保新创建的表也有权限
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO embyhub;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO embyhub;
