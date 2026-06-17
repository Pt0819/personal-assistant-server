package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRefreshToken(t *testing.T) {
	token, err := GenerateRefreshToken()
	require.NoError(t, err)
	assert.Len(t, token, 64, "refresh token should be 64 hex chars (32 bytes)")
}

func TestGenerateRefreshTokenUniqueness(t *testing.T) {
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, err := GenerateRefreshToken()
		require.NoError(t, err)
		assert.False(t, tokens[token], "tokens should be unique")
		tokens[token] = true
	}
}

func TestHashRefreshToken(t *testing.T) {
	raw := "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6a7b8c9d0e1f2a3b4c5d6a7b8c9d0e1f2"
	hash1 := HashRefreshToken(raw)
	hash2 := HashRefreshToken(raw)
	assert.Equal(t, hash1, hash2, "same input should produce same hash")
	assert.Len(t, hash1, 64, "SHA-256 hex output should be 64 chars")
}

func TestHashRefreshTokenDifferent(t *testing.T) {
	hash1 := HashRefreshToken("token1")
	hash2 := HashRefreshToken("token2")
	assert.NotEqual(t, hash1, hash2, "different inputs should produce different hashes")
}
