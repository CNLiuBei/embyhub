# API接口文档

## 接口概述

### Base URL

```
http://localhost:8080/api/v1
```

### 认证方式

使用JWT Bearer Token认证：

```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### 标准响应格式

```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

### 状态码说明

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权（需要登录） |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

## 用户接口

### 1. 发送注册验证码

发送邮箱验证码用于注册。

**接口地址**: `POST /user/send-register-code`

**请求参数**:

```json
{
  "email": "user@example.com"
}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "验证码已发送",
  "data": null
}
```

**限流规则**: 同一邮箱60秒内只能发送一次

---

### 2. 用户注册

使用邮箱验证码注册新用户。

**接口地址**: `POST /user/register`

**请求参数**:

```json
{
  "username": "testuser",
  "email": "user@example.com",
  "password": "Password123",
  "code": "123456"
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名，3-20个字符 |
| email | string | 是 | 邮箱地址 |
| password | string | 是 | 密码，至少6个字符 |
| code | string | 是 | 邮箱验证码 |

**响应示例**:

```json
{
  "code": 200,
  "message": "注册成功",
  "data": {
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "username": "testuser",
      "email": "user@example.com",
      "role": 0,
      "member_level": 0,
      "created_at": "2025-12-07T00:00:00Z"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

---

### 3. 用户登录

使用用户名/邮箱和密码登录。

**接口地址**: `POST /user/login`

**请求参数**:

```json
{
  "account": "testuser",
  "password": "Password123"
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| account | string | 是 | 用户名或邮箱 |
| password | string | 是 | 密码 |

**响应示例**:

```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "username": "testuser",
      "email": "user@example.com",
      "role": 0,
      "member_level": 1,
      "member_expire": "2026-01-07T00:00:00Z"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**限流规则**: 5次/分钟

---

### 4. 刷新Token

使用Refresh Token刷新Access Token。

**接口地址**: `POST /user/refresh-token`

**请求参数**:

```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "刷新成功",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

---

### 5. 获取用户信息

获取当前登录用户的详细信息。

**接口地址**: `GET /user/profile`

**请求头**: `Authorization: Bearer <token>`

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "testuser",
    "email": "user@example.com",
    "feiniu_username": "testuser",
    "role": 0,
    "member_level": 1,
    "member_expire": "2026-01-07T00:00:00Z",
    "last_login_at": "2025-12-07T00:00:00Z",
    "created_at": "2025-12-06T00:00:00Z"
  }
}
```

---

### 6. 登出

用户登出，清除会话。

**接口地址**: `POST /user/logout`

**请求头**: `Authorization: Bearer <token>`

**响应示例**:

```json
{
  "code": 200,
  "message": "登出成功",
  "data": null
}
```

---

## 会员接口

### 1. 获取会员信息

获取当前用户的会员详细信息。

**接口地址**: `GET /member/info`

**请求头**: `Authorization: Bearer <token>`

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "level": 1,
    "level_name": "银卡会员",
    "expire_at": "2026-01-07T00:00:00Z",
    "days_left": 31,
    "watch_limit": 10,
    "ad_free": true,
    "quality_4k": false
  }
}
```

---

### 2. 获取会员订单

获取会员充值/兑换历史记录。

**接口地址**: `GET /member/orders?page=1&page_size=10`

**请求头**: `Authorization: Bearer <token>`

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页数量，默认10 |

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "id": "order_001",
        "type": "card",
        "level": 1,
        "days": 30,
        "amount": 0,
        "status": "completed",
        "created_at": "2025-12-07T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10
  }
}
```

---

## 媒体库接口

### 1. 获取媒体库列表

获取用户可访问的所有媒体库。

**接口地址**: `GET /media/list`

**请求头**: `Authorization: Bearer <token>`

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "guid": "8fe47bc140884ede9d85f4324627f726",
      "title": "电视剧",
      "posters": [
        "/a8/13/RXFg9YOlYYTNwMynBkZifbn3VpVnzd401lk1CjS099E0CKLry1GZYopZ6ZWjQxxtHiUJb00oZMEWAjnXSGjRS4UMnX.webp"
      ],
      "category": "TV",
      "image_url": "/api/v1/image"
    },
    {
      "guid": "7124a7d33516491cb95b842238cb2e3b",
      "title": "动画电影",
      "posters": [
        "/b3/20/RXFg9YOlYYTNwMynBkZifbn3VpVnzd401lk1CjS099E0CKLrx61vRRSKke5Poyuvwqb8J1kzQrY41rhyMShInSPY4F.webp"
      ],
      "category": "Movie",
      "image_url": "/api/v1/image"
    }
  ]
}
```

---

### 2. 获取媒体库内容

获取指定媒体库中的影片列表。

**接口地址**: `GET /media/db/:guid/items`

**请求头**: `Authorization: Bearer <token>`

**路径参数**:

| 参数 | 类型 | 说明 |
|------|------|------|
| guid | string | 媒体库GUID |

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页数量，默认20 |

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "guid": "c8ad7e5ae16f4fd0b4d3fcbaaad11252",
        "title": "藏海传",
        "type": "TV",
        "poster": "/a8/13/RXFg9YOlYYTNwMynBkZifbn3VpVnzd401lk1CjS099E0CKLry1GZYopZ6ZWjQxxtHiUJb00oZMEWAjnXSGjRS4UMnX.webp",
        "vote_average": "7.464",
        "release_date": "2025-05-18",
        "overview": "《藏海传》讲述钦天监监正蒯铎之子身负血海深仇...",
        "number_of_episodes": 40,
        "local_number_of_episodes": 5
      }
    ],
    "total": 4,
    "page": 1,
    "page_size": 20,
    "image_url": "/api/v1/image"
  }
}
```

---

### 3. 搜索媒体

在所有媒体库中搜索影片。

**接口地址**: `GET /media/search`

**请求头**: `Authorization: Bearer <token>`

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| keyword | string | 是 | 搜索关键词 |
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页数量，默认20 |

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "guid": "xxx",
        "title": "命悬一生",
        "type": "TV",
        "poster": "/path/to/poster.webp",
        "vote_average": "8.8"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20,
    "image_url": "/api/v1/image"
  }
}
```

---

### 4. 图片代理

代理获取飞牛影视的图片资源。

**接口地址**: `GET /image/*path`

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| uid | string | 是 | 用户ID |

**示例URL**:

```
GET /api/v1/image/a8/13/RXFg9YOlYYTNwMynBkZifbn3VpVnzd401lk1CjS099E0CKLry1GZYopZ6ZWjQxxtHiUJb00oZMEWAjnXSGjRS4UMnX.webp?uid=550e8400-e29b-41d4-a716-446655440000
```

**响应**: 直接返回图片二进制数据

**Content-Type**: `image/webp` 或 `image/jpeg`

---

## 卡密接口

### 1. 兑换卡密

使用卡密兑换会员权益。

**接口地址**: `POST /card/redeem`

**请求头**: `Authorization: Bearer <token>`

**请求参数**:

```json
{
  "card_code": "XXXX-XXXX-XXXX-XXXX"
}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "兑换成功",
  "data": {
    "level": 1,
    "days": 30,
    "expire_at": "2026-01-07T00:00:00Z"
  }
}
```

---

### 2. 兑换历史

查看卡密兑换历史记录。

**接口地址**: `GET /card/history`

**请求头**: `Authorization: Bearer <token>`

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页数量，默认10 |

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "id": "history_001",
        "card_code": "XXXX-****-****-XXXX",
        "level": 1,
        "days": 30,
        "redeemed_at": "2025-12-07T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10
  }
}
```

---

## 管理接口

> 需要管理员权限 (role >= 1)

### 1. 获取用户列表

获取所有用户列表（分页）。

**接口地址**: `GET /admin/users`

**请求头**: `Authorization: Bearer <admin_token>`

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页数量，默认20 |
| keyword | string | 否 | 搜索关键词 |
| role | int | 否 | 角色筛选 |
| member_level | int | 否 | 会员等级筛选 |

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "username": "testuser",
        "email": "user@example.com",
        "role": 0,
        "member_level": 1,
        "member_expire": "2026-01-07T00:00:00Z",
        "last_login_at": "2025-12-07T00:00:00Z",
        "created_at": "2025-12-06T00:00:00Z"
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 2. 获取用户详情

获取指定用户的详细信息。

**接口地址**: `GET /admin/users/:id`

**请求头**: `Authorization: Bearer <admin_token>`

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "testuser",
    "email": "user@example.com",
    "feiniu_username": "testuser",
    "role": 0,
    "member_level": 1,
    "member_expire": "2026-01-07T00:00:00Z",
    "last_login_at": "2025-12-07T00:00:00Z",
    "last_login_ip": "127.0.0.1",
    "created_at": "2025-12-06T00:00:00Z",
    "updated_at": "2025-12-07T00:00:00Z"
  }
}
```

---

### 3. 更新用户信息

更新用户的会员等级、角色等信息。

**接口地址**: `PUT /admin/users/:id`

**请求头**: `Authorization: Bearer <admin_token>`

**请求参数**:

```json
{
  "role": 1,
  "member_level": 2,
  "member_expire": "2026-12-31T23:59:59Z"
}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "更新成功",
  "data": null
}
```

---

### 4. 生成卡密

批量生成会员卡密。

**接口地址**: `POST /admin/cards/generate`

**请求头**: `Authorization: Bearer <admin_token>`

**请求参数**:

```json
{
  "type": "monthly",
  "level": 1,
  "days": 30,
  "count": 10
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | 是 | 类型: monthly/yearly/permanent |
| level | int | 是 | 会员等级: 1-3 |
| days | int | 是 | 有效天数 |
| count | int | 是 | 生成数量，最多100 |

**响应示例**:

```json
{
  "code": 200,
  "message": "生成成功",
  "data": {
    "cards": [
      {
        "code": "ABCD-1234-EFGH-5678",
        "type": "monthly",
        "level": 1,
        "days": 30
      }
    ],
    "count": 10
  }
}
```

---

### 5. 获取卡密列表

查询所有卡密及使用状态。

**接口地址**: `GET /admin/cards`

**请求头**: `Authorization: Bearer <admin_token>`

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |
| status | string | 否 | unused/used/expired |
| type | string | 否 | 卡密类型 |

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "id": "card_001",
        "code": "ABCD-****-****-5678",
        "type": "monthly",
        "level": 1,
        "days": 30,
        "status": "unused",
        "created_at": "2025-12-07T00:00:00Z"
      }
    ],
    "total": 50,
    "page": 1,
    "page_size": 20
  }
}
```

---

## 错误码说明

### 业务错误码

| 错误码 | 说明 |
|--------|------|
| 1001 | 用户名已存在 |
| 1002 | 邮箱已被注册 |
| 1003 | 验证码错误 |
| 1004 | 验证码已过期 |
| 1005 | 账号或密码错误 |
| 1006 | 账号已被禁用 |
| 2001 | 会员已过期 |
| 2002 | 观看次数已用完 |
| 3001 | 卡密不存在 |
| 3002 | 卡密已使用 |
| 3003 | 卡密已过期 |
| 4001 | 权限不足 |

---

**最后更新**: 2025年12月
**文档版本**: v1.0
