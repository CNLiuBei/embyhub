#!/bin/bash

echo "=========================================="
echo "   Nginx SSL配置错误修复脚本"
echo "=========================================="
echo ""

CONF_FILE="/www/server/panel/vhost/nginx/xy.trueliu.com.conf"

if [ ! -f "$CONF_FILE" ]; then
    echo "❌ 配置文件不存在: $CONF_FILE"
    echo "请确认文件路径是否正确"
    exit 1
fi

echo "📋 当前配置文件: $CONF_FILE"
echo ""

# 检查是否有SSL配置
if grep -q "listen.*ssl" "$CONF_FILE"; then
    echo "✓ 检测到SSL配置"
    echo ""
    
    # 检查是否有证书配置
    if grep -q "ssl_certificate" "$CONF_FILE"; then
        echo "✓ 已配置SSL证书"
        echo ""
        echo "证书文件："
        grep "ssl_certificate" "$CONF_FILE"
    else
        echo "❌ 未配置SSL证书！"
        echo ""
        echo "请选择解决方案："
        echo ""
        echo "方案1：移除SSL，使用HTTP（适合测试）"
        echo "  sed -i 's/listen.*ssl/listen 80/' $CONF_FILE"
        echo ""
        echo "方案2：配置SSL证书（生产环境推荐）"
        echo "  1. 在宝塔面板中申请Let's Encrypt证书"
        echo "  2. 或手动添加证书配置："
        echo "     ssl_certificate /path/to/cert.pem;"
        echo "     ssl_certificate_key /path/to/key.pem;"
        echo ""
        echo "方案3：使用Let's Encrypt自动申请"
        echo "  certbot --nginx -d xy.trueliu.com"
        echo ""
        
        read -p "是否自动执行方案1（移除SSL）？[y/N] " -n 1 -r
        echo ""
        
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo "正在备份配置文件..."
            cp "$CONF_FILE" "${CONF_FILE}.backup.$(date +%Y%m%d_%H%M%S)"
            
            echo "正在移除SSL配置..."
            sed -i 's/listen.*443.*ssl/listen 80/' "$CONF_FILE"
            sed -i '/ssl_/d' "$CONF_FILE"
            
            echo "✓ SSL配置已移除"
            echo ""
            echo "测试配置..."
            nginx -t
            
            if [ $? -eq 0 ]; then
                echo ""
                echo "✓ 配置测试通过"
                read -p "是否重载Nginx配置？[y/N] " -n 1 -r
                echo ""
                if [[ $REPLY =~ ^[Yy]$ ]]; then
                    nginx -s reload
                    echo "✓ Nginx配置已重载"
                fi
            else
                echo ""
                echo "❌ 配置测试失败，请手动检查"
            fi
        else
            echo "已取消操作"
        fi
    fi
else
    echo "✓ 未使用SSL配置"
fi

echo ""
echo "=========================================="
echo "完成"
echo "=========================================="
