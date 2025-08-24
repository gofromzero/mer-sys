package service

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
)

// SMSService 短信服务接口
type SMSService interface {
	SendSMS(ctx context.Context, customerID uint64, templateCode, content string) error
}

// smsService 短信服务实现
type smsService struct{}

// NewSMSService 创建短信服务实例
func NewSMSService() SMSService {
	return &smsService{}
}

// SendSMS 发送短信
func (s *smsService) SendSMS(ctx context.Context, customerID uint64, templateCode, content string) error {
	// 获取配置
	cfg := g.Cfg()
	enabled := cfg.MustGet(ctx, "sms.enabled", false).Bool()

	if !enabled {
		g.Log().Info(ctx, "短信服务未启用，仅记录日志",
			"customer_id", customerID,
			"template_code", templateCode,
			"content", content)
		return nil
	}

	// Mock 阿里云短信服务集成
	// 在实际环境中，这里应该集成阿里云SMS SDK
	return s.sendAliyunSMS(ctx, customerID, templateCode, content)
}

// sendAliyunSMS Mock 阿里云短信发送
func (s *smsService) sendAliyunSMS(ctx context.Context, customerID uint64, templateCode, content string) error {
	// 获取阿里云短信配置
	cfg := g.Cfg()
	accessKeyId := cfg.MustGet(ctx, "sms.aliyun.access_key_id", "").String()
	accessKeySecret := cfg.MustGet(ctx, "sms.aliyun.access_key_secret", "").String()
	signName := cfg.MustGet(ctx, "sms.aliyun.sign_name", "商户系统").String()

	if accessKeyId == "" || accessKeySecret == "" {
		g.Log().Warning(ctx, "阿里云短信配置缺失，使用Mock模式")
		return s.mockSendSMS(ctx, customerID, templateCode, content)
	}

	// TODO: 实际集成阿里云SMS SDK
	// 这里应该使用官方的阿里云Go SDK发送短信
	g.Log().Info(ctx, "Mock: 通过阿里云发送短信",
		"customer_id", customerID,
		"sign_name", signName,
		"template_code", templateCode,
		"content", content)

	return nil
}

// mockSendSMS Mock 短信发送（开发环境使用）
func (s *smsService) mockSendSMS(ctx context.Context, customerID uint64, templateCode, content string) error {
	// 在开发环境中，我们可以将短信内容写入日志文件
	g.Log().Info(ctx, "Mock SMS发送成功",
		"customer_id", customerID,
		"template_code", templateCode,
		"content", content)

	return nil
}

// 短信模板常量
const (
	SMS_TEMPLATE_ORDER_CREATED   = "ORDER_CREATED"
	SMS_TEMPLATE_PAYMENT_SUCCESS = "PAYMENT_SUCCESS"
	SMS_TEMPLATE_PAYMENT_FAILURE = "PAYMENT_FAILURE"
	SMS_TEMPLATE_ORDER_COMPLETED = "ORDER_COMPLETED"
)

// getSMSTemplate 获取短信模板
func getSMSTemplate(templateCode string) (string, error) {
	templates := map[string]string{
		SMS_TEMPLATE_ORDER_CREATED:   "您的订单${orderNumber}已创建成功，金额¥${amount}，请及时支付。",
		SMS_TEMPLATE_PAYMENT_SUCCESS: "您的订单${orderNumber}支付成功，金额¥${amount}，我们将尽快处理。",
		SMS_TEMPLATE_PAYMENT_FAILURE: "您的订单${orderNumber}支付失败，请重新支付或联系客服。",
		SMS_TEMPLATE_ORDER_COMPLETED: "您的订单${orderNumber}已完成，感谢使用！",
	}

	template, exists := templates[templateCode]
	if !exists {
		return "", fmt.Errorf("短信模板不存在: %s", templateCode)
	}

	return template, nil
}
