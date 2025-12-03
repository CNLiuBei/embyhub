.PHONY: build run clean test docker-build

# 编译
build:
	go build -o embyhub cmd/main.go

# 运行
run:
	go run cmd/main.go

# 清理
clean:
	rm -f embyhub
	rm -rf logs/*

# 测试
test:
	go test -v ./...

# 下载依赖
deps:
	go mod download
	go mod tidy

# Docker构建
docker-build:
	docker build -t embyhub-backend:latest .

# 格式化代码
fmt:
	go fmt ./...

# 代码检查
lint:
	golangci-lint run

# 开发模式（热重载需要安装air）
dev:
	air
