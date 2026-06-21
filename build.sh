#!/bin/bash

# ================================
# 云值守系统 - 本地构建镜像脚本
# ================================

set -e

export DOCKER_BUILDKIT=1

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 版本号
VERSION=$(date +%Y%m%d%H%M%S)
OUTPUT_DIR="./dist"

# ================= 目标服务器配置 =================
REMOTE_USER="root"
REMOTE_IP="47.108.52.145"
REMOTE_DIR="/opt/yzs"
# 注意：SSL 证书需手动上传到 /etc/nginx/ssl/，本脚本不会自动上传
# =================================================

# 清理旧文件
clean() {
    log_info "清理旧的构建文件..."
    rm -rf $OUTPUT_DIR
    mkdir -p $OUTPUT_DIR
}

# 构建后端镜像
build_backend() {
    log_info "构建后端镜像 yzs-backend:$VERSION..."
    docker build -t yzs-backend:$VERSION -t yzs-backend:latest ./backend
    
    log_info "导出后端镜像..."
    docker save yzs-backend:$VERSION -o $OUTPUT_DIR/yzs-backend.tar
    gzip -f $OUTPUT_DIR/yzs-backend.tar
    
    log_info "后端镜像构建完成: $OUTPUT_DIR/yzs-backend.tar.gz"
}

# 构建前端镜像
build_frontend() {
    log_info "构建前端镜像 yzs-frontend:$VERSION..."
    docker build -t yzs-frontend:$VERSION -t yzs-frontend:latest ./frontend
    
    log_info "导出前端镜像..."
    docker save yzs-frontend:$VERSION -o $OUTPUT_DIR/yzs-frontend.tar
    gzip -f $OUTPUT_DIR/yzs-frontend.tar
    
    log_info "前端镜像构建完成: $OUTPUT_DIR/yzs-frontend.tar.gz"
}

# 复制配置文件
copy_configs() {
    log_info "复制配置文件..."
    cp nginx.conf $OUTPUT_DIR/
    if [[ ! -f .env ]]; then
        log_error ".env 文件不存在，请先创建（参考 .env.example）"
        exit 1
    fi
    cp .env $OUTPUT_DIR/
    
    # 创建 docker-compose.yml
    cat > $OUTPUT_DIR/docker-compose.yml << 'EOF'
services:
  mysql:
    image: mysql:8.1.0
    container_name: yzs-mysql
    restart: always
    pull_policy: never
    ports:
      - "3306:3306"
    environment:
      MYSQL_ROOT_PASSWORD: ${DB_PASSWORD:-123@123qwe}
      MYSQL_DATABASE: ${DB_NAME:-go_yzs}
    volumes:
      - mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "127.0.0.1", "-uroot", "-p${DB_PASSWORD:-123@123qwe}"]
      interval: 5s
      timeout: 5s
      retries: 10

  redis:
    image: redis:7.2
    container_name: yzs-redis
    restart: always
    pull_policy: never
    command: redis-server --maxmemory 256mb --maxmemory-policy allkeys-lru
    volumes:
      - redis_data:/data

  nginx:
    image: nginx:alpine
    container_name: yzs-nginx
    restart: always
    pull_policy: never
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/conf.d/default.conf:ro
      - /etc/nginx/ssl:/etc/nginx/ssl:ro
    extra_hosts:
      - "host.docker.internal:host-gateway"
    depends_on:
      - frontend
      - backend

  backend:
    image: yzs-backend:latest
    container_name: yzs-backend
    restart: always
    pull_policy: never
    env_file:
      - .env
    environment:
      REDIS_HOST: yzs-redis
      REDIS_PORT: 6379
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_started

  frontend:
    image: yzs-frontend:latest
    container_name: yzs-frontend
    restart: always
    pull_policy: never
    depends_on:
      - backend

volumes:
  mysql_data:
  redis_data:
EOF

    log_info "配置文件复制完成"
}

# 创建服务器端加载脚本
create_load_script() {
    cat > $OUTPUT_DIR/load.sh << 'EOF'
#!/bin/bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 检查 root 权限
if [[ $EUID -ne 0 ]]; then
    log_error "需要 root 权限运行"
    exit 1
fi

log_info "加载 Docker 镜像..."
BACKEND_LOADED=$(gunzip -c yzs-backend.tar.gz | docker load | grep -oP '(?<=Loaded image: ).*')
FRONTEND_LOADED=$(gunzip -c yzs-frontend.tar.gz | docker load | grep -oP '(?<=Loaded image: ).*')

log_info "镜像加载完成: $BACKEND_LOADED, $FRONTEND_LOADED"

log_info "设置 latest 标签..."
docker tag "$BACKEND_LOADED" yzs-backend:latest
docker tag "$FRONTEND_LOADED" yzs-frontend:latest

docker images | grep yzs

log_info "停止旧容器..."
docker compose down 2>/dev/null || true

log_info "启动服务..."
mkdir -p /etc/nginx/ssl
docker compose up -d

log_info "清理旧镜像..."
docker image prune -f
OLD_BACKEND=$(docker images yzs-backend --format '{{.Repository}}:{{.Tag}}' | grep -v latest | grep -v "$BACKEND_LOADED" || true)
OLD_FRONTEND=$(docker images yzs-frontend --format '{{.Repository}}:{{.Tag}}' | grep -v latest | grep -v "$FRONTEND_LOADED" || true)
for img in $OLD_BACKEND $OLD_FRONTEND; do
    docker rmi "$img" 2>/dev/null && log_info "已删除旧镜像: $img" || true
done
docker image prune -f

log_info "部署完成！"
docker compose ps
EOF

    chmod +x $OUTPUT_DIR/load.sh
    log_info "加载脚本创建完成: $OUTPUT_DIR/load.sh"
}

# 打包所有文件
package() {
    log_info "打包发布文件..."
    tar -czvf yzs-deploy-$VERSION.tar.gz -C $OUTPUT_DIR .
    log_info "发布包创建完成: yzs-deploy-$VERSION.tar.gz"
    
    # 显示文件大小
    echo ""
    echo "=========================================="
    echo "构建完成！"
    echo "=========================================="
    echo ""
    ls -lh $OUTPUT_DIR/
    ls -lh yzs-deploy-$VERSION.tar.gz
    echo ""
}

# 上传发布文件到服务器（依赖 SSH 免密；SSL 证书需手动上传，此处不处理）
upload() {
    log_info "上传发布文件到 ${REMOTE_USER}@${REMOTE_IP}:${REMOTE_DIR}/ ..."
    ssh "${REMOTE_USER}@${REMOTE_IP}" "mkdir -p ${REMOTE_DIR}"
    scp -C -r \
        "$OUTPUT_DIR/nginx.conf" \
        "$OUTPUT_DIR/.env" \
        "$OUTPUT_DIR/docker-compose.yml" \
        "$OUTPUT_DIR/yzs-backend.tar.gz" \
        "$OUTPUT_DIR/yzs-frontend.tar.gz" \
        "$OUTPUT_DIR/load.sh" \
        "${REMOTE_USER}@${REMOTE_IP}:${REMOTE_DIR}/"
    ssh "${REMOTE_USER}@${REMOTE_IP}" "chmod +x ${REMOTE_DIR}/load.sh"
    log_info "上传完成"
}

# 远程执行 load.sh 完成部署
deploy() {
    log_info "在远程服务器上执行 load.sh 进行部署..."
    ssh "${REMOTE_USER}@${REMOTE_IP}" "cd ${REMOTE_DIR} && ./load.sh"

    echo ""
    echo "=========================================="
    echo "  ✓ 部署完成！"
    echo "=========================================="
    echo ""
    echo "提示: (首次) SSL 证书需手动上传:"
    echo "  scp your.pem your.key ${REMOTE_USER}@${REMOTE_IP}:/etc/nginx/ssl/"
    echo ""
}

# 主函数
main() {
    echo ""
    echo "=========================================="
    echo "  云值守系统 - 本地构建"
    echo "=========================================="
    echo ""

    # 检查 Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装"
        exit 1
    fi

    clean
    build_backend
    build_frontend
    copy_configs
    create_load_script
    package
    upload
    deploy
}

main "$@"