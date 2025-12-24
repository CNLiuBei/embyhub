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

# 安装时区数据和 cloudflared
RUN apk --no-cache add ca-certificates tzdata \
    && apk add --no-cache --repository=https://dl-cdn.alpinelinux.org/alpine/edge/community cloudflared

# 设置时区
ENV TZ=Asia/Shanghai

# 从builder复制二进制文件
COPY --from=builder /app/server .

# 创建必要目录
RUN mkdir -p logs uploads/avatars config

EXPOSE 8080

CMD ["./server"]
