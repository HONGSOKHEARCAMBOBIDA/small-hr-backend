package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
)

func getKey(envVar string) ([]byte, error) {
	key := os.Getenv(envVar)
	if len(key) != 32 {
		return nil, fmt.Errorf("%s must be exactly 32 characters for AES-256", envVar)
	}
	return []byte(key), nil
}

func Encrypt(envVar, plainText string) (string, error) {
	key, err := getKey(envVar)
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

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := aesGCM.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func Decrypt(envVar, cipherTextB64 string) (string, error) {
	key, err := getKey(envVar)
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

	nonce, cipherText := cipherText[:nonceSize], cipherText[nonceSize:]

	plainText, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}

const (
	envBotToken = "ENCRYPTION_BOTTOKEN"
	envChatID   = "ENCRYPTION_CHATID"
)

func EncryptBotToken(v string) (string, error) { return Encrypt(envBotToken, v) }
func DecryptBotToken(v string) (string, error) { return Decrypt(envBotToken, v) }

func EncryptChatID(v string) (string, error) { return Encrypt(envChatID, v) }
func DecryptChatID(v string) (string, error) { return Decrypt(envChatID, v) }
