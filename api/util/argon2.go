package util

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"example_api/config"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	Memory      = 64 * 1024 // 64MB
	Iterations  = 3
	Parallelism = 2
	SaltLength  = 16
	KeyLength   = 32
)

// Argon2HashSecret generates an argon2 string-hash from a secret
func Argon2HashSecret(secret string) (string, error) {
	salt := make([]byte, SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	input := []byte(secret + config.GetServerPepper())

	hash := argon2.Key(input, salt, Iterations, Memory, Parallelism, KeyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf(
		"$argon2d$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, Memory, Iterations, Parallelism, b64Salt, b64Hash,
	)

	return encodedHash, nil
}

// Argon2CompareSecret checks if a secret matches an argon2 string-hash.
func Argon2CompareSecret(secret, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash format")
	}

	var mem uint32
	var iter uint32
	var par uint8
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &mem, &iter, &par)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	input := []byte(secret + config.GetServerPepper())
	comparisonHash := argon2.Key(input, salt, iter, mem, par, uint32(len(decodedHash)))

	return subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1, nil
}
