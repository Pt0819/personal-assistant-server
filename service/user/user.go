package user

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"personal-assistant-server/global"
	"personal-assistant-server/model"
	"personal-assistant-server/utils/avatar"
)

type UserService struct{}

func (s *UserService) GetProfile(ctx context.Context, userID uint) (*model.User, error) {
	var user model.User
	if err := global.GVA_DB.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

type UpdateProfileRequest struct {
	Nickname               string                `form:"nickname"`
	Avatar                 *multipart.FileHeader `form:"avatar"`
	DefaultReminderMinutes *int                  `form:"default_reminder_minutes"`
	OnboardingCompleted    *bool                 `form:"onboarding_completed"`
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uint, req *UpdateProfileRequest) (*model.User, error) {
	var user model.User
	if err := global.GVA_DB.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	needRegenAvatar := false

	if req.Nickname != "" && req.Nickname != user.Nickname {
		updates["nickname"] = req.Nickname
		user.Nickname = req.Nickname
		if req.Avatar == nil {
			needRegenAvatar = true
		}
	}

	if req.Avatar != nil {
		avatarURL, err := s.uploadCustomAvatar(ctx, userID, req.Avatar)
		if err != nil {
			return nil, fmt.Errorf("头像上传失败: %w", err)
		}
		updates["avatar_url"] = avatarURL
		needRegenAvatar = false
	}

	if needRegenAvatar {
		avatarURL, err := s.regenerateAvatar(userID, user.Nickname)
		if err != nil {
			global.GVA_LOG.Error("重新生成头像失败: " + err.Error())
		} else if avatarURL != "" {
			updates["avatar_url"] = avatarURL
		}
	}

	if req.DefaultReminderMinutes != nil {
		updates["default_reminder_minutes"] = *req.DefaultReminderMinutes
	}
	if req.OnboardingCompleted != nil {
		updates["onboarding_completed"] = *req.OnboardingCompleted
	}

	if len(updates) > 0 {
		if err := global.GVA_DB.WithContext(ctx).Model(&user).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	global.GVA_DB.WithContext(ctx).First(&user, userID)
	return &user, nil
}

func (s *UserService) uploadCustomAvatar(ctx context.Context, userID uint, fileHeader *multipart.FileHeader) (string, error) {
	if global.GVA_STORAGE == nil {
		return "", fmt.Errorf("OSS存储未初始化")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	ext := ".png"
	if fileHeader.Header.Get("Content-Type") == "image/jpeg" {
		ext = ".jpg"
	}

	key := fmt.Sprintf("%d_%d%s", userID, time.Now().Unix(), ext)
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/png"
	}

	return global.GVA_STORAGE.Upload(ctx, key, file, contentType)
}

func (s *UserService) regenerateAvatar(userID uint, nickname string) (string, error) {
	pngBytes, err := avatar.Generate(userID, nickname)
	if err != nil {
		return "", err
	}

	if global.GVA_STORAGE == nil {
		return "", nil
	}

	key := fmt.Sprintf("%d_%d.png", userID, time.Now().Unix())
	return global.GVA_STORAGE.Upload(context.Background(), key, bytes.NewReader(pngBytes), "image/png")
}

// ==================== Task 11: Phone Binding ====================

// BindPhoneRequest 绑定手机号请求
type BindPhoneRequest struct {
	Code string `json:"code"`
}

// BindPhoneResponse 绑定手机号响应
type BindPhoneResponse struct {
	Phone string `json:"phone"`
}

// BindPhone 绑定/更新手机号
func (s *UserService) BindPhone(ctx context.Context, userID uint, req *BindPhoneRequest) (*BindPhoneResponse, error) {
	if req.Code == "" {
		return nil, fmt.Errorf("code不能为空")
	}

	// 获取微信 server access_token
	accessToken, err := getWechatAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("微信接口调用失败,请稍后重试")
	}

	// 调用 getPhoneNumber API
	phoneNumber, err := getPhoneNumber(ctx, accessToken, req.Code)
	if err != nil {
		return nil, fmt.Errorf("手机号code已过期或已被使用")
	}

	// 更新 users.phone
	global.GVA_DB.Model(&model.User{}).Where("id = ?", userID).
		Update("phone", phoneNumber)

	masked := maskPhone(phoneNumber)
	return &BindPhoneResponse{Phone: masked}, nil
}

func getWechatAccessToken(ctx context.Context) (string, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		global.GVA_CONFIG.Wechat.AppID,
		global.GVA_CONFIG.Wechat.AppSecret,
	)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result struct {
		AccessToken string `json:"access_token"`
		ErrCode     int    `json:"errcode"`
	}
	json.Unmarshal(body, &result)
	if result.ErrCode != 0 {
		return "", fmt.Errorf("微信返回错误: %d", result.ErrCode)
	}
	return result.AccessToken, nil
}

func getPhoneNumber(ctx context.Context, accessToken, code string) (string, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=%s",
		accessToken,
	)
	reqBody, _ := json.Marshal(map[string]string{"code": code})
	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		ErrCode   int `json:"errcode"`
		PhoneInfo struct {
			PurePhoneNumber string `json:"purePhoneNumber"`
		} `json:"phone_info"`
	}
	json.Unmarshal(respBody, &result)
	if result.ErrCode != 0 {
		return "", fmt.Errorf("微信返回错误: %d", result.ErrCode)
	}
	return result.PhoneInfo.PurePhoneNumber, nil
}

func maskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

// ==================== Task 12: Session Management ====================

// SessionInfo 设备会话信息
type SessionInfo struct {
	DeviceID   string    `json:"device_id"`
	DeviceInfo string    `json:"device_info"`
	IsCurrent  bool      `json:"is_current"`
	LastUsedAt time.Time `json:"last_used_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// SessionsResponse 设备列表响应
type SessionsResponse struct {
	Sessions     []SessionInfo `json:"sessions"`
	MaxDevices   int           `json:"max_devices"`
	CurrentCount int           `json:"current_count"`
}

// GetSessions 获取用户的所有活跃会话
func (s *UserService) GetSessions(ctx context.Context, userID uint, currentDeviceID string) (*SessionsResponse, error) {
	var sessions []model.UserSession
	global.GVA_DB.Where("user_id = ?", userID).
		Order("last_used_at DESC").
		Find(&sessions)

	result := make([]SessionInfo, 0, len(sessions))
	for _, sess := range sessions {
		result = append(result, SessionInfo{
			DeviceID:   sess.DeviceID,
			DeviceInfo: sess.DeviceInfo,
			IsCurrent:  sess.DeviceID == currentDeviceID,
			LastUsedAt: sess.LastUsedAt,
			CreatedAt:  sess.CreatedAt,
		})
	}

	maxDevices := global.GVA_CONFIG.Security.MaxDevices
	if maxDevices <= 0 {
		maxDevices = 3
	}

	return &SessionsResponse{
		Sessions:     result,
		MaxDevices:   maxDevices,
		CurrentCount: len(result),
	}, nil
}

// KickSession 踢出指定设备
func (s *UserService) KickSession(ctx context.Context, userID uint, targetDeviceID, currentDeviceID string) error {
	if targetDeviceID == currentDeviceID {
		return fmt.Errorf("不能踢出当前设备,请使用登出接口")
	}

	global.GVA_DB.Where("user_id = ? AND device_id = ?", userID, targetDeviceID).
		Delete(&model.UserSession{})
	return nil
}
