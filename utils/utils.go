package utils

import (
	"crypto/sha256"
	"encoding/base64"
)

func HashString(token string) string {
	hasher := sha256.New()
	hasher.Write([]byte(token))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}
