package db

import (
	"time"
)

type DataStore struct {
	Auth AuthStore
}

type AuthStore interface {
	GetGuestSession(sessionID, clientIP string) (AuthSessionGuest, error)
	AddGuestSession(session AuthSessionGuest) error

	GetSession(sessionID string) (AuthSession, error)
	GetUserBySessionID(sessionID string, details bool) (AuthUser, error)
	AddSession(session AuthSession) error
	DeleteSession(sessionID string) error
	RotateSession(sessionOld, sessionNew AuthSession) error
	EnforceSessionPerUserLimit(userID int) error

	AddUserBasic(email, passwordHash string) error
	AddUserOAuth(email string, providerID int) error
	GetUserByID(userID int, details bool) (AuthUser, error)
	GetUserByEmail(email string, details bool) (AuthUser, error)
	DeleteUser(userID int) error
	IsUserActive(userID int) (bool, error)
	IsUserEmailVerified(userID int) (bool, error)
	RegisterUserLogon(userID, authType, deviceType int, deviceInfo string) error
	GetLastUserLogon(userID int) (time.Time, int, int, string, error)

	AddVerificationToken(userID, usageID int, tokenID, token string) error
	GetVerificationToken(tokenID string) (int, int, string, time.Time, error)
	SetVerificationDoneUserEmail(userID int) error
	DeleteVerificationToken(tokenID string) error
	RenewVerificationTokenForResend(userID, usageID int) (string, string, error)
}
