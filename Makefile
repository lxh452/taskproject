# Root Makefile - Quick start for the task project

PROJECT := task-project
API_DIR := task
BIN_DIR := $(API_DIR)/bin
BIN := $(BIN_DIR)/taskprojectapi

.PHONY: help deps tidy build run clean gen api model test fmt test-auto test-quick test-stress test-phase

help:
	@echo "Available commands:"
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




