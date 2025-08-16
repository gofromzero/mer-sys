# Multi-stage build for React application

# Build stage
FROM node:18-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 package 文件
COPY package*.json ./

# 安装依赖
RUN npm ci --only=production && npm cache clean --force

# 复制源代码
COPY . .

# 构建应用
RUN npm run build

# Development stage
FROM node:18-alpine AS development

# 安装开发工具
RUN apk add --no-cache git curl bash

# 设置工作目录
WORKDIR /app

# 复制 package 文件
COPY package*.json ./

# 安装所有依赖（包括开发依赖）
RUN npm install

# 暴露端口
EXPOSE 5173

# 开发命令
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]

# Production stage
FROM nginx:alpine AS production

# 安装必要的包
RUN apk --no-cache add curl

# 复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制 nginx 配置
COPY nginx.conf /etc/nginx/nginx.conf

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置权限
RUN chown -R appuser:appgroup /usr/share/nginx/html
RUN chown -R appuser:appgroup /var/cache/nginx
RUN chown -R appuser:appgroup /var/log/nginx
RUN chown -R appuser:appgroup /etc/nginx/conf.d
RUN touch /var/run/nginx.pid
RUN chown -R appuser:appgroup /var/run/nginx.pid

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 80

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:80 || exit 1

# 启动 nginx
CMD ["nginx", "-g", "daemon off;"]