# 企业任务交接与派发系统 API 生成脚本

# 生成统一的API服务
all:
	goctl api go -api ./task_project.api -dir . --style=goZero

# 清理生成的文件
clean:
	rm -rf ./internal/handler/* ./internal/logic/* ./internal/types/* ./internal/svc/* ./etc/*.yaml *.go

# 重新生成
rebuild: clean all

# 帮助信息
help:
	@echo "可用的命令:"
	@echo "  make all      - 生成统一API服务"
	@echo "  make clean    - 清理所有生成的文件"
	@echo "  make rebuild  - 清理后重新生成"
	@echo "  make help     - 显示此帮助信息"

