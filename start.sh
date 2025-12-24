#!/bin/bash

# EmbyHub用户管理系统启动脚本

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$PROJECT_DIR/backend"
FRONTEND_DIR="$PROJECT_DIR/frontend"

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    if ! command -v go &> /dev/null; then
        log_error "未找到Go环境，请先安装Go 1.21+"
        exit 1
    fi
    
    if ! command -v node &> /dev/null; then
        log_error "未找到Node.js环境，请先安装Node.js 18+"
        exit 1
    fi
    
    log_success "依赖检查通过"
}

# 初始化后端
init_backend() {
    log_info "初始化后端依赖..."
    cd "$BACKEND_DIR"
    go mod tidy
    log_success "后端依赖安装完成"
}

# 初始化前端
init_frontend() {
    log_info "初始化前端依赖..."
    cd "$FRONTEND_DIR"
    npm install
    log_success "前端依赖安装完成"
}

# 启动后端
start_backend() {
    log_info "启动后端服务..."
    cd "$BACKEND_DIR"
    
    # 创建日志目录
    mkdir -p logs
    
    # 后台启动
    nohup go run cmd/server/main.go > "$PROJECT_DIR/backend.log" 2>&1 &
    
    log_info "等待后端启动..."
    sleep 5
    
    if lsof -Pi :54680 -sTCP:LISTEN -t >/dev/null 2>&1; then
        log_success "后端服务已启动 (端口 54680)"
    else
        log_error "后端启动失败，请查看日志: $PROJECT_DIR/backend.log"
        exit 1
    fi
}

# 启动前端
start_frontend() {
    log_info "启动前端服务..."
    cd "$FRONTEND_DIR"
    
    nohup npm run dev > "$PROJECT_DIR/frontend.log" 2>&1 &
    
    sleep 3
    
    if lsof -Pi :54681 -sTCP:LISTEN -t >/dev/null 2>&1; then
        log_success "前端服务已启动 (端口 54681)"
    else
        log_error "前端启动失败，请查看日志: $PROJECT_DIR/frontend.log"
        exit 1
    fi
}

# 停止服务
stop_services() {
    log_info "停止服务..."
    
    # 停止后端
    pkill -f "go run cmd/server/main.go" 2>/dev/null || true
    fuser -k 54680/tcp 2>/dev/null || true
    
    # 停止前端
    pkill -f "vite" 2>/dev/null || true
    fuser -k 54681/tcp 2>/dev/null || true
    
    log_success "服务已停止"
}

# 显示状态
show_status() {
    echo ""
    echo "======================================"
    echo "       EmbyHub用户管理系统"
    echo "======================================"
    
    if lsof -Pi :54680 -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo -e "后端服务: ${GREEN}运行中${NC} (http://localhost:54680)"
    else
        echo -e "后端服务: ${RED}未运行${NC}"
    fi
    
    if lsof -Pi :54681 -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo -e "前端服务: ${GREEN}运行中${NC} (http://localhost:54681)"
    else
        echo -e "前端服务: ${RED}未运行${NC}"
    fi
    
    echo "======================================"
}

# 主函数
case "${1:-start}" in
    init)
        check_dependencies
        init_backend
        init_frontend
        log_success "初始化完成！运行 ./start.sh 启动服务"
        ;;
    start)
        check_dependencies
        stop_services
        start_backend
        start_frontend
        show_status
        echo ""
        log_info "访问地址: http://localhost:54681"
        log_info "API文档: http://localhost:54680/swagger/index.html"
        ;;
    stop)
        stop_services
        show_status
        ;;
    restart)
        $0 stop
        sleep 2
        $0 start
        ;;
    status)
        show_status
        ;;
    *)
        echo "用法: $0 {init|start|stop|restart|status}"
        echo ""
        echo "  init    - 初始化项目(安装依赖)"
        echo "  start   - 启动服务"
        echo "  stop    - 停止服务"
        echo "  restart - 重启服务"
        echo "  status  - 查看状态"
        exit 1
        ;;
esac
