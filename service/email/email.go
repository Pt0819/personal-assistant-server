package email

import (
	"context"
	"errors"
	"fmt"

	"personal-assistant-server/global"
)

type EmailService interface {
	SendVerificationCode(ctx context.Context, to, code string) error
}

type mockEmailService struct{}

func NewMockEmailService() EmailService {
	return &mockEmailService{}
}

func (s *mockEmailService) SendVerificationCode(ctx context.Context, to, code string) error {
	if to == "" {
		return errors.New("收件人邮箱不能为空")
	}
	global.GVA_LOG.Info(fmt.Sprintf("[Mock Email] 验证码 %s 已发送到 %s", code, to))
	return nil
}

// NewEmailService returns the email service based on config
func NewEmailService() EmailService {
	switch global.GVA_CONFIG.Email.Provider {
	case "smtp":
		return NewMockEmailService() // SMTP implementation deferred
	default:
		return NewMockEmailService()
	}
}
