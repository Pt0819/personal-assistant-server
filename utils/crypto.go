package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

var (
	ErrDecryptionFailed = errors.New("解密失败")
)

// EncryptAES256GCM encrypts plaintext using AES-256-GCM.
// The returned ciphertext is: nonce (12 bytes) || encrypted_data || tag (16 bytes)
func EncryptAES256GCM(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aead.NonceSize()) // 12 bytes
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return aead.Seal(nonce, nonce, plaintext, nil), nil
}

// DecryptAES256GCM decrypts data produced by EncryptAES256GCM.
func DecryptAES256GCM(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrDecryptionFailed
	}
	nonce, encrypted := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aead.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}
	return plaintext, nil
}
