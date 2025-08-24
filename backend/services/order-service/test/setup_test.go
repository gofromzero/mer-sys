package test

import (
	"github.com/gofromzero/mer-sys/backend/services/order-service/test/config"
	"github.com/gofromzero/mer-sys/backend/shared/auth"
)

func init() {
	// 初始化测试配置
	config.InitTestConfig()

	// 初始化JWT管理器
	auth.NewJWTManager()
}
