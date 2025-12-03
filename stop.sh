#!/bin/bash

# Emby用户管理系统停止脚本
# 功能：停止后端和前端服务

set -e

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

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 停止服务
stop_service() {
    local service_name=$1
    local pid_file=$2
    
    if [ ! -f "$pid_file" ]; then
        log_warn "$service_name 未在运行（PID文件不存在）"
        return 0
    fi
    
    local pid=$(cat "$pid_file")
    
    if ! ps -p $pid > /dev/null 2>&1; then
        log_warn "$service_name 进程不存在 (PID: $pid)"
        rm -f "$pid_file"
        return 0
    fi
    
    log_info "停止 $service_name (PID: $pid)..."
    kill $pid 2>/dev/null || true
    
    # 等待进程结束
    local count=0
    while ps -p $pid > /dev/null 2>&1 && [ $count -lt 10 ]; do
        sleep 1
        count=$((count + 1))
    done
    
    # 如果进程仍在运行，强制杀死
    if ps -p $pid > /dev/null 2>&1; then
        log_warn "$service_name 未响应，强制停止..."
        kill -9 $pid 2>/dev/null || true
        sleep 1
    fi
    
    if ps -p $pid > /dev/null 2>&1; then
        log_error "无法停止 $service_name (PID: $pid)"
        return 1
    else
        log_success "$service_name 已停止"
        rm -f "$pid_file"
        return 0
    fi
}

# 清理端口占用
cleanup_port() {
    local port=$1
    local service_name=$2
    
    log_info "检查端口 $port ($service_name)..."
    local pid=$(lsof -ti:$port 2>/dev/null || true)
    
    if [ -n "$pid" ]; then
        log_warn "端口 $port 被进程 $pid 占用，正在清理..."
        kill $pid 2>/dev/null || kill -9 $pid 2>/dev/null || true
        sleep 1
        log_success "端口 $port 已清理"
    fi
}

# 主函数
main() {
    echo "=========================================="
    echo "    Emby用户管理系统 - 停止脚本"
    echo "=========================================="
    echo ""
    
    # 停止服务
    stop_service "后端服务" "$BACKEND_PID_FILE"
    stop_service "前端服务" "$FRONTEND_PID_FILE"
    
    # 额外清理可能占用的端口
    cleanup_port 8080 "后端"
    cleanup_port 3000 "前端"
    cleanup_port 3001 "前端"
    
    echo ""
    log_success "所有服务已停止"
    echo ""
}

# 执行主函数
main "$@"
