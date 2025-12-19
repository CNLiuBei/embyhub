# 🚀 卡密系统 - 快速启动指南

**适用对象**: 开发人员、测试人员、项目经理  
**版本**: v1.0  
**更新日期**: 2025-12-07

---

## 📋 已完成功能

### ✅ 后端功能（已上线）
- [x] 自动过期检测定时任务
- [x] 批次统计自动更新
- [x] 会员过期自动处理
- [x] 多格式卡密导出（CSV/Excel/TXT/报告）
- [x] 防刷限流机制
- [x] IP黑名单功能
- [x] 完整的API接口

### 📋 设计文档（已完成）
- [x] UI/UX设计规范
- [x] 技术架构设计
- [x] 交互流程设计
- [x] 组件库规划

### ⏳ 前端功能（待开发）
- [ ] 管理员卡密管理界面
- [ ] 用户兑换界面
- [ ] 数据可视化图表
- [ ] 响应式适配

---

## 🏃 快速开始

### 1. 启动后端服务

```bash
cd /vol1/1000/FnHub/feiniu-user-system/backend

# 方式1: 使用启动脚本（推荐）
./start.sh

# 方式2: 直接运行
go run cmd/server/main.go
```

**预期输出**:
```
========================================
🕐 定时任务服务启动
========================================
✅ 卡密过期检测将在 1h 后执行
✅ 批次统计更新将在 30m 后执行
✅ 会员过期检测将在明天 02:00 执行
✅ 所有定时任务已启动
服务启动成功 :8080
```

### 2. 验证服务状态

```bash
# 健康检查
curl http://localhost:8080/health

# 预期响应
{"status":"ok"}
```

### 3. 测试超级管理员

**账号信息**:
- 用户名: `admin`
- 密码: `admin123`
- 角色: 超级管理员(3)
- 会员: 长期（100年）

**登录测试**:
```bash
curl -X POST http://localhost:8080/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

---

## 🧪 功能测试

### 1. 卡密生成测试

```bash
# 获取Token
TOKEN="your_jwt_token_here"

# 生成100张月卡
curl -X POST http://localhost:8080/api/v1/admin/card/batch \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "card_type": 1,
    "quantity": 100,
    "duration": 30
  }'
```

**预期响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "batch_no": "B20251207123456",
    "card_type": 1,
    "quantity": 100,
    "duration": 30
  }
}
```

### 2. 卡密导出测试

```bash
# 导出为CSV
curl -X GET "http://localhost:8080/api/v1/admin/card/export/csv?batch_no=B20251207123456" \
  -H "Authorization: Bearer $TOKEN" \
  --output cards.csv

# 导出为Excel
curl -X GET "http://localhost:8080/api/v1/admin/card/export/excel?batch_no=B20251207123456" \
  -H "Authorization: Bearer $TOKEN" \
  --output cards.xlsx

# 导出卡密码（纯文本）
curl -X GET "http://localhost:8080/api/v1/admin/card/export/codes?batch_no=B20251207123456" \
  -H "Authorization: Bearer $TOKEN" \
  --output codes.txt

# 生成使用报告
curl -X GET "http://localhost:8080/api/v1/admin/card/export/report?batch_no=B20251207123456" \
  -H "Authorization: Bearer $TOKEN" \
  --output report.xlsx
```

### 3. 卡密兑换测试

```bash
# 用户兑换卡密（会触发限流）
curl -X POST http://localhost:8080/api/v1/card/redeem \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "ABCD-EFGH-IJKL-MNOP"
  }'
```

**成功响应**:
```json
{
  "code": 200,
  "message": "兑换成功",
  "data": {
    "order_no": "R20251207123456000001",
    "duration": 30,
    "expire_time": "2026-01-07T10:20:30+08:00"
  }
}
```

**限流响应**:
```json
{
  "code": 429,
  "message": "兑换过于频繁，请在3500秒后再试"
}
```

### 4. 限流测试

```bash
# 快速请求5次（第6次应该被限流）
for i in {1..6}; do
  echo "第 $i 次请求:"
  curl -X POST http://localhost:8080/api/v1/card/redeem \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"code\": \"TEST-CODE-$i\"}"
  echo -e "\n---"
done
```

### 5. 定时任务测试

查看定时任务日志:
```bash
# 实时查看定时任务执行
tail -f backend/logs/backend_*.log | grep "定时任务\|卡密过期\|会员过期"

# 查看最近执行记录
grep "定时任务" backend/logs/backend_*.log | tail -20
```

---

## 📊 查看系统状态

### 1. 卡密统计

```bash
curl -X GET http://localhost:8080/api/v1/admin/card/stats \
  -H "Authorization: Bearer $TOKEN"
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total_cards": 1000,
    "unused_cards": 800,
    "used_cards": 180,
    "expired_cards": 20,
    "disabled_cards": 0,
    "total_batches": 10
  }
}
```

### 2. 批次列表

```bash
curl -X GET "http://localhost:8080/api/v1/admin/card/batch/list?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN"
```

### 3. 卡密列表

```bash
curl -X GET "http://localhost:8080/api/v1/admin/card/list?page=1&page_size=10&status=0" \
  -H "Authorization: Bearer $TOKEN"
```

---

## 🗄️ 数据库操作

### 查看数据

```bash
# 连接数据库
PGPASSWORD=fnuser123 psql -U fnuser -d feiniu_user

# 查看卡密统计
SELECT 
  status,
  COUNT(*) as count,
  CASE status
    WHEN 0 THEN '未使用'
    WHEN 1 THEN '已使用'
    WHEN 2 THEN '已过期'
    WHEN 3 THEN '已禁用'
  END as status_name
FROM cards
GROUP BY status;

# 查看批次信息
SELECT 
  batch_no,
  card_type,
  quantity,
  used_count,
  ROUND(used_count::numeric / quantity * 100, 2) as usage_rate
FROM card_batches
ORDER BY created_at DESC
LIMIT 10;

# 查看超级管理员
SELECT username, email, role, member_expire 
FROM users 
WHERE role = 3;
```

### 清理测试数据

```bash
# ⚠️ 危险操作：清空所有卡密数据
PGPASSWORD=fnuser123 psql -U fnuser -d feiniu_user -c "
  TRUNCATE cards CASCADE;
  TRUNCATE card_batches CASCADE;
  TRUNCATE member_orders CASCADE;
"
```

---

## 🔍 常见问题

### Q1: 服务启动失败，提示端口被占用？

**解决方法**:
```bash
# 查找占用端口的进程
lsof -ti:8080

# 杀死进程
lsof -ti:8080 | xargs kill -9

# 或使用停止脚本
./stop-all.sh
```

### Q2: 定时任务没有执行？

**检查步骤**:
1. 查看日志: `grep "定时任务" backend/logs/*.log`
2. 确认服务正常运行
3. 检查系统时间是否正确

### Q3: 导出文件乱码？

**原因**: Excel版本或编码问题

**解决方法**:
- CSV文件：用记事本打开，另存为UTF-8编码
- Excel文件：确保使用Excel 2016+版本

### Q4: 限流如何重置？

**限流自动重置**:
- 用户限流：1小时后自动重置
- IP限流：1小时后自动重置
- 黑名单：根据设置的时长自动解除

**手动重置**:
```bash
# 清空Redis限流数据
redis-cli FLUSHDB
```

### Q5: 如何查看API文档？

**Swagger文档**（如已配置）:
```
http://localhost:8080/swagger/index.html
```

**或查看代码注释**:
- 位置: `backend/internal/handler/*_handler.go`
- 格式: Swagger注释格式

---

## 📝 开发调试

### 启用Debug模式

修改 `config/config.yaml`:
```yaml
log:
  level: debug  # 改为debug
```

### 查看详细日志

```bash
# 实时查看所有日志
tail -f backend/logs/backend_*.log

# 只看错误日志
tail -f backend/logs/backend_*.log | grep "ERROR"

# 只看卡密相关日志
tail -f backend/logs/backend_*.log | grep "卡密\|card"
```

### 性能分析

```bash
# CPU性能分析
go tool pprof http://localhost:8080/debug/pprof/profile

# 内存分析
go tool pprof http://localhost:8080/debug/pprof/heap
```

---

## 🎯 下一步行动

### 立即可做
1. ✅ 测试所有后端API
2. ✅ 验证定时任务执行
3. ✅ 测试限流功能
4. ✅ 测试导出功能

### 准备开发
1. 📋 阅读UI/UX设计文档
2. 📋 阅读技术架构文档
3. 📋 准备前端开发环境
4. 📋 选择UI组件库

### 即将实施
1. ⏳ 实现管理员界面
2. ⏳ 实现用户兑换界面
3. ⏳ 添加数据可视化
4. ⏳ 响应式优化

---

## 📚 相关文档

### 设计文档
- [UI/UX设计方案](docs/UI_UX_DESIGN.md)
- [技术架构设计](docs/TECH_ARCHITECTURE.md)
- [项目总结](docs/PROJECT_SUMMARY.md)

### 功能文档
- [卡密系统完善需求](docs/CARD_SYSTEM_ENHANCEMENT.md)
- [卡密系统完成报告](docs/CARD_SYSTEM_COMPLETED.md)

### 运维文档
- [启动脚本使用说明](docs/SCRIPTS_README.md)

---

## 💡 提示

### 最佳实践
1. **定期备份数据库** - 每天自动备份
2. **监控日志** - 设置日志告警
3. **测试再上线** - 充分测试后部署
4. **文档同步** - 代码变更及时更新文档

### 安全提示
1. **修改默认密码** - admin账号首次登录后修改
2. **定期更新密钥** - JWT密钥定期轮换
3. **限制IP访问** - 生产环境添加IP白名单
4. **启用HTTPS** - 生产环境必须使用HTTPS

---

**🎉 祝你使用愉快！**

如有问题，请查阅相关文档或联系开发团队。

---

**版本**: v1.0  
**作者**: Cascade AI  
**更新**: 2025-12-07
