# 双重域名控制方案

## 🔒 完全阻止未授权访问

采用**前端Vite + 后端Go**双重控制，连登录页面都无法访问。

## 架构对比

### 方案1：只有后端控制（之前）

```
未授权域名访问
  ↓
Vite: allowedHosts: true
  ↓ 允许所有域名
✅ 前端页面正常加载
  ↓ 用户看到登录界面
  ↓ 尝试登录
API请求 → Go后端检查
  ↓
❌ 返回403错误

结果：能看到登录页面，但无法登录
```

### 方案2：前后端双重控制（现在）

```
未授权域名访问
  ↓
Vite: allowedHosts: ['localhost', '127.0.0.1']
  ↓ 检查域名
❌ Vite拒绝：显示Vite错误页面

结果：完全无法访问，连页面都看不到
```

## 配置方法

### 1. 前端Vite配置

**文件**: `frontend/vite.config.ts`

```typescript
server: {
  allowedHosts: [
    'localhost',
    '127.0.0.1',
    'fnhub.trueliu.com',  // 添加允许的域名
    '.trueliu.com',       // 支持通配符
  ],
}
```

**特点**:
- ✅ 阻止页面加载
- ❌ 需要重启前端服务才生效
- ❌ 需要手动同步域名列表

### 2. 后端Go配置

**文件**: 通过UI或API配置

```json
{
  "enabled": true,
  "domains": [
    "localhost",
    "127.0.0.1",
    "fnhub.trueliu.com"
  ]
}
```

**特点**:
- ✅ 阻止API访问
- ✅ 立即生效，无需重启
- ✅ 通过UI动态配置

## 完整工作流程

### 添加新域名步骤

#### 步骤1: 修改Vite配置

编辑 `frontend/vite.config.ts`:
```typescript
allowedHosts: [
  'localhost',
  '127.0.0.1',
  'fnhub.trueliu.com',  // ← 添加这行
],
```

#### 步骤2: 重启前端服务

```bash
./stop-all.sh
./start-all.sh
```

#### 步骤3: 通过UI添加到后端白名单

1. 访问 `http://localhost:3000/admin/settings`
2. 域名访问控制 → 添加 `fnhub.trueliu.com`
3. 点击"保存设置"

#### 步骤4: 验证

访问 `https://fnhub.trueliu.com`:
- ✅ 前端页面正常加载
- ✅ 登录功能正常
- ✅ 所有API正常

## 两种方案对比

| 特性 | 只后端控制 | 前后端双重控制 |
|------|-----------|--------------|
| 阻止页面加载 | ❌ 否 | ✅ 是 |
| 阻止API访问 | ✅ 是 | ✅ 是 |
| 配置复杂度 | ⭐ 简单 | ⭐⭐⭐ 复杂 |
| 立即生效 | ✅ 是 | ❌ 否（前端需重启） |
| 维护成本 | ⭐ 低 | ⭐⭐⭐ 高 |
| 安全性 | ⭐⭐⭐⭐ 高 | ⭐⭐⭐⭐⭐ 极高 |
| 用户体验 | 看到页面但无法登录 | 完全无法访问 |

## 测试效果

### 测试1: 授权域名

```bash
# 前端访问
http://localhost:3000
→ ✅ 页面正常加载

# 后端API
curl http://localhost:8080/health
→ ✅ {"status":"ok"}
```

### 测试2: 未授权域名

```bash
# 前端访问
http://unauthorized-domain.com:3000
→ ❌ Vite错误：
"Blocked request. This host is not allowed."

# 后端API
curl -H "Host: unauthorized-domain.com" http://localhost:8080/health
→ ❌ Go错误：
{"code":403,"message":"域名未授权访问"}
```

## Vite错误页面

当域名不在Vite白名单时：

```
403 | Blocked request

This host ("unauthorized-domain.com") is not allowed.

To allow this host, add "unauthorized-domain.com" 
to `server.allowedHosts` in vite.config.ts.
```

**效果**: 简单的文本错误页面，无法继续访问

## Go错误页面

当域名通过Vite但不在Go白名单时：

```html
╔══════════════════════════════╗
║          🚫                  ║
║      域名未授权               ║
║  Domain Not Authorized       ║
║                              ║
║  unauthorized-domain.com     ║
║                              ║
║  此域名未在系统白名单中...    ║
╚══════════════════════════════╝
```

**效果**: 美观的HTML错误页面

## 同步管理脚本

创建脚本自动同步Vite配置：

```bash
#!/bin/bash
# sync-domain-whitelist.sh

# 从后端API获取域名列表
DOMAINS=$(curl -s http://localhost:8080/api/v1/admin/settings/domain | jq -r '.data.domains[]')

# 更新vite.config.ts
# ... 自动更新逻辑 ...

# 重启前端服务
./stop-all.sh
./start-all.sh

echo "✅ 域名白名单已同步并重启服务"
```

## 推荐方案

### 开发环境

**推荐**: 只用后端控制

- 前端: `allowedHosts: true`
- 后端: 白名单关闭
- 理由: 方便开发调试

### 测试环境

**推荐**: 前后端双重控制

- 前端: 限制测试域名
- 后端: 白名单开启
- 理由: 模拟生产环境

### 生产环境

**推荐**: Nginx + 后端控制

```nginx
server {
    listen 80;
    server_name fnhub.trueliu.com;  # 只允许特定域名
    
    location / {
        proxy_pass http://localhost:3000;
    }
}
```

- Nginx: 网络层阻止
- 后端: 应用层验证
- 理由: 性能最优，安全性最高

## 注意事项

### 1. 开发体验

双重控制会降低开发体验：
- ❌ 每次添加域名需要重启前端
- ❌ 配置需要在两处维护
- ❌ 容易出现不同步问题

### 2. 紧急恢复

如果配置错误导致无法访问：

**方法1**: 临时改回单层控制
```typescript
// vite.config.ts
allowedHosts: true  // 临时允许所有
```

**方法2**: 通过localhost访问
```
http://localhost:3000/admin/settings
```

**方法3**: 直接修改配置文件
```bash
vim frontend/vite.config.ts
./start-all.sh
```

### 3. 通配符支持

Vite的通配符语法：
```typescript
allowedHosts: [
  '.trueliu.com',  // 匹配 *.trueliu.com
]
```

## 总结

### 单层控制（后端）

✅ **优点**:
- 配置简单
- 立即生效
- 维护方便

❌ **缺点**:
- 前端页面可以加载
- 能看到登录界面

### 双层控制（前端+后端）

✅ **优点**:
- 完全阻止访问
- 连页面都看不到
- 安全性更高

❌ **缺点**:
- 配置复杂
- 需要重启服务
- 维护成本高

### 建议

- **开发环境**: 单层控制（后端）
- **生产环境**: Nginx + 后端控制

现在的配置已经是**双重控制**，未授权域名完全无法访问！
