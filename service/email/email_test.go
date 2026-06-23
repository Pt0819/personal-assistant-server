package email

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

func TestMockEmailService_SendVerificationCode(t *testing.T) {
	svc := NewMockEmailService()
	err := svc.SendVerificationCode(t.Context(), "test@example.com", "123456")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestMockEmailService_SendVerificationCode_EmptyEmail(t *testing.T) {
	svc := NewMockEmailService()
	err := svc.SendVerificationCode(t.Context(), "", "123456")
	if err == nil {
		t.Fatal("expected error for empty email")
	}
}
