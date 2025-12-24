# 外部卡密API接口文档

## 概述

外部卡密API允许第三方系统（如闲鱼自动回复系统、发卡平台等）调用本系统获取会员卡密，实现自动发货功能。

## 基础信息

- **Base URL**: `http://your-domain:port/api/external`
- **认证方式**: Bearer Token (API密钥)
- **数据格式**: JSON
- **字符编码**: UTF-8

## 配置说明

### 1. 启用外部API

1. 登录管理后台
2. 进入 **外部API** 菜单
3. 开启 **启用外部API** 开关
4. 点击 **生成新密钥** 获取API密钥
5. 配置IP白名单（可选，留空表示不限制）
6. 保存设置

### 2. 配置项说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| enabled | 是否启用外部API | false |
| api_key | API密钥，用于认证 | 空 |
| allowed_ips | IP白名单，多个IP用逗号分隔 | 空（不限制） |
| rate_limit | 每分钟请求限制 | 60 |
| default_type | 默认卡密类型 | 1（月卡） |
| log_enabled | 是否记录请求日志 | true |

---

## API接口

### 1. 获取卡密

获取一张或多张未使用的会员卡密。支持 GET 和 POST 两种请求方式。

**请求**

```
GET  /api/external/card/fetch
POST /api/external/card/fetch
```

**请求头**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| Authorization | string | 是 | Bearer {API密钥} |
| Content-Type | string | POST时必填 | application/json |

**GET 请求参数（URL参数）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | int | 否 | 卡密类型：1=月卡, 2=季卡, 3=半年卡, 4=年卡。不传则使用默认类型 |
| count | int | 否 | 获取数量，范围1-10，默认1 |

**POST 请求参数（JSON Body）**

```json
{
  "type": 1,
  "count": 1
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | int | 否 | 卡密类型：1=月卡, 2=季卡, 3=半年卡, 4=年卡。不传则使用默认类型 |
| count | int | 否 | 获取数量，范围1-10，默认1 |

**响应 - 单张卡密 (count=1)**

```json
{
  "success": true,
  "message": "获取成功",
  "data": {
    "code": "True-ABCDEFGHJKLMNPQRSTUVWX",
    "type": 1,
    "type_name": "月卡",
    "duration": 30
  }
}
```

**响应 - 多张卡密 (count>1)**

```json
{
  "success": true,
  "message": "获取成功",
  "data": [
    {
      "code": "True-ABCDEFGHJKLMNPQRSTUVWX",
      "type": 1,
      "type_name": "月卡",
      "duration": 30
    },
    {
      "code": "True-ZYXWVUTSRQPNMLKJHGFEDC",
      "type": 1,
      "type_name": "月卡",
      "duration": 30
    }
  ],
  "count": 2
}
```

**响应字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| success | bool | 请求是否成功 |
| message | string | 响应消息 |
| data | object/array | 卡密数据 |
| data.code | string | 卡密码，用户兑换时使用 |
| data.type | int | 卡密类型编号 |
| data.type_name | string | 卡密类型名称 |
| data.duration | int | 有效天数 |
| count | int | 返回的卡密数量（仅批量获取时） |

---

### 2. 查询库存

查询各类型卡密的库存数量。

**请求**

```
GET /api/external/card/stock
```

**请求头**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| Authorization | string | 是 | Bearer {API密钥} |

**响应**

```json
{
  "success": true,
  "message": "获取成功",
  "data": {
    "月卡": 100,
    "季卡": 50,
    "半年卡": 30,
    "年卡": 20
  }
}
```

**响应字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| success | bool | 请求是否成功 |
| message | string | 响应消息 |
| data | object | 各类型卡密库存，key为类型名称，value为数量 |

---

## 错误响应

### 错误码说明

| HTTP状态码 | 说明 |
|------------|------|
| 200 | 请求成功（业务错误也返回200，通过success字段判断） |
| 401 | 认证失败（API未启用、密钥无效等） |
| 403 | IP不在白名单中 |
| 500 | 服务器内部错误 |

### 错误响应示例

**API未启用**
```json
{
  "success": false,
  "message": "外部API未启用"
}
```

**API密钥无效**
```json
{
  "success": false,
  "message": "API密钥无效"
}
```

**IP不在白名单**
```json
{
  "success": false,
  "message": "IP不在白名单中"
}
```

**没有可用卡密**
```json
{
  "success": false,
  "message": "没有可用的卡密"
}
```

---

## 调用示例

### cURL

```bash
# GET方式 - 获取单张月卡
curl -X GET "http://your-domain:54680/api/external/card/fetch?type=1" \
  -H "Authorization: Bearer your_api_key_here"

# GET方式 - 获取3张季卡
curl -X GET "http://your-domain:54680/api/external/card/fetch?type=2&count=3" \
  -H "Authorization: Bearer your_api_key_here"

# POST方式 - 获取单张月卡（推荐）
curl -X POST "http://your-domain:54680/api/external/card/fetch" \
  -H "Authorization: Bearer your_api_key_here" \
  -H "Content-Type: application/json" \
  -d '{"type": 1, "count": 1}'

# POST方式 - 获取3张季卡（推荐）
curl -X POST "http://your-domain:54680/api/external/card/fetch" \
  -H "Authorization: Bearer your_api_key_here" \
  -H "Content-Type: application/json" \
  -d '{"type": 2, "count": 3}'

# 查询库存
curl -X GET "http://your-domain:54680/api/external/card/stock" \
  -H "Authorization: Bearer your_api_key_here"
```

### Python

```python
import requests

API_URL = "http://your-domain:54680/api/external"
API_KEY = "your_api_key_here"

headers = {
    "Authorization": f"Bearer {API_KEY}",
    "Content-Type": "application/json"
}

# POST方式获取卡密（推荐）
def fetch_card_post(card_type=1, count=1):
    response = requests.post(
        f"{API_URL}/card/fetch",
        headers=headers,
        json={"type": card_type, "count": count}
    )
    return response.json()

# GET方式获取卡密
def fetch_card_get(card_type=1, count=1):
    response = requests.get(
        f"{API_URL}/card/fetch",
        headers=headers,
        params={"type": card_type, "count": count}
    )
    return response.json()

# 查询库存
def get_stock():
    response = requests.get(
        f"{API_URL}/card/stock",
        headers=headers
    )
    return response.json()

# 使用示例
if __name__ == "__main__":
    # POST方式获取一张月卡（推荐）
    result = fetch_card_post(card_type=1, count=1)
    if result["success"]:
        print(f"卡密: {result['data']['code']}")
    else:
        print(f"错误: {result['message']}")
    
    # 查询库存
    stock = get_stock()
    if stock["success"]:
        for card_type, count in stock["data"].items():
            print(f"{card_type}: {count}张")
```

### JavaScript/Node.js

```javascript
const axios = require('axios');

const API_URL = 'http://your-domain:54680/api/external';
const API_KEY = 'your_api_key_here';

const headers = {
  'Authorization': `Bearer ${API_KEY}`,
  'Content-Type': 'application/json'
};

// POST方式获取卡密（推荐）
async function fetchCardPost(type = 1, count = 1) {
  try {
    const response = await axios.post(`${API_URL}/card/fetch`, 
      { type, count },
      { headers }
    );
    return response.data;
  } catch (error) {
    return error.response?.data || { success: false, message: error.message };
  }
}

// GET方式获取卡密
async function fetchCardGet(type = 1, count = 1) {
  try {
    const response = await axios.get(`${API_URL}/card/fetch`, {
      headers,
      params: { type, count }
    });
    return response.data;
  } catch (error) {
    return error.response?.data || { success: false, message: error.message };
  }
}

// 查询库存
async function getStock() {
  try {
    const response = await axios.get(`${API_URL}/card/stock`, { headers });
    return response.data;
  } catch (error) {
    return error.response?.data || { success: false, message: error.message };
  }
}

// 使用示例
(async () => {
  // POST方式获取一张月卡（推荐）
  const result = await fetchCardPost(1, 1);
  if (result.success) {
    console.log(`卡密: ${result.data.code}`);
  } else {
    console.log(`错误: ${result.message}`);
  }
  
  // 查询库存
  const stock = await getStock();
  if (stock.success) {
    Object.entries(stock.data).forEach(([type, count]) => {
      console.log(`${type}: ${count}张`);
    });
  }
})();
```

### PHP

```php
<?php

$API_URL = 'http://your-domain:54680/api/external';
$API_KEY = 'your_api_key_here';

// POST方式获取卡密（推荐）
function fetchCardPost($type = 1, $count = 1) {
    global $API_URL, $API_KEY;
    
    $ch = curl_init();
    curl_setopt($ch, CURLOPT_URL, $API_URL . '/card/fetch');
    curl_setopt($ch, CURLOPT_POST, true);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_HTTPHEADER, [
        'Authorization: Bearer ' . $API_KEY,
        'Content-Type: application/json'
    ]);
    curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode([
        'type' => $type,
        'count' => $count
    ]));
    
    $response = curl_exec($ch);
    curl_close($ch);
    
    return json_decode($response, true);
}

// GET方式获取卡密
function fetchCardGet($type = 1, $count = 1) {
    global $API_URL, $API_KEY;
    
    $url = $API_URL . '/card/fetch?' . http_build_query([
        'type' => $type,
        'count' => $count
    ]);
    
    $ch = curl_init();
    curl_setopt($ch, CURLOPT_URL, $url);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_HTTPHEADER, [
        'Authorization: Bearer ' . $API_KEY
    ]);
    
    $response = curl_exec($ch);
    curl_close($ch);
    
    return json_decode($response, true);
}

function getStock() {
    global $API_URL, $API_KEY;
    
    $ch = curl_init();
    curl_setopt($ch, CURLOPT_URL, $API_URL . '/card/stock');
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_HTTPHEADER, [
        'Authorization: Bearer ' . $API_KEY
    ]);
    
    $response = curl_exec($ch);
    curl_close($ch);
    
    return json_decode($response, true);
}

// 使用示例 - POST方式（推荐）
$result = fetchCardPost(1, 1);
if ($result['success']) {
    echo "卡密: " . $result['data']['code'] . "\n";
} else {
    echo "错误: " . $result['message'] . "\n";
}

$stock = getStock();
if ($stock['success']) {
    foreach ($stock['data'] as $type => $count) {
        echo "$type: {$count}张\n";
    }
}
```

---

## 第三方系统对接指南

### 闲鱼自动回复系统对接

1. 在闲鱼自动回复系统中配置API地址：
   ```
   http://your-domain:54680/api/external/card/fetch
   ```

2. 配置认证方式为 Bearer Token，填入API密钥

3. 配置请求参数：
   - `type`: 根据商品类型设置（1=月卡, 2=季卡, 3=半年卡, 4=年卡）
   - `count`: 通常设置为1

4. 配置响应解析：
   - 成功时从 `data.code` 获取卡密
   - 失败时从 `message` 获取错误信息

### 发卡平台对接

大多数发卡平台支持自定义API对接，配置方式类似：

1. API地址：`http://your-domain:54680/api/external/card/fetch`
2. 请求方式：POST（推荐）或 GET
3. 认证头：`Authorization: Bearer {API密钥}`
4. Content-Type：`application/json`（POST方式）
5. 请求体：`{"type": 1, "count": 1}`（POST方式）
6. 卡密字段：`data.code`

---

## 注意事项

1. **卡密不会自动标记为已使用**：获取的卡密仍处于"未使用"状态，只有用户兑换后才会标记为已使用。如需防止重复发放，建议在第三方系统中记录已发放的卡密。

2. **库存管理**：定期检查库存，确保有足够的卡密供发放。可通过库存查询接口监控。

3. **安全建议**：
   - 定期更换API密钥
   - 配置IP白名单限制访问来源
   - 启用请求日志便于审计

4. **频率限制**：默认每分钟60次请求，超出限制会被拒绝。

5. **卡密类型**：
   | 类型值 | 名称 | 默认有效天数 |
   |--------|------|--------------|
   | 1 | 月卡 | 30天 |
   | 2 | 季卡 | 90天 |
   | 3 | 半年卡 | 180天 |
   | 4 | 年卡 | 365天 |

---

## 更新日志

- **v1.1.0** (2024-12-20): 新增POST请求方式支持，推荐使用POST方式调用
- **v1.0.0** (2024-12-20): 初始版本，支持获取卡密和查询库存