## 任务管理系统 - 后端 API 最简 Dockerfile
## 说明：
## - 只打包 Go 后端服务本身
## - MySQL / RabbitMQ / Redis / MongoDB 都使用外部服务（通过配置文件连接）

# ============ 构建阶段 ============
FROM golang:1.22-alpine AS builder

WORKDIR /build

ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# 先复制依赖文件，加速构建缓存
COPY go.mod ./
COPY go.sum* ./
RUN go mod download

# 再复制全部源码
COPY . .

# 编译后端二进制
RUN go build -ldflags="-s -w" -o /app/taskprojectapi ./task/taskprojectapi.go

# ============ 运行阶段 ============
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 拷贝编译好的二进制
COPY --from=builder /app/taskprojectapi ./taskprojectapi

# 拷贝配置目录（你可以在打包前修改 task/etc/taskprojectapi.yaml，把 MySQL/RabbitMQ 主机改成外部地址）
COPY --from=builder /build/task/etc ./task/etc

# 暴露 HTTP 端口
EXPOSE 8888

# 启动命令：
# -f 参数指定使用容器内的配置文件 task/etc/taskprojectapi.yaml
CMD ["./taskprojectapi", "-f", "task/etc/taskprojectapi.yaml"]




