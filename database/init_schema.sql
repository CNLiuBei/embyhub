-- Emby User Management System Database Schema
-- PostgreSQL 14+

-- 删除已存在的表（按依赖关系逆序删除）
DROP TABLE IF EXISTS access_records CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS system_configs CASCADE;

-- 角色表
CREATE TABLE roles (
    role_id SERIAL PRIMARY KEY,
    role_name VARCHAR(50) NOT NULL UNIQUE,
    description VARCHAR(200),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 权限表
CREATE TABLE permissions (
    permission_id SERIAL PRIMARY KEY,
    permission_name VARCHAR(50) NOT NULL UNIQUE,
    permission_key VARCHAR(50) NOT NULL UNIQUE,
    description VARCHAR(200)
);

-- 角色-权限关联表（多对多）
CREATE TABLE role_permissions (
    role_id INT NOT NULL,
    permission_id INT NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(role_id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(permission_id) ON DELETE CASCADE
);

-- 用户表
CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password_hash VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE,
    emby_user_id VARCHAR(50),
    role_id INT NOT NULL,
    status SMALLINT NOT NULL DEFAULT 1, -- 1-启用，0-禁用
    vip_level SMALLINT NOT NULL DEFAULT 0, -- 0-普通用户，1-VIP
    vip_expire_at TIMESTAMP, -- VIP过期时间
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (role_id) REFERENCES roles(role_id)
);

-- 访问记录表
CREATE TABLE access_records (
    record_id BIGSERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    access_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resource VARCHAR(200),
    ip_address VARCHAR(50),
    device_info VARCHAR(100),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- 系统配置表
CREATE TABLE system_configs (
    config_key VARCHAR(50) PRIMARY KEY,
    config_value TEXT NOT NULL,
    description VARCHAR(200),
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引优化查询性能
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_emby_user_id ON users(emby_user_id);
CREATE INDEX idx_access_records_user_id ON access_records(user_id);
CREATE INDEX idx_access_records_access_time ON access_records(access_time);
CREATE INDEX idx_access_records_user_time ON access_records(user_id, access_time);

-- 卡密表
CREATE TABLE card_keys (
    id SERIAL PRIMARY KEY,
    card_code VARCHAR(32) NOT NULL UNIQUE,
    card_type SMALLINT NOT NULL DEFAULT 1, -- 1=注册码 2=VIP升级码
    duration INT NOT NULL DEFAULT 30, -- 有效期（天）
    status SMALLINT NOT NULL DEFAULT 1, -- 0=已禁用 1=未使用 2=已使用
    used_by INT REFERENCES users(user_id),
    used_at TIMESTAMP,
    expire_at TIMESTAMP,
    remark VARCHAR(200),
    created_by INT NOT NULL REFERENCES users(user_id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_card_keys_card_code ON card_keys(card_code);
CREATE INDEX idx_card_keys_status ON card_keys(status);
CREATE INDEX idx_card_keys_card_type ON card_keys(card_type);

-- 添加注释
COMMENT ON TABLE users IS 'Emby用户信息表';
COMMENT ON TABLE roles IS '角色信息表';
COMMENT ON TABLE permissions IS '权限信息表';
COMMENT ON TABLE role_permissions IS '角色权限关联表';
COMMENT ON TABLE access_records IS '用户访问记录表';
COMMENT ON TABLE system_configs IS '系统配置表';
