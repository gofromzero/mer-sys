# Coding Standards

## Critical Fullstack Rules

- **Type Sharing:** 所有数据类型必须定义在shared包中，前后端从统一位置导入
- **API Calls:** 前端禁止直接HTTP调用，必须通过service层封装
- **Environment Variables:** 配置访问只能通过config对象，禁止直接使用process.env
- **Error Handling:** 所有API路由必须使用统一的错误处理器
- **State Updates:** 禁止直接变更state，必须使用proper的状态管理模式
- **Database Access:** 后端必须通过Repository模式访问数据库，禁止直接SQL操作
- **Tenant Isolation:** 所有数据查询必须自动添加tenant_id过滤，防止数据泄露
- **Permission Checks:** 所有敏感操作必须进行权限验证，使用中间件自动检查

## Naming Conventions

| Element | Frontend | Backend | Example |
|---------|----------|---------|---------|
| Components | PascalCase | - | `UserProfile.tsx` |
| Hooks | camelCase with 'use' | - | `useAuth.ts` |
| API Routes | - | kebab-case | `/api/user-profile` |
| Database Tables | - | snake_case | `user_profiles` |
| Go Packages | - | lowercase | `userservice` |
| Go Functions | - | PascalCase | `CreateUser` |
| React Props | camelCase | - | `onUserUpdate` |
