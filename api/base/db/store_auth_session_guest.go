package db

import (
	"crypto/subtle"
	"fmt"
	"example_api/config"
	"example_api/util"
	"log"
	"time"
)

type AuthSessionGuest struct {
	ID        string
	CSRFToken string
	ClientIP  string
	ExpiresAt time.Time
}

func (s *AuthSessionGuest) New() {
	s.ID = util.GenerateToken32()
	s.CSRFToken = util.GenerateToken32()

	s.ExpiresAt = time.Now().Add(config.TIMEOUT_SESSION_GUEST)
}

func (s *AuthSessionGuest) IsComplete() bool {
	if len(s.ID) != 43 || len(s.CSRFToken) != 43 {
		return false
	}
	if len(s.ClientIP) == 0 {
		return false
	}
	if s.ExpiresAt.IsZero() {
		return false
	}
	return true
}

func (s *AuthSessionGuest) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

func (s *AuthSessionGuest) CSRFMatches(csrf string) bool {
	return subtle.ConstantTimeCompare([]byte(csrf), []byte(s.CSRFToken)) == 1
}

// DB INTERACTION

func (s *DefaultAuthStore) GetGuestSession(sessionID, clientIP string) (AuthSessionGuest, error) {
	us := AuthSessionGuest{ID: sessionID, ClientIP: clientIP}
	err := s.DBConn.QueryRow(
		"SELECT csrf_token, expires_at, client_ip FROM sessions_guest WHERE id = $1 AND client_ip = $2",
		sessionID,
		clientIP,
	).Scan(&us.CSRFToken, &us.ExpiresAt, &us.ClientIP)

	if err == nil && time.Now().After(us.ExpiresAt) {
		err = fmt.Errorf("session expired")
	}
	return us, err
}

func (s *DefaultAuthStore) AddGuestSession(session AuthSessionGuest) error {
	if !session.IsComplete() {
		log.Printf("ERROR: AddGuestSession got incomplete session-struct: %+v", session)
		return fmt.Errorf("incomplete session")
	}
	_, err := s.DBConn.Exec(
		"INSERT INTO sessions_guest (id, client_ip, csrf_token, expires_at) VALUES ($1, $2, $3, $4)",
		session.ID,
		session.ClientIP,
		session.CSRFToken,
		session.ExpiresAt,
	)
	if !config.IsDeploymentProduction() {
		log.Printf("New guest-session: %s => %s (ip: %s)", session.ID, session.CSRFToken, session.ClientIP)
	}
	return err
}
