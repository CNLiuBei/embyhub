#!/bin/sh
# 启动脚本 - 在容器内运行

cd /app

# 等待数据库就绪
echo "等待数据库..."
sleep 5

# 修复前端文件权限
echo "修复文件权限..."
chmod -R 755 /usr/share/nginx/html

# 启动主服务
echo "启动主服务..."
chmod +x /app/server
/app/server &

sleep 2

# 启动Emby代理
echo "启动Emby代理..."
chmod +x /app/emby-proxy
/app/emby-proxy &

# 启动Nginx
echo "启动Nginx..."
nginx -g "daemon off;"
