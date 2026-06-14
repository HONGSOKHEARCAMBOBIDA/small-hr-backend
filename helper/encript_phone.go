package helper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

// aes AES encryption algorithm
// cipher GCM encryption mode
// rand Generate secure random bytes
// base64 Convert encrypted bytes to string
// io Read random data
// os Read environment variables
func getEncryptionPhone() ([]byte, error) {
	key := os.Getenv("ENCRYPTION_PHONE")
	if len(key) != 32 {
		return nil, errors.New("ENCRYPTION_PHONE must be exactly 32 characters for AES-256")
	}
	return []byte(key), nil
}

func EncryptPhone(plainText string) (string, error) {
	key, err := getEncryptionPhone()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Generate a random nonce
	nonce := make([]byte, aesGCM.NonceSize())
	// nonce លេខ Random ដែលប្រើតែម្តងសម្រាប់ Encryption មួយ
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt and prepend nonce to the ciphertext
	cipherText := aesGCM.Seal(nonce, nonce, []byte(plainText), nil)

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func DecryptPhone(cipherTextB64 string) (string, error) {
	key, err := getEncryptionPhone()
	if err != nil {
		return "", err
	}

	cipherText, err := base64.StdEncoding.DecodeString(cipherTextB64)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(cipherText) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	// Split nonce and actual ciphertext
	nonce, cipherText := cipherText[:nonceSize], cipherText[nonceSize:]

	plainText, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}
