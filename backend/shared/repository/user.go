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
