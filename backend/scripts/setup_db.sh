#!/bin/bash

# 飞牛影视用户管理系统 - 数据库初始化脚本
# 用法: ./setup_db.sh

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 数据库配置(与config.yaml保持一致)
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-fnuser}"
DB_PASS="${DB_PASS:-fnuser123}"
DB_NAME="${DB_NAME:-feiniu_user}"
PG_ADMIN_USER="${PG_ADMIN_USER:-postgres}"

log_info "开始初始化数据库..."

# 检查PostgreSQL是否可用
if ! command -v psql &> /dev/null; then
    log_error "未找到psql命令,请先安装PostgreSQL客户端"
    exit 1
fi

# 创建数据库用户
log_info "创建数据库用户 $DB_USER..."
sudo -u $PG_ADMIN_USER psql -c "CREATE USER $DB_USER WITH PASSWORD '$DB_PASS';" 2>/dev/null || log_warn "用户可能已存在"

# 创建数据库
log_info "创建数据库 $DB_NAME..."
sudo -u $PG_ADMIN_USER psql -c "CREATE DATABASE $DB_NAME OWNER $DB_USER;" 2>/dev/null || log_warn "数据库可能已存在"

# 授权
log_info "授权..."
sudo -u $PG_ADMIN_USER psql -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;"

# 配置pg_hba.conf (如果需要)
log_info "检查pg_hba.conf配置..."
PG_HBA=$(sudo -u $PG_ADMIN_USER psql -t -c "SHOW hba_file;" | tr -d ' ')
if [ -f "$PG_HBA" ]; then
    if ! grep -q "$DB_USER" "$PG_HBA"; then
        log_info "添加用户认证规则..."
        echo "host    $DB_NAME    $DB_USER    127.0.0.1/32    md5" | sudo tee -a "$PG_HBA" > /dev/null
        echo "host    $DB_NAME    $DB_USER    ::1/128         md5" | sudo tee -a "$PG_HBA" > /dev/null
        sudo -u $PG_ADMIN_USER psql -c "SELECT pg_reload_conf();"
    fi
fi

log_info "数据库初始化完成!"
echo ""
echo "数据库连接信息:"
echo "  主机: $DB_HOST"
echo "  端口: $DB_PORT"
echo "  用户: $DB_USER"
echo "  密码: $DB_PASS"
echo "  数据库: $DB_NAME"
echo ""
log_info "请确保 backend/config/config.yaml 中的数据库配置与以上信息一致"
