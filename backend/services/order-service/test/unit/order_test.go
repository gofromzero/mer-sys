package unit

import (
	"context"
	"testing"

	"github.com/gofromzero/mer-sys/backend/services/order-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	. "github.com/smartystreets/goconvey/convey"
)

func TestOrderService(t *testing.T) {
	Convey("订单服务测试", t, func() {
		ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))
		orderService := service.NewOrderService()

		Convey("获取订单确认信息", func() {
			req := &types.CreateOrderRequest{
				MerchantID: 1,
				Items: []struct {
					ProductID uint64 `json:"product_id" v:"required#商品ID不能为空"`
					Quantity  int    `json:"quantity" v:"required|min:1#数量不能为空|数量必须大于0"`
				}{
					{ProductID: 1001, Quantity: 2},
					{ProductID: 1002, Quantity: 1},
				},
			}

			confirmation, err := orderService.GetOrderConfirmation(ctx, 1, req)
			So(err, ShouldBeNil)
			So(confirmation, ShouldNotBeNil)
			So(confirmation.Items, ShouldHaveLength, 2)
			So(confirmation.TotalAmount, ShouldBeGreaterThan, 0)
			So(confirmation.TotalRightsCost, ShouldBeGreaterThan, 0)
		})

		Convey("创建订单", func() {
			req := &types.CreateOrderRequest{
				MerchantID: 1,
				Items: []struct {
					ProductID uint64 `json:"product_id" v:"required#商品ID不能为空"`
					Quantity  int    `json:"quantity" v:"required|min:1#数量不能为空|数量必须大于0"`
				}{
					{ProductID: 1001, Quantity: 1},
				},
			}

			order, err := orderService.CreateOrder(ctx, 1, req)
			So(err, ShouldBeNil)
			So(order, ShouldNotBeNil)
			So(order.CustomerID, ShouldEqual, 1)
			So(order.MerchantID, ShouldEqual, 1)
			So(order.Status, ShouldEqual, types.OrderStatusPending)
			So(order.OrderNumber, ShouldNotBeEmpty)
		})

		Convey("取消订单", func() {
			// 先创建订单
			req := &types.CreateOrderRequest{
				MerchantID: 1,
				Items: []struct {
					ProductID uint64 `json:"product_id" v:"required#商品ID不能为空"`
					Quantity  int    `json:"quantity" v:"required|min:1#数量不能为空|数量必须大于0"`
				}{
					{ProductID: 1001, Quantity: 1},
				},
			}

			order, err := orderService.CreateOrder(ctx, 1, req)
			So(err, ShouldBeNil)

			// 取消订单
			err = orderService.CancelOrder(ctx, order.ID)
			So(err, ShouldBeNil)
		})
	})
}
