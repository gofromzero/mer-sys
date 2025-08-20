package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gofromzero/mer-sys/backend/shared/audit"
	"github.com/gofromzero/mer-sys/backend/shared/auth"
	"github.com/gofromzero/mer-sys/backend/shared/notification"
	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// MerchantService 商户服务
type MerchantService struct {
	merchantRepo repository.MerchantRepository
}

// NewMerchantService 创建商户服务实例
func NewMerchantService() *MerchantService {
	return &MerchantService{
		merchantRepo: repository.NewMerchantRepository(),
	}
}

// RegisterMerchant 注册商户申请
func (s *MerchantService) RegisterMerchant(ctx context.Context, req *types.MerchantRegistrationRequest) (*types.Merchant, error) {
	// 验证商户代码是否已存在
	existing, err := s.merchantRepo.GetByCode(ctx, req.Code)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("检查商户代码失败: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("商户代码 %s 已存在", req.Code)
	}

	// 创建商户实体
	merchant := &types.Merchant{
		Name:         req.Name,
		Code:         req.Code,
		Status:       types.MerchantStatusPending,
		BusinessInfo: req.BusinessInfo,
		RightsBalance: &types.RightsBalance{
			TotalBalance:  0,
			UsedBalance:   0,
			FrozenBalance: 0,
		},
	}

	// 保存商户
	if err := s.merchantRepo.Create(ctx, merchant); err != nil {
		return nil, fmt.Errorf("创建商户失败: %w", err)
	}

	// 记录审计日志
	userInfo := auth.GetUserInfoFromContext(ctx)
	audit.LogOperation(ctx, "merchant", "register", g.Map{
		"merchant_id":   merchant.ID,
		"merchant_name": merchant.Name,
		"merchant_code": merchant.Code,
		"business_info": merchant.BusinessInfo,
		"operator_id":   userInfo.UserID,
	})

	return merchant, nil
}

// GetMerchantList 获取商户列表
func (s *MerchantService) GetMerchantList(ctx context.Context, query *types.MerchantListQuery) ([]*types.Merchant, int, error) {
	merchants, total, err := s.merchantRepo.FindPageWithFilter(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("查询商户列表失败: %w", err)
	}

	return merchants, total, nil
}

// GetMerchantByID 根据ID获取商户
func (s *MerchantService) GetMerchantByID(ctx context.Context, id uint64) (*types.Merchant, error) {
	merchant, err := s.merchantRepo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("商户不存在")
		}
		return nil, fmt.Errorf("获取商户信息失败: %w", err)
	}

	return merchant, nil
}

// UpdateMerchant 更新商户信息
func (s *MerchantService) UpdateMerchant(ctx context.Context, id uint64, req *types.MerchantUpdateRequest) (*types.Merchant, error) {
	// 获取现有商户信息
	merchant, err := s.merchantRepo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("商户不存在")
		}
		return nil, fmt.Errorf("获取商户信息失败: %w", err)
	}

	// 更新字段
	if req.Name != nil {
		merchant.Name = *req.Name
	}
	if req.BusinessInfo != nil {
		merchant.BusinessInfo = req.BusinessInfo
	}

	// 保存更新
	if err := s.merchantRepo.Update(ctx, merchant); err != nil {
		return nil, fmt.Errorf("更新商户信息失败: %w", err)
	}

	// 记录审计日志
	userInfo := auth.GetUserInfoFromContext(ctx)
	audit.LogOperation(ctx, "merchant", "update", g.Map{
		"merchant_id":   merchant.ID,
		"merchant_name": merchant.Name,
		"update_fields": req,
		"operator_id":   userInfo.UserID,
	})

	return merchant, nil
}

// UpdateMerchantStatus 更新商户状态
func (s *MerchantService) UpdateMerchantStatus(ctx context.Context, id uint64, status types.MerchantStatus, comment string) error {
	// 验证商户是否存在
	merchant, err := s.merchantRepo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("商户不存在")
		}
		return fmt.Errorf("获取商户信息失败: %w", err)
	}

	oldStatus := merchant.Status

	// 更新状态
	if err := s.merchantRepo.UpdateStatus(ctx, id, status); err != nil {
		return fmt.Errorf("更新商户状态失败: %w", err)
	}

	// 记录审计日志
	userInfo := auth.GetUserInfoFromContext(ctx)
	audit.LogOperation(ctx, "merchant", "status_change", g.Map{
		"merchant_id":   id,
		"merchant_name": merchant.Name,
		"old_status":    oldStatus,
		"new_status":    status,
		"comment":       comment,
		"operator_id":   userInfo.UserID,
	})

	// 发送状态变更通知
	err = notification.SendMerchantStatusChangedNotification(ctx, merchant, oldStatus, status, comment)
	if err != nil {
		g.Log().Errorf(ctx, "发送商户状态变更通知失败: %v", err)
		// 通知失败不影响业务流程，继续执行
	}

	return nil
}

// ApproveMerchant 审批商户申请
func (s *MerchantService) ApproveMerchant(ctx context.Context, id uint64, comment string) error {
	// 验证商户是否存在且状态为待审核
	merchant, err := s.merchantRepo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("商户不存在")
		}
		return fmt.Errorf("获取商户信息失败: %w", err)
	}

	if merchant.Status != types.MerchantStatusPending {
		return fmt.Errorf("商户状态为 %s，无法审批", merchant.Status)
	}

	// 获取审批人信息
	userInfo := auth.GetUserInfoFromContext(ctx)
	
	// 更新审批状态
	if err := s.merchantRepo.UpdateApproval(ctx, id, types.MerchantStatusActive, userInfo.UserID); err != nil {
		return fmt.Errorf("审批商户失败: %w", err)
	}

	// 记录审计日志
	audit.LogOperation(ctx, "merchant", "approve", g.Map{
		"merchant_id":   id,
		"merchant_name": merchant.Name,
		"comment":       comment,
		"operator_id":   userInfo.UserID,
		"approval_time": time.Now(),
	})

	// 发送审批通过通知
	err = notification.SendMerchantApprovedNotification(ctx, merchant, userInfo.Username)
	if err != nil {
		g.Log().Errorf(ctx, "发送商户审批通过通知失败: %v", err)
		// 通知失败不影响业务流程，继续执行
	}

	return nil
}

// RejectMerchant 拒绝商户申请
func (s *MerchantService) RejectMerchant(ctx context.Context, id uint64, comment string) error {
	// 验证商户是否存在且状态为待审核
	merchant, err := s.merchantRepo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("商户不存在")
		}
		return fmt.Errorf("获取商户信息失败: %w", err)
	}

	if merchant.Status != types.MerchantStatusPending {
		return fmt.Errorf("商户状态为 %s，无法拒绝", merchant.Status)
	}

	// 获取审批人信息
	userInfo := auth.GetUserInfoFromContext(ctx)
	
	// 更新审批状态为停用
	if err := s.merchantRepo.UpdateApproval(ctx, id, types.MerchantStatusDeactivated, userInfo.UserID); err != nil {
		return fmt.Errorf("拒绝商户失败: %w", err)
	}

	// 记录审计日志
	audit.LogOperation(ctx, "merchant", "reject", g.Map{
		"merchant_id":   id,
		"merchant_name": merchant.Name,
		"comment":       comment,
		"operator_id":   userInfo.UserID,
		"approval_time": time.Now(),
	})

	// 发送审批拒绝通知
	err = notification.SendMerchantRejectedNotification(ctx, merchant, comment)
	if err != nil {
		g.Log().Errorf(ctx, "发送商户审批拒绝通知失败: %v", err)
		// 通知失败不影响业务流程，继续执行
	}

	return nil
}

// GetMerchantAuditLog 获取商户操作历史
func (s *MerchantService) GetMerchantAuditLog(ctx context.Context, merchantID uint64) ([]g.Map, error) {
	// 验证商户是否存在
	_, err := s.merchantRepo.GetByID(ctx, merchantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("商户不存在")
		}
		return nil, fmt.Errorf("获取商户信息失败: %w", err)
	}

	// 获取审计日志
	logs, err := audit.GetOperationLogs(ctx, "merchant", fmt.Sprintf("%d", merchantID))
	if err != nil {
		return nil, fmt.Errorf("获取审计日志失败: %w", err)
	}

	return logs, nil
}