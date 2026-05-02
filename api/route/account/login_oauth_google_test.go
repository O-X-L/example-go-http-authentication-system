package account

import (
	"bytes"
	"encoding/json"
	"example_api/base/db"
	"example_api/config"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestHandleLoginOAuthGoogle_Success(t *testing.T) {
	originalVerifyFunc := VerifyGoogleIDTokenFunc
	defer func() { VerifyGoogleIDTokenFunc = originalVerifyFunc }()
	VerifyGoogleIDTokenFunc = func(token string) (string, error) {
		return "user@gmail.com", nil
	}

	mockAuth := &db.TestMockAuthStore{
		GetUserByEmailFunc: func(email string, d bool) (db.AuthUser, error) {
			us := db.AuthUser{ID: 42, PasswordHash: "random-unguessable-hash", AuthType: db.AUTH_TYPE_OAUTH_GOOGLE, IsActive: true}
			return us, nil
		},
		AddSessionFunc: func(db.AuthSession) error {
			return nil
		},
	}

	store := &db.DataStore{Auth: mockAuth}
	v := validator.New()
	handler := HandleLoginOAuthGoogle(store, v)

	body := OAuthGoogleRequest{IDToken: "valid-google-token"}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/a/session/login/oauth/google", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, status)
	}

	// Verify session cookie was attached to response
	cookies := rr.Result().Cookies()
	hasSessionCookie := false
	for _, c := range cookies {
		if c.Name == config.COOKIE_SESSION {
			hasSessionCookie = true
			break
		}
	}
	if !hasSessionCookie {
		t.Errorf("expected session cookie %s to be set", config.COOKIE_SESSION)
	}
}

func TestHandleLoginOAuthGoogle_WrongAuthType(t *testing.T) {
	originalVerifyFunc := VerifyGoogleIDTokenFunc
	defer func() { VerifyGoogleIDTokenFunc = originalVerifyFunc }()
	VerifyGoogleIDTokenFunc = func(token string) (string, error) {
		return "user@gmail.com", nil
	}

	mockAuth := &db.TestMockAuthStore{
		GetUserByEmailFunc: func(email string, d bool) (db.AuthUser, error) {
			us := db.AuthUser{ID: 42, PasswordHash: "real-password-hash", AuthType: db.AUTH_TYPE_BASIC, IsActive: true}
			return us, nil
		},
	}

	store := &db.DataStore{Auth: mockAuth}
	v := validator.New()
	handler := HandleLoginOAuthGoogle(store, v)

	body := OAuthGoogleRequest{IDToken: "valid-google-token"}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/a/session/login/oauth/google", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("expected status %v for wrong auth type, got %v", http.StatusUnauthorized, status)
	}
}

func TestHandleLoginOAuthGoogle_Inactive(t *testing.T) {
	originalVerifyFunc := VerifyGoogleIDTokenFunc
	defer func() { VerifyGoogleIDTokenFunc = originalVerifyFunc }()
	VerifyGoogleIDTokenFunc = func(token string) (string, error) {
		return "user@gmail.com", nil
	}

	mockAuth := &db.TestMockAuthStore{
		GetUserByEmailFunc: func(email string, d bool) (db.AuthUser, error) {
			us := db.AuthUser{ID: 42, PasswordHash: "random-unguessable-hash", AuthType: db.AUTH_TYPE_OAUTH_GOOGLE, IsActive: false}
			return us, nil
		},
		AddSessionFunc: func(session db.AuthSession) error {
			return nil
		},
	}

	store := &db.DataStore{Auth: mockAuth}
	v := validator.New()
	handler := HandleLoginOAuthGoogle(store, v)

	body := OAuthGoogleRequest{IDToken: "valid-google-token"}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/a/session/login/oauth/google", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("expected status %v for inactive user, got %v", http.StatusForbidden, status)
	}
}
