package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/os/gcache"

	"mer-demo/shared/types"
)

// NotificationChannel 通知渠道类型
type NotificationChannel string

const (
	ChannelEmail NotificationChannel = "email"
	ChannelSMS   NotificationChannel = "sms"
	ChannelSystem NotificationChannel = "system"
)

// NotificationService 通知服务接口
type NotificationService interface {
	SendAlert(ctx context.Context, alert *types.RightsAlert, channels []NotificationChannel) error
	SendEmail(ctx context.Context, to, subject, body string) error
	SendSMS(ctx context.Context, phone, message string) error
	SendSystemNotification(ctx context.Context, userID uint64, title, message string) error
	GetNotificationHistory(ctx context.Context, alertID uint64) ([]NotificationRecord, error)
}

// NotificationRecord 通知记录
type NotificationRecord struct {
	ID        uint64              `json:"id"`
	AlertID   uint64              `json:"alert_id"`
	Channel   NotificationChannel `json:"channel"`
	Target    string              `json:"target"`
	Subject   string              `json:"subject"`
	Message   string              `json:"message"`
	Status    string              `json:"status"` // sent, failed, pending
	SentAt    *time.Time          `json:"sent_at,omitempty"`
	Error     string              `json:"error,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
}

// notificationService 通知服务实现
type notificationService struct {
	cache      *gcache.Cache
	httpClient *gclient.Client
}

// NewNotificationService 创建通知服务实例
func NewNotificationService() NotificationService {
	return &notificationService{
		cache:      gcache.New(),
		httpClient: gclient.New(),
	}
}

// SendAlert 发送预警通知
func (s *notificationService) SendAlert(ctx context.Context, alert *types.RightsAlert, channels []NotificationChannel) error {
	// 检查通知限流
	if err := s.checkRateLimit(ctx, alert.MerchantID, alert.AlertType); err != nil {
		return err
	}

	var sentChannels []string
	var lastError error

	for _, channel := range channels {
		var err error
		switch channel {
		case ChannelEmail:
			err = s.sendEmailAlert(ctx, alert)
		case ChannelSMS:
			err = s.sendSMSAlert(ctx, alert)
		case ChannelSystem:
			err = s.sendSystemAlert(ctx, alert)
		}

		// 记录通知结果
		record := NotificationRecord{
			AlertID:   alert.ID,
			Channel:   channel,
			Message:   alert.Message,
			CreatedAt: time.Now(),
		}

		if err != nil {
			record.Status = "failed"
			record.Error = err.Error()
			lastError = err
			g.Log().Error(ctx, "Failed to send notification", g.Map{
				"channel":  channel,
				"alert_id": alert.ID,
				"error":    err,
			})
		} else {
			record.Status = "sent"
			record.SentAt = ptrOf(time.Now())
			sentChannels = append(sentChannels, string(channel))
			g.Log().Info(ctx, "Notification sent successfully", g.Map{
				"channel":  channel,
				"alert_id": alert.ID,
			})
		}

		// 保存通知记录
		s.saveNotificationRecord(ctx, &record)
	}

	// 更新预警的通知渠道记录
	if len(sentChannels) > 0 {
		alert.NotifiedChannels = sentChannels
	}

	// 只要有一个通道发送成功就认为成功
	if len(sentChannels) > 0 {
		return nil
	}

	return lastError
}

// SendEmail 发送邮件
func (s *notificationService) SendEmail(ctx context.Context, to, subject, body string) error {
	// 获取SMTP配置
	smtpConfig := g.Cfg().MustGet(ctx, "smtp")
	if smtpConfig.IsNil() {
		return fmt.Errorf("SMTP configuration not found")
	}

	// 这里应该实现实际的邮件发送逻辑
	// 为简化实现，这里只记录日志
	g.Log().Info(ctx, "Email would be sent", g.Map{
		"to":      to,
		"subject": subject,
		"body":    body,
	})

	return nil
}

// SendSMS 发送短信
func (s *notificationService) SendSMS(ctx context.Context, phone, message string) error {
	// 获取阿里云短信配置
	smsConfig := g.Cfg().MustGet(ctx, "aliyun.sms")
	if smsConfig.IsNil() {
		return fmt.Errorf("Aliyun SMS configuration not found")
	}

	accessKeyId := g.Cfg().MustGet(ctx, "aliyun.sms.access_key_id").String()
	accessKeySecret := g.Cfg().MustGet(ctx, "aliyun.sms.access_key_secret").String()
	signName := g.Cfg().MustGet(ctx, "aliyun.sms.sign_name").String()
	templateCode := g.Cfg().MustGet(ctx, "aliyun.sms.template_code").String()

	if accessKeyId == "" || accessKeySecret == "" {
		g.Log().Info(ctx, "SMS would be sent (no config)", g.Map{
			"phone":   phone,
			"message": message,
		})
		return nil
	}

	// 构建阿里云短信API请求
	params := map[string]interface{}{
		"PhoneNumbers":  phone,
		"SignName":      signName,
		"TemplateCode":  templateCode,
		"TemplateParam": fmt.Sprintf(`{"message":"%s"}`, message),
	}

	// 这里应该实现阿里云短信API调用
	// 为简化实现，这里只记录日志
	g.Log().Info(ctx, "SMS would be sent via Aliyun", g.Map{
		"phone":    phone,
		"message":  message,
		"params":   params,
	})

	return nil
}

// SendSystemNotification 发送系统内通知
func (s *notificationService) SendSystemNotification(ctx context.Context, userID uint64, title, message string) error {
	notification := g.Map{
		"user_id":    userID,
		"title":      title,
		"message":    message,
		"type":       "alert",
		"created_at": time.Now(),
		"read":       false,
	}

	// 这里应该保存到数据库的notifications表
	// 为简化实现，这里使用缓存模拟
	key := fmt.Sprintf("system_notification:%d:%d", userID, time.Now().Unix())
	s.cache.Set(ctx, key, notification, time.Hour*24*7) // 保存7天

	g.Log().Info(ctx, "System notification saved", g.Map{
		"user_id": userID,
		"title":   title,
		"key":     key,
	})

	return nil
}

// GetNotificationHistory 获取通知历史
func (s *notificationService) GetNotificationHistory(ctx context.Context, alertID uint64) ([]NotificationRecord, error) {
	// 这里应该从数据库查询通知记录
	// 为简化实现，返回空列表
	return []NotificationRecord{}, nil
}

// sendEmailAlert 发送邮件预警
func (s *notificationService) sendEmailAlert(ctx context.Context, alert *types.RightsAlert) error {
	// 获取商户管理员邮箱
	adminEmails := s.getMerchantAdminEmails(ctx, alert.MerchantID)
	if len(adminEmails) == 0 {
		return fmt.Errorf("no admin emails found for merchant %d", alert.MerchantID)
	}

	subject := fmt.Sprintf("权益预警通知 - %s", alert.AlertType.String())
	body := fmt.Sprintf(`
亲爱的管理员：

您的权益账户触发了预警条件：

预警类型：%s
当前值：%.2f
阈值：%.2f
触发时间：%s

详细信息：
%s

请及时登录系统查看详情并采取相应措施。

此邮件由系统自动发送，请勿回复。
`, alert.AlertType.String(), alert.CurrentValue, alert.ThresholdValue, alert.TriggeredAt.Format("2006-01-02 15:04:05"), alert.Message)

	var lastError error
	for _, email := range adminEmails {
		if err := s.SendEmail(ctx, email, subject, body); err != nil {
			lastError = err
		}
	}

	return lastError
}

// sendSMSAlert 发送短信预警
func (s *notificationService) sendSMSAlert(ctx context.Context, alert *types.RightsAlert) error {
	// 获取商户管理员手机号
	adminPhones := s.getMerchantAdminPhones(ctx, alert.MerchantID)
	if len(adminPhones) == 0 {
		return fmt.Errorf("no admin phones found for merchant %d", alert.MerchantID)
	}

	message := fmt.Sprintf("权益预警：%s，当前值%.2f，阈值%.2f，请及时处理。", 
		alert.AlertType.String(), alert.CurrentValue, alert.ThresholdValue)

	var lastError error
	for _, phone := range adminPhones {
		if err := s.SendSMS(ctx, phone, message); err != nil {
			lastError = err
		}
	}

	return lastError
}

// sendSystemAlert 发送系统内预警
func (s *notificationService) sendSystemAlert(ctx context.Context, alert *types.RightsAlert) error {
	// 获取商户管理员用户ID
	adminUserIDs := s.getMerchantAdminUserIDs(ctx, alert.MerchantID)
	if len(adminUserIDs) == 0 {
		return fmt.Errorf("no admin users found for merchant %d", alert.MerchantID)
	}

	title := fmt.Sprintf("权益预警 - %s", alert.AlertType.String())

	var lastError error
	for _, userID := range adminUserIDs {
		if err := s.SendSystemNotification(ctx, userID, title, alert.Message); err != nil {
			lastError = err
		}
	}

	return lastError
}

// checkRateLimit 检查通知限流
func (s *notificationService) checkRateLimit(ctx context.Context, merchantID uint64, alertType types.AlertType) error {
	key := fmt.Sprintf("rate_limit:merchant:%d:alert:%s", merchantID, alertType.String())
	
	// 每种预警类型每小时最多发送3次
	count, err := s.cache.Get(ctx, key)
	if err != nil {
		s.cache.Set(ctx, key, 1, time.Hour)
		return nil
	}

	if count.Int() >= 3 {
		return fmt.Errorf("rate limit exceeded for merchant %d alert type %s", merchantID, alertType.String())
	}

	s.cache.Set(ctx, key, count.Int()+1, time.Hour)
	return nil
}

// getMerchantAdminEmails 获取商户管理员邮箱
func (s *notificationService) getMerchantAdminEmails(ctx context.Context, merchantID uint64) []string {
	// 这里应该查询数据库获取商户管理员邮箱
	// 为简化实现，返回模拟数据
	return []string{fmt.Sprintf("admin@merchant%d.com", merchantID)}
}

// getMerchantAdminPhones 获取商户管理员手机号
func (s *notificationService) getMerchantAdminPhones(ctx context.Context, merchantID uint64) []string {
	// 这里应该查询数据库获取商户管理员手机号
	// 为简化实现，返回模拟数据
	return []string{fmt.Sprintf("1388888%04d", merchantID)}
}

// getMerchantAdminUserIDs 获取商户管理员用户ID
func (s *notificationService) getMerchantAdminUserIDs(ctx context.Context, merchantID uint64) []uint64 {
	// 这里应该查询数据库获取商户管理员用户ID
	// 为简化实现，返回模拟数据
	return []uint64{merchantID * 100} // 假设用户ID是商户ID*100
}

// saveNotificationRecord 保存通知记录
func (s *notificationService) saveNotificationRecord(ctx context.Context, record *NotificationRecord) {
	// 这里应该保存到数据库
	// 为简化实现，使用缓存
	key := fmt.Sprintf("notification_record:%d:%d", record.AlertID, time.Now().UnixNano())
	s.cache.Set(ctx, key, record, time.Hour*24*30) // 保存30天
}

// ptrOf helper function for creating pointers
func ptrOf[T any](v T) *T {
	return &v
}