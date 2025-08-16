#!/bin/bash

# =============================================================================
# MER System Docker 开发环境启动脚本
# =============================================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 函数：打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 函数：检查 Docker 是否安装
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi
    
    print_success "Docker 环境检查通过"
}

# 函数：检查环境文件
check_env_file() {
    if [ ! -f ".env" ]; then
        print_warning ".env 文件不存在，正在从 .env.example 复制..."
        cp .env.example .env
        print_info "请根据需要修改 .env 文件中的配置"
    fi
}

# 函数：创建必要的目录
create_directories() {
    print_info "创建必要的目录..."
    mkdir -p docker/mysql/data
    mkdir -p docker/redis/data
    mkdir -p logs
    mkdir -p uploads
    mkdir -p static
}

# 函数：检查端口是否被占用
check_ports() {
    local ports=(3306 6379 8080 8081 8082 5173)
    local occupied_ports=()
    
    for port in "${ports[@]}"; do
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            occupied_ports+=($port)
        fi
    done
    
    if [ ${#occupied_ports[@]} -gt 0 ]; then
        print_warning "以下端口被占用: ${occupied_ports[*]}"
        print_info "请检查是否有其他服务在使用这些端口"
        read -p "是否继续启动？(y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "启动已取消"
            exit 0
        fi
    fi
}

# 函数：启动服务
start_services() {
    print_info "启动 Docker 服务..."
    
    # 首先启动基础服务 (MySQL, Redis)
    print_info "启动基础服务 (MySQL, Redis)..."
    docker-compose up -d mysql redis
    
    # 等待基础服务启动
    print_info "等待基础服务启动..."
    sleep 10
    
    # 检查基础服务健康状态
    print_info "检查基础服务健康状态..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        mysql_health=$(docker-compose ps mysql --format json | jq -r '.[0].Health' 2>/dev/null || echo "unknown")
        redis_health=$(docker-compose ps redis --format json | jq -r '.[0].Health' 2>/dev/null || echo "unknown")
        
        if [[ "$mysql_health" == "healthy" && "$redis_health" == "healthy" ]]; then
            print_success "基础服务启动成功"
            break
        fi
        
        print_info "等待基础服务健康检查... ($attempt/$max_attempts)"
        sleep 5
        ((attempt++))
    done
    
    if [ $attempt -gt $max_attempts ]; then
        print_warning "基础服务健康检查超时，继续启动应用服务..."
    fi
    
    # 启动应用服务
    print_info "启动应用服务..."
    docker-compose up -d
    
    print_success "所有服务启动完成"
}

# 函数：显示服务状态
show_status() {
    print_info "服务状态："
    docker-compose ps
    
    echo
    print_info "服务访问地址："
    echo "  - 前端管理后台: http://localhost:5173"
    echo "  - API 网关:     http://localhost:8080"
    echo "  - 用户服务:     http://localhost:8081"
    echo "  - 租户服务:     http://localhost:8082"
    echo "  - MySQL:       localhost:3306"
    echo "  - Redis:       localhost:6379"
    
    echo
    print_info "健康检查："
    echo "  - API Health:  http://localhost:8080/api/v1/health"
    echo "  - 前端 Health: http://localhost:5173/health"
}

# 函数：显示日志
show_logs() {
    print_info "显示服务日志..."
    if [ $# -eq 0 ]; then
        docker-compose logs -f
    else
        docker-compose logs -f "$@"
    fi
}

# 函数：停止服务
stop_services() {
    print_info "停止所有服务..."
    docker-compose down
    print_success "服务已停止"
}

# 函数：重启服务
restart_services() {
    print_info "重启所有服务..."
    docker-compose restart
    print_success "服务已重启"
}

# 函数：清理环境
cleanup() {
    print_warning "这将删除所有容器、网络和卷！"
    read -p "确定要清理环境吗？(y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "清理 Docker 环境..."
        docker-compose down -v --remove-orphans
        docker system prune -f
        print_success "环境清理完成"
    else
        print_info "清理已取消"
    fi
}

# 函数：显示帮助信息
show_help() {
    echo "MER System Docker 开发环境管理脚本"
    echo
    echo "用法: $0 [命令]"
    echo
    echo "命令:"
    echo "  start     启动所有服务"
    echo "  stop      停止所有服务"
    echo "  restart   重启所有服务"
    echo "  status    显示服务状态"
    echo "  logs      显示服务日志"
    echo "  cleanup   清理环境（删除所有容器和卷）"
    echo "  help      显示此帮助信息"
    echo
    echo "示例:"
    echo "  $0 start                    # 启动所有服务"
    echo "  $0 logs gateway             # 显示网关服务日志"
    echo "  $0 logs gateway user-service # 显示多个服务日志"
}

# 主逻辑
main() {
    case "${1:-start}" in
        "start")
            check_docker
            check_env_file
            create_directories
            check_ports
            start_services
            show_status
            ;;
        "stop")
            stop_services
            ;;
        "restart")
            restart_services
            show_status
            ;;
        "status")
            show_status
            ;;
        "logs")
            shift
            show_logs "$@"
            ;;
        "cleanup")
            cleanup
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            print_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

# 脚本入口
main "$@"