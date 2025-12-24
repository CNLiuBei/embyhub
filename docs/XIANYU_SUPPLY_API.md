# 闲鱼货源系统 API 文档

## 概述

本文档描述了闲鱼货源系统的所有API端点，用于集成测试和验证。

## 基础信息

- 基础路径: `/api/v1/admin/supply`
- 认证方式: JWT Bearer Token
- 权限要求: 管理员 (role >= 2)

## API 端点列表

### 1. 配置管理

#### 获取配置
```
GET /api/v1/admin/supply/config
```

响应示例:
```json
{
  "code": 0,
  "data": {
    "configured": true,
    "app_key": 203413189371893,
    "base_url": "https://open.goofish.pro",
    "webhook_url": "https://your-domain.com/api/webhook/xianyu",
    "is_active": true
  }
}
```

#### 保存配置
```
POST /api/v1/admin/supply/config
Content-Type: application/json

{
  "app_key": 203413189371893,
  "app_secret": "your_app_secret",
  "base_url": "https://open.goofish.pro",
  "webhook_url": "https://your-domain.com/api/webhook/xianyu"
}
```

#### 测试连接
```
POST /api/v1/admin/supply/config/test
```

### 2. 店铺管理

#### 获取店铺列表
```
GET /api/v1/admin/supply/shops?refresh=false
```

#### 刷新店铺数据
```
POST /api/v1/admin/supply/shops/refresh
```

### 3. 商品类目和属性

#### 获取类目列表
```
GET /api/v1/admin/supply/categories?item_biz_type=2&sp_biz_type=1
```

#### 获取属性列表
```
GET /api/v1/admin/supply/properties?item_biz_type=2&sp_biz_type=1&channel_cat_id=xxx
```

### 4. 商品管理

#### 获取商品列表
```
GET /api/v1/admin/supply/products?page_no=1&page_size=20&product_status=22
```

#### 获取商品详情
```
GET /api/v1/admin/supply/products/{product_id}
```

#### 获取商品规格
```
GET /api/v1/admin/supply/products/{product_id}/skus
```

#### 创建商品
```
POST /api/v1/admin/supply/products
Content-Type: application/json

{
  "user_name": "闲鱼会员名",
  "province": 110000,
  "city": 110100,
  "district": 110101,
  "title": "商品标题",
  "content": "商品描述",
  "images": ["https://example.com/image1.jpg"],
  "item_biz_type": 2,
  "sp_biz_type": 1,
  "channel_cat_id": "xxx",
  "original_price": 10000,
  "price": 9900,
  "stock": 100
}
```

#### 编辑商品
```
PUT /api/v1/admin/supply/products/{product_id}
Content-Type: application/json

{
  "price": 8800,
  "stock": 50
}
```

#### 上架商品
```
POST /api/v1/admin/supply/products/{product_id}/online
```

#### 下架商品
```
POST /api/v1/admin/supply/products/{product_id}/offline
```

#### 更新库存
```
PUT /api/v1/admin/supply/products/{product_id}/stock
Content-Type: application/json

{
  "stock": 200
}
```

#### 删除商品
```
DELETE /api/v1/admin/supply/products/{product_id}
```

### 5. 订单管理

#### 获取订单列表
```
GET /api/v1/admin/supply/orders?page_no=1&page_size=20&order_status=12
```

#### 获取订单详情
```
GET /api/v1/admin/supply/orders/{order_no}
```

#### 获取订单卡密
```
GET /api/v1/admin/supply/orders/{order_no}/cards
```

#### 订单发货
```
POST /api/v1/admin/supply/orders/{order_no}/ship
Content-Type: application/json

{
  "waybill_no": "SF1234567890",
  "express_code": "SF",
  "express_name": "顺丰速运"
}
```

#### 修改订单价格
```
PUT /api/v1/admin/supply/orders/{order_no}/price
Content-Type: application/json

{
  "order_price": 9900,
  "express_fee": 0
}
```

### 6. 快递公司

#### 获取快递公司列表
```
GET /api/v1/admin/supply/express
```

### 7. 缓存管理

#### 清除缓存
```
POST /api/v1/admin/supply/cache/clear?type=category
```

type 可选值: `category`, `property`, 不传则清除所有缓存

### 8. 日志查询

#### 获取API调用日志
```
GET /api/v1/admin/supply/logs/api?page=1&page_size=20
```

#### 获取Webhook日志
```
GET /api/v1/admin/supply/logs/webhook?page=1&page_size=20
```

## Webhook 端点

### 商品回调通知
```
POST /api/webhook/xianyu/product/callback
```

### 商品推送通知
```
POST /api/webhook/xianyu/product/push
```

### 订单推送通知
```
POST /api/webhook/xianyu/order
```

## 枚举值说明

### 商品类型 (item_biz_type)
- 2: 普通商品
- 0: 已验货
- 10: 验货宝
- 16: 品牌授权
- 19: 闲鱼严选
- 24: 闲鱼特卖
- 26: 品牌捡漏

### 商品状态 (product_status)
- -1: 已删除
- 21: 待发布
- 22: 销售中
- 23: 已售罄
- 31: 手动下架
- 33: 售出下架
- 36: 自动下架

### 订单状态 (order_status)
- 11: 待付款
- 12: 待发货
- 21: 已发货
- 22: 交易成功
- 23: 已退款
- 24: 交易关闭

### 退款状态 (refund_status)
- 0: 未申请退款
- 1: 待商家处理
- 2: 待买家退货
- 3: 待商家收货
- 4: 退款关闭
- 5: 退款成功
- 6: 已拒绝退款
- 8: 待确认退货地址

## 测试步骤

### 1. 验证API配置功能
1. 访问配置页面 `/admin/supply/settings`
2. 输入闲管家开放平台的 AppKey 和 AppSecret
3. 点击"测试连接"验证配置是否正确
4. 保存配置

### 2. 验证店铺管理功能
1. 访问店铺页面 `/admin/supply/shops`
2. 点击"从API刷新"同步店铺数据
3. 验证店铺列表显示正确
4. 检查授权过期提醒

### 3. 验证商品管理功能
1. 访问商品页面 `/admin/supply/products`
2. 点击"发布商品"创建新商品
3. 选择类目、填写商品信息
4. 验证商品创建成功
5. 测试上架/下架/删除功能

### 4. 验证订单管理功能
1. 访问订单页面 `/admin/supply/orders`
2. 查看订单列表
3. 点击订单查看详情
4. 测试发货功能（需要真实订单）
