package utils

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPasswordManager(t *testing.T) {
	Convey("密码管理器测试", t, func() {
		pm := NewPasswordManager()

		Convey("密码加密和验证", func() {
			password := "test123456"
			
			hash, err := pm.HashPassword(password)
			So(err, ShouldBeNil)
			So(hash, ShouldNotBeEmpty)
			So(hash, ShouldNotEqual, password) // 哈希后应该不等于原密码
			
			// 验证正确密码
			isValid := pm.CheckPassword(password, hash)
			So(isValid, ShouldBeTrue)
			
			// 验证错误密码
			isValid = pm.CheckPassword("wrongpassword", hash)
			So(isValid, ShouldBeFalse)
		})

		Convey("随机密码生成", func() {
			// 测试不同长度
			lengths := []int{6, 8, 12, 16, 32}
			
			for _, length := range lengths {
				password, err := pm.GenerateRandomPassword(length)
				So(err, ShouldBeNil)
				So(len(password), ShouldEqual, length)
				So(password, ShouldNotBeEmpty)
			}
			
			// 测试两次生成的密码不同
			pass1, _ := pm.GenerateRandomPassword(10)
			pass2, _ := pm.GenerateRandomPassword(10)
			So(pass1, ShouldNotEqual, pass2)
		})

		Convey("密码强度验证", func() {
			// 有效密码
			validPasswords := []string{
				"123456",
				"password",
				"Test123!",
				"VeryLongPasswordWith123AndSymbols!@#",
			}
			
			for _, password := range validPasswords {
				err := pm.IsValidPassword(password)
				So(err, ShouldBeNil)
			}
			
			// 无效密码
			invalidPasswords := []string{
				"",      // 空密码
				"12345", // 太短
				strings.Repeat("a", 129), // 太长
			}
			
			for _, password := range invalidPasswords {
				err := pm.IsValidPassword(password)
				So(err, ShouldNotBeNil)
			}
		})

		Convey("边界值测试", func() {
			// 最小长度
			password, err := pm.GenerateRandomPassword(5) // 小于6应该调整为8
			So(err, ShouldBeNil)
			So(len(password), ShouldEqual, 8)
			
			// 最大长度
			password, err = pm.GenerateRandomPassword(50) // 大于32应该调整为32
			So(err, ShouldBeNil)
			So(len(password), ShouldEqual, 32)
		})
	})
}

func TestGlobalPasswordFunctions(t *testing.T) {
	Convey("全局密码函数测试", t, func() {
		password := "globaltest123"
		
		// 测试全局加密函数
		hash, err := HashPassword(password)
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeEmpty)
		
		// 测试全局验证函数
		isValid := CheckPassword(password, hash)
		So(isValid, ShouldBeTrue)
		
		// 测试全局随机密码生成
		randomPassword, err := GenerateRandomPassword(12)
		So(err, ShouldBeNil)
		So(len(randomPassword), ShouldEqual, 12)
		
		// 测试全局密码强度验证
		err = IsValidPassword(password)
		So(err, ShouldBeNil)
		
		err = IsValidPassword("12345")
		So(err, ShouldNotBeNil)
	})
}

func TestPasswordSecurity(t *testing.T) {
	Convey("密码安全性测试", t, func() {
		pm := NewPasswordManager()
		
		Convey("相同密码生成不同哈希", func() {
			password := "samepassword"
			
			hash1, _ := pm.HashPassword(password)
			hash2, _ := pm.HashPassword(password)
			
			// bcrypt每次生成的哈希都应该不同（因为有salt）
			So(hash1, ShouldNotEqual, hash2)
			
			// 但都应该能验证原密码
			So(pm.CheckPassword(password, hash1), ShouldBeTrue)
			So(pm.CheckPassword(password, hash2), ShouldBeTrue)
		})
		
		Convey("哈希不可逆性", func() {
			password := "irreversible123"
			hash, _ := pm.HashPassword(password)
			
			// 哈希值不应该包含原密码
			So(strings.Contains(hash, password), ShouldBeFalse)
			
			// 哈希长度应该固定（bcrypt固定60字符）
			So(len(hash), ShouldEqual, 60)
		})
	})
}

func BenchmarkPasswordOperations(b *testing.B) {
	pm := NewPasswordManager()
	password := "benchmarkpassword123"
	hash, _ := pm.HashPassword(password)
	
	b.Run("HashPassword", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			pm.HashPassword(password)
		}
	})
	
	b.Run("CheckPassword", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			pm.CheckPassword(password, hash)
		}
	})
	
	b.Run("GenerateRandomPassword", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			pm.GenerateRandomPassword(12)
		}
	})
}