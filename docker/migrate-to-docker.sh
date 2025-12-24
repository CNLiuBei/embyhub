#!/bin/bash
# EmbyHub - 从二进制部署迁移到Docker镜像部署
# 保留现有数据

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

WORK_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$WORK_DIR"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  EmbyHub - 迁移到Docker镜像部署${NC}"
echo -e "${GREEN}========================================${NC}"

# 1. 备份当前配置
echo -e "${YELLOW}[1/6] 备份当前配置...${NC}"
BACKUP_NAME="backup-$(date +%Y%m%d%H%M%S)"
mkdir -p "$BACKUP_NAME"
cp -r config "$BACKUP_NAME/" 2>/dev/null || true
cp -r uploads "$BACKUP_NAME/" 2>/dev/null || true
cp docker-compose.yml "$BACKUP_NAME/" 2>/dev/null || true
echo "备份已保存到: $BACKUP_NAME"

# 2. 停止当前服务
echo -e "${YELLOW}[2/6] 停止当前服务...${NC}"
docker compose down 2>/dev/null || docker-compose down 2>/dev/null || true

# 3. 更新配置文件（修改端口为8080）
echo -e "${YELLOW}[3/6] 更新配置文件...${NC}"
if [ -f config/config.yaml ]; then
    # 备份原配置
    cp config/config.yaml config/config.yaml.bak
    # 修改端口为8080（Docker内部端口）
    sed -i 's/port: 54680/port: 8080/' config/config.yaml 2>/dev/null || \
    sed -i '' 's/port: 54680/port: 8080/' config/config.yaml
    echo "配置文件已更新"
fi

# 4. 下载新的docker-compose.yml
echo -e "${YELLOW}[4/6] 下载新的docker-compose配置...${NC}"
cat > docker-compose.yml << 'EOF'
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: feiniu-postgres
    restart: always
    environment:
      POSTGRES_USER: fnuser
      POSTGRES_PASSWORD: fnuser123
      POSTGRES_DB: feiniu_user
    volumes:
      - ./data/postgres:/var/lib/postgresql/data
    ports:
      - "15432:5432"
    networks:
      - feiniu-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U fnuser -d feiniu_user"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: feiniu-redis
    restart: always
    command: redis-server --appendonly yes
    volumes:
      - ./data/redis:/data
    ports:
      - "16379:6379"
    networks:
      - feiniu-network

  backend:
    image: ghcr.io/cnliubei/embyhub-backend:latest
    container_name: feiniu-backend
    restart: always
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_started
    volumes:
      - ./config:/app/config
      - ./logs:/app/logs
      - ./uploads:/app/uploads
      - ./bin:/app/bin
    ports:
      - "54680:8080"
    networks:
      - feiniu-network

  frontend:
    image: ghcr.io/cnliubei/embyhub-frontend:latest
    container_name: feiniu-frontend
    restart: always
    depends_on:
      - backend
    ports:
      - "54681:80"
    networks:
      - feiniu-network

networks:
  feiniu-network:
    driver: bridge
EOF

# 5. 拉取最新镜像
echo -e "${YELLOW}[5/6] 拉取最新Docker镜像...${NC}"
docker compose pull 2>/dev/null || docker-compose pull

# 6. 启动服务
echo -e "${YELLOW}[6/6] 启动服务...${NC}"
docker compose up -d 2>/dev/null || docker-compose up -d

# 等待服务启动
echo "等待服务启动..."
sleep 10

# 检查服务状态
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  迁移完成！${NC}"
echo -e "${GREEN}========================================${NC}"
docker compose ps 2>/dev/null || docker-compose ps

echo ""
echo -e "${GREEN}访问地址:${NC}"
echo "  前端: http://localhost:54681"
echo "  后端: http://localhost:54680"
echo ""
echo -e "${YELLOW}如果遇到问题，可以恢复备份:${NC}"
echo "  cp $BACKUP_NAME/docker-compose.yml ."
echo "  cp $BACKUP_NAME/config/config.yaml config/"
