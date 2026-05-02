package db

import (
	"database/sql"
	"example_fe/config"
	"log"
	"time"
)

type DefaultAuthStore struct {
	DBConn *sql.DB
}

func (s *DefaultAuthStore) AddGuestSession(sessionID, clientIP, csrfToken string, expiresAt time.Time) error {
	_, err := s.DBConn.Exec(
		"INSERT INTO sessions_guest (id, client_ip, csrf_token, expires_at) VALUES ($1, $2, $3, $4)",
		sessionID,
		clientIP,
		csrfToken,
		expiresAt,
	)
	if !config.IsDeploymentProduction() {
		log.Printf("New guest-session: %s => %s (ip: %s)", sessionID, csrfToken, clientIP)
	}
	return err
}
