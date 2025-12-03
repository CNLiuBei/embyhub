# 快速启动指南

## 方式一：使用Docker Compose（推荐）

### 1. 一键启动所有服务

```bash
cd /vol1/1000/embyhub
docker-compose up -d
```

### 2. 查看启动状态

```bash
# 查看所有容器状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 3. 访问系统

- **前端界面**: http://localhost:3000
- **后端API**: http://localhost:8080
- **默认账号**: admin / 123456

## 方式二：本地开发运行

### 前置要求
- Go 1.21+
- Node.js 18+
- PostgreSQL 14+
- Redis 7.0+

### 步骤1：启动数据库服务

```bash
# 仅启动PostgreSQL和Redis
docker-compose up -d postgres redis

# 等待服务就绪
sleep 5

# 初始化数据库
docker-compose exec -T postgres psql -U postgres -c "CREATE DATABASE embyhub;"
docker-compose exec -T postgres psql -U postgres -d embyhub < database/init_schema.sql
docker-compose exec -T postgres psql -U postgres -d embyhub < database/init_data.sql
```

### 步骤2：启动后端

```bash
cd backend

# 下载依赖
go mod download

# 运行后端
go run cmd/main.go
```

后端将在 http://localhost:8080 启动

### 步骤3：启动前端

```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

前端将在 http://localhost:3000 启动

## 初次使用配置

### 1. 登录系统

使用默认账号登录：
- 用户名: `admin`
- 密码: `123456`

### 2. 修改密码

登录后立即修改默认密码！

### 3. 配置Emby连接

进入"系统设置"页面，配置：
- Emby服务器地址（如: http://192.168.1.100:8096）
- Emby API密钥
- 同步周期（默认3600秒）

### 4. 同步Emby用户

进入"Emby同步"页面，点击"立即同步"

## 常用命令

### Docker管理

```bash
# 启动所有服务
docker-compose up -d

# 停止所有服务
docker-compose down

# 重启服务
docker-compose restart

# 查看日志
docker-compose logs -f backend
docker-compose logs -f frontend

# 进入容器
docker-compose exec backend sh
docker-compose exec postgres psql -U postgres -d embyhub
```

### 数据库操作

```bash
# 备份数据库
docker-compose exec postgres pg_dump -U postgres embyhub > backup.sql

# 恢复数据库
docker-compose exec -T postgres psql -U postgres -d embyhub < backup.sql

# 连接数据库
docker-compose exec postgres psql -U postgres -d embyhub
```

### 后端开发

```bash
cd backend

# 运行
go run cmd/main.go

# 编译
go build -o embyhub cmd/main.go

# 测试
go test ./...

# 格式化代码
go fmt ./...
```

### 前端开发

```bash
cd frontend

# 开发模式
npm run dev

# 构建生产版本
npm run build

# 预览生产版本
npm run preview

# 代码检查
npm run lint
```

## 故障排查

### 端口被占用

```bash
# 检查端口占用
sudo lsof -i :3000
sudo lsof -i :8080
sudo lsof -i :5432
sudo lsof -i :6379

# 修改docker-compose.yml中的端口映射
```

### 数据库连接失败

```bash
# 检查PostgreSQL状态
docker-compose ps postgres
docker-compose logs postgres

# 测试连接
docker-compose exec postgres psql -U postgres -c "SELECT version();"
```

### Redis连接失败

```bash
# 检查Redis状态
docker-compose ps redis
docker-compose logs redis

# 测试连接
docker-compose exec redis redis-cli ping
```

### 后端无法启动

```bash
# 查看详细错误
docker-compose logs backend

# 检查配置文件
cat backend/config/config.yaml

# 重新下载依赖
cd backend && go mod tidy && go mod download
```

### 前端无法启动

```bash
# 清除node_modules重新安装
cd frontend
rm -rf node_modules package-lock.json
npm install

# 检查Node版本
node --version  # 需要 18+
```

## 开发技巧

### 热重载

后端使用 `air` 实现热重载：

```bash
cd backend
go install github.com/cosmtrek/air@latest
air
```

前端Vite自带热重载，修改代码后自动刷新。

### API测试

使用curl测试API：

```bash
# 登录获取Token
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"123456"}'

# 使用Token访问API
curl -X GET http://localhost:8080/api/users \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 数据库迁移

添加新表或修改表结构：

1. 修改 `database/init_schema.sql`
2. 创建迁移SQL文件
3. 执行迁移

```bash
docker-compose exec -T postgres psql -U postgres -d embyhub < database/migration_xxx.sql
```

## 下一步

1. ✅ 系统已启动并运行
2. 📝 阅读 [README.md](README.md) 了解功能特性
3. 📖 查看 [DEPLOYMENT.md](DEPLOYMENT.md) 了解生产部署
4. 🎨 根据需求自定义界面和功能
5. 🔒 配置生产环境安全设置

## 获取帮助

遇到问题？
1. 查看日志文件
2. 阅读故障排查章节
3. 提交Issue并附上详细信息
