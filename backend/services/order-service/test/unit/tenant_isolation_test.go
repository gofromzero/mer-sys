package unit

import (
	"context"
	"testing"

	"github.com/gofromzero/mer-sys/backend/services/order-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTenantIsolation(t *testing.T) {
	Convey("多租户隔离测试", t, func() {

		Convey("订单多租户隔离", func() {
			// 创建两个不同租户的上下文
			ctx1 := context.WithValue(context.Background(), "tenant_id", uint64(1))
			ctx2 := context.WithValue(context.Background(), "tenant_id", uint64(2))

			orderService := service.NewOrderService()

			// 租户1创建订单
			req1 := &types.CreateOrderRequest{
				MerchantID: 1,
				Items: []struct {
					ProductID uint64 `json:"product_id" v:"required#商品ID不能为空"`
					Quantity  int    `json:"quantity" v:"required|min:1#数量不能为空|数量必须大于0"`
				}{
					{ProductID: 1001, Quantity: 1},
				},
			}

			order1, err := orderService.CreateOrder(ctx1, 1, req1)
			So(err, ShouldBeNil)
			So(order1, ShouldNotBeNil)
			So(order1.TenantID, ShouldEqual, 1)

			// 租户2创建订单
			req2 := &types.CreateOrderRequest{
				MerchantID: 1,
				Items: []struct {
					ProductID uint64 `json:"product_id" v:"required#商品ID不能为空"`
					Quantity  int    `json:"quantity" v:"required|min:1#数量不能为空|数量必须大于0"`
				}{
					{ProductID: 1002, Quantity: 2},
				},
			}

			order2, err := orderService.CreateOrder(ctx2, 2, req2)
			So(err, ShouldBeNil)
			So(order2, ShouldNotBeNil)
			So(order2.TenantID, ShouldEqual, 2)

			// 租户1无法获取租户2的订单
			_, err = orderService.GetOrder(ctx1, order2.ID)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "订单不存在")

			// 租户2无法获取租户1的订单
			_, err = orderService.GetOrder(ctx2, order1.ID)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "订单不存在")

			// 每个租户只能看到自己的订单列表
			orders1, total1, err := orderService.ListOrders(ctx1, 1, "", 1, 10)
			So(err, ShouldBeNil)
			So(len(orders1), ShouldEqual, 1)
			So(total1, ShouldEqual, 1)
			So(orders1[0].TenantID, ShouldEqual, 1)

			orders2, total2, err := orderService.ListOrders(ctx2, 2, "", 1, 10)
			So(err, ShouldBeNil)
			So(len(orders2), ShouldEqual, 1)
			So(total2, ShouldEqual, 1)
			So(orders2[0].TenantID, ShouldEqual, 2)
		})

		Convey("购物车多租户隔离", func() {
			ctx1 := context.WithValue(context.Background(), "tenant_id", uint64(1))
			ctx2 := context.WithValue(context.Background(), "tenant_id", uint64(2))

			cartService := service.NewCartService()

			// 租户1添加商品到购物车
			err := cartService.AddItem(ctx1, 1, 1001, 2)
			So(err, ShouldBeNil)

			// 租户2添加商品到购物车
			err = cartService.AddItem(ctx2, 2, 1002, 1)
			So(err, ShouldBeNil)

			// 获取购物车，验证租户隔离
			cart1, err := cartService.GetCart(ctx1, 1)
			So(err, ShouldBeNil)
			So(cart1, ShouldNotBeNil)
			So(cart1.TenantID, ShouldEqual, 1)
			So(len(cart1.Items), ShouldEqual, 1)
			So(cart1.Items[0].ProductID, ShouldEqual, 1001)

			cart2, err := cartService.GetCart(ctx2, 2)
			So(err, ShouldBeNil)
			So(cart2, ShouldNotBeNil)
			So(cart2.TenantID, ShouldEqual, 2)
			So(len(cart2.Items), ShouldEqual, 1)
			So(cart2.Items[0].ProductID, ShouldEqual, 1002)
		})

		Convey("Repository层租户隔离", func() {
			ctx1 := context.WithValue(context.Background(), "tenant_id", uint64(1))
			ctx2 := context.WithValue(context.Background(), "tenant_id", uint64(2))

			orderRepo := repository.NewOrderRepository()

			// 创建两个不同租户的订单
			order1 := &types.Order{
				TenantID:    1,
				MerchantID:  1,
				CustomerID:  1,
				OrderNumber: "TEST001",
				Status:      types.OrderStatusPending,
				Items:       []types.OrderItem{},
				TotalAmount: types.Money{Amount: 100.0, Currency: "CNY"},
			}

			order2 := &types.Order{
				TenantID:    2,
				MerchantID:  1,
				CustomerID:  1,
				OrderNumber: "TEST002",
				Status:      types.OrderStatusPending,
				Items:       []types.OrderItem{},
				TotalAmount: types.Money{Amount: 200.0, Currency: "CNY"},
			}

			err := orderRepo.Create(ctx1, order1)
			So(err, ShouldBeNil)

			err = orderRepo.Create(ctx2, order2)
			So(err, ShouldBeNil)

			// 租户1无法查询到租户2的订单
			_, err = orderRepo.GetByID(ctx1, order2.ID)
			So(err, ShouldNotBeNil)

			// 租户2无法查询到租户1的订单
			_, err = orderRepo.GetByID(ctx2, order1.ID)
			So(err, ShouldNotBeNil)

			// 每个租户只能查询到自己的订单
			fetchedOrder1, err := orderRepo.GetByID(ctx1, order1.ID)
			So(err, ShouldBeNil)
			So(fetchedOrder1.TenantID, ShouldEqual, 1)

			fetchedOrder2, err := orderRepo.GetByID(ctx2, order2.ID)
			So(err, ShouldBeNil)
			So(fetchedOrder2.TenantID, ShouldEqual, 2)
		})

		Convey("上下文租户ID验证", func() {
			// 缺少租户ID的上下文应该返回错误
			ctxNoTenant := context.Background()
			orderRepo := repository.NewOrderRepository()

			order := &types.Order{
				MerchantID:  1,
				CustomerID:  1,
				OrderNumber: "TEST003",
				Status:      types.OrderStatusPending,
				Items:       []types.OrderItem{},
				TotalAmount: types.Money{Amount: 100.0, Currency: "CNY"},
			}

			err := orderRepo.Create(ctxNoTenant, order)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "tenant_id")
		})

		Convey("支付多租户隔离", func() {
			ctx1 := context.WithValue(context.Background(), "tenant_id", uint64(1))
			ctx2 := context.WithValue(context.Background(), "tenant_id", uint64(2))

			paymentService := service.NewPaymentService()
			orderService := service.NewOrderService()

			// 租户1创建订单
			req1 := &types.CreateOrderRequest{
				MerchantID: 1,
				Items: []struct {
					ProductID uint64 `json:"product_id" v:"required#商品ID不能为空"`
					Quantity  int    `json:"quantity" v:"required|min:1#数量不能为空|数量必须大于0"`
				}{
					{ProductID: 1001, Quantity: 1},
				},
			}

			order1, err := orderService.CreateOrder(ctx1, 1, req1)
			So(err, ShouldBeNil)

			// 租户2创建订单
			req2 := &types.CreateOrderRequest{
				MerchantID: 1,
				Items: []struct {
					ProductID uint64 `json:"product_id" v:"required#商品ID不能为空"`
					Quantity  int    `json:"quantity" v:"required|min:1#数量不能为空|数量必须大于0"`
				}{
					{ProductID: 1002, Quantity: 1},
				},
			}

			order2, err := orderService.CreateOrder(ctx2, 2, req2)
			So(err, ShouldBeNil)

			// 租户1无法为租户2的订单发起支付
			_, err = paymentService.InitiatePayment(ctx1, order2.ID, types.PaymentMethodAlipay, "")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "订单不存在")

			// 租户2无法为租户1的订单发起支付
			_, err = paymentService.InitiatePayment(ctx2, order1.ID, types.PaymentMethodAlipay, "")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "订单不存在")

			// 租户只能为自己的订单发起支付
			paymentInfo1, err := paymentService.InitiatePayment(ctx1, order1.ID, types.PaymentMethodAlipay, "")
			So(err, ShouldBeNil)
			So(paymentInfo1, ShouldNotBeNil)

			paymentInfo2, err := paymentService.InitiatePayment(ctx2, order2.ID, types.PaymentMethodAlipay, "")
			So(err, ShouldBeNil)
			So(paymentInfo2, ShouldNotBeNil)
		})
	})
}
