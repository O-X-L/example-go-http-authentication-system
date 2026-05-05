package account

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"example_api/base/db"
	"example_api/types"

	"github.com/go-playground/validator/v10"
)

func TestHandlePasswordResetRequest_Success(t *testing.T) {
	tokenGenerated := false
	mockAuth := &db.TestMockAuthStore{
		GetUserByEmailFunc: func(email string, details bool) (db.AuthUser, error) {
			if email == "test@example.com" {
				return db.AuthUser{ID: 1, IsActive: true, AuthType: db.AUTH_TYPE_BASIC, Email: email}, nil
			}
			return db.AuthUser{}, sql.ErrNoRows
		},
		AddVerificationTokenFunc: func(userID, usageID int, tokenID, token string) error {
			tokenGenerated = true
			if usageID != db.VERIFICATION_LINK_USAGE_PASSWORD_RESET {
				t.Errorf("expected usageID %d, got %d", db.VERIFICATION_LINK_USAGE_PASSWORD_RESET, usageID)
			}
			return nil
		},
	}
	store := &db.DataStore{Auth: mockAuth}
	v := validator.New()
	handler := HandlePasswordResetRequest(store, v)

	body := PasswordResetRequest{Email: "test@example.com"}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/a/password_reset/request", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status OK, got %v", status)
	}

	if !tokenGenerated {
		t.Errorf("expected verification token to be generated")
	}

	var response types.MessageResponse
	json.NewDecoder(rr.Body).Decode(&response)
	if response.Message == "" {
		t.Errorf("expected a response message")
	}
}

func TestHandlePasswordResetRequest_UserNotFound_Security(t *testing.T) {
	// Should return 200 OK even if user doesn't exist to prevent email enumeration
	mockAuth := &db.TestMockAuthStore{
		GetUserByEmailFunc: func(email string, details bool) (db.AuthUser, error) {
			return db.AuthUser{}, sql.ErrNoRows
		},
	}
	store := &db.DataStore{Auth: mockAuth}
	v := validator.New()
	handler := HandlePasswordResetRequest(store, v)

	body := PasswordResetRequest{Email: "unknown@example.com"}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/a/password_reset/request", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status OK for non-existent user, got %v", status)
	}
}

func TestHandlePasswordResetConfirm_Success(t *testing.T) {
	var passwordUpdated, tokenDeleted, sessionsCleared bool

	validToken := "validtoken123456789012345678901234567890123"
	mockAuth := &db.TestMockAuthStore{
		GetVerificationTokenFunc: func(tokenID string) (int, int, string, time.Time, error) {
			return 1, db.VERIFICATION_LINK_USAGE_PASSWORD_RESET, validToken, time.Now().Add(1 * time.Hour), nil
		},
		UpdateUserPasswordFunc: func(userID int, passwordHash string) error {
			passwordUpdated = true
			return nil
		},
		DeleteVerificationTokenFunc: func(tokenID string) error {
			tokenDeleted = true
			return nil
		},
		DeleteAllUserSessionsFunc: func(userID int) error {
			sessionsCleared = true
			return nil
		},
	}

	store := &db.DataStore{Auth: mockAuth}
	v := validator.New()
	handler := HandlePasswordResetConfirm(store, v)

	body := PasswordResetConfirmRequest{
		TokenID:     "validid123456789012345678901234567890123456",
		Token:       validToken,
		NewPassword: "new-strong-password",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/a/password_reset/confirm", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status OK, got %v", status)
	}
	if !passwordUpdated {
		t.Errorf("expected password to be updated")
	}
	if !tokenDeleted {
		t.Errorf("expected token to be deleted")
	}
	if !sessionsCleared {
		t.Errorf("expected user sessions to be cleared")
	}
}

func TestHandlePasswordResetConfirm_WrongUsage(t *testing.T) {
	// Trying to use an email verification token for a password reset
	validToken := "validtoken123456789012345678901234567890123"
	mockAuth := &db.TestMockAuthStore{
		GetVerificationTokenFunc: func(tokenID string) (int, int, string, time.Time, error) {
			return 1, db.VERIFICATION_LINK_USAGE_VERIFY_EMAIL, validToken, time.Now().Add(1 * time.Hour), nil
		},
	}

	store := &db.DataStore{Auth: mockAuth}
	v := validator.New()
	handler := HandlePasswordResetConfirm(store, v)

	body := PasswordResetConfirmRequest{
		TokenID:     "validid123456789012345678901234567890123456",
		Token:       validToken,
		NewPassword: "new-strong-password",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/a/password_reset/confirm", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("expected status Forbidden for wrong token usage, got %v", status)
	}
}
