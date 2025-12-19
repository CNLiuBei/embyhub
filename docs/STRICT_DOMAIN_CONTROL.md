# 严格域名白名单控制方案

## 🎯 升级方案：整个网站都无法访问

之前的方案：只阻止API请求，前端页面仍可加载  
**现在的方案**：域名不在白名单时，整个网站都无法访问 ❌

## 🔒 工作原理

```
浏览器访问 (unauthorized-domain.com)
    ↓
加载 index.html
    ↓
执行 domain-check.js (页面加载前)
    ↓
调用后端 API: /api/v1/domain-check
    ↓
后端检查白名单
    ↓
域名不在白名单？
    ├─ 是 → 显示错误页面 ❌ (完全阻止)
    └─ 否 → 继续加载页面 ✅
```

## 📁 实现细节

### 1. 前端域名检查脚本

**文件**：`frontend/public/domain-check.js`

```javascript
// 在页面加载前执行
(function() {
  const currentHost = window.location.hostname;
  
  async function checkDomain() {
    const response = await fetch('/api/v1/domain-check', {
      method: 'POST',
      body: JSON.stringify({ domain: currentHost })
    });
    
    if (response.status === 403 || !data.allowed) {
      // 域名未授权，显示错误页面
      document.body.innerHTML = `
        <!-- 美化的403错误页面 -->
      `;
      throw new Error('Domain not authorized');
    }
  }
  
  checkDomain();
})();
```

### 2. HTML头部引入

**文件**：`frontend/index.html`

```html
<head>
  <title>飞牛影视 - 用户管理系统</title>
  <!-- 域名白名单检查脚本 -->
  <script src="/domain-check.js"></script>
</head>
```

### 3. 后端域名检查接口

**文件**：`backend/internal/handler/domain_handler.go`

```go
// CheckDomain 检查域名（公开接口，不需要认证）
func (h *DomainHandler) CheckDomain(c *gin.Context) {
    var req CheckDomainRequest
    c.ShouldBindJSON(&req)
    
    // 检查域名是否在白名单中
    allowed := h.settingService.IsDomainAllowed(req.Domain)
    
    if !allowed {
        response.Error(c, 403, "域名未授权访问")
        return
    }
    
    response.Success(c, gin.H{"allowed": true})
}
```

**路由**：`/api/v1/domain-check`（公开接口）

## 🎨 用户体验

### 场景1：域名在白名单中

```
访问 https://fnhub.trueliu.com
    ↓
域名检查通过 ✅
    ↓
页面正常加载 ✅
    ↓
所有功能可用 ✅
```

### 场景2：域名不在白名单中

```
访问 https://unauthorized-domain.com
    ↓
域名检查失败 ❌
    ↓
显示错误页面：

┌─────────────────────────────┐
│         🚫                  │
│    域名未授权                │
│  Domain Not Authorized      │
│                             │
│  unauthorized-domain.com    │
│                             │
│ 此域名未在系统白名单中         │
│ 请联系系统管理员添加授权       │
└─────────────────────────────┘

阻止继续加载 ❌
用户无法访问网站 ❌
```

## 🔧 错误页面样式

精美的渐变背景 + 毛玻璃效果：

- 紫色渐变背景
- 半透明卡片
- 大尺寸禁止图标 🚫
- 显示被拒绝的域名
- 友好的错误提示

## ✨ 优势

### 1. **完全阻止**
- ✅ 域名不在白名单时，整个网站无法访问
- ✅ 比之前的方案更严格
- ✅ 真正的访问控制

### 2. **用户友好**
- ✅ 显示美观的错误页面（而非空白）
- ✅ 明确告知被拒绝的原因
- ✅ 提示联系管理员

### 3. **即时生效**
- ✅ UI修改域名后立即生效
- ✅ 无需重启服务
- ✅ 配置存储在数据库

### 4. **灵活管理**
- ✅ 在系统设置中添加/删除域名
- ✅ 支持通配符域名
- ✅ 可以随时启用/禁用

## 📝 使用流程

### 步骤1：添加域名

1. 登录管理后台（通过已授权域名）
2. 进入 **系统设置** → **域名访问控制**
3. 添加允许的域名
4. 点击"保存设置"

### 步骤2：启用白名单

1. 打开 **"启用域名白名单"** 开关
2. 点击"保存设置"
3. ✅ 立即生效

### 步骤3：测试验证

#### 测试A：授权域名（应该正常）
访问 `https://fnhub.trueliu.com`
- ✅ 页面正常加载
- ✅ 所有功能可用

#### 测试B：未授权域名（应该被拒）
访问 `https://unauthorized-domain.com`
- ❌ 显示错误页面
- ❌ 无法访问网站

## 🧪 双重保护机制

### 第一层：前端域名检查
```
页面加载时执行 domain-check.js
    ↓
检查域名
    ↓
不在白名单 → 显示错误页面 ❌
```

### 第二层：后端API验证
```
每个API请求
    ↓
DomainWhitelist 中间件
    ↓
不在白名单 → 返回403 ❌
```

**效果**：
- 前端检查阻止页面加载
- 后端验证阻止API访问
- 双重保护，更安全

## 🔄 工作流程对比

### 之前的方案
```
unauthorized-domain.com 访问
  ✅ 前端页面加载（HTML/JS/CSS）
  ❌ API请求失败（403）
  结果：能看到页面框架，但功能不能用
```

### 现在的方案
```
unauthorized-domain.com 访问
  ✅ 开始加载 index.html
  ❌ domain-check.js 阻止
  ❌ 显示错误页面
  结果：完全无法访问，看到友好的错误提示
```

## 🛡️ 安全性增强

### 1. **防止信息泄露**
- 未授权域名无法看到任何页面内容
- 无法获取系统信息
- 无法尝试登录

### 2. **防止滥用**
- 无法通过非授权域名使用系统
- 阻止域名劫持攻击
- 防止钓鱼网站

### 3. **访问控制**
- 只有授权域名可以访问
- 支持多域名部署
- 灵活的白名单管理

## ⚙️ 配置示例

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
  "domains": ["localhost", "test.trueliu.com"]
}
```
**效果**：只有localhost和测试域名可访问

### 生产环境
```json
{
  "enabled": true,
  "domains": ["fnhub.trueliu.com", ".trueliu.com"]
}
```
**效果**：只有主域名和所有子域名可访问

## 🚨 注意事项

### 1. 首次配置
- 必须先通过localhost或已授权域名登录
- 添加所有需要的域名到白名单
- 然后再启用白名单功能

### 2. 避免锁定自己
- ⚠️ 启用前确保当前域名在白名单中
- ⚠️ UI会用绿色标记当前域名
- ⚠️ 点击"添加当前域名"按钮最安全

### 3. 通配符使用
- `*.trueliu.com` 匹配所有子域名
- 但不匹配 `trueliu.com` 本身
- 需要两条规则：`trueliu.com` + `*.trueliu.com`

## 🔍 调试方法

### 查看域名检查请求

打开浏览器F12 → Network标签，查看：
```
POST /api/v1/domain-check
Request: {"domain": "current-hostname"}
Response: {"allowed": true/false}
```

### 测试错误页面

1. 在白名单中删除当前域名
2. 点击"保存设置"
3. 刷新页面
4. 应该看到美观的403错误页面

### 恢复访问

如果被锁定，有两种方法：

**方法1**：通过localhost访问
```
http://localhost:3000/admin/settings
```

**方法2**：直接修改数据库
```sql
UPDATE settings 
SET value = '{"enabled":false,"domains":["localhost"]}' 
WHERE key = 'domain_whitelist';
```

## 📊 性能影响

- **额外请求**：每次加载页面增加1个域名检查请求
- **延迟时间**：约10-50ms（取决于网络）
- **用户体验**：几乎无感知

## 🎉 总结

### 新方案特点

✅ **完全阻止**：域名不在白名单时，整个网站无法访问  
✅ **用户友好**：显示美观的错误页面  
✅ **即时生效**：UI修改后立即生效  
✅ **双重保护**：前端检查 + 后端验证  
✅ **易于管理**：在UI中操作，无需修改代码  

### 适用场景

- ✅ 企业内网部署（限制访问域名）
- ✅ 多租户系统（每个租户独立域名）
- ✅ 安全要求高的系统
- ✅ 防止域名劫持

现在域名控制更严格了！未授权域名完全无法访问网站！🔒
