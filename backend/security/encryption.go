package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"strings"
)

const cipherPrefix = "enc:"

// Encrypt seals plaintext with AES-256-GCM using $DATA_ENC_KEY (32 bytes, base64-encoded).
// Output: "enc:" + base64(nonce || sealed). The prefix lets Encrypt be idempotent
// and lets Decrypt detect legacy unencrypted rows from before this change landed.
func Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	if strings.HasPrefix(plaintext, cipherPrefix) {
		return plaintext, nil
	}
	key, err := loadKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	return cipherPrefix + base64.StdEncoding.EncodeToString(append(nonce, sealed...)), nil
}

// Decrypt reverses Encrypt. Values without the "enc:" prefix pass through
// unchanged so existing unencrypted rows keep working until migrated.
func Decrypt(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	if !strings.HasPrefix(value, cipherPrefix) {
		return value, nil
	}
	raw, err := base64.StdEncoding.DecodeString(value[len(cipherPrefix):])
	if err != nil {
		return "", err
	}
	key, err := loadKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}
	nonce, sealed := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func loadKey() ([]byte, error) {
	raw := os.Getenv("DATA_ENC_KEY")
	if raw == "" {
		return nil, errors.New("DATA_ENC_KEY is not set")
	}
	key, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}
	if len(key) != 32 {
		return nil, errors.New("DATA_ENC_KEY must decode to 32 bytes for AES-256")
	}
	return key, nil
}
