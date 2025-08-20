package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gofromzero/mer-sys/backend/shared/audit"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// MerchantRepository 商户仓储接口
type MerchantRepository interface {
	Create(ctx context.Context, merchant *types.Merchant) error
	GetByID(ctx context.Context, id uint64) (*types.Merchant, error)
	GetByCode(ctx context.Context, code string) (*types.Merchant, error)
	GetByTenantID(ctx context.Context, tenantID uint64) ([]*types.Merchant, error)
	Update(ctx context.Context, merchant *types.Merchant) error
	UpdateStatus(ctx context.Context, id uint64, status types.MerchantStatus) error
	UpdateApproval(ctx context.Context, id uint64, status types.MerchantStatus, approvedBy uint64) error
	Delete(ctx context.Context, id uint64) error
	FindPage(ctx context.Context, page, pageSize int, condition interface{}, args ...interface{}) ([]*types.Merchant, int, error)
	FindPageWithFilter(ctx context.Context, query *types.MerchantListQuery) ([]*types.Merchant, int, error)
	Count(ctx context.Context) (int, error)
	CountByStatus(ctx context.Context, status types.MerchantStatus) (int, error)
}

// merchantRepository 商户仓储实现
type merchantRepository struct {
	*BaseRepository
}

// NewMerchantRepository 创建商户仓储实例
func NewMerchantRepository() MerchantRepository {
	return &merchantRepository{
		BaseRepository: NewBaseRepository("merchants"),
	}
}

// Create 创建商户
func (r *merchantRepository) Create(ctx context.Context, merchant *types.Merchant) error {
	// 确保商户属于当前租户
	tenantID, err := r.GetTenantID(ctx)
	if err != nil {
		return err
	}
	
	merchant.TenantID = tenantID
	
	// 设置注册时间为当前时间
	now := time.Now()
	merchant.RegistrationTime = &now
	
	// 新注册的商户默认为待审核状态
	if merchant.Status == "" {
		merchant.Status = types.MerchantStatusPending
	}
	
	_, err = r.Insert(ctx, merchant)
	return err
}

// GetByID 根据ID获取商户
func (r *merchantRepository) GetByID(ctx context.Context, id uint64) (*types.Merchant, error) {
	record, err := r.FindOne(ctx, "id", id)
	if err != nil {
		return nil, err
	}
	
	if record.IsEmpty() {
		return nil, sql.ErrNoRows
	}
	
	var merchant types.Merchant
	if err := record.Struct(&merchant); err != nil {
		return nil, err
	}
	
	return &merchant, nil
}

// GetByCode 根据代码获取商户
func (r *merchantRepository) GetByCode(ctx context.Context, code string) (*types.Merchant, error) {
	record, err := r.FindOne(ctx, "code", code)
	if err != nil {
		return nil, err
	}
	
	if record.IsEmpty() {
		return nil, sql.ErrNoRows
	}
	
	var merchant types.Merchant
	if err := record.Struct(&merchant); err != nil {
		return nil, err
	}
	
	return &merchant, nil
}

// GetByTenantID 根据租户ID获取所有商户
func (r *merchantRepository) GetByTenantID(ctx context.Context, tenantID uint64) ([]*types.Merchant, error) {
	// 验证请求的租户ID是否与上下文中的租户ID一致
	contextTenantID, err := r.GetTenantID(ctx)
	if err != nil {
		return nil, err
	}
	
	if contextTenantID != tenantID {
		// 记录跨租户访问尝试
		audit.LogCrossTenantAttempt(ctx, contextTenantID, tenantID, "merchants", "query", 
			map[string]interface{}{
				"method": "GetByTenantID",
				"requested_tenant": tenantID,
				"user_tenant": contextTenantID,
			})
		return nil, types.ErrCrossTenantAccess
	}
	
	records, err := r.FindAll(ctx, nil)
	if err != nil {
		return nil, err
	}
	
	var merchants []*types.Merchant
	if err := records.Structs(&merchants); err != nil {
		return nil, err
	}
	
	return merchants, nil
}

// Update 更新商户信息
func (r *merchantRepository) Update(ctx context.Context, merchant *types.Merchant) error {
	// 确保只能更新当前租户的商户
	_, err := r.BaseRepository.Update(ctx, merchant, "id", merchant.ID)
	return err
}

// UpdateStatus 更新商户状态
func (r *merchantRepository) UpdateStatus(ctx context.Context, id uint64, status types.MerchantStatus) error {
	_, err := r.BaseRepository.Update(ctx, gdb.Map{"status": status}, "id", id)
	return err
}

// Delete 删除商户
func (r *merchantRepository) Delete(ctx context.Context, id uint64) error {
	_, err := r.BaseRepository.Delete(ctx, "id", id)
	return err
}

// FindPage 分页查询商户
func (r *merchantRepository) FindPage(ctx context.Context, page, pageSize int, condition interface{}, args ...interface{}) ([]*types.Merchant, int, error) {
	records, total, err := r.BaseRepository.FindPage(ctx, page, pageSize, condition, args...)
	if err != nil {
		return nil, 0, err
	}
	
	var merchants []*types.Merchant
	if err := records.Structs(&merchants); err != nil {
		return nil, 0, err
	}
	
	return merchants, total, nil
}

// Count 统计商户数量
func (r *merchantRepository) Count(ctx context.Context) (int, error) {
	return r.BaseRepository.Count(ctx, nil)
}

// UpdateApproval 更新商户审批状态
func (r *merchantRepository) UpdateApproval(ctx context.Context, id uint64, status types.MerchantStatus, approvedBy uint64) error {
	updateData := gdb.Map{
		"status":       status,
		"approved_by":  approvedBy,
		"approval_time": "NOW()",
	}
	_, err := r.BaseRepository.Update(ctx, updateData, "id", id)
	return err
}

// CountByStatus 根据状态统计商户数量
func (r *merchantRepository) CountByStatus(ctx context.Context, status types.MerchantStatus) (int, error) {
	return r.BaseRepository.Count(ctx, "status = ?", status)
}

// FindPageWithFilter 根据查询条件分页查询商户
func (r *merchantRepository) FindPageWithFilter(ctx context.Context, query *types.MerchantListQuery) ([]*types.Merchant, int, error) {
	// 构建查询条件
	conditions := make([]string, 0)
	args := make([]interface{}, 0)
	
	if query.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, query.Status)
	}
	
	if query.Name != "" {
		conditions = append(conditions, "name LIKE ?")
		args = append(args, "%"+query.Name+"%")
	}
	
	if query.Search != "" {
		conditions = append(conditions, "(name LIKE ? OR code LIKE ? OR JSON_EXTRACT(business_info, '$.contact_name') LIKE ?)")
		searchPattern := "%" + query.Search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}
	
	// 组合条件
	var whereClause interface{}
	if len(conditions) > 0 {
		whereClause = conditions[0]
		for i := 1; i < len(conditions); i++ {
			whereClause = whereClause.(string) + " AND " + conditions[i]
		}
	}
	
	return r.FindPage(ctx, query.Page, query.PageSize, whereClause, args...)
}