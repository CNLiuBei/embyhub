# Emby Hub Docker 镜像
# 多阶段构建 - 使用国内源加速

# ===== 阶段1: 构建前端 =====
FROM node:18-alpine AS frontend-builder

# 使用国内 Alpine 镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci --registry=https://registry.npmmirror.com
COPY frontend/ ./
RUN npm run build

# ===== 阶段2: 构建后端 =====
FROM golang:1.23-alpine AS backend-builder

# 使用国内 Alpine 镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装构建依赖
RUN apk add --no-cache git

# 设置 Go 代理（国内）
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn

WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /embyhub ./cmd/...

# ===== 阶段3: 最终镜像 =====
FROM alpine:3.19

# 使用国内 Alpine 镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

WORKDIR /app

# 复制后端可执行文件
COPY --from=backend-builder /embyhub ./embyhub

# 复制前端静态文件
COPY --from=frontend-builder /app/frontend/dist ./frontend

# 复制数据库脚本
COPY database/ ./database/

# 创建必要目录
RUN mkdir -p config logs

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/setup/status || exit 1

# 启动命令
CMD ["./embyhub"]
