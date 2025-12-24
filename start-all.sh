#!/bin/bash

# Emby用户系统 - 一键启动脚本
# 作者: Cascade AI
# 说明: 同时启动前端和后端服务

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="${PROJECT_ROOT}/backend"
FRONTEND_DIR="${PROJECT_ROOT}/frontend"

clear
echo -e "${PURPLE}╔════════════════════════════════════════╗${NC}"
echo -e "${PURPLE}║   Emby用户系统 - 一键启动脚本         ║${NC}"
echo -e "${PURPLE}╚════════════════════════════════════════╝${NC}"
echo ""

# 检查目录是否存在
echo -e "${CYAN}🔍 检查项目结构...${NC}"
if [ ! -d "${BACKEND_DIR}" ]; then
    echo -e "${RED}✗ 错误: 后端目录不存在: ${BACKEND_DIR}${NC}"
    exit 1
fi
echo -e "${GREEN}✓ 后端目录存在${NC}"

if [ ! -d "${FRONTEND_DIR}" ]; then
    echo -e "${RED}✗ 错误: 前端目录不存在: ${FRONTEND_DIR}${NC}"
    exit 1
fi
echo -e "${GREEN}✓ 前端目录存在${NC}"
echo ""

# 询问是否先停止旧服务
if pgrep -f "cmd/server/main" >/dev/null 2>&1 || pgrep -f "vite" >/dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  检测到服务正在运行${NC}"
    read -p "是否先停止旧服务？[Y/n] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]] || [[ -z $REPLY ]]; then
        echo -e "${CYAN}🛑 停止旧服务...${NC}"
        "${PROJECT_ROOT}/stop-all.sh"
        sleep 2
    fi
    echo ""
fi

# 启动后端
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}[1/2] 🚀 启动后端服务${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
cd "${BACKEND_DIR}" || exit 1

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo -e "${RED}✗ 错误: 未找到Go环境，请先安装Go${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Go版本: $(go version | awk '{print $3}')${NC}"

# 停止旧的后端进程
if pgrep -f "cmd/server/main" >/dev/null 2>&1; then
    echo -e "${YELLOW}🛑 停止旧的后端服务...${NC}"
    pkill -15 -f "cmd/server/main" 2>/dev/null || true
    sleep 2
    pkill -9 -f "cmd/server/main" 2>/dev/null || true
    sleep 1
fi

# 清理端口占用
if lsof -ti:54680 >/dev/null 2>&1; then
    echo -e "${YELLOW}🧹 清理端口54680占用...${NC}"
    lsof -ti:54680 | xargs kill -9 2>/dev/null || true
    sleep 1
fi

# 创建日志目录
mkdir -p logs

# 启动后端服务
echo -e "${CYAN}⏳ 启动后端服务...${NC}"
nohup go run cmd/server/main.go > logs/backend_$(date +%Y%m%d_%H%M%S).log 2>&1 &
BACKEND_PID=$!
echo $BACKEND_PID > logs/app.pid

# 等待后端启动
echo -e "${CYAN}⏳ 等待后端启动...${NC}"
sleep 3

# 检查后端是否启动成功
if ps -p $BACKEND_PID > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 后端服务启动成功 (PID: $BACKEND_PID)${NC}"
else
    echo -e "${RED}✗ 后端启动失败！${NC}"
    exit 1
fi

echo ""
echo -e "${CYAN}⏳ 等待后端服务就绪...${NC}"
sleep 3

# 验证后端是否成功启动
if curl -s http://localhost:54680/health >/dev/null 2>&1; then
    echo -e "${GREEN}✓ 后端健康检查通过${NC}"
else
    echo -e "${YELLOW}⚠️  后端健康检查未通过，可能还在启动中${NC}"
fi

echo ""

# 启动 Emby 代理服务
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}[2/3] 🎬 启动 Emby 代理服务${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# 停止旧的 emby-proxy 进程
if pgrep -f "cmd/emby-proxy/main" >/dev/null 2>&1; then
    echo -e "${YELLOW}🛑 停止旧的 Emby 代理服务...${NC}"
    pkill -15 -f "cmd/emby-proxy/main" 2>/dev/null || true
    sleep 2
    pkill -9 -f "cmd/emby-proxy/main" 2>/dev/null || true
    sleep 1
fi

# 清理端口占用
if lsof -ti:54682 >/dev/null 2>&1; then
    echo -e "${YELLOW}🧹 清理端口54682占用...${NC}"
    lsof -ti:54682 | xargs kill -9 2>/dev/null || true
    sleep 1
fi

# 启动 Emby 代理服务
echo -e "${CYAN}⏳ 启动 Emby 代理服务...${NC}"
nohup go run cmd/emby-proxy/main.go > logs/emby_proxy_$(date +%Y%m%d_%H%M%S).log 2>&1 &
EMBY_PROXY_PID=$!
echo $EMBY_PROXY_PID > logs/emby_proxy.pid

# 等待启动
sleep 3

# 检查是否启动成功
if ps -p $EMBY_PROXY_PID > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Emby 代理服务启动成功 (PID: $EMBY_PROXY_PID)${NC}"
else
    echo -e "${YELLOW}⚠️  Emby 代理服务可能还在启动中...${NC}"
fi

echo ""

# 启动前端
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}[3/3] 🎨 启动前端服务${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
cd "${FRONTEND_DIR}" || exit 1

# 检查Node.js环境
if ! command -v node &> /dev/null; then
    echo -e "${RED}✗ 错误: 未找到Node.js，请先安装${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Node版本: $(node -v)${NC}"

# 检查依赖
if [ ! -d "node_modules" ]; then
    echo -e "${YELLOW}⏳ 未找到依赖，开始安装...${NC}"
    npm install || {
        echo -e "${RED}✗ 依赖安装失败！${NC}"
        exit 1
    }
fi

# 停止旧的前端进程
if pgrep -f "vite" >/dev/null 2>&1; then
    echo -e "${YELLOW}🛑 停止旧的前端服务...${NC}"
    pkill -9 -f "vite" 2>/dev/null || true
    pkill -9 -f "npm run dev" 2>/dev/null || true
    sleep 1
fi

# 创建日志目录
mkdir -p logs

# 启动前端开发服务器
echo -e "${CYAN}⏳ 启动前端开发服务器...${NC}"
nohup npm run dev > logs/frontend_$(date +%Y%m%d_%H%M%S).log 2>&1 &
FRONTEND_PID=$!
echo $FRONTEND_PID > logs/npm.pid

# 等待前端启动
sleep 5

# 检查前端是否启动成功
if pgrep -f "vite" >/dev/null 2>&1; then
    echo -e "${GREEN}✓ 前端服务启动成功 (PID: $FRONTEND_PID)${NC}"
else
    echo -e "${YELLOW}⚠️  前端可能还在启动中...${NC}"
fi

echo ""
echo -e "${PURPLE}╔════════════════════════════════════════╗${NC}"
echo -e "${PURPLE}║  ${GREEN}🎉 所有服务启动完成！${PURPLE}                 ║${NC}"
echo -e "${PURPLE}╚════════════════════════════════════════╝${NC}"
echo ""
echo -e "${CYAN}📍 访问地址:${NC}"
echo -e "  ${GREEN}前端: ${BLUE}http://localhost:54681${NC}"
echo -e "  ${GREEN}后端: ${BLUE}http://localhost:54680${NC}"
echo -e "  ${GREEN}Emby代理: ${BLUE}http://localhost:54682${NC}"
echo ""
echo -e "${CYAN}🔑 默认账号:${NC}"
echo -e "  ${GREEN}用户名: ${BLUE}admin${NC}"
echo -e "  ${GREEN}密码:   ${BLUE}admin123${NC}"
echo -e "  ${RED}⚠️  首次登录后请立即修改密码！${NC}"
echo ""
echo -e "${CYAN}📊 管理命令:${NC}"
echo -e "  ${GREEN}健康检查: ${BLUE}curl http://localhost:54680/health${NC}"
echo -e "  ${GREEN}后端日志: ${BLUE}tail -f ${BACKEND_DIR}/logs/backend_*.log${NC}"
echo -e "  ${GREEN}前端日志: ${BLUE}tail -f ${FRONTEND_DIR}/logs/frontend_*.log${NC}"
echo -e "  ${GREEN}停止服务: ${BLUE}./stop-all.sh${NC}"
echo ""
echo -e "${PURPLE}════════════════════════════════════════${NC}"
echo ""
