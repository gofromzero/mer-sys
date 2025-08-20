package repository

import (
	"context"
	"fmt"

	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// UserRepository 用户仓储
type UserRepository struct {
	*BaseRepository
}

// NewUserRepository 创建用户仓储实例
func NewUserRepository() *UserRepository {
	return &UserRepository{
		BaseRepository: NewBaseRepository("users"),
	}
}

// GetByID 根据ID查找用户（自动添加租户隔离）- 兼容性别名
func (r *UserRepository) GetByID(ctx context.Context, id uint64) (*types.User, error) {
	return r.FindByID(ctx, id)
}

// FindByID 根据ID查找用户（自动添加租户隔离）
func (r *UserRepository) FindByID(ctx context.Context, id uint64) (*types.User, error) {
	record, err := r.FindOne(ctx, "id", id)
	if err != nil {
		return nil, err
	}

	if record.IsEmpty() {
		return nil, fmt.Errorf("用户不存在: %d", id)
	}

	var user types.User
	if err := record.Struct(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// FindByUUID 根据UUID查找用户（自动添加租户隔离）
func (r *UserRepository) FindByUUID(ctx context.Context, uuid string) (*types.User, error) {
	record, err := r.FindOne(ctx, "uuid", uuid)
	if err != nil {
		return nil, err
	}

	if record.IsEmpty() {
		return nil, fmt.Errorf("用户不存在: %s", uuid)
	}

	var user types.User
	if err := record.Struct(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// FindByUsername 根据用户名查找用户（自动添加租户隔离）
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*types.User, error) {
	record, err := r.FindOne(ctx, "username", username)
	if err != nil {
		return nil, err
	}

	if record.IsEmpty() {
		return nil, fmt.Errorf("用户不存在: %s", username)
	}

	var user types.User
	if err := record.Struct(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// FindByEmail 根据邮箱查找用户（自动添加租户隔离）
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*types.User, error) {
	record, err := r.FindOne(ctx, "email", email)
	if err != nil {
		return nil, err
	}

	if record.IsEmpty() {
		return nil, fmt.Errorf("用户不存在: %s", email)
	}

	var user types.User
	if err := record.Struct(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// Create 创建用户（自动添加租户ID）
func (r *UserRepository) Create(ctx context.Context, user *types.User) error {
	tenantID, err := r.GetTenantID(ctx)
	if err != nil {
		return err
	}

	// 检查用户名是否在当前租户下已存在
	exists, err := r.Exists(ctx, "username", user.Username)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("用户名已存在: %s", user.Username)
	}

	// 检查邮箱是否在当前租户下已存在
	exists, err = r.Exists(ctx, "email", user.Email)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("邮箱已存在: %s", user.Email)
	}

	// 设置租户ID
	user.TenantID = tenantID

	// 插入用户数据
	_, err = r.Insert(ctx, user)
	return err
}

// UpdateByID 根据ID更新用户（自动添加租户隔离）
func (r *UserRepository) UpdateByID(ctx context.Context, id uint64, data interface{}) error {
	_, err := r.Update(ctx, data, "id", id)
	return err
}

// UpdateByUUID 根据UUID更新用户（自动添加租户隔离）
func (r *UserRepository) UpdateByUUID(ctx context.Context, uuid string, data interface{}) error {
	_, err := r.Update(ctx, data, "uuid", uuid)
	return err
}

// DeleteByID 根据ID删除用户（自动添加租户隔离）
func (r *UserRepository) DeleteByID(ctx context.Context, id uint64) error {
	_, err := r.Delete(ctx, "id", id)
	return err
}

// FindAllByTenant 查找租户下的所有用户
func (r *UserRepository) FindAllByTenant(ctx context.Context) ([]*types.User, error) {
	records, err := r.FindAll(ctx, nil)
	if err != nil {
		return nil, err
	}

	var users []*types.User
	for _, record := range records {
		var user types.User
		if err := record.Struct(&user); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}

// FindPageByTenant 分页查找租户下的用户
func (r *UserRepository) FindPageByTenant(ctx context.Context, page, pageSize int, condition interface{}, args ...interface{}) ([]*types.User, int, error) {
	records, total, err := r.FindPage(ctx, page, pageSize, condition, args...)
	if err != nil {
		return nil, 0, err
	}

	var users []*types.User
	for _, record := range records {
		var user types.User
		if err := record.Struct(&user); err != nil {
			return nil, 0, err
		}
		users = append(users, &user)
	}

	return users, total, nil
}

// UpdateStatus 更新用户状态（自动添加租户隔离）
func (r *UserRepository) UpdateStatus(ctx context.Context, id uint64, status types.UserStatus) error {
	_, err := r.Update(ctx, gdb.Map{
		"status": status,
	}, "id", id)
	return err
}

// UpdateLastLogin 更新最后登录时间（自动添加租户隔离）
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uint64) error {
	_, err := r.Update(ctx, gdb.Map{
		"last_login_at": "NOW()",
	}, "id", id)
	return err
}

// FindByUsernameAndTenant 根据用户名和租户ID查找用户
func (r *UserRepository) FindByUsernameAndTenant(ctx context.Context, username string, tenantID uint64) (*types.User, error) {
	// 根据用户名和租户ID查找用户
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	record, err := model.Where("username = ? AND tenant_id = ?", username, tenantID).One()
	if err != nil {
		return nil, err
	}

	if record.IsEmpty() {
		return nil, fmt.Errorf("用户不存在: %s", username)
	}

	var user types.User
	if err := record.Struct(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// FindByEmailAndTenant 根据邮箱和租户ID查找用户
func (r *UserRepository) FindByEmailAndTenant(ctx context.Context, email string, tenantID uint64) (*types.User, error) {
	// 根据邮箱和租户ID查找用户
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	record, err := model.Where("email = ? AND tenant_id = ?", email, tenantID).One()
	if err != nil {
		return nil, err
	}

	if record.IsEmpty() {
		return nil, fmt.Errorf("用户不存在: %s", email)
	}

	var user types.User
	if err := record.Struct(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// FindByIDAndTenant 根据ID和租户ID查找用户
func (r *UserRepository) FindByIDAndTenant(ctx context.Context, userID, tenantID uint64) (*types.User, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	record, err := model.Where("id = ? AND tenant_id = ?", userID, tenantID).One()
	if err != nil {
		return nil, err
	}

	if record.IsEmpty() {
		return nil, fmt.Errorf("用户不存在: %d", userID)
	}

	var user types.User
	if err := record.Struct(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateLastLoginTime 更新用户最后登录时间
func (r *UserRepository) UpdateLastLoginTime(ctx context.Context, userID uint64, loginTime *gtime.Time) error {
	_, err := r.Update(ctx, gdb.Map{
		"last_login_at": loginTime,
		"updated_at":    gtime.Now(),
	}, "id", userID)
	return err
}

// UpdatePassword 更新用户密码
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uint64, passwordHash string) error {
	_, err := r.Update(ctx, gdb.Map{
		"password_hash": passwordHash,
		"updated_at":    gtime.Now(),
	}, "id", userID)
	return err
}

// GetUserRoles 获取用户角色列表
func (r *UserRepository) GetUserRoles(ctx context.Context, userID, tenantID uint64) ([]types.RoleType, error) {
	// 从user_roles表查询用户角色
	records, err := g.DB().Model("user_roles").
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Fields("role_type").
		All()

	if err != nil {
		g.Log().Errorf(ctx, "查询用户角色失败: %v", err)
		return nil, err
	}

	var roles []types.RoleType
	for _, record := range records {
		roleType := types.RoleType(record["role_type"].String())
		roles = append(roles, roleType)
	}

	return roles, nil
}

// AssignRole 为用户分配角色
func (r *UserRepository) AssignRole(ctx context.Context, userID, tenantID uint64, roleType types.RoleType) error {
	// 检查角色是否已存在
	exists, err := g.DB().Model("user_roles").
		Where("user_id = ? AND tenant_id = ? AND role_type = ?", userID, tenantID, roleType).
		Count()

	if err != nil {
		return err
	}

	if exists > 0 {
		return fmt.Errorf("用户已拥有角色: %s", roleType)
	}

	// 插入用户角色记录
	_, err = g.DB().Model("user_roles").Insert(gdb.Map{
		"user_id":    userID,
		"tenant_id":  tenantID,
		"role_type":  roleType,
		"created_at": gtime.Now(),
		"updated_at": gtime.Now(),
	})

	return err
}

// RemoveRole 移除用户角色
func (r *UserRepository) RemoveRole(ctx context.Context, userID, tenantID uint64, roleType types.RoleType) error {
	_, err := g.DB().Model("user_roles").
		Where("user_id = ? AND tenant_id = ? AND role_type = ?", userID, tenantID, roleType).
		Delete()

	return err
}

// ========================= 商户用户专用方法 =========================

// CreateMerchantUser 创建商户用户
func (r *UserRepository) CreateMerchantUser(ctx context.Context, user *types.User, roleType types.RoleType) error {
	if user.MerchantID == nil {
		return fmt.Errorf("商户用户必须关联商户")
	}

	// 验证商户用户数据
	if err := user.ValidateMerchantUser(); err != nil {
		return err
	}

	tenantID, err := r.GetTenantID(ctx)
	if err != nil {
		return err
	}

	// 检查用户名在当前租户+商户下是否已存在
	exists, err := r.MerchantUserExists(ctx, "username", user.Username, *user.MerchantID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("用户名在该商户下已存在: %s", user.Username)
	}

	// 检查邮箱在当前租户+商户下是否已存在
	exists, err = r.MerchantUserExists(ctx, "email", user.Email, *user.MerchantID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("邮箱在该商户下已存在: %s", user.Email)
	}

	// 设置租户ID
	user.TenantID = tenantID

	// 开始事务
	tx, err := g.DB().Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// 插入用户数据
	result, err := tx.Model("users").Insert(user)
	if err != nil {
		return err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// 分配商户角色
	_, err = tx.Model("user_roles").Insert(gdb.Map{
		"user_id":     userID,
		"tenant_id":   tenantID,
		"role_type":   roleType,
		"resource_id": user.MerchantID, // 商户角色的resource_id是merchant_id
		"created_at":  gtime.Now(),
		"updated_at":  gtime.Now(),
	})

	return err
}

// FindMerchantUsers 查找商户下的用户列表
func (r *UserRepository) FindMerchantUsers(ctx context.Context, merchantID uint64, page, pageSize int, searchKeyword string) ([]*types.User, int, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, 0, err
	}

	// 构建查询条件
	query := model.Where("merchant_id = ?", merchantID)

	// 如果有搜索关键词，添加搜索条件
	if searchKeyword != "" {
		query = query.Where("username LIKE ? OR email LIKE ? OR phone LIKE ?", 
			"%"+searchKeyword+"%", "%"+searchKeyword+"%", "%"+searchKeyword+"%")
	}

	// 获取总数
	total, err := query.Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	records, err := query.
		Page(page, pageSize).
		OrderDesc("created_at").
		All()
	if err != nil {
		return nil, 0, err
	}

	var users []*types.User
	for _, record := range records {
		var user types.User
		if err := record.Struct(&user); err != nil {
			return nil, 0, err
		}
		users = append(users, &user)
	}

	return users, total, nil
}

// FindMerchantUserByID 根据ID查找商户用户
func (r *UserRepository) FindMerchantUserByID(ctx context.Context, userID, merchantID uint64) (*types.User, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}

	record, err := model.Where("id = ? AND merchant_id = ?", userID, merchantID).One()
	if err != nil {
		return nil, err
	}

	if record.IsEmpty() {
		return nil, fmt.Errorf("商户用户不存在: %d", userID)
	}

	var user types.User
	if err := record.Struct(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateMerchantUserStatus 更新商户用户状态
func (r *UserRepository) UpdateMerchantUserStatus(ctx context.Context, userID, merchantID uint64, status types.UserStatus) error {
	model, err := r.Model(ctx)
	if err != nil {
		return err
	}

	_, err = model.Where("id = ? AND merchant_id = ?", userID, merchantID).
		Update(gdb.Map{
			"status":     status,
			"updated_at": gtime.Now(),
		})

	return err
}

// UpdateMerchantUser 更新商户用户信息
func (r *UserRepository) UpdateMerchantUser(ctx context.Context, userID, merchantID uint64, data interface{}) error {
	model, err := r.Model(ctx)
	if err != nil {
		return err
	}

	_, err = model.Where("id = ? AND merchant_id = ?", userID, merchantID).
		Update(data)

	return err
}

// ResetMerchantUserPassword 重置商户用户密码
func (r *UserRepository) ResetMerchantUserPassword(ctx context.Context, userID, merchantID uint64, newPasswordHash string) error {
	model, err := r.Model(ctx)
	if err != nil {
		return err
	}

	_, err = model.Where("id = ? AND merchant_id = ?", userID, merchantID).
		Update(gdb.Map{
			"password_hash": newPasswordHash,
			"updated_at":    gtime.Now(),
		})

	return err
}

// MerchantUserExists 检查商户用户是否存在
func (r *UserRepository) MerchantUserExists(ctx context.Context, field string, value interface{}, merchantID uint64) (bool, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return false, err
	}

	count, err := model.Where(fmt.Sprintf("%s = ? AND merchant_id = ?", field), value, merchantID).Count()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetMerchantUserRoles 获取商户用户角色列表
func (r *UserRepository) GetMerchantUserRoles(ctx context.Context, userID, merchantID uint64) ([]types.RoleType, error) {
	tenantID, err := r.GetTenantID(ctx)
	if err != nil {
		return nil, err
	}

	// 从user_roles表查询商户用户角色
	records, err := g.DB().Model("user_roles").
		Where("user_id = ? AND tenant_id = ? AND resource_id = ?", userID, tenantID, merchantID).
		Fields("role_type").
		All()

	if err != nil {
		g.Log().Errorf(ctx, "查询商户用户角色失败: %v", err)
		return nil, err
	}

	var roles []types.RoleType
	for _, record := range records {
		roleType := types.RoleType(record["role_type"].String())
		roles = append(roles, roleType)
	}

	return roles, nil
}

// AssignMerchantUserRole 为商户用户分配角色
func (r *UserRepository) AssignMerchantUserRole(ctx context.Context, userID, merchantID uint64, roleType types.RoleType) error {
	tenantID, err := r.GetTenantID(ctx)
	if err != nil {
		return err
	}

	// 检查角色是否已存在
	exists, err := g.DB().Model("user_roles").
		Where("user_id = ? AND tenant_id = ? AND resource_id = ? AND role_type = ?", 
			userID, tenantID, merchantID, roleType).
		Count()

	if err != nil {
		return err
	}

	if exists > 0 {
		return fmt.Errorf("用户已拥有该商户角色: %s", roleType)
	}

	// 插入用户角色记录
	_, err = g.DB().Model("user_roles").Insert(gdb.Map{
		"user_id":     userID,
		"tenant_id":   tenantID,
		"resource_id": merchantID,
		"role_type":   roleType,
		"created_at":  gtime.Now(),
		"updated_at":  gtime.Now(),
	})

	return err
}
