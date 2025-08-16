package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/spume/mer-sys/backend/shared/cache"
	"github.com/spume/mer-sys/backend/shared/types"
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
	UserID    uint64    `json:"user_id"`
	TenantID  uint64    `json:"tenant_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	TokenType string    `json:"token_type"` // access, refresh
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// GenerateToken 生成JWT令牌
func (j *JWTManager) GenerateToken(ctx context.Context, user *types.User, tokenType string) (string, error) {
	now := time.Now()
	expireTime := j.expireTime
	if tokenType == "refresh" {
		expireTime = j.expireTime * 7 // 刷新令牌7倍时长
	}

	claims := &TokenClaims{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		Username:  user.Username,
		Email:     user.Email,
		TokenType: tokenType,
		IssuedAt:  now,
		ExpiresAt: now.Add(expireTime),
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
	var claims TokenClaims
	err := j.cache.GetStruct(ctx, token, &claims)
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

	// 创建新的访问令牌
	user := &types.User{
		ID:       claims.UserID,
		TenantID: claims.TenantID,
		Username: claims.Username,
		Email:    claims.Email,
	}

	return j.GenerateToken(ctx, user, "access")
}

// RevokeToken 撤销令牌
func (j *JWTManager) RevokeToken(ctx context.Context, token string) error {
	// 获取令牌信息
	claims, err := j.ValidateToken(ctx, token)
	if err != nil {
		return err
	}

	// 删除令牌
	err = j.cache.Delete(ctx, token)
	if err != nil {
		return err
	}

	// 从用户令牌列表中移除
	return j.removeUserToken(ctx, claims.UserID, token)
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
