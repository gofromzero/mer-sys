package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gofromzero/mer-sys/backend/shared/audit"
	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

type ITenantService interface {
	CreateTenant(ctx context.Context, req *types.CreateTenantRequest) (*types.TenantResponse, error)
	ListTenants(ctx context.Context, req *types.ListTenantsRequest) (*types.ListTenantsResponse, error)
	GetTenantByID(ctx context.Context, id uint64) (*types.TenantResponse, error)
	UpdateTenant(ctx context.Context, id uint64, req *types.UpdateTenantRequest) (*types.TenantResponse, error)
	UpdateTenantStatus(ctx context.Context, id uint64, req *types.UpdateTenantStatusRequest) error
	GetTenantConfig(ctx context.Context, id uint64) (*types.TenantConfig, error)
	UpdateTenantConfig(ctx context.Context, id uint64, config *types.TenantConfig) error
	GetConfigChangeNotification(ctx context.Context, id uint64) (map[string]interface{}, error)
}

type tenantService struct {
	tenantRepo    repository.ITenantRepository
	configCache   *TenantConfigCache
}

func NewTenantService() ITenantService {
	return &tenantService{
		tenantRepo:  repository.NewTenantRepository(),
		configCache: NewTenantConfigCache(),
	}
}

func (s *tenantService) CreateTenant(ctx context.Context, req *types.CreateTenantRequest) (*types.TenantResponse, error) {
	// 检查租户代码是否已存在
	existing, err := s.tenantRepo.GetByCode(ctx, req.Code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("租户代码已存在")
	}

	// 检查联系邮箱是否已存在
	existingEmail, err := s.tenantRepo.GetByContactEmail(ctx, req.ContactEmail)
	if err != nil {
		return nil, err
	}
	if existingEmail != nil {
		return nil, errors.New("联系邮箱已被使用")
	}

	// 创建租户实体
	now := gtime.Now()
	tenant := &types.Tenant{
		Name:             req.Name,
		Code:             req.Code,
		Status:           types.TenantStatusActive,
		BusinessType:     req.BusinessType,
		ContactPerson:    req.ContactPerson,
		ContactEmail:     req.ContactEmail,
		ContactPhone:     req.ContactPhone,
		Address:          req.Address,
		RegistrationTime: &now.Time,
		ActivationTime:   &now.Time,
		Config:           `{"max_users": 100, "max_merchants": 50, "features": ["basic"], "settings": {}}`,
		CreatedAt:        now.Time,
		UpdatedAt:        now.Time,
	}

	// 保存到数据库
	id, err := s.tenantRepo.Create(ctx, tenant)
	if err != nil {
		return nil, err
	}

	// 记录租户创建审计日志
	audit.LogTenantAccess(ctx, id, "tenant", "create", map[string]interface{}{
		"tenant_name": tenant.Name,
		"tenant_code": tenant.Code,
		"business_type": tenant.BusinessType,
		"contact_email": tenant.ContactEmail,
	})

	// 返回创建的租户信息
	tenant.ID = id
	return s.convertToResponse(tenant), nil
}

func (s *tenantService) ListTenants(ctx context.Context, req *types.ListTenantsRequest) (*types.ListTenantsResponse, error) {
	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	tenants, total, err := s.tenantRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	responses := make([]types.TenantResponse, len(tenants))
	for i, tenant := range tenants {
		responses[i] = *s.convertToResponse(&tenant)
	}

	return &types.ListTenantsResponse{
		Total:   total,
		Page:    req.Page,
		Size:    len(responses),
		Tenants: responses,
	}, nil
}

func (s *tenantService) GetTenantByID(ctx context.Context, id uint64) (*types.TenantResponse, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if tenant == nil {
		return nil, nil
	}

	return s.convertToResponse(tenant), nil
}

func (s *tenantService) UpdateTenant(ctx context.Context, id uint64, req *types.UpdateTenantRequest) (*types.TenantResponse, error) {
	// 获取现有租户
	tenant, err := s.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if tenant == nil {
		return nil, errors.New("租户不存在")
	}

	// 更新字段
	if req.Name != "" {
		tenant.Name = req.Name
	}
	if req.BusinessType != "" {
		tenant.BusinessType = req.BusinessType
	}
	if req.ContactPerson != "" {
		tenant.ContactPerson = req.ContactPerson
	}
	if req.ContactEmail != "" {
		// 检查新邮箱是否已被其他租户使用
		existing, err := s.tenantRepo.GetByContactEmail(ctx, req.ContactEmail)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, errors.New("联系邮箱已被其他租户使用")
		}
		tenant.ContactEmail = req.ContactEmail
	}
	if req.ContactPhone != "" {
		tenant.ContactPhone = req.ContactPhone
	}
	if req.Address != "" {
		tenant.Address = req.Address
	}

	tenant.UpdatedAt = gtime.Now().Time

	// 更新数据库
	err = s.tenantRepo.Update(ctx, tenant)
	if err != nil {
		return nil, err
	}

	return s.convertToResponse(tenant), nil
}

func (s *tenantService) UpdateTenantStatus(ctx context.Context, id uint64, req *types.UpdateTenantStatusRequest) error {
	// 获取现有租户
	tenant, err := s.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if tenant == nil {
		return errors.New("租户不存在")
	}

	// 更新状态
	oldStatus := tenant.Status
	tenant.Status = req.Status
	tenant.UpdatedAt = gtime.Now().Time

	// 状态转换逻辑
	if oldStatus != req.Status {
		switch req.Status {
		case types.TenantStatusActive:
			if tenant.ActivationTime == nil {
				now := gtime.Now()
				tenant.ActivationTime = &now.Time
			}
		case types.TenantStatusSuspended, types.TenantStatusExpired:
			// 可以添加暂停或过期的业务逻辑
		}
	}

	// 更新数据库
	err = s.tenantRepo.Update(ctx, tenant)
	if err != nil {
		return err
	}

	// 记录状态变更审计日志
	audit.LogTenantAccess(ctx, id, "tenant", "status_change", map[string]interface{}{
		"old_status": string(oldStatus),
		"new_status": string(req.Status),
		"reason": req.Reason,
		"tenant_name": tenant.Name,
		"tenant_code": tenant.Code,
	})

	g.Log().Infof(ctx, "Tenant status changed: tenant_id=%d, old_status=%s, new_status=%s, reason=%s", 
		id, oldStatus, req.Status, req.Reason)

	return nil
}

func (s *tenantService) GetTenantConfig(ctx context.Context, id uint64) (*types.TenantConfig, error) {
	// 首先尝试从缓存获取
	config, err := s.configCache.GetConfig(ctx, id)
	if err == nil && config != nil {
		return config, nil
	}

	// 缓存未命中，从数据库获取
	tenant, err := s.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if tenant == nil {
		return nil, errors.New("租户不存在")
	}

	var tenantConfig types.TenantConfig
	if tenant.Config != "" {
		err = json.Unmarshal([]byte(tenant.Config), &tenantConfig)
		if err != nil {
			return nil, errors.New("租户配置格式错误")
		}
	}

	// 将配置写入缓存
	s.configCache.SetConfig(ctx, id, &tenantConfig)

	return &tenantConfig, nil
}

func (s *tenantService) UpdateTenantConfig(ctx context.Context, id uint64, config *types.TenantConfig) error {
	// 获取现有租户
	tenant, err := s.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if tenant == nil {
		return errors.New("租户不存在")
	}

	// 序列化配置
	configJSON, err := json.Marshal(config)
	if err != nil {
		return errors.New("配置序列化失败")
	}

	// 更新配置
	tenant.Config = string(configJSON)
	tenant.UpdatedAt = gtime.Now().Time

	// 更新数据库
	err = s.tenantRepo.Update(ctx, tenant)
	if err != nil {
		return err
	}

	// 使配置缓存失效
	s.configCache.InvalidateConfig(ctx, id)

	// 设置配置变更通知
	changeInfo := map[string]interface{}{
		"tenant_id":   id,
		"tenant_name": tenant.Name,
		"tenant_code": tenant.Code,
		"changed_at":  tenant.UpdatedAt.Format("2006-01-02 15:04:05"),
		"new_config":  config,
	}
	s.configCache.SetConfigChangeNotification(ctx, id, changeInfo)

	// 记录配置变更审计日志
	audit.LogTenantAccess(ctx, id, "tenant", "config_update", map[string]interface{}{
		"tenant_name": tenant.Name,
		"tenant_code": tenant.Code,
		"new_config":  config,
	})

	g.Log().Infof(ctx, "Tenant config updated: tenant_id=%d", id)

	return nil
}

func (s *tenantService) GetConfigChangeNotification(ctx context.Context, id uint64) (map[string]interface{}, error) {
	return s.configCache.GetConfigChangeNotification(ctx, id)
}

func (s *tenantService) convertToResponse(tenant *types.Tenant) *types.TenantResponse {
	return &types.TenantResponse{
		ID:               tenant.ID,
		Name:             tenant.Name,
		Code:             tenant.Code,
		Status:           string(tenant.Status),
		BusinessType:     tenant.BusinessType,
		ContactPerson:    tenant.ContactPerson,
		ContactEmail:     tenant.ContactEmail,
		ContactPhone:     tenant.ContactPhone,
		Address:          tenant.Address,
		RegistrationTime: tenant.RegistrationTime,
		ActivationTime:   tenant.ActivationTime,
		CreatedAt:        tenant.CreatedAt,
		UpdatedAt:        tenant.UpdatedAt,
	}
}