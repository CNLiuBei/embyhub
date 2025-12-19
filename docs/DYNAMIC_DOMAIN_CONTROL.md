# 动态域名白名单方案

## 🎯 最佳方案架构

采用**前端开放 + 后端动态验证**的架构：

```
浏览器请求 (任意域名)
    ↓
Vite Dev Server (allowedHosts: true)
    ↓ 允许所有域名访问前端页面
    ↓
前端页面正常加载 ✅
    ↓
API请求 (/api/*)
    ↓ Vite Proxy (保持原始Host头)
    ↓
后端 DomainWhitelist 中间件
    ↓ 从数据库读取白名单
    ↓
域名在白名单中？
    ├─ 是 → API正常响应 ✅
    └─ 否 → 返回403错误 ❌
```

## ✨ 优势

### 1. **即时生效**
- ✅ UI修改后立即生效
- ✅ 无需重启任何服务
- ✅ 配置存储在数据库中

### 2. **灵活管理**
- ✅ 在系统设置UI中添加/删除域名
- ✅ 支持通配符域名（*.example.com）
- ✅ 可以随时启用/禁用白名单

### 3. **开发友好**
- ✅ 不需要修改配置文件
- ✅ 前端页面始终可以加载
- ✅ 只控制API访问

### 4. **生产就绪**
- ✅ 适合多环境部署
- ✅ 配置可以通过API管理
- ✅ 支持反向代理环境

## 📋 完整使用流程

### 步骤1：访问设置页面

1. 登录管理后台
2. 进入 **系统设置** → **域名访问控制**

### 步骤2：添加域名

#### 方法A：快速添加当前域名
点击 **"添加当前域名"** 按钮，自动添加当前访问的域名

#### 方法B：手动输入域名
1. 在输入框输入域名（如：`fnhub.trueliu.com`）
2. 按Enter或点击"添加"按钮
3. 看到域名标签出现

### 步骤3：启用白名单

1. 打开 **"启用域名白名单"** 开关
2. 点击 **"保存设置"**
3. ✅ 立即生效！

### 步骤4：测试验证

#### 测试A：当前域名应该正常工作
- 刷新页面
- 所有功能正常 ✅

#### 测试B：删除当前域名测试拒绝
1. 删除当前域名标签
2. 点击"保存设置"
3. 刷新页面
4. 前端页面正常加载
5. 但所有API请求返回403错误 ❌

## 🎨 UI功能说明

### 域名标签颜色

| 颜色 | 含义 |
|------|------|
| 🟢 绿色 | 当前访问域名（已在白名单中） |
| 🔵 蓝色 | 其他白名单域名 |

### 功能按钮

| 按钮 | 功能 |
|------|------|
| 添加当前域名 | 快速添加当前访问的域名 |
| 添加 | 手动添加输入的域名 |
| ✗ (标签) | 删除该域名 |
| 保存设置 | 保存配置到数据库并立即生效 |

### 开关状态

| 状态 | 效果 |
|------|------|
| 关闭 | 允许所有域名访问（开发模式） |
| 开启 | 只允许白名单中的域名访问（生产模式） |

## 🔧 工作原理

### 前端层（Vite）

```typescript
// vite.config.ts
allowedHosts: true  // 允许所有域名访问前端

proxy: {
  '/api': {
    changeOrigin: false,  // 保持原始Host头
    configure: (proxy) => {
      proxy.on('proxyReq', (proxyReq, req) => {
        // 转发原始Host到后端
        proxyReq.setHeader('X-Forwarded-Host', req.headers.host);
      });
    },
  },
}
```

**效果**：
- 前端HTML/JS/CSS可以从任意域名加载
- Host头被正确转发到后端

### 后端层（Go Gin）

```go
// 中间件从数据库读取配置
func DomainWhitelist(db *gorm.DB) gin.HandlerFunc {
    settingService := service.NewSettingService(db)
    return func(c *gin.Context) {
        // 获取Host（支持X-Forwarded-Host）
        host := getHostFromRequest(c)
        
        // 动态检查白名单
        if !settingService.IsDomainAllowed(host) {
            c.JSON(403, gin.H{
                "code": 403,
                "message": "域名未授权访问",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

**效果**：
- 每次API请求都检查当前数据库配置
- UI修改后立即生效
- 返回详细的403错误信息

### 数据库存储

```json
{
  "key": "domain_whitelist",
  "value": {
    "enabled": true,
    "domains": ["localhost", "127.0.0.1", "fnhub.trueliu.com"]
  }
}
```

**特点**：
- 配置持久化
- 可通过API修改
- 支持多服务器同步

## 📝 域名格式支持

### 完整域名
```
localhost
127.0.0.1
fnhub.trueliu.com
api.example.com
```

### 通配符域名
```
*.trueliu.com
*.example.com
```

**通配符规则**：
- `*.trueliu.com` 匹配：
  - ✅ `sub.trueliu.com`
  - ✅ `api.trueliu.com`
  - ✅ `test.trueliu.com`
  - ❌ `trueliu.com`（需单独添加）

## 🧪 测试场景

### 场景1：开发环境（白名单关闭）

**配置**：
```json
{
  "enabled": false,
  "domains": ["localhost"]
}
```

**效果**：
- 任意域名都可以访问 ✅
- 适合本地开发和调试

### 场景2：生产环境（白名单开启）

**配置**：
```json
{
  "enabled": true,
  "domains": ["fnhub.trueliu.com", ".trueliu.com"]
}
```

**效果**：
- `fnhub.trueliu.com` → ✅ 允许
- `api.trueliu.com` → ✅ 允许（通配符匹配）
- `other-domain.com` → ❌ 403错误

### 场景3：多域名部署

**配置**：
```json
{
  "enabled": true,
  "domains": [
    "localhost",
    "fnhub.trueliu.com",
    "app.example.com",
    "*.internal.com"
  ]
}
```

**效果**：
- 支持多个主域名
- 支持内网域名通配符

## 🚀 实际访问效果

### 白名单关闭时

```
访问 https://fnhub.trueliu.com
  ✅ 前端页面正常加载
  ✅ API请求正常工作
  ✅ 所有功能可用

访问 https://any-domain.com
  ✅ 前端页面正常加载
  ✅ API请求正常工作
  ✅ 所有功能可用
```

### 白名单开启时（fnhub.trueliu.com在列表中）

```
访问 https://fnhub.trueliu.com
  ✅ 前端页面正常加载
  ✅ API请求正常工作
  ✅ 所有功能可用

访问 https://unauthorized-domain.com
  ✅ 前端页面正常加载（HTML/JS/CSS）
  ❌ API请求返回403错误
  ❌ 功能不可用（无法获取数据）
```

## 🔍 调试方法

### 查看403错误详情

打开浏览器F12 → Network标签，查看被拒绝的请求：

```json
{
  "code": 403,
  "message": "域名未授权访问",
  "data": {
    "requestHost": "unauthorized-domain.com:3000",
    "forwardedHost": "unauthorized-domain.com",
    "checkedHost": "unauthorized-domain.com",
    "clientIP": "::1",
    "requestPath": "/api/v1/user/info"
  }
}
```

### 查看当前配置

方法1：UI查看
- 进入系统设置 → 域名访问控制

方法2：API查询
```bash
curl http://localhost:8080/api/v1/admin/settings/domain
```

方法3：数据库查询
```sql
SELECT * FROM settings WHERE key = 'domain_whitelist';
```

## 💡 常见问题

### Q1: 修改配置后需要重启服务吗？

❌ **不需要**！配置立即生效。

### Q2: 前端页面能加载但功能不能用？

✅ **这是正常的**！域名不在白名单中：
- 前端HTML/JS/CSS可以加载
- 但API请求被拒绝（403）

### Q3: 如何快速测试白名单功能？

1. 添加当前域名到白名单
2. 开启白名单开关
3. 保存设置
4. 删除当前域名
5. 保存设置
6. 刷新页面 → 应该看到API 403错误

### Q4: 通配符域名如何使用？

- UI输入：`*.trueliu.com` 或 `.trueliu.com`
- 匹配所有子域名，但不匹配主域名本身

### Q5: 生产环境建议配置？

```json
{
  "enabled": true,
  "domains": [
    "yourdomain.com",
    "*.yourdomain.com"
  ]
}
```

同时建议在Nginx层也做域名限制（双重保护）

## 📊 性能影响

- **数据库查询**：每次API请求会查询一次数据库
- **缓存优化**：后续可以添加Redis缓存提升性能
- **影响程度**：可忽略（查询非常简单且频率可控）

## 🎉 总结

### 当前方案特点

✅ **灵活**：UI实时修改，立即生效  
✅ **安全**：只控制API访问，不影响前端加载  
✅ **友好**：开发环境不受限制  
✅ **完整**：支持通配符、多域名  
✅ **可调试**：详细的403错误信息  

### 推荐使用场景

- ✅ 开发环境：关闭白名单（方便调试）
- ✅ 测试环境：开启白名单（限制测试域名）
- ✅ 生产环境：开启白名单 + Nginx双重保护

这就是目前最灵活方便的域名白名单方案！🚀
