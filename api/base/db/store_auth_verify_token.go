package db

import (
	"example_api/config"
	"example_api/util"
	"fmt"
	"time"
)

func (s *DefaultAuthStore) AddVerificationToken(userID, usageID int, tokenID, token string) error {
	expiresAt := time.Now().Add(config.TIMEOUT_VERIFICATION_TOKEN)
	_, err := s.DBConn.Exec(
		"INSERT INTO verification_token (id, user_id, token, usage_id, expires_at) VALUES ($1, $2, $3, $4, $5)",
		tokenID,
		userID,
		token,
		usageID,
		expiresAt,
	)
	return err
}

func (s *DefaultAuthStore) GetVerificationToken(tokenID string) (int, int, string, time.Time, error) {
	var userID int
	var token string
	var usageID int
	var expiresAt time.Time
	err := s.DBConn.QueryRow(
		"SELECT user_id, token, usage_id, expires_at FROM verification_token WHERE id = $1",
		tokenID,
	).Scan(&userID, &token, &usageID, &expiresAt)
	return userID, usageID, token, expiresAt, err
}

func (s *DefaultAuthStore) SetVerificationDoneUserEmail(userID int) error {
	_, err := s.DBConn.Exec("UPDATE users SET email_verified = true, active = true WHERE id = $1", userID)
	return err
}

func (s *DefaultAuthStore) DeleteVerificationToken(tokenID string) error {
	_, err := s.DBConn.Exec("DELETE FROM verification_token WHERE id = $1", tokenID)
	return err
}

func (s *DefaultAuthStore) RenewVerificationTokenForResend(userID, usageID int) (string, string, error) {
	newToken := util.GenerateToken32()
	newTokenID := util.GenerateToken32()

	tx, err := s.DBConn.Begin()
	if err != nil {
		return "", "", err
	}
	defer tx.Rollback()

	// check if token exists; if so - query resend_count
	var resendCounter int
	err = tx.QueryRow(
		`SELECT resend_count FROM verification_token WHERE user_id = $1 AND usage_id = $2 FOR UPDATE`,
		userID,
		usageID,
	).Scan(&resendCounter)

	if err != nil {
		// add new one if it does not exist (maybe it was cleaned-up)
		err = s.AddVerificationToken(userID, usageID, newTokenID, newToken)
		if err != nil {
			return "", "", err
		}
		return newTokenID, newToken, nil
	}

	// ensure resend-ratelimit is enforced
	if resendCounter >= (config.VERIFICATION_LINK_MAX_RESEND - 1) {
		return "", "", fmt.Errorf("resend limit exceeded")
	}

	// apply newly generated tokens
	expiresAt := time.Now().Add(config.TIMEOUT_VERIFICATION_TOKEN)
	_, err = tx.Exec(
		`UPDATE verification_token SET id = $1, token = $2, expires_at = $3, resend_count = resend_count + 1 WHERE user_id = $4 AND usage_id = $5`,
		newTokenID, newToken, expiresAt, userID, usageID,
	)
	if err != nil {
		return "", "", err
	}

	if err := tx.Commit(); err != nil {
		return "", "", err
	}

	return newTokenID, newToken, nil
}
