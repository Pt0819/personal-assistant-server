package push

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"personal-assistant-server/global"
	"personal-assistant-server/model"
)

type PushService struct{}

// Subscribe 订阅消息推送
func (s *PushService) Subscribe(ctx context.Context, userID uint, openID, templateID string) (*model.PushSubscription, error) {
	// 检查是否已订阅
	var existing model.PushSubscription
	err := global.GVA_DB.WithContext(ctx).Where("user_id = ? AND template_id = ?", userID, templateID).First(&existing).Error
	if err == nil {
		// 已存在，更新为启用
		global.GVA_DB.Model(&existing).Updates(map[string]interface{}{
			"is_active":  true,
			"openid":     openID,
			"updated_at": time.Now(),
		})
		return &existing, nil
	}

	sub := &model.PushSubscription{
		UserID:     userID,
		OpenID:     openID,
		TemplateID: templateID,
		IsActive:   true,
	}
	if err := global.GVA_DB.WithContext(ctx).Create(sub).Error; err != nil {
		return nil, err
	}
	return sub, nil
}

// Unsubscribe 取消订阅
func (s *PushService) Unsubscribe(ctx context.Context, userID uint, templateID string) error {
	return global.GVA_DB.WithContext(ctx).
		Model(&model.PushSubscription{}).
		Where("user_id = ? AND template_id = ?", userID, templateID).
		Update("is_active", false).Error
}

// GetSubscriptions 获取用户订阅列表
func (s *PushService) GetSubscriptions(ctx context.Context, userID uint) ([]model.PushSubscription, error) {
	var subs []model.PushSubscription
	err := global.GVA_DB.WithContext(ctx).Where("user_id = ? AND is_active = ?", userID, true).Find(&subs).Error
	return subs, err
}

// ScanAndPushReminders 扫描待提醒日程并推送（由定时任务调用）
func (s *PushService) ScanAndPushReminders() error {
	ctx := context.Background()

	// 查询未来35分钟内开始且未推送的活跃日程
	var schedules []model.Schedule
	err := global.GVA_DB.WithContext(ctx).
		Where("status = 'active' AND reminded = ? AND start_time > ? AND start_time <= ?",
			false,
			time.Now(),
			time.Now().Add(35*time.Minute)).
		Find(&schedules).Error
	if err != nil {
		global.GVA_LOG.Error("扫描推送提醒失败: " + err.Error())
		return err
	}

	for _, schedule := range schedules {
		if err := s.sendReminder(ctx, &schedule); err != nil {
			global.GVA_LOG.Error(fmt.Sprintf("推送日程[%d]提醒失败: %s", schedule.ID, err.Error()))
			continue
		}
		// 标记已推送
		global.GVA_DB.Model(&schedule).Update("reminded", true)
	}

	return nil
}

// sendReminder 发送单个日程提醒
func (s *PushService) sendReminder(ctx context.Context, schedule *model.Schedule) error {
	// 查询用户订阅
	var subs []model.PushSubscription
	if err := global.GVA_DB.WithContext(ctx).Where("user_id = ? AND is_active = ?", schedule.UserID, true).Find(&subs).Error; err != nil {
		return err
	}

	if len(subs) == 0 {
		return nil // 用户未订阅推送
	}

	// 获取微信 access_token
	accessToken, err := s.getAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("获取access_token失败: %w", err)
	}

	for _, sub := range subs {
		if err := s.sendWechatMessage(accessToken, sub.OpenID, sub.TemplateID, schedule); err != nil {
			global.GVA_LOG.Error(fmt.Sprintf("发送微信消息失败: %s", err.Error()))
		}
	}

	return nil
}

// sendWechatMessage 调用微信订阅消息API
func (s *PushService) sendWechatMessage(accessToken, openID, templateID string, schedule *model.Schedule) error {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token=%s", accessToken)

	body := map[string]interface{}{
		"touser":      openID,
		"template_id": templateID,
		"page":        "/pages/index/index",
		"data": map[string]map[string]string{
			"thing1": {"value": schedule.Title},
			"time2":  {"value": schedule.StartTime.Format("2006-01-02 15:04")},
			"thing3": {"value": schedule.Location},
		},
	}

	jsonBody, _ := json.Marshal(body)
	resp, err := http.Post(url, "application/json", io.NopCloser(http.NoBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_ = jsonBody // use in real request
	// TODO: Replace with actual HTTP POST with jsonBody
	return nil
}

// getAccessToken 获取微信 access_token（带 Redis 缓存）
func (s *PushService) getAccessToken(ctx context.Context) (string, error) {
	// 先尝试从 Redis 获取
	if global.GVA_REDIS != nil {
		token, err := global.GVA_REDIS.Get(ctx, "wechat:access_token").Result()
		if err == nil && token != "" {
			return token, nil
		}
	}

	// 从微信API获取
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

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", err
	}

	if result.ErrCode != 0 {
		return "", fmt.Errorf("获取access_token失败: %d - %s", result.ErrCode, result.ErrMsg)
	}

	// 缓存到 Redis（提前5分钟过期）
	if global.GVA_REDIS != nil {
		ttl := time.Duration(result.ExpiresIn-300) * time.Second
		if ttl > 0 {
			global.GVA_REDIS.Set(ctx, "wechat:access_token", result.AccessToken, ttl)
		}
	}

	return result.AccessToken, nil
}
