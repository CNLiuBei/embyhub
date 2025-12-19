# 卡密系统技术架构设计

**架构师**: Cascade AI (Technical Architect)  
**日期**: 2025-12-07  
**版本**: v1.0

---

## 🏗️ 整体架构

### 系统架构图

```
┌──────────────────────────────────────────────────────────┐
│                       用户层                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐               │
│  │ 管理员   │  │ 普通用户 │  │ 移动端   │               │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘               │
└───────┼─────────────┼─────────────┼────────────────────┘
        │             │             │
┌───────┼─────────────┼─────────────┼────────────────────┐
│       │             │             │   CDN层             │
│       └─────────────┴─────────────┘                    │
│              Nginx (反向代理)                           │
└─────────────────────┬──────────────────────────────────┘
                      │
┌─────────────────────┼──────────────────────────────────┐
│                     │         应用层                    │
│  ┌──────────────────┴──────────────────────────┐       │
│  │           前端 (React + TypeScript)         │       │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  │       │
│  │  │ 管理后台 │  │ 用户中心 │  │ 移动端   │  │       │
│  │  └──────────┘  └──────────┘  └──────────┘  │       │
│  │                                              │       │
│  │  • Ant Design / Material-UI                │       │
│  │  • React Query (数据管理)                  │       │
│  │  • Redux Toolkit (状态管理)                │       │
│  │  • ECharts (数据可视化)                    │       │
│  └──────────────────┬──────────────────────────┘       │
│                     │ REST API                          │
│  ┌──────────────────┴──────────────────────────┐       │
│  │           后端 (Go + Gin)                   │       │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  │       │
│  │  │ API网关  │  │ 业务服务 │  │ 定时任务 │  │       │
│  │  └──────────┘  └──────────┘  └──────────┘  │       │
│  │                                              │       │
│  │  • JWT认证                                  │       │
│  │  • 限流中间件                               │       │
│  │  • 日志监控                                 │       │
│  └──────────────────┬──────────────────────────┘       │
└─────────────────────┼──────────────────────────────────┘
                      │
┌─────────────────────┼──────────────────────────────────┐
│                     │         数据层                    │
│  ┌──────────────────┴──────────────────────────┐       │
│  │  PostgreSQL (主数据库)                      │       │
│  │  • 用户数据                                  │       │
│  │  • 卡密数据                                  │       │
│  │  • 订单数据                                  │       │
│  └──────────────────┬──────────────────────────┘       │
│                     │                                   │
│  ┌──────────────────┴──────────────────────────┐       │
│  │  Redis (缓存 + 会话 + 限流)                 │       │
│  │  • Session存储                               │       │
│  │  • 限流计数器                                │       │
│  │  • 热点数据缓存                              │       │
│  └──────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────┘
```

---

## 🎯 技术选型

### 前端技术栈

#### 核心框架
| 技术 | 版本 | 用途 | 优势 |
|-----|------|------|------|
| React | 18.x | UI框架 | 组件化、生态丰富 |
| TypeScript | 5.x | 类型系统 | 类型安全、代码提示 |
| Vite | 5.x | 构建工具 | 快速启动、HMR |

#### UI组件库
| 组件库 | 用途 | 选择理由 |
|--------|------|----------|
| Ant Design | 管理后台 | 企业级、组件丰富 |
| Tailwind CSS | 样式系统 | 灵活、可定制 |
| shadcn/ui | 通用组件 | 现代、可复制 |

#### 状态管理
| 工具 | 用途 | 场景 |
|-----|------|------|
| Redux Toolkit | 全局状态 | 用户信息、主题配置 |
| React Query | 服务端状态 | API数据缓存、同步 |
| Zustand | 轻量状态 | 临时状态、UI状态 |

#### 数据可视化
| 库 | 用途 | 特点 |
|----|------|------|
| ECharts | 图表展示 | 功能强大、中文文档 |
| Recharts | 简单图表 | React友好、响应式 |

### 后端技术栈

#### 核心框架
| 技术 | 版本 | 用途 | 优势 |
|-----|------|------|------|
| Go | 1.24+ | 服务端语言 | 高性能、并发优秀 |
| Gin | 1.9+ | Web框架 | 轻量、快速 |
| GORM | 1.25+ | ORM框架 | 功能完善、易用 |

#### 数据库
| 数据库 | 用途 | 特点 |
|--------|------|------|
| PostgreSQL | 主数据库 | 可靠、功能强大 |
| Redis | 缓存/会话 | 高性能、丰富数据结构 |

#### 中间件
| 工具 | 用途 | 功能 |
|-----|------|------|
| JWT | 认证授权 | 无状态、跨域友好 |
| Zap | 日志系统 | 高性能、结构化 |
| Validator | 参数验证 | 强类型、规则丰富 |

---

## 📦 模块设计

### 前端模块结构

```
frontend/
├── src/
│   ├── components/          # 通用组件
│   │   ├── Card/           # 卡片组件
│   │   ├── StatusBadge/    # 状态标签
│   │   ├── StatCard/       # 统计卡片
│   │   ├── ExportButton/   # 导出按钮
│   │   └── ...
│   │
│   ├── features/           # 功能模块
│   │   ├── card/          # 卡密模块
│   │   │   ├── components/     # 卡密相关组件
│   │   │   │   ├── CardList.tsx
│   │   │   │   ├── CardForm.tsx
│   │   │   │   ├── BatchList.tsx
│   │   │   │   └── ExportDialog.tsx
│   │   │   ├── hooks/          # 自定义钩子
│   │   │   │   ├── useCardList.ts
│   │   │   │   ├── useCardCreate.ts
│   │   │   │   └── useCardExport.ts
│   │   │   ├── services/       # API服务
│   │   │   │   └── cardApi.ts
│   │   │   ├── types/          # 类型定义
│   │   │   │   └── card.types.ts
│   │   │   └── utils/          # 工具函数
│   │   │       └── cardHelpers.ts
│   │   │
│   │   ├── redeem/        # 兑换模块
│   │   │   ├── components/
│   │   │   │   ├── RedeemForm.tsx
│   │   │   │   ├── RedeemHistory.tsx
│   │   │   │   └── RedeemSuccess.tsx
│   │   │   └── hooks/
│   │   │       └── useRedeem.ts
│   │   │
│   │   └── analytics/     # 分析模块
│   │       ├── components/
│   │       │   ├── TrendChart.tsx
│   │       │   ├── PieChart.tsx
│   │       │   └── StatsOverview.tsx
│   │       └── hooks/
│   │           └── useAnalytics.ts
│   │
│   ├── pages/             # 页面组件
│   │   ├── admin/
│   │   │   ├── CardManagement.tsx
│   │   │   ├── BatchManagement.tsx
│   │   │   └── Analytics.tsx
│   │   └── user/
│   │       ├── Redeem.tsx
│   │       └── RedeemHistory.tsx
│   │
│   ├── hooks/             # 全局钩子
│   │   ├── useAuth.ts
│   │   ├── usePermission.ts
│   │   └── useNotification.ts
│   │
│   ├── store/             # 状态管理
│   │   ├── slices/
│   │   │   ├── authSlice.ts
│   │   │   └── uiSlice.ts
│   │   └── store.ts
│   │
│   ├── services/          # API服务
│   │   ├── api.ts        # Axios实例
│   │   └── endpoints.ts   # API端点定义
│   │
│   ├── utils/             # 工具函数
│   │   ├── format.ts     # 格式化
│   │   ├── validate.ts   # 验证
│   │   └── constants.ts  # 常量
│   │
│   └── types/             # 全局类型
│       ├── api.types.ts
│       └── common.types.ts
│
└── public/                # 静态资源
```

### 后端模块结构

```
backend/
├── cmd/
│   └── server/
│       └── main.go                    # 入口文件
│
├── internal/
│   ├── handler/                       # 处理器层
│   │   ├── card_handler.go
│   │   ├── batch_handler.go
│   │   └── analytics_handler.go
│   │
│   ├── service/                       # 业务逻辑层
│   │   ├── card_service.go
│   │   ├── card_export_service.go
│   │   ├── scheduler_service.go
│   │   └── analytics_service.go
│   │
│   ├── repository/                    # 数据访问层
│   │   ├── card_repository.go
│   │   ├── batch_repository.go
│   │   └── cache_repository.go
│   │
│   ├── middleware/                    # 中间件
│   │   ├── auth.go
│   │   ├── rate_limit.go
│   │   ├── logger.go
│   │   └── cors.go
│   │
│   ├── models/                        # 数据模型
│   │   ├── card.go
│   │   ├── batch.go
│   │   └── order.go
│   │
│   ├── config/                        # 配置
│   │   └── config.go
│   │
│   └── router/                        # 路由
│       └── router.go
│
├── pkg/                               # 公共包
│   ├── response/                      # 响应封装
│   ├── utils/                         # 工具函数
│   └── validator/                     # 验证器
│
└── config/                            # 配置文件
    └── config.yaml
```

---

## 🔄 数据流设计

### 卡密生成流程

```
┌──────────┐
│ 管理员   │
│ 点击生成 │
└────┬─────┘
     │
     ▼
┌────────────────────────┐
│ 1. 前端表单验证       │
│    • 验证数量范围     │
│    • 验证类型选择     │
│    • 验证有效期      │
└────┬───────────────────┘
     │
     ▼
┌────────────────────────┐
│ 2. 调用API           │
│    POST /api/v1/admin/│
│    card/batch         │
└────┬───────────────────┘
     │
     ▼
┌────────────────────────┐
│ 3. 后端处理          │
│    • 权限验证        │
│    • 参数校验        │
│    • 限流检查        │
└────┬───────────────────┘
     │
     ▼
┌────────────────────────┐
│ 4. 业务处理          │
│    • 生成批次号      │
│    • 生成卡密码      │
│    • 批量插入数据库  │
└────┬───────────────────┘
     │
     ▼
┌────────────────────────┐
│ 5. 记录操作日志      │
│    • 管理员ID        │
│    • 操作时间        │
│    • 生成数量        │
└────┬───────────────────┘
     │
     ▼
┌────────────────────────┐
│ 6. 返回结果          │
│    • 批次信息        │
│    • 成功数量        │
│    • 批次号          │
└────┬───────────────────┘
     │
     ▼
┌────────────────────────┐
│ 7. 前端展示          │
│    • 成功提示        │
│    • 刷新列表        │
│    • 可立即导出      │
└────────────────────────┘
```

### 卡密兑换流程

```
┌──────────┐
│ 用户     │
│ 输入卡密 │
└────┬─────┘
     │
     ▼
┌────────────────────────┐
│ 1. 前端实时验证       │
│    • 格式检查         │
│    • 自动补全横杠     │
│    • 大写转换         │
└────┬───────────────────┘
     │
     ▼
┌────────────────────────┐
│ 2. 限流检查           │
│    • 检查用户兑换次数 │
│    • 检查IP频率       │
│    • 显示剩余次数     │
└────┬───────────────────┘
     │
     ▼
┌────────────────────────┐
│ 3. 调用API            │
│    POST /api/v1/card/ │
│    redeem             │
└────┬───────────────────┘
     │
     ▼
┌────────────────────────┐
│ 4. 后端验证           │
│    • 用户认证         │
│    • 限流验证         │
│    • 参数验证         │
└────┬───────────────────┘
     │
     ▼
┌────────────────────────┐
│ 5. 卡密验证           │
│    • 查询卡密         │
│    • 检查状态         │
│    • 检查过期         │
└────┬───────────────────┘
     │
     ▼
┌────────────────────────┐
│ 6. 事务处理           │
│    • 更新卡密状态     │
│    • 更新用户会员     │
│    • 创建兑换记录     │
│    • 更新批次统计     │
└────┬───────────────────┘
     │
     ▼
┌────────────────────────┐
│ 7. 发送通知           │
│    • 站内消息         │
│    • 邮件通知(可选)   │
└────┬───────────────────┘
     │
     ▼
┌────────────────────────┐
│ 8. 返回结果           │
│    • 兑换信息         │
│    • 会员到期时间     │
│    • 享受的权益       │
└────┬───────────────────┘
     │
     ▼
┌────────────────────────┐
│ 9. 前端展示           │
│    • 成功动画         │
│    • 权益说明         │
│    • 跳转提示         │
└────────────────────────┘
```

---

## 🔐 安全设计

### 认证授权

#### JWT Token结构
```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "user_id": "uuid",
    "username": "admin",
    "role": 3,
    "exp": 1234567890,
    "iat": 1234567890
  }
}
```

#### 权限级别
```go
const (
    RoleUser       = 0  // 普通用户 - 可兑换卡密
    RoleMember     = 1  // 会员用户 - 可兑换卡密
    RoleAdmin      = 2  // 管理员 - 所有操作
    RoleSuperAdmin = 3  // 超级管理员 - 所有操作 + 系统配置
)
```

#### API权限矩阵
| API端点 | 普通用户 | 会员 | 管理员 | 超级管理员 |
|---------|---------|------|--------|-----------|
| POST /card/redeem | ✅ | ✅ | ✅ | ✅ |
| GET /card/history | ✅ | ✅ | ✅ | ✅ |
| POST /admin/card/batch | ❌ | ❌ | ✅ | ✅ |
| GET /admin/card/list | ❌ | ❌ | ✅ | ✅ |
| GET /admin/card/export/* | ❌ | ❌ | ✅ | ✅ |
| DELETE /admin/user/:id | ❌ | ❌ | ❌ | ✅ |

### 数据安全

#### 敏感数据加密
```go
// 密码加密 - bcrypt
hashedPassword := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)

// 卡密生成 - 随机 + 去混淆字符
func GenerateCardCode() string {
    const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
    // 去除 0OI1 等易混淆字符
}
```

#### SQL注入防护
```go
// 使用参数化查询
db.Where("code = ?", code).First(&card)

// 禁止动态SQL拼接
// 错误：db.Raw("SELECT * FROM cards WHERE code = " + code)
```

#### XSS防护
```tsx
// React自动转义
<div>{userInput}</div>  // 自动HTML转义

// 使用 DOMPurify 处理富文本
import DOMPurify from 'dompurify';
const clean = DOMPurify.sanitize(dirty);
```

### 限流策略

#### 多层限流
```go
// 1. 全局限流 - 100请求/分钟/IP
GlobalRateLimit(100, time.Minute)

// 2. 用户限流 - 兑换 5次/小时
RedeemRateLimit(5, time.Hour)

// 3. IP限流 - 防止换号刷
IPRateLimit(5, time.Hour)
```

#### 熔断机制
```go
// 连续失败5次 -> 封禁1小时
if violations > 5 {
    blacklist.Add(ip, time.Hour)
}
```

---

## 📊 性能优化

### 前端优化

#### 代码分割
```tsx
// 路由懒加载
const CardManagement = lazy(() => import('./pages/CardManagement'));

// 组件懒加载
const ExportDialog = lazy(() => import('./components/ExportDialog'));
```

#### 缓存策略
```tsx
// React Query 缓存配置
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,     // 5分钟
      cacheTime: 10 * 60 * 1000,    // 10分钟
      refetchOnWindowFocus: false,
    },
  },
});
```

#### 虚拟滚动
```tsx
// 长列表优化 - react-window
import { FixedSizeList } from 'react-window';

<FixedSizeList
  height={600}
  itemCount={1000}
  itemSize={50}
>
  {Row}
</FixedSizeList>
```

### 后端优化

#### 数据库优化
```sql
-- 添加索引
CREATE INDEX idx_card_status ON cards(status);
CREATE INDEX idx_card_batch ON cards(batch_no);
CREATE INDEX idx_card_used_by ON cards(used_by);
CREATE INDEX idx_card_created ON cards(created_at);

-- 复合索引
CREATE INDEX idx_card_status_batch ON cards(status, batch_no);
```

#### 缓存策略
```go
// 热点数据缓存
// 1. 统计数据 - 缓存5分钟
cache.Set("card:stats", stats, 5*time.Minute)

// 2. 批次列表 - 缓存10分钟
cache.Set("batch:list", batches, 10*time.Minute)

// 3. 用户会员信息 - 缓存30分钟
cache.Set(fmt.Sprintf("user:%s:member", userID), member, 30*time.Minute)
```

#### 批量操作
```go
// 批量插入 - 每100条一批
db.CreateInBatches(cards, 100)

// 批量查询 - IN查询
db.Where("id IN ?", ids).Find(&cards)
```

#### 异步处理
```go
// 导出任务异步化
go func() {
    data := generateExport()
    // 存储到临时文件
    // 通知用户下载
}()
```

---

## 🔍 监控告警

### 日志系统

#### 日志级别
```go
// Debug - 调试信息
logger.Debug("生成卡密", zap.Int("count", 100))

// Info - 正常信息
logger.Info("用户兑换成功", zap.String("user_id", uid))

// Warn - 警告信息
logger.Warn("兑换频繁", zap.String("ip", ip))

// Error - 错误信息
logger.Error("数据库错误", zap.Error(err))
```

#### 日志格式
```json
{
  "level": "INFO",
  "time": "2025-12-07T10:20:30+08:00",
  "caller": "service/card_service.go:123",
  "msg": "卡密兑换成功",
  "user_id": "uuid",
  "card_id": 12345,
  "duration": "23ms"
}
```

### 性能监控

#### 关键指标
- **API响应时间**: P50 < 100ms, P95 < 500ms, P99 < 1s
- **数据库查询时间**: P95 < 50ms
- **缓存命中率**: > 80%
- **错误率**: < 0.1%

#### 告警规则
```yaml
alerts:
  - name: "API响应慢"
    condition: "p95_latency > 1000ms"
    duration: "5m"
    action: "发送钉钉通知"
  
  - name: "错误率高"
    condition: "error_rate > 1%"
    duration: "1m"
    action: "发送邮件 + 短信"
  
  - name: "数据库连接数高"
    condition: "db_connections > 80%"
    duration: "5m"
    action: "发送告警"
```

---

## 🚀 部署方案

### Docker部署

```yaml
# docker-compose.yml
version: '3.8'

services:
  # 后端服务
  backend:
    build: ./backend
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
  
  # 前端服务
  frontend:
    build: ./frontend
    ports:
      - "3000:80"
    depends_on:
      - backend
  
  # 数据库
  postgres:
    image: postgres:15
    environment:
      - POSTGRES_PASSWORD=secret
    volumes:
      - pgdata:/var/lib/postgresql/data
  
  # 缓存
  redis:
    image: redis:7
    volumes:
      - redisdata:/data

volumes:
  pgdata:
  redisdata:
```

### CI/CD流程

```
代码提交 → GitHub
    ↓
自动测试 (GitHub Actions)
    ↓
构建镜像 (Docker Build)
    ↓
推送镜像 (Docker Registry)
    ↓
部署到服务器 (Docker Deploy)
    ↓
健康检查
    ↓
完成 ✅
```

---

## 📚 接口文档

### API规范

#### 请求格式
```http
POST /api/v1/admin/card/batch HTTP/1.1
Host: api.example.com
Content-Type: application/json
Authorization: Bearer <token>

{
  "card_type": 1,
  "quantity": 100,
  "duration": 30
}
```

#### 响应格式
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "batch_no": "B20251207123456",
    "quantity": 100,
    "cards": [...]
  }
}
```

#### 错误响应
```json
{
  "code": 400,
  "message": "参数错误：数量不能超过1000"
}
```

### 核心API列表

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| POST | /api/v1/card/redeem | 兑换卡密 | 用户 |
| GET | /api/v1/card/history | 兑换历史 | 用户 |
| POST | /api/v1/admin/card/batch | 生成卡密 | 管理员 |
| GET | /api/v1/admin/card/list | 卡密列表 | 管理员 |
| GET | /api/v1/admin/card/export/csv | 导出CSV | 管理员 |
| GET | /api/v1/admin/card/export/excel | 导出Excel | 管理员 |
| GET | /api/v1/admin/card/stats | 统计数据 | 管理员 |

---

## 📖 总结

### 架构优势
1. **模块化设计** - 高内聚低耦合
2. **分层清晰** - 易于维护扩展
3. **性能优化** - 多层缓存、异步处理
4. **安全可靠** - 多层防护、限流熔断
5. **监控完善** - 日志、告警、追踪

### 扩展性
- 水平扩展：无状态设计，可部署多实例
- 垂直扩展：模块化设计，易于功能扩展
- 数据库扩展：读写分离、分库分表

### 最佳实践
- 遵循RESTful API设计
- 使用TypeScript类型安全
- 代码规范统一（Prettier + ESLint）
- Git工作流规范（Git Flow）
- 完善的单元测试和集成测试

---

**架构师签名**: Cascade AI  
**审核**: 待审核  
**版本**: v1.0
