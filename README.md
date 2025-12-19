# Emby用户管理系统

基于 React + Go + PostgreSQL + Redis 技术栈开发的Emby用户管理系统。

## 技术栈

### 后端
- Go 1.21+
- Gin Web框架
- GORM ORM
- PostgreSQL 15+
- Redis 7+
- JWT认证
- Zap日志

### 前端
- React 18
- TypeScript
- Redux Toolkit
- React Query
- Ant Design v5
- Tailwind CSS
- Vite

## 功能模块

### 用户端
- 用户登录/注册(支持账号或邮箱)
- 个人信息管理
- 密码修改
- 会员中心(卡密兑换)
- 观影记录
- 收藏管理

### 管理后台
- 数据统计概览
- 用户列表(分页/筛选/批量操作)
- 卡密管理(批量生成/查询/禁用/导出)
- 操作日志审计

### 会员体系
- **普通用户**: 基础功能
- **月卡用户**: 30天会员权益
- **年卡用户**: 365天会员权益

## 快速开始

### 1. 环境要求
- Go 1.21+
- Node.js 18+
- PostgreSQL 15+
- Redis 7+

### 2. 配置数据库

```sql
-- 创建数据库和用户
CREATE DATABASE emby_user;
CREATE USER embyuser WITH PASSWORD 'embyuser123';
GRANT ALL PRIVILEGES ON DATABASE emby_user TO embyuser;
```

修改 `backend/config/config.yaml` 中的数据库连接配置。

### 3. 初始化项目

```bash
cd emby-user-system
chmod +x start.sh
./start.sh init
```

### 4. 启动服务

```bash
./start.sh start
```

### 5. 访问系统

- 前端: http://localhost:3000
- 后端API: http://localhost:8080/api/v1
- 健康检查: http://localhost:8080/health

## 项目结构

```
emby-user-system/
├── backend/                    # Go后端
│   ├── cmd/server/            # 入口
│   ├── config/                # 配置文件
│   ├── internal/
│   │   ├── config/            # 配置解析
│   │   ├── database/          # 数据库连接
│   │   ├── handler/           # HTTP处理器
│   │   ├── middleware/        # 中间件
│   │   ├── models/            # 数据模型
│   │   ├── router/            # 路由
│   │   └── service/           # 业务逻辑
│   └── pkg/
│       ├── auth/              # JWT认证
│       ├── response/          # 响应格式
│       └── utils/             # 工具函数
├── frontend/                   # React前端
│   ├── src/
│   │   ├── layouts/           # 布局组件
│   │   ├── pages/             # 页面
│   │   ├── services/          # API服务
│   │   └── store/             # Redux状态
│   ├── package.json
│   └── vite.config.ts
├── start.sh                    # 启动脚本
└── README.md
```

## API接口

### 公开接口
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/user/login | 用户登录(账号/邮箱) |
| POST | /api/v1/user/register | 用户注册 |
| POST | /api/v1/user/refresh-token | 刷新Token |

### 用户接口 (需认证)
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/user/info | 获取用户信息 |
| PUT | /api/v1/user/update | 更新用户信息 |
| PUT | /api/v1/user/password | 修改密码 |
| POST | /api/v1/user/logout | 退出登录 |

### 卡密接口 (需认证)
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/card/redeem | 兑换卡密 |
| GET | /api/v1/card/history | 兑换记录 |

### 管理员接口 (需管理员权限)
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/admin/user/list | 用户列表 |
| GET | /api/v1/admin/user/:id | 用户详情 |
| PUT | /api/v1/admin/user/:id/status | 修改状态 |
| POST | /api/v1/admin/card/batch | 批量生成卡密 |
| GET | /api/v1/admin/card/list | 卡密列表 |
| GET | /api/v1/admin/card/export | 导出卡密 |
| GET | /api/v1/admin/stat/user | 用户统计 |

## 安全特性

- 密码 bcrypt 加盐哈希存储
- JWT Token 认证
- 接口限流 (登录5次/分钟，API 60次/分钟)

## 默认账户

执行 `backend/scripts/init.sql` 初始化后：

| 角色 | 账号 | 邮箱 | 密码 |
|------|------|------|------|
| 管理员 | admin | admin@emby.local | Admin123! |
| 普通用户 | testuser | user@emby.local | User123! |

## License

MIT
