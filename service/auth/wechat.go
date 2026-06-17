package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"personal-assistant-server/global"
	"personal-assistant-server/model"
	"personal-assistant-server/utils"
	"personal-assistant-server/utils/avatar"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthService struct{}

// WechatSessionResponse 微信 code2session API 响应
type WechatSessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// LoginRequest 登录请求参数
type LoginRequest struct {
	Code       string `json:"code"`
	Nickname   string `json:"nickname"`
	AvatarURL  string `json:"avatar_url"`
	DeviceID   string `json:"device_id"`
	DeviceInfo string `json:"device_info"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int64       `json:"expires_in"`
	User         *model.User `json:"user"`
}

// Login 微信小程序登录（无昵称/头像的简单模式）
func (s *AuthService) Login(ctx context.Context, code string) (*LoginResponse, error) {
	return s.LoginWithProfile(ctx, LoginRequest{Code: code})
}

// LoginWithProfile 微信小程序登录（可携带昵称和头像URL）
func (s *AuthService) LoginWithProfile(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// 1. 调用微信 code2session 接口
	sessionResp, err := code2session(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("微信登录失败: %w", err)
	}

	// 2. 查找或创建用户
	user, isNew, err := s.findOrCreateUser(sessionResp)
	if err != nil {
		return nil, fmt.Errorf("用户处理失败: %w", err)
	}

	if user.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}

	// 3. 新用户：处理昵称和头像
	if isNew {
		nickname := req.Nickname
		if nickname == "" {
			nickname = "微信用户"
		}
		user.Nickname = nickname

		avatarURL, err := s.generateAndUploadAvatar(user.ID, nickname)
		if err != nil {
			global.GVA_LOG.Error("生成头像失败: " + err.Error())
		} else {
			user.AvatarURL = avatarURL
		}

		global.GVA_DB.Model(user).Updates(map[string]interface{}{
			"nickname":   user.Nickname,
			"avatar_url": user.AvatarURL,
		})
	}

	// 4. 加密 session_key
	var encryptedSessionKey []byte
	if len(global.GVA_ENCRYPTION_KEY) == 32 {
		encryptedSessionKey, err = utils.EncryptAES256GCM([]byte(sessionResp.SessionKey), global.GVA_ENCRYPTION_KEY)
		if err != nil {
			global.GVA_LOG.Error("加密session_key失败: " + err.Error())
		}
	}

	// 5. 处理 device_id
	deviceID := req.DeviceID
	if deviceID == "" {
		deviceID = fmt.Sprintf("unknown_%s", uuid.New().String()[:16])
	}

	// 6. 生成 refresh_token
	rawRefreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("生成refresh_token失败: %w", err)
	}
	refreshTokenHash := utils.HashRefreshToken(rawRefreshToken)

	// 7. 计算过期时间
	accessTTL, _ := utils.ParseDuration(global.GVA_CONFIG.JWT.ExpiresTime)
	refreshTTL, _ := utils.ParseDuration(global.GVA_CONFIG.JWT.RefreshExpiresTime)
	now := time.Now()
	accessExpiresAt := now.Add(accessTTL)
	refreshExpiresAt := now.Add(refreshTTL)

	// 8. 事务：多设备管理 + 创建/更新会话
	tx := global.GVA_DB.Begin()

	var currentCount int64
	tx.Model(&model.UserSession{}).Where("user_id = ?", user.ID).Count(&currentCount)

	maxDevices := global.GVA_CONFIG.Security.MaxDevices
	if maxDevices <= 0 {
		maxDevices = 3
	}
	if currentCount >= int64(maxDevices) {
		tx.Where("user_id = ?", user.ID).
			Order("last_used_at ASC").
			Limit(int(currentCount - int64(maxDevices) + 1)).
			Delete(&model.UserSession{})
	}

	session := model.UserSession{
		UserID:              user.ID,
		RefreshTokenHash:    refreshTokenHash,
		DeviceID:            deviceID,
		DeviceInfo:          req.DeviceInfo,
		SessionKeyEncrypted: encryptedSessionKey,
		AccessExpiresAt:     accessExpiresAt,
		RefreshExpiresAt:    refreshExpiresAt,
		LastUsedAt:          now,
	}
	if err := tx.Where("user_id = ? AND device_id = ?", user.ID, deviceID).
		Assign(session).
		FirstOrCreate(&session).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建会话失败: %w", err)
	}

	tx.Commit()

	// 9. 生成 JWT access_token
	j := utils.NewJWT()
	claims := j.CreateClaims(user.ID, user.OpenID, deviceID)
	accessToken, err := j.CreateToken(claims)
	if err != nil {
		return nil, fmt.Errorf("生成token失败: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		ExpiresIn:    int64(accessTTL.Seconds()),
		User:         user,
	}, nil
}

// GetUserProfile 获取用户信息
func (s *AuthService) GetUserProfile(ctx context.Context, userID uint) (*model.User, error) {
	var user model.User
	if err := global.GVA_DB.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// generateAndUploadAvatar 生成头像并上传到 OSS
func (s *AuthService) generateAndUploadAvatar(userID uint, nickname string) (string, error) {
	pngBytes, err := avatar.Generate(userID, nickname)
	if err != nil {
		return "", fmt.Errorf("生成头像图片失败: %w", err)
	}

	if global.GVA_STORAGE == nil {
		global.GVA_LOG.Warn("OSS存储未初始化，跳过头像上传")
		return "", nil
	}

	key := fmt.Sprintf("%d_%d.png", userID, time.Now().Unix())
	url, err := global.GVA_STORAGE.Upload(context.Background(), key, bytes.NewReader(pngBytes), "image/png")
	if err != nil {
		return "", fmt.Errorf("上传头像失败: %w", err)
	}
	return url, nil
}

// findOrCreateUser 查找用户，不存在则创建
// 返回值：user, isNew, error
func (s *AuthService) findOrCreateUser(session *WechatSessionResponse) (*model.User, bool, error) {
	var user model.User
	err := global.GVA_DB.Where("openid = ?", session.OpenID).First(&user).Error
	if err == nil {
		// 用户已存在，更新 unionid（如果有）
		if session.UnionID != "" && user.UnionID == "" {
			global.GVA_DB.Model(&user).Update("unionid", session.UnionID)
		}
		return &user, false, nil
	}

	// 创建新用户
	user = model.User{
		OpenID:  session.OpenID,
		UnionID: session.UnionID,
	}
	if err := global.GVA_DB.Create(&user).Error; err != nil {
		return nil, false, err
	}
	return &user, true, nil
}

// code2session 调用微信小程序登录接口
func code2session(ctx context.Context, code string) (*WechatSessionResponse, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		global.GVA_CONFIG.Wechat.AppID,
		global.GVA_CONFIG.Wechat.AppSecret,
		code,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求微信接口失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取微信响应失败: %w", err)
	}

	var sessionResp WechatSessionResponse
	if err := json.Unmarshal(body, &sessionResp); err != nil {
		return nil, fmt.Errorf("解析微信响应失败: %w", err)
	}

	if sessionResp.ErrCode != 0 {
		return nil, fmt.Errorf("微信返回错误: %d - %s", sessionResp.ErrCode, sessionResp.ErrMsg)
	}

	return &sessionResp, nil
}

// RefreshTokenRequest 刷新 token 请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshTokenResponse 刷新 token 响应
type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// RefreshToken 用 refresh token 换取新的 access token
func (s *AuthService) RefreshToken(ctx context.Context, req RefreshTokenRequest) (*RefreshTokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, errors.New("缺少refresh_token参数")
	}

	// 1. 计算 hash 并查找会话
	incomingHash := utils.HashRefreshToken(req.RefreshToken)
	var session model.UserSession
	if err := global.GVA_DB.Where("refresh_token_hash = ?", incomingHash).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("refresh_token无效或已登出")
		}
		return nil, fmt.Errorf("查询会话失败: %w", err)
	}

	// 2. 检查过期
	if time.Now().After(session.RefreshExpiresAt) {
		return nil, fmt.Errorf("refresh_token已过期,请重新登录")
	}

	// 3. 生成新 access_token
	var user model.User
	if err := global.GVA_DB.First(&user, session.UserID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	j := utils.NewJWT()
	claims := j.CreateClaims(user.ID, user.OpenID, session.DeviceID)
	accessToken, err := j.CreateToken(claims)
	if err != nil {
		return nil, fmt.Errorf("生成token失败: %w", err)
	}

	// 4. 更新会话
	accessTTL, _ := utils.ParseDuration(global.GVA_CONFIG.JWT.ExpiresTime)
	global.GVA_DB.Model(&session).Updates(map[string]interface{}{
		"access_expires_at": time.Now().Add(accessTTL),
		"last_used_at":      time.Now(),
	})

	return &RefreshTokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   int64(accessTTL.Seconds()),
	}, nil
}

// Logout 登出：黑名单 access token + 删除 refresh session
func (s *AuthService) Logout(ctx context.Context, userID uint, deviceID string, jti string) error {
	// 1. 黑名单 access token (存 jti)
	blacklist := model.JwtBlacklist{
		Jwt:       jti,
		ExpiresAt: time.Now().Add(2 * time.Hour),
	}
	global.GVA_DB.Create(&blacklist)

	// 2. 删除 user_session
	global.GVA_DB.Where("user_id = ? AND device_id = ?", userID, deviceID).
		Delete(&model.UserSession{})

	return nil
}
