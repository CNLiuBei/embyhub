#!/bin/bash
# EmbyHub - Docker部署升级脚本
# 用法: ./deploy-docker.sh

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

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  EmbyHub - Docker部署升级${NC}"
echo -e "${GREEN}  目标服务器: ${REMOTE_HOST}${NC}"
echo -e "${GREEN}  部署路径: ${REMOTE_PATH}${NC}"
echo -e "${GREEN}========================================${NC}"

# 打包源码（排除不需要的文件）
echo -e "${YELLOW}[1/4] 打包源码...${NC}"
tar -czvf /tmp/embyhub-src.tar.gz \
    --exclude='node_modules' \
    --exclude='dist' \
    --exclude='.git' \
    --exclude='*.log' \
    -C "${PROJECT_DIR}" \
    backend frontend docker-compose.yml

# 上传到服务器
echo -e "${YELLOW}[2/4] 上传到服务器...${NC}"
sshpass -p "${REMOTE_PASS}" scp -o StrictHostKeyChecking=no \
    /tmp/embyhub-src.tar.gz \
    "${REMOTE_HOST}:${REMOTE_PATH}/"

# 在服务器上执行升级
echo -e "${YELLOW}[3/4] 在服务器上执行升级...${NC}"
sshpass -p "${REMOTE_PASS}" ssh -o StrictHostKeyChecking=no "${REMOTE_HOST}" << REMOTE_SCRIPT
cd ${REMOTE_PATH}

echo "解压源码..."
tar -xzvf embyhub-src.tar.gz

echo "停止旧容器..."
docker-compose down || true

echo "重新构建并启动..."
docker-compose up -d --build

echo "清理..."
rm -f embyhub-src.tar.gz

echo "查看容器状态..."
docker-compose ps
REMOTE_SCRIPT

# 清理本地临时文件
echo -e "${YELLOW}[4/4] 清理临时文件...${NC}"
rm -f /tmp/embyhub-src.tar.gz

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  升级完成！${NC}"
echo -e "${GREEN}========================================${NC}"
