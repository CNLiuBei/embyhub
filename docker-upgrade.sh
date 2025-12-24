#!/bin/bash
# EmbyHub - Docker镜像升级脚本
# 用法: ./docker-upgrade.sh

set -e

# 配置
REMOTE_HOST="treuliu@10.10.10.68"
REMOTE_PATH="/vol1/1000/embyhub"
REMOTE_PASS="Liubei00"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
VERSION=$(date +%Y%m%d%H%M%S)

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  EmbyHub - Docker镜像升级${NC}"
echo -e "${GREEN}  目标服务器: ${REMOTE_HOST}${NC}"
echo -e "${GREEN}  部署路径: ${REMOTE_PATH}${NC}"
echo -e "${GREEN}========================================${NC}"

# 检查sshpass
if ! command -v sshpass &> /dev/null; then
    echo -e "${YELLOW}安装 sshpass...${NC}"
    brew install hudochenkov/sshpass/sshpass 2>/dev/null || {
        echo -e "${RED}请先安装 sshpass: brew install hudochenkov/sshpass/sshpass${NC}"
        exit 1
    }
fi

# 构建后端镜像
echo -e "${YELLOW}[1/6] 构建后端镜像...${NC}"
docker build -t embyhub-backend:latest -f "${PROJECT_DIR}/backend/Dockerfile" "${PROJECT_DIR}/backend"

# 构建前端镜像
echo -e "${YELLOW}[2/6] 构建前端镜像...${NC}"
docker build -t embyhub-frontend:latest -f "${PROJECT_DIR}/frontend/Dockerfile" "${PROJECT_DIR}/frontend"

# 导出镜像
echo -e "${YELLOW}[3/6] 导出镜像...${NC}"
mkdir -p "${PROJECT_DIR}/dist"
docker save embyhub-backend:latest | gzip > "${PROJECT_DIR}/dist/embyhub-backend.tar.gz"
docker save embyhub-frontend:latest | gzip > "${PROJECT_DIR}/dist/embyhub-frontend.tar.gz"

# 上传镜像到服务器
echo -e "${YELLOW}[4/6] 上传镜像到服务器...${NC}"
sshpass -p "${REMOTE_PASS}" scp -o StrictHostKeyChecking=no \
    "${PROJECT_DIR}/dist/embyhub-backend.tar.gz" \
    "${PROJECT_DIR}/dist/embyhub-frontend.tar.gz" \
    "${REMOTE_HOST}:${REMOTE_PATH}/"

# 上传docker-compose.yml
echo -e "${YELLOW}[5/6] 上传docker-compose配置...${NC}"
sshpass -p "${REMOTE_PASS}" scp -o StrictHostKeyChecking=no \
    "${PROJECT_DIR}/docker-compose.yml" \
    "${REMOTE_HOST}:${REMOTE_PATH}/"

# 在服务器上执行升级
echo -e "${YELLOW}[6/6] 在服务器上执行升级...${NC}"
sshpass -p "${REMOTE_PASS}" ssh -o StrictHostKeyChecking=no "${REMOTE_HOST}" << 'REMOTE_SCRIPT'
cd /vol1/1000/embyhub

echo "加载新镜像..."
docker load < embyhub-backend.tar.gz
docker load < embyhub-frontend.tar.gz

echo "停止旧容器..."
docker-compose down || true

echo "启动新容器..."
docker-compose up -d

echo "清理镜像文件..."
rm -f embyhub-backend.tar.gz embyhub-frontend.tar.gz

echo "查看容器状态..."
docker-compose ps

echo "升级完成!"
REMOTE_SCRIPT

# 清理本地临时文件
rm -f "${PROJECT_DIR}/dist/embyhub-backend.tar.gz"
rm -f "${PROJECT_DIR}/dist/embyhub-frontend.tar.gz"

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  升级完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "访问地址: http://10.10.10.68:3000"
