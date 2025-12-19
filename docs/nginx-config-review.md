# 宝塔Nginx反向代理配置检查

## ✅ 配置分析

### 你的配置总结

```nginx
server {
    listen 443 ssl;
    server_name fnhub.trueliu.com;
    
    # SSL证书
    ssl_certificate /www/server/panel/vhost/cert/fnhub.trueliu.com/fullchain.pem;
    ssl_certificate_key /www/server/panel/vhost/cert/fnhub.trueliu.com/privkey.pem;
    
    # 反向代理到前端
    location ^~ / {
        proxy_pass http://127.0.0.1:3000;
        proxy_set_header Host $http_host;
        proxy_set_header X-Forwarded-Host $host;  # ✅ Go后端需要这个
        proxy_set_header X-Forwarded-Proto $scheme;
        # ... 其他配置
    }
}
```

### ✅ 配置检查结果

| 项目 | 状态 | 说明 |
|------|------|------|
| SSL证书 | ✅ 已配置 | Let's Encrypt证书 |
| HTTP重定向 | ✅ 已启用 | 自动跳转HTTPS |
| 反向代理 | ✅ 已配置 | 代理到localhost:3000 |
| X-Forwarded-Host | ✅ 已配置 | Go后端域名验证需要 |
| WebSocket | ✅ 已支持 | 支持实时连接 |

### 🔧 配置说明

#### 1. 反向代理路径

```nginx
location ^~ / {
    proxy_pass http://127.0.0.1:3000;  # Vite前端服务器
}
```

**工作流程**：
```
用户访问 https://fnhub.trueliu.com
    ↓
Nginx接收请求（443端口）
    ↓
SSL终结（解密HTTPS）
    ↓
代理到 localhost:3000（HTTP）
    ↓
Vite开发服务器
    ↓
API请求 → Vite代理 → Go后端(8080)
```

#### 2. 关键请求头

```nginx
proxy_set_header Host $http_host;              # 原始Host
proxy_set_header X-Forwarded-Host $host;       # ✅ Go后端用这个验证域名
proxy_set_header X-Forwarded-Proto $scheme;    # https
proxy_set_header X-Real-IP $remote_addr;       # 真实IP
```

**Go后端会检查**：`X-Forwarded-Host: fnhub.trueliu.com`

#### 3. HTTP/HTTPS重定向

```nginx
if ($server_port != 443) {
    rewrite ^(/.*)$ https://$host$1 permanent;
}
```

**效果**：访问 http://fnhub.trueliu.com 自动跳转到 https://

## 📝 当前架构

### 完整请求流程

```
用户浏览器
    ↓ HTTPS (443)
Nginx反向代理
    ↓ HTTP (3000)
Vite开发服务器
    ↓ /api/* 请求
Vite Proxy
    ↓ HTTP (8080)
Go后端服务器
    ↓ 域名白名单检查
    ✅ X-Forwarded-Host: fnhub.trueliu.com
```

### 端口说明

| 端口 | 服务 | 说明 |
|------|------|------|
| 80 | Nginx HTTP | 自动重定向到443 |
| 443 | Nginx HTTPS | SSL终结，代理到3000 |
| 3000 | Vite前端 | React开发服务器 |
| 8080 | Go后端 | API服务器 |

## ✅ 需要做的事

### 步骤1: 确认服务运行

检查3000和8080端口是否运行：
```bash
netstat -tlnp | grep -E ":3000|:8080"
# 或
ss -tlnp | grep -E ":3000|:8080"
```

期望结果：
```
tcp  0.0.0.0:3000  LISTEN  (vite/node)
tcp  :::8080       LISTEN  (go后端)
```

### 步骤2: 添加域名到白名单

1. 访问：`http://localhost:3000/admin/settings`
2. 域名访问控制标签
3. 添加域名：`fnhub.trueliu.com`
4. 打开"启用域名白名单"开关
5. 点击"保存设置"

**重要**：必须先通过localhost访问添加域名，否则会被锁定！

### 步骤3: 测试访问

```bash
# 测试HTTPS访问
curl -I https://fnhub.trueliu.com

# 期望看到：
# HTTP/2 200
# server: nginx
```

## 🐛 可能的问题

### 问题1: 503 Bad Gateway

**原因**：3000端口服务未运行

**解决**：
```bash
cd /vol1/1000/FnHub/feiniu-user-system
./start-all.sh
```

### 问题2: 域名白名单拒绝

**现象**：能看到登录页面，但无法登录

**原因**：域名未添加到白名单

**解决**：
```bash
# 通过localhost访问
http://localhost:3000/admin/settings
# 添加 fnhub.trueliu.com 到白名单
```

### 问题3: SSL证书错误

**现象**：浏览器提示证书无效

**原因**：证书过期或配置错误

**解决**：
```bash
# 宝塔面板 → 网站 → fnhub.trueliu.com → SSL
# 重新申请Let's Encrypt证书
```

## 🔒 安全性检查

### 当前安全层级

1. **Nginx层** (网络层)
   - ✅ SSL/TLS加密
   - ✅ HTTP强制跳转HTTPS
   - ✅ 只允许 fnhub.trueliu.com 域名

2. **Go后端层** (应用层)
   - ✅ 域名白名单验证
   - ✅ 动态配置，立即生效
   - ✅ 检查 X-Forwarded-Host

### 双重保护效果

```
未授权域名访问:
  ↓
Nginx: 配置中只有fnhub.trueliu.com
  → 如果换其他域名，Nginx不会匹配这个server块
  
Go后端: 域名白名单检查
  → 如果X-Forwarded-Host不在白名单中，返回403
```

## 🎯 配置优化建议

### 可选优化1: 添加API直接代理

如果需要Nginx直接代理API（绕过Vite）：

```nginx
# 在 location ^~ / 之前添加
location ^~ /api/ {
    proxy_pass http://127.0.0.1:8080;
    proxy_set_header Host $http_host;
    proxy_set_header X-Forwarded-Host $host;
    proxy_set_header X-Forwarded-Proto $scheme;
}

# 前端静态资源
location ^~ / {
    proxy_pass http://127.0.0.1:3000;
    # ...
}
```

**好处**：
- 减少一层代理
- 提高API性能

### 可选优化2: 添加缓存

已配置缓存：
```nginx
proxy_cache_path /www/wwwroot/fnhub.trueliu.com/proxy_cache_dir 
    levels=1:2 
    keys_zone=fnhub_trueliu_com_cache:20m 
    inactive=1d 
    max_size=5g;
```

可以为静态资源启用缓存：
```nginx
location ~* \.(jpg|jpeg|png|gif|ico|css|js)$ {
    proxy_pass http://127.0.0.1:3000;
    proxy_cache fnhub_trueliu_com_cache;
    proxy_cache_valid 200 1d;
}
```

## ✅ 测试清单

- [ ] 服务运行检查
  ```bash
  ./check_services.sh
  ```

- [ ] 域名白名单添加
  ```
  localhost:3000/admin/settings → 添加fnhub.trueliu.com
  ```

- [ ] HTTPS访问测试
  ```bash
  curl -I https://fnhub.trueliu.com
  ```

- [ ] 登录功能测试
  ```
  打开 https://fnhub.trueliu.com
  尝试登录
  ```

- [ ] 未授权域名测试
  ```bash
  curl -H "Host: unauthorized.com" http://localhost:3000/api/v1/health
  # 应该返回403
  ```

## 🎉 总结

你的Nginx配置已经很完善了！

**当前状态**：
- ✅ SSL证书配置正确
- ✅ 反向代理配置正确
- ✅ 请求头转发正确
- ✅ 支持WebSocket
- ✅ HTTP自动跳转HTTPS

**需要做的**：
1. 确保服务运行（./start-all.sh）
2. 添加域名到白名单（localhost:3000/admin/settings）
3. 测试访问（https://fnhub.trueliu.com）

配置完全没问题，可以直接使用！🚀
