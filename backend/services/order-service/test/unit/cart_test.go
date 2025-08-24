package unit

import (
	"context"
	"testing"

	"github.com/gofromzero/mer-sys/backend/services/order-service/internal/service"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCartService(t *testing.T) {
	Convey("购物车服务测试", t, func() {
		ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))
		cartService := service.NewCartService()

		Convey("获取购物车", func() {
			cart, err := cartService.GetCart(ctx, 1)
			So(err, ShouldBeNil)
			So(cart, ShouldNotBeNil)
			So(cart.CustomerID, ShouldEqual, 1)
			So(cart.TenantID, ShouldEqual, 1)
		})

		Convey("添加商品到购物车", func() {
			err := cartService.AddItem(ctx, 1, 1001, 2)
			So(err, ShouldBeNil)
		})

		Convey("添加商品时数量必须大于0", func() {
			err := cartService.AddItem(ctx, 1, 1001, 0)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "商品数量必须大于0")
		})

		Convey("更新购物车商品数量", func() {
			// 先添加商品
			err := cartService.AddItem(ctx, 1, 1002, 1)
			So(err, ShouldBeNil)

			// 更新数量
			err = cartService.UpdateItemQuantity(ctx, 1, 3)
			So(err, ShouldBeNil)
		})

		Convey("更新数量时必须大于0", func() {
			err := cartService.UpdateItemQuantity(ctx, 1, -1)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "商品数量必须大于0")
		})

		Convey("清空购物车", func() {
			err := cartService.ClearCart(ctx, 1)
			So(err, ShouldBeNil)
		})
	})
}
