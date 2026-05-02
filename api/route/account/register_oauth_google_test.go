package account

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"example_api/base/db"
	"example_api/types"

	"github.com/go-playground/validator/v10"
)

func TestHandleRegisterOAuthGoogle_Success(t *testing.T) {
	// Mock the Google ID token verification
	originalVerifyFunc := VerifyGoogleIDTokenFunc
	defer func() { VerifyGoogleIDTokenFunc = originalVerifyFunc }()
	VerifyGoogleIDTokenFunc = func(token string) (string, error) {
		if token == "valid-google-token" {
			return "user@gmail.com", nil
		}
		return "", fmt.Errorf("invalid token")
	}

	mockAuth := &db.TestMockAuthStore{
		AddUserOAuthFunc: func(email string, authType int) error {
			if email != "user@gmail.com" || authType != db.AUTH_TYPE_OAUTH_GOOGLE {
				t.Errorf("Unexpected values passed to DB: email=%s, authType=%d", email, authType)
			}
			return nil
		},
	}

	store := &db.DataStore{Auth: mockAuth}
	v := validator.New()
	handler := HandleRegisterOAuthGoogle(store, v)

	body := OAuthGoogleRequest{IDToken: "valid-google-token"}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/a/register/oauth/google", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("expected status %v, got %v", http.StatusCreated, status)
	}

	var response types.MessageResponse
	json.NewDecoder(rr.Body).Decode(&response)
	if response.Message != "User registered successfully using Google-OAuth" {
		t.Errorf("unexpected response message: %s", response.Message)
	}
}

func TestHandleRegisterOAuthGoogle_Conflict(t *testing.T) {
	originalVerifyFunc := VerifyGoogleIDTokenFunc
	defer func() { VerifyGoogleIDTokenFunc = originalVerifyFunc }()
	VerifyGoogleIDTokenFunc = func(token string) (string, error) {
		return "existinguser@gmail.com", nil
	}

	mockAuth := &db.TestMockAuthStore{
		AddUserOAuthFunc: func(email string, authType int) error {
			return sql.ErrNoRows // Simulate DB uniqueness constraint failure
		},
	}

	store := &db.DataStore{Auth: mockAuth}
	v := validator.New()
	handler := HandleRegisterOAuthGoogle(store, v)

	body := OAuthGoogleRequest{IDToken: "valid-google-token"}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/a/register/oauth/google", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusConflict {
		t.Errorf("expected status %v for existing user, got %v", http.StatusConflict, status)
	}
}
