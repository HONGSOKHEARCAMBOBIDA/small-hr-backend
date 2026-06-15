package helper

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func HashPhone(phone string) string {
	secret := []byte("h;H.3}.9DYA>+iZ^R!Ifn),v*:G4AUzI>+Z.5aa")

	h := hmac.New(sha256.New, secret)
	h.Write([]byte(phone))

	return hex.EncodeToString(h.Sum(nil))
}
