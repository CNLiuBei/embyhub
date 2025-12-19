# VIP购买功能 - 完整实现指南

## ✨ 功能特性

### 1. 金融级别安全设计
- ✅ **金额使用分（int64）存储**，完全避免浮点精度问题
- ✅ **数据库事务**保证ACID特性
- ✅ **SELECT FOR UPDATE**行锁防止并发超卖
- ✅ **完整的余额流水**记录每一笔资金变动
- ✅ **幂等性设计**预留订单号唯一索引

### 2. 架构设计
```
Controller (HTTP层)
    ↓
Service (业务逻辑层)
    ↓
Repository (数据访问层)
    ↓
Model (数据模型层)
```

### 3. 核心业务流程

```
1. 用户发起购买请求
   ↓
2. 验证用户身份（JWT）
   ↓
3. 开始数据库事务
   ↓
4. SELECT FOR UPDATE 锁定用户余额
   ↓
5. 校验余额是否充足
   ↓
6. 扣减用户余额
   ↓
7. 记录余额流水
   ↓
8. SELECT FOR UPDATE 锁定用户VIP记录
   ↓
9. 计算新的到期时间
   ↓
10. 更新VIP到期时间
   ↓
11. 创建订单记录
   ↓
12. 提交事务
   ↓
13. 返回成功响应
```

---

## 📦 集成步骤

### Step 1: 创建数据库表

在 MySQL 中执行以下 SQL：

```sql
-- 1. VIP套餐表
CREATE TABLE `vip_plans` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(50) NOT NULL COMMENT '套餐名称',
  `price` bigint NOT NULL COMMENT '价格（分）',
  `duration_days` int NOT NULL COMMENT '会员天数',
  `is_active` tinyint(1) DEFAULT 1 COMMENT '是否启用',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_active` (`is_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='VIP套餐表';

-- 2. 用户会员表
CREATE TABLE `user_vips` (
  `user_id` char(36) NOT NULL COMMENT '用户ID（UUID）',
  `vip_expire_at` timestamp NULL COMMENT '会员到期时间',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`user_id`),
  KEY `idx_expire` (`vip_expire_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户会员表';

-- 3. VIP订单表
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

-- 4. 余额流水表
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
  KEY `idx_type` (`type`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='余额流水表';

-- 插入测试套餐数据
INSERT INTO `vip_plans` (`name`, `price`, `duration_days`, `is_active`) VALUES
('月度会员', 1500, 30, 1),   -- 15元
('季度会员', 4000, 90, 1),   -- 40元
('年度会员', 12000, 365, 1); -- 120元
```

### Step 2: 注册路由

在 `internal/router/router.go` 中添加：

```go
// 在Setup函数中添加
func Setup() *gin.Engine {
	// ... 其他初始化代码

	// VIP服务初始化
	vipService := service.NewVipPurchaseService(database.DB)
	vipHandler := handler.NewVipPurchaseHandler(vipService)

	// VIP路由组
	vipGroup := r.Group("/api/v1/vip")
	vipGroup.Use(middleware.JWTAuth(jwtManager))
	{
		vipGroup.GET("/plans", vipHandler.GetVipPlans)      // 获取套餐列表
		vipGroup.GET("/info", vipHandler.GetUserVipInfo)    // 获取VIP信息
		vipGroup.POST("/purchase", vipHandler.PurchaseVip)  // 购买VIP
	}

	return r
}
```

### Step 3: GORM自动迁移（可选）

在 `internal/database/postgres.go` 的 `autoMigrate` 函数中添加：

```go
func autoMigrate() error {
	return DB.AutoMigrate(
		&models.User{},
		// ... 其他模型
		&models.VipPlan{},
		&models.UserVip{},
		&models.VipOrder{},
		&models.BalanceLog{},
	)
}
```

---

## 🧪 测试方法

### 1. 准备测试数据

```sql
-- 创建测试用户（假设users表已存在）
-- 注意：余额字段假设使用float存储元，实际查询时会转换为分
UPDATE users SET balance = 100.00 WHERE username = 'test_user'; -- 100元 = 10000分
```

### 2. 使用cURL测试

```bash
# 1. 登录获取token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"account":"test_user","password":"password"}' | jq -r '.data.access_token')

# 2. 查看套餐列表
curl -X GET "http://localhost:8080/api/v1/vip/plans"

# 3. 查看当前VIP状态
curl -X GET "http://localhost:8080/api/v1/vip/info" \
  -H "Authorization: Bearer $TOKEN"

# 4. 购买月度会员
curl -X POST "http://localhost:8080/api/v1/vip/purchase" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"plan_id": 1}'

# 5. 再次查看VIP状态
curl -X GET "http://localhost:8080/api/v1/vip/info" \
  -H "Authorization: Bearer $TOKEN"
```

### 3. 验证数据

```sql
-- 查看订单
SELECT * FROM vip_orders WHERE user_id = 'your_user_uuid' ORDER BY created_at DESC LIMIT 5;

-- 查看余额流水
SELECT * FROM balance_logs WHERE user_id = 'your_user_uuid' ORDER BY created_at DESC LIMIT 5;

-- 查看VIP状态
SELECT * FROM user_vips WHERE user_id = 'your_user_uuid';

-- 查看用户余额
SELECT balance FROM users WHERE id = 'your_user_uuid';
```

---

## ⚠️ 重要注意事项

### 1. 金额单位转换

**数据库存储（假设users.balance是float元）→ 业务逻辑（int64分）**

```go
// Repository层的转换
balanceInCents := int64(user.Balance * 100)  // 元 → 分

newBalanceInYuan := float64(newBalanceInCents) / 100.0  // 分 → 元
```

**前端显示**
```javascript
// 分 → 元
const yuan = cents / 100;

// 元 → 分
const cents = yuan * 100;
```

### 2. 时区处理

所有时间统一使用 UTC：

```go
time.Now().UTC()
```

前端显示时转换为本地时区。

### 3. 并发安全

**必须使用行锁：**
```go
tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user).Error
```

### 4. 错误处理

服务层返回的所有错误都会被控制器捕获并转换为HTTP响应。

---

## 🔐 安全检查清单

- [ ] JWT认证已启用
- [ ] 余额校验在事务内完成
- [ ] 使用SELECT FOR UPDATE防止并发
- [ ] 订单号唯一索引已创建
- [ ] 所有金额使用int64（分）
- [ ] 事务隔离级别 >= READ COMMITTED
- [ ] 敏感操作已记录日志
- [ ] API限流已配置
- [ ] 输入参数已验证

---

## 📊 性能优化建议

1. **索引优化**
   - `vip_orders.order_no` 唯一索引
   - `balance_logs.user_id` 索引
   - `user_vips.vip_expire_at` 索引

2. **缓存策略**
   - VIP套餐列表可缓存（Redis）
   - 用户VIP状态可缓存（5分钟TTL）

3. **数据库连接池**
   ```go
   sqlDB.SetMaxIdleConns(10)
   sqlDB.SetMaxOpenConns(100)
   sqlDB.SetConnMaxLifetime(time.Hour)
   ```

---

## 🐛 常见问题

### Q1: 余额扣除了但VIP没开通？
**A:** 检查事务是否正确提交，查看 `vip_orders.status` 字段。

### Q2: 并发购买导致余额超扣？
**A:** 确认使用了 `SELECT FOR UPDATE` 行锁。

### Q3: 时间计算不正确？
**A:** 检查是否统一使用 UTC 时区。

### Q4: 金额显示错误？
**A:** 确认前后端都正确进行了分↔元转换。

---

## 📚 扩展功能建议

1. **幂等性增强**
   - 添加请求ID字段
   - 实现重复请求检测

2. **退款功能**
   - 实现VIP退订
   - 余额返还

3. **优惠券系统**
   - 折扣码
   - 满减活动

4. **分级会员**
   - 普通会员
   - 高级会员
   - 超级会员

---

## 📞 技术支持

如有问题，请提交Issue或联系开发团队。

**祝您使用愉快！** 🎉
