package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEmail(t *testing.T) {
	assert.True(t, isEmail("test@example.com"))
	assert.True(t, isEmail("a@b.co"))
	assert.False(t, isEmail("notanemail"))
	assert.False(t, isEmail("13800138000"))
}

func TestIsPhone(t *testing.T) {
	assert.True(t, isPhone("13800138000"))
	assert.True(t, isPhone("+8613800138000"))
	assert.False(t, isPhone("12345"))
	assert.False(t, isPhone("test@example.com"))
}

func TestGenerateVerificationCode(t *testing.T) {
	for i := 0; i < 20; i++ {
		code := generateVerificationCode()
		assert.Len(t, code, 6)
		for _, c := range code {
			assert.True(t, c >= '0' && c <= '9', "code contains non-digit: %s", code)
		}
	}
}
