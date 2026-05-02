package db

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// LOGIC: AuthSessionGuest

func TestAuthSessionGuest_Lifecycle(t *testing.T) {
	t.Run("New Guest Session", func(t *testing.T) {
		s := AuthSessionGuest{ClientIP: "127.0.0.1"}
		s.New()

		if !s.IsComplete() {
			t.Errorf("expected new guest session to be complete")
		}
		if s.IsExpired() {
			t.Errorf("new guest session should not be expired")
		}
	})

	t.Run("IsComplete scenarios", func(t *testing.T) {
		s := AuthSessionGuest{} // completely empty
		if s.IsComplete() {
			t.Errorf("empty session should not be complete")
		}

		s.New() // generates ID, CSRF, Dates, but IP is empty
		if s.IsComplete() {
			t.Errorf("session without ClientIP should not be complete")
		}
	})

	t.Run("Expiration", func(t *testing.T) {
		s := AuthSessionGuest{ClientIP: "127.0.0.1"}
		s.New()
		s.ExpiresAt = time.Now().Add(-1 * time.Minute)

		if !s.IsExpired() {
			t.Errorf("session with past ExpiresAt should be expired")
		}
	})
}

func TestAuthSessionGuest_IncompleteGuard(t *testing.T) {
	store, _, teardown := setupAuthStore(t)
	defer teardown()

	t.Run("AddGuestSession Guard", func(t *testing.T) {
		err := store.AddGuestSession(AuthSessionGuest{}) // Empty/Incomplete
		if err == nil || err.Error() != "incomplete session" {
			t.Errorf("expected incomplete session error, got: %v", err)
		}
	})
}

// DB INTERACTION

func TestDefaultAuthStore_AddGuestSession(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	session := AuthSessionGuest{ClientIP: "192.168.1.1"}
	session.New()

	mock.ExpectExec(`INSERT INTO sessions_guest \(id, client_ip, csrf_token, expires_at\) VALUES \(\$1, \$2, \$3, \$4\)`).
		WithArgs(session.ID, session.ClientIP, session.CSRFToken, session.ExpiresAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := store.AddGuestSession(session)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled database expectations: %s", err)
	}
}

func TestDefaultAuthStore_GetGuestSession(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	sessionID := "guest-session-id"
	clientIP := "192.168.1.1"
	expectedCSRF := "csrf-token-123"
	expectedExpires := time.Now().Add(1 * time.Hour)

	rows := sqlmock.NewRows([]string{"csrf_token", "expires_at", "client_ip"}).
		AddRow(expectedCSRF, expectedExpires, clientIP)

	mock.ExpectQuery(`SELECT csrf_token, expires_at, client_ip FROM sessions_guest WHERE id = \$1 AND client_ip = \$2`).
		WithArgs(sessionID, clientIP).
		WillReturnRows(rows)

	session, err := store.GetGuestSession(sessionID, clientIP)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if session.CSRFToken != expectedCSRF {
		t.Errorf("expected CSRF %s, got %s", expectedCSRF, session.CSRFToken)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled database expectations: %s", err)
	}
}

func TestDefaultAuthStore_ExpiredGuestSessionGuard(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	t.Run("GetGuestSession Expired", func(t *testing.T) {
		sessionID := "expired-guest-session"
		clientIP := "127.0.0.1"
		pastTime := time.Now().Add(-1 * time.Hour)

		rows := sqlmock.NewRows([]string{"csrf_token", "expires_at", "client_ip"}).
			AddRow("csrf", pastTime, clientIP)

		mock.ExpectQuery(`SELECT csrf_token, expires_at, client_ip FROM sessions_guest WHERE id = \$1 AND client_ip = \$2`).
			WithArgs(sessionID, clientIP).
			WillReturnRows(rows)

		_, err := store.GetGuestSession(sessionID, clientIP)
		if err == nil || err.Error() != "session expired" {
			t.Errorf("expected 'session expired' error, got: %v", err)
		}
	})
}
