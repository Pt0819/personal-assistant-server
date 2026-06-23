package auth

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateUsername(t *testing.T) {
	name := GenerateUsername()
	assert.True(t, utf8.RuneCountInString(name) >= 9, "username too short: %s", name)
	assert.True(t, utf8.RuneCountInString(name) <= 12, "username too long: %s", name)
	assert.Contains(t, name, "小助手_")
}

func TestGenerateUsername_Uniqueness(t *testing.T) {
	names := make(map[string]bool)
	for i := 0; i < 100; i++ {
		name := GenerateUsername()
		require.False(t, names[name], "duplicate username: %s", name)
		names[name] = true
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  string
	}{
		{"valid", "abcd1234", ""},
		{"valid long", "abcdefghijklmn12345678", ""},
		{"too short", "a1", "密码至少需要8个字符"},
		{"only letters", "abcdefgh", "密码需同时包含字母和数字"},
		{"only digits", "12345678", "密码需同时包含字母和数字"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
			}
		})
	}
}
