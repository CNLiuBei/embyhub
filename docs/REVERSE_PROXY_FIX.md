# 反向代理环境下的域名白名单修复

## 问题原因

您使用了反向代理架构：
```
浏览器 → https://fnhub.trueliu.com (反向代理)
           ↓
       http://localhost:3000 (Vite Dev Server)
           ↓ (Vite proxy with changeOrigin: true)
       http://localhost:8080 (后端API)
```

**问题**：`changeOrigin: true` 会将Host头改写为`localhost:8080`，导致域名白名单无法识别真实的访问域名。

## 已修复的内容

### 1. 前端Vite配置修复

**修改前**（vite.config.ts）：
```typescript
proxy: {
  '/api': {
    target: 'http://localhost:8080',
    changeOrigin: true,  // ❌ 问题所在
  },
}
```

**修改后**：
```typescript
proxy: {
  '/api': {
    target: 'http://localhost:8080',
    changeOrigin: false,  // ✅ 保持原始Host
    // 配置代理转发原始Host头
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

### 2. 后端中间件增强

后端`DomainWhitelist`中间件现在支持：
- 检查`X-Forwarded-Host`头（反向代理设置）
- 检查`X-Original-Host`头（备用）
- 详细的调试信息

**工作流程**：
```go
1. 获取Request.Host（可能是localhost:8080）
2. 检查X-Forwarded-Host（fnhub.trueliu.com）✅
3. 使用X-Forwarded-Host进行白名单验证
4. 如果不在白名单→返回403 + 详细调试信息
```

## 完整的请求流程

### 成功的请求（域名在白名单中）

```
浏览器请求:
  https://fnhub.trueliu.com/api/v1/user/info

↓ 您的Nginx/Caddy反向代理
  Host: fnhub.trueliu.com
  
↓ Vite Dev Server (localhost:3000)
  收到: Host: fnhub.trueliu.com
  
↓ Vite Proxy转发到后端
  Host: fnhub.trueliu.com (changeOrigin: false保留)
  X-Forwarded-Host: fnhub.trueliu.com (configure添加)
  
↓ 后端DomainWhitelist中间件
  检查X-Forwarded-Host: fnhub.trueliu.com
  查询白名单: ["localhost", "127.0.0.1", "fnhub.trueliu.com"]
  结果: ✅ 在白名单中，允许访问
  
↓ 后端API处理
  返回: 200 OK + 数据
```

### 失败的请求（域名不在白名单中）

```
浏览器请求:
  https://fnhub.trueliu.com/api/v1/user/info
  (fnhub.trueliu.com已从白名单删除)

↓ ... (同上流程)
  
↓ 后端DomainWhitelist中间件
  检查X-Forwarded-Host: fnhub.trueliu.com
  查询白名单: ["localhost", "127.0.0.1"]
  结果: ❌ 不在白名单中
  
↓ 返回403错误
  {
    "code": 403,
    "message": "域名未授权访问",
    "data": {
      "requestHost": "localhost:8080",
      "forwardedHost": "fnhub.trueliu.com",
      "checkedHost": "fnhub.trueliu.com",
      ...
    }
  }
```

## 测试验证

### 测试1：白名单中有fnhub.trueliu.com

1. 添加`fnhub.trueliu.com`到白名单
2. 点击"保存设置"
3. 访问 `https://fnhub.trueliu.com`
4. 打开F12 → Network标签
5. 查看`/api/`请求 → 应该全部返回200 ✅

### 测试2：白名单中没有fnhub.trueliu.com

1. 删除`fnhub.trueliu.com`标签
2. 点击"保存设置"
3. 刷新页面
4. 打开F12 → Network标签
5. 查看`/api/`请求 → 应该全部返回403 ❌

**403响应示例**：
```json
{
  "code": 403,
  "message": "域名未授权访问",
  "data": {
    "requestHost": "localhost:8080",
    "forwardedHost": "fnhub.trueliu.com",
    "checkedHost": "fnhub.trueliu.com",
    "clientIP": "::1",
    "requestPath": "/api/v1/user/info"
  }
}
```

## 您的反向代理配置建议

### 如果使用Nginx

确保Nginx正确转发Host头：

```nginx
server {
    listen 443 ssl http2;
    server_name fnhub.trueliu.com;
    
    # SSL配置...
    
    location / {
        proxy_pass http://localhost:3000;
        
        # 保持原始Host头（重要！）
        proxy_set_header Host $host;
        
        # 其他常用头
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host $host;
    }
}
```

### 如果使用Caddy

Caddy默认会正确转发头部：

```caddy
fnhub.trueliu.com {
    reverse_proxy localhost:3000
    # Caddy自动处理Host和X-Forwarded-*头
}
```

## 当前白名单配置

从日志中提取的最新配置：
```json
{
  "enabled": true,
  "domains": ["localhost", "127.0.0.1"]
}
```

**建议配置**：
```json
{
  "enabled": true,
  "domains": [
    "localhost",
    "127.0.0.1",
    "fnhub.trueliu.com"  ← 添加您的域名
  ]
}
```

## 调试技巧

### 查看实际的Host头

如果403错误，查看返回的`data`字段：

```json
{
  "code": 403,
  "data": {
    "requestHost": "...",      // 原始Request.Host
    "forwardedHost": "...",    // X-Forwarded-Host值
    "checkedHost": "...",      // 实际验证的域名
    ...
  }
}
```

这些信息会告诉您：
- 后端实际收到的Host是什么
- 是否有X-Forwarded-Host头
- 白名单验证使用的是哪个域名

### 测试curl命令

```bash
# 测试1：模拟localhost访问（应该成功）
curl http://localhost:8080/health

# 测试2：模拟fnhub.trueliu.com访问（取决于白名单）
curl -H "Host: fnhub.trueliu.com" http://localhost:8080/health

# 测试3：带X-Forwarded-Host（模拟反向代理）
curl -H "Host: localhost:8080" \
     -H "X-Forwarded-Host: fnhub.trueliu.com" \
     http://localhost:8080/health
```

## 修复验证清单

- [x] 修改vite.config.ts - changeOrigin设为false
- [x] 添加configure配置转发X-Forwarded-Host
- [x] 后端中间件支持X-Forwarded-Host
- [x] 后端添加详细的403调试信息
- [x] 重启前后端服务
- [ ] 添加fnhub.trueliu.com到白名单
- [ ] 测试访问验证

## 下一步

1. **刷新浏览器页面**（清除缓存：Ctrl+Shift+R）
2. **添加fnhub.trueliu.com到白名单**
3. **点击"保存设置"**
4. **测试访问**：应该能正常使用所有功能
5. **测试删除域名**：删除后应该看到403错误

现在域名白名单应该能正确工作了！🎉
