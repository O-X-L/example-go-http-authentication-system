package account_util

import (
	"testing"
	"time"

	"example_api/base/db"
)

func TestIsValidVerificationToken(t *testing.T) {
	validToken := "securetoken123"
	futureDate := time.Now().Add(1 * time.Hour)
	pastDate := time.Now().Add(-1 * time.Hour)

	tests := []struct {
		name        string
		inputToken  string
		dbToken     string
		expiresAt   time.Time
		expectedRes bool
	}{
		{"Valid Token", validToken, validToken, futureDate, true},
		{"Expired Token", validToken, validToken, pastDate, false},
		{"Mismatched Token", "wrongtoken", validToken, futureDate, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := IsValidVerificationToken(tt.inputToken, tt.dbToken, tt.expiresAt)
			if res != tt.expectedRes {
				t.Errorf("expected %v, got %v", tt.expectedRes, res)
			}
		})
	}
}

func TestAddVerificationToken(t *testing.T) {
	called := false
	mockAuth := &db.TestMockAuthStore{
		AddVerificationTokenFunc: func(userID, usageID int, tokenID, token string) error {
			called = true
			if userID != 123 || usageID != db.VERIFICATION_LINK_USAGE_VERIFY_EMAIL {
				t.Errorf("unexpected args")
			}
			return nil
		},
	}
	store := &db.DataStore{Auth: mockAuth}

	tokenID, token, err := AddVerificationToken(store, 123, db.VERIFICATION_LINK_USAGE_VERIFY_EMAIL)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !called {
		t.Errorf("expected AddVerificationToken to be called on store")
	}
	if tokenID == "" || token == "" {
		t.Errorf("expected tokens to be generated")
	}
}
