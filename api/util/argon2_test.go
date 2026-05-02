package util

import (
	"example_api/config"
	"os"
	"strings"
	"testing"
)

// TestArgon2SuccessPath validates that the hashing and comparison works
// correctly with the expected environment variable.
func TestArgon2SuccessPath(t *testing.T) {
	// Setup environment
	originalPepper := os.Getenv(config.ENV_PEPPER)
	testPepper := "test-internal-pepper-value"
	os.Setenv(config.ENV_PEPPER, testPepper)

	// Restore environment after test
	defer os.Setenv(config.ENV_PEPPER, originalPepper)

	secret := "user-password-123"

	// 1. Test Hashing
	hash, err := Argon2HashSecret(secret)
	if err != nil {
		t.Fatalf("Argon2HashSecret failed: %v", err)
	}

	if !strings.HasPrefix(hash, "$argon2d$") {
		t.Errorf("Expected argon2d prefix, got: %s", hash)
	}

	// 2. Test Successful Comparison
	match, err := Argon2CompareSecret(secret, hash)
	if err != nil {
		t.Fatalf("Argon2CompareSecret returned error: %v", err)
	}
	if !match {
		t.Error("Valid secret failed to match stored hash")
	}
}

// TestArgon2FailureCases validates that the system correctly rejects
// incorrect passwords, incorrect peppers, and malformed hashes.
func TestArgon2FailureCases(t *testing.T) {
	testPepper := "security-test-pepper"
	os.Setenv(config.ENV_PEPPER, testPepper)
	secret := "secure-password"

	hash, _ := Argon2HashSecret(secret)

	// 1. Test Wrong Secret
	match, _ := Argon2CompareSecret("not-the-right-password", hash)
	if match {
		t.Error("Security breach: matched with incorrect password")
	}

	// 2. Test Wrong Pepper (Changing the environment variable)
	os.Setenv(config.ENV_PEPPER, "different-pepper-value")
	match, _ = Argon2CompareSecret(secret, hash)
	if match {
		t.Error("Security breach: matched with incorrect pepper")
	}
	// Reset pepper for next sub-test
	os.Setenv(config.ENV_PEPPER, testPepper)

	// 3. Test Malformed Hash String
	_, err := Argon2CompareSecret(secret, "invalid:hash:format")
	if err == nil {
		t.Error("Expected error for malformed hash format, got nil")
	}

	// 4. Test Tampered Hash Content
	// Replacing a character in the Base64 hash portion
	tampered := hash[:len(hash)-5] + "A" + hash[len(hash)-4:]

	if tampered == hash {
		tampered = hash[:len(hash)-5] + "B" + hash[len(hash)-4:]
	}

	match, _ = Argon2CompareSecret(secret, tampered)
	if match {
		t.Error("Security breach: matched with tampered hash bytes")
	}
}
