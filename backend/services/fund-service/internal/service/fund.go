package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"mer-demo/shared/audit"
	"mer-demo/shared/repository"
	"mer-demo/shared/types"
)

// FundService 资金管理服务接口
type FundService interface {
	// 充值相关
	Deposit(ctx context.Context, req *types.DepositRequest, operatorID uint64) (*types.Fund, error)
	BatchDeposit(ctx context.Context, req *types.BatchDepositRequest, operatorID uint64) ([]*types.Fund, error)
	
	// 权益分配
	Allocate(ctx context.Context, req *types.AllocateRequest, operatorID uint64) (*types.Fund, error)
	
	// 余额查询
	GetMerchantBalance(ctx context.Context, merchantID uint64) (*types.RightsBalance, error)
	
	// 流转记录查询
	ListTransactions(ctx context.Context, query *types.FundTransactionQuery) ([]*types.FundTransaction, int64, error)
	
	// 统计
	GetFundSummary(ctx context.Context, merchantID *uint64) (*types.FundSummary, error)
	
	// 冻结/解冻
	FreezeMerchantBalance(ctx context.Context, merchantID uint64, action string, amount float64, reason string, operatorID uint64) error
}

// fundService 资金管理服务实现
type fundService struct {
	fundRepo repository.FundRepository
}

// NewFundService 创建资金管理服务
func NewFundService() FundService {
	return &fundService{
		fundRepo: repository.NewFundRepository(),
	}
}

// Deposit 单笔资金充值
func (s *fundService) Deposit(ctx context.Context, req *types.DepositRequest, operatorID uint64) (*types.Fund, error) {
	// 获取租户ID
	tenantID := getTenantIDFromContext(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("无效的租户上下文")
	}

	var fund *types.Fund
	var err error

	// 在事务中执行充值操作
	err = s.fundRepo.WithTransaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 创建资金记录
		fund = &types.Fund{
			TenantID:   tenantID,
			MerchantID: req.MerchantID,
			FundType:   types.FundTypeDeposit,
			Amount:     req.Amount,
			Currency:   req.Currency,
			Status:     types.FundStatusPending,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		if err := s.fundRepo.CreateFund(ctx, fund); err != nil {
			return fmt.Errorf("创建资金记录失败: %v", err)
		}

		// 获取当前商户余额
		balance, err := s.fundRepo.GetMerchantBalance(ctx, tenantID, req.MerchantID)
		if err != nil {
			return fmt.Errorf("获取商户余额失败: %v", err)
		}

		// 创建资金流转记录
		transaction := &types.FundTransaction{
			TenantID:        tenantID,
			MerchantID:      req.MerchantID,
			FundID:          fund.ID,
			TransactionType: types.TransactionTypeCredit,
			Amount:          req.Amount,
			BalanceBefore:   balance.TotalBalance,
			BalanceAfter:    balance.TotalBalance + req.Amount,
			OperatorID:      operatorID,
			Description:     req.Description,
			CreatedAt:       time.Now(),
		}

		if err := s.fundRepo.CreateFundTransaction(ctx, transaction); err != nil {
			return fmt.Errorf("创建流转记录失败: %v", err)
		}

		// 更新商户余额
		balance.TotalBalance += req.Amount
		balance.UpdateAvailableBalance()

		if err := s.fundRepo.UpdateMerchantBalance(ctx, tenantID, req.MerchantID, balance); err != nil {
			return fmt.Errorf("更新商户余额失败: %v", err)
		}

		// 更新资金记录状态为已确认
		if err := s.fundRepo.UpdateFundStatus(ctx, tenantID, fund.ID, types.FundStatusConfirmed); err != nil {
			return fmt.Errorf("更新资金状态失败: %v", err)
		}

		fund.Status = types.FundStatusConfirmed
		
		// 记录审计日志
		audit.LogFundDeposit(ctx, tenantID, req.MerchantID, operatorID, req.Amount, req.Currency, fund.ID, req.Description)
		
		return nil
	})

	if err != nil {
		return nil, err
	}

	return fund, nil
}

// BatchDeposit 批量资金充值
func (s *fundService) BatchDeposit(ctx context.Context, req *types.BatchDepositRequest, operatorID uint64) ([]*types.Fund, error) {
	var results []*types.Fund
	totalAmount := 0.0
	var fundIDs []uint64

	// 逐个处理充值请求
	for i, deposit := range req.Deposits {
		fund, err := s.Deposit(ctx, &deposit, operatorID)
		if err != nil {
			return nil, fmt.Errorf("第%d笔充值失败: %v", i+1, err)
		}
		results = append(results, fund)
		totalAmount += deposit.Amount
		fundIDs = append(fundIDs, fund.ID)
	}

	// 记录批量充值审计日志
	tenantID := getTenantIDFromContext(ctx)
	audit.LogFundBatchDeposit(ctx, tenantID, operatorID, len(req.Deposits), totalAmount, fundIDs, req)

	return results, nil
}

// Allocate 权益分配
func (s *fundService) Allocate(ctx context.Context, req *types.AllocateRequest, operatorID uint64) (*types.Fund, error) {
	// 获取租户ID
	tenantID := getTenantIDFromContext(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("无效的租户上下文")
	}

	var fund *types.Fund
	var err error

	// 在事务中执行分配操作
	err = s.fundRepo.WithTransaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 创建资金记录
		fund = &types.Fund{
			TenantID:   tenantID,
			MerchantID: req.MerchantID,
			FundType:   types.FundTypeAllocation,
			Amount:     req.Amount,
			Currency:   "CNY", // 默认人民币
			Status:     types.FundStatusPending,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		if err := s.fundRepo.CreateFund(ctx, fund); err != nil {
			return fmt.Errorf("创建资金记录失败: %v", err)
		}

		// 获取当前商户余额
		balance, err := s.fundRepo.GetMerchantBalance(ctx, tenantID, req.MerchantID)
		if err != nil {
			return fmt.Errorf("获取商户余额失败: %v", err)
		}

		// 创建资金流转记录
		transaction := &types.FundTransaction{
			TenantID:        tenantID,
			MerchantID:      req.MerchantID,
			FundID:          fund.ID,
			TransactionType: types.TransactionTypeCredit,
			Amount:          req.Amount,
			BalanceBefore:   balance.TotalBalance,
			BalanceAfter:    balance.TotalBalance + req.Amount,
			OperatorID:      operatorID,
			Description:     req.Description,
			CreatedAt:       time.Now(),
		}

		if err := s.fundRepo.CreateFundTransaction(ctx, transaction); err != nil {
			return fmt.Errorf("创建流转记录失败: %v", err)
		}

		// 更新商户余额
		balance.TotalBalance += req.Amount
		balance.UpdateAvailableBalance()

		if err := s.fundRepo.UpdateMerchantBalance(ctx, tenantID, req.MerchantID, balance); err != nil {
			return fmt.Errorf("更新商户余额失败: %v", err)
		}

		// 更新资金记录状态为已确认
		if err := s.fundRepo.UpdateFundStatus(ctx, tenantID, fund.ID, types.FundStatusConfirmed); err != nil {
			return fmt.Errorf("更新资金状态失败: %v", err)
		}

		fund.Status = types.FundStatusConfirmed
		
		// 记录审计日志
		audit.LogFundAllocate(ctx, tenantID, req.MerchantID, operatorID, req.Amount, fund.ID, req.Description)
		
		return nil
	})

	if err != nil {
		return nil, err
	}

	return fund, nil
}

// GetMerchantBalance 获取商户权益余额
func (s *fundService) GetMerchantBalance(ctx context.Context, merchantID uint64) (*types.RightsBalance, error) {
	tenantID := getTenantIDFromContext(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("无效的租户上下文")
	}

	balance, err := s.fundRepo.GetMerchantBalance(ctx, tenantID, merchantID)
	if err != nil {
		return nil, err
	}
	
	// 记录余额查询审计日志
	operatorID := getOperatorIDFromContext(ctx)
	balanceMap := map[string]interface{}{
		"total_balance":     balance.TotalBalance,
		"used_balance":      balance.UsedBalance,
		"frozen_balance":    balance.FrozenBalance,
		"available_balance": balance.AvailableBalance,
	}
	audit.LogFundBalanceQuery(ctx, tenantID, merchantID, operatorID, &balanceMap)
	
	return balance, nil
}

// ListTransactions 查询资金流转记录
func (s *fundService) ListTransactions(ctx context.Context, query *types.FundTransactionQuery) ([]*types.FundTransaction, int64, error) {
	transactions, total, err := s.fundRepo.ListFundTransactions(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	
	// 记录交易查询审计日志
	tenantID := getTenantIDFromContext(ctx)
	operatorID := getOperatorIDFromContext(ctx)
	queryMap := map[string]interface{}{
		"merchant_id":      query.MerchantID,
		"fund_id":          query.FundID,
		"transaction_type": query.TransactionType,
		"operator_id":      query.OperatorID,
		"start_time":       query.StartTime,
		"end_time":         query.EndTime,
		"page":             query.Page,
		"page_size":        query.PageSize,
	}
	audit.LogFundTransactionQuery(ctx, tenantID, operatorID, queryMap, int(total))
	
	return transactions, total, nil
}

// GetFundSummary 获取资金概览统计
func (s *fundService) GetFundSummary(ctx context.Context, merchantID *uint64) (*types.FundSummary, error) {
	tenantID := getTenantIDFromContext(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("无效的租户上下文")
	}

	return s.fundRepo.GetFundSummary(ctx, tenantID, merchantID)
}

// FreezeMerchantBalance 冻结/解冻商户权益
func (s *fundService) FreezeMerchantBalance(ctx context.Context, merchantID uint64, action string, amount float64, reason string, operatorID uint64) error {
	tenantID := getTenantIDFromContext(ctx)
	if tenantID == 0 {
		return fmt.Errorf("无效的租户上下文")
	}

	err := s.fundRepo.WithTransaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 获取当前商户余额
		balance, err := s.fundRepo.GetMerchantBalance(ctx, tenantID, merchantID)
		if err != nil {
			return fmt.Errorf("获取商户余额失败: %v", err)
		}

		var balanceBefore, balanceAfter float64
		balanceBefore = balance.FrozenBalance

		if action == "freeze" {
			// 冻结权益
			if balance.GetAvailableBalance() < amount {
				return fmt.Errorf("可用余额不足，无法冻结")
			}
			balance.FrozenBalance += amount
			balanceAfter = balance.FrozenBalance
		} else {
			// 解冻权益
			if balance.FrozenBalance < amount {
				return fmt.Errorf("冻结余额不足，无法解冻")
			}
			balance.FrozenBalance -= amount
			balanceAfter = balance.FrozenBalance
		}

		// 更新可用余额
		balance.UpdateAvailableBalance()

		// 更新商户余额
		if err := s.fundRepo.UpdateMerchantBalance(ctx, tenantID, merchantID, balance); err != nil {
			return fmt.Errorf("更新商户余额失败: %v", err)
		}

		// 创建冻结/解冻操作的资金记录
		fund := &types.Fund{
			TenantID:   tenantID,
			MerchantID: merchantID,
			FundType:   types.FundTypeAllocation, // 使用分配类型记录冻结操作
			Amount:     amount,
			Currency:   "CNY",
			Status:     types.FundStatusConfirmed,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		if err := s.fundRepo.CreateFund(ctx, fund); err != nil {
			return fmt.Errorf("创建资金记录失败: %v", err)
		}

		// 创建资金流转记录
		transactionType := types.TransactionTypeDebit
		if action == "unfreeze" {
			transactionType = types.TransactionTypeCredit
		}

		transaction := &types.FundTransaction{
			TenantID:        tenantID,
			MerchantID:      merchantID,
			FundID:          fund.ID,
			TransactionType: transactionType,
			Amount:          amount,
			BalanceBefore:   balanceBefore,
			BalanceAfter:    balanceAfter,
			OperatorID:      operatorID,
			Description:     fmt.Sprintf("%s权益: %s", action, reason),
			CreatedAt:       time.Now(),
		}

		if err := s.fundRepo.CreateFundTransaction(ctx, transaction); err != nil {
			return fmt.Errorf("创建流转记录失败: %v", err)
		}

		return nil
	})
	
	if err != nil {
		return err
	}
	
	// 记录审计日志
	if action == "freeze" {
		audit.LogFundFreeze(ctx, tenantID, merchantID, operatorID, amount, reason, nil)
	} else {
		audit.LogFundUnfreeze(ctx, tenantID, merchantID, operatorID, amount, reason, nil)
	}
	
	return nil
}

// getTenantIDFromContext 从上下文获取租户ID
func getTenantIDFromContext(ctx context.Context) uint64 {
	if tenantID := ctx.Value("tenant_id"); tenantID != nil {
		if id, ok := tenantID.(uint64); ok {
			return id
		}
	}
	return 0
}

// getOperatorIDFromContext 从上下文获取操作人ID
func getOperatorIDFromContext(ctx context.Context) uint64 {
	if operatorID := ctx.Value("user_id"); operatorID != nil {
		if id, ok := operatorID.(uint64); ok {
			return id
		}
	}
	return 0
}