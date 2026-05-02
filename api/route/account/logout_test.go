package account

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"example_api/base/db"
	"example_api/config"
)

func TestHandleLogout_Success(t *testing.T) {
	sessionID := "valid-session-id"

	mockAuth := &db.TestMockAuthStore{
		DeleteSessionFunc: func(sid string) error {
			if sid == sessionID {
				return nil
			}
			return sql.ErrNoRows
		},
	}

	store := &db.DataStore{Auth: mockAuth}
	handler := HandleLogout(store)

	req, _ := http.NewRequest("DELETE", "/session/logout", nil)
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

func TestHandleLogout_DatabaseError(t *testing.T) {
	mockAuth := &db.TestMockAuthStore{
		DeleteSessionFunc: func(sid string) error {
			return sql.ErrConnDone // Simulate DB failure
		},
	}

	store := &db.DataStore{Auth: mockAuth}
	handler := HandleLogout(store)

	req, _ := http.NewRequest("DELETE", "/session/logout", nil)
	req.AddCookie(&http.Cookie{Name: config.COOKIE_SESSION, Value: "some-session-id"})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("expected status %v for database failure, got %v", http.StatusInternalServerError, status)
	}
}
