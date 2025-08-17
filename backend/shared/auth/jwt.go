package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/cache"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/frame/g"
)

// JWTManager JWT管理器
type JWTManager struct {
	cache      *cache.Cache
	secret     string
	expireTime time.Duration
}

// NewJWTManager 创建JWT管理器
func NewJWTManager() *JWTManager {
	secret := g.Cfg().MustGet(context.Background(), "jwt.secret", "mer-system-jwt-secret").String()
	expire := g.Cfg().MustGet(context.Background(), "jwt.expire", 24).Int()

	return &JWTManager{
		cache:      cache.NewCache("jwt"),
		secret:     secret,
		expireTime: time.Duration(expire) * time.Hour,
	}
}

// TokenClaims JWT令牌声明
type TokenClaims struct {
	UserID      uint64             `json:"user_id"`
	TenantID    uint64             `json:"tenant_id"`
	Username    string             `json:"username"`
	Email       string             `json:"email"`
	Roles       []types.RoleType   `json:"roles"`
	Permissions []types.Permission `json:"permissions"`
	TokenType   string             `json:"token_type"` // access, refresh
	IssuedAt    time.Time          `json:"issued_at"`
	ExpiresAt   time.Time          `json:"expires_at"`
}

// GenerateToken 生成JWT令牌（兼容旧接口）
func (j *JWTManager) GenerateToken(ctx context.Context, user *types.User, tokenType string) (string, error) {
	// 使用空的权限信息，仅为兼容性
	userPermissions := &types.UserPermissions{
		UserID:      user.ID,
		TenantID:    user.TenantID,
		Roles:       []types.RoleType{},
		Permissions: []types.Permission{},
	}
	return j.GenerateTokenWithPermissions(ctx, user, userPermissions, tokenType)
}

// GenerateTokenWithPermissions 生成包含角色和权限的JWT令牌
func (j *JWTManager) GenerateTokenWithPermissions(ctx context.Context, user *types.User, userPermissions *types.UserPermissions, tokenType string) (string, error) {
	now := time.Now()
	expireTime := j.expireTime
	if tokenType == "refresh" {
		expireTime = j.expireTime * 7 // 刷新令牌7倍时长
	}

	claims := &TokenClaims{
		UserID:      user.ID,
		TenantID:    user.TenantID,
		Username:    user.Username,
		Email:       user.Email,
		Roles:       userPermissions.Roles,
		Permissions: userPermissions.Permissions,
		TokenType:   tokenType,
		IssuedAt:    now,
		ExpiresAt:   now.Add(expireTime),
	}

	// 生成令牌ID（使用用户ID、类型和时间戳）
	tokenID := j.generateTokenID(user.ID, tokenType, now)

	// 将令牌信息存储到Redis
	err := j.cache.Set(ctx, tokenID, claims, expireTime)
	if err != nil {
		return "", fmt.Errorf("存储令牌失败: %v", err)
	}

	// 维护用户的活跃令牌列表
	err = j.addUserToken(ctx, user.ID, tokenID, tokenType)
	if err != nil {
		return "", fmt.Errorf("维护用户令牌列表失败: %v", err)
	}

	return tokenID, nil
}

// ValidateToken 验证JWT令牌
func (j *JWTManager) ValidateToken(ctx context.Context, token string) (*TokenClaims, error) {
	// 首先检查令牌是否在黑名单中
	isBlacklisted, err := j.IsTokenBlacklisted(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("检查令牌黑名单失败: %v", err)
	}
	if isBlacklisted {
		return nil, fmt.Errorf("令牌已被撤销")
	}

	var claims TokenClaims
	err = j.cache.GetStruct(ctx, token, &claims)
	if err != nil {
		return nil, fmt.Errorf("令牌无效或已过期: %v", err)
	}

	// 检查是否过期
	if time.Now().After(claims.ExpiresAt) {
		j.cache.Delete(ctx, token) // 清理过期令牌
		return nil, fmt.Errorf("令牌已过期")
	}

	return &claims, nil
}

// RefreshToken 刷新访问令牌
func (j *JWTManager) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	// 验证刷新令牌
	claims, err := j.ValidateToken(ctx, refreshToken)
	if err != nil {
		return "", fmt.Errorf("刷新令牌无效: %v", err)
	}

	if claims.TokenType != "refresh" {
		return "", fmt.Errorf("令牌类型错误")
	}

	// 创建新的访问令牌，保持原有的角色和权限
	user := &types.User{
		ID:       claims.UserID,
		TenantID: claims.TenantID,
		Username: claims.Username,
		Email:    claims.Email,
	}

	userPermissions := &types.UserPermissions{
		UserID:      claims.UserID,
		TenantID:    claims.TenantID,
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
	}

	return j.GenerateTokenWithPermissions(ctx, user, userPermissions, "access")
}

// RefreshTokenWithRotation 刷新令牌并轮换刷新令牌（安全增强版）
func (j *JWTManager) RefreshTokenWithRotation(ctx context.Context, refreshToken string) (accessToken, newRefreshToken string, err error) {
	// 验证刷新令牌
	claims, err := j.ValidateToken(ctx, refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("刷新令牌无效: %v", err)
	}

	if claims.TokenType != "refresh" {
		return "", "", fmt.Errorf("令牌类型错误")
	}

	// 检查刷新令牌是否接近过期（在过期前1小时内才允许刷新）
	if time.Until(claims.ExpiresAt) > time.Hour {
		return "", "", fmt.Errorf("刷新令牌尚未到刷新时间")
	}

	// 立即撤销旧的刷新令牌，防止重放攻击
	err = j.RevokeToken(ctx, refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("撤销旧刷新令牌失败: %v", err)
	}

	// 创建用户和权限信息
	user := &types.User{
		ID:       claims.UserID,
		TenantID: claims.TenantID,
		Username: claims.Username,
		Email:    claims.Email,
	}

	userPermissions := &types.UserPermissions{
		UserID:      claims.UserID,
		TenantID:    claims.TenantID,
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
	}

	// 生成新的访问令牌
	accessToken, err = j.GenerateTokenWithPermissions(ctx, user, userPermissions, "access")
	if err != nil {
		return "", "", fmt.Errorf("生成访问令牌失败: %v", err)
	}

	// 生成新的刷新令牌
	newRefreshToken, err = j.GenerateTokenWithPermissions(ctx, user, userPermissions, "refresh")
	if err != nil {
		// 如果生成新刷新令牌失败，撤销刚生成的访问令牌
		j.RevokeToken(ctx, accessToken)
		return "", "", fmt.Errorf("生成新刷新令牌失败: %v", err)
	}

	return accessToken, newRefreshToken, nil
}

// RevokeToken 撤销令牌
func (j *JWTManager) RevokeToken(ctx context.Context, token string) error {
	// 获取令牌信息（跳过黑名单检查，因为我们正在撤销它）
	var claims TokenClaims
	err := j.cache.GetStruct(ctx, token, &claims)
	if err != nil {
		// 令牌可能已经不存在，但仍需要添加到黑名单
		g.Log().Warningf(ctx, "撤销不存在的令牌: %s", token)
	}

	// 将令牌添加到黑名单
	err = j.AddTokenToBlacklist(ctx, token, claims.ExpiresAt)
	if err != nil {
		return fmt.Errorf("添加令牌到黑名单失败: %v", err)
	}

	// 删除令牌
	err = j.cache.Delete(ctx, token)
	if err != nil {
		g.Log().Warningf(ctx, "删除令牌失败: %v", err)
	}

	// 从用户令牌列表中移除
	if claims.UserID > 0 {
		err = j.removeUserToken(ctx, claims.UserID, token)
		if err != nil {
			g.Log().Warningf(ctx, "从用户令牌列表移除失败: %v", err)
		}
	}

	return nil
}

// RevokeUserTokens 撤销用户的所有令牌
func (j *JWTManager) RevokeUserTokens(ctx context.Context, userID uint64) error {
	// 获取用户的所有令牌
	tokens, err := j.getUserTokens(ctx, userID)
	if err != nil {
		return err
	}

	// 删除所有令牌
	for _, token := range tokens {
		j.cache.Delete(ctx, token)
	}

	// 清空用户令牌列表
	return j.clearUserTokens(ctx, userID)
}

// generateTokenID 生成令牌ID
func (j *JWTManager) generateTokenID(userID uint64, tokenType string, issuedAt time.Time) string {
	data := fmt.Sprintf("%d-%s-%d-%s", userID, tokenType, issuedAt.Unix(), j.secret)
	return gmd5.MustEncryptString(data)
}

// addUserToken 添加用户令牌到列表
func (j *JWTManager) addUserToken(ctx context.Context, userID uint64, token, tokenType string) error {
	key := fmt.Sprintf("user_tokens:%d", userID)
	field := fmt.Sprintf("%s:%s", tokenType, token)
	return j.cache.HSet(ctx, key, field, time.Now().Unix())
}

// removeUserToken 从用户令牌列表中移除
func (j *JWTManager) removeUserToken(ctx context.Context, userID uint64, token string) error {
	key := fmt.Sprintf("user_tokens:%d", userID)

	// 获取所有字段，找到匹配的令牌
	tokens, err := j.cache.HGetAll(ctx, key)
	if err != nil {
		return err
	}

	for field := range tokens {
		// 检查字段是否包含目标令牌
		if strings.HasSuffix(field, ":"+token) {
			return j.cache.HDel(ctx, key, field)
		}
	}

	return nil
}

// getUserTokens 获取用户的所有令牌
func (j *JWTManager) getUserTokens(ctx context.Context, userID uint64) ([]string, error) {
	key := fmt.Sprintf("user_tokens:%d", userID)

	tokens, err := j.cache.HGetAll(ctx, key)
	if err != nil {
		return nil, err
	}

	var result []string
	for field := range tokens {
		// 从字段名中提取令牌（格式：type:token）
		// field 格式：access:token 或 refresh:token
		if colonIndex := strings.Index(field, ":"); colonIndex != -1 && colonIndex < len(field)-1 {
			token := field[colonIndex+1:]
			result = append(result, token)
		}
	}

	return result, nil
}

// clearUserTokens 清空用户令牌列表
func (j *JWTManager) clearUserTokens(ctx context.Context, userID uint64) error {
	key := fmt.Sprintf("user_tokens:%d", userID)
	return j.cache.Delete(ctx, key)
}

// GetTokenInfo 获取令牌信息
func (j *JWTManager) GetTokenInfo(ctx context.Context, token string) (*TokenClaims, error) {
	return j.ValidateToken(ctx, token)
}

// ExtendTokenExpiry 延长令牌有效期
func (j *JWTManager) ExtendTokenExpiry(ctx context.Context, token string, duration time.Duration) error {
	// 获取当前令牌信息
	claims, err := j.ValidateToken(ctx, token)
	if err != nil {
		return err
	}

	// 更新过期时间
	claims.ExpiresAt = time.Now().Add(duration)

	// 重新存储
	return j.cache.Set(ctx, token, claims, duration)
}

// CleanupExpiredTokens 清理过期令牌（定时任务调用）
func (j *JWTManager) CleanupExpiredTokens(ctx context.Context) error {
	// 获取所有JWT令牌
	pattern := "jwt:*"
	keys, err := j.cache.Keys(ctx, pattern)
	if err != nil {
		return err
	}

	now := time.Now()
	cleanedCount := 0

	for _, key := range keys {
		var claims TokenClaims
		err := j.cache.GetStruct(ctx, key[4:], &claims) // 去掉前缀"jwt:"
		if err != nil {
			// 无法解析的令牌，直接删除
			j.cache.Delete(ctx, key[4:])
			cleanedCount++
			continue
		}

		// 检查是否过期
		if now.After(claims.ExpiresAt) {
			j.cache.Delete(ctx, key[4:])
			j.removeUserToken(ctx, claims.UserID, key[4:])
			cleanedCount++
		}
	}

	g.Log().Infof(ctx, "清理了 %d 个过期令牌", cleanedCount)
	return nil
}

// HasPermission 检查令牌是否拥有指定权限
func (tc *TokenClaims) HasPermission(permission types.Permission) bool {
	for _, p := range tc.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasRole 检查令牌是否拥有指定角色
func (tc *TokenClaims) HasRole(role types.RoleType) bool {
	for _, r := range tc.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// GetUserPermissions 从令牌获取用户权限信息
func (tc *TokenClaims) GetUserPermissions() *types.UserPermissions {
	return &types.UserPermissions{
		UserID:      tc.UserID,
		TenantID:    tc.TenantID,
		Roles:       tc.Roles,
		Permissions: tc.Permissions,
	}
}

// AddTokenToBlacklist 添加令牌到黑名单
func (j *JWTManager) AddTokenToBlacklist(ctx context.Context, token string, expiresAt time.Time) error {
	key := j.getBlacklistKey(token)

	// 计算黑名单过期时间（令牌原过期时间）
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		// 如果令牌已过期，设置短暂的黑名单时间以防重放攻击
		ttl = time.Hour
	}

	// 存储黑名单记录，值为撤销时间
	revokedAt := time.Now().Unix()
	return j.cache.Set(ctx, key, revokedAt, ttl)
}

// IsTokenBlacklisted 检查令牌是否在黑名单中
func (j *JWTManager) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	key := j.getBlacklistKey(token)
	exists, err := j.cache.Exists(ctx, key)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// RemoveTokenFromBlacklist 从黑名单移除令牌（管理员功能）
func (j *JWTManager) RemoveTokenFromBlacklist(ctx context.Context, token string) error {
	key := j.getBlacklistKey(token)
	return j.cache.Delete(ctx, key)
}

// GetBlacklistedTokenInfo 获取黑名单令牌信息
func (j *JWTManager) GetBlacklistedTokenInfo(ctx context.Context, token string) (revokedAt time.Time, err error) {
	key := j.getBlacklistKey(token)
	timestamp, err := j.cache.GetInt(ctx, key)
	if err != nil {
		return time.Time{}, err
	}

	if timestamp == 0 {
		return time.Time{}, fmt.Errorf("令牌不在黑名单中")
	}

	return time.Unix(int64(timestamp), 0), nil
}

// CleanupExpiredBlacklistTokens 清理过期的黑名单令牌
func (j *JWTManager) CleanupExpiredBlacklistTokens(ctx context.Context) error {
	pattern := j.getBlacklistKey("*")
	keys, err := j.cache.Keys(ctx, pattern)
	if err != nil {
		return err
	}

	cleanedCount := 0
	for _, key := range keys {
		// 检查TTL，如果已过期Redis会自动删除
		// 从完整键中提取实际的令牌部分
		tokenKey := key
		if len(key) > 10 { // "blacklist:" 前缀长度
			tokenKey = key[10:] // 去掉 "blacklist:" 前缀
		}

		ttl, err := j.cache.TTL(ctx, tokenKey)
		if err != nil {
			continue
		}

		if ttl == -2 { // -2表示key不存在（已过期）
			cleanedCount++
		}
	}

	g.Log().Infof(ctx, "黑名单清理完成，清理了 %d 个过期记录", cleanedCount)
	return nil
}

// getBlacklistKey 生成黑名单键
func (j *JWTManager) getBlacklistKey(token string) string {
	return fmt.Sprintf("blacklist:%s", token)
}

// RevokeAllUserTokens 撤销用户的所有令牌并加入黑名单
func (j *JWTManager) RevokeAllUserTokens(ctx context.Context, userID uint64) error {
	// 获取用户的所有令牌
	tokens, err := j.getUserTokens(ctx, userID)
	if err != nil {
		return fmt.Errorf("获取用户令牌失败: %v", err)
	}

	// 逐个撤销令牌
	for _, token := range tokens {
		err = j.RevokeToken(ctx, token)
		if err != nil {
			g.Log().Warningf(ctx, "撤销用户令牌失败: %v", err)
		}
	}

	// 清空用户令牌列表
	return j.clearUserTokens(ctx, userID)
}
