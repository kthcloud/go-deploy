package utils

import "crypto/sha256"

func HashString(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return string(h.Sum(nil))
}
