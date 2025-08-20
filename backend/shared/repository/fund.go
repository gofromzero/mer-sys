package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"mer-demo/shared/types"
)

// FundRepository 资金仓储接口
type FundRepository interface {
	// Fund相关操作
	CreateFund(ctx context.Context, fund *types.Fund) error
	GetFundByID(ctx context.Context, tenantID, fundID uint64) (*types.Fund, error)
	UpdateFundStatus(ctx context.Context, tenantID, fundID uint64, status types.FundStatus) error
	ListFunds(ctx context.Context, tenantID, merchantID uint64, page, pageSize int) ([]*types.Fund, int64, error)
	
	// FundTransaction相关操作
	CreateFundTransaction(ctx context.Context, transaction *types.FundTransaction) error
	ListFundTransactions(ctx context.Context, query *types.FundTransactionQuery) ([]*types.FundTransaction, int64, error)
	
	// 权益余额相关操作
	GetMerchantBalance(ctx context.Context, tenantID, merchantID uint64) (*types.RightsBalance, error)
	UpdateMerchantBalance(ctx context.Context, tenantID, merchantID uint64, balance *types.RightsBalance) error
	
	// 统计相关操作
	GetFundSummary(ctx context.Context, tenantID uint64, merchantID *uint64) (*types.FundSummary, error)
	
	// 事务操作
	WithTransaction(ctx context.Context, fn func(ctx context.Context, tx gdb.TX) error) error
}

// fundRepository 资金仓储实现
type fundRepository struct {
	db gdb.DB
}

// NewFundRepository 创建资金仓储实例
func NewFundRepository() FundRepository {
	return &fundRepository{
		db: g.DB(),
	}
}

// CreateFund 创建资金记录
func (r *fundRepository) CreateFund(ctx context.Context, fund *types.Fund) error {
	// 验证数据
	if err := fund.Validate(); err != nil {
		return fmt.Errorf("资金记录验证失败: %v", err)
	}
	
	// 确保租户隔离
	tenantID := GetTenantIDFromContext(ctx)
	if tenantID == 0 {
		return fmt.Errorf("无效的租户上下文")
	}
	fund.TenantID = tenantID
	
	result, err := r.db.Model("funds").Ctx(ctx).Insert(fund)
	if err != nil {
		return fmt.Errorf("创建资金记录失败: %v", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取资金记录ID失败: %v", err)
	}
	fund.ID = uint64(id)
	
	return nil
}

// GetFundByID 根据ID获取资金记录
func (r *fundRepository) GetFundByID(ctx context.Context, tenantID, fundID uint64) (*types.Fund, error) {
	if tenantID == 0 || fundID == 0 {
		return nil, fmt.Errorf("无效的参数")
	}
	
	var fund types.Fund
	err := r.db.Model("funds").Ctx(ctx).
		Where("id = ? AND tenant_id = ?", fundID, tenantID).
		Scan(&fund)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询资金记录失败: %v", err)
	}
	
	return &fund, nil
}

// UpdateFundStatus 更新资金状态
func (r *fundRepository) UpdateFundStatus(ctx context.Context, tenantID, fundID uint64, status types.FundStatus) error {
	if tenantID == 0 || fundID == 0 {
		return fmt.Errorf("无效的参数")
	}
	
	result, err := r.db.Model("funds").Ctx(ctx).
		Where("id = ? AND tenant_id = ?", fundID, tenantID).
		Update(g.Map{"status": status})
	
	if err != nil {
		return fmt.Errorf("更新资金状态失败: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("资金记录不存在或无权限")
	}
	
	return nil
}

// ListFunds 获取资金记录列表
func (r *fundRepository) ListFunds(ctx context.Context, tenantID, merchantID uint64, page, pageSize int) ([]*types.Fund, int64, error) {
	if tenantID == 0 {
		return nil, 0, fmt.Errorf("无效的租户ID")
	}
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	
	// 构建查询条件
	whereCondition := "tenant_id = ?"
	whereArgs := []interface{}{tenantID}
	
	if merchantID > 0 {
		whereCondition += " AND merchant_id = ?"
		whereArgs = append(whereArgs, merchantID)
	}
	
	// 查询总数
	count, err := r.db.Model("funds").Ctx(ctx).
		Where(whereCondition, whereArgs...).
		Count()
	if err != nil {
		return nil, 0, fmt.Errorf("查询资金记录总数失败: %v", err)
	}
	
	// 查询列表
	var funds []*types.Fund
	err = r.db.Model("funds").Ctx(ctx).
		Where(whereCondition, whereArgs...).
		OrderDesc("created_at").
		Limit((page-1)*pageSize, pageSize).
		Scan(&funds)
	
	if err != nil {
		return nil, 0, fmt.Errorf("查询资金记录列表失败: %v", err)
	}
	
	return funds, int64(count), nil
}

// CreateFundTransaction 创建资金流转记录
func (r *fundRepository) CreateFundTransaction(ctx context.Context, transaction *types.FundTransaction) error {
	// 验证数据
	if err := transaction.Validate(); err != nil {
		return fmt.Errorf("资金流转记录验证失败: %v", err)
	}
	
	// 确保租户隔离
	tenantID := GetTenantIDFromContext(ctx)
	if tenantID == 0 {
		return fmt.Errorf("无效的租户上下文")
	}
	transaction.TenantID = tenantID
	
	result, err := r.db.Model("fund_transactions").Ctx(ctx).Insert(transaction)
	if err != nil {
		return fmt.Errorf("创建资金流转记录失败: %v", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取资金流转记录ID失败: %v", err)
	}
	transaction.ID = uint64(id)
	
	return nil
}

// ListFundTransactions 查询资金流转记录
func (r *fundRepository) ListFundTransactions(ctx context.Context, query *types.FundTransactionQuery) ([]*types.FundTransaction, int64, error) {
	// 确保租户隔离
	tenantID := GetTenantIDFromContext(ctx)
	if tenantID == 0 {
		return nil, 0, fmt.Errorf("无效的租户上下文")
	}
	query.TenantID = tenantID
	
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}
	
	// 构建查询条件
	model := r.db.Model("fund_transactions").Ctx(ctx).Where("tenant_id = ?", tenantID)
	
	if query.MerchantID > 0 {
		model = model.Where("merchant_id = ?", query.MerchantID)
	}
	if query.FundID > 0 {
		model = model.Where("fund_id = ?", query.FundID)
	}
	if query.TransactionType > 0 {
		model = model.Where("transaction_type = ?", query.TransactionType)
	}
	if query.OperatorID > 0 {
		model = model.Where("operator_id = ?", query.OperatorID)
	}
	if query.StartTime != nil {
		model = model.Where("created_at >= ?", query.StartTime)
	}
	if query.EndTime != nil {
		model = model.Where("created_at <= ?", query.EndTime)
	}
	
	// 查询总数
	count, err := model.Count()
	if err != nil {
		return nil, 0, fmt.Errorf("查询资金流转记录总数失败: %v", err)
	}
	
	// 查询列表
	var transactions []*types.FundTransaction
	err = model.OrderDesc("created_at").
		Limit((query.Page-1)*query.PageSize, query.PageSize).
		Scan(&transactions)
	
	if err != nil {
		return nil, 0, fmt.Errorf("查询资金流转记录列表失败: %v", err)
	}
	
	return transactions, int64(count), nil
}

// GetMerchantBalance 获取商户权益余额
func (r *fundRepository) GetMerchantBalance(ctx context.Context, tenantID, merchantID uint64) (*types.RightsBalance, error) {
	if tenantID == 0 || merchantID == 0 {
		return nil, fmt.Errorf("无效的参数")
	}
	
	var balance types.RightsBalance
	err := r.db.Model("merchants").Ctx(ctx).
		Fields("rights_balance").
		Where("id = ? AND tenant_id = ?", merchantID, tenantID).
		Scan(&balance)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("商户不存在")
		}
		return nil, fmt.Errorf("查询商户权益余额失败: %v", err)
	}
	
	return &balance, nil
}

// UpdateMerchantBalance 更新商户权益余额
func (r *fundRepository) UpdateMerchantBalance(ctx context.Context, tenantID, merchantID uint64, balance *types.RightsBalance) error {
	if tenantID == 0 || merchantID == 0 {
		return fmt.Errorf("无效的参数")
	}
	
	// 更新可用余额
	balance.UpdateAvailableBalance()
	
	result, err := r.db.Model("merchants").Ctx(ctx).
		Where("id = ? AND tenant_id = ?", merchantID, tenantID).
		Update(g.Map{"rights_balance": balance})
	
	if err != nil {
		return fmt.Errorf("更新商户权益余额失败: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("商户不存在或无权限")
	}
	
	return nil
}

// GetFundSummary 获取资金概览统计
func (r *fundRepository) GetFundSummary(ctx context.Context, tenantID uint64, merchantID *uint64) (*types.FundSummary, error) {
	if tenantID == 0 {
		return nil, fmt.Errorf("无效的租户ID")
	}
	
	// 构建查询条件
	whereCondition := "tenant_id = ? AND status = ?"
	whereArgs := []interface{}{tenantID, types.FundStatusConfirmed}
	
	if merchantID != nil && *merchantID > 0 {
		whereCondition += " AND merchant_id = ?"
		whereArgs = append(whereArgs, *merchantID)
	}
	
	// 查询各类型资金统计
	var summary types.FundSummary
	
	// 充值总额
	var totalDeposits sql.NullFloat64
	err := r.db.Model("funds").Ctx(ctx).
		Fields("COALESCE(SUM(amount), 0) as total").
		Where(whereCondition+" AND fund_type = ?", append(whereArgs, types.FundTypeDeposit)...).
		Scan(&totalDeposits)
	if err != nil {
		return nil, fmt.Errorf("查询充值总额失败: %v", err)
	}
	summary.TotalDeposits = totalDeposits.Float64
	
	// 分配总额
	var totalAllocations sql.NullFloat64
	err = r.db.Model("funds").Ctx(ctx).
		Fields("COALESCE(SUM(amount), 0) as total").
		Where(whereCondition+" AND fund_type = ?", append(whereArgs, types.FundTypeAllocation)...).
		Scan(&totalAllocations)
	if err != nil {
		return nil, fmt.Errorf("查询分配总额失败: %v", err)
	}
	summary.TotalAllocations = totalAllocations.Float64
	
	// 消费总额
	var totalConsumption sql.NullFloat64
	err = r.db.Model("funds").Ctx(ctx).
		Fields("COALESCE(SUM(amount), 0) as total").
		Where(whereCondition+" AND fund_type = ?", append(whereArgs, types.FundTypeConsumption)...).
		Scan(&totalConsumption)
	if err != nil {
		return nil, fmt.Errorf("查询消费总额失败: %v", err)
	}
	summary.TotalConsumption = totalConsumption.Float64
	
	// 退款总额
	var totalRefunds sql.NullFloat64
	err = r.db.Model("funds").Ctx(ctx).
		Fields("COALESCE(SUM(amount), 0) as total").
		Where(whereCondition+" AND fund_type = ?", append(whereArgs, types.FundTypeRefund)...).
		Scan(&totalRefunds)
	if err != nil {
		return nil, fmt.Errorf("查询退款总额失败: %v", err)
	}
	summary.TotalRefunds = totalRefunds.Float64
	
	// 计算可用余额
	summary.AvailableBalance = summary.TotalDeposits + summary.TotalAllocations - summary.TotalConsumption - summary.TotalRefunds
	
	return &summary, nil
}

// WithTransaction 在事务中执行操作
func (r *fundRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context, tx gdb.TX) error) error {
	return r.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		return fn(ctx, tx)
	})
}

// GetTenantIDFromContext 从上下文获取租户ID
func GetTenantIDFromContext(ctx context.Context) uint64 {
	if tenantID := ctx.Value("tenant_id"); tenantID != nil {
		if id, ok := tenantID.(uint64); ok {
			return id
		}
	}
	return 0
}