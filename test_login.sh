#!/bin/bash

echo "=== 测试登录功能 ==="
echo ""

echo "1. 测试健康检查"
curl -s http://localhost:8080/health | jq .
echo ""

echo "2. 测试登录（admin/Liubei00）"
RESPONSE=$(curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Liubei00"}' \
  -s)
echo "$RESPONSE" | jq .

# 提取token
TOKEN=$(echo "$RESPONSE" | jq -r '.data.token // empty')

if [ -n "$TOKEN" ]; then
  echo ""
  echo "3. 登录成功！Token: ${TOKEN:0:50}..."
  echo ""
  
  echo "4. 测试获取当前用户信息"
  curl -s http://localhost:8080/api/auth/current \
    -H "Authorization: Bearer $TOKEN" | jq .
  echo ""
  
  echo "5. 测试获取角色列表"
  curl -s http://localhost:8080/api/roles \
    -H "Authorization: Bearer $TOKEN" | jq .
  echo ""
  
  echo "6. 测试获取用户列表"
  curl -s http://localhost:8080/api/users?page=1\&page_size=10 \
    -H "Authorization: Bearer $TOKEN" | jq .
else
  echo ""
  echo "❌ 登录失败"
fi
