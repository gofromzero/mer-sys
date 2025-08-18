package repository

import (
	"context"

	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// ITenantRepository 租户仓储接口
type ITenantRepository interface {
	Create(ctx context.Context, tenant *types.Tenant) (uint64, error)
	GetByID(ctx context.Context, id uint64) (*types.Tenant, error)
	GetByCode(ctx context.Context, code string) (*types.Tenant, error)
	GetByContactEmail(ctx context.Context, email string) (*types.Tenant, error)
	Update(ctx context.Context, tenant *types.Tenant) error
	List(ctx context.Context, req *types.ListTenantsRequest) ([]types.Tenant, int, error)
}

// TenantRepository 租户仓储
type TenantRepository struct {
	*BaseRepository
}

// NewTenantRepository 创建租户仓储实例
func NewTenantRepository() ITenantRepository {
	return &TenantRepository{
		BaseRepository: NewBaseRepository("tenants"),
	}
}

// Create 创建租户
func (r *TenantRepository) Create(ctx context.Context, tenant *types.Tenant) (uint64, error) {
	// 插入租户数据（不需要tenant_id，因为这就是租户表）
	result, err := r.ModelWithoutTenant().Insert(tenant)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint64(id), nil
}

// GetByID 根据ID查找租户
func (r *TenantRepository) GetByID(ctx context.Context, id uint64) (*types.Tenant, error) {
	record, err := r.ModelWithoutTenant().Where("id", id).One()
	if err != nil {
		return nil, err
	}

	if record.IsEmpty() {
		return nil, nil
	}

	var tenant types.Tenant
	if err := record.Struct(&tenant); err != nil {
		return nil, err
	}

	return &tenant, nil
}

// GetByCode 根据代码查找租户
func (r *TenantRepository) GetByCode(ctx context.Context, code string) (*types.Tenant, error) {
	record, err := r.ModelWithoutTenant().Where("code", code).One()
	if err != nil {
		return nil, err
	}

	if record.IsEmpty() {
		return nil, nil
	}

	var tenant types.Tenant
	if err := record.Struct(&tenant); err != nil {
		return nil, err
	}

	return &tenant, nil
}

// GetByContactEmail 根据联系邮箱查找租户
func (r *TenantRepository) GetByContactEmail(ctx context.Context, email string) (*types.Tenant, error) {
	record, err := r.ModelWithoutTenant().Where("contact_email", email).One()
	if err != nil {
		return nil, err
	}

	if record.IsEmpty() {
		return nil, nil
	}

	var tenant types.Tenant
	if err := record.Struct(&tenant); err != nil {
		return nil, err
	}

	return &tenant, nil
}

// Update 更新租户
func (r *TenantRepository) Update(ctx context.Context, tenant *types.Tenant) error {
	_, err := r.ModelWithoutTenant().Where("id", tenant.ID).Update(tenant)
	return err
}

// List 分页查询租户列表
func (r *TenantRepository) List(ctx context.Context, req *types.ListTenantsRequest) ([]types.Tenant, int, error) {
	model := r.ModelWithoutTenant()

	// 构建查询条件
	if req.Status != "" {
		model = model.Where("status", req.Status)
	}
	if req.BusinessType != "" {
		model = model.Where("business_type", req.BusinessType)
	}
	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		model = model.Where(
			"name LIKE ? OR code LIKE ? OR contact_person LIKE ? OR contact_email LIKE ?", 
			searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	// 获取总数
	total, err := model.Clone().Count()
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (req.Page - 1) * req.PageSize
	records, err := model.Offset(offset).Limit(req.PageSize).OrderDesc("created_at").All()
	if err != nil {
		return nil, 0, err
	}

	var tenants []types.Tenant
	for _, record := range records {
		var tenant types.Tenant
		if err := record.Struct(&tenant); err != nil {
			return nil, 0, err
		}
		tenants = append(tenants, tenant)
	}

	return tenants, total, nil
}
