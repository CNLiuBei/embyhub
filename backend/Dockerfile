# 飞牛影视用户管理系统 - 后端 Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git

# 复制go.mod和go.sum
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# 运行镜像
FROM alpine:3.19

WORKDIR /app

# 安装时区数据
RUN apk --no-cache add ca-certificates tzdata curl

# 下载 cloudflared (支持 amd64 和 arm64)
RUN ARCH=$(uname -m) && \
    if [ "$ARCH" = "x86_64" ]; then \
        CLOUDFLARED_ARCH="amd64"; \
    elif [ "$ARCH" = "aarch64" ]; then \
        CLOUDFLARED_ARCH="arm64"; \
    else \
        CLOUDFLARED_ARCH="amd64"; \
    fi && \
    curl -L -o /usr/local/bin/cloudflared \
        "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-${CLOUDFLARED_ARCH}" && \
    chmod +x /usr/local/bin/cloudflared

# 设置时区
ENV TZ=Asia/Shanghai

# 从builder复制二进制文件
COPY --from=builder /app/server .

# 创建必要目录
RUN mkdir -p logs uploads/avatars config

# 复制默认配置文件（Docker环境专用）
COPY config/config.docker.yaml ./config/config.yaml

EXPOSE 8080

CMD ["./server"]
