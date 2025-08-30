package unit

import (
	"context"
	"testing"
	"time"

	"github.com/gofromzero/mer-sys/backend/services/order-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	. "github.com/smartystreets/goconvey/convey"
)

func TestOrderStatusService(t *testing.T) {
	Convey("订单状态管理服务测试", t, func() {
		ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))
		ctx = context.WithValue(ctx, "user_id", uint64(100))
		orderStatusService := service.NewOrderStatusService()

		Convey("订单状态流转验证", func() {
			Convey("有效状态流转", func() {
				// 待支付 -> 已支付
				valid := types.OrderStatusIntPending.IsValidTransition(types.OrderStatusIntPaid)
				So(valid, ShouldBeTrue)

				// 已支付 -> 处理中
				valid = types.OrderStatusIntPaid.IsValidTransition(types.OrderStatusIntProcessing)
				So(valid, ShouldBeTrue)

				// 处理中 -> 已完成
				valid = types.OrderStatusIntProcessing.IsValidTransition(types.OrderStatusIntCompleted)
				So(valid, ShouldBeTrue)

				// 待支付 -> 已取消
				valid = types.OrderStatusIntPending.IsValidTransition(types.OrderStatusIntCancelled)
				So(valid, ShouldBeTrue)

				// 已支付 -> 已取消
				valid = types.OrderStatusIntPaid.IsValidTransition(types.OrderStatusIntCancelled)
				So(valid, ShouldBeTrue)
			})

			Convey("无效状态流转", func() {
				// 待支付 -> 处理中（跳过已支付）
				valid := types.OrderStatusIntPending.IsValidTransition(types.OrderStatusIntProcessing)
				So(valid, ShouldBeFalse)

				// 已完成 -> 处理中（反向流转）
				valid = types.OrderStatusIntCompleted.IsValidTransition(types.OrderStatusIntProcessing)
				So(valid, ShouldBeFalse)

				// 已取消 -> 已支付（死状态复活）
				valid = types.OrderStatusIntCancelled.IsValidTransition(types.OrderStatusIntPaid)
				So(valid, ShouldBeFalse)

				// 已完成 -> 已取消（最终状态变更）
				valid = types.OrderStatusIntCompleted.IsValidTransition(types.OrderStatusIntCancelled)
				So(valid, ShouldBeFalse)
			})
		})

		Convey("单个订单状态更新", func() {
			// 首先创建一个测试订单
			orderService := service.NewOrderService()
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
			So(order.Status, ShouldEqual, types.OrderStatusPending)

			Convey("成功更新订单状态", func() {
				updateReq := &types.UpdateOrderStatusRequest{
					Status:       types.OrderStatusIntPaid,
					Reason:       "客户完成支付",
					OperatorType: types.OrderStatusOperatorTypeCustomer,
				}

				err = orderStatusService.UpdateOrderStatus(ctx, order.ID, updateReq)
				So(err, ShouldBeNil)

				// 验证状态历史记录
				history, err := orderStatusService.GetOrderStatusHistory(ctx, order.ID)
				So(err, ShouldBeNil)
				So(len(history), ShouldBeGreaterThan, 0)

				latestHistory := history[0] // 按创建时间倒序
				So(latestHistory.FromStatus, ShouldEqual, types.OrderStatusIntPending)
				So(latestHistory.ToStatus, ShouldEqual, types.OrderStatusIntPaid)
				So(latestHistory.Reason, ShouldEqual, "客户完成支付")
				So(latestHistory.OperatorType, ShouldEqual, types.OrderStatusOperatorTypeCustomer)
			})

			Convey("状态流转验证失败", func() {
				updateReq := &types.UpdateOrderStatusRequest{
					Status:       types.OrderStatusIntProcessing, // 跳过已支付状态
					Reason:       "直接处理订单",
					OperatorType: types.OrderStatusOperatorTypeMerchant,
				}

				err := orderStatusService.UpdateOrderStatus(ctx, order.ID, updateReq)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "状态转换")
			})

			Convey("重复状态更新", func() {
				// 先更新到已支付
				updateReq := &types.UpdateOrderStatusRequest{
					Status:       types.OrderStatusIntPaid,
					Reason:       "客户完成支付",
					OperatorType: types.OrderStatusOperatorTypeCustomer,
				}
				err := orderStatusService.UpdateOrderStatus(ctx, order.ID, updateReq)
				So(err, ShouldBeNil)

				// 再次更新到相同状态
				err = orderStatusService.UpdateOrderStatus(ctx, order.ID, updateReq)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("批量订单状态更新", func() {
			// 创建多个测试订单
			orderService := service.NewOrderService()
			req := &types.CreateOrderRequest{
				MerchantID: 1,
				Items: []struct {
					ProductID uint64 `json:"product_id" v:"required#商品ID不能为空"`
					Quantity  int    `json:"quantity" v:"required|min:1#数量不能为空|数量必须大于0"`
				}{
					{ProductID: 1001, Quantity: 1},
				},
			}

			var orderIDs []uint64
			for i := 0; i < 3; i++ {
				order, err := orderService.CreateOrder(ctx, 1, req)
				So(err, ShouldBeNil)
				orderIDs = append(orderIDs, order.ID)
			}

			Convey("成功批量更新状态", func() {
				batchReq := &types.BatchUpdateOrderStatusRequest{
					OrderIDs:     orderIDs,
					Status:       types.OrderStatusIntCancelled,
					Reason:       "批量取消订单",
					OperatorType: types.OrderStatusOperatorTypeMerchant,
				}

				result, err := orderStatusService.BatchUpdateOrderStatus(ctx, batchReq)
				So(err, ShouldBeNil)
				So(result.SuccessCount, ShouldEqual, 3)
				So(result.FailCount, ShouldEqual, 0)
				So(len(result.Errors), ShouldEqual, 0)

				// 验证所有订单状态都已更新
				for _, orderID := range orderIDs {
					history, err := orderStatusService.GetOrderStatusHistory(ctx, orderID)
					So(err, ShouldBeNil)
					So(len(history), ShouldBeGreaterThan, 0)

					latestHistory := history[0]
					So(latestHistory.ToStatus, ShouldEqual, types.OrderStatusIntCancelled)
					So(latestHistory.Reason, ShouldEqual, "批量取消订单")
				}
			})

			Convey("部分订单更新失败", func() {
				// 先将一个订单设为已完成
				firstOrderUpdateReq := &types.UpdateOrderStatusRequest{
					Status:       types.OrderStatusIntPaid,
					Reason:       "支付完成",
					OperatorType: types.OrderStatusOperatorTypeCustomer,
				}
				err := orderStatusService.UpdateOrderStatus(ctx, orderIDs[0], firstOrderUpdateReq)
				So(err, ShouldBeNil)

				firstOrderUpdateReq.Status = types.OrderStatusIntProcessing
				firstOrderUpdateReq.Reason = "开始处理"
				firstOrderUpdateReq.OperatorType = types.OrderStatusOperatorTypeMerchant
				err = orderStatusService.UpdateOrderStatus(ctx, orderIDs[0], firstOrderUpdateReq)
				So(err, ShouldBeNil)

				firstOrderUpdateReq.Status = types.OrderStatusIntCompleted
				firstOrderUpdateReq.Reason = "处理完成"
				err = orderStatusService.UpdateOrderStatus(ctx, orderIDs[0], firstOrderUpdateReq)
				So(err, ShouldBeNil)

				// 尝试批量取消（已完成的订单无法取消）
				batchReq := &types.BatchUpdateOrderStatusRequest{
					OrderIDs:     orderIDs,
					Status:       types.OrderStatusIntCancelled,
					Reason:       "批量取消订单",
					OperatorType: types.OrderStatusOperatorTypeMerchant,
				}

				result, err := orderStatusService.BatchUpdateOrderStatus(ctx, batchReq)
				So(err, ShouldBeNil)
				So(result.SuccessCount, ShouldEqual, 2)
				So(result.FailCount, ShouldEqual, 1)
				So(len(result.Errors), ShouldEqual, 1)
				So(result.Errors[0].OrderID, ShouldEqual, orderIDs[0])
				So(result.Errors[0].Error, ShouldContainSubstring, "状态转换")
			})
		})

		Convey("订单状态历史查询", func() {
			// 创建订单并进行多次状态变更
			orderService := service.NewOrderService()
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

			// 模拟完整的订单流程
			statusUpdates := []struct {
				status       types.OrderStatusInt
				reason       string
				operatorType types.OrderStatusOperatorType
			}{
				{types.OrderStatusIntPaid, "客户完成支付", types.OrderStatusOperatorTypeCustomer},
				{types.OrderStatusIntProcessing, "商户开始处理", types.OrderStatusOperatorTypeMerchant},
				{types.OrderStatusIntCompleted, "订单处理完成", types.OrderStatusOperatorTypeMerchant},
			}

			for _, update := range statusUpdates {
				updateReq := &types.UpdateOrderStatusRequest{
					Status:       update.status,
					Reason:       update.reason,
					OperatorType: update.operatorType,
				}
				err := orderStatusService.UpdateOrderStatus(ctx, order.ID, updateReq)
				So(err, ShouldBeNil)
			}

			Convey("获取完整状态历史", func() {
				history, err := orderStatusService.GetOrderStatusHistory(ctx, order.ID)
				So(err, ShouldBeNil)
				So(len(history), ShouldEqual, 3) // 3次状态变更

				// 验证历史记录按时间倒序排列
				So(history[0].ToStatus, ShouldEqual, types.OrderStatusIntCompleted)
				So(history[1].ToStatus, ShouldEqual, types.OrderStatusIntProcessing)
				So(history[2].ToStatus, ShouldEqual, types.OrderStatusIntPaid)

				// 验证操作者类型
				So(history[0].OperatorType, ShouldEqual, types.OrderStatusOperatorTypeMerchant)
				So(history[1].OperatorType, ShouldEqual, types.OrderStatusOperatorTypeMerchant)
				So(history[2].OperatorType, ShouldEqual, types.OrderStatusOperatorTypeCustomer)
			})
		})

		Convey("多租户数据隔离", func() {
			// 租户1创建订单
			tenant1Ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))
			tenant2Ctx := context.WithValue(context.Background(), "tenant_id", uint64(2))

			orderService := service.NewOrderService()
			req := &types.CreateOrderRequest{
				MerchantID: 1,
				Items: []struct {
					ProductID uint64 `json:"product_id" v:"required#商品ID不能为空"`
					Quantity  int    `json:"quantity" v:"required|min:1#数量不能为空|数量必须大于0"`
				}{
					{ProductID: 1001, Quantity: 1},
				},
			}

			tenant1Order, err := orderService.CreateOrder(tenant1Ctx, 1, req)
			So(err, ShouldBeNil)

			// 租户2尝试访问租户1的订单状态
			_, err = orderStatusService.GetOrderStatusHistory(tenant2Ctx, tenant1Order.ID)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "订单不存在或无权限访问")

			// 租户2尝试更新租户1的订单状态
			updateReq := &types.UpdateOrderStatusRequest{
				Status:       types.OrderStatusIntCancelled,
				Reason:       "尝试跨租户操作",
				OperatorType: types.OrderStatusOperatorTypeSystem,
			}

			err = orderStatusService.UpdateOrderStatus(tenant2Ctx, tenant1Order.ID, updateReq)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "订单不存在")
		})
	})
}