package utils

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestKey() []byte {
	key := make([]byte, 32)
	_, _ = rand.Read(key)
	return key
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := newTestKey()
	plaintext := []byte("wx_session_key_24_bytes")

	ciphertext, err := EncryptAES256GCM(plaintext, key)
	require.NoError(t, err)
	assert.Greater(t, len(ciphertext), len(plaintext), "ciphertext should include nonce+tag overhead")

	decrypted, err := DecryptAES256GCM(ciphertext, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestDecryptWithWrongKey(t *testing.T) {
	key1 := newTestKey()
	key2 := newTestKey()
	plaintext := []byte("secret data")

	ciphertext, err := EncryptAES256GCM(plaintext, key1)
	require.NoError(t, err)

	_, err = DecryptAES256GCM(ciphertext, key2)
	assert.ErrorIs(t, err, ErrDecryptionFailed)
}

func TestEncryptDecryptEmptyPlaintext(t *testing.T) {
	key := newTestKey()

	ciphertext, err := EncryptAES256GCM([]byte{}, key)
	require.NoError(t, err)

	decrypted, err := DecryptAES256GCM(ciphertext, key)
	require.NoError(t, err)
	assert.Empty(t, decrypted)
}

func TestEncryptDecryptLargePlaintext(t *testing.T) {
	key := newTestKey()
	plaintext := make([]byte, 2048)
	_, _ = rand.Read(plaintext)

	ciphertext, err := EncryptAES256GCM(plaintext, key)
	require.NoError(t, err)

	decrypted, err := DecryptAES256GCM(ciphertext, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestDecryptTamperedCiphertext(t *testing.T) {
	key := newTestKey()
	plaintext := []byte("original message")

	ciphertext, err := EncryptAES256GCM(plaintext, key)
	require.NoError(t, err)

	// Flip a byte in the ciphertext (after nonce)
	tampered := make([]byte, len(ciphertext))
	copy(tampered, ciphertext)
	tampered[len(tampered)-1] ^= 0xFF

	_, err = DecryptAES256GCM(tampered, key)
	assert.ErrorIs(t, err, ErrDecryptionFailed)
}

func TestDecryptTooShort(t *testing.T) {
	key := newTestKey()
	_, err := DecryptAES256GCM([]byte("short"), key)
	assert.ErrorIs(t, err, ErrDecryptionFailed)
}

func TestEncryptInvalidKeySize(t *testing.T) {
	_, err := EncryptAES256GCM([]byte("data"), []byte("not-32-bytes"))
	assert.Error(t, err)
}
