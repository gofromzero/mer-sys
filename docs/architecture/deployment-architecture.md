# Deployment Architecture

## Deployment Strategy

**Frontend Deployment:**
- **Platform:** 阿里云OSS + CDN
- **Build Command:** npm run build
- **Output Directory:** dist/
- **CDN/Edge:** 阿里云CDN全球加速

**Backend Deployment:**
- **Platform:** 阿里云ECS + Docker
- **Build Command:** go build
- **Deployment Method:** Docker容器部署

## Environments

| Environment | Frontend URL | Backend URL | Purpose |
|-------------|--------------|-------------|---------|
| Development | http://localhost:3000 | http://localhost:8080 | 本地开发环境 |
| Staging | https://staging.mer-demo.com | https://api-staging.mer-demo.com | 预发布测试环境 |
| Production | https://mer-demo.com | https://api.mer-demo.com | 生产环境 |
