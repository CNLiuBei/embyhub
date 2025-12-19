#!/bin/bash

echo "========== 域名白名单配置检查 =========="
echo ""

# 检查数据库配置
echo "1. 数据库中的配置："
echo "---"

# 从最新的日志中查找域名白名单配置
LOG_FILE=$(ls -t logs/backend_*.log 2>/dev/null | head -1)

if [ -n "$LOG_FILE" ]; then
    # 查找最近的UPDATE语句
    LATEST_CONFIG=$(grep "domain_whitelist" "$LOG_FILE" | grep "UPDATE\|INSERT" | tail -1)
    
    if [ -n "$LATEST_CONFIG" ]; then
        echo "最新保存的配置："
        echo "$LATEST_CONFIG" | grep -o '{"enabled[^}]*}' | python3 -m json.tool 2>/dev/null || echo "$LATEST_CONFIG"
    else
        echo "⚠️  数据库中可能还没有域名白名单配置"
        echo "   默认状态：enabled=false (白名单关闭)"
    fi
else
    echo "⚠️  未找到后端日志文件"
fi

echo ""
echo "2. 当前白名单状态："
echo "---"

# 通过API获取当前配置
RESPONSE=$(curl -s http://localhost:8080/api/v1/admin/settings/domain 2>/dev/null)

if [ -n "$RESPONSE" ]; then
    echo "$RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"
else
    echo "⚠️  无法连接到后端API"
fi

echo ""
echo "3. 白名单工作逻辑："
echo "---"
echo "如果 enabled = false (关闭)："
echo "  → ✅ 允许所有域名访问"
echo ""
echo "如果 enabled = true (开启)："
echo "  → 只允许 domains 列表中的域名访问"
echo "  → 其他域名显示403错误页面"
echo ""

echo "4. 当前测试："
echo "---"
echo "测试localhost访问："
curl -s http://localhost:8080/health > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "  ✅ localhost 可以访问"
else
    echo "  ❌ localhost 无法访问"
fi

echo ""
echo "测试未授权域名访问："
RESPONSE=$(curl -s -H "Host: unauthorized-domain.com" http://localhost:8080/health 2>&1)
if echo "$RESPONSE" | grep -q "ok"; then
    echo "  ⚠️  unauthorized-domain.com 可以访问"
    echo "  原因：白名单功能未启用 (enabled=false)"
elif echo "$RESPONSE" | grep -q "403\|未授权"; then
    echo "  ✅ unauthorized-domain.com 被拒绝 (403)"
    echo "  白名单功能正常工作"
else
    echo "  ⚠️  测试结果不明确"
fi

echo ""
echo "========== 检查完成 =========="
echo ""
echo "💡 提示："
echo "  1. 如果要限制域名访问，需要先在UI中'启用域名白名单'开关"
echo "  2. 添加域名到列表后，必须打开开关才会生效"
echo "  3. 开关打开后，只有列表中的域名可以访问"
