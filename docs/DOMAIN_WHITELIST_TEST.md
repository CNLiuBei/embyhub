# 域名白名单功能测试说明

## 工作原理

域名白名单功能通过中间件检查每个HTTP请求的Host头部：

```
请求 → 域名白名单中间件 → 检查是否启用 → 检查域名 → 允许/拒绝
```

### 验证逻辑

1. **未启用白名单** (`enabled: false`)：允许所有域名访问 ✅
2. **已启用白名单** (`enabled: true`)：
   - 域名在白名单中 → 允许访问 ✅
   - 域名不在白名单中 → 403错误 ❌

## 测试步骤

### 测试1：验证白名单功能（未启用）

**预期结果**：无论域名在不在列表中，都能正常访问

1. 确保"启用域名白名单"开关是**关闭**状态
2. 域名列表可以为空或有任意域名
3. 点击"保存设置"
4. 刷新页面
5. ✅ 应该能正常访问

### 测试2：验证白名单功能（已启用，有当前域名）

**预期结果**：当前域名在白名单中，能正常访问

1. 点击"添加当前域名"按钮（会添加 `localhost` 或实际域名）
2. 确认域名列表中有当前访问的域名（绿色标签）
3. **打开**"启用域名白名单"开关
4. 点击"保存设置"
5. 刷新页面
6. ✅ 应该能正常访问

### 测试3：验证白名单功能（已启用，无当前域名）

**预期结果**：当前域名不在白名单中，403错误

1. 确保"启用域名白名单"开关是**打开**状态
2. **删除**当前访问域名的标签
3. 点击"保存设置"
4. 刷新页面
5. ❌ 应该看到403错误："域名未授权访问"

### 测试4：恢复访问

**如果被锁定无法访问**：

方法1：使用其他允许的域名访问
```bash
# 如果白名单中有 localhost
http://localhost:3000/admin/settings

# 如果白名单中有其他域名
http://other-domain.com/admin/settings
```

方法2：通过后端API禁用白名单
```bash
# 需要有效的管理员token
curl -X PUT http://localhost:8080/api/v1/admin/settings/domain \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"enabled":false,"domains":["localhost"]}'
```

方法3：直接修改数据库
```sql
-- 连接到PostgreSQL数据库
psql -U username -d feiniu_user_system

-- 禁用域名白名单
UPDATE settings 
SET value = '{"enabled":false,"domains":["localhost","127.0.0.1"]}' 
WHERE key = 'domain_whitelist';
```

## 当前设置查看

### 方法1：前端页面
访问：系统设置 → 域名访问控制

### 方法2：后端日志
```bash
tail -f logs/backend_*.log | grep domain_whitelist
```

### 方法3：数据库查询
```sql
SELECT * FROM settings WHERE key = 'domain_whitelist';
```

示例输出：
```json
{
  "enabled": true,
  "domains": ["localhost", "127.0.0.1", "fnhub.trueliu.com"]
}
```

## 常见问题

### Q1: 为什么我删除了域名还能访问？

**A**: 检查"启用域名白名单"开关是否打开。只有开关打开时，白名单才会生效。

### Q2: 如何测试通配符域名？

**A**: 
- 添加 `*.example.com` 到白名单
- 从 `sub.example.com` 访问 → 允许 ✅
- 从 `another.example.com` 访问 → 允许 ✅
- 从 `example.com` 访问 → 拒绝 ❌（需单独添加）

### Q3: 开发环境建议配置？

**A**: 建议添加：
```
localhost
127.0.0.1
```

### Q4: 生产环境建议配置？

**A**: 只添加实际使用的域名：
```
yourdomain.com
*.yourdomain.com  （如果有子域名）
```

## 安全建议

1. **生产环境务必启用**：防止通过非法域名访问
2. **最小化白名单**：只添加必要的域名
3. **定期审查**：移除不再使用的域名
4. **保留localhost**：方便本地调试和紧急处理

## 技术实现

- 后端中间件：`middleware.DomainWhitelist()`
- 服务层：`service.SettingService.IsDomainAllowed()`
- 数据库表：`settings` (key='domain_whitelist')
- 前端页面：`pages/admin/Settings.tsx`

## 日志示例

**正常访问**（域名在白名单中）：
```
HTTP请求 GET /api/v1/user/info → 200 OK
```

**拒绝访问**（域名不在白名单中）：
```
HTTP请求 GET /api/v1/user/info → 403 Forbidden
响应: {"code":403,"message":"域名未授权访问"}
```
