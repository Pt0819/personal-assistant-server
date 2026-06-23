package sms

import (
	"os"
	"testing"

	"personal-assistant-server/global"

	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	global.GVA_LOG = zap.NewNop()
	os.Exit(m.Run())
}

func TestMockSMSService_SendVerificationCode(t *testing.T) {
	svc := NewMockSMSService("123456")
	err := svc.SendVerificationCode(t.Context(), "13800138000", "123456")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestMockSMSService_SendVerificationCode_EmptyPhone(t *testing.T) {
	svc := NewMockSMSService("123456")
	err := svc.SendVerificationCode(t.Context(), "", "123456")
	if err == nil {
		t.Fatal("expected error for empty phone")
	}
}
