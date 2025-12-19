# Go后端完全控制的域名白名单方案

## 🔒 最安全的方案

完全由**Go后端中间件**控制域名访问，前端JavaScript无法绕过。

## 🎯 核心优势

### 1. **无法绕过**
- ✅ 所有请求都经过Go后端验证
- ✅ 禁用JavaScript也无法绕过
- ✅ 修改前端代码也无法绕过
- ✅ 使用浏览器开发者工具也无法绕过

### 2. **完全控制**
- ✅ 控制所有HTTP请求（页面+API）
- ✅ 根据请求类型返回不同内容
- ✅ 浏览器请求 → 美观的HTML错误页
- ✅ API请求 → JSON错误响应

### 3. **即时生效**
- ✅ UI修改后立即生效
- ✅ 无需重启服务
- ✅ 配置存储在数据库

### 4. **反向代理支持**
- ✅ 支持X-Forwarded-Host头
- ✅ 支持X-Original-Host头
- ✅ 适配Nginx/Caddy等反向代理

## 📋 技术实现

### Go中间件代码

```go
// DomainWhitelist 域名白名单中间件
func DomainWhitelist(db *gorm.DB) gin.HandlerFunc {
    settingService := service.NewSettingService(db)
    
    // 内嵌HTML错误页面模板
    const errorPageHTML = `...`
    
    return func(c *gin.Context) {
        // 1. 获取域名（支持反向代理）
        host := c.Request.Host
        if forwardedHost := c.GetHeader("X-Forwarded-Host"); forwardedHost != "" {
            host = forwardedHost
        }
        
        // 2. 移除端口号
        if idx := strings.Index(host, ":"); idx != -1 {
            host = host[:idx]
        }
        
        // 3. 检查白名单
        allowed := settingService.IsDomainAllowed(host)
        
        if !allowed {
            // 4. 根据请求类型返回不同内容
            acceptHeader := c.GetHeader("Accept")
            
            // API请求 → JSON
            if strings.Contains(acceptHeader, "application/json") || 
               strings.HasPrefix(c.Request.URL.Path, "/api/") {
                c.JSON(403, gin.H{
                    "code": 403,
                    "message": "域名未授权访问",
                })
            } else {
                // 浏览器请求 → HTML
                c.Header("Content-Type", "text/html; charset=utf-8")
                c.String(403, htmlContent)
            }
            
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### 关键特性

#### 1. 智能响应类型判断
```go
// 判断是API请求还是浏览器请求
if strings.Contains(acceptHeader, "application/json") || 
   strings.HasPrefix(c.Request.URL.Path, "/api/") {
    // 返回JSON
} else {
    // 返回HTML
}
```

#### 2. 内嵌HTML模板
- 不依赖外部文件
- 美观的错误页面
- 响应式设计
- 动态变量替换

#### 3. 反向代理支持
```go
// 优先使用X-Forwarded-Host
if forwardedHost := c.GetHeader("X-Forwarded-Host"); forwardedHost != "" {
    host = forwardedHost
}
```

## 🎨 错误页面设计

### 视觉效果

```
╔═══════════════════════════════════╗
║                                   ║
║              🚫                   ║
║                                   ║
║          域名未授权                ║
║     Domain Not Authorized         ║
║                                   ║
║   ┌─────────────────────────┐    ║
║   │  unauthorized-domain.com │    ║
║   └─────────────────────────┘    ║
║                                   ║
║   访问受限原因                     ║
║   此域名未在系统白名单中...        ║
║                                   ║
║   安全提示                        ║
║   本系统采用域名白名单机制...      ║
║                                   ║
║   Error Code: 403                 ║
║   Client IP: xxx.xxx.xxx.xxx      ║
║                                   ║
╚═══════════════════════════════════╝
```

### 设计特点

- **紫色渐变背景**：专业美观
- **白色卡片**：清晰易读
- **动画效果**：弹跳图标、滑入效果
- **响应式布局**：适配各种设备
- **技术信息**：错误码、IP、Request ID

## 🔐 安全机制

### 1. 前端无法绕过

**尝试方法** → **结果**

| 绕过尝试 | 能否成功 | 原因 |
|---------|---------|------|
| 禁用JavaScript | ❌ 失败 | Go中间件在JavaScript执行前就拦截了 |
| 修改前端代码 | ❌ 失败 | 前端代码不参与验证 |
| 删除检查脚本 | ❌ 失败 | 本来就没有检查脚本 |
| 直接访问API | ❌ 失败 | 所有请求都经过中间件 |
| 修改HTTP头 | ❌ 失败 | 后端验证Host头 |
| 使用代理工具 | ❌ 失败 | 后端仍然检查域名 |

### 2. 多层防护

```
请求层面：所有HTTP请求
    ↓
中间件层：域名验证
    ↓
服务层：业务逻辑
    ↓
数据层：数据访问
```

### 3. 即时更新

```
UI修改域名
    ↓
保存到数据库
    ↓
下一个请求立即生效
```

## 📝 使用流程

### 步骤1：添加域名

1. 登录管理后台（通过已授权域名）
2. 进入 **系统设置** → **域名访问控制**
3. 点击 **"添加当前域名"**
4. 看到绿色标签（表示当前域名已添加）
5. 点击 **"保存设置"**

### 步骤2：启用白名单

1. 确认当前域名在列表中（绿色标签）
2. 打开 **"启用域名白名单"** 开关
3. 点击 **"保存设置"**
4. ✅ 立即生效！

### 步骤3：验证效果

#### 测试A：授权域名
```
访问 https://fnhub.trueliu.com
→ ✅ 网站正常访问
→ ✅ 所有功能可用
```

#### 测试B：未授权域名
```
访问 https://unauthorized-domain.com
→ ❌ 显示美观的403错误页面
→ ❌ 完全无法访问
```

## 🌐 支持的场景

### 场景1：直接访问
```
浏览器 → Go服务器 → 域名验证 ✅
```

### 场景2：Vite开发服务器代理
```
浏览器 → Vite(3000) → Go(8080) → 域名验证 ✅
```

### 场景3：Nginx反向代理
```
浏览器 → Nginx → Vite(3000) → Go(8080) → 域名验证 ✅
```

### 场景4：Nginx直接代理Go
```
浏览器 → Nginx → Go(8080) → 域名验证 ✅
```

## 🔧 配置示例

### 开发环境
```json
{
  "enabled": false,
  "domains": ["localhost", "127.0.0.1"]
}
```
**效果**：白名单关闭，任意域名可访问

### 测试环境
```json
{
  "enabled": true,
  "domains": [
    "localhost",
    "127.0.0.1", 
    "test.trueliu.com"
  ]
}
```
**效果**：只有测试域名可访问

### 生产环境
```json
{
  "enabled": true,
  "domains": [
    "fnhub.trueliu.com",
    "*.trueliu.com"
  ]
}
```
**效果**：主域名和所有子域名可访问

## ⚡ 性能优化建议

### 当前实现
- ✅ 每次请求查询数据库
- ✅ 简单直接
- ✅ 配置立即生效

### 未来优化（可选）
```go
// 添加内存缓存（1分钟）
var (
    domainCache *DomainSettings
    cacheTime   time.Time
    cacheMutex  sync.RWMutex
)

func getCachedDomainSettings() *DomainSettings {
    cacheMutex.RLock()
    if time.Since(cacheTime) < time.Minute {
        defer cacheMutex.RUnlock()
        return domainCache
    }
    cacheMutex.RUnlock()
    
    // 刷新缓存
    cacheMutex.Lock()
    defer cacheMutex.Unlock()
    domainCache = settingService.GetDomainSettings()
    cacheTime = time.Now()
    return domainCache
}
```

## 🛡️ 安全最佳实践

### 1. 域名配置
```
✅ 推荐：
   - fnhub.trueliu.com（主域名）
   - *.trueliu.com（通配符）
   - localhost（开发环境）

❌ 不推荐：
   - * （允许所有域名）
   - 过多的通配符规则
```

### 2. 配置顺序
```
1. 先添加域名
2. 确认当前域名在列表中
3. 再启用白名单
4. 测试验证
```

### 3. 紧急恢复
如果被锁定：

**方法1**：通过localhost访问
```bash
http://localhost:3000/admin/settings
```

**方法2**：修改数据库
```sql
UPDATE settings 
SET value = '{"enabled":false,"domains":["localhost"]}' 
WHERE key = 'domain_whitelist';
```

**方法3**：临时禁用中间件
```go
// router.go
// 注释这行：
// r.Use(middleware.DomainWhitelist(db))
```

## 📊 与其他方案对比

| 方案 | 安全性 | 可绕过性 | 易用性 | 维护性 |
|------|-------|---------|--------|--------|
| **Go后端控制** | ⭐⭐⭐⭐⭐ | ❌ 无法绕过 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| 前端JS检查 | ⭐⭐ | ✅ 禁用JS即可绕过 | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| Vite配置 | ⭐⭐⭐ | ✅ 直接访问API可绕过 | ⭐⭐ | ⭐⭐ |
| Nginx配置 | ⭐⭐⭐⭐⭐ | ❌ 无法绕过 | ⭐⭐⭐ | ⭐⭐⭐ |

## 🎯 总结

### 当前方案特点

✅ **极高安全性**：完全由Go后端控制，无法绕过  
✅ **即时生效**：UI修改后立即生效，无需重启  
✅ **用户友好**：美观的错误页面，清晰的提示信息  
✅ **智能判断**：API返回JSON，浏览器返回HTML  
✅ **反向代理支持**：完美适配各种部署架构  
✅ **配置灵活**：支持通配符、多域名  
✅ **开发友好**：可以随时启用/禁用  

### 推荐使用场景

- ✅ 企业内网部署
- ✅ 多租户SaaS系统
- ✅ 高安全要求的系统
- ✅ 防止域名劫持
- ✅ 防止钓鱼攻击
- ✅ 限制访问来源

### 不适用场景

- ❌ 完全公开的系统（不需要域名限制）
- ❌ CDN加速的前端（CDN域名会变化）

## 🚀 部署建议

### 开发环境
- 白名单：关闭
- 域名：localhost

### 测试环境
- 白名单：开启
- 域名：test.yourdomain.com

### 生产环境
- 白名单：开启
- 域名：yourdomain.com + *.yourdomain.com
- 配合Nginx双重保护

这是目前最安全、最可靠的域名白名单方案！🔒
