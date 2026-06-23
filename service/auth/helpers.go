package auth

import (
	"crypto/rand"
	"errors"
	"math/big"
	"strings"
	"unicode/utf8"

	"personal-assistant-server/global"
	"personal-assistant-server/model"

	"github.com/google/uuid"
)

// usernameChars excludes confusing characters i/l/o/0/1, leaving 28 chars
const usernameChars = "abcdefghjkmnpqrstuvwxyz23456789"

// GenerateUsername generates a random username, format "小助手_" + 6 random chars
func GenerateUsername() string {
	const suffixLen = 6
	suffix := make([]byte, suffixLen)
	for i := range suffixLen {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(usernameChars))))
		if err != nil {
			suffix[i] = usernameChars[i%len(usernameChars)]
			continue
		}
		suffix[i] = usernameChars[n.Int64()]
	}
	return "小助手_" + string(suffix)
}

// GenerateUniqueUsername generates a unique username, retrying on collision
func GenerateUniqueUsername() string {
	for i := 0; i < 5; i++ {
		name := GenerateUsername()
		var count int64
		global.GVA_DB.Model(&model.User{}).Where("username = ?", name).Count(&count)
		if count == 0 {
			return name
		}
	}
	// 5 collisions (extremely unlikely) — use UUID-based fallback
	return "u_" + uuid.New().String()[:8]
}

// ValidatePassword validates password strength: at least 8 chars, must contain both letters and digits
func ValidatePassword(password string) error {
	if utf8.RuneCountInString(password) < 8 {
		return errors.New("密码至少需要8个字符")
	}
	hasLetter := strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasDigit := strings.ContainsAny(password, "0123456789")
	if !hasLetter || !hasDigit {
		return errors.New("密码需同时包含字母和数字")
	}
	return nil
}
