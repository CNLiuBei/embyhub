# VIP购买功能 - 前端使用文档

## 📦 已创建的文件

### 1. 核心页面
- ✅ `src/pages/user/VipPurchase.tsx` - VIP购买主页面

### 2. 已更新的文件
- ✅ `src/App.tsx` - 添加路由配置
- ✅ `src/layouts/UserLayout.tsx` - 添加侧边栏菜单

---

## 🎨 页面功能

### 1. VIP状态展示
- 显示当前VIP状态（是否为会员）
- 显示会员到期时间
- 显示账户余额

### 2. VIP权益展示
- 无广告体验
- 高清画质（1080P）
- 抢先观看新片
- 专属客服

### 3. VIP套餐列表
- 月度会员（30天）- ¥15
- 季度会员（90天）- ¥40 【最受欢迎】
- 年度会员（365天）- ¥120

### 4. 购买流程
1. 用户点击"立即开通"
2. 弹出确认弹窗
3. 显示套餐详情和支付后余额
4. 确认购买
5. 扣除余额，开通/续期VIP
6. 显示成功提示

---

## 🔌 API接口对接

### 后端接口要求

```typescript
// 1. 获取VIP套餐列表
GET /api/v1/vip/plans
Response: {
  code: 200,
  data: [
    {
      id: 1,
      name: "月度会员",
      price: 1500,        // 单位：分
      original: 2000,     // 原价（分）
      duration_days: 30,
      is_active: true,
      discount: "限时优惠"
    }
  ]
}

// 2. 获取用户VIP信息
GET /api/v1/vip/info
Authorization: Bearer {token}
Response: {
  code: 200,
  data: {
    user_id: "uuid",
    vip_expire_at: "2025-12-31T23:59:59Z",
    is_vip: true
  }
}

// 3. 购买VIP
POST /api/v1/vip/purchase
Authorization: Bearer {token}
Body: {
  plan_id: 1
}
Response: {
  code: 200,
  message: "购买成功",
  data: {
    vip_expire_at: "2026-01-31T23:59:59Z",
    balance: 8500,      // 剩余余额（分）
    order_no: "VIP20250101120000abc123"
  }
}

// 4. 获取用户余额（已有接口）
GET /api/v1/balance
Authorization: Bearer {token}
Response: {
  code: 200,
  data: {
    balance: 100.00     // 单位：元
  }
}
```

---

## 💰 金额处理规则

### 前端显示
```typescript
// 后端返回的是分（int64），前端需要转换为元显示
const yuan = cents / 100;

// 示例
price: 1500 分 → 显示为 ¥15.00
balance: 8500 分 → 显示为 ¥85.00
```

### 数据转换
```typescript
// VipPlan接口
interface VipPlan {
  price: number       // 单位：分
  original: number    // 单位：分
}

// 显示时
<span>¥{(plan.price / 100).toFixed(2)}</span>

// UserVip接口
interface UserVip {
  vip_expire_at: string | null  // ISO 8601格式
  is_vip: boolean
}

// 时间显示
dayjs(vipExpireAt).format('YYYY-MM-DD HH:mm')

// 余额数据（从balance接口）
interface Balance {
  balance: number     // 单位：元
}
```

---

## 🎯 页面路由

### 访问地址
```
http://localhost:3000/user/vip-purchase
```

### 路由配置
```tsx
// App.tsx
<Route path="vip-purchase" element={<VipPurchase />} />
```

### 侧边栏菜单
```tsx
// UserLayout.tsx
{ key: '/user/vip-purchase', icon: <CrownOutlined />, label: 'VIP购买' }
```

---

## 🎨 UI设计特点

### 1. 渐变背景
- VIP状态卡片：黄-橙渐变
- 最受欢迎套餐：金色边框 + 阴影
- 购买按钮：渐变色

### 2. 视觉层次
- 使用Ant Design的Card、Badge、Tag组件
- 响应式布局（Col xs/sm/lg）
- 图标突出显示（CrownOutlined）

### 3. 交互反馈
- Loading状态
- 成功/失败提示
- 余额不足警告
- 续期提示

---

## 🧪 本地测试步骤

### 1. 启动后端
```bash
cd backend
./backend
```

### 2. 启动前端
```bash
cd frontend
npm run dev
```

### 3. 测试流程
1. 登录账号（http://localhost:3000/login）
2. 点击侧边栏"VIP购买"
3. 查看套餐列表
4. 点击"立即开通"
5. 确认购买弹窗
6. 点击"确认支付"
7. 查看成功提示
8. 验证VIP状态已更新

---

## ⚠️ 注意事项

### 1. 金额单位
- **后端存储和传输：分（int64）**
- **前端显示：元（number，保留2位小数）**
- 转换公式：`元 = 分 / 100`

### 2. 时间格式
- 后端返回：ISO 8601格式（UTC）
- 前端显示：`YYYY-MM-DD HH:mm` 本地时区
- 使用dayjs库处理

### 3. 错误处理
```typescript
onError: (error: any) => {
  message.error(error.response?.data?.message || '购买失败，请稍后重试')
}
```

### 4. 数据刷新
```typescript
// 购买成功后刷新
queryClient.invalidateQueries({ queryKey: ['vipInfo'] })
queryClient.invalidateQueries({ queryKey: ['balance'] })
```

---

## 🚀 功能扩展建议

### 1. 支付方式
- [ ] 支付宝支付
- [ ] 微信支付
- [ ] 银行卡支付

### 2. 优惠券
- [ ] 优惠券列表
- [ ] 优惠券选择
- [ ] 折扣计算

### 3. 订单历史
- [ ] VIP购买记录
- [ ] 订单详情查看
- [ ] 订单状态追踪

### 4. 自动续费
- [ ] 开通自动续费
- [ ] 续费管理
- [ ] 余额提醒

---

## 📱 响应式设计

```tsx
<Col xs={24} sm={12} lg={8}>
  {/* 套餐卡片 */}
</Col>
```

- **xs (< 576px)**: 单列显示
- **sm (≥ 576px)**: 两列显示
- **lg (≥ 992px)**: 三列显示

---

## 🎨 主题适配

页面使用Ant Design组件，自动适配主题：
- 默认主题
- 暗色主题
- 粉色主题

通过 `ThemeContext` 统一管理。

---

## 📞 技术支持

如有问题，请查看：
1. 浏览器控制台（Console）
2. Network标签页（查看API请求）
3. React Query Devtools（查看缓存状态）

---

**🎉 前端页面已完成，可直接使用！**
