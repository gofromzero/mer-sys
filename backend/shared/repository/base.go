package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gofromzero/mer-sys/backend/shared/audit"
)

// BaseRepository 基础仓储类，提供多租户隔离支持
type BaseRepository struct {
	db       gdb.DB
	tableName string
}

// NewBaseRepository 创建基础仓储实例
func NewBaseRepository() *BaseRepository {
	return &BaseRepository{
		db: g.DB(),
	}
}

// GetDB 获取数据库实例
func (r *BaseRepository) GetDB() gdb.DB {
	return r.db
}

// GetTableName 获取表名
func (r *BaseRepository) GetTableName() string {
	return r.tableName
}

// GetTenantID 从上下文中获取租户ID
func (r *BaseRepository) GetTenantID(ctx context.Context) uint64 {
	tenantID := ctx.Value("tenant_id")
	if tenantID == nil {
		return 0
	}
	
	switch v := tenantID.(type) {
	case uint64:
		return v
	case int64:
		return uint64(v)
	case int:
		return uint64(v)
	case string:
		return gconv.Uint64(v)
	default:
		return 0
	}
}

// GetUserID 从上下文中获取用户ID
func (r *BaseRepository) GetUserID(ctx context.Context) uint64 {
	userID := ctx.Value("user_id")
	if userID == nil {
		return 0
	}
	
	switch v := userID.(type) {
	case uint64:
		return v
	case int64:
		return uint64(v)
	case int:
		return uint64(v)
	case string:
		return gconv.Uint64(v)
	default:
		return 0
	}
}

// GetMerchantID 从上下文中获取商户ID
func (r *BaseRepository) GetMerchantID(ctx context.Context) uint64 {
	merchantID := ctx.Value("merchant_id")
	if merchantID == nil {
		return 0
	}
	
	switch v := merchantID.(type) {
	case uint64:
		return v
	case int64:
		return uint64(v)
	case int:
		return uint64(v)
	case string:
		return gconv.Uint64(v)
	default:
		return 0
	}
}

// Model 获取带租户隔离的模型
func (r *BaseRepository) Model(ctx context.Context) (*gdb.Model, error) {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		// 记录无效的租户访问尝试
		audit.LogSecurityViolation(ctx, 0, "invalid_tenant_context", 
			"尝试在无效租户上下文中访问数据", map[string]interface{}{
				"table": r.tableName,
			})
		return nil, fmt.Errorf("missing tenant_id in context")
	}
	
	// 记录正常的租户数据访问
	audit.LogTenantAccess(ctx, tenantID, r.tableName, "query", nil)
	
	return r.db.Model(r.tableName).Where("tenant_id", tenantID), nil
}

// ModelWithoutTenant 获取不带租户隔离的模型（慎用）
func (r *BaseRepository) ModelWithoutTenant() *gdb.Model {
	return r.db.Model(r.tableName)
}

// Insert 插入数据（自动添加租户ID）
func (r *BaseRepository) Insert(ctx context.Context, data interface{}) (sql.Result, error) {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("missing tenant_id in context")
	}
	
	// 确保数据包含租户ID
	dataMap := gconv.Map(data)
	dataMap["tenant_id"] = tenantID
	
	return r.db.Model(r.tableName).Insert(dataMap)
}

// InsertAndGetId 插入数据并返回ID
func (r *BaseRepository) InsertAndGetId(ctx context.Context, data interface{}) (int64, error) {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return 0, fmt.Errorf("missing tenant_id in context")
	}
	
	dataMap := gconv.Map(data)
	dataMap["tenant_id"] = tenantID
	
	return r.db.Model(r.tableName).InsertAndGetId(dataMap)
}

// Update 更新数据（自动添加租户隔离）
func (r *BaseRepository) Update(ctx context.Context, data interface{}, condition interface{}, args ...interface{}) (sql.Result, error) {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("missing tenant_id in context")
	}
	
	model := r.db.Model(r.tableName).Where("tenant_id", tenantID)
	if condition != nil {
		model = model.Where(condition, args...)
	}
	
	return model.Update(data)
}

// Delete 删除数据（自动添加租户隔离）
func (r *BaseRepository) Delete(ctx context.Context, condition interface{}, args ...interface{}) (sql.Result, error) {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("missing tenant_id in context")
	}
	
	model := r.db.Model(r.tableName).Where("tenant_id", tenantID)
	if condition != nil {
		model = model.Where(condition, args...)
	}
	
	return model.Delete()
}

// FindOne 查询单条数据（自动添加租户隔离）
func (r *BaseRepository) FindOne(ctx context.Context, condition interface{}, args ...interface{}) (gdb.Record, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	if condition != nil {
		model = model.Where(condition, args...)
	}
	
	return model.One()
}

// FindAll 查询所有数据（自动添加租户隔离）
func (r *BaseRepository) FindAll(ctx context.Context, condition interface{}, args ...interface{}) (gdb.Result, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	if condition != nil {
		model = model.Where(condition, args...)
	}
	
	return model.All()
}

// FindPage 分页查询（自动添加租户隔离）
func (r *BaseRepository) FindPage(ctx context.Context, page, pageSize int, condition interface{}, args ...interface{}) (gdb.Result, int, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, 0, err
	}
	
	if condition != nil {
		model = model.Where(condition, args...)
	}
	
	// 获取总数
	countModel := model.Clone()
	total, err := countModel.Count()
	if err != nil {
		return nil, 0, err
	}
	
	// 获取分页数据
	result, err := model.Page(page, pageSize).All()
	if err != nil {
		return nil, 0, err
	}
	
	return result, total, nil
}

// Count 统计数据（自动添加租户隔离）
func (r *BaseRepository) Count(ctx context.Context, condition interface{}, args ...interface{}) (int, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return 0, err
	}
	
	if condition != nil {
		model = model.Where(condition, args...)
	}
	
	return model.Count()
}

// Exists 检查数据是否存在（自动添加租户隔离）
func (r *BaseRepository) Exists(ctx context.Context, condition interface{}, args ...interface{}) (bool, error) {
	count, err := r.Count(ctx, condition, args...)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}