package util

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func GenerateToken32() string {
	return GenerateToken(32)
}
