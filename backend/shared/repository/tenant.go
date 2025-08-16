package repository

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/spume/mer-sys/shared/types"
)

// TenantRepository 租户仓储
type TenantRepository struct {
	*BaseRepository
}

// NewTenantRepository 创建租户仓储实例
func NewTenantRepository() *TenantRepository {
	return &TenantRepository{
		BaseRepository: NewBaseRepository("tenants"),
	}
}

// FindByCode 根据代码查找租户
func (r *TenantRepository) FindByCode(ctx context.Context, code string) (*types.Tenant, error) {
	// 租户表不需要租户隔离，因为它本身就是租户数据
	record, err := r.ModelWithoutTenant().Where("code", code).One()
	if err != nil {
		return nil, err
	}

	if record.IsEmpty() {
		return nil, fmt.Errorf("租户不存在: %s", code)
	}

	var tenant types.Tenant
	if err := record.Struct(&tenant); err != nil {
		return nil, err
	}

	return &tenant, nil
}

// FindByID 根据ID查找租户
func (r *TenantRepository) FindByID(ctx context.Context, id uint64) (*types.Tenant, error) {
	record, err := r.ModelWithoutTenant().Where("id", id).One()
	if err != nil {
		return nil, err
	}

	if record.IsEmpty() {
		return nil, fmt.Errorf("租户不存在: %d", id)
	}

	var tenant types.Tenant
	if err := record.Struct(&tenant); err != nil {
		return nil, err
	}

	return &tenant, nil
}

// Create 创建租户
func (r *TenantRepository) Create(ctx context.Context, tenant *types.Tenant) error {
	// 检查代码是否已存在
	exists, err := r.ModelWithoutTenant().Where("code", tenant.Code).Count()
	if err != nil {
		return err
	}
	if exists > 0 {
		return fmt.Errorf("租户代码已存在: %s", tenant.Code)
	}

	// 插入租户数据（不需要tenant_id）
	_, err = r.ModelWithoutTenant().Insert(tenant)
	return err
}

// UpdateByID 根据ID更新租户
func (r *TenantRepository) UpdateByID(ctx context.Context, id uint64, data interface{}) error {
	_, err := r.ModelWithoutTenant().Where("id", id).Update(data)
	return err
}

// DeleteByID 根据ID删除租户
func (r *TenantRepository) DeleteByID(ctx context.Context, id uint64) error {
	_, err := r.ModelWithoutTenant().Where("id", id).Delete()
	return err
}

// FindAll 查找所有租户
func (r *TenantRepository) FindAll(ctx context.Context) ([]*types.Tenant, error) {
	records, err := r.ModelWithoutTenant().All()
	if err != nil {
		return nil, err
	}

	var tenants []*types.Tenant
	for _, record := range records {
		var tenant types.Tenant
		if err := record.Struct(&tenant); err != nil {
			return nil, err
		}
		tenants = append(tenants, &tenant)
	}

	return tenants, nil
}

// FindPage 分页查找租户
func (r *TenantRepository) FindPage(ctx context.Context, page, pageSize int, condition interface{}, args ...interface{}) ([]*types.Tenant, int, error) {
	model := r.ModelWithoutTenant()
	if condition != nil {
		model = model.Where(condition, args...)
	}

	// 获取总数
	total, err := model.Clone().Count()
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	records, err := model.Page(page, pageSize).All()
	if err != nil {
		return nil, 0, err
	}

	var tenants []*types.Tenant
	for _, record := range records {
		var tenant types.Tenant
		if err := record.Struct(&tenant); err != nil {
			return nil, 0, err
		}
		tenants = append(tenants, &tenant)
	}

	return tenants, total, nil
}

// UpdateStatus 更新租户状态
func (r *TenantRepository) UpdateStatus(ctx context.Context, id uint64, status types.TenantStatus) error {
	_, err := r.ModelWithoutTenant().Where("id", id).Update(gdb.Map{
		"status": status,
	})
	return err
}
