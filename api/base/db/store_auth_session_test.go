package db

import (
	"example_api/config"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// LOGIC: AuthSession

func TestAuthSession_Lifecycle(t *testing.T) {
	t.Run("New Session", func(t *testing.T) {
		s := AuthSession{UserID: 1}
		s.New()

		if !s.IsComplete() {
			t.Errorf("expected new session to be complete")
		}
		if s.IsExpired() {
			t.Errorf("new session should not be expired")
		}
		if s.ShouldRotateToken() {
			t.Errorf("new session should not need rotation")
		}
	})

	t.Run("FromExisting", func(t *testing.T) {
		old := AuthSession{UserID: 5}
		old.New()

		s := AuthSession{}
		s.FromExisting(old)

		if s.UserID != old.UserID {
			t.Errorf("expected UserID to carry over")
		}
		if s.ID == old.ID || s.CSRFToken == old.CSRFToken {
			t.Errorf("expected new ID and CSRF tokens to be generated")
		}
	})

	t.Run("IsComplete scenarios", func(t *testing.T) {
		s := AuthSession{} // completely empty
		if s.IsComplete() {
			t.Errorf("empty session should not be complete")
		}

		s.New() // generates ID, CSRF, dates, but UserID is 0
		if s.IsComplete() {
			t.Errorf("session without UserID should not be complete")
		}
	})

	t.Run("Expirations and Rotation", func(t *testing.T) {
		s := AuthSession{UserID: 1}
		s.New()

		// 1. Test Rotation
		// Simulate a token age that is just past the rotation threshold
		tokenAge := config.TIMEOUT_SESSION_ROTATE + time.Minute
		s.ExpiresAt = time.Now().Add(config.TIMEOUT_SESSION_IDLE - tokenAge)

		if !s.ShouldRotateToken() {
			t.Errorf("session token older than ROTATE timeout should require rotation")
		}

		// 2. Test Force Expiration
		// Simulate a session chain that exceeds the absolute maximum lifetime
		s.StartedAt = time.Now().Add(-config.TIMEOUT_SESSION_FORCE - time.Minute)
		if !s.IsExpiredForce() || !s.IsExpired() {
			t.Errorf("session chain older than FORCE timeout should be expired")
		}

		// Edge case: an expired session shouldn't rotate, it should just die
		if s.ShouldRotateToken() {
			t.Errorf("expired session should NOT be eligible for rotation")
		}
	})
}

func TestAuthSession_IncompleteGuard(t *testing.T) {
	store, _, teardown := setupAuthStore(t)
	defer teardown()

	t.Run("AddSession Guard", func(t *testing.T) {
		err := store.AddSession(AuthSession{}) // Empty/Incomplete
		if err == nil || err.Error() != "incomplete session" {
			t.Errorf("expected incomplete session error, got: %v", err)
		}
	})

	t.Run("RotateSession Guards", func(t *testing.T) {
		validSession := AuthSession{UserID: 1}
		validSession.New()

		// Test old session incomplete
		err := store.RotateSession(AuthSession{}, validSession)
		if err == nil || err.Error() != "incomplete session" {
			t.Errorf("expected error when old session is incomplete")
		}

		// Test new session incomplete
		err = store.RotateSession(validSession, AuthSession{})
		if err == nil || err.Error() != "incomplete session" {
			t.Errorf("expected error when new session is incomplete")
		}
	})
}

// DB INTERACTION

func TestDefaultAuthStore_GetSession(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	session := AuthSession{UserID: 123}
	session.New()

	rows := sqlmock.NewRows([]string{"user_id", "csrf_token", "expires_at", "started_at"}).
		AddRow(session.UserID, session.CSRFToken, session.ExpiresAt, session.StartedAt)

	mock.ExpectQuery(`SELECT user_id, csrf_token, expires_at, started_at FROM sessions WHERE id = \$1`).
		WithArgs(session.ID).
		WillReturnRows(rows)

	session, err := store.GetSession(session.ID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if session.CSRFToken != session.CSRFToken {
		t.Errorf("expected CSRF %s, got %s", session.CSRFToken, session.CSRFToken)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled database expectations: %s", err)
	}
}

func TestDefaultAuthStore_RotateSession(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	oldSession := AuthSession{UserID: 123}
	oldSession.New()

	newSession := AuthSession{}
	newSession.FromExisting(oldSession)

	// Expect the insertion of the new session
	mock.ExpectExec(`INSERT INTO sessions \(id, user_id, csrf_token, expires_at, started_at\) VALUES \(\$1, \$2, \$3, \$4, \$5\)`).
		WithArgs(newSession.ID, newSession.UserID, newSession.CSRFToken, newSession.ExpiresAt, newSession.StartedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect the update (grace period invalidation) of the old session
	mock.ExpectExec(`UPDATE sessions SET expires_at = \$1 WHERE id = \$2`).
		WithArgs(sqlmock.AnyArg(), oldSession.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := store.RotateSession(oldSession, newSession)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled database expectations: %s", err)
	}
}

func TestDefaultAuthStore_AddSession(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	session := AuthSession{
		UserID:    123,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	session.New()

	mock.ExpectExec(`INSERT INTO sessions \(id, user_id, csrf_token, expires_at, started_at\) VALUES \(\$1, \$2, \$3, \$4, \$5\)`).
		WithArgs(session.ID, session.UserID, session.CSRFToken, session.ExpiresAt, session.StartedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := store.AddSession(session)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled database expectations: %s", err)
	}
}

func TestDefaultAuthStore_DeleteSession(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	sessionID := "valid-session-id"

	mock.ExpectExec(`DELETE FROM sessions WHERE id = \$1`).
		WithArgs(sessionID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := store.DeleteSession(sessionID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled database expectations: %s", err)
	}
}

func TestDefaultAuthStore_ExpiredSessionGuard(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	t.Run("GetSession Expired", func(t *testing.T) {
		sessionID := "expired-user-session"
		pastTime := time.Now().Add(-1 * time.Hour) // Force expiration

		rows := sqlmock.NewRows([]string{"user_id", "csrf_token", "expires_at", "started_at"}).
			AddRow(123, "csrf", pastTime, pastTime)

		mock.ExpectQuery(`SELECT user_id, csrf_token, expires_at, started_at FROM sessions WHERE id = \$1`).
			WithArgs(sessionID).
			WillReturnRows(rows)

		_, err := store.GetSession(sessionID)
		if err == nil || err.Error() != "session expired" {
			t.Errorf("expected 'session expired' error, got: %v", err)
		}
	})
}

func TestDefaultAuthStore_EnforceSessionPerUserLimit(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	userID := 123

	// Note: We use \s+ in the regex to flexibly match the spaces/newlines in the query block
	mock.ExpectExec(`DELETE FROM sessions WHERE user_id = \$1 AND id NOT IN \( SELECT id FROM sessions WHERE user_id = \$1 AND expires_at > NOW\(\) ORDER BY started_at DESC, expires_at DESC LIMIT \$2 \)`).
		WithArgs(userID, config.MAX_CONCURRENT_SESSIONS).
		WillReturnResult(sqlmock.NewResult(0, 2)) // Pretending 2 old sessions were deleted

	err := store.EnforceSessionPerUserLimit(userID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled database expectations: %s", err)
	}
}
