# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git

# 复制go.mod和go.sum
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建
RUN CGO_ENABLED=0 GOOS=linux go build -o embyhub cmd/main.go

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 安装必要工具
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 从构建阶段复制二进制文件
COPY --from=builder /app/embyhub .
COPY --from=builder /app/config ./config

# 创建日志目录
RUN mkdir -p logs

# 暴露端口
EXPOSE 8080

# 启动应用
CMD ["./embyhub"]
