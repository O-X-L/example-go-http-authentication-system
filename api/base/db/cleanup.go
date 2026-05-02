package db

import (
	"database/sql"
	"log"
	"time"
)

func cleanupExpiredSessions(dbConn *sql.DB) {
	_, err := dbConn.Exec("DELETE FROM sessions WHERE expires_at < now() - interval '2 hours'")
	if err != nil {
		log.Printf("error while running cleaning-up old sessions: %v\n", err)
	}
}

func cleanupExpiredGuestSessions(dbConn *sql.DB) {
	_, err := dbConn.Exec("DELETE FROM sessions_guest WHERE expires_at < now() - interval '2 hours'")
	if err != nil {
		log.Printf("error while running cleaning-up old guest-sessions: %v\n", err)
	}
}

func cleanupExpiredVerificationTokens(dbConn *sql.DB) {
	// note: keep tokens longer because of resend-counter/ratelimit
	_, err := dbConn.Exec("DELETE FROM verification_token WHERE expires_at < now() - interval '24 hours'")
	if err != nil {
		log.Printf("error while running cleaning-up old verification-tokens: %v\n", err)
	}
}

func CleanupTasks(dbConn *sql.DB) {
	cleanupTasksHourly(dbConn)
}

func cleanupTasksHourly(dbConn *sql.DB) {
	for {
		time.Sleep(time.Hour)
		cleanupExpiredSessions(dbConn)
		cleanupExpiredGuestSessions(dbConn)
		cleanupExpiredVerificationTokens(dbConn)
	}
}
