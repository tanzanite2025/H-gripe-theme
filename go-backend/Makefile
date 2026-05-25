.PHONY: help build run test clean docker-build docker-up docker-down migrate lint fmt vet install dev

# 变量定义
APP_NAME=tanzanite-backend
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.3.0")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION=$(shell go version | awk '{print $$3}')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# 默认目标
help: ## 显示帮助信息
	@echo "可用的命令:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

install: ## 安装依赖
	@echo "安装 Go 依赖..."
	go mod download
	go mod tidy
	@echo "依赖安装完成!"

build: ## 编译应用
	@echo "编译 $(APP_NAME)..."
	go build $(LDFLAGS) -o bin/$(APP_NAME) cmd/server/main.go
	@echo "编译完成: bin/$(APP_NAME)"

build-linux: ## 编译 Linux 版本
	@echo "编译 Linux 版本..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-linux-amd64 cmd/server/main.go
	@echo "编译完成: bin/$(APP_NAME)-linux-amd64"

build-windows: ## 编译 Windows 版本
	@echo "编译 Windows 版本..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-windows-amd64.exe cmd/server/main.go
	@echo "编译完成: bin/$(APP_NAME)-windows-amd64.exe"

build-mac: ## 编译 macOS 版本
	@echo "编译 macOS 版本..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-darwin-amd64 cmd/server/main.go
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(APP_NAME)-darwin-arm64 cmd/server/main.go
	@echo "编译完成: bin/$(APP_NAME)-darwin-*"

build-all: build-linux build-windows build-mac ## 编译所有平台版本

run: ## 运行应用
	@echo "启动 $(APP_NAME)..."
	go run cmd/server/main.go

dev: ## 开发模式 (使用 air 热重载)
	@echo "启动开发模式..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "air 未安装，请运行: go install github.com/cosmtrek/air@latest"; \
		echo "或直接运行: make run"; \
	fi

test: ## 运行测试
	@echo "运行测试..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo "测试完成!"

test-coverage: test ## 生成测试覆盖率报告
	@echo "生成覆盖率报告..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告: coverage.html"

benchmark: ## 运行性能测试
	@echo "运行性能测试..."
	go test -bench=. -benchmem ./...

lint: ## 运行代码检查
	@echo "运行 golangci-lint..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint 未安装，请访问: https://golangci-lint.run/usage/install/"; \
	fi

fmt: ## 格式化代码
	@echo "格式化代码..."
	go fmt ./...
	@echo "格式化完成!"

vet: ## 运行 go vet
	@echo "运行 go vet..."
	go vet ./...
	@echo "检查完成!"

clean: ## 清理构建文件
	@echo "清理构建文件..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "清理完成!"

migrate: ## 运行数据库迁移
	@echo "运行数据库迁移..."
	go run cmd/server/main.go migrate
	@echo "迁移完成!"

migrate-down: ## 回滚数据库迁移
	@echo "回滚数据库迁移..."
	go run cmd/server/main.go migrate-down
	@echo "回滚完成!"

seed: ## 填充测试数据
	@echo "填充测试数据..."
	go run cmd/server/main.go seed
	@echo "数据填充完成!"

docker-build: ## 构建 Docker 镜像
	@echo "构建 Docker 镜像..."
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest
	@echo "镜像构建完成: $(APP_NAME):$(VERSION)"

docker-up: ## 启动 Docker Compose
	@echo "启动 Docker Compose..."
	docker-compose up -d
	@echo "服务已启动!"
	@echo "查看日志: docker-compose logs -f"

docker-down: ## 停止 Docker Compose
	@echo "停止 Docker Compose..."
	docker-compose down
	@echo "服务已停止!"

docker-logs: ## 查看 Docker 日志
	docker-compose logs -f

docker-ps: ## 查看 Docker 容器状态
	docker-compose ps

docker-restart: docker-down docker-up ## 重启 Docker 服务

swagger: ## 生成 Swagger 文档
	@echo "生成 Swagger 文档..."
	@if command -v swag > /dev/null; then \
		swag init -g cmd/server/main.go -o docs/swagger; \
	else \
		echo "swag 未安装，请运行: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

deps-update: ## 更新依赖
	@echo "更新依赖..."
	go get -u ./...
	go mod tidy
	@echo "依赖更新完成!"

deps-check: ## 检查依赖安全性
	@echo "检查依赖安全性..."
	@if command -v govulncheck > /dev/null; then \
		govulncheck ./...; \
	else \
		echo "govulncheck 未安装，请运行: go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi

security: ## 运行安全检查
	@echo "运行安全检查..."
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "gosec 未安装，请运行: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
	fi

deploy-staging: build-linux ## 部署到测试环境
	@echo "部署到测试环境..."
	@echo "TODO: 实现部署脚本"

deploy-production: build-linux ## 部署到生产环境
	@echo "部署到生产环境..."
	@echo "TODO: 实现部署脚本"

version: ## 显示版本信息
	@echo "应用名称: $(APP_NAME)"
	@echo "版本: $(VERSION)"
	@echo "构建时间: $(BUILD_TIME)"
	@echo "Go 版本: $(GO_VERSION)"

info: version ## 显示项目信息
	@echo ""
	@echo "项目结构:"
	@tree -L 2 -I 'vendor|node_modules|bin' || ls -la

setup: install ## 初始化项目
	@echo "初始化项目..."
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "已创建 .env 文件，请编辑配置"; \
	fi
	@mkdir -p bin uploads logs
	@echo "项目初始化完成!"
	@echo ""
	@echo "下一步:"
	@echo "  1. 编辑 .env 文件配置数据库和其他服务"
	@echo "  2. 运行 'make docker-up' 启动数据库和 Redis"
	@echo "  3. 运行 'make migrate' 初始化数据库"
	@echo "  4. 运行 'make run' 启动应用"

check: fmt vet lint test ## 运行所有检查

ci: check build ## CI 流程

all: clean install build test ## 完整构建流程

.DEFAULT_GOAL := help
