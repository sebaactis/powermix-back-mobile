package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func HashToken(pepper []byte, raw string) string {
	mac := hmac.New(sha256.New, pepper)
	mac.Write([]byte(raw))
	return hex.EncodeToString(mac.Sum(nil))
}
