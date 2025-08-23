package test

import (
	"log"
	"os"
	"testing"

	"github.com/gofromzero/mer-sys/backend/services/product-service/test/config"
)

// TestMain 测试入口点，用于设置和清理测试环境
func TestMain(m *testing.M) {
	// 设置测试环境
	if err := setupTestEnvironment(); err != nil {
		log.Fatalf("测试环境设置失败: %v", err)
	}

	// 运行测试
	code := m.Run()

	// 清理测试环境
	if err := cleanupTestEnvironment(); err != nil {
		log.Printf("测试环境清理失败: %v", err)
	}

	os.Exit(code)
}

// setupTestEnvironment 设置测试环境
func setupTestEnvironment() error {
	// 1. 初始化测试数据库连接
	if err := config.InitTestDatabase(); err != nil {
		// 如果数据库连接失败，记录警告但允许测试继续（使用Mock）
		log.Printf("警告: 测试数据库初始化失败，将使用Mock测试: %v", err)
		return nil
	}

	// 2. 创建测试需要的数据表
	if err := config.SetupTestTables(); err != nil {
		log.Printf("警告: 测试表创建失败: %v", err)
		return nil
	}

	// 3. 插入测试数据
	if err := config.InsertTestData(); err != nil {
		log.Printf("警告: 测试数据插入失败: %v", err)
		return nil
	}

	log.Println("测试环境设置完成")
	return nil
}

// cleanupTestEnvironment 清理测试环境
func cleanupTestEnvironment() error {
	// 清理测试数据
	if err := config.CleanupTestTables(); err != nil {
		log.Printf("警告: 测试数据清理失败: %v", err)
	}

	log.Println("测试环境清理完成")
	return nil
}