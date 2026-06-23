package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
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
	openID := ""
	if user.OpenID != nil {
		openID = *user.OpenID
	}
	claims := j.CreateClaims(user.ID, user.Username, openID, deviceID)
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
		if session.UnionID != "" && (user.UnionID == nil || *user.UnionID == "") {
			global.GVA_DB.Model(&user).Update("unionid", session.UnionID)
		}
		return &user, false, nil
	}

	// 创建新用户
	user = model.User{
		OpenID:  &session.OpenID,
		UnionID: &session.UnionID,
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
	openID := ""
	if user.OpenID != nil {
		openID = *user.OpenID
	}
	claims := j.CreateClaims(user.ID, user.Username, openID, session.DeviceID)
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

// WebQrcodeRequest 网页端获取二维码请求（空 body）
type WebQrcodeRequest struct{}

// WebQrcodeResponse 网页端获取二维码响应
type WebQrcodeResponse struct {
	QrcodeURL string `json:"qrcode_url"`
	TempToken string `json:"temp_token"`
	ExpiresIn int    `json:"expires_in"`
}

// WebStatusResponse 网页端轮询登录状态响应
type WebStatusResponse struct {
	Status      string      `json:"status"`
	AccessToken string      `json:"access_token,omitempty"`
	RefreshToken string     `json:"refresh_token,omitempty"`
	ExpiresIn   int64       `json:"expires_in,omitempty"`
	User        *model.User `json:"user,omitempty"`
	Nickname    string      `json:"nickname,omitempty"`
	AvatarURL   string      `json:"avatar_url,omitempty"`
}

// WebQrcode 生成网页端微信扫码登录二维码
func (s *AuthService) WebQrcode(ctx context.Context, req WebQrcodeRequest) (*WebQrcodeResponse, error) {
	// 1. 生成 state（防 CSRF）和 temp_token（轮询标识）
	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("生成state失败: %w", err)
	}
	tempToken, err := generateTempToken()
	if err != nil {
		return nil, fmt.Errorf("生成temp_token失败: %w", err)
	}

	// 2. 存储 state → temp_token 到 Redis（TTL 300s）
	if global.GVA_REDIS != nil {
		key := "wechat_oauth:" + state
		val := fmt.Sprintf(`{"temp_token":"%s","created_at":%d}`, tempToken, time.Now().Unix())
		if err := global.GVA_REDIS.Set(ctx, key, val, 300*time.Second).Err(); err != nil {
			return nil, fmt.Errorf("存储state失败: %w", err)
		}
	}

	// 3. 构建二维码 URL
	qrcodeURL := fmt.Sprintf(
		"https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s",
		global.GVA_CONFIG.Wechat.OpenPlatformAppID,
		global.GVA_CONFIG.Wechat.WebRedirectURI,
		state,
	)

	return &WebQrcodeResponse{
		QrcodeURL: qrcodeURL,
		TempToken: tempToken,
		ExpiresIn: 300,
	}, nil
}

// WebCallback 处理微信 OAuth 回调，验证 state，用 code 换取用户信息，生成 token，结果存入 Redis
// 返回前端重定向 URL
func (s *AuthService) WebCallback(ctx context.Context, code, state string) (string, error) {
	// 1. 从 Redis 验证 state
	if global.GVA_REDIS == nil {
		return "", fmt.Errorf("Redis未初始化")
	}
	key := "wechat_oauth:" + state
	val, err := global.GVA_REDIS.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("state无效或已过期")
	}
	var data struct {
		TempToken string `json:"temp_token"`
		CreatedAt int64  `json:"created_at"`
	}
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return "", fmt.Errorf("state数据损坏")
	}
	// 删除一次性 state
	global.GVA_REDIS.Del(ctx, key)

	// 2. 用 code 换取用户信息
	user, err := s.oauthAccessToken(ctx, code)
	if err != nil {
		return "", fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 3. 生成双 token
	deviceID := fmt.Sprintf("web_%s", uuid.New().String()[:16])
	rawRefreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return "", fmt.Errorf("生成refresh_token失败: %w", err)
	}
	refreshTokenHash := utils.HashRefreshToken(rawRefreshToken)

	accessTTL, _ := utils.ParseDuration(global.GVA_CONFIG.JWT.ExpiresTime)
	refreshTTL, _ := utils.ParseDuration(global.GVA_CONFIG.JWT.RefreshExpiresTime)
	now := time.Now()

	// 4. 创建会话（设备管理 — 单条插入，网页端不限制设备数）
	session := model.UserSession{
		UserID:           user.ID,
		RefreshTokenHash: refreshTokenHash,
		DeviceID:         deviceID,
		DeviceInfo:       "Web 网页端",
		AccessExpiresAt:  now.Add(accessTTL),
		RefreshExpiresAt: now.Add(refreshTTL),
		LastUsedAt:       now,
	}
	global.GVA_DB.Create(&session)

	// 5. 生成 JWT
	j := utils.NewJWT()
	openID := ""
	if user.OpenID != nil {
		openID = *user.OpenID
	}
	claims := j.CreateClaims(user.ID, user.Username, openID, deviceID)
	accessToken, err := j.CreateToken(claims)
	if err != nil {
		return "", fmt.Errorf("生成token失败: %w", err)
	}

	// 6. 存储 temp_token → 登录结果到 Redis（TTL 60s 供轮询）
	loginData := fmt.Sprintf(
		`{"status":"confirmed","access_token":"%s","refresh_token":"%s","expires_in":%d,"user":{"id":%d,"nickname":"%s","avatar_url":"%s","openid":"%s"}}`,
		accessToken, rawRefreshToken, int64(accessTTL.Seconds()),
		user.ID, user.Nickname, user.AvatarURL, openID,
	)
	global.GVA_REDIS.Set(ctx, "wechat_login:"+data.TempToken, loginData, 60*time.Second)

	// 7. 返回重定向 URL
	return fmt.Sprintf("/login?temp_token=%s", data.TempToken), nil
}

// WebStatus 轮询登录状态
func (s *AuthService) WebStatus(ctx context.Context, tempToken string) (*WebStatusResponse, error) {
	if tempToken == "" {
		return nil, fmt.Errorf("缺少temp_token参数")
	}
	if global.GVA_REDIS == nil {
		return &WebStatusResponse{Status: "pending"}, nil
	}
	val, err := global.GVA_REDIS.Get(ctx, "wechat_login:"+tempToken).Result()
	if err != nil {
		return &WebStatusResponse{Status: "pending"}, nil
	}
	var data struct {
		Status       string      `json:"status"`
		AccessToken  string      `json:"access_token"`
		RefreshToken string      `json:"refresh_token"`
		ExpiresIn    int64       `json:"expires_in"`
		User         *model.User `json:"user"`
	}
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return &WebStatusResponse{Status: "pending"}, nil
	}
	return &WebStatusResponse{
		Status:       data.Status,
		AccessToken:  data.AccessToken,
		RefreshToken: data.RefreshToken,
		ExpiresIn:    data.ExpiresIn,
		User:         data.User,
	}, nil
}

// generateState 生成随机 state（防 CSRF）
func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// generateTempToken 生成临时 token（轮询标识）
func generateTempToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// oauthAccessToken 用 code 换取微信用户信息（开放平台 OAuth）
func (s *AuthService) oauthAccessToken(ctx context.Context, code string) (*model.User, error) {
	// 1. 用 code 换取 access_token + openid
	tokenURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		global.GVA_CONFIG.Wechat.OpenPlatformAppID,
		global.GVA_CONFIG.Wechat.OpenPlatformAppSecret,
		code,
	)
	resp, err := http.Get(tokenURL)
	if err != nil {
		return nil, fmt.Errorf("请求微信接口失败: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		OpenID      string `json:"openid"`
		UnionID     string `json:"unionid"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("解析微信响应失败: %w", err)
	}
	if tokenResp.ErrCode != 0 {
		return nil, fmt.Errorf("微信返回错误: %d - %s", tokenResp.ErrCode, tokenResp.ErrMsg)
	}

	// 2. 获取用户信息（昵称、头像）
	userInfoURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s",
		tokenResp.AccessToken, tokenResp.OpenID,
	)
	uiResp, err := http.Get(userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}
	defer uiResp.Body.Close()
	uiBody, _ := io.ReadAll(uiResp.Body)
	var userInfo struct {
		OpenID     string `json:"openid"`
		Nickname   string `json:"nickname"`
		HeadImgURL string `json:"headimgurl"`
		UnionID    string `json:"unionid"`
		ErrCode    int    `json:"errcode"`
	}
	if err := json.Unmarshal(uiBody, &userInfo); err != nil {
		return nil, fmt.Errorf("解析用户信息失败: %w", err)
	}
	if userInfo.ErrCode != 0 {
		return nil, fmt.Errorf("获取用户信息错误: %d", userInfo.ErrCode)
	}

	// 3. 查找或创建用户
	var user model.User
	err = global.GVA_DB.Where("openid = ?", userInfo.OpenID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user = model.User{
				OpenID:    &userInfo.OpenID,
				UnionID:   &userInfo.UnionID,
				Nickname:  userInfo.Nickname,
				AvatarURL: userInfo.HeadImgURL,
			}
			if err := global.GVA_DB.Create(&user).Error; err != nil {
				return nil, fmt.Errorf("创建用户失败: %w", err)
			}
			return &user, nil
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	// 补充 unionid 和头像
	if userInfo.UnionID != "" && (user.UnionID == nil || *user.UnionID == "") {
		global.GVA_DB.Model(&user).Updates(map[string]interface{}{
			"unionid":    userInfo.UnionID,
			"avatar_url": userInfo.HeadImgURL,
		})
	}
	return &user, nil
}

// DevLogin 开发环境免扫码登录，仅 system.env == "local" 时可用
func (s *AuthService) DevLogin(ctx context.Context) (*LoginResponse, error) {
	if global.GVA_CONFIG.System.Env != "local" {
		return nil, fmt.Errorf("开发登录仅在 local 环境可用")
	}

	devOpenID := "dev_user_001"
	var user model.User
	err := global.GVA_DB.Where("openid = ?", devOpenID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user = model.User{
				OpenID:    &devOpenID,
				Nickname:  "Dev",
				AvatarURL: "",
			}
			if err := global.GVA_DB.Create(&user).Error; err != nil {
				return nil, fmt.Errorf("创建开发用户失败: %w", err)
			}
		} else {
			return nil, fmt.Errorf("查询开发用户失败: %w", err)
		}
	}

	return s.buildLoginResponse(&user, "dev_device_001", "Dev Browser")
}

// buildLoginResponse 构建双 token 登录响应
func (s *AuthService) buildLoginResponse(user *model.User, deviceID, deviceInfo string) (*LoginResponse, error) {
	rawRefreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("生成refresh_token失败: %w", err)
	}
	refreshTokenHash := utils.HashRefreshToken(rawRefreshToken)

	accessTTL, _ := utils.ParseDuration(global.GVA_CONFIG.JWT.ExpiresTime)
	refreshTTL, _ := utils.ParseDuration(global.GVA_CONFIG.JWT.RefreshExpiresTime)
	now := time.Now()

	session := model.UserSession{
		UserID:           user.ID,
		RefreshTokenHash: refreshTokenHash,
		DeviceID:         deviceID,
		DeviceInfo:       deviceInfo,
		AccessExpiresAt:  now.Add(accessTTL),
		RefreshExpiresAt: now.Add(refreshTTL),
		LastUsedAt:       now,
	}
	global.GVA_DB.Create(&session)

	j := utils.NewJWT()
	openID := ""
	if user.OpenID != nil {
		openID = *user.OpenID
	}
	claims := j.CreateClaims(user.ID, user.Username, openID, deviceID)
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
