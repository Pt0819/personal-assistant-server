package sms

import (
	"context"
	"errors"
	"fmt"

	"personal-assistant-server/global"
)

type SMSService interface {
	SendVerificationCode(ctx context.Context, phone, code string) error
}

type mockSMSService struct {
	fixedCode string
}

func NewMockSMSService(fixedCode string) SMSService {
	if fixedCode == "" {
		fixedCode = "123456"
	}
	return &mockSMSService{fixedCode: fixedCode}
}

func (s *mockSMSService) SendVerificationCode(ctx context.Context, phone, code string) error {
	if phone == "" {
		return errors.New("手机号不能为空")
	}
	global.GVA_LOG.Info(fmt.Sprintf("[Mock SMS] 验证码 %s 已发送到 %s", code, phone))
	return nil
}

// NewSMSService returns the SMS service based on config
func NewSMSService() SMSService {
	switch global.GVA_CONFIG.SMS.Provider {
	case "montnets":
		return NewMockSMSService(global.GVA_CONFIG.SMS.Mock.FixedCode) // Montnets implementation deferred
	default:
		return NewMockSMSService(global.GVA_CONFIG.SMS.Mock.FixedCode)
	}
}
