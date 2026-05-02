package db

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func setupAuthStore(t *testing.T) (*DefaultAuthStore, sqlmock.Sqlmock, func()) {
	dbConn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening stub db: %v", err)
	}
	store := &DefaultAuthStore{DBConn: dbConn}
	return store, mock, func() { dbConn.Close() }
}

func TestDefaultAuthStore_GetUserByEmail(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	email := "test@example.com"
	expectedHash := "hashedpassword"
	expectedAuthType := AUTH_TYPE_BASIC

	// UPDATED: Added new columns to match the updated SQL query
	rows := sqlmock.NewRows([]string{"id", "auth_type", "password_hash", "email_verified", "active"}).
		AddRow(1, expectedAuthType, expectedHash, true, true)

	mock.ExpectQuery(`SELECT id, auth_type, password_hash, email_verified, active FROM users WHERE email = \$1`).
		WithArgs(email).
		WillReturnRows(rows)

	user, err := store.GetUserByEmail(email, false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if user.ID != 1 || user.PasswordHash != expectedHash || user.AuthType != expectedAuthType {
		t.Errorf("expected id=1, hash=%s, authType=%d, got id=%d, hash=%s, authType=%d",
			expectedHash, expectedAuthType, user.ID, user.PasswordHash, user.AuthType)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled database expectations: %s", err)
	}
}

func TestDefaultAuthStore_GetUserByEmail_NotFound(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	email := "unknown@example.com"
	mock.ExpectQuery(`SELECT id, auth_type, password_hash, email_verified, active FROM users WHERE email = \$1`).
		WithArgs(email).
		WillReturnError(sql.ErrNoRows)

	_, err := store.GetUserByEmail(email, false)
	if err != sql.ErrNoRows {
		t.Errorf("expected ErrNoRows, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled database expectations: %s", err)
	}
}

func TestDefaultAuthStore_GetUserByID(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	userID := 123
	expectedEmail := "test@example.com"

	rows := sqlmock.NewRows([]string{"email", "auth_type", "password_hash", "email_verified", "active"}).
		AddRow(expectedEmail, AUTH_TYPE_BASIC, "hash", true, true)

	mock.ExpectQuery(`SELECT email, auth_type, password_hash, email_verified, active FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnRows(rows)

	user, err := store.GetUserByID(userID, false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if user.Email != expectedEmail {
		t.Errorf("expected email %s, got %s", expectedEmail, user.Email)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled database expectations: %s", err)
	}
}

func TestDefaultAuthStore_GetUserByID_NotFound(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	userID := 999
	mock.ExpectQuery(`SELECT email, auth_type, password_hash, email_verified, active FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	_, err := store.GetUserByID(userID, false)
	if err != sql.ErrNoRows {
		t.Errorf("expected ErrNoRows, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled database expectations: %s", err)
	}
}

func TestDefaultAuthStore_GetUserByEmail_WithDetails(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	email := "detailed@example.com"
	expectedID := 1
	expectedAuthType := AUTH_TYPE_BASIC
	expectedLogonTime := time.Now()
	expectedLogonAuthType := AUTH_TYPE_OAUTH_GOOGLE // Made distinct to verify mapping
	expectedLogonDeviceType := LOGON_DEVICE_TYPE_WEB
	expectedLogonDeviceInfo := "Chrome abc"

	// Expectation 1: Fetch the core user profile
	userRows := sqlmock.NewRows([]string{"id", "auth_type", "password_hash", "email_verified", "active"}).
		AddRow(expectedID, expectedAuthType, "hash", true, true)

	mock.ExpectQuery(`SELECT id, auth_type, password_hash, email_verified, active FROM users WHERE email = \$1`).
		WithArgs(email).
		WillReturnRows(userRows)

	// Expectation 2: Fetch the last logon details (triggered by details=true)
	logonRows := sqlmock.NewRows([]string{"time", "auth_type", "device_type", "device_info"}).
		AddRow(expectedLogonTime, expectedLogonAuthType, expectedLogonDeviceType, expectedLogonDeviceInfo)

	mock.ExpectQuery(`SELECT time, auth_type, device_type, device_info FROM user_logons WHERE user_id = \$1 ORDER BY time DESC LIMIT 1`).
		WithArgs(expectedID).
		WillReturnRows(logonRows)

	// Call the function with details: true
	user, err := store.GetUserByEmail(email, true)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if user.ID != expectedID {
		t.Errorf("expected id %d, got %d", expectedID, user.ID)
	}
	if user.LastLogonAuthType != expectedLogonAuthType {
		t.Errorf("expected last logon auth type %d, got %d", expectedLogonAuthType, user.LastLogonAuthType)
	}
	if user.LastLogonTime.Unix() != expectedLogonTime.Unix() {
		t.Errorf("expected logon time %v, got %v", expectedLogonTime, user.LastLogonTime)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled database expectations: %s", err)
	}
}

func TestDefaultAuthStore_GetUserByID_WithDetails(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	userID := 123
	expectedEmail := "detailsbyid@example.com"
	expectedLogonTime := time.Now()
	expectedLogonAuthType := AUTH_TYPE_BASIC
	expectedLogonDeviceType := LOGON_DEVICE_TYPE_WEB
	expectedLogonDeviceInfo := "Chrome abc"

	// Expectation 1: Fetch the core user profile
	userRows := sqlmock.NewRows([]string{"email", "auth_type", "password_hash", "email_verified", "active"}).
		AddRow(expectedEmail, AUTH_TYPE_BASIC, "hash", true, true)

	mock.ExpectQuery(`SELECT email, auth_type, password_hash, email_verified, active FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnRows(userRows)

	// Expectation 2: Fetch the last logon details (triggered by details=true)
	logonRows := sqlmock.NewRows([]string{"time", "auth_type", "device_type", "device_info"}).
		AddRow(expectedLogonTime, expectedLogonAuthType, expectedLogonDeviceType, expectedLogonDeviceInfo)

	mock.ExpectQuery(`SELECT time, auth_type, device_type, device_info FROM user_logons WHERE user_id = \$1 ORDER BY time DESC LIMIT 1`).
		WithArgs(userID).
		WillReturnRows(logonRows)

	// Call the function with details: true
	user, err := store.GetUserByID(userID, true)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if user.Email != expectedEmail {
		t.Errorf("expected email %s, got %s", expectedEmail, user.Email)
	}
	if user.LastLogonAuthType != expectedLogonAuthType {
		t.Errorf("expected last logon auth type %d, got %d", expectedLogonAuthType, user.LastLogonAuthType)
	}
	if user.LastLogonTime.Unix() != expectedLogonTime.Unix() {
		t.Errorf("expected logon time %v, got %v", expectedLogonTime, user.LastLogonTime)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled database expectations: %s", err)
	}
}

func TestDefaultAuthStore_GetUserBySessionID(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	sessionID := "valid-session"
	expectedID := 1
	expectedEmail := "test@example.com"
	expectedAuthType := AUTH_TYPE_BASIC
	expectedHash := "hash"
	expectedCSRF := "csrf-token"

	rows := sqlmock.NewRows([]string{"id", "email", "auth_type", "password_hash", "email_verified", "active", "csrf_token"}).
		AddRow(expectedID, expectedEmail, expectedAuthType, expectedHash, true, true, expectedCSRF)

	mock.ExpectQuery(`SELECT u.id, u.email, u.auth_type, u.password_hash, u.email_verified, u.active, s.csrf_token FROM users u JOIN sessions s ON u.id = s.user_id WHERE s.id = \$1`).
		WithArgs(sessionID).
		WillReturnRows(rows)

	user, err := store.GetUserBySessionID(sessionID, false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if user.ID != expectedID || user.Email != expectedEmail || user.SessionCSRFToken != expectedCSRF {
		t.Errorf("unexpected return values: %+v", user)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled database expectations: %s", err)
	}
}

func TestDefaultAuthStore_RegisterUserLogon(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	userID := 123
	authType := AUTH_TYPE_BASIC
	deviceType := LOGON_DEVICE_TYPE_WEB
	deviceInfo := "Chrome abc"

	mock.ExpectExec(`INSERT INTO user_logons \(user_id, auth_type, device_type, device_info\) VALUES \(\$1, \$2, \$3, \$4\)`).
		WithArgs(userID, authType, deviceType, deviceInfo).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := store.RegisterUserLogon(userID, authType, deviceType, deviceInfo)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled database expectations: %s", err)
	}
}

func TestDefaultAuthStore_GetLastUserLogon(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	userID := 123
	expectedTime := time.Now()
	expectedAuthType := AUTH_TYPE_BASIC
	expectedDeviceType := LOGON_DEVICE_TYPE_WEB
	expectedDeviceInfo := "Chrome abc"

	rows := sqlmock.NewRows([]string{"time", "auth_type", "device_type", "device_info"}).
		AddRow(expectedTime, expectedAuthType, expectedDeviceType, expectedDeviceInfo)

	mock.ExpectQuery(`SELECT time, auth_type, device_type, device_info FROM user_logons WHERE user_id = \$1 ORDER BY time DESC LIMIT 1`).
		WithArgs(userID).
		WillReturnRows(rows)

	logonTime, authType, deviceType, deviceInfo, err := store.GetLastUserLogon(userID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if authType != expectedAuthType {
		t.Errorf("expected auth type %d, got %d", expectedAuthType, authType)
	}
	if logonTime.Unix() != expectedTime.Unix() {
		t.Errorf("expected time %v, got %v", expectedTime, logonTime)
	}
	if deviceType != expectedDeviceType {
		t.Errorf("expected device type %d, got %d", expectedDeviceType, deviceType)
	}
	if deviceInfo != expectedDeviceInfo {
		t.Errorf("expected device info %s, got %s", expectedDeviceInfo, deviceInfo)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled database expectations: %s", err)
	}
}

func TestDefaultAuthStore_GetLastUserLogon_NoRows(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	userID := 999
	mock.ExpectQuery(`SELECT time, auth_type, device_type, device_info FROM user_logons WHERE user_id = \$1 ORDER BY time DESC LIMIT 1`).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	_, _, _, _, err := store.GetLastUserLogon(userID)
	if err != sql.ErrNoRows {
		t.Errorf("expected ErrNoRows, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled database expectations: %s", err)
	}
}

func TestDefaultAuthStore_AddUserBasic(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	mock.ExpectExec(`INSERT INTO users \(email, password_hash, auth_type\) VALUES \(\$1, \$2, \$3\)`).
		WithArgs("new@example.com", "hash123", AUTH_TYPE_BASIC).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := store.AddUserBasic("new@example.com", "hash123")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled database expectations: %s", err)
	}
}

func TestDefaultAuthStore_AddUserOAuth(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	mock.ExpectExec(`INSERT INTO users \(email, password_hash, auth_type, email_verified\) VALUES \(\$1, \$2, \$3, true\)`).
		WithArgs("oauth@example.com", sqlmock.AnyArg(), AUTH_TYPE_OAUTH_GOOGLE).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := store.AddUserOAuth("oauth@example.com", AUTH_TYPE_OAUTH_GOOGLE)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled database expectations: %s", err)
	}
}

func TestDefaultAuthStore_DeleteUser(t *testing.T) {
	store, mock, teardown := setupAuthStore(t)
	defer teardown()

	userID := 123

	mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := store.DeleteUser(userID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled database expectations: %s", err)
	}
}
