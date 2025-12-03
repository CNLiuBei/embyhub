# 部署指南

本文档详细介绍Emby用户管理系统的部署步骤。

## 系统要求

### 硬件要求
- CPU: 2核心及以上
- 内存: 4GB及以上
- 磁盘: 20GB及以上可用空间

### 软件要求
- Docker 20.10+
- Docker Compose 2.0+
- （可选）Node.js 18+, Go 1.21+（用于本地开发）

## Docker部署（生产环境推荐）

### 1. 准备工作

```bash
# 克隆项目
git clone <repository-url>
cd embyhub

# 创建环境变量文件
cp .env.example .env
```

### 2. 配置环境变量

编辑 `.env` 文件：

```bash
# 数据库配置
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_secure_password  # 修改为强密码
DB_NAME=embyhub

# Redis配置
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password  # 修改为强密码

# JWT配置
JWT_SECRET=your_jwt_secret_key  # 修改为随机字符串

# Emby配置
EMBY_SERVER_URL=http://your-emby-server:8096
EMBY_API_KEY=your_emby_api_key
```

### 3. 启动服务

```bash
# 构建并启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 4. 访问系统

- 前端界面: http://your-server-ip:3000
- 后端API: http://your-server-ip:8080
- 默认账号: admin / 123456

### 5. 初始化配置

首次部署后：
1. 使用默认账号登录
2. 立即修改admin密码
3. 配置Emby服务器连接
4. 创建其他管理员账号

## 本地开发部署

### 1. 启动依赖服务

```bash
# 只启动PostgreSQL和Redis
docker-compose up -d postgres redis
```

### 2. 初始化数据库

```bash
# 进入数据库容器
docker-compose exec postgres psql -U postgres

# 创建数据库
CREATE DATABASE embyhub;
\q

# 执行初始化脚本
docker-compose exec -T postgres psql -U postgres -d embyhub < database/init_schema.sql
docker-compose exec -T postgres psql -U postgres -d embyhub < database/init_data.sql
```

### 3. 启动后端

```bash
cd backend

# 下载依赖
go mod download

# 运行
go run cmd/main.go
```

### 4. 启动前端

```bash
cd frontend

# 安装依赖
npm install

# 开发模式
npm run dev
```

## 生产环境优化

### 1. 使用Nginx反向代理

创建 `nginx.conf`:

```nginx
upstream backend {
    server localhost:8080;
}

upstream frontend {
    server localhost:3000;
}

server {
    listen 80;
    server_name your-domain.com;

    # 前端
    location / {
        proxy_pass http://frontend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # API
    location /api {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 2. 启用HTTPS

```bash
# 使用Let's Encrypt
sudo certbot --nginx -d your-domain.com
```

### 3. 数据备份

```bash
# 备份PostgreSQL
docker-compose exec postgres pg_dump -U postgres embyhub > backup_$(date +%Y%m%d).sql

# 备份Redis
docker-compose exec redis redis-cli SAVE
docker cp embyhub-redis:/data/dump.rdb redis_backup_$(date +%Y%m%d).rdb
```

### 4. 监控和日志

```bash
# 查看服务资源使用
docker stats

# 持续查看日志
docker-compose logs -f backend

# 日志轮转（在docker-compose.yml中配置）
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

## 升级部署

### 1. 备份数据

```bash
# 备份数据库
docker-compose exec postgres pg_dump -U postgres embyhub > backup_before_upgrade.sql

# 备份配置
cp backend/config/config.yaml config_backup.yaml
```

### 2. 拉取新代码

```bash
git pull origin main
```

### 3. 重新构建

```bash
# 停止服务
docker-compose down

# 重新构建镜像
docker-compose build

# 启动服务
docker-compose up -d
```

### 4. 验证升级

```bash
# 检查服务状态
docker-compose ps

# 检查日志
docker-compose logs backend frontend
```

## 故障排查

### 服务无法启动

```bash
# 查看详细日志
docker-compose logs backend
docker-compose logs frontend

# 检查端口占用
netstat -tulpn | grep -E '3000|8080|5432|6379'

# 重新创建容器
docker-compose down -v
docker-compose up -d
```

### 数据库连接失败

```bash
# 检查数据库状态
docker-compose ps postgres

# 进入数据库容器
docker-compose exec postgres psql -U postgres

# 检查数据库配置
docker-compose exec postgres cat /var/lib/postgresql/data/postgresql.conf
```

### 性能问题

```bash
# 检查资源使用
docker stats

# 优化PostgreSQL配置
# 编辑postgresql.conf
shared_buffers = 256MB
effective_cache_size = 1GB
maintenance_work_mem = 64MB
```

## 安全建议

1. **修改默认密码**
   - 数据库密码
   - Redis密码
   - Admin账号密码

2. **限制网络访问**
   - 使用防火墙限制端口访问
   - 只开放必要的端口（80, 443）

3. **定期备份**
   - 每日自动备份数据库
   - 保留至少7天的备份

4. **更新依赖**
   - 定期更新Docker镜像
   - 及时修复安全漏洞

5. **监控日志**
   - 定期检查访问日志
   - 设置异常告警

## 容量规划

### 用户规模与资源配置

| 用户数 | CPU | 内存 | 磁盘 |
|--------|-----|------|------|
| <1000  | 2核 | 4GB  | 20GB |
| 1000-5000 | 4核 | 8GB  | 50GB |
| 5000-10000 | 8核 | 16GB | 100GB |

### 数据库优化

```sql
-- 创建额外索引（大量数据时）
CREATE INDEX CONCURRENTLY idx_access_records_user_time_resource 
ON access_records(user_id, access_time, resource);

-- 清理旧数据
DELETE FROM access_records WHERE access_time < NOW() - INTERVAL '90 days';
VACUUM ANALYZE access_records;
```

## 维护计划

### 日常维护
- 检查服务状态
- 查看错误日志
- 监控磁盘空间

### 每周维护
- 数据库备份验证
- 清理旧日志
- 检查安全更新

### 每月维护
- 性能分析
- 容量评估
- 清理旧访问记录

## 支持

如遇到部署问题，请：
1. 查看日志文件
2. 参考故障排查章节
3. 提交Issue并附上详细日志
