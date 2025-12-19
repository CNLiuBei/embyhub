# 🚀 前后端联调测试指南

**日期**: 2025-12-07  
**状态**: 准备就绪

---

## ✅ 已完成功能

### 后端 (100%)
- ✅ 卡密生成API
- ✅ 卡密兑换API（带限流）
- ✅ 多格式导出API (CSV/Excel/TXT/Report)
- ✅ 统计查询API
- ✅ 定时任务（自动过期检测）

### 前端 (核心功能完成)
- ✅ 类型定义 (`types/card.types.ts`)
- ✅ API服务 (`services/cardApi.ts`)
- ✅ 工具函数 (`utils/cardHelpers.ts`)
- ✅ 通用组件 (`StatusBadge`, `StatCard`)
- ✅ 用户兑换页面 (`pages/user/RedeemCard.tsx`)
- 📝 管理员页面（使用现有，待更新）

---

## 🏃 快速启动

### 1. 启动后端服务

```bash
# 进入后端目录
cd /vol1/1000/FnHub/feiniu-user-system/backend

# 启动服务
./start.sh

# 或者直接运行
go run cmd/server/main.go
```

**验证后端**:
```bash
curl http://localhost:8080/health
# 应返回: {"status":"ok"}
```

### 2. 启动前端服务

```bash
# 进入前端目录
cd /vol1/1000/FnHub/feiniu-user-system/frontend

# 安装依赖（首次）
npm install

# 启动开发服务器
npm run dev

# 或使用启动脚本
./start.sh
```

**访问地址**:
- 前端: http://localhost:3000
- 后端: http://localhost:8080

---

## 🧪 测试流程

### 测试1: 管理员生成卡密

#### 1.1 登录超级管理员
```
用户名: admin
密码: admin123
```

#### 1.2 进入卡密管理
```
路由: /admin/card
```

#### 1.3 生成测试卡密
- 点击"生成卡密"
- 选择"月卡"
- 数量: 10张
- 点击"确认生成"

**预期结果**:
- ✅ 成功提示
- ✅ 自动下载卡密文件
- ✅ 列表刷新显示新卡密

#### 1.4 测试导出功能
- 找到刚生成的批次
- 点击"导出"按钮
- 尝试不同格式:
  - CSV导出
  - Excel导出
  - 纯文本导出

### 测试2: 用户兑换卡密

#### 2.1 创建测试用户或登录普通用户

#### 2.2 进入兑换页面
```
路由: /user/redeem 或 /redeem
```

#### 2.3 输入卡密兑换
- 复制刚才生成的卡密
- 粘贴到输入框
- 点击"立即兑换"

**预期结果**:
- ✅ 显示成功页面
- ✅ 显示会员到期时间
- ✅ 显示订单号
- ✅ 显示会员权益

#### 2.4 测试限流
- 连续快速兑换5次以上

**预期结果**:
- ❌ 第6次提示"兑换过于频繁"
- ⏱️ 显示剩余等待时间

### 测试3: 导出功能

#### 3.1 导出CSV
```bash
# 使用curl测试
TOKEN="your_admin_token"

curl -X GET "http://localhost:8080/api/v1/admin/card/export/csv?batch_no=B20251207123456" \
  -H "Authorization: Bearer $TOKEN" \
  --output test.csv

# 检查文件
file test.csv
cat test.csv
```

#### 3.2 导出Excel
```bash
curl -X GET "http://localhost:8080/api/v1/admin/card/export/excel?batch_no=B20251207123456" \
  -H "Authorization: Bearer $TOKEN" \
  --output test.xlsx

# 用Excel打开查看
```

#### 3.3 导出纯文本
```bash
curl -X GET "http://localhost:8080/api/v1/admin/card/export/codes?batch_no=B20251207123456" \
  -H "Authorization: Bearer $TOKEN" \
  --output codes.txt

cat codes.txt
```

### 测试4: 定时任务

#### 4.1 查看定时任务日志
```bash
# 实时查看日志
tail -f backend/logs/backend_*.log | grep "定时任务"

# 预期输出（示例）:
# 2025-12-07 10:40:00 ⏰ 执行定时任务: 卡密过期检测
# 2025-12-07 10:40:00 ℹ️  没有需要标记的过期卡密
```

#### 4.2 验证过期检测
1. 创建一个已过期的卡密（修改数据库expire_at）
2. 等待1小时或手动触发
3. 检查卡密状态是否更新为"已过期"

---

## 📊 API测试清单

### 管理员API

| API | 方法 | 路径 | 状态 |
|-----|------|------|------|
| 生成卡密 | POST | `/admin/card/batch` | ✅ |
| 卡密列表 | GET | `/admin/card/list` | ✅ |
| 批次列表 | GET | `/admin/card/batch/list` | ✅ |
| 禁用卡密 | POST | `/admin/card/:id/disable` | ✅ |
| 启用卡密 | POST | `/admin/card/:id/enable` | ✅ |
| 统计数据 | GET | `/admin/card/stats` | ✅ |
| 导出CSV | GET | `/admin/card/export/csv` | ✅ |
| 导出Excel | GET | `/admin/card/export/excel` | ✅ |
| 导出文本 | GET | `/admin/card/export/codes` | ✅ |
| 生成报告 | GET | `/admin/card/export/report` | ✅ |

### 用户API

| API | 方法 | 路径 | 状态 |
|-----|------|------|------|
| 兑换卡密 | POST | `/card/redeem` | ✅ |
| 兑换历史 | GET | `/card/history` | ✅ |

---

## 🐛 常见问题

### 问题1: 前端无法连接后端

**症状**: API请求失败，CORS错误

**解决**:
```bash
# 检查后端是否运行
curl http://localhost:8080/health

# 检查前端代理配置
cat frontend/vite.config.ts

# 确保proxy配置正确
proxy: {
  '/api': {
    target: 'http://localhost:8080',
    changeOrigin: true
  }
}
```

### 问题2: Token验证失败

**症状**: 401 Unauthorized

**解决**:
1. 检查localStorage中的access_token
2. 重新登录获取新token
3. 检查token是否过期

```javascript
// 浏览器控制台
console.log(localStorage.getItem('access_token'))
```

### 问题3: TypeScript类型错误

**症状**: IDE中显示类型错误

**临时解决**:
```bash
# 忽略类型检查运行
npm run dev -- --no-check

# 或在文件中添加
// @ts-nocheck
```

**正确解决**:
- 后续会修复所有类型定义
- 当前专注功能实现

### 问题4: 限流测试不work

**症状**: 可以无限兑换

**检查**:
1. 路由是否添加了限流中间件
```go
card.POST("/redeem", redeemRateLimit, cardHandler.Redeem)
```

2. Redis是否正常运行
```bash
redis-cli PING
# 应返回: PONG
```

---

## 📝 测试检查表

### 后端功能
- [ ] 服务正常启动
- [ ] 健康检查通过
- [ ] 超级管理员可登录
- [ ] 生成卡密成功
- [ ] 导出CSV成功
- [ ] 导出Excel成功
- [ ] 导出文本成功
- [ ] 统计数据正确
- [ ] 定时任务运行
- [ ] 限流生效

### 前端功能
- [ ] 页面正常加载
- [ ] 登录功能正常
- [ ] 卡密管理页面显示
- [ ] 生成卡密表单工作
- [ ] 卡密列表显示
- [ ] 兑换页面显示
- [ ] 兑换功能正常
- [ ] 成功页面显示
- [ ] 兑换历史显示
- [ ] 错误提示友好

### 前后端联调
- [ ] API调用成功
- [ ] 数据格式正确
- [ ] Token认证工作
- [ ] 限流正确触发
- [ ] 错误处理完善
- [ ] 加载状态显示

---

## 🎯 下一步开发

### 优先级1 - 核心功能完善
- [ ] 修复TypeScript类型错误
- [ ] 完善管理员页面使用新API
- [ ] 添加更多导出选项
- [ ] 优化错误提示

### 优先级2 - 用户体验
- [ ] 添加加载动画
- [ ] 优化表单验证
- [ ] 添加数据可视化图表
- [ ] 响应式设计优化

### 优先级3 - 高级功能
- [ ] 批量操作
- [ ] 高级筛选
- [ ] 数据统计报表
- [ ] 实时通知

---

## 💡 开发提示

### 快速测试API
使用Postman或curl快速测试：

```bash
# 获取token
TOKEN=$(curl -X POST http://localhost:8080/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  | jq -r '.data.access_token')

# 使用token测试
curl -X GET http://localhost:8080/api/v1/admin/card/stats \
  -H "Authorization: Bearer $TOKEN"
```

### 快速重置数据
```bash
# 清空所有卡密
cd backend
psql -U fnuser -d feiniu_user -c "TRUNCATE cards CASCADE;"
```

### 快速查看日志
```bash
# 后端日志
tail -f backend/logs/backend_*.log

# 只看错误
tail -f backend/logs/backend_*.log | grep "ERROR"

# 只看卡密相关
tail -f backend/logs/backend_*.log | grep "卡密\|card\|Card"
```

---

## 🎉 成功标准

### 基础功能测试通过
- ✅ 后端所有API正常响应
- ✅ 前端页面正常显示
- ✅ 卡密生成、兑换流程完整
- ✅ 导出功能正常工作
- ✅ 限流机制生效

### 用户体验良好
- ✅ 操作流畅无卡顿
- ✅ 错误提示清晰友好
- ✅ 界面美观易用
- ✅ 响应速度快（< 1秒）

---

**准备就绪！开始测试吧！** 🚀

如有问题，参考:
- 后端文档: `docs/CARD_SYSTEM_COMPLETED.md`
- 快速指南: `QUICK_START.md`
- UI设计: `docs/UI_UX_DESIGN.md`
