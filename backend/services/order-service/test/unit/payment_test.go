package unit

import (
	"context"
	"testing"

	"github.com/gofromzero/mer-sys/backend/services/order-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPaymentService(t *testing.T) {
	Convey("支付服务测试", t, func() {
		ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))
		paymentService := service.NewPaymentService()
		orderService := service.NewOrderService()

		Convey("发起支付", func() {
			// 先创建一个订单
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

			Convey("支付宝支付", func() {
				paymentInfo, err := paymentService.InitiatePayment(ctx, order.ID, types.PaymentMethodAlipay, "http://localhost:3000/payment/success")
				So(err, ShouldBeNil)
				So(paymentInfo, ShouldNotBeNil)
				So(paymentInfo.PaymentMethod, ShouldEqual, types.PaymentMethodAlipay)
				So(paymentInfo.PaymentURL, ShouldNotBeEmpty)
			})

			Convey("不支持的支付方式", func() {
				paymentInfo, err := paymentService.InitiatePayment(ctx, order.ID, types.PaymentMethodWechat, "")
				So(err, ShouldNotBeNil)
				So(paymentInfo, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "暂不支持微信支付")
			})
		})

		Convey("查询支付状态", func() {
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

			status, err := paymentService.GetPaymentStatus(ctx, order.ID)
			So(err, ShouldBeNil)
			So(status, ShouldEqual, "pending")
		})

		Convey("支付回调处理", func() {
			// 先创建订单并发起支付
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

			_, err = paymentService.InitiatePayment(ctx, order.ID, types.PaymentMethodAlipay, "")
			So(err, ShouldBeNil)

			Convey("支付成功回调", func() {
				callbackData := map[string]interface{}{
					"out_trade_no": order.OrderNumber,
					"trade_status": "TRADE_SUCCESS",
					"total_amount": "100.00",
				}

				err := paymentService.HandleAlipayCallback(ctx, callbackData)
				So(err, ShouldBeNil)

				// 验证订单状态已更新
				updatedOrder, err := orderService.GetOrder(ctx, order.ID)
				So(err, ShouldBeNil)
				So(updatedOrder.Status, ShouldEqual, types.OrderStatusPaid)
			})

			Convey("无效的回调数据", func() {
				callbackData := map[string]interface{}{
					"invalid_field": "invalid_value",
				}

				err := paymentService.HandleAlipayCallback(ctx, callbackData)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("重新支付", func() {
			// 创建订单
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

			// 重新支付
			paymentInfo, err := paymentService.RetryPayment(ctx, order.ID, types.PaymentMethodAlipay, "http://localhost:3000/payment/retry")
			So(err, ShouldBeNil)
			So(paymentInfo, ShouldNotBeNil)
			So(paymentInfo.PaymentMethod, ShouldEqual, types.PaymentMethodAlipay)
		})

		Convey("对已支付订单重新支付应失败", func() {
			// 先创建并支付订单
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

			// 模拟支付成功
			callbackData := map[string]interface{}{
				"out_trade_no": order.OrderNumber,
				"trade_status": "TRADE_SUCCESS",
				"total_amount": "100.00",
			}
			paymentService.HandleAlipayCallback(ctx, callbackData)

			// 尝试重新支付已支付订单
			paymentInfo, err := paymentService.RetryPayment(ctx, order.ID, types.PaymentMethodAlipay, "")
			So(err, ShouldNotBeNil)
			So(paymentInfo, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "订单状态不正确")
		})
	})
}
