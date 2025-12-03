-- Emby User Management System Initial Data
-- 默认角色、权限和管理员账号

-- 插入默认角色
INSERT INTO roles (role_name, description) VALUES
('超级管理员', '拥有系统所有权限'),
('普通管理员', '拥有用户管理和访问统计权限'),
('访客管理员', '仅拥有查看权限');

-- 插入默认权限
INSERT INTO permissions (permission_name, permission_key, description) VALUES
-- 用户管理权限
('查看用户', 'user:view', '查看用户列表和详情'),
('创建用户', 'user:create', '创建新用户'),
('编辑用户', 'user:edit', '编辑用户信息'),
('删除用户', 'user:delete', '删除用户'),
('导入用户', 'user:import', '批量导入用户'),
('导出用户', 'user:export', '批量导出用户'),

-- 角色管理权限
('查看角色', 'role:view', '查看角色列表和详情'),
('创建角色', 'role:create', '创建新角色'),
('编辑角色', 'role:edit', '编辑角色信息'),
('删除角色', 'role:delete', '删除角色'),

-- 权限管理权限
('查看权限', 'permission:view', '查看权限列表'),
('分配权限', 'permission:assign', '为角色分配权限'),

-- 系统设置权限
('查看系统设置', 'system:view', '查看系统配置'),
('修改系统设置', 'system:edit', '修改系统配置'),

-- 访问统计权限
('查看统计数据', 'stats:view', '查看访问统计数据'),
('导出统计数据', 'stats:export', '导出访问日志'),

-- Emby同步权限
('查看同步状态', 'emby:view', '查看Emby同步状态'),
('执行同步', 'emby:sync', '手动触发Emby数据同步'),
('配置Emby', 'emby:config', '配置Emby连接参数'),

-- 卡密管理权限
('查看卡密', 'cardkey:view', '查看卡密列表'),
('生成卡密', 'cardkey:create', '生成新卡密'),
('删除卡密', 'cardkey:delete', '删除卡密'),
('导出卡密', 'cardkey:export', '导出卡密列表');

-- 为超级管理员分配所有权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT 1, permission_id FROM permissions;

-- 为普通管理员分配部分权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT 2, permission_id FROM permissions
WHERE permission_key IN (
    'user:view', 'user:create', 'user:edit', 'user:delete',
    'stats:view', 'stats:export',
    'emby:view', 'emby:sync'
);

-- 为访客管理员分配查看权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT 3, permission_id FROM permissions
WHERE permission_key IN (
    'user:view', 'role:view', 'permission:view',
    'system:view', 'stats:view', 'emby:view'
);

-- 插入默认超级管理员账号
-- 用户名：admin
-- 密码：Liubei00（bcrypt加密后的哈希值）
INSERT INTO users (username, password_hash, email, role_id, status) VALUES
('admin', '$2a$10$NIDdLWXZi/0cv3yuQcoyjulmEOqynzjUQgsjtLxrWypD33wGClaX6', 'admin@embyhub.com', 1, 1);

-- 插入默认系统配置
INSERT INTO system_configs (config_key, config_value, description) VALUES
('emby_server_url', 'http://localhost:8096', 'Emby服务器地址'),
('emby_api_key', '', 'Emby API密钥'),
('emby_sync_interval', '3600', 'Emby数据同步周期（秒）'),
('jwt_secret', 'emby-ums-secret-key-change-in-production', 'JWT签名密钥'),
('jwt_expire_hours', '24', 'JWT Token过期时间（小时）'),
('password_min_length', '6', '密码最小长度'),
('log_retention_days', '30', '日志保留天数'),
('session_timeout_minutes', '120', '管理员会话超时时间（分钟）');

-- 插入测试访问记录（可选）
INSERT INTO access_records (user_id, resource, ip_address, device_info) VALUES
(1, '电影库/复仇者联盟', '192.168.1.100', 'Chrome/Windows 10'),
(1, '电视剧库/权力的游戏 S01E01', '192.168.1.100', 'Chrome/Windows 10'),
(1, '音乐库/周杰伦专辑', '192.168.1.100', 'Chrome/Windows 10');
