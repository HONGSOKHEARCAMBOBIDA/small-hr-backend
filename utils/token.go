package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

func HashToken(token string) string {
	mac := hmac.New(sha256.New, Jwtkey)
	//បង្កើត HMAC object ថ្មី
	mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil))
}

func VerifyToken(storedHash, providedToken string) bool {
	computedHash := HashToken(providedToken)
	return subtle.ConstantTimeCompare([]byte(storedHash), []byte(computedHash)) == 1
}
