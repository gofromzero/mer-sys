# API Specification

## REST API Specification

```yaml
openapi: 3.0.0
info:
  title: 多租户商户管理SaaS系统 API
  version: 1.0.0
  description: 支持三层B2B2C架构的完整业务API

servers:
  - url: https://api.mer-demo.com/v1
    description: 生产环境API服务器

paths:
  /auth/login:
    post:
      summary: 用户登录
      tags: [Authentication]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                username:
                  type: string
                password:
                  type: string
                tenant_id:
                  type: string
      responses:
        '200':
          description: 登录成功
          content:
            application/json:
              schema:
                type: object
                properties:
                  access_token:
                    type: string
                  user:
                    $ref: '#/components/schemas/User'

  /users:
    get:
      summary: 获取用户列表
      tags: [Users]
      security:
        - bearerAuth: []
      responses:
        '200':
          description: 用户列表
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/User'

  /merchants:
    get:
      summary: 获取商户列表
      tags: [Merchants]
      security:
        - bearerAuth: []
      responses:
        '200':
          description: 商户列表

  /products:
    get:
      summary: 获取商品列表
      tags: [Products]
      parameters:
        - in: query
          name: merchant_id
          schema:
            type: string
      responses:
        '200':
          description: 商品列表

  /orders:
    post:
      summary: 创建订单
      tags: [Orders]
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                merchant_id:
                  type: string
                items:
                  type: array
                  items:
                    type: object
                    properties:
                      product_id:
                        type: string
                      quantity:
                        type: integer
      responses:
        '201':
          description: 订单创建成功

components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
      
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
        username:
          type: string
        email:
          type: string
        status:
          type: string
```
