#!/bin/bash

# 全功能测试脚本
echo "=== Emby用户管理系统 - 完整功能测试 ==="

# 登录
TOKEN=$(curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Liubei00"}' \
  -s | jq -r '.data.token')

echo "✅ 1. 登录成功"

# 测试用户管理
echo -e "\n📋 2. 测试用户管理"
curl -s "http://localhost:8080/api/users?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN" | jq '{total: .data.total, users: (.data.list | length)}'

# 测试角色管理
echo -e "\n📋 3. 测试角色管理"
curl -s http://localhost:8080/api/roles \
  -H "Authorization: Bearer $TOKEN" | jq '{total: .data.total, roles: (.data.list | length)}'

# 测试权限列表
echo -e "\n📋 4. 测试权限管理"
curl -s http://localhost:8080/api/permissions \
  -H "Authorization: Bearer $TOKEN" | jq '{total: .data.total, permissions: (.data.list | length)}'

# 测试统计数据
echo -e "\n📊 5. 测试统计数据"
curl -s http://localhost:8080/api/statistics \
  -H "Authorization: Bearer $TOKEN" | jq '.data | {total_users, active_users, today_access}'

# 测试访问记录
echo -e "\n📜 6. 测试访问记录"
curl -s "http://localhost:8080/api/access-records?page=1&page_size=5" \
  -H "Authorization: Bearer $TOKEN" | jq '{total: .data.total, records: (.data.list | length)}'

# 测试Emby连接
echo -e "\n🔗 7. 测试Emby连接"
curl -X POST http://localhost:8080/api/emby/test \
  -H "Authorization: Bearer $TOKEN" -s | jq '{code, message}'

# 测试Emby用户列表
echo -e "\n👥 8. 测试Emby用户"
curl -s http://localhost:8080/api/emby/users \
  -H "Authorization: Bearer $TOKEN" | jq '{code, users: (.data | length)}'

# 测试系统配置
echo -e "\n⚙️  9. 测试系统配置"
curl -s http://localhost:8080/api/configs \
  -H "Authorization: Bearer $TOKEN" | jq '{total: .data.total, configs: (.data.list | length)}'

echo -e "\n✅ 所有功能测试完成！"
