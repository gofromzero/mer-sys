package service

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/gogf/gf/v2/frame/g"
)

// EmailService 邮件服务接口
type EmailService interface {
	SendEmail(ctx context.Context, customerID uint64, subject, content string) error
}

// emailService 邮件服务实现
type emailService struct{}

// NewEmailService 创建邮件服务实例
func NewEmailService() EmailService {
	return &emailService{}
}

// SendEmail 发送邮件
func (s *emailService) SendEmail(ctx context.Context, customerID uint64, subject, content string) error {
	// 获取配置
	cfg := g.Cfg()
	enabled := cfg.MustGet(ctx, "email.enabled", false).Bool()

	if !enabled {
		g.Log().Info(ctx, "邮件服务未启用，仅记录日志",
			"customer_id", customerID,
			"subject", subject)
		return nil
	}

	// 获取用户邮箱地址（这里应该从用户服务获取）
	userEmail, err := s.getUserEmail(ctx, customerID)
	if err != nil {
		g.Log().Error(ctx, "获取用户邮箱失败", "error", err, "customer_id", customerID)
		return err
	}

	return s.sendSMTPEmail(ctx, userEmail, subject, content)
}

// getUserEmail 获取用户邮箱地址
func (s *emailService) getUserEmail(ctx context.Context, customerID uint64) (string, error) {
	// TODO: 从用户服务获取用户邮箱
	// 这里使用Mock数据
	return fmt.Sprintf("customer_%d@example.com", customerID), nil
}

// sendSMTPEmail 通过SMTP发送邮件
func (s *emailService) sendSMTPEmail(ctx context.Context, to, subject, content string) error {
	cfg := g.Cfg()

	smtpHost := cfg.MustGet(ctx, "email.smtp.host", "").String()
	smtpPort := cfg.MustGet(ctx, "email.smtp.port", "587").String()
	username := cfg.MustGet(ctx, "email.smtp.username", "").String()
	password := cfg.MustGet(ctx, "email.smtp.password", "").String()
	fromEmail := cfg.MustGet(ctx, "email.from", "").String()

	if smtpHost == "" || username == "" || password == "" {
		g.Log().Warning(ctx, "SMTP配置缺失，使用Mock模式")
		return s.mockSendEmail(ctx, to, subject, content)
	}

	// 构建邮件内容
	msg := s.buildEmailMessage(fromEmail, to, subject, content)

	// 发送邮件
	auth := smtp.PlainAuth("", username, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, fromEmail, []string{to}, []byte(msg))

	if err != nil {
		g.Log().Error(ctx, "发送邮件失败", "error", err, "to", to, "subject", subject)
		return err
	}

	g.Log().Info(ctx, "邮件发送成功", "to", to, "subject", subject)
	return nil
}

// buildEmailMessage 构建邮件消息
func (s *emailService) buildEmailMessage(from, to, subject, content string) string {
	return fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s\r\n",
		from, to, subject, content)
}

// mockSendEmail Mock 邮件发送（开发环境使用）
func (s *emailService) mockSendEmail(ctx context.Context, to, subject, content string) error {
	g.Log().Info(ctx, "Mock Email发送成功",
		"to", to,
		"subject", subject,
		"content_length", len(content))

	return nil
}
