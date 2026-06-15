package user

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
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
