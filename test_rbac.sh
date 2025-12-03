#!/bin/bash

echo "=== RBAC功能测试 ==="

# 1. 使用超级管理员登录
echo -e "\n1. 超级管理员登录（admin）"
ADMIN_TOKEN=$(curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Liubei00"}' \
  -s | jq -r '.data.token')

if [ -n "$ADMIN_TOKEN" ] && [ "$ADMIN_TOKEN" != "null" ]; then
  echo "✅ 超级管理员登录成功"
  
  # 获取用户信息
  curl -s http://localhost:8080/api/auth/current \
    -H "Authorization: Bearer $ADMIN_TOKEN" | jq '{
      username: .data.username,
      role: .data.role.role_name,
      permission_count: (.data.role.permissions | length)
    }'
else
  echo "❌ 登录失败"
fi

# 2. 使用新注册的普通用户登录
echo -e "\n2. 普通用户登录（newuser）"
USER_TOKEN=$(curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"newuser","password":"newpass123"}' \
  -s | jq -r '.data.token')

if [ -n "$USER_TOKEN" ] && [ "$USER_TOKEN" != "null" ]; then
  echo "✅ 普通用户登录成功"
  
  # 获取用户信息
  curl -s http://localhost:8080/api/auth/current \
    -H "Authorization: Bearer $USER_TOKEN" | jq '{
      username: .data.username,
      role: .data.role.role_name,
      permission_count: (.data.role.permissions | length),
      permissions: [.data.role.permissions[] | .permission_key]
    }'
else
  echo "❌ 登录失败"
fi

# 3. 测试超级管理员访问所有资源
echo -e "\n3. 测试超级管理员权限"
echo "- 访问用户列表："
curl -s http://localhost:8080/api/users \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq '{code, total: .data.total}'

echo "- 访问角色列表："
curl -s http://localhost:8080/api/roles \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq '{code, total: .data.total}'

echo "- 访问系统配置："
curl -s http://localhost:8080/api/configs \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq '{code, total: .data.total}'

# 4. 测试普通用户权限
echo -e "\n4. 测试普通用户权限"
echo "- 访问用户列表："
curl -s http://localhost:8080/api/users \
  -H "Authorization: Bearer $USER_TOKEN" | jq '{code, total: .data.total}'

echo "- 访问统计数据："
curl -s http://localhost:8080/api/statistics \
  -H "Authorization: Bearer $USER_TOKEN" | jq '{code}'

echo "- 尝试访问系统配置（应该受限）："
curl -s http://localhost:8080/api/configs \
  -H "Authorization: Bearer $USER_TOKEN" | jq '{code, message}'

echo -e "\n✅ RBAC测试完成！"
