package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// IOrderRepository 订单仓储接口
type IOrderRepository interface {
	Create(ctx context.Context, order *types.Order) error
	GetByID(ctx context.Context, id uint64) (*types.Order, error)
	GetByOrderNumber(ctx context.Context, orderNumber string) (*types.Order, error)
	List(ctx context.Context, customerID uint64, status types.OrderStatus, page, limit int) ([]*types.Order, int, error)
	Update(ctx context.Context, order *types.Order) error
	UpdateStatus(ctx context.Context, id uint64, status types.OrderStatus) error
	GenerateOrderNumber(ctx context.Context) (string, error)
}

// OrderRepository 订单仓储实现
type OrderRepository struct {
	*BaseRepository
}

// NewOrderRepository 创建订单仓储实例
func NewOrderRepository() IOrderRepository {
	return &OrderRepository{
		BaseRepository: NewBaseRepository(),
	}
}

// Create 创建订单
func (r *OrderRepository) Create(ctx context.Context, order *types.Order) error {
	tenantID := r.GetTenantID(ctx)
	
	itemsJSON, err := json.Marshal(order.Items)
	if err != nil {
		return fmt.Errorf("序列化订单项失败: %v", err)
	}
	
	paymentInfoJSON, err := json.Marshal(order.PaymentInfo)
	if err != nil {
		return fmt.Errorf("序列化支付信息失败: %v", err)
	}
	
	verificationInfoJSON := "null"
	if order.VerificationInfo != nil {
		verificationBytes, err := json.Marshal(order.VerificationInfo)
		if err != nil {
			return fmt.Errorf("序列化核销信息失败: %v", err)
		}
		verificationInfoJSON = string(verificationBytes)
	}
	
	result, err := g.DB().Model("orders").Ctx(ctx).Insert(gdb.Map{
		"tenant_id":           tenantID,
		"merchant_id":         order.MerchantID,
		"customer_id":         order.CustomerID,
		"order_number":        order.OrderNumber,
		"status":              order.Status,
		"items":               string(itemsJSON),
		"payment_info":        string(paymentInfoJSON),
		"verification_info":   verificationInfoJSON,
		"total_amount":        order.TotalAmount,
		"total_rights_cost":   order.TotalRightsCost,
		"created_at":          gtime.Now(),
		"updated_at":          gtime.Now(),
	})
	
	if err != nil {
		return fmt.Errorf("创建订单失败: %v", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取订单ID失败: %v", err)
	}
	
	order.ID = uint64(id)
	return nil
}

// GetByID 根据ID获取订单
func (r *OrderRepository) GetByID(ctx context.Context, id uint64) (*types.Order, error) {
	tenantID := r.GetTenantID(ctx)
	
	var orderData struct {
		types.Order
		ItemsJSON           string `db:"items"`
		PaymentInfoJSON     string `db:"payment_info"`
		VerificationInfoJSON string `db:"verification_info"`
	}
	
	err := g.DB().Model("orders").Ctx(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Scan(&orderData)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("订单不存在")
		}
		return nil, fmt.Errorf("查询订单失败: %v", err)
	}
	
	// 反序列化订单项
	if err := json.Unmarshal([]byte(orderData.ItemsJSON), &orderData.Order.Items); err != nil {
		return nil, fmt.Errorf("反序列化订单项失败: %v", err)
	}
	
	// 反序列化支付信息
	if orderData.PaymentInfoJSON != "" && orderData.PaymentInfoJSON != "null" {
		if err := json.Unmarshal([]byte(orderData.PaymentInfoJSON), &orderData.Order.PaymentInfo); err != nil {
			return nil, fmt.Errorf("反序列化支付信息失败: %v", err)
		}
	}
	
	// 反序列化核销信息
	if orderData.VerificationInfoJSON != "" && orderData.VerificationInfoJSON != "null" {
		if err := json.Unmarshal([]byte(orderData.VerificationInfoJSON), &orderData.Order.VerificationInfo); err != nil {
			return nil, fmt.Errorf("反序列化核销信息失败: %v", err)
		}
	}
	
	return &orderData.Order, nil
}

// GetByOrderNumber 根据订单号获取订单
func (r *OrderRepository) GetByOrderNumber(ctx context.Context, orderNumber string) (*types.Order, error) {
	tenantID := r.GetTenantID(ctx)
	
	var orderData struct {
		types.Order
		ItemsJSON           string `db:"items"`
		PaymentInfoJSON     string `db:"payment_info"`
		VerificationInfoJSON string `db:"verification_info"`
	}
	
	err := g.DB().Model("orders").Ctx(ctx).
		Where("order_number = ? AND tenant_id = ?", orderNumber, tenantID).
		Scan(&orderData)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("订单不存在")
		}
		return nil, fmt.Errorf("查询订单失败: %v", err)
	}
	
	// 反序列化订单项
	if err := json.Unmarshal([]byte(orderData.ItemsJSON), &orderData.Order.Items); err != nil {
		return nil, fmt.Errorf("反序列化订单项失败: %v", err)
	}
	
	// 反序列化支付信息
	if orderData.PaymentInfoJSON != "" && orderData.PaymentInfoJSON != "null" {
		if err := json.Unmarshal([]byte(orderData.PaymentInfoJSON), &orderData.Order.PaymentInfo); err != nil {
			return nil, fmt.Errorf("反序列化支付信息失败: %v", err)
		}
	}
	
	// 反序列化核销信息
	if orderData.VerificationInfoJSON != "" && orderData.VerificationInfoJSON != "null" {
		if err := json.Unmarshal([]byte(orderData.VerificationInfoJSON), &orderData.Order.VerificationInfo); err != nil {
			return nil, fmt.Errorf("反序列化核销信息失败: %v", err)
		}
	}
	
	return &orderData.Order, nil
}

// List 获取订单列表
func (r *OrderRepository) List(ctx context.Context, customerID uint64, status types.OrderStatus, page, limit int) ([]*types.Order, int, error) {
	tenantID := r.GetTenantID(ctx)
	
	query := g.DB().Model("orders").Ctx(ctx).Where("tenant_id = ?", tenantID)
	
	if customerID > 0 {
		query = query.Where("customer_id = ?", customerID)
	}
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	// 获取总数
	count, err := query.Count()
	if err != nil {
		return nil, 0, fmt.Errorf("查询订单总数失败: %v", err)
	}
	
	// 分页查询
	var orderDataList []struct {
		types.Order
		ItemsJSON           string `db:"items"`
		PaymentInfoJSON     string `db:"payment_info"`
		VerificationInfoJSON string `db:"verification_info"`
	}
	
	err = query.Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Scan(&orderDataList)
	
	if err != nil {
		return nil, 0, fmt.Errorf("查询订单列表失败: %v", err)
	}
	
	// 构造返回结果
	orders := make([]*types.Order, 0, len(orderDataList))
	for _, orderData := range orderDataList {
		// 反序列化订单项
		if err := json.Unmarshal([]byte(orderData.ItemsJSON), &orderData.Order.Items); err != nil {
			return nil, 0, fmt.Errorf("反序列化订单项失败: %v", err)
		}
		
		// 反序列化支付信息
		if orderData.PaymentInfoJSON != "" && orderData.PaymentInfoJSON != "null" {
			if err := json.Unmarshal([]byte(orderData.PaymentInfoJSON), &orderData.Order.PaymentInfo); err != nil {
				return nil, 0, fmt.Errorf("反序列化支付信息失败: %v", err)
			}
		}
		
		// 反序列化核销信息
		if orderData.VerificationInfoJSON != "" && orderData.VerificationInfoJSON != "null" {
			if err := json.Unmarshal([]byte(orderData.VerificationInfoJSON), &orderData.Order.VerificationInfo); err != nil {
				return nil, 0, fmt.Errorf("反序列化核销信息失败: %v", err)
			}
		}
		
		orders = append(orders, &orderData.Order)
	}
	
	return orders, count, nil
}

// Update 更新订单
func (r *OrderRepository) Update(ctx context.Context, order *types.Order) error {
	tenantID := r.GetTenantID(ctx)
	
	itemsJSON, err := json.Marshal(order.Items)
	if err != nil {
		return fmt.Errorf("序列化订单项失败: %v", err)
	}
	
	paymentInfoJSON, err := json.Marshal(order.PaymentInfo)
	if err != nil {
		return fmt.Errorf("序列化支付信息失败: %v", err)
	}
	
	verificationInfoJSON := "null"
	if order.VerificationInfo != nil {
		verificationBytes, err := json.Marshal(order.VerificationInfo)
		if err != nil {
			return fmt.Errorf("序列化核销信息失败: %v", err)
		}
		verificationInfoJSON = string(verificationBytes)
	}
	
	_, err = g.DB().Model("orders").Ctx(ctx).
		Where("id = ? AND tenant_id = ?", order.ID, tenantID).
		Update(gdb.Map{
			"merchant_id":         order.MerchantID,
			"customer_id":         order.CustomerID,
			"order_number":        order.OrderNumber,
			"status":              order.Status,
			"items":               string(itemsJSON),
			"payment_info":        string(paymentInfoJSON),
			"verification_info":   verificationInfoJSON,
			"total_amount":        order.TotalAmount,
			"total_rights_cost":   order.TotalRightsCost,
			"updated_at":          gtime.Now(),
		})
	
	if err != nil {
		return fmt.Errorf("更新订单失败: %v", err)
	}
	
	return nil
}

// UpdateStatus 更新订单状态
func (r *OrderRepository) UpdateStatus(ctx context.Context, id uint64, status types.OrderStatus) error {
	tenantID := r.GetTenantID(ctx)
	
	_, err := g.DB().Model("orders").Ctx(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Update(gdb.Map{
			"status":     status,
			"updated_at": gtime.Now(),
		})
	
	if err != nil {
		return fmt.Errorf("更新订单状态失败: %v", err)
	}
	
	return nil
}

// GenerateOrderNumber 生成订单号
func (r *OrderRepository) GenerateOrderNumber(ctx context.Context) (string, error) {
	// 订单号格式：年月日时分秒 + 6位随机数
	now := gtime.Now()
	timePrefix := now.Format("20060102150405")
	
	// 使用数据库的自增特性来确保唯一性
	count, err := g.DB().Model("orders").Ctx(ctx).
		Where("created_at >= ?", now.Format("2006-01-02 00:00:00")).
		Count()
	
	if err != nil {
		return "", fmt.Errorf("查询订单计数失败: %v", err)
	}
	
	return fmt.Sprintf("%s%06d", timePrefix, count+1), nil
}