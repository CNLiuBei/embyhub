# 域名访问控制功能

## 功能概述

添加了域名白名单功能，管理员可以在系统设置中配置允许访问的域名，增强系统安全性。

## 功能特性

### 1. 域名白名单管理
- ✅ 支持启用/禁用域名白名单
- ✅ 支持添加多个域名
- ✅ 支持通配符子域名（如 `*.example.com`）
- ✅ 可视化域名列表（Tag标签显示）
- ✅ 一键删除域名

### 2. 安全特性
- ⚠️ 启用白名单后，只有列表中的域名可以访问前端
- 🔐 防止非法域名访问
- 🛡️ 保护生产环境安全

## 技术实现

### 后端实现

#### 1. 数据模型（`models/user.go`）
```go
// Setting 系统设置模型
type Setting struct {
    ID        uint64    `gorm:"primaryKey" json:"id"`
    Key       string    `gorm:"size:64;uniqueIndex;not null" json:"key"`
    Value     string    `gorm:"type:text" json:"value"`
    Type      string    `gorm:"size:32" json:"type"`
    UpdatedBy uuid.UUID `gorm:"type:uuid" json:"updated_by"`
    UpdatedAt time.Time `json:"updated_at"`
    CreatedAt time.Time `json:"created_at"`
}
```

#### 2. 服务层（`service/setting_service.go`）
- `GetDomainSettings()` - 获取域名白名单设置
- `SaveDomainSettings()` - 保存域名白名单设置
- `IsDomainAllowed()` - 检查域名是否允许访问

域名设置结构：
```go
type DomainSettings struct {
    Enabled bool     `json:"enabled"`
    Domains []string `json:"domains"`
}
```

#### 3. API层（`handler/setting_handler.go`）
- `GET /api/v1/admin/settings/domain` - 获取域名设置
- `PUT /api/v1/admin/settings/domain` - 保存域名设置

### 前端实现

#### 1. API服务（`services/api.ts`）
```typescript
getDomainSettings: () => api.get('/admin/settings/domain')
saveDomainSettings: (data) => api.put('/admin/settings/domain', data)
```

#### 2. UI界面（`pages/admin/Settings.tsx`）
- 域名白名单开关
- 域名输入框（支持回车快捷添加）
- 域名列表（Tag形式展示）
- 保存按钮

## 使用说明

### 1. 访问设置页面
1. 使用管理员账号登录（role >= 2）
2. 进入 **系统设置** 页面
3. 找到 **域名访问控制** 卡片

### 2. 配置域名白名单
1. 点击开关启用域名白名单
2. 在输入框中输入允许的域名
3. 按 **Enter** 或点击 **添加** 按钮
4. 重复步骤2-3添加更多域名
5. 点击 **保存设置** 按钮

### 3. 域名格式说明
- **完整域名**：`example.com`、`www.example.com`
- **IP地址**：`127.0.0.1`、`localhost`
- **通配符子域名**：`*.example.com`（匹配所有子域名）

### 4. 建议配置
开发环境：
```
localhost
127.0.0.1
```

生产环境：
```
fnhub.trueliu.com
*.trueliu.com
```

## 注意事项

⚠️ **重要提醒**

1. **启用前务必添加正确的域名**
   - 如果域名配置错误，可能导致无法访问系统
   - 建议先添加当前访问域名，再启用白名单

2. **通配符使用**
   - `*.example.com` 匹配所有 `example.com` 的子域名
   - 不匹配 `example.com` 本身，需要单独添加

3. **数据持久化**
   - 配置保存在数据库 `settings` 表中
   - Key为 `domain_whitelist`
   - Value为JSON格式

## 数据库结构

```sql
CREATE TABLE settings (
    id BIGSERIAL PRIMARY KEY,
    key VARCHAR(64) UNIQUE NOT NULL,
    value TEXT,
    type VARCHAR(32),
    updated_by UUID,
    updated_at TIMESTAMP,
    created_at TIMESTAMP
);
```

示例数据：
```json
{
  "key": "domain_whitelist",
  "value": "{\"enabled\":true,\"domains\":[\"localhost\",\"*.trueliu.com\"]}",
  "type": "domain"
}
```

## API文档

### 获取域名设置
```http
GET /api/v1/admin/settings/domain
Authorization: Bearer {token}
```

响应：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "enabled": true,
    "domains": ["localhost", "*.trueliu.com"]
  }
}
```

### 保存域名设置
```http
PUT /api/v1/admin/settings/domain
Authorization: Bearer {token}
Content-Type: application/json

{
  "enabled": true,
  "domains": ["localhost", "*.trueliu.com", "fnhub.trueliu.com"]
}
```

响应：
```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

## 修改文件清单

### 后端
- ✅ `internal/models/user.go` - 添加Setting模型
- ✅ `internal/service/setting_service.go` - 新建服务层
- ✅ `internal/handler/setting_handler.go` - 新建处理层
- ✅ `internal/router/router.go` - 添加API路由
- ✅ `internal/database/postgres.go` - 添加数据库迁移

### 前端
- ✅ `src/services/api.ts` - 添加API方法
- ✅ `src/pages/admin/Settings.tsx` - 添加UI界面

## 安全建议

1. **生产环境必须启用**
   - 防止通过非法域名访问系统
   - 减少安全风险

2. **定期审查域名列表**
   - 移除不再使用的域名
   - 确保域名列表最小化

3. **结合其他安全措施**
   - HTTPS强制
   - CORS配置
   - 防火墙规则
   - CDN保护

## 常见问题

### Q: 启用后无法访问怎么办？
A: 通过数据库直接修改配置：
```sql
UPDATE settings 
SET value = '{"enabled":false,"domains":[]}'
WHERE key = 'domain_whitelist';
```

### Q: 支持端口号吗？
A: 当前只匹配域名/IP，不包含端口号

### Q: 如何临时禁用？
A: 在设置页面关闭开关即可，域名列表会保留

---

**版本**: v1.0  
**更新时间**: 2025-12-08  
**作者**: Cascade AI Assistant
