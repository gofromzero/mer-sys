package service

import (
	"testing"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/types"
)

func TestNotificationTemplateManager(t *testing.T) {
	manager := NewNotificationTemplateManager()

	// 测试获取客户端短信模板
	template := manager.GetTemplate(
		NotificationMethodTypeSMS,
		NotificationCategoryCustomer,
		NotificationEventOrderCreated,
		"zh-CN",
	)

	if template == nil {
		t.Fatal("客户端订单创建短信模板不存在")
	}

	if template.Type != NotificationMethodTypeSMS {
		t.Errorf("模板类型不正确，期望: %s, 实际: %s", NotificationMethodTypeSMS, template.Type)
	}

	if template.Category != NotificationCategoryCustomer {
		t.Errorf("模板分类不正确，期望: %s, 实际: %s", NotificationCategoryCustomer, template.Category)
	}

	if template.Event != NotificationEventOrderCreated {
		t.Errorf("模板事件不正确，期望: %s, 实际: %s", NotificationEventOrderCreated, template.Event)
	}

	if !template.Enabled {
		t.Error("模板应该是启用状态")
	}

	// 测试模板渲染
	testOrder := &types.Order{
		OrderNumber:      "ORD20250828001",
		TotalAmount:      299.99,
		TotalRightsCost:  10.5,
		CreatedAt:        time.Now(),
	}

	data := manager.BuildOrderDataMap(testOrder, nil)
	subject, content, err := manager.RenderTemplate(template, data)

	if err != nil {
		t.Fatalf("模板渲染失败: %v", err)
	}

	if subject != "" {
		t.Error("短信模板不应该有主题")
	}

	expectedContent := "您的订单 ORD20250828001 已创建成功，金额 ¥299.99，请及时支付。"
	if content != expectedContent {
		t.Errorf("模板内容不正确，期望: %s, 实际: %s", expectedContent, content)
	}
}

func TestNotificationTemplateManagerEmailTemplate(t *testing.T) {
	manager := NewNotificationTemplateManager()

	// 测试获取客户端邮件模板
	template := manager.GetTemplate(
		NotificationMethodTypeEmail,
		NotificationCategoryCustomer,
		NotificationEventOrderCreated,
		"zh-CN",
	)

	if template == nil {
		t.Fatal("客户端订单创建邮件模板不存在")
	}

	// 测试模板渲染
	testOrder := &types.Order{
		OrderNumber:      "ORD20250828001",
		TotalAmount:      299.99,
		TotalRightsCost:  10.5,
		CreatedAt:        time.Now(),
	}

	data := manager.BuildOrderDataMap(testOrder, nil)
	subject, content, err := manager.RenderTemplate(template, data)

	if err != nil {
		t.Fatalf("邮件模板渲染失败: %v", err)
	}

	expectedSubject := "订单确认 - ORD20250828001"
	if subject != expectedSubject {
		t.Errorf("邮件主题不正确，期望: %s, 实际: %s", expectedSubject, subject)
	}

	if content == "" {
		t.Error("邮件内容不能为空")
	}

	// 验证内容包含关键信息
	if !contains(content, "ORD20250828001") {
		t.Error("邮件内容应该包含订单号")
	}

	if !contains(content, "¥299.99") {
		t.Error("邮件内容应该包含订单金额")
	}
}

func TestNotificationTemplateManagerMerchantTemplate(t *testing.T) {
	manager := NewNotificationTemplateManager()

	// 测试获取商户端邮件模板
	template := manager.GetTemplate(
		NotificationMethodTypeEmail,
		NotificationCategoryMerchant,
		NotificationEventOrderStatusChanged,
		"zh-CN",
	)

	if template == nil {
		t.Fatal("商户端订单状态变更邮件模板不存在")
	}

	// 测试模板渲染
	testOrder := &types.Order{
		OrderNumber:      "ORD20250828001",
		TotalAmount:      299.99,
		CustomerID:       12345,
		CreatedAt:        time.Now(),
	}

	statusHistory := &types.OrderStatusHistory{
		FromStatus:   types.OrderStatusIntPaid,
		ToStatus:     types.OrderStatusIntProcessing,
		Reason:       "商户开始处理订单",
		OperatorType: types.OrderStatusOperatorTypeMerchant,
		CreatedAt:    time.Now(),
	}

	data := manager.BuildOrderDataMap(testOrder, statusHistory)
	subject, content, err := manager.RenderTemplate(template, data)

	if err != nil {
		t.Fatalf("商户端模板渲染失败: %v", err)
	}

	expectedSubject := "商户订单状态变更 - ORD20250828001"
	if subject != expectedSubject {
		t.Errorf("商户端邮件主题不正确，期望: %s, 实际: %s", expectedSubject, subject)
	}

	// 验证内容包含关键信息
	if !contains(content, "ORD20250828001") {
		t.Error("商户端邮件内容应该包含订单号")
	}

	if !contains(content, "已支付") {
		t.Error("商户端邮件内容应该包含源状态")
	}

	if !contains(content, "处理中") {
		t.Error("商户端邮件内容应该包含目标状态")
	}
}

func TestNotificationTemplateManagerTemplateToggle(t *testing.T) {
	manager := NewNotificationTemplateManager()

	templateID := "customer_sms_order_created_zh_cn"

	// 测试禁用模板
	err := manager.DisableTemplate(templateID)
	if err != nil {
		t.Fatalf("禁用模板失败: %v", err)
	}

	template := manager.GetTemplate(
		NotificationMethodTypeSMS,
		NotificationCategoryCustomer,
		NotificationEventOrderCreated,
		"zh-CN",
	)

	if template == nil || template.Enabled {
		t.Error("模板应该被禁用")
	}

	// 测试启用模板
	err = manager.EnableTemplate(templateID)
	if err != nil {
		t.Fatalf("启用模板失败: %v", err)
	}

	template = manager.GetTemplate(
		NotificationMethodTypeSMS,
		NotificationCategoryCustomer,
		NotificationEventOrderCreated,
		"zh-CN",
	)

	if template == nil || !template.Enabled {
		t.Error("模板应该被启用")
	}
}

// contains 检查字符串是否包含子串的辅助函数
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}