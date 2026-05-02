package base

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"example_api/base/db"
	"example_api/config"

	"github.com/go-playground/validator/v10"
)

func TestGuestMiddleware_Scenarios(t *testing.T) {
	v := validator.New()
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	validSessionID := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" // 43 chars
	validCSRFToken := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" // 43 chars

	mockAuth := &db.TestMockAuthStore{
		AddGuestSessionFunc: func(session db.AuthSessionGuest) error {
			return nil
		},
		GetGuestSessionFunc: func(sid, ip string) (db.AuthSessionGuest, error) {
			if sid == validSessionID {
				return db.AuthSessionGuest{CSRFToken: validCSRFToken}, nil
			}
			return db.AuthSessionGuest{}, sql.ErrNoRows
		},
		// Added to test the fallback logic when a user hits a guest route
		GetSessionFunc: func(sid string) (db.AuthSession, error) {
			if sid == validSessionID {
				return db.AuthSession{CSRFToken: validCSRFToken}, nil
			}
			return db.AuthSession{}, sql.ErrNoRows
		},
	}

	store := &db.DataStore{Auth: mockAuth}
	handler := GuestMiddleware(store, v, nextHandler)

	t.Run("simple GET to register (creates guest session)", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/a/register/basic", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200 OK, got %d", rr.Code)
		}
	})

	t.Run("user session falls back to AuthMiddleware", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/a/register/basic", nil)
		// Attach a VALID USER session cookie to a GUEST request
		req.AddCookie(&http.Cookie{Name: config.COOKIE_SESSION, Value: validSessionID})
		req.RemoteAddr = "127.0.0.1:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		// It should pass through the AuthMiddleware successfully
		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200 OK, got %d", rr.Code)
		}
	})

	t.Run("simple POST to register (missing CSRF)", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/a/register/basic", nil)
		req.AddCookie(&http.Cookie{Name: config.COOKIE_SESSION_PREAUTH, Value: validSessionID})
		req.RemoteAddr = "127.0.0.1:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status 400 Bad Request, got %d", rr.Code)
		}
	})

	t.Run("session-id in wrong format", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/a/register/basic", nil)
		req.AddCookie(&http.Cookie{Name: config.COOKIE_SESSION_PREAUTH, Value: "wrong_format_too_short"})
		req.RemoteAddr = "127.0.0.1:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401 Unauthorized, got %d", rr.Code)
		}
	})

	t.Run("session-id valid and exists - correct csrf-header", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/a/register/basic", nil)
		req.AddCookie(&http.Cookie{Name: config.COOKIE_SESSION_PREAUTH, Value: validSessionID})
		req.Header.Set(config.HEADER_CSRF_PREAUTH, validCSRFToken)
		req.RemoteAddr = "127.0.0.1:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200 OK, got %d", rr.Code)
		}
	})

	t.Run("session-id valid and exists - wrong csrf-header", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/a/register/basic", nil)
		req.AddCookie(&http.Cookie{Name: config.COOKIE_SESSION_PREAUTH, Value: validSessionID})
		req.Header.Set(config.HEADER_CSRF_PREAUTH, "ccccccccccccccccccccccccccccccccccccccccccc")
		req.RemoteAddr = "127.0.0.1:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		// Middleware returns 403 Forbidden for mismatched tokens
		if rr.Code != http.StatusForbidden {
			t.Errorf("expected status 403 Forbidden, got %d", rr.Code)
		}
	})
}

func TestAuthMiddleware_Scenarios(t *testing.T) {
	v := validator.New()
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	validSessionID := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"  // 43 chars
	rotateSessionID := "rotateaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" // 43 chars
	validCSRFToken := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"  // 43 chars

	var rotateCalled bool

	mockAuth := &db.TestMockAuthStore{
		GetSessionFunc: func(sid string) (db.AuthSession, error) {
			if sid == validSessionID {
				// Normal session, not old enough to rotate
				s := db.AuthSession{CSRFToken: validCSRFToken, StartedAt: time.Now(), ExpiresAt: time.Now().Add(config.TIMEOUT_SESSION_IDLE)}
				return s, nil
			}
			if sid == rotateSessionID {
				// Old session that should trigger ShouldRotate()
				s := db.AuthSession{UserID: 1, CSRFToken: validCSRFToken}
				s.New() // Generate IDs
				s.StartedAt = time.Now().Add(-config.TIMEOUT_SESSION_ROTATE - time.Minute)
				s.ExpiresAt = time.Now().Add(config.TIMEOUT_SESSION_IDLE - config.TIMEOUT_SESSION_ROTATE - time.Minute)
				return s, nil
			}
			return db.AuthSession{}, sql.ErrNoRows
		},
		RotateSessionFunc: func(sessionOld, sessionNew db.AuthSession) error {
			rotateCalled = true
			return nil
		},
	}

	store := &db.DataStore{Auth: mockAuth}
	handler := AuthMiddleware(store, v, nextHandler)

	t.Run("no session cookie present", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/a/protected", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401 Unauthorized, got %d", rr.Code)
		}
	})

	t.Run("GET request - session valid and exists", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/a/protected", nil)
		req.AddCookie(&http.Cookie{Name: config.COOKIE_SESSION, Value: validSessionID})
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200 OK, got %d", rr.Code)
		}
	})

	t.Run("GET request - session triggers rotation", func(t *testing.T) {
		rotateCalled = false // reset state
		req, _ := http.NewRequest("GET", "/a/protected", nil)
		req.AddCookie(&http.Cookie{Name: config.COOKIE_SESSION, Value: rotateSessionID})
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200 OK, got %d", rr.Code)
		}
		if !rotateCalled {
			t.Errorf("expected RotateSession to be called for old token")
		}
	})

	t.Run("POST request - session valid and exists - correct csrf-header", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/a/protected/action", nil)
		req.AddCookie(&http.Cookie{Name: config.COOKIE_SESSION, Value: validSessionID})
		req.Header.Set(config.HEADER_CSRF_POSTAUTH, validCSRFToken)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200 OK, got %d", rr.Code)
		}
	})

	t.Run("POST request - session valid and exists - missing csrf-header", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/a/protected/action", nil)
		req.AddCookie(&http.Cookie{Name: config.COOKIE_SESSION, Value: validSessionID})
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status 400 Bad Request, got %d", rr.Code)
		}
	})

	t.Run("POST request - session valid and exists - wrong csrf-header", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/a/protected/action", nil)
		req.AddCookie(&http.Cookie{Name: config.COOKIE_SESSION, Value: validSessionID})
		req.Header.Set(config.HEADER_CSRF_POSTAUTH, "ccccccccccccccccccccccccccccccccccccccccccc")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		// Middleware returns 403 Forbidden for mismatched tokens
		if rr.Code != http.StatusForbidden {
			t.Errorf("expected status 403 Forbidden, got %d", rr.Code)
		}
	})

	t.Run("session not found in DB", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/a/protected", nil)
		req.AddCookie(&http.Cookie{Name: config.COOKIE_SESSION, Value: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"})
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401 Unauthorized, got %d", rr.Code)
		}
	})
}
