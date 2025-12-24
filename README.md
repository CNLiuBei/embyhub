# EmbyHub - Emby用户管理系统

基于 React + Go + PostgreSQL + Redis 的 Emby 用户管理系统，支持会员管理、卡密系统、支付宝支付、闲管家对接等功能。

## 功能特性

- 🎬 **Emby 集成** - 自动同步用户到 Emby，支持设备管理、会话控制
- 👥 **用户管理** - 注册登录、个人中心、邀请好友
- 💳 **会员系统** - 卡密兑换、支付宝购买、积分兑换
- 🎫 **卡密管理** - 批量生成、导出、外部API对接
- 🛒 **闲管家对接** - 虚拟货源自动发货
- 💬 **社区功能** - 论坛、私信、关注
- 🔒 **安全特性** - JWT认证、接口限流、IP黑名单

## GitHub Actions 自动构建

推送代码到 GitHub 后会自动触发构建，需要在仓库 Settings → Secrets 中配置：

| Secret 名称 | 说明 |
|------------|------|
| ALIYUN_REGISTRY_USERNAME | 阿里云镜像仓库用户名 |
| ALIYUN_REGISTRY_PASSWORD | 阿里云镜像仓库密码 |

构建产物：
- 多平台二进制文件 (linux/darwin/windows, amd64/arm64)
- Docker 镜像推送到 GHCR 和阿里云镜像仓库

## 快速部署 (Docker)

### 1. 创建目录

```bash
mkdir -p embyhub && cd embyhub
mkdir -p config logs uploads bin data/postgres data/redis
```

### 2. 下载配置文件

```bash
# 下载 docker-compose.yml
curl -O https://raw.githubusercontent.com/CNLiuBei/embyhub/main/docker/docker-compose.simple.yml
mv docker-compose.simple.yml docker-compose.yml

# 下载示例配置
curl -o config/config.yaml https://raw.githubusercontent.com/CNLiuBei/embyhub/main/backend/config/config.example.yaml
```

### 3. 修改配置

编辑 `config/config.yaml`，主要修改：

```yaml
# JWT密钥（必须修改为随机字符串）
jwt:
  secret: "your-random-secret-key-change-this"

# Emby配置（修改为你的Emby服务器）
emby:
  enabled: true
  base_url: "http://YOUR_EMBY_IP:8096"
  api_key: "YOUR_EMBY_API_KEY"
```

### 4. 启动服务

```bash
docker compose up -d
```

### 5. 访问系统

- 前端：http://localhost:54681
- 后端：http://localhost:54680
- 默认管理员：`admin` / `Admin123!`

## 升级方法

```bash
cd embyhub

# 拉取最新镜像
docker compose pull

# 重启服务（数据会保留）
docker compose down
docker compose up -d
```

## 常用命令

```bash
# 查看日志
docker compose logs -f backend
docker compose logs -f frontend

# 重启服务
docker compose restart

# 停止服务
docker compose down

# 查看状态
docker compose ps
```

## 数据备份

```bash
# 备份数据库
docker exec embyhub-postgres pg_dump -U embyhub embyhub > backup.sql

# 恢复数据库
cat backup.sql | docker exec -i embyhub-postgres psql -U embyhub embyhub
```

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go 1.21+, Gin, GORM |
| 前端 | React 18, TypeScript, Ant Design |
| 数据库 | PostgreSQL 15 |
| 缓存 | Redis 7 |
| 部署 | Docker, GitHub Actions |

## 项目结构

```
embyhub/
├── backend/                # Go后端
│   ├── cmd/server/        # 入口
│   ├── config/            # 配置
│   └── internal/          # 业务代码
├── frontend/              # React前端
│   └── src/
├── docker/                # Docker配置
└── docs/                  # 文档
```

## API文档

### 公开接口
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/user/login | 用户登录 |
| POST | /api/v1/user/register | 用户注册 |
| POST | /api/v1/card/renew | 卡密续费 |

### 用户接口 (需认证)
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/user/info | 获取用户信息 |
| POST | /api/v1/card/redeem | 兑换卡密 |
| GET | /api/v1/member/info | 会员信息 |

### 管理接口 (需管理员)
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/admin/user/list | 用户列表 |
| POST | /api/v1/admin/card/batch | 批量生成卡密 |
| GET | /api/v1/admin/stat/user | 用户统计 |

## 外部API对接

### 闲管家虚拟货源
系统支持闲管家虚拟货源对接，可在管理后台配置商户信息和商品映射。

### 外部卡密API
提供外部卡密获取接口，支持第三方系统调用：
```
GET /api/external/card/fetch?api_key=xxx&type=month
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| DB_HOST | 数据库地址 | postgres |
| DB_USER | 数据库用户 | embyhub |
| DB_PASSWORD | 数据库密码 | embyhub123 |
| REDIS_HOST | Redis地址 | redis |

## License

MIT
