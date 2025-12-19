#!/bin/bash
# EmbyHub - Docker部署包打包脚本

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  EmbyHub - Docker部署包打包${NC}"
echo -e "${GREEN}========================================${NC}"

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
BUILD_DIR="${PROJECT_DIR}/dist-docker"
VERSION=$(date +%Y%m%d%H%M%S)

# 清理
rm -rf "${BUILD_DIR}"
mkdir -p "${BUILD_DIR}/feiniu-docker"

# 构建前端
echo -e "${YELLOW}[1/4] 构建前端...${NC}"
cd "${PROJECT_DIR}/frontend"
npm install --legacy-peer-deps
npm run build
mkdir -p "${BUILD_DIR}/feiniu-docker/frontend"
cp -r dist/* "${BUILD_DIR}/feiniu-docker/frontend/"

# 构建后端 (Linux amd64)
echo -e "${YELLOW}[2/4] 构建后端...${NC}"
cd "${PROJECT_DIR}/backend"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "${BUILD_DIR}/feiniu-docker/server" ./cmd/server
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "${BUILD_DIR}/feiniu-docker/emby-proxy" ./cmd/emby-proxy

# 复制Docker配置
echo -e "${YELLOW}[3/4] 复制配置文件...${NC}"
cp "${PROJECT_DIR}/docker/docker-compose.simple.yml" "${BUILD_DIR}/feiniu-docker/docker-compose.yml"
cp "${PROJECT_DIR}/docker/nginx.conf" "${BUILD_DIR}/feiniu-docker/"
cp "${PROJECT_DIR}/docker/start.sh" "${BUILD_DIR}/feiniu-docker/"
mkdir -p "${BUILD_DIR}/feiniu-docker/config"
cp "${PROJECT_DIR}/docker/config/config.yaml" "${BUILD_DIR}/feiniu-docker/config/config.yaml"

mkdir -p "${BUILD_DIR}/feiniu-docker/data/postgres"
mkdir -p "${BUILD_DIR}/feiniu-docker/data/redis"
mkdir -p "${BUILD_DIR}/feiniu-docker/logs"
mkdir -p "${BUILD_DIR}/feiniu-docker/uploads/avatars"
mkdir -p "${BUILD_DIR}/feiniu-docker/uploads/logo"

chmod +x "${BUILD_DIR}/feiniu-docker/server"
chmod +x "${BUILD_DIR}/feiniu-docker/emby-proxy"
chmod +x "${BUILD_DIR}/feiniu-docker/start.sh"

# 设置前端文件权限，避免nginx权限问题
chmod -R 755 "${BUILD_DIR}/feiniu-docker/frontend"

# 创建部署说明
cat > "${BUILD_DIR}/feiniu-docker/README.md" << 'EOF'
# EmbyHub - Docker部署

## 快速开始

```bash
# 1. 修改配置文件
nano config/config.yaml

# 2. 启动服务
docker-compose up -d

# 3. 查看日志
docker-compose logs -f
```

## 访问地址
- 前端: http://YOUR_IP:54681
- 后端API: http://YOUR_IP:54680
- Emby代理: http://YOUR_IP:54682

## 默认账号
- 用户名: admin
- 密码: admin123

## 配置说明
编辑 `config/config.yaml`，修改以下内容：
- `emby.base_url`: 你的Emby服务器地址
- `emby.api_key`: Emby API密钥

## 管理命令
```bash
docker-compose up -d      # 启动
docker-compose down       # 停止
docker-compose restart    # 重启
docker-compose logs -f    # 查看日志
```
EOF

# 打包
echo -e "${YELLOW}[4/4] 打包...${NC}"
cd "${BUILD_DIR}"
tar -czvf "feiniu-docker-${VERSION}.tar.gz" feiniu-docker

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  打包完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "输出文件: ${BUILD_DIR}/feiniu-docker-${VERSION}.tar.gz"
echo ""
echo "部署步骤:"
echo "  1. 上传到服务器"
echo "  2. tar -xzvf feiniu-docker-*.tar.gz"
echo "  3. cd feiniu-docker"
echo "  4. nano config/config.yaml  # 修改Emby配置"
echo "  5. docker-compose up -d"
echo ""
echo "端口说明:"
echo "  - 54681: 前端"
echo "  - 54680: 后端API"
echo "  - 54682: Emby代理"
