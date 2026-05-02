package account

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example_api/base/db"
	"example_api/config"
)

func TestHandleStatus_Success(t *testing.T) {
	sessionID := "valid-session-id"
	mockAuth := &db.TestMockAuthStore{
		GetUserBySessionIDFunc: func(sid string, d bool) (db.AuthUser, error) {
			if sid == sessionID {
				return db.AuthUser{ID: 123, Email: "test@example.com", EmailVerified: true}, nil
			}
			return db.AuthUser{}, sql.ErrNoRows
		},
	}

	store := &db.DataStore{Auth: mockAuth}
	handler := HandleStatus(store)

	req, _ := http.NewRequest("GET", "/a/status", nil)
	req.AddCookie(&http.Cookie{Name: config.COOKIE_SESSION, Value: sessionID})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status OK, got %v", status)
	}

	var response UserStatusResponse
	json.NewDecoder(rr.Body).Decode(&response)
	if response.Email != "test@example.com" || response.EmailVerified != true {
		t.Errorf("unexpected response: %+v", response)
	}
}

func TestHandleStatus_InvalidSession(t *testing.T) {
	mockAuth := &db.TestMockAuthStore{
		GetUserBySessionIDFunc: func(sid string, d bool) (db.AuthUser, error) {
			return db.AuthUser{}, sql.ErrNoRows // Force failure to simulate bad cookie
		},
	}

	store := &db.DataStore{Auth: mockAuth}
	handler := HandleStatus(store)

	req, _ := http.NewRequest("GET", "/a/status", nil)
	req.AddCookie(&http.Cookie{Name: config.COOKIE_SESSION, Value: "invalid-or-expired"})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("expected status Unauthorized, got %v", status)
	}
}
