.PHONY: help build run test clean docker-build docker-run lint fmt migrate dev

# 变量定义
APP_NAME=tanzanite-api
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-w -s -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"
DOCKER_IMAGE=$(APP_NAME):$(VERSION)
DOCKER_REGISTRY?=ghcr.io/your-org

# Go相关变量
GO=go
GOFLAGS=-v
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

# 输出目录
BIN_DIR=bin
COVERAGE_DIR=coverage

help: ## 显示帮助信息
	@echo "Tanzanite API - Makefile命令"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "\033[36m\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ 开发命令

build: ## 编译应用
	@echo "Building $(APP_NAME) $(VERSION)..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/server ./cmd/server
	@echo "Build complete: $(BIN_DIR)/server"

run: ## 运行应用
	@echo "Running $(APP_NAME)..."
	$(GO) run ./cmd/server

dev: ## 开发模式（使用air热重载）
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "air not installed. Installing..."; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

test: ## 运行所有测试
	@echo "Running tests..."
	@mkdir -p $(COVERAGE_DIR)
	$(GO) test -v -race -coverprofile=$(COVERAGE_DIR)/coverage.txt -covermode=atomic ./...
	@echo "Coverage report: $(COVERAGE_DIR)/coverage.txt"

test-coverage: test ## 生成测试覆盖率报告
	$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.txt -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report: $(COVERAGE_DIR)/coverage.html"

test-integration: ## 运行集成测试
	@echo "Running integration tests..."
	$(GO) test -v -tags=integration ./...

bench: ## 运行基准测试
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem ./...

lint: ## 运行代码检查
	@echo "Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "golangci-lint not installed. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run --timeout=5m; \
	fi

fmt: ## 格式化代码
	@echo "Formatting code..."
	$(GO) fmt ./...
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	fi

vet: ## 运行go vet
	@echo "Running go vet..."
	$(GO) vet ./...

tidy: ## 整理依赖
	@echo "Tidying dependencies..."
	$(GO) mod tidy
	$(GO) mod verify

##@ 数据库命令

migrate-up: ## 运行数据库迁移（向上）
	@echo "Running migrations up..."
	$(GO) run ./cmd/server --migrate

migrate-down: ## 回滚数据库迁移
	@echo "Rolling back migrations..."
	$(GO) run ./cmd/server --migrate-down

migrate-status: ## 查看迁移状态
	@echo "Migration status..."
	$(GO) run ./cmd/server --migrate-status

##@ Docker命令

docker-build: ## 构建Docker镜像
	@echo "Building Docker image: $(DOCKER_IMAGE)..."
	docker build -t $(DOCKER_IMAGE) \
		--build-arg VERSION=$(VERSION) \
		-f Dockerfile .
	@echo "Docker image built: $(DOCKER_IMAGE)"

docker-build-prod: ## 构建生产Docker镜像
	@echo "Building production Docker image..."
	docker build -t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE) \
		--build-arg VERSION=$(VERSION) \
		--platform linux/amd64 \
		-f Dockerfile .

docker-push: docker-build-prod ## 推送Docker镜像到registry
	@echo "Pushing Docker image to $(DOCKER_REGISTRY)..."
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE)

docker-run: ## 运行Docker容器
	@echo "Running Docker container..."
	docker run --rm -it \
		-p 9000:9000 \
		-e SERVER_MODE=debug \
		-e DB_HOST=host.docker.internal \
		-e REDIS_HOST=host.docker.internal \
		--name $(APP_NAME) \
		$(DOCKER_IMAGE)

docker-compose-up: ## 使用docker-compose启动所有服务
	docker-compose up -d

docker-compose-down: ## 停止docker-compose服务
	docker-compose down

docker-compose-logs: ## 查看docker-compose日志
	docker-compose logs -f

##@ Kubernetes命令

k8s-deploy: ## 部署到Kubernetes
	@echo "Deploying to Kubernetes..."
	kubectl apply -f k8s/deployment.yaml
	kubectl apply -f k8s/hpa.yaml
	kubectl apply -f k8s/ingress.yaml

k8s-delete: ## 从Kubernetes删除
	@echo "Deleting from Kubernetes..."
	kubectl delete -f k8s/ingress.yaml
	kubectl delete -f k8s/hpa.yaml
	kubectl delete -f k8s/deployment.yaml

k8s-status: ## 查看Kubernetes部署状态
	@echo "Kubernetes status..."
	kubectl get pods -l app=$(APP_NAME)
	kubectl get svc $(APP_NAME)
	kubectl get hpa $(APP_NAME)-hpa

k8s-logs: ## 查看Kubernetes日志
	@echo "Fetching logs..."
	kubectl logs -f deployment/$(APP_NAME) --tail=100

k8s-describe: ## 查看Pod详情
	@echo "Describing pods..."
	kubectl describe pods -l app=$(APP_NAME)

##@ 清理命令

clean: ## 清理构建产物
	@echo "Cleaning..."
	rm -rf $(BIN_DIR)
	rm -rf $(COVERAGE_DIR)
	rm -f server
	$(GO) clean -cache -testcache

clean-docker: ## 清理Docker镜像
	@echo "Cleaning Docker images..."
	docker rmi $(DOCKER_IMAGE) || true

clean-all: clean clean-docker ## 清理所有

##@ 工具命令

install-tools: ## 安装开发工具
	@echo "Installing development tools..."
	$(GO) install github.com/cosmtrek/air@latest
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install golang.org/x/tools/cmd/goimports@latest

gen-swagger: ## 生成Swagger文档
	@echo "Generating Swagger docs..."
	@if command -v swag > /dev/null; then \
		swag init -g cmd/server/main.go -o docs/swagger; \
	else \
		echo "swag not installed. Run: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

gen-mocks: ## 生成Mock代码
	@echo "Generating mocks..."
	@if command -v mockgen > /dev/null; then \
		go generate ./...; \
	else \
		echo "mockgen not installed. Run: go install github.com/golang/mock/mockgen@latest"; \
	fi

##@ 环境设置

setup-dev: install-tools ## 设置开发环境
	@echo "Setting up development environment..."
	cp config/config.example.yaml config/config.yaml
	@echo "Please update config/config.yaml with your settings"

setup-git-hooks: ## 设置Git hooks
	@echo "Setting up Git hooks..."
	@mkdir -p .git/hooks
	@echo '#!/bin/sh\nmake fmt lint test' > .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Git hooks installed"

##@ 信息命令

version: ## 显示版本信息
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(shell go version)"
	@echo "GOOS: $(GOOS)"
	@echo "GOARCH: $(GOARCH)"

deps: ## 显示依赖树
	@echo "Dependency tree..."
	$(GO) mod graph

deps-update: ## 更新所有依赖
	@echo "Updating dependencies..."
	$(GO) get -u ./...
	$(GO) mod tidy

##@ 性能分析

profile-cpu: ## CPU性能分析
	@echo "Running CPU profiling..."
	$(GO) test -cpuprofile=cpu.prof -bench=. ./...
	$(GO) tool pprof -http=:8080 cpu.prof

profile-mem: ## 内存性能分析
	@echo "Running memory profiling..."
	$(GO) test -memprofile=mem.prof -bench=. ./...
	$(GO) tool pprof -http=:8080 mem.prof

##@ CI/CD

ci: lint test build ## 运行CI流程
	@echo "CI pipeline completed successfully!"

cd: docker-build docker-push k8s-deploy ## 运行CD流程
	@echo "CD pipeline completed successfully!"

.DEFAULT_GOAL := help
