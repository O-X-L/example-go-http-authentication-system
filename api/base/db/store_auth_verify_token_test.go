package db

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestDefaultAuthStore_AddVerificationToken(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	userID := 123
	tokenID := "token-id-123"
	token := "secure-token"

	mock.ExpectExec(`INSERT INTO verification_token \(id, user_id, token, usage_id, expires_at\) VALUES \(\$1, \$2, \$3, \$4, \$5\)`).
		WithArgs(tokenID, userID, token, 1, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := store.AddVerificationToken(userID, VERIFICATION_LINK_USAGE_VERIFY_EMAIL, tokenID, token)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled database expectations: %s", err)
	}
}

func TestDefaultAuthStore_GetVerificationToken(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	tokenID := "token-id-123"
	expectedUserID := 123
	expectedToken := "secure-token"
	expectedUsageID := 1
	expectedExpiresAt := time.Now().Add(1 * time.Hour)

	rows := sqlmock.NewRows([]string{"user_id", "token", "usage_id", "expires_at"}).
		AddRow(expectedUserID, expectedToken, expectedUsageID, expectedExpiresAt)

	mock.ExpectQuery(`SELECT user_id, token, usage_id, expires_at FROM verification_token WHERE id = \$1`).
		WithArgs(tokenID).
		WillReturnRows(rows)

	userID, usageID, token, expiresAt, err := store.GetVerificationToken(tokenID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if userID != expectedUserID || token != expectedToken || usageID != VERIFICATION_LINK_USAGE_VERIFY_EMAIL || !expiresAt.Equal(expectedExpiresAt) {
		t.Errorf("unexpected return values")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled database expectations: %s", err)
	}
}

func TestDefaultAuthStore_SetVerificationDoneUserEmail(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	mock.ExpectExec(`UPDATE users SET email_verified = true, active = true WHERE id = \$1`).
		WithArgs(123).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := store.SetVerificationDoneUserEmail(123)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled database expectations: %s", err)
	}
}

func TestDefaultAuthStore_RenewVerificationTokenForResend_Success(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	userID := 123
	usageID := VERIFICATION_LINK_USAGE_VERIFY_EMAIL

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT resend_count FROM verification_token WHERE user_id = \$1 AND usage_id = \$2 FOR UPDATE`).
		WithArgs(userID, usageID).
		WillReturnRows(sqlmock.NewRows([]string{"resend_count"}).AddRow(0))

	mock.ExpectExec(`UPDATE verification_token SET id = \$1, token = \$2, expires_at = \$3, resend_count = resend_count \+ 1 WHERE user_id = \$4 AND usage_id = \$5`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), userID, usageID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	tokenID, token, err := store.RenewVerificationTokenForResend(userID, usageID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if tokenID == "" || token == "" {
		t.Errorf("expected generated tokens, got empty strings")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled database expectations: %s", err)
	}
}

func TestDefaultAuthStore_RenewVerificationTokenForResend_RateLimit(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	userID := 123
	usageID := VERIFICATION_LINK_USAGE_VERIFY_EMAIL

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT resend_count FROM verification_token WHERE user_id = \$1 AND usage_id = \$2 FOR UPDATE`).
		WithArgs(userID, usageID).
		WillReturnRows(sqlmock.NewRows([]string{"resend_count"}).AddRow(2)) // Matches config.VERIFICATION_LINK_MAX_RESEND

	// The transaction should rollback when the error is returned
	mock.ExpectRollback()

	_, _, err := store.RenewVerificationTokenForResend(userID, usageID)
	if err == nil || err.Error() != "resend limit exceeded" {
		t.Errorf("expected rate limit error, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled database expectations: %s", err)
	}
}
