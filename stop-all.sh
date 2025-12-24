#!/bin/bash

# EmbyHub - 停止服务脚本
# 作者: Cascade AI  
# 说明: 停止前端和后端服务

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

echo -e "${PURPLE}========================================"
echo -e "   EmbyHub - 停止服务"
echo -e "========================================${NC}"
echo ""

# 停止后端服务
echo -e "${YELLOW}正在停止后端服务...${NC}"

# 检查PID文件
PID_FILE="./backend/logs/app.pid"
if [ -f "$PID_FILE" ]; then
    PID=$(cat "$PID_FILE" 2>/dev/null)
    if [ -n "$PID" ] && kill -0 "$PID" 2>/dev/null; then
        kill -15 "$PID" 2>/dev/null || true
        sleep 2
        if kill -0 "$PID" 2>/dev/null; then
            kill -9 "$PID" 2>/dev/null || true
        fi
        rm -f "$PID_FILE"
        echo -e "${GREEN}✓ 后端服务已停止 (PID: $PID)${NC}"
    fi
fi

# 停止所有Go后端进程
PIDS=$(pgrep -f "cmd/server/main")
if [ -n "$PIDS" ]; then
    echo -e "${YELLOW}停止残留后端进程: $PIDS${NC}"
    kill -15 $PIDS 2>/dev/null || true
    sleep 1
    pkill -9 -f "cmd/server/main" 2>/dev/null || true
    echo -e "${GREEN}✓ 所有后端进程已停止${NC}"
else
    echo -e "${BLUE}  后端服务未运行${NC}"
fi

echo ""

# 停止 Emby 代理服务
echo -e "${YELLOW}正在停止 Emby 代理服务...${NC}"

# 检查PID文件
EMBY_PID_FILE="./backend/logs/emby_proxy.pid"
if [ -f "$EMBY_PID_FILE" ]; then
    EMBY_PID=$(cat "$EMBY_PID_FILE" 2>/dev/null)
    if [ -n "$EMBY_PID" ] && kill -0 "$EMBY_PID" 2>/dev/null; then
        kill -15 "$EMBY_PID" 2>/dev/null || true
        sleep 2
        if kill -0 "$EMBY_PID" 2>/dev/null; then
            kill -9 "$EMBY_PID" 2>/dev/null || true
        fi
        rm -f "$EMBY_PID_FILE"
        echo -e "${GREEN}✓ Emby 代理服务已停止 (PID: $EMBY_PID)${NC}"
    fi
fi

# 停止所有 emby-proxy 进程
EMBY_PIDS=$(pgrep -f "cmd/emby-proxy/main")
if [ -n "$EMBY_PIDS" ]; then
    echo -e "${YELLOW}停止残留 Emby 代理进程: $EMBY_PIDS${NC}"
    kill -15 $EMBY_PIDS 2>/dev/null || true
    sleep 1
    pkill -9 -f "cmd/emby-proxy/main" 2>/dev/null || true
    echo -e "${GREEN}✓ 所有 Emby 代理进程已停止${NC}"
else
    echo -e "${BLUE}  Emby 代理服务未运行${NC}"
fi

echo ""

# 停止前端服务
echo -e "${YELLOW}正在停止前端服务...${NC}"

# 停止vite进程
VITE_PIDS=$(pgrep -f "vite")
if [ -n "$VITE_PIDS" ]; then
    echo -e "${YELLOW}停止vite进程: $VITE_PIDS${NC}"
    kill -15 $VITE_PIDS 2>/dev/null || true
    sleep 1
    pkill -9 -f "vite" 2>/dev/null || true
    echo -e "${GREEN}✓ vite进程已停止${NC}"
else
    echo -e "${BLUE}  vite未运行${NC}"
fi

# 停止npm进程
NPM_PIDS=$(pgrep -f "npm run dev")
if [ -n "$NPM_PIDS" ]; then
    echo -e "${YELLOW}停止npm进程: $NPM_PIDS${NC}"
    kill -15 $NPM_PIDS 2>/dev/null || true
    sleep 1
    pkill -9 -f "npm run dev" 2>/dev/null || true
    echo -e "${GREEN}✓ npm进程已停止${NC}"
else
    echo -e "${BLUE}  npm未运行${NC}"
fi

# 停止node进程（前端相关）
NODE_PIDS=$(pgrep -f "node.*vite")
if [ -n "$NODE_PIDS" ]; then
    kill -15 $NODE_PIDS 2>/dev/null || true
    sleep 1
    kill -9 $NODE_PIDS 2>/dev/null || true
    echo -e "${GREEN}✓ node进程已停止${NC}"
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}✓ 所有服务已停止${NC}"
echo -e "${GREEN}========================================${NC}"
