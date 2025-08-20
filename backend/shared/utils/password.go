package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

// PasswordManager 密码管理器
type PasswordManager struct {
	cost int // bcrypt cost factor (默认12)
}

// NewPasswordManager 创建密码管理器
func NewPasswordManager() *PasswordManager {
	return &PasswordManager{
		cost: 12, // 推荐的安全级别
	}
}

// HashPassword 使用bcrypt加密密码
func (pm *PasswordManager) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), pm.cost)
	if err != nil {
		return "", fmt.Errorf("密码加密失败: %v", err)
	}
	return string(bytes), nil
}

// CheckPassword 验证密码
func (pm *PasswordManager) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomPassword 生成随机密码
func (pm *PasswordManager) GenerateRandomPassword(length int) (string, error) {
	if length < 6 {
		length = 8 // 最小长度8位
	}
	if length > 32 {
		length = 32 // 最大长度32位
	}

	// 包含大小写字母、数字和特殊字符
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, length)
	
	for i := range password {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("生成随机密码失败: %v", err)
		}
		password[i] = charset[num.Int64()]
	}
	
	return string(password), nil
}

// IsValidPassword 验证密码强度
func (pm *PasswordManager) IsValidPassword(password string) error {
	if len(password) < 6 {
		return fmt.Errorf("密码长度至少6位")
	}
	if len(password) > 128 {
		return fmt.Errorf("密码长度不能超过128位")
	}
	
	// 可以添加更多密码复杂度检查
	return nil
}

// 全局密码管理器实例
var defaultPasswordManager = NewPasswordManager()

// HashPassword 全局密码加密函数
func HashPassword(password string) (string, error) {
	return defaultPasswordManager.HashPassword(password)
}

// CheckPassword 全局密码验证函数
func CheckPassword(password, hash string) bool {
	return defaultPasswordManager.CheckPassword(password, hash)
}

// GenerateRandomPassword 全局随机密码生成函数
func GenerateRandomPassword(length int) (string, error) {
	return defaultPasswordManager.GenerateRandomPassword(length)
}

// IsValidPassword 全局密码强度验证函数
func IsValidPassword(password string) error {
	return defaultPasswordManager.IsValidPassword(password)
}