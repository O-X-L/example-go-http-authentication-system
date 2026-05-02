package account

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"example_api/base/db"
	"example_api/config"
)

func TestHandleDelete_Success(t *testing.T) {
	sessionID := "valid-session-id"
	userID := 123

	mockAuth := &db.TestMockAuthStore{
		GetUserBySessionIDFunc: func(sid string, d bool) (db.AuthUser, error) {
			if sid == sessionID {
				return db.AuthUser{ID: userID}, nil
			}
			return db.AuthUser{}, sql.ErrNoRows
		},
		DeleteUserFunc: func(uID int) error {
			if uID == userID {
				return nil
			}
			return sql.ErrNoRows
		},
	}

	store := &db.DataStore{Auth: mockAuth}
	handler := HandleDelete(store)

	req, _ := http.NewRequest("DELETE", "/delete", nil)
	req.AddCookie(&http.Cookie{Name: config.COOKIE_SESSION, Value: sessionID})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	cookies := rr.Result().Cookies()
	if len(cookies) == 0 || cookies[0].MaxAge != -1 {
		t.Error("expected session cookie to be cleared (MaxAge=-1)")
	}
}

func TestHandleDelete_InvalidSession(t *testing.T) {
	mockAuth := &db.TestMockAuthStore{
		GetUserBySessionIDFunc: func(sid string, d bool) (db.AuthUser, error) {
			return db.AuthUser{}, sql.ErrNoRows // Invalid session
		},
	}

	store := &db.DataStore{Auth: mockAuth}
	handler := HandleDelete(store)

	req, _ := http.NewRequest("DELETE", "/delete", nil)
	req.AddCookie(&http.Cookie{Name: config.COOKIE_SESSION, Value: "invalid-session-id"})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("expected status %v for invalid session, got %v", http.StatusUnauthorized, status)
	}
}
