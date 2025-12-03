#!/bin/bash

# 初始化测试数据脚本

echo "=== Emby用户管理系统 - 初始化测试数据 ==="

# 获取Token
echo "1. 登录系统..."
TOKEN=$(curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Liubei00"}' \
  -s | jq -r '.data.token')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo "❌ 登录失败"
  exit 1
fi

echo "✅ 登录成功"

# 创建测试用户
echo -e "\n2. 创建测试用户..."
curl -X POST http://localhost:8080/api/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "test_user",
    "password": "test123",
    "email": "test@example.com",
    "role_id": 2
  }' -s | jq '{code, message}'

# 创建访问记录
echo -e "\n3. 创建访问记录..."
for i in {1..10}; do
  curl -X POST http://localhost:8080/api/access-records \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"user_id\": 1,
      \"resource\": \"测试资源-$i\",
      \"ip_address\": \"127.0.0.1\",
      \"device_info\": \"Chrome/Linux\"
    }" -s > /dev/null
done
echo "✅ 创建了10条访问记录"

# 获取统计数据
echo -e "\n4. 获取统计数据..."
curl -s http://localhost:8080/api/statistics \
  -H "Authorization: Bearer $TOKEN" | jq '{
    total_users: .data.total_users,
    active_users: .data.active_users,
    today_access: .data.today_access,
    top_users_count: (.data.top_users | length),
    trend_days: (.data.access_trend | length)
  }'

echo -e "\n✅ 测试数据初始化完成！"
