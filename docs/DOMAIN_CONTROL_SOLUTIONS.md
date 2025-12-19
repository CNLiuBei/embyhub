# 域名控制方案对比

## 需求
1. ❌ 完全阻止未授权域名访问（连登录页面都看不到）
2. ✅ 不需要重启服务
3. ✅ 通过UI动态配置
4. ✅ 防止外人入侵

## 方案对比

### 方案1：Vite allowedHosts（当前）

```typescript
// vite.config.ts
allowedHosts: ['localhost', '127.0.0.1']
```

✅ **优点**：
- 完全阻止前端页面加载
- 安全性高

❌ **缺点**：
- **必须重启服务才能生效**
- 无法动态配置
- 开发体验差

**结论**：不满足"不需要重启"的需求 ❌

---

### 方案2：只用Go后端控制

```
前端：allowedHosts: true（允许所有）
后端：动态域名白名单
```

✅ **优点**：
- ✅ 不需要重启服务
- ✅ UI动态配置
- ✅ 立即生效

❌ **缺点**：
- ❌ 前端页面可以加载（能看到登录界面）
- ❌ 只是API被拒绝

**结论**：不满足"完全阻止访问"的需求 ❌

---

### 方案3：生产环境使用Nginx ⭐ 推荐

```nginx
# nginx.conf
server {
    listen 80;
    server_name fnhub.trueliu.com;  # 只允许这个域名
    
    location / {
        proxy_pass http://localhost:3000;
    }
}

# 其他域名nginx直接返回403
server {
    listen 80 default_server;
    return 403;
}
```

✅ **优点**：
- ✅ 完全阻止未授权域名
- ✅ 不需要重启应用
- ✅ 网络层拦截，性能最好
- ✅ 安全性最高

❌ **缺点**：
- 需要Nginx配置
- 修改Nginx配置需要reload（但很快）

**结论**：最佳方案！ ✅

---

### 方案4：Nginx动态读取配置 ⭐⭐ 完美方案

使用Nginx + Lua动态读取数据库配置：

```nginx
location / {
    access_by_lua_block {
        -- 从数据库或Redis读取域名白名单
        local host = ngx.var.host
        local allowed = check_domain_whitelist(host)
        
        if not allowed then
            ngx.exit(403)
        end
    }
    
    proxy_pass http://localhost:3000;
}
```

✅ **优点**：
- ✅ 完全阻止未授权域名
- ✅ 不需要重启任何服务
- ✅ UI动态配置
- ✅ 立即生效
- ✅ 安全性最高

❌ **缺点**：
- 需要安装OpenResty或lua-nginx-module
- 配置相对复杂

**结论**：完美方案！ ✅✅

---

### 方案5：开发/生产分离 ⭐ 实用方案

**开发环境**：
```typescript
// vite.config.ts
allowedHosts: true  // 开发时允许所有
```
- 只用Go后端控制
- 前端页面能加载但API被拒
- 方便开发调试

**生产环境**：
```nginx
# 生产环境使用Nginx
server {
    server_name fnhub.trueliu.com;
    location / {
        proxy_pass http://localhost:3000;
    }
}
```
- Nginx层完全阻止
- 不需要重启应用
- 通过后端API动态配置Nginx

✅ **优点**：
- 开发环境方便
- 生产环境安全
- 各取所长

**结论**：最实用的方案！ ✅

---

## 推荐方案

### 立即可用方案

**修改为只用Go后端控制**：

```typescript
// vite.config.ts
allowedHosts: true  // 允许所有域名
```

```go
// 后端中间件继续工作
middleware.DomainWhitelist(db)
```

**效果**：
- ✅ 不需要重启
- ✅ UI动态配置
- ⚠️ 前端页面可以加载，但无法登录

**权衡**：
接受"能看到登录页面"的代价，换取"不需要重启"的便利性。

### 生产环境方案

**使用Nginx反向代理**：

```bash
# 安装Nginx
apt install nginx

# 配置文件
cat > /etc/nginx/sites-available/fnhub << 'EOF'
server {
    listen 80;
    server_name fnhub.trueliu.com;
    
    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Host $host;
    }
}

# 默认拒绝其他域名
server {
    listen 80 default_server;
    return 403 "Domain not authorized";
}
EOF

# 启用配置
ln -s /etc/nginx/sites-available/fnhub /etc/nginx/sites-enabled/
nginx -t
systemctl reload nginx
```

**效果**：
- ✅ 完全阻止未授权域名
- ✅ 只需要reload nginx（很快）
- ✅ 应用不需要重启

### 完美方案（需要额外配置）

**Nginx + Redis + 后端API**：

1. 后端API修改域名白名单时，同时更新Redis
2. Nginx使用Lua从Redis读取白名单
3. 所有请求在Nginx层就被拦截

```lua
-- nginx.conf
http {
    lua_shared_dict domain_whitelist 1m;
    
    server {
        listen 80;
        
        access_by_lua_block {
            local redis = require "resty.redis"
            local red = redis:new()
            red:connect("127.0.0.1", 6379)
            
            local host = ngx.var.host
            local allowed = red:sismember("domain_whitelist", host)
            
            if allowed == 0 then
                ngx.exit(403)
            end
        }
        
        location / {
            proxy_pass http://localhost:3000;
        }
    }
}
```

## 建议

### 开发阶段（现在）

```typescript
// vite.config.ts
allowedHosts: true  // 改回允许所有
```

**理由**：
- 方便开发调试
- 不需要频繁重启
- 后端仍然有保护

### 部署生产（推荐）

```nginx
# 使用Nginx控制域名
server {
    server_name fnhub.trueliu.com;
    # ...
}
```

**理由**：
- 完全阻止未授权访问
- 不需要重启应用
- 性能最优

## 操作建议

### 立即修改（恢复便利性）

```bash
# 修改vite.config.ts
sed -i "s/allowedHosts: \[/allowedHosts: true, \/\/ [/" frontend/vite.config.ts

# 重启服务（最后一次重启）
./stop-all.sh
./start-all.sh
```

之后：
- ✅ 添加域名不需要重启
- ✅ UI配置立即生效
- ⚠️ 前端页面可以访问（但API被拒绝）

### 生产环境部署

使用Nginx配置域名白名单，应用层不需要关心。

## 总结

| 需求 | Vite控制 | Go后端控制 | Nginx控制 |
|------|---------|-----------|----------|
| 完全阻止访问 | ✅ | ❌ | ✅ |
| 不需要重启 | ❌ | ✅ | ✅ |
| UI动态配置 | ❌ | ✅ | ⚠️ |
| 开发友好 | ❌ | ✅ | ⚠️ |
| 生产推荐 | ❌ | ❌ | ✅ |

**最终建议**：
- **开发环境**：只用Go后端控制（方便）
- **生产环境**：Nginx + Go后端（安全）
