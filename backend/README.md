# Emby用户管理系统 - 后端

基于 Go + Gin + PostgreSQL + Redis 的 Emby 用户管理系统后端服务。

## 技术栈

- **框架**: Gin 1.9+
- **语言**: Go 1.21+
- **ORM**: GORM 2.x
- **数据库**: PostgreSQL 14+
- **缓存**: Redis 7.0+
- **认证**: JWT (golang-jwt/jwt v5)
- **日志**: Zap

## 项目结构

```
backend/
├── cmd/                    # 程序入口
│   └── main.go
├── config/                 # 配置文件
│   ├── config.go          # 配置结构定义
│   └── config.yaml        # 配置文件
├── internal/               # 内部代码
│   ├── dao/               # 数据访问层
│   ├── handler/           # 控制器层
│   ├── middleware/        # 中间件
│   ├── model/             # 数据模型
│   ├── router/            # 路由
│   ├── service/           # 服务层
│   └── util/              # 工具函数
├── pkg/                    # 公共包
│   ├── database/          # 数据库连接
│   ├── emby/              # Emby客户端
│   └── redis/             # Redis连接
├── go.mod                  # 依赖管理
├── go.sum
├── Dockerfile             # Docker构建文件
└── Makefile              # Make命令
```

## 快速开始

### 1. 安装依赖

```bash
cd backend
go mod download
```

### 2. 配置数据库

修改 `config/config.yaml` 中的数据库配置：

```yaml
database:
  host: localhost
  port: 5432
  user: embyhub
  password: embyhub123
  dbname: embyhub
```

### 3. 初始化数据库

```bash
# 执行数据库初始化脚本
psql -U embyhub -d embyhub -f ../database/init_schema.sql
psql -U embyhub -d embyhub -f ../database/init_data.sql
```

### 4. 启动服务

```bash
# 方式1：直接运行
go run cmd/main.go

# 方式2：编译后运行
make build
./embyhub
```

服务将在 `http://localhost:8080` 启动

## API文档

### 认证相关

- `POST /api/auth/login` - 管理员登录
- `POST /api/auth/logout` - 管理员登出
- `GET /api/auth/current` - 获取当前用户信息

### 用户管理

- `GET /api/users` - 获取用户列表
- `POST /api/users` - 创建用户
- `GET /api/users/:id` - 获取用户详情
- `PUT /api/users/:id` - 更新用户
- `DELETE /api/users/:id` - 删除用户
- `PUT /api/users/:id/password` - 重置用户密码

### 角色管理

- `GET /api/roles` - 获取角色列表
- `POST /api/roles` - 创建角色
- `GET /api/roles/:id` - 获取角色详情
- `PUT /api/roles/:id` - 更新角色
- `DELETE /api/roles/:id` - 删除角色
- `POST /api/roles/:id/permissions` - 为角色分配权限

### 权限管理

- `GET /api/permissions` - 获取权限列表

### 访问记录

- `GET /api/access-records` - 获取访问记录列表
- `POST /api/access-records` - 创建访问记录

### 统计数据

- `GET /api/statistics` - 获取统计数据

### 系统配置

- `GET /api/configs` - 获取系统配置列表
- `PUT /api/configs/:key` - 更新系统配置

### Emby同步

- `POST /api/emby/test` - 测试Emby连接
- `POST /api/emby/sync` - 同步Emby用户
- `GET /api/emby/users` - 获取Emby用户列表

## 默认账号

- 用户名: `admin`
- 密码: `Liubei00`

## 开发命令

```bash
# 下载依赖
make deps

# 编译
make build

# 运行
make run

# 清理
make clean

# 测试
make test

# 代码格式化
make fmt

# Docker构建
make docker-build
```

## 环境变量

可通过环境变量覆盖配置文件：

- `DB_HOST` - 数据库地址
- `DB_PORT` - 数据库端口
- `DB_USER` - 数据库用户名
- `DB_PASSWORD` - 数据库密码
- `DB_NAME` - 数据库名称
- `REDIS_HOST` - Redis地址
- `REDIS_PORT` - Redis端口
- `REDIS_PASSWORD` - Redis密码
- `JWT_SECRET` - JWT密钥
- `EMBY_SERVER_URL` - Emby服务器地址
- `EMBY_API_KEY` - Emby API密钥

## License

MIT
