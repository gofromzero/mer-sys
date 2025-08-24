package unit

import (
	"testing"

	"github.com/gofromzero/mer-sys/backend/shared/types"
	. "github.com/smartystreets/goconvey/convey"
)

func TestOrderTypes(t *testing.T) {
	Convey("订单类型测试", t, func() {

		Convey("订单状态转换", func() {
			status := types.OrderStatusIntPending
			So(status.ToOrderStatus(), ShouldEqual, types.OrderStatusPending)

			status = types.OrderStatusIntPaid
			So(status.ToOrderStatus(), ShouldEqual, types.OrderStatusPaid)

			status = types.OrderStatusIntCompleted
			So(status.ToOrderStatus(), ShouldEqual, types.OrderStatusCompleted)
		})

		Convey("订单创建请求验证", func() {
			req := &types.CreateOrderRequest{
				MerchantID: 1,
				Items: []struct {
					ProductID uint64 `json:"product_id" v:"required#商品ID不能为空"`
					Quantity  int    `json:"quantity" v:"required|min:1#数量不能为空|数量必须大于0"`
				}{
					{ProductID: 1001, Quantity: 2},
				},
			}

			So(req.MerchantID, ShouldEqual, 1)
			So(len(req.Items), ShouldEqual, 1)
			So(req.Items[0].ProductID, ShouldEqual, 1001)
			So(req.Items[0].Quantity, ShouldEqual, 2)
		})

		Convey("订单确认信息", func() {
			confirmation := &types.OrderConfirmation{
				Items: []types.OrderConfirmationItem{
					{
						ProductID:          1001,
						ProductName:        "测试商品",
						Quantity:           2,
						UnitPrice:          50.0,
						UnitRightsCost:     10.0,
						SubtotalAmount:     100.0,
						SubtotalRightsCost: 20.0,
					},
				},
				TotalAmount:     100.0,
				TotalRightsCost: 20.0,
				CanCreate:       true,
			}

			So(confirmation.CanCreate, ShouldBeTrue)
			So(confirmation.TotalAmount, ShouldEqual, 100.0)
			So(confirmation.TotalRightsCost, ShouldEqual, 20.0)
			So(len(confirmation.Items), ShouldEqual, 1)
		})

		Convey("购物车项测试", func() {
			cartItem := &types.CartItem{
				ID:        1,
				CartID:    1,
				ProductID: 1001,
				Quantity:  2,
			}

			So(cartItem.ID, ShouldEqual, 1)
			So(cartItem.ProductID, ShouldEqual, 1001)
			So(cartItem.Quantity, ShouldEqual, 2)
		})

		Convey("支付方法测试", func() {
			So(string(types.PaymentMethodAlipay), ShouldEqual, "alipay")
			So(string(types.PaymentMethodWechat), ShouldEqual, "wechat")
			So(string(types.PaymentMethodBalance), ShouldEqual, "balance")
		})

		Convey("支付状态测试", func() {
			So(types.PaymentStatusUnpaid, ShouldEqual, types.PaymentStatus(1))
			So(types.PaymentStatusPaying, ShouldEqual, types.PaymentStatus(2))
			So(types.PaymentStatusPaid, ShouldEqual, types.PaymentStatus(3))
			So(types.PaymentStatusFailed, ShouldEqual, types.PaymentStatus(4))
			So(types.PaymentStatusRefunded, ShouldEqual, types.PaymentStatus(5))

			// 测试String()方法
			So(types.PaymentStatusUnpaid.String(), ShouldEqual, "unpaid")
			So(types.PaymentStatusPaid.String(), ShouldEqual, "paid")
		})
	})
}
