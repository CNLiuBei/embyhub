# VIP购买功能模块文档

## 🎯 功能概述

本模块实现了使用账户余额购买VIP会员的完整功能，包含：
- ✅ 金额使用分（int64）存储，防止浮点精度问题
- ✅ 数据库事务保证ACID特性
- ✅ SELECT FOR UPDATE 防止并发超卖
- ✅ 完整的余额流水记录
- ✅ 清晰的分层架构（Controller/Service/Repository）

---

## 📁 文件结构

```
backend/
├── internal/
│   ├── models/
│   │   └── vip_models.go          # VIP相关数据模型
│   ├── repository/
│   │   └── vip_repository.go      # 数据访问层
│   ├── service/
│   │   └── vip_purchase_service.go # 业务逻辑层
│   └── handler/
│       └── vip_purchase_handler.go # 控制器层
```

---

## 🗄️ 数据库表设计

### 1. VIP套餐表 (vip_plans)
```sql
CREATE TABLE `vip_plans` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(50) NOT NULL COMMENT '套餐名称',
  `price` bigint NOT NULL COMMENT '价格（分）',
  `duration_days` int NOT NULL COMMENT '会员天数',
  `is_active` tinyint(1) DEFAULT 1 COMMENT '是否启用',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='VIP套餐表';

-- 插入测试数据
INSERT INTO `vip_plans` (`name`, `price`, `duration_days`, `is_active`) VALUES
('月度会员', 1500, 30, 1),   -- 15元 = 1500分
('季度会员', 4000, 90, 1),   -- 40元 = 4000分
('年度会员', 12000, 365, 1); -- 120元 = 12000分
```

### 2. 用户会员表 (user_vips)
```sql
CREATE TABLE `user_vips` (
  `user_id` char(36) NOT NULL COMMENT '用户ID（UUID）',
  `vip_expire_at` timestamp NULL COMMENT '会员到期时间',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`user_id`),
  KEY `idx_expire` (`vip_expire_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户会员表';
```

### 3. VIP订单表 (vip_orders)
```sql
CREATE TABLE `vip_orders` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `order_no` varchar(64) NOT NULL COMMENT '订单号',
  `user_id` char(36) NOT NULL COMMENT '用户ID',
  `plan_id` bigint unsigned NOT NULL COMMENT '套餐ID',
  `amount` bigint NOT NULL COMMENT '支付金额（分）',
  `status` varchar(20) NOT NULL COMMENT 'success/failed/pending',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_no` (`order_no`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='VIP订单表';
```

### 4. 余额流水表 (balance_logs)
```sql
CREATE TABLE `balance_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` char(36) NOT NULL COMMENT '用户ID',
  `change_amount` bigint NOT NULL COMMENT '变动金额（分，负数表示扣款）',
  `before_balance` bigint NOT NULL COMMENT '变动前余额（分）',
  `after_balance` bigint NOT NULL COMMENT '变动后余额（分）',
  `type` varchar(50) NOT NULL COMMENT '类型：vip_purchase/recharge/refund',
  `order_no` varchar(64) DEFAULT NULL COMMENT '关联订单号',
  `remark` varchar(255) DEFAULT NULL COMMENT '备注',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_order_no` (`order_no`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='余额流水表';
```

---

## 🔌 路由注册示例

在 `router.go` 中添加以下代码：

```go
package router

import (
	"feiniu-user-system/internal/handler"
	"feiniu-user-system/internal/middleware"
	"feiniu-user-system/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupVipRoutes(r *gin.Engine, db *gorm.DB, jwtManager *auth.JWTManager) {
	// 初始化服务和处理器
	vipService := service.NewVipPurchaseService(db)
	vipHandler := handler.NewVipPurchaseHandler(vipService)

	// VIP相关路由
	vipGroup := r.Group("/api/v1/vip")
	vipGroup.Use(middleware.JWTAuth(jwtManager)) // 需要登录
	{
		vipGroup.GET("/plans", vipHandler.GetVipPlans)           // 获取套餐列表
		vipGroup.GET("/info", vipHandler.GetUserVipInfo)         // 获取用户VIP信息
		vipGroup.POST("/purchase", vipHandler.PurchaseVip)       // 购买VIP
	}
}
```

---

## 📡 API接口文档

### 1. 购买VIP会员

**请求：**
```http
POST /api/v1/vip/purchase
Content-Type: application/json
Authorization: Bearer {token}

{
  "plan_id": 1
}
```

**成功响应：**
```json
{
  "code": 200,
  "message": "购买成功",
  "data": {
    "vip_expire_at": "2026-01-01T00:00:00Z",
    "balance": 8500,
    "order_no": "VIP20250101120000abc123"
  }
}
```

**失败响应：**
```json
{
  "code": 400,
  "message": "余额不足：当前余额 1000 分，需要 1500 分"
}
```

### 2. 获取VIP套餐列表

**请求：**
```http
GET /api/v1/vip/plans
```

**响应：**
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "月度会员",
      "price": 1500,
      "duration_days": 30,
      "is_active": true
    }
  ]
}
```

### 3. 获取用户VIP信息

**请求：**
```http
GET /api/v1/vip/info
Authorization: Bearer {token}
```

**响应：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "vip_expire_at": "2025-12-31T23:59:59Z",
    "is_vip": true
  }
}
```

---

## 🧪 测试脚本

```bash
#!/bin/bash

# 配置
API_BASE="http://localhost:8080/api/v1"
TOKEN="your_jwt_token_here"

# 1. 获取套餐列表
echo "=== 获取VIP套餐列表 ==="
curl -s -X GET "${API_BASE}/vip/plans" | jq

# 2. 购买VIP
echo -e "\n=== 购买VIP会员 ==="
curl -s -X POST "${API_BASE}/vip/purchase" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{"plan_id": 1}' | jq

# 3. 查看VIP信息
echo -e "\n=== 查看VIP信息 ==="
curl -s -X GET "${API_BASE}/vip/info" \
  -H "Authorization: Bearer ${TOKEN}" | jq
```

---

## ⚠️ 重要说明

### 金额处理
- ✅ **所有金额使用分（int64）存储**
- ✅ 1元 = 100分
- ✅ 前端传入时需要转换：15元 → 1500分
- ✅ 前端显示时需要转换：1500分 → 15元

### 并发安全
- ✅ 使用 `SELECT ... FOR UPDATE` 行锁
- ✅ 事务保证ACID特性
- ✅ 防止余额超卖

### 会员时间计算规则
- 未开通 → 从当前时间开始计算
- 已过期 → 从当前时间重新计算
- 未过期 → 从到期时间顺延

---

## 🚀 部署检查清单

- [ ] 数据库表已创建
- [ ] 索引已创建
- [ ] 测试数据已插入
- [ ] 路由已注册
- [ ] JWT认证已配置
- [ ] 余额单位已确认（分 vs 元）
- [ ] 事务隔离级别已确认
- [ ] 并发测试已通过

---

## 📞 联系方式

如有问题，请联系开发团队。
