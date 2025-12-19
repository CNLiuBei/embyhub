# Vite域名控制配置说明

## 架构决策

已改回由**Vite前端**控制域名访问，后端不再验证域名。

### 为什么改回Vite控制？

1. **配置简单**：只需修改一个配置文件
2. **立即生效**：重启前端服务即可
3. **开发友好**：不需要在后端和前端之间同步配置
4. **适合开发环境**：Vite本身就有域名控制功能

## 当前配置

### Vite配置（vite.config.ts）

```typescript
server: {
  port: 3000,
  host: '0.0.0.0',
  allowedHosts: [
    'localhost',
    'fnhub.trueliu.com',
    '.trueliu.com',  // 允许所有 trueliu.com 的子域名
  ],
  proxy: {
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: true,
    },
  },
}
```

### 后端配置

域名白名单中间件已**禁用**：
```go
// 域名白名单验证已移至前端Vite层控制
// 如需后端验证，取消注释下面这行：
// r.Use(middleware.DomainWhitelist(db))
```

## 工作原理

```
浏览器访问
    ↓
Vite Dev Server
    ↓
检查 Host 头
    ↓
在 allowedHosts 列表中？
    ├─ 是 → 允许访问 ✅
    └─ 否 → 拒绝访问 ❌
         返回: Blocked request
```

## 修改域名白名单

### 方法1：修改vite.config.ts（推荐）

1. 打开 `frontend/vite.config.ts`
2. 修改 `allowedHosts` 数组：
```typescript
allowedHosts: [
  'localhost',
  'your-domain.com',      // 添加新域名
  '*.your-domain.com',    // 通配符支持
],
```
3. 重启前端服务：
```bash
./stop-all.sh
./start-all.sh
```

### 方法2：通过UI记录（仅供参考）

系统设置页面的"域名访问控制"仍然可以使用：
- ✅ 保存域名配置到数据库
- ✅ 查看和管理域名列表
- ❌ **不会**实际控制访问（需手动同步到vite.config.ts）

**用途**：作为配置记录和参考，方便团队成员查看允许的域名

## 支持的域名格式

### 完整域名
```typescript
'localhost'
'fnhub.trueliu.com'
'api.example.com'
```

### 通配符域名
```typescript
'.trueliu.com'      // 匹配所有 *.trueliu.com
'.example.com'      // 匹配所有 *.example.com
```

**注意**：
- `.trueliu.com` 会匹配 `sub.trueliu.com`、`api.trueliu.com` 等
- 但**不匹配** `trueliu.com` 本身（需要单独添加）

## 测试验证

### 测试1：允许的域名（应该成功）

访问：`https://fnhub.trueliu.com`
- ✅ 页面正常加载
- ✅ API请求正常工作

### 测试2：未授权的域名（应该失败）

访问：`https://unauthorized-domain.com`
- ❌ Vite返回：
```
Blocked request. This host ("unauthorized-domain.com") is not allowed.
To allow this host, add "unauthorized-domain.com" to `server.allowedHosts` in vite.config.js.
```

## 开发环境 vs 生产环境

### 开发环境（当前配置）

**Vite Dev Server**控制域名：
- 配置文件：`vite.config.ts`
- 控制范围：前端页面 + API代理请求
- 重启方式：`./start-all.sh`

### 生产环境（建议）

**Nginx/Caddy反向代理**控制域名：

#### Nginx示例
```nginx
server {
    listen 80;
    server_name fnhub.trueliu.com;  # 只允许这个域名
    
    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
    }
}

# 其他域名会被Nginx直接拒绝
```

#### Caddy示例
```caddy
fnhub.trueliu.com {
    reverse_proxy localhost:3000
}
```

**优势**：
- 性能更好
- 更安全（在网络层阻止）
- 支持HTTPS
- 更灵活的访问控制

## 常见场景配置

### 场景1：本地开发
```typescript
allowedHosts: [
  'localhost',
  '127.0.0.1',
]
```

### 场景2：内网测试
```typescript
allowedHosts: [
  'localhost',
  '192.168.1.100',
  'dev.internal.com',
]
```

### 场景3：多域名生产环境
```typescript
allowedHosts: [
  'localhost',           // 本地调试
  'fnhub.trueliu.com',  // 主站
  '.trueliu.com',       // 所有子域名
]
```

## 如何添加新域名

### 步骤1：确定域名
假设要添加：`new.example.com`

### 步骤2：修改配置
编辑 `frontend/vite.config.ts`：
```typescript
allowedHosts: [
  'localhost',
  'fnhub.trueliu.com',
  '.trueliu.com',
  'new.example.com',  // ← 添加这行
],
```

### 步骤3：重启服务
```bash
./stop-all.sh
./start-all.sh
```

### 步骤4：测试
访问 `https://new.example.com`，应该能正常访问

## UI配置页面说明

访问：**系统设置** → **域名访问控制**

### 功能说明

#### ✅ 可以做的：
- 查看允许的域名列表
- 添加/删除域名记录
- 保存到数据库作为配置记录

#### ❌ 不能做的：
- **实时控制访问**（需要手动同步到vite.config.ts）
- 自动重启前端服务

### 建议的工作流程

1. 在UI中添加域名（记录配置）
2. 点击"保存设置"
3. 手动更新 `vite.config.ts`
4. 重启前端服务：`./start-all.sh`

## 回滚到后端控制

如果将来需要改回后端控制：

### 步骤1：启用后端中间件
编辑 `backend/internal/router/router.go`：
```go
// 取消注释这行
r.Use(middleware.DomainWhitelist(db))
```

### 步骤2：修改Vite配置
编辑 `frontend/vite.config.ts`：
```typescript
allowedHosts: true,  // 允许所有域名
proxy: {
  '/api': {
    target: 'http://localhost:8080',
    changeOrigin: false,  // 保持原始Host头
    configure: (proxy, options) => {
      proxy.on('proxyReq', (proxyReq, req, res) => {
        if (req.headers.host) {
          proxyReq.setHeader('X-Forwarded-Host', req.headers.host);
        }
      });
    },
  },
}
```

### 步骤3：重启服务
```bash
./stop-all.sh
./start-all.sh
```

## 总结

### 当前架构
- ✅ 域名控制：Vite前端
- ✅ 配置文件：`vite.config.ts`
- ✅ UI配置：仅作记录参考
- ✅ 适用场景：开发环境

### 优点
- 配置简单
- 立即生效（重启前端即可）
- 不需要前后端同步

### 缺点
- 需要手动修改配置文件
- 不能通过UI实时控制
- 生产环境建议使用Nginx

### 推荐方案
- **开发环境**：使用Vite控制（当前配置）✅
- **生产环境**：使用Nginx/Caddy控制
