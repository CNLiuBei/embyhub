#!/bin/bash
# 飞牛用户系统 - Docker一键部署脚本

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  飞牛用户系统 - Docker部署${NC}"
echo -e "${GREEN}========================================${NC}"

# 检查Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Docker 未安装，请先安装 Docker${NC}"
    exit 1
fi

if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo -e "${RED}Docker Compose 未安装${NC}"
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

# 检查前端文件
if [ ! -d "../dist/feiniu-user-system/frontend" ]; then
    echo -e "${YELLOW}前端文件不存在，请先构建前端${NC}"
    echo "运行: cd ../frontend && npm install && npm run build"
    exit 1
fi

# 创建必要目录
mkdir -p config

# 复制配置文件
if [ ! -f "config/config.yaml" ]; then
    cp config.yaml.example config/config.yaml 2>/dev/null || cp config/config.yaml config/config.yaml
    echo -e "${YELLOW}请编辑 config/config.yaml 配置文件${NC}"
fi

# 启动服务
echo -e "${YELLOW}启动服务...${NC}"

# 使用 docker compose (新版) 或 docker-compose (旧版)
if docker compose version &> /dev/null; then
    docker compose up -d --build
else
    docker-compose up -d --build
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  部署完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "访问地址:"
echo "  - 前端: http://YOUR_IP"
echo "  - 后端API: http://YOUR_IP:54680"
echo "  - Emby代理: http://YOUR_IP:54682"
echo ""
echo -e "默认管理员:"
echo "  - 用户名: admin"
echo "  - 密码: admin123"
echo ""
echo -e "${YELLOW}重要: 请修改 config/config.yaml 中的 Emby 配置${NC}"
echo ""
echo "管理命令:"
echo "  docker compose logs -f        # 查看日志"
echo "  docker compose restart        # 重启服务"
echo "  docker compose down           # 停止服务"
