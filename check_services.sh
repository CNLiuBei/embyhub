#!/bin/bash

echo "========== 服务状态检查 =========="
echo ""

# 检查后端服务
echo "1. 后端服务 (Port 8080):"
if netstat -tlnp 2>/dev/null | grep -q ":8080" || ss -tlnp 2>/dev/null | grep -q ":8080"; then
    echo "   ✅ 后端服务正在运行"
    echo "   测试后端API:"
    curl -s http://localhost:8080/health && echo " ✅" || echo " ❌"
else
    echo "   ❌ 后端服务未运行"
fi

echo ""

# 检查前端服务
echo "2. 前端服务 (Port 3000):"
if netstat -tlnp 2>/dev/null | grep -q ":3000" || ss -tlnp 2>/dev/null | grep -q ":3000"; then
    echo "   ✅ 前端服务正在运行"
else
    echo "   ❌ 前端服务未运行"
fi

echo ""

# 检查域名白名单配置
echo "3. 域名白名单配置:"
LOG_FILE=$(ls -t logs/backend_*.log 2>/dev/null | head -1)
if [ -n "$LOG_FILE" ]; then
    DOMAIN_CONFIG=$(grep "domain_whitelist" "$LOG_FILE" | grep "UPDATE" | tail -1)
    if [ -n "$DOMAIN_CONFIG" ]; then
        echo "   最新配置:"
        echo "$DOMAIN_CONFIG" | grep -o '{"enabled[^}]*}' | head -1
    else
        echo "   未找到域名白名单配置"
    fi
fi

echo ""

# 检查最近的API错误
echo "4. 最近的API错误:"
if [ -n "$LOG_FILE" ]; then
    ERRORS=$(tail -100 "$LOG_FILE" | grep -E "ERROR|403|FATAL" | tail -5)
    if [ -n "$ERRORS" ]; then
        echo "$ERRORS"
    else
        echo "   ✅ 无错误"
    fi
fi

echo ""

# 测试域名白名单API
echo "5. 测试域名白名单API:"
echo "   测试从localhost访问:"
RESPONSE=$(curl -s -w "\n%{http_code}" http://localhost:8080/health 2>&1)
HTTP_CODE=$(echo "$RESPONSE" | tail -1)
if [ "$HTTP_CODE" = "200" ]; then
    echo "   ✅ localhost访问成功"
else
    echo "   ❌ localhost访问失败 (HTTP $HTTP_CODE)"
fi

echo ""
echo "========== 检查完成 =========="
