package db

import "time"

type DataStore struct {
	Auth AuthStore
}

type AuthStore interface {
	AddGuestSession(sessionID, clientIP, csrfToken string, expiresAt time.Time) error
}
