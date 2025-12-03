#!/bin/bash

# Emby用户管理系统状态查看脚本

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目路径
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PID_DIR="$PROJECT_ROOT/.pids"
BACKEND_PID_FILE="$PID_DIR/backend.pid"
FRONTEND_PID_FILE="$PID_DIR/frontend.pid"

# 端口配置
BACKEND_PORT=8080
FRONTEND_PORT_1=3000
FRONTEND_PORT_2=3001
DB_PORT=5432
REDIS_PORT=6379

# 检查端口
check_port() {
    lsof -Pi :$1 -sTCP:LISTEN -t >/dev/null 2>&1
}

# 获取端口占用的PID
get_port_pid() {
    lsof -ti:$1 2>/dev/null || echo ""
}

# 检查服务状态
check_service() {
    local service_name=$1
    local pid_file=$2
    local port=$3
    
    echo -n "$service_name: "
    
    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        if ps -p $pid > /dev/null 2>&1; then
            echo -e "${GREEN}运行中${NC} (PID: $pid)"
            if [ -n "$port" ] && check_port $port; then
                echo "  └─ 端口: $port"
            fi
            return 0
        else
            echo -e "${RED}已停止${NC} (PID文件存在但进程不存在)"
            return 1
        fi
    else
        if [ -n "$port" ] && check_port $port; then
            local pid=$(get_port_pid $port)
            echo -e "${YELLOW}运行中${NC} (端口: $port, PID: $pid) ${YELLOW}[无PID文件]${NC}"
            return 0
        else
            echo -e "${RED}未运行${NC}"
            return 1
        fi
    fi
}

# 检查数据库
check_database() {
    echo -n "PostgreSQL: "
    if check_port $DB_PORT; then
        if PGPASSWORD=embyhub123 psql -U embyhub -h localhost -d embyhub -c "SELECT 1" >/dev/null 2>&1; then
            echo -e "${GREEN}运行中${NC} (端口: $DB_PORT) ${GREEN}[可连接]${NC}"
            return 0
        else
            echo -e "${YELLOW}运行中${NC} (端口: $DB_PORT) ${RED}[连接失败]${NC}"
            return 1
        fi
    else
        echo -e "${RED}未运行${NC}"
        return 1
    fi
}

# 检查Redis
check_redis() {
    echo -n "Redis: "
    if check_port $REDIS_PORT; then
        if redis-cli -a embyhub123 ping >/dev/null 2>&1; then
            echo -e "${GREEN}运行中${NC} (端口: $REDIS_PORT) ${GREEN}[可连接]${NC}"
            return 0
        else
            echo -e "${YELLOW}运行中${NC} (端口: $REDIS_PORT) ${RED}[连接失败]${NC}"
            return 1
        fi
    else
        echo -e "${RED}未运行${NC}"
        return 1
    fi
}

# 测试API
test_api() {
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  API 健康检查"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    if check_port $BACKEND_PORT; then
        echo -n "后端 API: "
        if curl -s http://localhost:$BACKEND_PORT/health >/dev/null 2>&1; then
            echo -e "${GREEN}正常${NC}"
            echo "  └─ http://localhost:$BACKEND_PORT/health"
        else
            echo -e "${RED}无响应${NC}"
        fi
    fi
    
    if check_port $FRONTEND_PORT_2; then
        echo -n "前端应用: "
        if curl -s http://localhost:$FRONTEND_PORT_2 >/dev/null 2>&1; then
            echo -e "${GREEN}正常${NC}"
            echo "  └─ http://localhost:$FRONTEND_PORT_2"
        else
            echo -e "${RED}无响应${NC}"
        fi
    elif check_port $FRONTEND_PORT_1; then
        echo -n "前端应用: "
        if curl -s http://localhost:$FRONTEND_PORT_1 >/dev/null 2>&1; then
            echo -e "${GREEN}正常${NC}"
            echo "  └─ http://localhost:$FRONTEND_PORT_1"
        else
            echo -e "${RED}无响应${NC}"
        fi
    fi
}

# 显示日志信息
show_logs() {
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  日志文件"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    if [ -f "$PROJECT_ROOT/logs/backend.log" ]; then
        local size=$(du -h "$PROJECT_ROOT/logs/backend.log" | cut -f1)
        echo "后端日志: $PROJECT_ROOT/logs/backend.log ($size)"
    fi
    
    if [ -f "$PROJECT_ROOT/logs/frontend.log" ]; then
        local size=$(du -h "$PROJECT_ROOT/logs/frontend.log" | cut -f1)
        echo "前端日志: $PROJECT_ROOT/logs/frontend.log ($size)"
    fi
}

# 主函数
main() {
    echo "=========================================="
    echo "    Emby用户管理系统 - 服务状态"
    echo "=========================================="
    echo ""
    
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  应用服务"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    check_service "后端服务" "$BACKEND_PID_FILE" "$BACKEND_PORT"
    check_service "前端服务" "$FRONTEND_PID_FILE" ""
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  基础设施"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    check_database
    check_redis
    
    test_api
    show_logs
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  快速命令"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "启动服务: ./start.sh"
    echo "停止服务: ./stop.sh"
    echo "查看后端日志: tail -f logs/backend.log"
    echo "查看前端日志: tail -f logs/frontend.log"
    echo ""
}

# 执行主函数
main "$@"
