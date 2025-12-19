#!/bin/bash
# EmbyHub - 打包脚本
# 用法: ./build.sh [linux|darwin] [amd64|arm64]

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 默认目标平台
TARGET_OS=${1:-linux}
TARGET_ARCH=${2:-amd64}

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  EmbyHub - 打包脚本${NC}"
echo -e "${GREEN}  目标平台: ${TARGET_OS}/${TARGET_ARCH}${NC}"
echo -e "${GREEN}========================================${NC}"

# 项目根目录
PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
BUILD_DIR="${PROJECT_DIR}/dist"
VERSION=$(date +%Y%m%d%H%M%S)

# 清理旧的构建
echo -e "${YELLOW}[1/5] 清理旧的构建...${NC}"
rm -rf "${BUILD_DIR}"
mkdir -p "${BUILD_DIR}/feiniu-user-system"

# 构建前端
echo -e "${YELLOW}[2/5] 构建前端...${NC}"
cd "${PROJECT_DIR}/frontend"
npm install --legacy-peer-deps
npm run build

# 复制前端构建产物
mkdir -p "${BUILD_DIR}/feiniu-user-system/frontend"
cp -r dist/* "${BUILD_DIR}/feiniu-user-system/frontend/"

# 构建后端
echo -e "${YELLOW}[3/5] 构建后端...${NC}"
cd "${PROJECT_DIR}/backend"

# 主服务
echo "  - 构建主服务 (server)..."
CGO_ENABLED=0 GOOS=${TARGET_OS} GOARCH=${TARGET_ARCH} go build -ldflags="-s -w" -o "${BUILD_DIR}/feiniu-user-system/server" ./cmd/server

# Emby代理服务
echo "  - 构建Emby代理 (emby-proxy)..."
CGO_ENABLED=0 GOOS=${TARGET_OS} GOARCH=${TARGET_ARCH} go build -ldflags="-s -w" -o "${BUILD_DIR}/feiniu-user-system/emby-proxy" ./cmd/emby-proxy

# 复制配置文件
echo -e "${YELLOW}[4/5] 复制配置文件...${NC}"
mkdir -p "${BUILD_DIR}/feiniu-user-system/config"
mkdir -p "${BUILD_DIR}/feiniu-user-system/logs"
mkdir -p "${BUILD_DIR}/feiniu-user-system/uploads/avatars"
mkdir -p "${BUILD_DIR}/feiniu-user-system/uploads/logo"

# 复制配置模板
cp "${PROJECT_DIR}/backend/config/config.yaml" "${BUILD_DIR}/feiniu-user-system/config/config.yaml.example"

# 复制logo文件(如果存在)
if [ -d "${PROJECT_DIR}/backend/uploads/logo" ]; then
    cp -r "${PROJECT_DIR}/backend/uploads/logo/"* "${BUILD_DIR}/feiniu-user-system/uploads/logo/" 2>/dev/null || true
fi

# 复制静态文件目录
if [ -d "${PROJECT_DIR}/backend/static" ]; then
    cp -r "${PROJECT_DIR}/backend/static" "${BUILD_DIR}/feiniu-user-system/"
fi

# 创建部署脚本
echo -e "${YELLOW}[5/5] 创建部署脚本...${NC}"

# 创建一键部署脚本
cat > "${BUILD_DIR}/feiniu-user-system/deploy.sh" << 'DEPLOY_SCRIPT'
#!/bin/bash
# EmbyHub - 一键部署脚本
# 用法: ./deploy.sh

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

INSTALL_DIR="/opt/feiniu-user-system"
SERVICE_USER="feiniu"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  EmbyHub - 一键部署${NC}"
echo -e "${GREEN}========================================${NC}"

# 检查root权限
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}请使用 root 权限运行此脚本${NC}"
    echo "sudo ./deploy.sh"
    exit 1
fi

# 检测系统
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
else
    OS=$(uname -s)
fi

echo -e "${BLUE}检测到系统: ${OS}${NC}"

# 安装依赖
install_deps() {
    echo -e "${YELLOW}[1/7] 安装系统依赖...${NC}"
    case $OS in
        ubuntu|debian)
            apt-get update
            apt-get install -y postgresql postgresql-contrib redis-server nginx
            ;;
        centos|rhel|fedora)
            if command -v dnf &> /dev/null; then
                dnf install -y postgresql-server postgresql-contrib redis nginx
            else
                yum install -y postgresql-server postgresql-contrib redis nginx
            fi
            # 初始化PostgreSQL
            postgresql-setup --initdb 2>/dev/null || true
            ;;
        *)
            echo -e "${YELLOW}未知系统，请手动安装: PostgreSQL, Redis, Nginx${NC}"
            ;;
    esac
}

# 配置PostgreSQL
setup_postgres() {
    echo -e "${YELLOW}[2/7] 配置PostgreSQL...${NC}"
    
    # 启动服务
    systemctl enable postgresql
    systemctl start postgresql
    
    # 创建数据库和用户
    sudo -u postgres psql -c "CREATE USER fnuser WITH PASSWORD 'fnuser123';" 2>/dev/null || echo "用户已存在"
    sudo -u postgres psql -c "CREATE DATABASE feiniu_user OWNER fnuser;" 2>/dev/null || echo "数据库已存在"
    sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE feiniu_user TO fnuser;"
    
    # 配置认证方式
    PG_HBA=$(sudo -u postgres psql -t -c "SHOW hba_file;" | tr -d ' ')
    if ! grep -q "fnuser" "$PG_HBA" 2>/dev/null; then
        echo "host    feiniu_user    fnuser    127.0.0.1/32    md5" >> "$PG_HBA"
        systemctl reload postgresql
    fi
    
    echo -e "${GREEN}PostgreSQL 配置完成${NC}"
}

# 配置Redis
setup_redis() {
    echo -e "${YELLOW}[3/7] 配置Redis...${NC}"
    systemctl enable redis-server 2>/dev/null || systemctl enable redis
    systemctl start redis-server 2>/dev/null || systemctl start redis
    echo -e "${GREEN}Redis 配置完成${NC}"
}

# 安装应用
install_app() {
    echo -e "${YELLOW}[4/7] 安装应用...${NC}"
    
    # 创建安装目录
    mkdir -p ${INSTALL_DIR}
    
    # 复制文件
    SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
    cp -r "${SCRIPT_DIR}"/* ${INSTALL_DIR}/
    
    # 创建配置文件
    if [ ! -f "${INSTALL_DIR}/config/config.yaml" ]; then
        cp "${INSTALL_DIR}/config/config.yaml.example" "${INSTALL_DIR}/config/config.yaml"
    fi
    
    # 设置权限
    chmod +x ${INSTALL_DIR}/server
    chmod +x ${INSTALL_DIR}/emby-proxy
    
    echo -e "${GREEN}应用安装完成${NC}"
}

# 创建systemd服务
create_services() {
    echo -e "${YELLOW}[5/7] 创建系统服务...${NC}"
    
    # 主服务
    cat > /etc/systemd/system/feiniu-server.service << EOF
[Unit]
Description=Feiniu User System Server
After=network.target postgresql.service redis.service

[Service]
Type=simple
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/server
Restart=always
RestartSec=5
StandardOutput=append:${INSTALL_DIR}/logs/server.log
StandardError=append:${INSTALL_DIR}/logs/server-error.log

[Install]
WantedBy=multi-user.target
EOF

    # Emby代理服务
    cat > /etc/systemd/system/feiniu-emby-proxy.service << EOF
[Unit]
Description=Feiniu Emby Proxy
After=network.target feiniu-server.service

[Service]
Type=simple
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/emby-proxy
Restart=always
RestartSec=5
StandardOutput=append:${INSTALL_DIR}/logs/emby-proxy.log
StandardError=append:${INSTALL_DIR}/logs/emby-proxy-error.log

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    echo -e "${GREEN}系统服务创建完成${NC}"
}

# 配置Nginx
setup_nginx() {
    echo -e "${YELLOW}[6/7] 配置Nginx...${NC}"
    
    cat > /etc/nginx/sites-available/feiniu 2>/dev/null || cat > /etc/nginx/conf.d/feiniu.conf << 'NGINX_CONF'
server {
    listen 80;
    server_name _;
    
    # 前端静态文件
    location / {
        root /opt/feiniu-user-system/frontend;
        index index.html;
        try_files $uri $uri/ /index.html;
    }
    
    # 后端API
    location /api/ {
        proxy_pass http://127.0.0.1:54680;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # 静态资源
    location /uploads/ {
        proxy_pass http://127.0.0.1:54680;
    }
    
    location /static/ {
        proxy_pass http://127.0.0.1:54680;
    }
}

# Emby代理 (可选，如需要通过Nginx代理Emby)
server {
    listen 8097;
    server_name _;
    
    location / {
        proxy_pass http://127.0.0.1:54682;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_buffering off;
        proxy_request_buffering off;
    }
}
NGINX_CONF

    # 启用站点 (Debian/Ubuntu)
    if [ -d /etc/nginx/sites-enabled ]; then
        ln -sf /etc/nginx/sites-available/feiniu /etc/nginx/sites-enabled/
        rm -f /etc/nginx/sites-enabled/default 2>/dev/null || true
    fi
    
    nginx -t && systemctl reload nginx
    systemctl enable nginx
    
    echo -e "${GREEN}Nginx 配置完成${NC}"
}

# 启动服务
start_services() {
    echo -e "${YELLOW}[7/7] 启动服务...${NC}"
    
    systemctl enable feiniu-server
    systemctl enable feiniu-emby-proxy
    systemctl start feiniu-server
    sleep 2
    systemctl start feiniu-emby-proxy
    
    echo -e "${GREEN}服务启动完成${NC}"
}

# 显示信息
show_info() {
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  部署完成！${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo -e "${BLUE}访问地址:${NC}"
    echo "  - 前端: http://YOUR_IP"
    echo "  - 后端API: http://YOUR_IP:54680"
    echo "  - Emby代理: http://YOUR_IP:54682"
    echo ""
    echo -e "${BLUE}默认管理员:${NC}"
    echo "  - 用户名: admin"
    echo "  - 密码: Liubei00"
    echo ""
    echo -e "${BLUE}数据库信息:${NC}"
    echo "  - 数据库: feiniu_user"
    echo "  - 用户名: fnuser"
    echo "  - 密码: fnuser123"
    echo ""
    echo -e "${BLUE}配置文件:${NC}"
    echo "  - ${INSTALL_DIR}/config/config.yaml"
    echo ""
    echo -e "${BLUE}服务管理:${NC}"
    echo "  - systemctl status feiniu-server"
    echo "  - systemctl status feiniu-emby-proxy"
    echo "  - systemctl restart feiniu-server"
    echo ""
    echo -e "${YELLOW}重要: 请修改配置文件中的 Emby 服务器地址和 API Key${NC}"
}

# 主流程
main() {
    install_deps
    setup_postgres
    setup_redis
    install_app
    create_services
    setup_nginx
    start_services
    show_info
}

main "$@"
DEPLOY_SCRIPT

chmod +x "${BUILD_DIR}/feiniu-user-system/deploy.sh"


# 创建管理脚本
cat > "${BUILD_DIR}/feiniu-user-system/manage.sh" << 'MANAGE_SCRIPT'
#!/bin/bash
# EmbyHub - 管理脚本

INSTALL_DIR="/opt/feiniu-user-system"

case "$1" in
    start)
        echo "启动服务..."
        systemctl start feiniu-server
        systemctl start feiniu-emby-proxy
        echo "服务已启动"
        ;;
    stop)
        echo "停止服务..."
        systemctl stop feiniu-emby-proxy
        systemctl stop feiniu-server
        echo "服务已停止"
        ;;
    restart)
        echo "重启服务..."
        systemctl restart feiniu-server
        systemctl restart feiniu-emby-proxy
        echo "服务已重启"
        ;;
    status)
        echo "=== 主服务状态 ==="
        systemctl status feiniu-server --no-pager
        echo ""
        echo "=== Emby代理状态 ==="
        systemctl status feiniu-emby-proxy --no-pager
        ;;
    logs)
        echo "=== 最近日志 ==="
        tail -50 ${INSTALL_DIR}/logs/server.log
        ;;
    update)
        echo "更新应用..."
        SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
        systemctl stop feiniu-emby-proxy
        systemctl stop feiniu-server
        cp "${SCRIPT_DIR}/server" ${INSTALL_DIR}/server
        cp "${SCRIPT_DIR}/emby-proxy" ${INSTALL_DIR}/emby-proxy
        cp -r "${SCRIPT_DIR}/frontend/"* ${INSTALL_DIR}/frontend/
        systemctl start feiniu-server
        systemctl start feiniu-emby-proxy
        echo "更新完成"
        ;;
    *)
        echo "用法: $0 {start|stop|restart|status|logs|update}"
        exit 1
        ;;
esac
MANAGE_SCRIPT

chmod +x "${BUILD_DIR}/feiniu-user-system/manage.sh"

# 创建README
cat > "${BUILD_DIR}/feiniu-user-system/README.md" << 'README'
# EmbyHub

Emby/Jellyfin 用户管理系统，支持会员卡密、播放限制、速率控制等功能。

## 快速部署

```bash
# 一键部署 (需要root权限)
sudo ./deploy.sh
```

## 手动部署

### 1. 安装依赖
- PostgreSQL 12+
- Redis 6+
- Nginx

### 2. 配置数据库
```sql
CREATE USER fnuser WITH PASSWORD 'fnuser123';
CREATE DATABASE feiniu_user OWNER fnuser;
```

### 3. 修改配置
```bash
cp config/config.yaml.example config/config.yaml
vim config/config.yaml
```

### 4. 启动服务
```bash
./server &
./emby-proxy &
```

## 端口说明
- 54680: 主服务API
- 54681: 前端开发服务 (生产环境使用Nginx)
- 54682: Emby代理服务

## 默认账号
- 用户名: admin
- 密码: Liubei00

## 目录结构
```
├── server          # 主服务
├── emby-proxy      # Emby代理
├── frontend/       # 前端静态文件
├── config/         # 配置文件
├── logs/           # 日志目录
├── uploads/        # 上传文件
├── deploy.sh       # 一键部署脚本
└── manage.sh       # 管理脚本
```

## 管理命令
```bash
./manage.sh start    # 启动
./manage.sh stop     # 停止
./manage.sh restart  # 重启
./manage.sh status   # 状态
./manage.sh logs     # 查看日志
./manage.sh update   # 更新
```
README

# 打包
echo -e "${YELLOW}打包中...${NC}"
cd "${BUILD_DIR}"
tar -czvf "feiniu-user-system-${TARGET_OS}-${TARGET_ARCH}-${VERSION}.tar.gz" feiniu-user-system

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  打包完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "输出文件: ${BUILD_DIR}/feiniu-user-system-${TARGET_OS}-${TARGET_ARCH}-${VERSION}.tar.gz"
echo ""
echo -e "部署步骤:"
echo "  1. 上传到目标服务器"
echo "  2. tar -xzvf feiniu-user-system-*.tar.gz"
echo "  3. cd feiniu-user-system"
echo "  4. sudo ./deploy.sh"
