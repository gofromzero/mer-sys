package controller

import (
	"fmt"
	"strconv"

	"github.com/gofromzero/mer-sys/backend/shared/audit"
	"github.com/gofromzero/mer-sys/backend/shared/middleware"
	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gofromzero/mer-sys/backend/shared/utils"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/grand"
)

// MerchantUserController 商户用户控制器
type MerchantUserController struct {
	userRepo *repository.UserRepository
}

// NewMerchantUserController 创建商户用户控制器
func NewMerchantUserController() *MerchantUserController {
	return &MerchantUserController{
		userRepo: repository.NewUserRepository(),
	}
}

// CreateMerchantUser 创建商户用户
func (c *MerchantUserController) CreateMerchantUser(r *ghttp.Request) {
	var req types.CreateMerchantUserRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "请求参数解析失败",
			"data":    nil,
		})
		return
	}

	ctx := r.GetCtx()
	
	// 生成UUID
	uuid := grand.S(32)
	
	// 使用bcrypt生成安全的密码哈希
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		g.Log().Errorf(ctx, "密码加密失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "密码加密失败",
			"data":    nil,
		})
		return
	}
	
	// 构建用户对象
	user := &types.User{
		UUID:         uuid,
		Username:     req.Username,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: passwordHash,
		MerchantID:   &req.MerchantID,
		Status:       types.UserStatusActive,
		Profile:      req.Profile,
		CreatedAt:    gtime.Now().Time,
		UpdatedAt:    gtime.Now().Time,
	}

	// 创建商户用户
	if err := c.userRepo.CreateMerchantUser(ctx, user, req.RoleType); err != nil {
		g.Log().Errorf(ctx, "创建商户用户失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// 记录审计日志
	operatorUserID := uint64(0)
	tenantID := uint64(1) // 从上下文获取
	if uid, ok := middleware.GetMerchantUserFromContext(ctx); ok && uid != nil {
		operatorUserID = uid.ID
		tenantID = uid.TenantID
	}
	
	audit.LogMerchantUserCreate(ctx, tenantID, req.MerchantID, operatorUserID, user.ID, user.Username, map[string]interface{}{
		"role_type": req.RoleType,
		"email":     user.Email,
		"phone":     user.Phone,
	})

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "创建商户用户成功",
		"data": g.Map{
			"uuid":        user.UUID,
			"username":    user.Username,
			"email":       user.Email,
			"merchant_id": user.MerchantID,
			"role_type":   req.RoleType,
		},
	})
}

// ListMerchantUsers 查询商户用户列表
func (c *MerchantUserController) ListMerchantUsers(r *ghttp.Request) {
	merchantIDStr := r.Get("merchant_id").String()
	if merchantIDStr == "" {
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "商户ID不能为空",
			"data":    nil,
		})
		return
	}

	merchantID, err := strconv.ParseUint(merchantIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "商户ID格式不正确",
			"data":    nil,
		})
		return
	}

	page := r.Get("page", 1).Int()
	pageSize := r.Get("pageSize", 20).Int()
	searchKeyword := r.Get("search", "").String()

	ctx := r.GetCtx()

	// 查询商户用户列表
	users, total, err := c.userRepo.FindMerchantUsers(ctx, merchantID, page, pageSize, searchKeyword)
	if err != nil {
		g.Log().Errorf(ctx, "查询商户用户列表失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "查询商户用户列表失败",
			"data":    nil,
		})
		return
	}

	// 构建返回数据
	var userList []g.Map
	for _, user := range users {
		// 获取用户角色
		roles, _ := c.userRepo.GetMerchantUserRoles(ctx, user.ID, merchantID)
		
		userInfo := g.Map{
			"id":           user.ID,
			"uuid":         user.UUID,
			"username":     user.Username,
			"email":        user.Email,
			"phone":        user.Phone,
			"status":       user.Status,
			"merchant_id":  user.MerchantID,
			"roles":        roles,
			"profile":      user.Profile,
			"created_at":   user.CreatedAt,
			"updated_at":   user.UpdatedAt,
			"last_login_at": user.LastLoginAt,
		}
		userList = append(userList, userInfo)
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "查询成功",
		"data": g.Map{
			"list":      userList,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetMerchantUser 获取商户用户详情
func (c *MerchantUserController) GetMerchantUser(r *ghttp.Request) {
	userIDStr := r.Get("id").String()
	merchantIDStr := r.Get("merchant_id").String()

	if userIDStr == "" || merchantIDStr == "" {
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "用户ID和商户ID不能为空",
			"data":    nil,
		})
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "用户ID格式不正确",
			"data":    nil,
		})
		return
	}

	merchantID, err := strconv.ParseUint(merchantIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "商户ID格式不正确",
			"data":    nil,
		})
		return
	}

	ctx := r.GetCtx()

	// 查询商户用户
	user, err := c.userRepo.FindMerchantUserByID(ctx, userID, merchantID)
	if err != nil {
		g.Log().Errorf(ctx, "查询商户用户失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// 获取用户角色
	roles, _ := c.userRepo.GetMerchantUserRoles(ctx, userID, merchantID)

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "查询成功",
		"data": g.Map{
			"id":           user.ID,
			"uuid":         user.UUID,
			"username":     user.Username,
			"email":        user.Email,
			"phone":        user.Phone,
			"status":       user.Status,
			"merchant_id":  user.MerchantID,
			"roles":        roles,
			"profile":      user.Profile,
			"created_at":   user.CreatedAt,
			"updated_at":   user.UpdatedAt,
			"last_login_at": user.LastLoginAt,
		},
	})
}

// UpdateMerchantUser 更新商户用户信息
func (c *MerchantUserController) UpdateMerchantUser(r *ghttp.Request) {
	userIDStr := r.Get("id").String()
	merchantIDStr := r.Get("merchant_id").String()

	if userIDStr == "" || merchantIDStr == "" {
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "用户ID和商户ID不能为空",
			"data":    nil,
		})
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "用户ID格式不正确",
			"data":    nil,
		})
		return
	}

	merchantID, err := strconv.ParseUint(merchantIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "商户ID格式不正确",
			"data":    nil,
		})
		return
	}

	var req types.UpdateMerchantUserRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "请求参数解析失败",
			"data":    nil,
		})
		return
	}

	ctx := r.GetCtx()

	// 构建更新数据
	updateData := g.Map{
		"updated_at": gtime.Now(),
	}

	if req.Username != "" {
		updateData["username"] = req.Username
	}
	if req.Email != "" {
		updateData["email"] = req.Email
	}
	if req.Phone != "" {
		updateData["phone"] = req.Phone
	}
	if req.Status != "" {
		updateData["status"] = req.Status
	}
	if req.Profile != nil {
		updateData["profile"] = req.Profile
	}

	// 更新商户用户信息
	if err := c.userRepo.UpdateMerchantUser(ctx, userID, merchantID, updateData); err != nil {
		g.Log().Errorf(ctx, "更新商户用户失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "更新成功",
		"data":    nil,
	})
}

// UpdateMerchantUserStatus 更新商户用户状态
func (c *MerchantUserController) UpdateMerchantUserStatus(r *ghttp.Request) {
	userIDStr := r.Get("id").String()
	merchantIDStr := r.Get("merchant_id").String()

	if userIDStr == "" || merchantIDStr == "" {
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "用户ID和商户ID不能为空",
			"data":    nil,
		})
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "用户ID格式不正确",
			"data":    nil,
		})
		return
	}

	merchantID, err := strconv.ParseUint(merchantIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "商户ID格式不正确",
			"data":    nil,
		})
		return
	}

	status := types.UserStatus(r.Get("status").String())
	if status == "" {
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "状态不能为空",
			"data":    nil,
		})
		return
	}

	ctx := r.GetCtx()

	// 先获取用户信息用于审计日志
	targetUser, err := c.userRepo.FindMerchantUserByID(ctx, userID, merchantID)
	if err != nil {
		g.Log().Errorf(ctx, "获取商户用户信息失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "用户不存在",
			"data":    nil,
		})
		return
	}

	oldStatus := targetUser.Status

	// 更新商户用户状态
	if err := c.userRepo.UpdateMerchantUserStatus(ctx, userID, merchantID, status); err != nil {
		g.Log().Errorf(ctx, "更新商户用户状态失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// 记录状态变更审计日志
	operatorUserID := uint64(0)
	tenantID := targetUser.TenantID
	if uid, ok := middleware.GetMerchantUserFromContext(ctx); ok && uid != nil {
		operatorUserID = uid.ID
		tenantID = uid.TenantID
	}
	
	audit.LogMerchantUserStatusChange(ctx, tenantID, merchantID, operatorUserID, userID, targetUser.Username, string(oldStatus), string(status), nil)

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "状态更新成功",
		"data":    nil,
	})
}

// ResetMerchantUserPassword 重置商户用户密码
func (c *MerchantUserController) ResetMerchantUserPassword(r *ghttp.Request) {
	userIDStr := r.Get("id").String()
	merchantIDStr := r.Get("merchant_id").String()

	if userIDStr == "" || merchantIDStr == "" {
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "用户ID和商户ID不能为空",
			"data":    nil,
		})
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "用户ID格式不正确",
			"data":    nil,
		})
		return
	}

	merchantID, err := strconv.ParseUint(merchantIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "商户ID格式不正确",
			"data":    nil,
		})
		return
	}

	ctx := r.GetCtx()

	// 生成随机密码（8位）
	newPassword, err := utils.GenerateRandomPassword(8)
	if err != nil {
		g.Log().Errorf(ctx, "生成随机密码失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "生成随机密码失败",
			"data":    nil,
		})
		return
	}

	// 使用bcrypt生成安全的密码哈希
	passwordHash, err := utils.HashPassword(newPassword)
	if err != nil {
		g.Log().Errorf(ctx, "密码加密失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "密码加密失败",
			"data":    nil,
		})
		return
	}

	// 先获取用户信息用于审计日志
	targetUser, err := c.userRepo.FindMerchantUserByID(ctx, userID, merchantID)
	if err != nil {
		g.Log().Errorf(ctx, "获取商户用户信息失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "用户不存在",
			"data":    nil,
		})
		return
	}

	// 重置商户用户密码
	if err := c.userRepo.ResetMerchantUserPassword(ctx, userID, merchantID, passwordHash); err != nil {
		g.Log().Errorf(ctx, "重置商户用户密码失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// 记录敏感操作审计日志
	operatorUserID := uint64(0)
	tenantID := targetUser.TenantID
	if uid, ok := middleware.GetMerchantUserFromContext(ctx); ok && uid != nil {
		operatorUserID = uid.ID
		tenantID = uid.TenantID
	}
	
	audit.LogMerchantUserPasswordReset(ctx, tenantID, merchantID, operatorUserID, userID, targetUser.Username, "admin_reset")

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "密码重置成功",
		"data": g.Map{
			"new_password": newPassword,
		},
	})
}


// BatchCreateMerchantUsers 批量创建商户用户
func (c *MerchantUserController) BatchCreateMerchantUsers(r *ghttp.Request) {
	type BatchCreateRequest struct {
		MerchantID uint64                            `json:"merchant_id" validate:"required"`
		Users      []types.CreateMerchantUserRequest `json:"users" validate:"required,min=1,max=50"`
	}

	var req BatchCreateRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "请求参数解析失败",
			"data":    nil,
		})
		return
	}

	ctx := r.GetCtx()

	var successCount int
	var failedUsers []g.Map

	// 逐个创建用户
	for i, userReq := range req.Users {
		userReq.MerchantID = req.MerchantID
		
		// 生成UUID
		uuid := grand.S(32)
		
		// 使用bcrypt生成安全的密码哈希
		passwordHash, err := utils.HashPassword(userReq.Password)
		if err != nil {
			g.Log().Errorf(ctx, "批量创建商户用户-密码加密失败[%d]: %v", i, err)
			failedUsers = append(failedUsers, g.Map{
				"index":    i + 1,
				"username": userReq.Username,
				"email":    userReq.Email,
				"error":    "密码加密失败",
			})
			continue
		}
		
		// 构建用户对象
		user := &types.User{
			UUID:         uuid,
			Username:     userReq.Username,
			Email:        userReq.Email,
			Phone:        userReq.Phone,
			PasswordHash: passwordHash,
			MerchantID:   &userReq.MerchantID,
			Status:       types.UserStatusActive,
			Profile:      userReq.Profile,
			CreatedAt:    gtime.Now().Time,
			UpdatedAt:    gtime.Now().Time,
		}

		// 创建商户用户
		if err := c.userRepo.CreateMerchantUser(ctx, user, userReq.RoleType); err != nil {
			g.Log().Errorf(ctx, "批量创建商户用户失败[%d]: %v", i, err)
			failedUsers = append(failedUsers, g.Map{
				"index":    i + 1,
				"username": userReq.Username,
				"email":    userReq.Email,
				"error":    err.Error(),
			})
		} else {
			successCount++
		}
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": fmt.Sprintf("批量创建完成，成功%d个，失败%d个", successCount, len(failedUsers)),
		"data": g.Map{
			"success_count": successCount,
			"failed_count":  len(failedUsers),
			"failed_users":  failedUsers,
		},
	})
}

// GetMerchantUserAuditLogs 获取商户用户审计日志
func (c *MerchantUserController) GetMerchantUserAuditLogs(r *ghttp.Request) {
	merchantIDStr := r.Get("merchant_id").String()
	if merchantIDStr == "" {
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "商户ID不能为空",
			"data":    nil,
		})
		return
	}

	merchantID, err := strconv.ParseUint(merchantIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "商户ID格式不正确",
			"data":    nil,
		})
		return
	}

	ctx := r.GetCtx()

	// 解析查询参数
	var userID *uint64
	if userIDStr := r.Get("user_id").String(); userIDStr != "" {
		if uid, err := strconv.ParseUint(userIDStr, 10, 64); err == nil {
			userID = &uid
		}
	}

	// 解析事件类型
	var eventTypes []audit.AuditEventType
	if eventTypesStr := r.Get("event_types").String(); eventTypesStr != "" {
		// 简单解析，实际应该使用JSON或逗号分隔
		eventTypes = append(eventTypes, audit.AuditEventType(eventTypesStr))
	}

	// 解析分页参数
	page := r.Get("page").Int()
	if page <= 0 {
		page = 1
	}
	pageSize := r.Get("page_size").Int()
	if pageSize <= 0 {
		pageSize = 20
	}

	// 获取审计日志
	logs, total, err := audit.GetMerchantUserAuditLogs(ctx, merchantID, userID, eventTypes, nil, nil, page, pageSize)
	if err != nil {
		g.Log().Errorf(ctx, "获取商户用户审计日志失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "获取审计日志失败",
			"data":    nil,
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "获取审计日志成功",
		"data": g.Map{
			"logs":      logs,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
			"total_pages": (total + pageSize - 1) / pageSize,
		},
	})
}

// GetMerchantUserOperationHistory 获取特定商户用户的操作历史
func (c *MerchantUserController) GetMerchantUserOperationHistory(r *ghttp.Request) {
	userIDStr := r.Get("user_id").String()
	merchantIDStr := r.Get("merchant_id").String()

	if userIDStr == "" || merchantIDStr == "" {
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "用户ID和商户ID不能为空",
			"data":    nil,
		})
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "用户ID格式不正确",
			"data":    nil,
		})
		return
	}

	merchantID, err := strconv.ParseUint(merchantIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "商户ID格式不正确",
			"data":    nil,
		})
		return
	}

	ctx := r.GetCtx()

	// 解析分页参数
	page := r.Get("page").Int()
	if page <= 0 {
		page = 1
	}
	pageSize := r.Get("page_size").Int()
	if pageSize <= 0 {
		pageSize = 10
	}

	// 获取用户操作历史
	logs, total, err := audit.GetMerchantUserAuditLogs(ctx, merchantID, &userID, nil, nil, nil, page, pageSize)
	if err != nil {
		g.Log().Errorf(ctx, "获取用户操作历史失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": "获取用户操作历史失败",
			"data":    nil,
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "获取用户操作历史成功",
		"data": g.Map{
			"user_id":     userID,
			"merchant_id": merchantID,
			"logs":        logs,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + pageSize - 1) / pageSize,
		},
	})
}