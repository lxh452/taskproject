## 任务管理系统

# ============ 构建阶段 ============
FROM golang:1.23-alpine AS builder

# 安装必要的构建工具
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# 设置 Go 环境变量
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# 先复制依赖文件，利用 Docker 缓存
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# 复制源码
COPY . .

# 编译二进制，添加版本信息
ARG VERSION=dev
ARG BUILD_TIME
RUN BUILD_TIME=${BUILD_TIME:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")} && \
    go build -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}" \
    -o /app/taskprojectapi ./task/taskprojectapi.go

# ============ 运行阶段 ============
FROM alpine:3.19 AS runtime

# 安装运行时依赖
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && apk del tzdata

# 创建非 root 用户
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser

WORKDIR /app

# 创建必要目录
RUN mkdir -p /app/logs /app/task/etc /app/task/internal/templates && \
    chown -R appuser:appgroup /app

# 从构建阶段复制文件
COPY --from=builder --chown=appuser:appgroup /app/taskprojectapi ./taskprojectapi
COPY --from=builder --chown=appuser:appgroup /build/task/etc ./task/etc
COPY --from=builder --chown=appuser:appgroup /build/task/internal/templates ./task/internal/templates

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8888

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8888/health || exit 1

# 启动命令
ENTRYPOINT ["./taskprojectapi"]
CMD ["-f", "task/etc/taskprojectapi.yaml"]
