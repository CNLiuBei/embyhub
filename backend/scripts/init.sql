-- Emby用户管理系统 - 数据库初始化脚本
-- 执行方式: psql -U postgres -d emby_user -f init.sql

-- 创建管理员账户 (密码: Admin123!)
-- bcrypt hash for "Admin123!": $2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqxK8MbKCxqMnBfSQCVBVbHGnFGDi
INSERT INTO users (
    id, username, email, password, nickname, avatar, status, role, member_level, register_ip, created_at, updated_at
) VALUES (
    'a0000000-0000-0000-0000-000000000001',
    'admin',
    'admin@emby.local',
    '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqxK8MbKCxqMnBfSQCVBVbHGnFGDi',
    '管理员',
    '',
    1,
    1,  -- 管理员角色
    2,  -- 年卡会员
    '127.0.0.1',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- 创建测试普通用户 (密码: User123!)
INSERT INTO users (
    id, username, email, password, nickname, avatar, status, role, member_level, register_ip, created_at, updated_at
) VALUES (
    'a0000000-0000-0000-0000-000000000002',
    'testuser',
    'user@emby.local',
    '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqxK8MbKCxqMnBfSQCVBVbHGnFGDi',
    '测试用户',
    '',
    1,
    0,  -- 普通用户角色
    0,  -- 普通用户
    '127.0.0.1',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_member_level ON users(member_level);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

CREATE INDEX IF NOT EXISTS idx_login_logs_user_id ON login_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_login_logs_created_at ON login_logs(created_at);

CREATE INDEX IF NOT EXISTS idx_watch_histories_user_id ON watch_histories(user_id);
CREATE INDEX IF NOT EXISTS idx_watch_histories_video_id ON watch_histories(video_id);

CREATE INDEX IF NOT EXISTS idx_favorites_user_id ON favorites(user_id);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);

CREATE INDEX IF NOT EXISTS idx_operation_logs_admin_id ON operation_logs(admin_id);
CREATE INDEX IF NOT EXISTS idx_operation_logs_created_at ON operation_logs(created_at);

CREATE INDEX IF NOT EXISTS idx_member_orders_user_id ON member_orders(user_id);
CREATE INDEX IF NOT EXISTS idx_member_orders_order_no ON member_orders(order_no);

-- 卡密索引
CREATE INDEX IF NOT EXISTS idx_cards_code ON cards(code);
CREATE INDEX IF NOT EXISTS idx_cards_batch_no ON cards(batch_no);
CREATE INDEX IF NOT EXISTS idx_cards_status ON cards(status);
CREATE INDEX IF NOT EXISTS idx_card_batches_batch_no ON card_batches(batch_no);
