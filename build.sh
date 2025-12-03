#!/bin/bash
# Emby Hub 打包脚本

set -e

VERSION=${1:-"1.0.0"}
OUTPUT_DIR="dist/embyhub-${VERSION}"

echo "=== Emby Hub 打包脚本 v${VERSION} ==="

# 清理旧的构建
rm -rf dist
mkdir -p ${OUTPUT_DIR}

# 1. 构建前端
echo ">>> 构建前端..."
cd frontend
npm install
npm run build
cd ..

# 复制前端构建产物
cp -r frontend/dist ${OUTPUT_DIR}/frontend

# 2. 构建后端（多平台）
echo ">>> 构建后端..."
cd backend

# Linux AMD64
echo "  - Linux AMD64"
GOOS=linux GOARCH=amd64 go build -o ../dist/embyhub-${VERSION}/embyhub-linux-amd64 ./cmd/...

# Linux ARM64
echo "  - Linux ARM64"
GOOS=linux GOARCH=arm64 go build -o ../dist/embyhub-${VERSION}/embyhub-linux-arm64 ./cmd/...

# macOS AMD64
echo "  - macOS AMD64"
GOOS=darwin GOARCH=amd64 go build -o ../dist/embyhub-${VERSION}/embyhub-darwin-amd64 ./cmd/...

# macOS ARM64 (Apple Silicon)
echo "  - macOS ARM64"
GOOS=darwin GOARCH=arm64 go build -o ../dist/embyhub-${VERSION}/embyhub-darwin-arm64 ./cmd/...

# Windows AMD64
echo "  - Windows AMD64"
GOOS=windows GOARCH=amd64 go build -o ../dist/embyhub-${VERSION}/embyhub-windows-amd64.exe ./cmd/...

cd ..

# 3. 复制数据库脚本
echo ">>> 复制数据库脚本..."
mkdir -p ${OUTPUT_DIR}/database
cp database/init_schema.sql ${OUTPUT_DIR}/database/
cp database/init_data.sql ${OUTPUT_DIR}/database/

# 4. 创建配置目录
mkdir -p ${OUTPUT_DIR}/config
mkdir -p ${OUTPUT_DIR}/logs

# 5. 创建启动脚本
echo ">>> 创建启动脚本..."

# Linux/macOS 启动脚本
cat > ${OUTPUT_DIR}/start.sh << 'SCRIPT'
#!/bin/bash
# Emby Hub 启动脚本

# 检测系统和架构
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
esac

BINARY="./embyhub-${OS}-${ARCH}"

if [ ! -f "$BINARY" ]; then
    echo "错误: 找不到可执行文件 $BINARY"
    exit 1
fi

chmod +x $BINARY
echo "启动 Emby Hub..."
$BINARY
SCRIPT
chmod +x ${OUTPUT_DIR}/start.sh

# Windows 启动脚本
cat > ${OUTPUT_DIR}/start.bat << 'SCRIPT'
@echo off
echo 启动 Emby Hub...
embyhub-windows-amd64.exe
pause
SCRIPT

# 6. 创建部署说明
cat > ${OUTPUT_DIR}/README.md << 'README'
# Emby Hub 部署指南

## 系统要求

- PostgreSQL 12+
- Redis 6+ (可选，用于缓存)
- Node.js 18+ (仅开发时需要)

## 快速开始

### 1. 创建数据库

```bash
# 登录 PostgreSQL
psql -U postgres

# 创建数据库
CREATE DATABASE embyhub;
\q
```

### 2. 启动后端服务

**Linux/macOS:**
```bash
chmod +x start.sh
./start.sh
```

**Windows:**
```
双击 start.bat
```

### 3. 访问安装向导

打开浏览器访问: `http://localhost:8080/setup`

按照向导完成：
1. 授权验证
2. 数据库配置
3. Emby 服务器配置
4. 邮件服务配置
5. 管理员账户设置

### 4. 前端部署（可选）

如需单独部署前端，将 `frontend/` 目录部署到 Web 服务器（如 Nginx）。

Nginx 配置示例：
```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        root /path/to/embyhub/frontend;
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 配置文件

配置文件位于 `config/config.yaml`，首次运行会通过安装向导自动生成。

## 目录结构

```
embyhub/
├── embyhub-*          # 后端可执行文件
├── frontend/          # 前端静态文件
├── database/          # 数据库脚本
├── config/            # 配置文件目录
├── logs/              # 日志目录
├── start.sh           # Linux/macOS 启动脚本
└── start.bat          # Windows 启动脚本
```

## 常见问题

**Q: 端口被占用？**
A: 修改 config.yaml 中的 server.port

**Q: 数据库连接失败？**
A: 检查 PostgreSQL 服务是否启动，用户名密码是否正确

**Q: Emby 连接失败？**
A: 确认 Emby 服务器地址和 API Key 正确

## 技术支持

如有问题，请联系技术支持。
README

# 7. 打包
echo ">>> 打包..."
cd dist
tar -czvf embyhub-${VERSION}.tar.gz embyhub-${VERSION}
cd ..

echo ""
echo "=== 打包完成 ==="
echo "输出目录: ${OUTPUT_DIR}"
echo "压缩包: dist/embyhub-${VERSION}.tar.gz"
echo ""
echo "文件列表:"
ls -la ${OUTPUT_DIR}/
