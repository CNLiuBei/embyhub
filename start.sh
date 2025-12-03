#!/bin/bash

# Emby用户管理系统启动脚本
# 功能：启动后端和前端服务，并检查是否重复启动

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目路径
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$PROJECT_ROOT/backend"
FRONTEND_DIR="$PROJECT_ROOT/frontend"

# 端口配置
BACKEND_PORT=8080
FRONTEND_PORT=3001
DB_PORT=5432
REDIS_PORT=6379

# PID文件路径
PID_DIR="$PROJECT_ROOT/.pids"
BACKEND_PID_FILE="$PID_DIR/backend.pid"
FRONTEND_PID_FILE="$PID_DIR/frontend.pid"

# 创建PID目录
mkdir -p "$PID_DIR"

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

# 检查端口是否被占用
check_port() {
    local port=$1
    if nc -z localhost $port >/dev/null 2>&1; then
        return 0  # 端口被占用
    else
        return 1  # 端口空闲
    fi
}

# 检查服务是否在运行
check_service_running() {
    local service_name=$1
    local pid_file=$2
    
    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        if ps -p $pid > /dev/null 2>&1; then
            return 0  # 服务正在运行
        else
            # PID文件存在但进程不存在，清理PID文件
            rm -f "$pid_file"
            return 1
        fi
    fi
    return 1
}

# 检查数据库连接
check_database() {
    log_info "检查 PostgreSQL 连接..."
    if check_port $DB_PORT; then
        if PGPASSWORD=embyhub123 psql -U embyhub -h localhost -d embyhub -c "SELECT 1" >/dev/null 2>&1; then
            log_success "PostgreSQL 连接正常"
            return 0
        else
            log_error "PostgreSQL 连接失败，请检查数据库配置"
            return 1
        fi
    else
        log_error "PostgreSQL 未运行在端口 $DB_PORT"
        return 1
    fi
}

# 检查Redis连接
check_redis() {
    log_info "检查 Redis 连接..."
    if check_port $REDIS_PORT; then
        if redis-cli -a embyhub123 ping >/dev/null 2>&1; then
            log_success "Redis 连接正常"
            return 0
        else
            log_error "Redis 连接失败，请检查Redis配置"
            return 1
        fi
    else
        log_error "Redis 未运行在端口 $REDIS_PORT"
        return 1
    fi
}

# 启动后端服务
start_backend() {
    log_info "检查后端服务状态..."
    
    if check_service_running "后端" "$BACKEND_PID_FILE"; then
        log_warn "后端服务已在运行 (PID: $(cat $BACKEND_PID_FILE))"
        return 0
    fi
    
    if check_port $BACKEND_PORT; then
        log_error "端口 $BACKEND_PORT 已被占用，无法启动后端"
        return 1
    fi
    
    log_info "启动后端服务..."
    cd "$BACKEND_DIR"
    
    # 检查依赖
    if [ ! -d "vendor" ] && [ ! -f "go.sum" ]; then
        log_info "下载 Go 依赖..."
        go mod download
    fi
    
    # 后台启动后端
    nohup go run cmd/main.go > "$PROJECT_ROOT/logs/backend.log" 2>&1 &
    local pid=$!
    echo $pid > "$BACKEND_PID_FILE"
    
    # 等待服务启动
    sleep 3
    
    if ps -p $pid > /dev/null; then
        log_success "后端服务启动成功 (PID: $pid, 端口: $BACKEND_PORT)"
        return 0
    else
        log_error "后端服务启动失败，查看日志: $PROJECT_ROOT/logs/backend.log"
        rm -f "$BACKEND_PID_FILE"
        return 1
    fi
}

# 启动前端服务
start_frontend() {
    log_info "检查前端服务状态..."
    
    if check_service_running "前端" "$FRONTEND_PID_FILE"; then
        log_warn "前端服务已在运行 (PID: $(cat $FRONTEND_PID_FILE))"
        return 0
    fi
    
    log_info "启动前端服务..."
    cd "$FRONTEND_DIR"
    
    # 检查依赖
    if [ ! -d "node_modules" ]; then
        log_info "安装 npm 依赖..."
        npm install
    fi
    
    # 后台启动前端
    nohup npm run dev > "$PROJECT_ROOT/logs/frontend.log" 2>&1 &
    local pid=$!
    echo $pid > "$FRONTEND_PID_FILE"
    
    # 等待服务启动
    sleep 5
    
    if ps -p $pid > /dev/null; then
        log_success "前端服务启动成功 (PID: $pid)"
        
        # 查找实际运行的端口
        sleep 2
        if check_port 3001; then
            log_success "前端运行在: http://localhost:3001"
        elif check_port 3000; then
            log_success "前端运行在: http://localhost:3000"
        fi
        return 0
    else
        log_error "前端服务启动失败，查看日志: $PROJECT_ROOT/logs/frontend.log"
        rm -f "$FRONTEND_PID_FILE"
        return 1
    fi
}

# 显示服务状态
show_status() {
    echo ""
    echo "=========================================="
    echo "           服务运行状态"
    echo "=========================================="
    
    # 后端状态
    if check_service_running "后端" "$BACKEND_PID_FILE"; then
        echo -e "${GREEN}✓${NC} 后端服务: 运行中 (PID: $(cat $BACKEND_PID_FILE), 端口: $BACKEND_PORT)"
    else
        echo -e "${RED}✗${NC} 后端服务: 未运行"
    fi
    
    # 前端状态
    if check_service_running "前端" "$FRONTEND_PID_FILE"; then
        echo -e "${GREEN}✓${NC} 前端服务: 运行中 (PID: $(cat $FRONTEND_PID_FILE))"
    else
        echo -e "${RED}✗${NC} 前端服务: 未运行"
    fi
    
    # 数据库状态
    if check_port $DB_PORT; then
        echo -e "${GREEN}✓${NC} PostgreSQL: 运行中 (端口: $DB_PORT)"
    else
        echo -e "${RED}✗${NC} PostgreSQL: 未运行"
    fi
    
    # Redis状态
    if check_port $REDIS_PORT; then
        echo -e "${GREEN}✓${NC} Redis: 运行中 (端口: $REDIS_PORT)"
    else
        echo -e "${RED}✗${NC} Redis: 未运行"
    fi
    
    echo "=========================================="
    echo ""
}

# 主函数
main() {
    echo "=========================================="
    echo "    Emby用户管理系统 - 启动脚本"
    echo "=========================================="
    echo ""
    
    # 创建日志目录
    mkdir -p "$PROJECT_ROOT/logs"
    
    # 检查依赖服务
    if ! check_database; then
        log_error "数据库检查失败，请先启动 PostgreSQL"
        exit 1
    fi
    
    if ! check_redis; then
        log_error "Redis检查失败，请先启动 Redis"
        exit 1
    fi
    
    # 启动服务
    start_backend
    start_frontend
    
    # 显示状态
    show_status
    
    log_success "所有服务启动完成！"
    echo ""
    echo "访问地址："
    echo "  - 前端: http://localhost:3001 (或 3000)"
    echo "  - 后端: http://localhost:8080"
    echo "  - 健康检查: http://localhost:8080/health"
    echo ""
    echo "默认登录账号："
    echo "  - 用户名: admin"
    echo "  - 密码: Liubei00"
    echo ""
    echo "查看日志："
    echo "  - 后端: tail -f $PROJECT_ROOT/logs/backend.log"
    echo "  - 前端: tail -f $PROJECT_ROOT/logs/frontend.log"
    echo ""
    echo "停止服务: ./stop.sh"
    echo ""
}

# 执行主函数
main "$@"
