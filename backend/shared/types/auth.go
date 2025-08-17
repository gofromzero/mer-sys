package types

// RoleType 角色类型枚举
type RoleType string

const (
	// RoleTenantAdmin 租户管理员
	RoleTenantAdmin RoleType = "tenant_admin"
	// RoleMerchant 商户
	RoleMerchant RoleType = "merchant"
	// RoleCustomer 客户
	RoleCustomer RoleType = "customer"
)

// Permission 权限枚举
type Permission string

const (
	// 用户管理权限
	PermissionUserManage Permission = "user:manage"
	PermissionUserView   Permission = "user:view"
	PermissionUserCreate Permission = "user:create"
	PermissionUserUpdate Permission = "user:update"
	PermissionUserDelete Permission = "user:delete"

	// 商户管理权限
	PermissionMerchantManage Permission = "merchant:manage"
	PermissionMerchantView   Permission = "merchant:view"
	PermissionMerchantCreate Permission = "merchant:create"
	PermissionMerchantUpdate Permission = "merchant:update"
	PermissionMerchantDelete Permission = "merchant:delete"

	// 订单管理权限
	PermissionOrderManage Permission = "order:manage"
	PermissionOrderView   Permission = "order:view"
	PermissionOrderCreate Permission = "order:create"
	PermissionOrderUpdate Permission = "order:update"
	PermissionOrderDelete Permission = "order:delete"

	// 商品管理权限
	PermissionProductManage Permission = "product:manage"
	PermissionProductView   Permission = "product:view"
	PermissionProductCreate Permission = "product:create"
	PermissionProductUpdate Permission = "product:update"
	PermissionProductDelete Permission = "product:delete"

	// 租户管理权限
	PermissionTenantManage Permission = "tenant:manage"
	PermissionTenantView   Permission = "tenant:view"
	PermissionTenantCreate Permission = "tenant:create"
	PermissionTenantUpdate Permission = "tenant:update"
	PermissionTenantDelete Permission = "tenant:delete"

	// 报表权限
	PermissionReportView   Permission = "report:view"
	PermissionReportExport Permission = "report:export"
	PermissionReportCreate Permission = "report:create"
	PermissionReportDelete Permission = "report:delete"

	// 资金管理权限
	PermissionFundView     Permission = "fund:view"
	PermissionFundManage   Permission = "fund:manage"
	PermissionFundWithdraw Permission = "fund:withdraw"
	PermissionFundTransfer Permission = "fund:transfer"

	// 权益管理权限
	PermissionBenefitView   Permission = "benefit:view"
	PermissionBenefitManage Permission = "benefit:manage"
	PermissionBenefitCreate Permission = "benefit:create"
	PermissionBenefitUpdate Permission = "benefit:update"
	PermissionBenefitDelete Permission = "benefit:delete"

	// 系统管理权限
	PermissionSystemConfig Permission = "system:config"
	PermissionSystemAudit  Permission = "system:audit"
	PermissionSystemLog    Permission = "system:log"

	// 角色管理权限
	PermissionRoleView   Permission = "role:view"
	PermissionRoleManage Permission = "role:manage"
	PermissionRoleAssign Permission = "role:assign"
)

// Role 角色定义
type Role struct {
	Type        RoleType     `json:"type"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
}

// GetDefaultRoles 获取默认角色配置
func GetDefaultRoles() map[RoleType]Role {
	return map[RoleType]Role{
		RoleTenantAdmin: {
			Type:        RoleTenantAdmin,
			Name:        "租户管理员",
			Description: "拥有租户内所有权限",
			Permissions: []Permission{
				// 用户管理
				PermissionUserManage, PermissionUserView, PermissionUserCreate, PermissionUserUpdate, PermissionUserDelete,
				// 商户管理
				PermissionMerchantManage, PermissionMerchantView, PermissionMerchantCreate, PermissionMerchantUpdate, PermissionMerchantDelete,
				// 订单管理
				PermissionOrderManage, PermissionOrderView, PermissionOrderCreate, PermissionOrderUpdate, PermissionOrderDelete,
				// 商品管理
				PermissionProductManage, PermissionProductView, PermissionProductCreate, PermissionProductUpdate, PermissionProductDelete,
				// 租户管理
				PermissionTenantView, PermissionTenantUpdate,
				// 报表管理
				PermissionReportView, PermissionReportExport, PermissionReportCreate, PermissionReportDelete,
				// 资金管理
				PermissionFundView, PermissionFundManage, PermissionFundWithdraw, PermissionFundTransfer,
				// 权益管理
				PermissionBenefitView, PermissionBenefitManage, PermissionBenefitCreate, PermissionBenefitUpdate, PermissionBenefitDelete,
				// 系统管理
				PermissionSystemConfig, PermissionSystemAudit, PermissionSystemLog,
				// 角色管理
				PermissionRoleView, PermissionRoleManage, PermissionRoleAssign,
			},
		},
		RoleMerchant: {
			Type:        RoleMerchant,
			Name:        "商户",
			Description: "管理自己的商品和订单",
			Permissions: []Permission{
				// 订单管理
				PermissionOrderView, PermissionOrderCreate, PermissionOrderUpdate,
				// 商品管理
				PermissionProductManage, PermissionProductView, PermissionProductCreate, PermissionProductUpdate, PermissionProductDelete,
				// 报表查看
				PermissionReportView,
				// 资金查看
				PermissionFundView, PermissionFundWithdraw,
				// 权益查看
				PermissionBenefitView,
			},
		},
		RoleCustomer: {
			Type:        RoleCustomer,
			Name:        "客户",
			Description: "查看商品和管理自己的订单",
			Permissions: []Permission{
				PermissionProductView,
				PermissionOrderView, PermissionOrderCreate,
			},
		},
	}
}

// HasPermission 检查角色是否拥有指定权限
func (r Role) HasPermission(permission Permission) bool {
	for _, p := range r.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// UserPermissions 用户权限信息
type UserPermissions struct {
	UserID      uint64       `json:"user_id"`
	TenantID    uint64       `json:"tenant_id"`
	Roles       []RoleType   `json:"roles"`
	Permissions []Permission `json:"permissions"`
}

// HasPermission 检查用户是否拥有指定权限
func (up UserPermissions) HasPermission(permission Permission) bool {
	for _, p := range up.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasRole 检查用户是否拥有指定角色
func (up UserPermissions) HasRole(role RoleType) bool {
	for _, r := range up.Roles {
		if r == role {
			return true
		}
	}
	return false
}