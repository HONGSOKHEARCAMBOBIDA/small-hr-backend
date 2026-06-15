package helper

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func HashQrtoken(phone string) string {
	secret := []byte(">Z71.i*?YGN.^:]Hv!+dSJ;JfR$TS:nW{1Uk^o6^tjGc")

	h := hmac.New(sha256.New, secret)
	h.Write([]byte(phone))

	return hex.EncodeToString(h.Sum(nil))
}
