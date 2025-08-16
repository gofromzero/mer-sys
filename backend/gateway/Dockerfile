# Multi-stage build for Go application

# Build stage
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

# 复制 Go modules 文件
COPY go.work go.work.sum ./
COPY shared/go.mod shared/go.sum ./shared/
COPY gateway/go.mod gateway/go.sum ./gateway/

# 下载依赖
RUN go mod download

# 复制源代码
COPY shared/ ./shared/
COPY gateway/ ./gateway/

# 构建应用
WORKDIR /app/gateway
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Development stage
FROM golang:1.21-alpine AS development

# 安装开发工具
RUN apk add --no-cache git ca-certificates tzdata curl bash

# 安装 air for hot reload
RUN go install github.com/cosmtrek/air@latest

# 设置工作目录
WORKDIR /app

# 设置时区
ENV TZ=Asia/Shanghai

# 开发环境端口
EXPOSE 8080

# 开发命令将通过 docker-compose 覆盖
CMD ["go", "run", "main.go"]

# Production stage
FROM alpine:latest AS production

# 安装必要的包
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/gateway/main .
COPY --from=builder /app/config ./config

# 改变文件所有权
RUN chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/api/v1/health || exit 1

# 运行应用
CMD ["./main"]