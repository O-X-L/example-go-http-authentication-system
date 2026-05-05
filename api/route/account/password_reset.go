package account

import (
	"encoding/json"
	"log"
	"net/http"

	"example_api/base/db"
	"example_api/types"
	"example_api/util"
	account_util "example_api/util/account"

	"github.com/go-playground/validator/v10"
)

type PasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// HandlePasswordResetRequest godoc
// @Summary      Request a password reset email
// @Router       /a/password_reset/request [post]
func HandlePasswordResetRequest(store *db.DataStore, v *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PasswordResetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if err := v.Struct(req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		user, err := store.Auth.GetUserByEmail(req.Email, false)

		// Security: If the user doesn't exist, we STILL return a success message.
		// This prevents attackers from using this endpoint to enumerate valid email addresses.
		if err == nil && user.IsActive && user.AuthType == db.AUTH_TYPE_BASIC {
			verifyTokenID, verifyToken, err := account_util.AddVerificationToken(store, user.ID, db.VERIFICATION_LINK_USAGE_PASSWORD_RESET)
			if err != nil {
				log.Printf("failed to create token for password reset: %v", err)
			} else {
				go account_util.SendEmailVerificationTokenPerEmail(verifyTokenID, verifyToken, &user, db.VERIFICATION_LINK_USAGE_PASSWORD_RESET)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(types.MessageResponse{
			Message: "If the email is registered, a reset link has been sent.",
		})
	}
}

type PasswordResetConfirmRequest struct {
	TokenID     string `json:"token_id" validate:"required,len=43,base64rawurl"`
	Token       string `json:"token" validate:"required,len=43,base64rawurl"`
	NewPassword string `json:"password" validate:"required,min=8"`
}

// HandlePasswordResetConfirm godoc
// @Summary      Confirm and execute password reset
// @Router       /a/password_reset/confirm [post]
func HandlePasswordResetConfirm(store *db.DataStore, v *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PasswordResetConfirmRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if err := v.Struct(req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		userID, usageID, dbToken, expiresAt, err := store.Auth.GetVerificationToken(req.TokenID)
		if err != nil || dbToken == "" || usageID != db.VERIFICATION_LINK_USAGE_PASSWORD_RESET {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		valid := account_util.IsValidVerificationToken(req.Token, dbToken, expiresAt)
		if !valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		passwordHash, err := util.Argon2HashSecret(req.NewPassword)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if err := store.Auth.UpdateUserPassword(userID, string(passwordHash)); err != nil {
			http.Error(w, "Failed to update password", http.StatusInternalServerError)
			return
		}

		store.Auth.DeleteVerificationToken(req.TokenID)
		store.Auth.DeleteAllUserSessions(userID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(types.MessageResponse{
			Message: "Password has been successfully reset",
		})
	}
}
