package test

import (
	"testing"
	
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/smartystreets/goconvey/convey"
)

func TestProductTypes(t *testing.T) {
	convey.Convey("商品类型定义测试", t, func() {
		convey.Convey("商品状态枚举应该包含所有需要的状态", func() {
			convey.So(string(types.ProductStatusDraft), convey.ShouldEqual, "draft")
			convey.So(string(types.ProductStatusActive), convey.ShouldEqual, "active")
			convey.So(string(types.ProductStatusInactive), convey.ShouldEqual, "inactive")
			convey.So(string(types.ProductStatusDeleted), convey.ShouldEqual, "deleted")
		})
		
		convey.Convey("分类状态枚举应该正确", func() {
			convey.So(types.CategoryStatusActive, convey.ShouldEqual, 1)
			convey.So(types.CategoryStatusInactive, convey.ShouldEqual, 0)
		})
		
		convey.Convey("变更操作类型应该完整", func() {
			convey.So(string(types.ChangeOperationCreate), convey.ShouldEqual, "create")
			convey.So(string(types.ChangeOperationUpdate), convey.ShouldEqual, "update")
			convey.So(string(types.ChangeOperationDelete), convey.ShouldEqual, "delete")
			convey.So(string(types.ChangeOperationStatusChange), convey.ShouldEqual, "status_change")
		})
	})
}

func TestProductValidation(t *testing.T) {
	convey.Convey("商品数据验证测试", t, func() {
		convey.Convey("创建商品请求应该有效", func() {
			req := &types.CreateProductRequest{
				Name:        "测试商品",
				Description: "测试商品描述",
				Price: types.Money{
					Amount:   9999, // 99.99元，以分为单位
					Currency: "CNY",
				},
				RightsCost: 1000, // 10.00元权益
				Inventory: types.InventoryInfo{
					StockQuantity:    100,
					ReservedQuantity: 0,
					TrackInventory:   true,
				},
			}
			
			convey.So(req.Name, convey.ShouldEqual, "测试商品")
			convey.So(req.Price.Amount, convey.ShouldEqual, 9999)
			convey.So(req.Price.Currency, convey.ShouldEqual, "CNY")
			convey.So(req.Inventory.StockQuantity, convey.ShouldEqual, 100)
		})
	})
}