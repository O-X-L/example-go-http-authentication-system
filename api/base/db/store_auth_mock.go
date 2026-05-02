package db

import (
	"time"
)

// mockAuthStore implements db.AuthStore for testing
type TestMockAuthStore struct {
	GetGuestSessionFunc func(sessionID, clientIP string) (AuthSessionGuest, error)
	AddGuestSessionFunc func(session AuthSessionGuest) error

	GetSessionFunc                 func(sessionID string) (AuthSession, error)
	GetUserBySessionIDFunc         func(sessionID string, details bool) (AuthUser, error)
	AddSessionFunc                 func(session AuthSession) error
	DeleteSessionFunc              func(sessionID string) error
	RotateSessionFunc              func(sessionOld, sessionNew AuthSession) error
	EnforceSessionPerUserLimitFunc func(userID int) error

	AddUserBasicFunc        func(email, passwordHash string) error
	AddUserOAuthFunc        func(email string, providerID int) error
	GetUserByIDFunc         func(userID int, details bool) (AuthUser, error)
	GetUserByEmailFunc      func(email string, details bool) (AuthUser, error)
	DeleteUserFunc          func(userID int) error
	IsUserActiveFunc        func(userID int) (bool, error)
	IsUserEmailVerifiedFunc func(userID int) (bool, error)
	RegisterUserLogonFunc   func(userID, authType, deviceType int, deviceInfo string) error
	GetLastUserLogonFunc    func(userID int) (time.Time, int, int, string, error)

	AddVerificationTokenFunc            func(userID, usageID int, tokenID, token string) error
	GetVerificationTokenFunc            func(tokenID string) (int, int, string, time.Time, error)
	SetVerificationDoneUserEmailFunc    func(userID int) error
	DeleteVerificationTokenFunc         func(tokenID string) error
	RenewVerificationTokenForResendFunc func(userID, usageID int) (string, string, error)
}

func (m *TestMockAuthStore) GetGuestSession(sessionID, clientIP string) (AuthSessionGuest, error) {
	if m.GetGuestSessionFunc != nil {
		return m.GetGuestSessionFunc(sessionID, clientIP)
	}
	return AuthSessionGuest{}, nil
}

func (m *TestMockAuthStore) AddGuestSession(session AuthSessionGuest) error {
	if m.AddGuestSessionFunc != nil {
		return m.AddGuestSessionFunc(session)
	}
	return nil
}

func (m *TestMockAuthStore) GetSession(sessionID string) (AuthSession, error) {
	if m.GetSessionFunc != nil {
		return m.GetSessionFunc(sessionID)
	}
	return AuthSession{}, nil
}

func (m *TestMockAuthStore) GetUserBySessionID(sessionID string, details bool) (AuthUser, error) {
	if m.GetUserBySessionIDFunc != nil {
		return m.GetUserBySessionIDFunc(sessionID, details)
	}
	return AuthUser{}, nil
}

func (m *TestMockAuthStore) AddSession(session AuthSession) error {
	if m.AddSessionFunc != nil {
		return m.AddSessionFunc(session)
	}
	return nil
}

func (m *TestMockAuthStore) DeleteSession(sessionID string) error {
	if m.DeleteSessionFunc != nil {
		return m.DeleteSessionFunc(sessionID)
	}
	return nil
}

func (m *TestMockAuthStore) RotateSession(sessionOld, sessionNew AuthSession) error {
	if m.RotateSessionFunc != nil {
		return m.RotateSessionFunc(sessionOld, sessionNew)
	}
	return nil
}

func (m *TestMockAuthStore) EnforceSessionPerUserLimit(userID int) error {
	if m.EnforceSessionPerUserLimitFunc != nil {
		return m.EnforceSessionPerUserLimitFunc(userID)
	}
	return nil
}

func (m *TestMockAuthStore) AddUserBasic(email, passwordHash string) error {
	if m.AddUserBasicFunc != nil {
		return m.AddUserBasicFunc(email, passwordHash)
	}
	return nil
}

func (m *TestMockAuthStore) AddUserOAuth(email string, providerID int) error {
	if m.AddUserOAuthFunc != nil {
		return m.AddUserOAuthFunc(email, providerID)
	}
	return nil
}

func (m *TestMockAuthStore) GetUserByEmail(email string, details bool) (AuthUser, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(email, details)
	}
	return AuthUser{}, nil
}

func (m *TestMockAuthStore) GetUserByID(userID int, details bool) (AuthUser, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(userID, details)
	}
	return AuthUser{}, nil
}

func (m *TestMockAuthStore) DeleteUser(userID int) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(userID)
	}
	return nil
}

func (m *TestMockAuthStore) IsUserActive(userID int) (bool, error) {
	if m.IsUserActiveFunc != nil {
		return m.IsUserActiveFunc(userID)
	}
	return false, nil
}

func (m *TestMockAuthStore) IsUserEmailVerified(userID int) (bool, error) {
	if m.IsUserEmailVerifiedFunc != nil {
		return m.IsUserEmailVerifiedFunc(userID)
	}
	return false, nil
}

func (m *TestMockAuthStore) RegisterUserLogon(userID, authType, deviceType int, deviceInfo string) error {
	if m.RegisterUserLogonFunc != nil {
		return m.RegisterUserLogonFunc(userID, authType, deviceType, deviceInfo)
	}
	return nil
}

func (m *TestMockAuthStore) GetLastUserLogon(userID int) (time.Time, int, int, string, error) {
	if m.GetLastUserLogonFunc != nil {
		return m.GetLastUserLogonFunc(userID)
	}
	return time.Time{}, 0, 0, "", nil
}

func (m *TestMockAuthStore) AddVerificationToken(userID, usageID int, tokenID, token string) error {
	if m.AddVerificationTokenFunc != nil {
		return m.AddVerificationTokenFunc(userID, usageID, tokenID, token)
	}
	return nil
}

func (m *TestMockAuthStore) GetVerificationToken(tokenID string) (int, int, string, time.Time, error) {
	if m.GetVerificationTokenFunc != nil {
		return m.GetVerificationTokenFunc(tokenID)
	}
	return 0, 0, "", time.Time{}, nil
}

func (m *TestMockAuthStore) SetVerificationDoneUserEmail(userID int) error {
	if m.SetVerificationDoneUserEmailFunc != nil {
		return m.SetVerificationDoneUserEmailFunc(userID)
	}
	return nil
}

func (m *TestMockAuthStore) DeleteVerificationToken(tokenID string) error {
	if m.DeleteVerificationTokenFunc != nil {
		return m.DeleteVerificationTokenFunc(tokenID)
	}
	return nil
}

func (m *TestMockAuthStore) RenewVerificationTokenForResend(userID, usageID int) (string, string, error) {
	if m.RenewVerificationTokenForResendFunc != nil {
		return m.RenewVerificationTokenForResendFunc(userID, usageID)
	}
	return "", "", nil
}
