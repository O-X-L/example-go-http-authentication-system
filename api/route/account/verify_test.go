package account

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"

	"example_api/base/db"
	"example_api/config"
	"example_api/types"
)

func TestHandleVerify_InvalidPayload(t *testing.T) {
	v := validator.New()
	handler := HandleVerify(nil, v)

	req, _ := http.NewRequest("POST", "/a/verify", bytes.NewBuffer([]byte(`{invalid-json}`)))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("expected %v, got %v", http.StatusBadRequest, status)
	}
}

func TestHandleVerify_Success(t *testing.T) {
	mockAuth := &db.TestMockAuthStore{
		GetVerificationTokenFunc: func(tokenID string) (int, int, string, time.Time, error) {
			return 1, db.VERIFICATION_LINK_USAGE_VERIFY_EMAIL, "validtoken123456789012345678901234567890123", time.Now().Add(1 * time.Hour), nil
		},
		SetVerificationDoneUserEmailFunc: func(userID int) error {
			return nil
		},
		DeleteVerificationTokenFunc: func(tokenID string) error {
			return nil
		},
	}

	store := &db.DataStore{Auth: mockAuth}
	v := validator.New()
	handler := HandleVerify(store, v)

	body := VerifyRequest{
		TokenID: "validid123456789012345678901234567890123456",
		Token:   "validtoken123456789012345678901234567890123",
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/a/verify", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status OK, got %v", status)
	}

	var response types.MessageResponse
	json.NewDecoder(rr.Body).Decode(&response)
	if response.Message != "Verified successfully" {
		t.Errorf("unexpected response message: %s", response.Message)
	}
}

func TestHandleVerifyResend_Success(t *testing.T) {
	sessionID := "valid-session"
	userID := 123

	mockAuth := &db.TestMockAuthStore{
		GetUserBySessionIDFunc: func(sid string, details bool) (db.AuthUser, error) {
			if sid == sessionID {
				return db.AuthUser{ID: userID, Email: "test@example.com"}, nil
			}
			return db.AuthUser{}, sql.ErrNoRows
		},
		RenewVerificationTokenForResendFunc: func(uID, usageID int) (string, string, error) {
			if uID == userID && usageID == db.VERIFICATION_LINK_USAGE_VERIFY_EMAIL {
				return "new-token-id", "new-token", nil
			}
			return "", "", errors.New("unexpected args")
		},
	}

	store := &db.DataStore{Auth: mockAuth}
	v := validator.New()
	handler := HandleVerifyResend(store, v)

	body := VerifyResendRequest{Usage: "email"}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/a/verify_resend", bytes.NewBuffer(jsonBody))
	req.AddCookie(&http.Cookie{Name: config.COOKIE_SESSION, Value: sessionID})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status OK, got %v", status)
	}
	var response types.MessageResponse
	json.NewDecoder(rr.Body).Decode(&response)
	if response.Message != "Verified successfully" { // Note: Your handler returns this exact string right now
		t.Errorf("unexpected response message: %s", response.Message)
	}
}

func TestHandleVerifyResend_RateLimit(t *testing.T) {
	sessionID := "valid-session"

	mockAuth := &db.TestMockAuthStore{
		GetUserBySessionIDFunc: func(sid string, details bool) (db.AuthUser, error) {
			return db.AuthUser{ID: 123, Email: "test@example.com"}, nil
		},
		RenewVerificationTokenForResendFunc: func(uID, usageID int) (string, string, error) {
			return "", "", errors.New("resend limit exceeded")
		},
	}

	store := &db.DataStore{Auth: mockAuth}
	v := validator.New()
	handler := HandleVerifyResend(store, v)

	body := VerifyResendRequest{Usage: "email"}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/a/verify_resend", bytes.NewBuffer(jsonBody))
	req.AddCookie(&http.Cookie{Name: config.COOKIE_SESSION, Value: sessionID})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusTooManyRequests {
		t.Errorf("expected status InternalServerError, got %v", status)
	}
}

func TestHandleVerifyResend_Error(t *testing.T) {
	sessionID := "valid-session"

	mockAuth := &db.TestMockAuthStore{
		GetUserBySessionIDFunc: func(sid string, details bool) (db.AuthUser, error) {
			return db.AuthUser{ID: 123, Email: "test@example.com"}, nil
		},
		RenewVerificationTokenForResendFunc: func(uID, usageID int) (string, string, error) {
			return "", "", errors.New("random error")
		},
	}

	store := &db.DataStore{Auth: mockAuth}
	v := validator.New()
	handler := HandleVerifyResend(store, v)

	body := VerifyResendRequest{Usage: "email"}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/a/verify_resend", bytes.NewBuffer(jsonBody))
	req.AddCookie(&http.Cookie{Name: config.COOKIE_SESSION, Value: sessionID})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("expected status InternalServerError, got %v", status)
	}
}

func TestHandleVerify_PasswordReset(t *testing.T) {
	tokenDeleted := false

	mockAuth := &db.TestMockAuthStore{
		GetVerificationTokenFunc: func(tokenID string) (int, int, string, time.Time, error) {
			return 1, db.VERIFICATION_LINK_USAGE_PASSWORD_RESET, "validtoken123456789012345678901234567890123", time.Now().Add(1 * time.Hour), nil
		},
		DeleteVerificationTokenFunc: func(tokenID string) error {
			tokenDeleted = true
			return nil
		},
	}

	store := &db.DataStore{Auth: mockAuth}
	v := validator.New()
	handler := HandleVerify(store, v)

	body := VerifyRequest{
		TokenID: "validid123456789012345678901234567890123456",
		Token:   "validtoken123456789012345678901234567890123",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/a/verify", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status OK, got %v", status)
	}

	// For password resets, the token MUST NOT be deleted during the HandleVerify stage
	if tokenDeleted {
		t.Errorf("expected token NOT to be deleted for password reset pre-verification")
	}
}
