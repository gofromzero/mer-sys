package controller

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthController(t *testing.T) {
	Convey("认证控制器测试", t, func() {
		// 创建控制器实例
		authController := NewAuthController()
		
		Convey("控制器创建", func() {
			So(authController, ShouldNotBeNil)
			So(authController.authService, ShouldNotBeNil)
			So(authController.jwtManager, ShouldNotBeNil)
		})
	})
}

func TestLoginRequest(t *testing.T) {
	Convey("登录请求结构测试", t, func() {
		req := LoginRequest{
			Username: "testuser",
			Password: "testpass123",
			TenantID: 1,
		}
		
		Convey("请求结构字段", func() {
			So(req.Username, ShouldEqual, "testuser")
			So(req.Password, ShouldEqual, "testpass123")
			So(req.TenantID, ShouldEqual, 1)
		})
	})
}

func TestLoginResponse(t *testing.T) {
	Convey("登录响应结构测试", t, func() {
		resp := LoginResponse{
			AccessToken:  "access-token-123",
			RefreshToken: "refresh-token-456",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}
		
		Convey("响应结构字段", func() {
			So(resp.AccessToken, ShouldEqual, "access-token-123")
			So(resp.RefreshToken, ShouldEqual, "refresh-token-456")
			So(resp.ExpiresIn, ShouldEqual, 3600)
			So(resp.TokenType, ShouldEqual, "Bearer")
		})
	})
}

func TestLogoutRequest(t *testing.T) {
	Convey("登出请求结构测试", t, func() {
		req := LogoutRequest{
			Token:        "access-token-123",
			RefreshToken: "refresh-token-456",
		}
		
		Convey("请求结构字段", func() {
			So(req.Token, ShouldEqual, "access-token-123")
			So(req.RefreshToken, ShouldEqual, "refresh-token-456")
		})
	})
}