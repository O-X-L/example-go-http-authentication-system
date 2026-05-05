package db

import (
	"database/sql"
	"example_api/util"
	"time"
)

const (
	AUTH_TYPE_BASIC        = 1
	AUTH_TYPE_OAUTH_GOOGLE = 2

	VERIFICATION_LINK_USAGE_VERIFY_EMAIL   = 1
	VERIFICATION_LINK_USAGE_PASSWORD_RESET = 2

	LOGON_DEVICE_TYPE_WEB = 1
)

type AuthUser struct {
	ID                  int
	Email               string
	AuthType            int
	PasswordHash        string
	EmailVerified       bool
	IsActive            bool
	LastLogonTime       time.Time
	LastLogonAuthType   int
	LastLogonDeviceType int
	LastLogonDeviceInfo string
	SessionID           string
	SessionCSRFToken    string
}

type DefaultAuthStore struct {
	DBConn *sql.DB
}

// DB INTERACTION

func (s *DefaultAuthStore) AddUserBasic(email, passwordHash string) error {
	_, err := s.DBConn.Exec(
		"INSERT INTO users (email, password_hash, auth_type) VALUES ($1, $2, $3)",
		email,
		passwordHash,
		AUTH_TYPE_BASIC,
	)
	return err
}

func (s *DefaultAuthStore) AddUserOAuth(email string, providerID int) error {
	randomPasswordHash, err := util.Argon2HashSecret(util.GenerateToken32())
	if err != nil {
		return err
	}

	_, err = s.DBConn.Exec(
		"INSERT INTO users (email, password_hash, auth_type, email_verified) VALUES ($1, $2, $3, true)",
		email,
		randomPasswordHash,
		providerID,
	)
	return err
}

func (s *DefaultAuthStore) populateUserDetails(us *AuthUser) error {
	llTime, llAuthType, llDeviceType, llDeviceInfo, err := s.GetLastUserLogon(us.ID)
	if err == nil {
		us.LastLogonTime = llTime
		us.LastLogonAuthType = llAuthType
		us.LastLogonDeviceType = llDeviceType
		us.LastLogonDeviceInfo = llDeviceInfo
	}
	return err
}

func (s *DefaultAuthStore) GetUserByEmail(email string, details bool) (AuthUser, error) {
	us := AuthUser{Email: email}
	err := s.DBConn.QueryRow(
		"SELECT id, auth_type, password_hash, email_verified, active FROM users WHERE email = $1",
		email,
	).Scan(&us.ID, &us.AuthType, &us.PasswordHash, &us.EmailVerified, &us.IsActive)

	if err == nil && details {
		err = s.populateUserDetails(&us)
	}
	return us, err
}

func (s *DefaultAuthStore) GetUserByID(userID int, details bool) (AuthUser, error) {
	us := AuthUser{ID: userID}
	err := s.DBConn.QueryRow(
		"SELECT email, auth_type, password_hash, email_verified, active FROM users WHERE id = $1",
		userID,
	).Scan(&us.Email, &us.AuthType, &us.PasswordHash, &us.EmailVerified, &us.IsActive)

	if err == nil && details {
		err = s.populateUserDetails(&us)
	}
	return us, err
}

func (s *DefaultAuthStore) GetUserBySessionID(sessionID string, details bool) (AuthUser, error) {
	us := AuthUser{SessionID: sessionID}
	err := s.DBConn.QueryRow(
		`SELECT
			u.id,
			u.email,
			u.auth_type,
			u.password_hash,
			u.email_verified,
			u.active,
			s.csrf_token
		FROM users u
		JOIN sessions s ON u.id = s.user_id
		WHERE s.id = $1`,
		sessionID,
	).Scan(
		&us.ID, &us.Email, &us.AuthType, &us.PasswordHash, &us.EmailVerified, &us.IsActive,
		&us.SessionCSRFToken,
	)

	if err == nil && details {
		err = s.populateUserDetails(&us)
	}
	return us, err
}

func (s *DefaultAuthStore) DeleteUser(userID int) error {
	_, err := s.DBConn.Exec("DELETE FROM users WHERE id = $1", userID)
	return err
}

func (s *DefaultAuthStore) IsUserActive(userID int) (bool, error) {
	var active bool
	err := s.DBConn.QueryRow("SELECT active FROM users WHERE id = $1", userID).Scan(&active)
	return active, err
}

func (s *DefaultAuthStore) IsUserEmailVerified(userID int) (bool, error) {
	var verified bool
	err := s.DBConn.QueryRow("SELECT email_verified FROM users WHERE id = $1", userID).Scan(&verified)
	return verified, err
}

func (s *DefaultAuthStore) RegisterUserLogon(userID, authType, deviceType int, deviceInfo string) error {
	_, err := s.DBConn.Exec(
		"INSERT INTO user_logons (user_id, auth_type, device_type, device_info) VALUES ($1, $2, $3, $4)",
		userID,
		authType,
		deviceType,
		deviceInfo,
	)
	return err
}

func (s *DefaultAuthStore) GetLastUserLogon(userID int) (time.Time, int, int, string, error) {
	var logonTime time.Time
	var authType int
	var deviceType int
	var deviceInfo string
	err := s.DBConn.QueryRow(
		"SELECT time, auth_type, device_type, device_info FROM user_logons WHERE user_id = $1 ORDER BY time DESC LIMIT 1",
		userID,
	).Scan(&logonTime, &authType, &deviceType, &deviceInfo)
	return logonTime, authType, deviceType, deviceInfo, err
}

func (s *DefaultAuthStore) UpdateUserPassword(userID int, passwordHash string) error {
	_, err := s.DBConn.Exec("UPDATE users SET password_hash = $1 WHERE id = $2", passwordHash, userID)
	return err
}

func (s *DefaultAuthStore) DeleteAllUserSessions(userID int) error {
	_, err := s.DBConn.Exec("DELETE FROM sessions WHERE user_id = $1", userID)
	return err
}
