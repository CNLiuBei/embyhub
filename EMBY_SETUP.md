# Emby 用户管理系统配置指南

本系统已改造为支持 Emby 官方 API，可以同步管理 Emby 用户。

## 配置方式

### 通过管理后台配置（推荐）

1. 登录管理后台（默认账号：admin / admin123）
2. 进入 **系统设置** → **媒体服务配置**
3. 启用媒体服务
4. 选择服务模式（Emby 官方 或 飞牛影视）
5. 填写服务器地址和认证信息
6. 设置模板用户（可选，新用户将复制此用户的权限）
7. 点击 **测试连接** 验证配置
8. 保存设置

配置保存后立即生效，无需重启服务。

## 获取 Emby API 密钥

1. 登录 Emby 管理后台
2. 进入 **设置** → **高级** → **API 密钥**
3. 点击 **新建 API 密钥**
4. 输入应用名称（如 "用户管理系统"）
5. 复制生成的 API 密钥

## 模板用户配置

为了让新注册用户获得合适的权限，建议在 Emby 中创建一个模板用户（如 `test`）：

1. 在 Emby 管理后台创建一个普通用户（如 `test`）
2. 配置该用户的权限：
   - 媒体库访问权限
   - 远程访问权限
   - 播放权限等
3. 在本系统的媒体服务配置中，将 `test` 设置为模板用户
4. 新注册用户将自动获得与 `test` 用户相同的权限配置

## 功能说明

### 用户同步
- 注册：在本系统注册时，会自动在 Emby 创建对应用户，并复制模板用户的权限
- 登录：登录时会同步验证 Emby 账号
- 密码修改：修改密码时会同步更新 Emby 密码
- 状态管理：禁用/启用用户时会同步更新 Emby 用户状态
- 删除：删除用户时会同步删除 Emby 用户

### 媒体库访问
- 会员用户可以浏览 Emby 媒体库
- 支持电影、电视剧分类浏览
- 支持搜索功能
- 支持查看媒体详情、季、剧集信息

### Emby 用户列表
- 管理后台可以查看 Emby 服务器上的所有用户
- 显示用户状态、权限信息
- 方便选择模板用户

## 支持的 Emby API

- 用户管理：创建、删除、修改密码、启用/禁用、复制权限
- 用户认证：登录验证
- 媒体库：获取媒体库列表、媒体项列表
- 媒体详情：获取电影/剧集详情
- 剧集信息：获取季列表、剧集列表
- 搜索：全局搜索媒体
- 图片：代理 Emby 图片资源

## 兼容模式

如果你使用的是飞牛影视而非 Emby，可以在管理后台切换到飞牛模式：

1. 进入 **系统设置** → **媒体服务配置**
2. 服务模式选择 **飞牛影视**
3. 填写飞牛服务器地址（如：http://your-feiniu-server:8005/v/api/v1）
4. 填写管理员用户名和密码
5. 保存设置

## 启动服务

```bash
# 启动所有服务
./start-all.sh

# 或分别启动
cd backend && go run cmd/server/main.go
cd frontend && npm run dev
```

## 访问地址

- 前端：http://localhost:3000
- 后端 API：http://localhost:8080
- 默认管理员：admin / admin123

## API 接口

### Emby 设置 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/admin/settings/emby | 获取 Emby 设置 |
| PUT | /api/v1/admin/settings/emby | 保存 Emby 设置 |
| POST | /api/v1/admin/settings/emby/test | 测试 Emby 连接 |
| GET | /api/v1/admin/emby/users | 获取 Emby 用户列表 |

### 请求示例

```json
// PUT /api/v1/admin/settings/emby
{
  "enabled": true,
  "mode": "emby",
  "base_url": "http://192.168.1.100:8096",
  "api_key": "your-api-key",
  "template_user": "test"
}
```

### 响应示例

```json
// GET /api/v1/admin/emby/users
{
  "code": 200,
  "data": [
    {
      "id": "xxx",
      "name": "admin",
      "is_admin": true,
      "is_disabled": false,
      "enable_media_playback": true,
      "enable_remote_access": true,
      "enable_all_folders": true
    },
    {
      "id": "yyy",
      "name": "test",
      "is_admin": false,
      "is_disabled": false,
      "enable_media_playback": true,
      "enable_remote_access": true,
      "enable_all_folders": true
    }
  ]
}
```
