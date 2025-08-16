# External APIs

## 阿里云短信服务 API

- **Purpose:** 发送验证码、通知短信
- **Documentation:** https://help.aliyun.com/product/44282.html
- **Base URL:** https://dysmsapi.aliyuncs.com
- **Authentication:** AccessKey/SecretKey签名认证
- **Rate Limits:** 1000条/天（免费额度）

**Key Endpoints Used:**
- `POST /` - 发送短信验证码
- `POST /` - 发送订单通知短信

## 支付宝支付 API

- **Purpose:** 在线支付处理
- **Documentation:** https://opendocs.alipay.com/
- **Base URL:** https://openapi.alipay.com/gateway.do
- **Authentication:** RSA2签名
- **Rate Limits:** 20000笔/天

**Key Endpoints Used:**
- `alipay.trade.create` - 创建交易订单
- `alipay.trade.query` - 查询交易状态
