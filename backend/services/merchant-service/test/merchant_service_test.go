package test

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/gofromzero/mer-sys/backend/services/merchant-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

func TestMerchantService(t *testing.T) {
	Convey("商户服务测试", t, func() {
		ctx := context.Background()
		
		// 设置测试上下文，模拟认证用户
		ctx = context.WithValue(ctx, "user_id", uint64(1))
		ctx = context.WithValue(ctx, "tenant_id", uint64(1))
		ctx = context.WithValue(ctx, "username", "test_user")
		
		merchantService := service.NewMerchantService()

		Convey("注册商户测试", func() {
			Convey("正常注册应该成功", func() {
				req := &types.MerchantRegistrationRequest{
					Name: "测试商户",
					Code: "TEST001",
					BusinessInfo: &types.BusinessInfo{
						Type:         "retail",
						Category:     "零售",
						License:      "123456789",
						LegalName:    "张三",
						ContactName:  "张三",
						ContactPhone: "13800138000",
						ContactEmail: "test@example.com",
						Address:      "测试地址",
						Scope:        "零售业务",
						Description:  "测试商户",
					},
				}

				merchant, err := merchantService.RegisterMerchant(ctx, req)
				
				So(err, ShouldBeNil)
				So(merchant, ShouldNotBeNil)
				So(merchant.Name, ShouldEqual, req.Name)
				So(merchant.Code, ShouldEqual, req.Code)
				So(merchant.Status, ShouldEqual, types.MerchantStatusPending)
				So(merchant.TenantID, ShouldEqual, uint64(1))
			})

			Convey("重复代码应该失败", func() {
				// 第一次注册
				req := &types.MerchantRegistrationRequest{
					Name: "测试商户1",
					Code: "DUPLICATE001",
					BusinessInfo: &types.BusinessInfo{
						Type:         "retail",
						Category:     "零售",
						License:      "123456789",
						LegalName:    "张三",
						ContactName:  "张三",
						ContactPhone: "13800138000",
						ContactEmail: "test@example.com",
						Address:      "测试地址",
						Scope:        "零售业务",
						Description:  "测试商户",
					},
				}

				_, err := merchantService.RegisterMerchant(ctx, req)
				So(err, ShouldBeNil)

				// 第二次注册相同代码应该失败
				req2 := &types.MerchantRegistrationRequest{
					Name: "测试商户2",
					Code: "DUPLICATE001", // 相同代码
					BusinessInfo: req.BusinessInfo,
				}

				_, err = merchantService.RegisterMerchant(ctx, req2)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "已存在")
			})
		})

		Convey("获取商户列表测试", func() {
			Convey("正常查询应该成功", func() {
				query := &types.MerchantListQuery{
					Page:     1,
					PageSize: 20,
				}

				merchants, total, err := merchantService.GetMerchantList(ctx, query)
				
				So(err, ShouldBeNil)
				So(merchants, ShouldNotBeNil)
				So(total, ShouldBeGreaterThanOrEqualTo, 0)
			})

			Convey("按状态筛选应该工作", func() {
				query := &types.MerchantListQuery{
					Page:     1,
					PageSize: 20,
					Status:   types.MerchantStatusPending,
				}

				merchants, _, err := merchantService.GetMerchantList(ctx, query)
				
				So(err, ShouldBeNil)
				for _, merchant := range merchants {
					So(merchant.Status, ShouldEqual, types.MerchantStatusPending)
				}
			})
		})

		Convey("商户审批测试", func() {
			// 先创建一个待审核的商户
			req := &types.MerchantRegistrationRequest{
				Name: "待审批商户",
				Code: "APPROVAL001",
				BusinessInfo: &types.BusinessInfo{
					Type:         "retail",
					Category:     "零售",
					License:      "123456789",
					LegalName:    "李四",
					ContactName:  "李四",
					ContactPhone: "13900139000",
					ContactEmail: "approval@example.com",
					Address:      "审批测试地址",
					Scope:        "零售业务",
					Description:  "待审批测试商户",
				},
			}

			merchant, err := merchantService.RegisterMerchant(ctx, req)
			So(err, ShouldBeNil)
			So(merchant.Status, ShouldEqual, types.MerchantStatusPending)

			Convey("审批通过应该成功", func() {
				err := merchantService.ApproveMerchant(ctx, merchant.ID, "审批通过")
				So(err, ShouldBeNil)

				// 验证状态已更新
				updatedMerchant, err := merchantService.GetMerchantByID(ctx, merchant.ID)
				So(err, ShouldBeNil)
				So(updatedMerchant.Status, ShouldEqual, types.MerchantStatusActive)
			})

			Convey("拒绝审批应该成功", func() {
				err := merchantService.RejectMerchant(ctx, merchant.ID, "不符合要求")
				So(err, ShouldBeNil)

				// 验证状态已更新
				updatedMerchant, err := merchantService.GetMerchantByID(ctx, merchant.ID)
				So(err, ShouldBeNil)
				So(updatedMerchant.Status, ShouldEqual, types.MerchantStatusDeactivated)
			})
		})

		Convey("商户状态管理测试", func() {
			// 创建并审批一个商户
			req := &types.MerchantRegistrationRequest{
				Name: "状态测试商户",
				Code: "STATUS001",
				BusinessInfo: &types.BusinessInfo{
					Type:         "retail",
					Category:     "零售",
					License:      "123456789",
					LegalName:    "王五",
					ContactName:  "王五",
					ContactPhone: "13700137000",
					ContactEmail: "status@example.com",
					Address:      "状态测试地址",
					Scope:        "零售业务",
					Description:  "状态测试商户",
				},
			}

			merchant, err := merchantService.RegisterMerchant(ctx, req)
			So(err, ShouldBeNil)

			err = merchantService.ApproveMerchant(ctx, merchant.ID, "测试审批")
			So(err, ShouldBeNil)

			Convey("状态变更应该成功", func() {
				err := merchantService.UpdateMerchantStatus(ctx, merchant.ID, types.MerchantStatusSuspended, "测试暂停")
				So(err, ShouldBeNil)

				// 验证状态已更新
				updatedMerchant, err := merchantService.GetMerchantByID(ctx, merchant.ID)
				So(err, ShouldBeNil)
				So(updatedMerchant.Status, ShouldEqual, types.MerchantStatusSuspended)
			})
		})

		Convey("商户信息更新测试", func() {
			// 先创建一个商户
			req := &types.MerchantRegistrationRequest{
				Name: "更新测试商户",
				Code: "UPDATE001",
				BusinessInfo: &types.BusinessInfo{
					Type:         "retail",
					Category:     "零售",
					License:      "123456789",
					LegalName:    "赵六",
					ContactName:  "赵六",
					ContactPhone: "13600136000",
					ContactEmail: "update@example.com",
					Address:      "更新测试地址",
					Scope:        "零售业务",
					Description:  "更新测试商户",
				},
			}

			merchant, err := merchantService.RegisterMerchant(ctx, req)
			So(err, ShouldBeNil)

			Convey("更新商户信息应该成功", func() {
				newName := "更新后的商户名称"
				updateReq := &types.MerchantUpdateRequest{
					Name: &newName,
					BusinessInfo: &types.BusinessInfo{
						Type:         "wholesale", // 改为批发
						Category:     "批发",
						License:      req.BusinessInfo.License,
						LegalName:    req.BusinessInfo.LegalName,
						ContactName:  req.BusinessInfo.ContactName,
						ContactPhone: req.BusinessInfo.ContactPhone,
						ContactEmail: req.BusinessInfo.ContactEmail,
						Address:      req.BusinessInfo.Address,
						Scope:        "批发业务", // 更新经营范围
						Description:  req.BusinessInfo.Description,
					},
				}

				updatedMerchant, err := merchantService.UpdateMerchant(ctx, merchant.ID, updateReq)
				So(err, ShouldBeNil)
				So(updatedMerchant.Name, ShouldEqual, newName)
				So(updatedMerchant.BusinessInfo.Type, ShouldEqual, "wholesale")
				So(updatedMerchant.BusinessInfo.Scope, ShouldEqual, "批发业务")
			})
		})

		Convey("错误处理测试", func() {
			Convey("获取不存在的商户应该返回错误", func() {
				_, err := merchantService.GetMerchantByID(ctx, 99999)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "不存在")
			})

			Convey("审批不存在的商户应该返回错误", func() {
				err := merchantService.ApproveMerchant(ctx, 99999, "测试")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "不存在")
			})

			Convey("更新不存在的商户应该返回错误", func() {
				name := "测试"
				updateReq := &types.MerchantUpdateRequest{Name: &name}
				_, err := merchantService.UpdateMerchant(ctx, 99999, updateReq)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "不存在")
			})
		})
	})
}