package account_util

import (
	"crypto/subtle"
	"fmt"
	"example_api/base/db"
	"example_api/config"
	"example_api/util"
	"log"
	"time"
)

var MAPPING_VERIFICATION_FRONTEND_LOCATION = map[int]string{
	db.VERIFICATION_LINK_USAGE_VERIFY_EMAIL:   "/a/verify/email",
	db.VERIFICATION_LINK_USAGE_PASSWORD_RESET: "/a/verify/password_reset",
}
var MAPPING_VERIFICATION_LOG_NAME = map[int]string{
	db.VERIFICATION_LINK_USAGE_VERIFY_EMAIL:   "email-verification",
	db.VERIFICATION_LINK_USAGE_PASSWORD_RESET: "password-reset",
}

func AddVerificationToken(store *db.DataStore, userID, usageID int) (string, string, error) {
	verifyTokenID := util.GenerateToken32()
	verifyToken := util.GenerateToken32()
	err := store.Auth.AddVerificationToken(userID, usageID, verifyTokenID, verifyToken)
	return verifyTokenID, verifyToken, err
}

func generateVerificationLinkWithToken(tokenID, token string, usageID int) (string, error) {
	location := MAPPING_VERIFICATION_FRONTEND_LOCATION[usageID]
	return fmt.Sprintf("%s%s?id=%s&token=%s", config.GetDomainFE(), location, tokenID, token), nil
}

func IsValidVerificationToken(token, dbToken string, expiresAt time.Time) bool {
	expired := time.Now().After(expiresAt)
	matching := subtle.ConstantTimeCompare([]byte(token), []byte(dbToken)) == 1
	return !expired && matching
}

func logEmailSendError(usageID int, email string, err error) {
	if err != nil {
		log.Printf("failed to send %s email to '%s': %v", MAPPING_VERIFICATION_LOG_NAME[usageID], email, err)
	}
}

func SendEmailVerificationTokenPerEmail(verifyTokenID, verifyToken string, user *db.AuthUser, usageID int) {
	verifyLink, err := generateVerificationLinkWithToken(verifyTokenID, verifyToken, usageID)
	if err != nil {
		logEmailSendError(usageID, user.Email, err)
		return
	}
	if usageID == db.VERIFICATION_LINK_USAGE_PASSWORD_RESET {
		err = util.SendPasswordResetVerificationEmail(user.Email, verifyLink)
	} else {
		err = util.SendRegistrationVerificationEmail(user.Email, verifyLink)
	}
	if err != nil {
		logEmailSendError(usageID, user.Email, err)
	}
}
