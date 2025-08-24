package config

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

// InitTestConfig 初始化测试配置
func InitTestConfig() {
	ctx := gctx.GetInitCtx()

	// 设置测试数据库配置
	g.Cfg().SetData(g.Map{
		"database": g.Map{
			"link":  "mysql:mer_user:mer_password@tcp(127.0.0.1:3306)/mer_system_test",
			"debug": true,
		},
		"redis": g.Map{
			"address": "127.0.0.1:6379",
			"db":      2, // 使用不同的Redis数据库进行测试
		},
		"jwt": g.Map{
			"sign_key": "test_jwt_secret_key",
			"expire":   3600,
		},
	}, ctx)
}
