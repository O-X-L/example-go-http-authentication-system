package account

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example_api/base/db"
	"example_api/util"

	"github.com/go-playground/validator/v10"
)

func TestLoginRequestValidation(t *testing.T) {
	v := validator.New()

	tests := []struct {
		name    string
		payload LoginRequest
		wantErr bool
	}{
		{"Valid Request", LoginRequest{Email: "test@example.com", Password: "password123"}, false},
		{"Invalid Email", LoginRequest{Email: "not-an-email", Password: "password123"}, true},
		{"Missing Password", LoginRequest{Email: "test@example.com", Password: ""}, true},
		{"Missing Email", LoginRequest{Email: "", Password: "password123"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Struct(tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoginRequest validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHandleLogin_InvalidJSON(t *testing.T) {
	v := validator.New()
	handler := HandleLogin(nil, v) // dbConn nil for this specific test

	req, _ := http.NewRequest("POST", "/a/session/login/basic", bytes.NewBuffer([]byte(`{invalid-json}`)))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestHandleLogin_Success(t *testing.T) {
	email := "test@example.com"
	password := "securepassword123"
	userID := 42

	hash, _ := util.Argon2HashSecret(password)

	mockAuth := &db.TestMockAuthStore{
		GetUserByEmailFunc: func(e string, d bool) (db.AuthUser, error) {
			if e == email {
				us := db.AuthUser{ID: userID, PasswordHash: string(hash), AuthType: db.AUTH_TYPE_BASIC, IsActive: true}
				return us, nil
			}
			return db.AuthUser{}, sql.ErrNoRows
		},
		AddSessionFunc: func(session db.AuthSession) error {
			return nil
		},
	}

	store := &db.DataStore{Auth: mockAuth}
	v := validator.New()
	handler := HandleLogin(store, v)

	body := LoginRequest{Email: email, Password: password}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/session/login/basic", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestHandleLogin_InvalidCredentials_UserNotFound(t *testing.T) {
	mockAuth := &db.TestMockAuthStore{
		GetUserByEmailFunc: func(email string, d bool) (db.AuthUser, error) {
			return db.AuthUser{}, sql.ErrNoRows
		},
	}

	store := &db.DataStore{Auth: mockAuth}
	v := validator.New()
	handler := HandleLogin(store, v)

	body := LoginRequest{Email: "unknown@example.com", Password: "anypassword"}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/session/login/basic", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("expected status %v for missing user, got %v", http.StatusUnauthorized, status)
	}
}

func TestHandleLogin_InactiveUser(t *testing.T) {
	email := "inactive@example.com"
	password := "securepassword123"
	userID := 43

	hash, _ := util.Argon2HashSecret(password)

	mockAuth := &db.TestMockAuthStore{
		GetUserByEmailFunc: func(e string, d bool) (db.AuthUser, error) {
			us := db.AuthUser{ID: userID, PasswordHash: string(hash), AuthType: db.AUTH_TYPE_BASIC, IsActive: false}
			return us, nil
		},
	}

	store := &db.DataStore{Auth: mockAuth}
	v := validator.New()
	handler := HandleLogin(store, v)

	body := LoginRequest{Email: email, Password: password}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/session/login/basic", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("expected status %v for inactive user, got %v", http.StatusForbidden, status)
	}
}

func TestHandleLogin_DatabaseErrorOnSessionInsert(t *testing.T) {
	email := "test@example.com"
	hash, _ := util.Argon2HashSecret("securepassword123")

	mockAuth := &db.TestMockAuthStore{
		GetUserByEmailFunc: func(e string, d bool) (db.AuthUser, error) {
			us := db.AuthUser{ID: 1, PasswordHash: string(hash), AuthType: db.AUTH_TYPE_BASIC, IsActive: true}
			return us, nil
		},
		AddSessionFunc: func(session db.AuthSession) error {
			return sql.ErrConnDone // Simulate DB error
		},
	}

	store := &db.DataStore{Auth: mockAuth}
	v := validator.New()
	handler := HandleLogin(store, v)

	body := LoginRequest{Email: email, Password: "securepassword123"}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/session/login/basic", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("expected status %v for failed session insert, got %v", http.StatusInternalServerError, status)
	}
}
