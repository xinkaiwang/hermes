# 构建阶段
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git make

# 复制 go.mod 和 go.sum 先下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN go build -ldflags "-X github.com/xinkaiwang/hermes/internal/common.Version=$(cat VERSION) -X github.com/xinkaiwang/hermes/internal/common.GitCommit=$(git rev-parse --short HEAD || echo 'unknown') -X github.com/xinkaiwang/hermes/internal/common.BuildTime=$(date -u '+%Y-%m-%d_%H:%M:%S')" -o bin/hermes ./service/hermes

# 使用轻量级基础镜像
FROM alpine:latest

# 安装运行时依赖和调试工具
RUN apk add --no-cache ca-certificates tzdata bash vim curl wget bind-tools tcpdump htop busybox-extras

# 设置 bash 为默认 shell
SHELL ["/bin/bash", "-c"]

# 设置默认时区为西雅图（美国太平洋时区）
ENV TZ=America/Los_Angeles

# 创建非 root 用户
RUN adduser -D -u 1000 appuser

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/bin/hermes .

# 使用非 root 用户运行
USER appuser

# 设置默认环境变量
ENV API_PORT=8080 \
    METRICS_PORT=9090 \
    LOG_LEVEL=info \
    LOG_FORMAT=json

# 暴露端口
EXPOSE ${API_PORT} ${METRICS_PORT}

# 设置健康检查
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget --no-verbose --tries=1 --spider http://localhost:${API_PORT}/api/ping || exit 1

# 运行应用
CMD ["./hermes"] 