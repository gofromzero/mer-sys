package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// IOrderRepository 订单仓储接口
type IOrderRepository interface {
	Create(ctx context.Context, order *types.Order) error
	GetByID(ctx context.Context, id uint64) (*types.Order, error)
	GetByIDWithHistory(ctx context.Context, id uint64) (*types.Order, error)
	GetByOrderNumber(ctx context.Context, orderNumber string) (*types.Order, error)
	List(ctx context.Context, customerID uint64, status types.OrderStatus, page, limit int) ([]*types.Order, int, error)
	QueryList(ctx context.Context, req *types.OrderQueryRequest) (*types.OrderListResponse, error)
	Update(ctx context.Context, order *types.Order) error
	UpdateStatus(ctx context.Context, id uint64, status types.OrderStatus) error
	UpdateStatusWithHistory(ctx context.Context, id uint64, status types.OrderStatusInt, reason string, operatorType types.OrderStatusOperatorType, operatorID *uint64, metadata interface{}) error
	BatchUpdateStatus(ctx context.Context, req *types.BatchUpdateOrderStatusRequest, operatorID *uint64) (*types.BatchUpdateOrderStatusResponse, error)
	GenerateOrderNumber(ctx context.Context) (string, error)
	GetTimeoutOrders(ctx context.Context, timeoutConfig *types.OrderTimeoutConfig) ([]*types.Order, error)
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

// GetByIDWithHistory 根据ID获取订单（包含状态历史）
func (r *OrderRepository) GetByIDWithHistory(ctx context.Context, id uint64) (*types.Order, error) {
	order, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// 获取状态历史
	historyRepo := NewOrderStatusHistoryRepository()
	history, err := historyRepo.GetByOrderID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取订单状态历史失败: %v", err)
	}
	
	order.StatusHistory = history
	return order, nil
}

// QueryList 根据查询条件获取订单列表
func (r *OrderRepository) QueryList(ctx context.Context, req *types.OrderQueryRequest) (*types.OrderListResponse, error) {
	tenantID := r.GetTenantID(ctx)
	
	// 构建查询条件
	query := g.DB().Model("orders o").Ctx(ctx).Where("o.tenant_id = ?", tenantID)
	
	if req.MerchantID != nil {
		query = query.Where("o.merchant_id = ?", *req.MerchantID)
	}
	
	if req.CustomerID != nil {
		query = query.Where("o.customer_id = ?", *req.CustomerID)
	}
	
	if len(req.Status) > 0 {
		statusValues := make([]interface{}, len(req.Status))
		for i, status := range req.Status {
			statusValues[i] = status
		}
		query = query.Where("o.status IN (?)", statusValues)
	}
	
	if req.StartDate != nil {
		query = query.Where("o.created_at >= ?", req.StartDate.Format("2006-01-02 00:00:00"))
	}
	
	if req.EndDate != nil {
		query = query.Where("o.created_at <= ?", req.EndDate.Format("2006-01-02 23:59:59"))
	}
	
	if req.SearchKeyword != nil && *req.SearchKeyword != "" {
		keyword := "%" + *req.SearchKeyword + "%"
		query = query.Where("(o.order_number LIKE ? OR JSON_UNQUOTE(JSON_EXTRACT(o.items, '$[*].product_name')) LIKE ?)", keyword, keyword)
	}
	
	// 获取总数
	total, err := query.Count()
	if err != nil {
		return nil, fmt.Errorf("获取订单总数失败: %v", err)
	}
	
	// 构建完整查询（包含用户和商户名称）
	// 构建查询参数
	queryParams := []interface{}{tenantID}
	queryParams = append(queryParams, r.buildWhereParams(req)...)
	queryParams = append(queryParams, req.PageSize, (req.Page-1)*req.PageSize)
	
	fullQuery := g.DB().Ctx(ctx).Raw(`
		SELECT 
			o.id, o.order_number, o.status, o.total_amount, o.created_at, o.updated_at,
			COUNT(JSON_EXTRACT(o.items, '$[*]')) as item_count,
			u.username as customer_name,
			m.name as merchant_name
		FROM orders o
		LEFT JOIN users u ON o.customer_id = u.id AND u.tenant_id = o.tenant_id
		LEFT JOIN merchants m ON o.merchant_id = m.id AND m.tenant_id = o.tenant_id
		WHERE o.tenant_id = ? ` + r.buildWhereClause(req) + `
		ORDER BY o.` + req.SortBy + ` ` + strings.ToUpper(req.SortOrder) + `
		LIMIT ? OFFSET ?
	`, queryParams...)
	
	var summaries []types.OrderSummary
	err = fullQuery.Scan(&summaries)
	if err != nil {
		return nil, fmt.Errorf("查询订单列表失败: %v", err)
	}
	
	// 批量获取最新状态变更历史
	orderIDs := make([]uint64, len(summaries))
	for i, summary := range summaries {
		orderIDs[i] = summary.ID
	}
	
	historyRepo := NewOrderStatusHistoryRepository()
	historyMap, err := historyRepo.GetByOrderIDs(ctx, orderIDs)
	if err != nil {
		return nil, fmt.Errorf("获取订单状态历史失败: %v", err)
	}
	
	// 填充状态历史
	for i := range summaries {
		if history, exists := historyMap[summaries[i].ID]; exists {
			summaries[i].LatestStatusChange = history
		}
	}
	
	return &types.OrderListResponse{
		Items:    summaries,
		Total:    int64(total),
		Page:     req.Page,
		PageSize: req.PageSize,
		HasNext:  int64((req.Page)*req.PageSize) < int64(total),
	}, nil
}

// buildWhereClause 构建WHERE子句（不含WHERE关键字）
func (r *OrderRepository) buildWhereClause(req *types.OrderQueryRequest) string {
	conditions := []string{}
	
	if req.MerchantID != nil {
		conditions = append(conditions, "AND o.merchant_id = ?")
	}
	
	if req.CustomerID != nil {
		conditions = append(conditions, "AND o.customer_id = ?")
	}
	
	if len(req.Status) > 0 {
		placeholders := strings.Repeat("?,", len(req.Status))
		placeholders = placeholders[:len(placeholders)-1] // 去掉最后一个逗号
		conditions = append(conditions, "AND o.status IN ("+placeholders+")")
	}
	
	if req.StartDate != nil {
		conditions = append(conditions, "AND o.created_at >= ?")
	}
	
	if req.EndDate != nil {
		conditions = append(conditions, "AND o.created_at <= ?")
	}
	
	if req.SearchKeyword != nil && *req.SearchKeyword != "" {
		conditions = append(conditions, "AND (o.order_number LIKE ? OR JSON_UNQUOTE(JSON_EXTRACT(o.items, '$[*].product_name')) LIKE ?)")
	}
	
	return strings.Join(conditions, " ")
}

// buildWhereParams 构建WHERE参数
func (r *OrderRepository) buildWhereParams(req *types.OrderQueryRequest) []interface{} {
	params := []interface{}{}
	
	if req.MerchantID != nil {
		params = append(params, *req.MerchantID)
	}
	
	if req.CustomerID != nil {
		params = append(params, *req.CustomerID)
	}
	
	if len(req.Status) > 0 {
		for _, status := range req.Status {
			params = append(params, status)
		}
	}
	
	if req.StartDate != nil {
		params = append(params, req.StartDate.Format("2006-01-02 00:00:00"))
	}
	
	if req.EndDate != nil {
		params = append(params, req.EndDate.Format("2006-01-02 23:59:59"))
	}
	
	if req.SearchKeyword != nil && *req.SearchKeyword != "" {
		keyword := "%" + *req.SearchKeyword + "%"
		params = append(params, keyword, keyword)
	}
	
	return params
}

// UpdateStatusWithHistory 更新订单状态并记录历史
func (r *OrderRepository) UpdateStatusWithHistory(ctx context.Context, id uint64, status types.OrderStatusInt, reason string, operatorType types.OrderStatusOperatorType, operatorID *uint64, metadata interface{}) error {
	// 获取当前订单状态
	currentOrder, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("获取订单失败: %v", err)
	}
	
	// 转换当前状态为数字格式进行比较
	currentStatusInt := r.orderStatusToInt(currentOrder.Status)
	
	// 验证状态转换是否合法
	if !currentStatusInt.IsValidTransition(status) {
		return fmt.Errorf("不允许从状态 %s 转换到 %s", currentStatusInt.String(), status.String())
	}
	
	// 开启事务
	tx, err := g.DB().Begin(ctx)
	if err != nil {
		return fmt.Errorf("开启事务失败: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	
	// 更新订单状态
	newStatus := status.ToOrderStatus()
	tenantID := r.GetTenantID(ctx)
	
	_, err = tx.Model("orders").Ctx(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Update(gdb.Map{
			"status":            newStatus,
			"status_updated_at": gtime.Now(),
			"updated_at":       gtime.Now(),
		})
	if err != nil {
		return fmt.Errorf("更新订单状态失败: %v", err)
	}
	
	// 创建状态历史记录
	history := &types.OrderStatusHistory{
		TenantID:     tenantID,
		OrderID:      id,
		FromStatus:   currentStatusInt,
		ToStatus:     status,
		Reason:       reason,
		OperatorID:   operatorID,
		OperatorType: operatorType,
		Metadata:     metadata,
		CreatedAt:    time.Now(),
	}
	
	_, err = tx.Model("order_status_history").Ctx(ctx).Data(history).Insert()
	if err != nil {
		return fmt.Errorf("创建状态历史记录失败: %v", err)
	}
	
	return nil
}

// orderStatusToInt 将字符串状态转换为数字状态
func (r *OrderRepository) orderStatusToInt(status types.OrderStatus) types.OrderStatusInt {
	switch status {
	case "pending":
		return types.OrderStatusIntPending
	case "paid":
		return types.OrderStatusIntPaid
	case "processing":
		return types.OrderStatusIntProcessing
	case "completed":
		return types.OrderStatusIntCompleted
	case "cancelled":
		return types.OrderStatusIntCancelled
	default:
		return types.OrderStatusIntPending
	}
}

// BatchUpdateStatus 批量更新订单状态
func (r *OrderRepository) BatchUpdateStatus(ctx context.Context, req *types.BatchUpdateOrderStatusRequest, operatorID *uint64) (*types.BatchUpdateOrderStatusResponse, error) {
	response := &types.BatchUpdateOrderStatusResponse{
		Errors: []types.OrderStatusValidationError{},
	}
	
	// 开启事务
	tx, err := g.DB().Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("开启事务失败: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	
	for _, orderID := range req.OrderIDs {
		err := r.UpdateStatusWithHistory(ctx, orderID, req.Status, req.Reason, req.OperatorType, operatorID, req.Metadata)
		if err != nil {
			// 记录失败的订单
			currentOrder, getErr := r.GetByID(ctx, orderID)
			currentStatus := types.OrderStatusIntPending
			if getErr == nil {
				currentStatus = r.orderStatusToInt(currentOrder.Status)
			}
			
			response.Errors = append(response.Errors, types.OrderStatusValidationError{
				OrderID:    orderID,
				FromStatus: currentStatus,
				ToStatus:   req.Status,
				Message:    err.Error(),
			})
			response.FailCount++
		} else {
			response.SuccessCount++
		}
	}
	
	return response, nil
}

// GetTimeoutOrders 获取超时的订单
func (r *OrderRepository) GetTimeoutOrders(ctx context.Context, timeoutConfig *types.OrderTimeoutConfig) ([]*types.Order, error) {
	tenantID := r.GetTenantID(ctx)
	
	// 计算超时时间点
	now := time.Now()
	paymentTimeout := now.Add(-time.Duration(timeoutConfig.PaymentTimeoutMinutes) * time.Minute)
	processingTimeout := now.Add(-time.Duration(timeoutConfig.ProcessingTimeoutHours) * time.Hour)
	
	var orderDataList []struct {
		types.Order
		ItemsJSON           string `db:"items"`
		PaymentInfoJSON     string `db:"payment_info"`
		VerificationInfoJSON string `db:"verification_info"`
	}
	
	// 查找超时的订单
	query := g.DB().Model("orders").Ctx(ctx).Where("tenant_id = ?", tenantID)
	
	if timeoutConfig.MerchantID != nil {
		query = query.Where("merchant_id = ?", *timeoutConfig.MerchantID)
	}
	
	// 查找待支付超时的订单或处理中超时的订单
	query = query.Where("(status = ? AND status_updated_at < ?) OR (status = ? AND status_updated_at < ?)",
		types.OrderStatusIntPending, paymentTimeout.Format("2006-01-02 15:04:05"),
		types.OrderStatusIntProcessing, processingTimeout.Format("2006-01-02 15:04:05"))
	
	err := query.Scan(&orderDataList)
	if err != nil {
		return nil, fmt.Errorf("查询超时订单失败: %v", err)
	}
	
	// 构造返回结果
	orders := make([]*types.Order, 0, len(orderDataList))
	for _, orderData := range orderDataList {
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
		
		orders = append(orders, &orderData.Order)
	}
	
	return orders, nil
}