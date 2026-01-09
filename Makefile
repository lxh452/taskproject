# Root Makefile - Quick start for the task project

PROJECT := task-project
API_DIR := task
BIN_DIR := $(API_DIR)/bin
BIN := $(BIN_DIR)/taskprojectapi

.PHONY: help deps tidy build run clean gen api model test fmt test-auto test-quick test-stress test-phase \
       docker-build docker-run docker-stop docker-logs docker-shell docker-push docker-clean \
       docker-up docker-down docker-restart docker-ps

# Docker 配置
DOCKER_IMAGE := task-project-backend
DOCKER_TAG := latest
DOCKER_REGISTRY := # 你的镜像仓库地址，例如: registry.cn-hangzhou.aliyuncs.com/yournamespace

help:
	@echo "Available commands:"
	@echo ""
	@echo "=== 开发命令 ==="
	@echo "  make deps     - Download dependencies"
	@echo "  make tidy     - Tidy go.mod"
	@echo "  make build    - Build API binary"
	@echo "  make run      - Run API service"
	@echo "  make clean    - Clean build artifacts"
	@echo "  make api      - Generate API code (goctl)"
	@echo "  make model    - Generate model code (delegates to model/Makefile)"
	@echo "  make gen      - Generate API + model code"
	@echo "  make test     - Run unit tests"
	@echo "  make fmt      - go fmt"
	@echo ""
	@echo "=== Docker 命令 ==="
	@echo "  make docker-build     - 构建 Docker 镜像"
	@echo "  make docker-run       - 单独运行后端容器"
	@echo "  make docker-stop      - 停止后端容器"
	@echo "  make docker-logs      - 查看后端容器日志"
	@echo "  make docker-shell     - 进入后端容器 shell"
	@echo "  make docker-push      - 推送镜像到仓库"
	@echo "  make docker-clean     - 清理 Docker 镜像和缓存"
	@echo ""
	@echo "=== Docker Compose 命令 ==="
	@echo "  make docker-up        - 启动所有服务 (后台)"
	@echo "  make docker-down      - 停止并移除所有服务"
	@echo "  make docker-restart   - 重启所有服务"
	@echo "  make docker-ps        - 查看服务状态"
	@echo ""
	@echo "=== 测试命令 ==="
	@echo "  make test-auto    - Run full automated tests"
	@echo "  make test-quick   - Run quick automated tests"
	@echo "  make test-stress  - Run stress tests"
	@echo "  make test-phase   - Run specific phase tests"

deps:
	@cd $(API_DIR) && go mod download
	@cd model && go mod download

tidy:
	@cd $(API_DIR) && go mod tidy
	@cd model && go mod tidy

build:
	@mkdir -p $(BIN_DIR)
	@cd $(API_DIR) && go build -o ../$(BIN) ./taskprojectapi.go
	@echo "Built $(BIN)"

run:
	@cd $(API_DIR) && go run taskprojectapi.go -f etc/taskprojectapi.yaml

clean:
	@rm -rf $(BIN_DIR)
	@find . -name "*.log" -delete

api:
	@cd $(API_DIR) && goctl api go -api task_project.api -dir . --style=goZero

model:
	@cd model && make gen

gen: api model

test:
	@cd $(API_DIR) && go test ./... -v
	@cd model && go test ./... -v

fmt:
	@go fmt ./...

# 自动化测试目标
test-auto:
	@echo "运行完整自动化测试..."
	@if command -v newman >/dev/null 2>&1; then \
		newman run postman_automated_workflow.json \
			--environment postman_env.json \
			--reporters cli,html,json \
			--reporter-html-export test_reports/full_test_$(shell date +%Y%m%d_%H%M%S).html \
			--reporter-json-export test_reports/full_test_$(shell date +%Y%m%d_%H%M%S).json \
			--timeout-request 30000 \
			--timeout-script 30000 \
			--delay-request 1000 \
			--verbose; \
	else \
		echo "错误: Newman 未安装，请运行: npm install -g newman"; \
		exit 1; \
	fi

test-quick:
	@echo "运行快速自动化测试..."
	@if command -v newman >/dev/null 2>&1; then \
		newman run postman_automated_workflow.json \
			--environment postman_env.json \
			--reporters cli,html \
			--reporter-html-export test_reports/quick_test_$(shell date +%Y%m%d_%H%M%S).html \
			--timeout-request 15000 \
			--timeout-script 15000 \
			--delay-request 500 \
			--iteration-count 1; \
	else \
		echo "错误: Newman 未安装，请运行: npm install -g newman"; \
		exit 1; \
	fi

test-stress:
	@echo "运行压力测试..."
	@if command -v newman >/dev/null 2>&1; then \
		newman run postman_automated_workflow.json \
			--environment postman_env.json \
			--reporters cli,html,json \
			--reporter-html-export test_reports/stress_test_$(shell date +%Y%m%d_%H%M%S).html \
			--reporter-json-export test_reports/stress_test_$(shell date +%Y%m%d_%H%M%S).json \
			--timeout-request 60000 \
			--timeout-script 60000 \
			--delay-request 100 \
			--iteration-count 5; \
	else \
		echo "错误: Newman 未安装，请运行: npm install -g newman"; \
		exit 1; \
	fi

test-phase:
	@echo "运行阶段测试: $(PHASE)"
	@if [ -z "$(PHASE)" ]; then \
		echo "错误: 请指定测试阶段，例如: make test-phase PHASE=\"Phase 1\""; \
		exit 1; \
	fi
	@if command -v newman >/dev/null 2>&1; then \
		newman run postman_automated_workflow.json \
			--environment postman_env.json \
			--reporters cli,html \
			--reporter-html-export test_reports/phase_$(PHASE)_$(shell date +%Y%m%d_%H%M%S).html \
			--timeout-request 20000 \
			--timeout-script 20000 \
			--delay-request 1000 \
			--folder "$(PHASE)"; \
	else \
		echo "错误: Newman 未安装，请运行: npm install -g newman"; \
		exit 1; \
	fi






# ============ Docker 命令 ============

# 构建 Docker 镜像
docker-build:
	@echo "构建 Docker 镜像: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	docker build \
		--build-arg VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev") \
		--build-arg BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ") \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		-f Dockerfile .
	@echo "镜像构建完成: $(DOCKER_IMAGE):$(DOCKER_TAG)"

# 构建镜像 (无缓存)
docker-build-nocache:
	@echo "构建 Docker 镜像 (无缓存): $(DOCKER_IMAGE):$(DOCKER_TAG)"
	docker build --no-cache \
		--build-arg VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev") \
		--build-arg BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ") \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		-f Dockerfile .

# 单独运行后端容器 (需要外部数据库服务)
docker-run:
	@echo "启动后端容器..."
	docker run -d \
		--name task_project_backend \
		--network task_project_network \
		-p 8888:8888 \
		-v $(PWD)/logs/backend:/app/logs \
		-e TZ=Asia/Shanghai \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

# 停止后端容器
docker-stop:
	@echo "停止后端容器..."
	docker stop task_project_backend 2>/dev/null || true
	docker rm task_project_backend 2>/dev/null || true

# 查看后端容器日志
docker-logs:
	docker logs -f task_project_backend

# 查看最近 100 行日志
docker-logs-tail:
	docker logs --tail 100 -f task_project_backend

# 进入后端容器 shell
docker-shell:
	docker exec -it task_project_backend /bin/sh

# 推送镜像到仓库
docker-push:
ifdef DOCKER_REGISTRY
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "镜像已推送: $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)"
else
	@echo "错误: 请设置 DOCKER_REGISTRY 变量"
	@echo "示例: make docker-push DOCKER_REGISTRY=registry.cn-hangzhou.aliyuncs.com/yournamespace"
endif

# 清理 Docker 镜像和缓存
docker-clean:
	@echo "清理 Docker 资源..."
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true
	docker builder prune -f

# ============ Docker Compose 命令 ============

# 启动所有服务 (后台运行)
docker-up:
	@echo "启动所有服务..."
	docker-compose up -d --build
	@echo "服务启动完成，使用 'make docker-ps' 查看状态"

# 启动所有服务 (前台运行，显示日志)
docker-up-fg:
	docker-compose up --build

# 停止并移除所有服务
docker-down:
	@echo "停止所有服务..."
	docker-compose down

# 停止并移除所有服务 (包括数据卷)
docker-down-v:
	@echo "停止所有服务并删除数据卷..."
	docker-compose down -v

# 重启所有服务
docker-restart:
	@echo "重启所有服务..."
	docker-compose restart

# 只重启后端服务
docker-restart-backend:
	@echo "重启后端服务..."
	docker-compose restart backend

# 重新构建并启动后端服务
docker-rebuild-backend:
	@echo "重新构建并启动后端服务..."
	docker-compose up -d --build backend

# 查看服务状态
docker-ps:
	docker-compose ps

# 查看所有服务日志
docker-compose-logs:
	docker-compose logs -f

# 查看后端服务日志
docker-backend-logs:
	docker-compose logs -f backend

# 查看 MySQL 日志
docker-mysql-logs:
	docker-compose logs -f mysql

# 进入 MySQL 容器
docker-mysql-shell:
	docker-compose exec mysql mysql -u root -p

# 进入 Redis 容器
docker-redis-shell:
	docker-compose exec redis redis-cli
