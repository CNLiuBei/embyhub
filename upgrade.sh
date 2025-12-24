#!/bin/bash
# EmbyHub Docker 升级脚本
set -e

DEPLOY_DIR="/vol1/1000/embyhub"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "=========================================="
echo "  EmbyHub Docker 升级"
echo "=========================================="

# 停止服务
echo "[1/4] 停止服务..."
docker stop feiniu-server feiniu-emby-proxy feiniu-nginx 2>/dev/null || true

# 备份旧文件
echo "[2/4] 备份旧文件..."
BACKUP_DIR="${DEPLOY_DIR}/backup_$(date +%Y%m%d%H%M%S)"
mkdir -p "$BACKUP_DIR"
cp -f "${DEPLOY_DIR}/server" "$BACKUP_DIR/" 2>/dev/null || true
cp -f "${DEPLOY_DIR}/emby-proxy" "$BACKUP_DIR/" 2>/dev/null || true

# 更新文件
echo "[3/4] 更新文件..."
cp -f "${SCRIPT_DIR}/server" "${DEPLOY_DIR}/"
cp -f "${SCRIPT_DIR}/emby-proxy" "${DEPLOY_DIR}/"
cp -rf "${SCRIPT_DIR}/frontend/"* "${DEPLOY_DIR}/dist/feiniu-user-system/frontend/" 2>/dev/null || \
cp -rf "${SCRIPT_DIR}/frontend/"* "${DEPLOY_DIR}/frontend/" 2>/dev/null || true

chmod +x "${DEPLOY_DIR}/server" "${DEPLOY_DIR}/emby-proxy"

# 重启服务
echo "[4/4] 重启服务..."
docker start feiniu-server feiniu-emby-proxy feiniu-nginx

echo ""
echo "=========================================="
echo "  升级完成！"
echo "=========================================="
docker ps | grep feiniu
