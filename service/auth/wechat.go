package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"personal-assistant-server/global"
	"personal-assistant-server/model"
	"personal-assistant-server/utils"
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

// LoginResponse 登录响应
type LoginResponse struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
}

// Login 微信小程序登录
func (s *AuthService) Login(ctx context.Context, code string) (*LoginResponse, error) {
	// 1. 调用微信 code2session 接口
	sessionResp, err := code2session(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("微信登录失败: %w", err)
	}

	// 2. 查找或创建用户
	user, err := s.findOrCreateUser(sessionResp)
	if err != nil {
		return nil, fmt.Errorf("用户处理失败: %w", err)
	}

	if user.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}

	// 3. 生成 JWT
	j := utils.NewJWT()
	claims := j.CreateClaims(user.ID, user.OpenID)
	token, err := j.CreateToken(claims)
	if err != nil {
		return nil, fmt.Errorf("生成token失败: %w", err)
	}

	return &LoginResponse{
		Token: token,
		User:  user,
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

// findOrCreateUser 查找用户，不存在则创建
func (s *AuthService) findOrCreateUser(session *WechatSessionResponse) (*model.User, error) {
	var user model.User
	err := global.GVA_DB.Where("openid = ?", session.OpenID).First(&user).Error
	if err == nil {
		// 用户已存在，更新 unionid（如果有）
		if session.UnionID != "" && user.UnionID == "" {
			global.GVA_DB.Model(&user).Update("unionid", session.UnionID)
		}
		return &user, nil
	}

	// 创建新用户
	user = model.User{
		OpenID:  session.OpenID,
		UnionID: session.UnionID,
	}
	if err := global.GVA_DB.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
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
