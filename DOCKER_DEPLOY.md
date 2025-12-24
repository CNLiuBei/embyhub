# EmbyHub Docker 部署指南

## 快速开始

### 方式一：使用 docker-compose（推荐）

1. 创建部署目录并下载配置文件：

```bash
mkdir -p embyhub && cd embyhub

# 下载 docker-compose.yml
curl -O https://raw.githubusercontent.com/CNLiuBei/embyhub/main/docker-compose.yml

# 创建配置目录
mkdir -p backend/config backend/logs backend/uploads backend/bin
```

2. 创建配置文件 `backend/config/config.yaml`：

```yaml
server:
  port: 8080
  mode: release

database:
  host: postgres
  port: 5432
  user: embyhub
  password: embyhub123
  dbname: embyhub
  sslmode: disable

redis:
  host: redis
  port: 6379
  password: ""
  db: 0

jwt:
  secret: your-jwt-secret-change-this
  access_expire: 24h
  refresh_expire: 168h
  issuer: embyhub

security:
  rate_limit:
    login: 5
    api: 60

emby:
  server_url: http://your-emby-server:8096
  api_key: your-emby-api-key
```

3. 创建环境变量文件 `.env`（可选）：

```bash
# 数据库配置
DB_USER=embyhub
DB_PASSWORD=embyhub123
DB_NAME=embyhub

# JWT密钥（请修改为随机字符串）
JWT_SECRET=your-super-secret-key-change-this
```

4. 启动服务：

```bash
docker-compose up -d
```

5. 访问系统：
   - 前端：http://localhost:54681
   - 后端API：http://localhost:54680/api/v1
   - 默认管理员：admin / Admin123!

### 方式二：手动拉取镜像

```bash
# 拉取镜像
docker pull ghcr.io/cnliubei/embyhub-backend:latest
docker pull ghcr.io/cnliubei/embyhub-frontend:latest

# 运行后端
docker run -d \
  --name embyhub-backend \
  -p 54680:8080 \
  -v ./backend/config:/app/config \
  -v ./backend/logs:/app/logs \
  -v ./backend/uploads:/app/uploads \
  ghcr.io/cnliubei/embyhub-backend:latest

# 运行前端
docker run -d \
  --name embyhub-frontend \
  -p 54681:80 \
  ghcr.io/cnliubei/embyhub-frontend:latest
```

## 升级

```bash
# 拉取最新镜像
docker-compose pull

# 重启服务
docker-compose up -d
```

## 数据备份

```bash
# 备份数据库
docker exec embyhub-postgres pg_dump -U embyhub embyhub > backup.sql

# 备份配置和上传文件
tar -czvf embyhub-data.tar.gz backend/config backend/uploads
```

## 数据恢复

```bash
# 恢复数据库
cat backup.sql | docker exec -i embyhub-postgres psql -U embyhub embyhub

# 恢复配置和上传文件
tar -xzvf embyhub-data.tar.gz
```

## 常用命令

```bash
# 查看日志
docker-compose logs -f backend
docker-compose logs -f frontend

# 重启服务
docker-compose restart

# 停止服务
docker-compose down

# 停止并删除数据卷（慎用）
docker-compose down -v
```

## 反向代理配置（Nginx）

```nginx
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    # 前端
    location / {
        proxy_pass http://127.0.0.1:54681;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # 后端API
    location /api/ {
        proxy_pass http://127.0.0.1:54680;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Emby代理
    location /emby/ {
        proxy_pass http://127.0.0.1:54680;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_buffering off;
    }
}
```

## 常见问题

### 1. 数据库连接失败
确保 PostgreSQL 容器已启动并健康：
```bash
docker-compose ps
docker-compose logs postgres
```

### 2. 前端无法访问后端
检查后端是否正常运行：
```bash
curl http://localhost:54680/health
```

### 3. 镜像拉取慢
可以配置 Docker 镜像加速器，或使用代理。

## 支持的架构

- linux/amd64
- linux/arm64
