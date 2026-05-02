package db

import (
	"crypto/subtle"
	"example_api/config"
	"example_api/util"
	"fmt"
	"log"
	"time"
)

type AuthSession struct {
	UserID    int
	ID        string
	CSRFToken string
	StartedAt time.Time
	ExpiresAt time.Time
}

func (s *AuthSession) New() {
	s.ID = util.GenerateToken32()
	s.CSRFToken = util.GenerateToken32()

	s.StartedAt = time.Now()
	s.ExpiresAt = time.Now().Add(config.TIMEOUT_SESSION_IDLE)
}

func (s *AuthSession) FromExisting(existingSession AuthSession) {
	s.ID = util.GenerateToken32()
	s.CSRFToken = util.GenerateToken32()
	s.ExpiresAt = time.Now().Add(config.TIMEOUT_SESSION_IDLE)

	s.UserID = existingSession.UserID
	s.StartedAt = existingSession.StartedAt
}

func (s *AuthSession) IsComplete() bool {
	if s.UserID == 0 {
		return false
	}
	if len(s.ID) != 43 || len(s.CSRFToken) != 43 {
		return false
	}
	if s.StartedAt.IsZero() || s.ExpiresAt.IsZero() {
		return false
	}
	return true
}

func (s *AuthSession) IsExpired() bool {
	return s.IsExpiredIdle() || s.IsExpiredForce()
}

func (s *AuthSession) IsExpiredIdle() bool {
	return time.Now().After(s.ExpiresAt)
}

func (s *AuthSession) IsExpiredForce() bool {
	return time.Now().After(s.StartedAt.Add(config.TIMEOUT_SESSION_FORCE))
}

func (s *AuthSession) TokenAge() time.Duration {
	if s.IsExpiredIdle() {
		return config.TIMEOUT_SESSION_IDLE
	}
	timeLeft := s.ExpiresAt.Sub(time.Now())
	return config.TIMEOUT_SESSION_IDLE - timeLeft
}

func (s *AuthSession) ShouldRotateToken() bool {
	if s.IsExpired() {
		return false
	}
	return s.TokenAge() > config.TIMEOUT_SESSION_ROTATE
}

func (s *AuthSession) CSRFMatches(csrf string) bool {
	return subtle.ConstantTimeCompare([]byte(csrf), []byte(s.CSRFToken)) == 1
}

// DB INTERACTION

func (s *DefaultAuthStore) GetSession(sessionID string) (AuthSession, error) {
	us := AuthSession{ID: sessionID}
	err := s.DBConn.QueryRow(
		"SELECT user_id, csrf_token, expires_at, started_at FROM sessions WHERE id = $1",
		sessionID,
	).Scan(&us.UserID, &us.CSRFToken, &us.ExpiresAt, &us.StartedAt)

	if err == nil && us.IsExpired() {
		err = fmt.Errorf("session expired")
	}
	return us, err
}

func (s *DefaultAuthStore) AddSession(session AuthSession) error {
	if !session.IsComplete() {
		log.Printf("ERROR: AddSession got incomplete session-struct: %+v", session)
		return fmt.Errorf("incomplete session")
	}
	_, err := s.DBConn.Exec(
		"INSERT INTO sessions (id, user_id, csrf_token, expires_at, started_at) VALUES ($1, $2, $3, $4, $5)",
		session.ID,
		session.UserID,
		session.CSRFToken,
		session.ExpiresAt,
		session.StartedAt,
	)
	if !config.IsDeploymentProduction() {
		log.Printf("New user-session: %s => %s", session.ID, session.CSRFToken)
	}
	return err
}

func (s *DefaultAuthStore) DeleteSession(sessionID string) error {
	_, err := s.DBConn.Exec("DELETE FROM sessions WHERE id = $1", sessionID)
	return err
}

func (s *DefaultAuthStore) RotateSession(sessionOld AuthSession, sessionNew AuthSession) error {
	if !sessionOld.IsComplete() {
		log.Printf("ERROR: RotateSession got incomplete session-struct (old): %+v", sessionOld)
		return fmt.Errorf("incomplete session")
	}
	if !sessionNew.IsComplete() {
		log.Printf("ERROR: RotateSession got incomplete session-struct (new): %+v", sessionNew)
		return fmt.Errorf("incomplete session")
	}
	_, err := s.DBConn.Exec(
		"INSERT INTO sessions (id, user_id, csrf_token, expires_at, started_at) VALUES ($1, $2, $3, $4, $5)",
		sessionNew.ID,
		sessionNew.UserID,
		sessionNew.CSRFToken,
		sessionNew.ExpiresAt,
		sessionNew.StartedAt,
	)
	if err != nil {
		return err
	}

	oldSessionExpiration := time.Now().Add(60 * time.Second) // invalidate session after grace-period
	_, err = s.DBConn.Exec(
		"UPDATE sessions SET expires_at = $1 WHERE id = $2",
		oldSessionExpiration, sessionOld.ID,
	)
	if err != nil {
		return err
	}

	if !config.IsDeploymentProduction() {
		log.Printf("Rotated user-session: %s => %s (%s)", sessionOld.ID, sessionNew.ID, sessionNew.CSRFToken)
	}

	return nil
}

func (s *DefaultAuthStore) EnforceSessionPerUserLimit(userID int) error {
	_, err := s.DBConn.Exec(
		`DELETE FROM sessions 
		WHERE user_id = $1 
	    AND id NOT IN (
	 	  SELECT id FROM sessions 
	 	  WHERE user_id = $1 AND expires_at > NOW()
		  ORDER BY started_at DESC, expires_at DESC 
		  LIMIT $2
		)`,
		userID,
		config.MAX_CONCURRENT_SESSIONS,
	)
	if err != nil {
		log.Printf("ERROR: EnforceSessionLimit failed for user %d: %v", userID, err)
	}
	return err
}
