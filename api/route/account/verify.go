package account

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"

	"example_api/base/db"
	"example_api/config"
	"example_api/types"
	account_util "example_api/util/account"
)

type VerifyRequest struct {
	TokenID string `json:"id" validate:"required,len=43,base64rawurl"`
	Token   string `json:"token" validate:"required,len=43,base64rawurl"`
}

// HandleVerify godoc
// @Summary      Validate account-verification-token
// @Router       /a/verify [post]
func HandleVerify(store *db.DataStore, v *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req VerifyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if err := v.Struct(req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		userID, usageID, dbToken, expiresAt, err := store.Auth.GetVerificationToken(req.TokenID)
		if err != nil || dbToken == "" {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		valid := account_util.IsValidVerificationToken(req.Token, dbToken, expiresAt)
		if !valid {
			http.Error(w, "Token mismatch or expired", http.StatusUnauthorized)
			return
		}

		if usageID == db.VERIFICATION_LINK_USAGE_VERIFY_EMAIL {
			store.Auth.SetVerificationDoneUserEmail(userID)
		}
		// todo: password_reset

		store.Auth.DeleteVerificationToken(req.TokenID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(types.MessageResponse{
			Message: "Verified successfully",
		})
	}
}

type VerifyResendRequest struct {
	Usage string `json:"usage" validate:"required"`
}

var MAPPING_RESEND_USAGE = map[string]int{
	"email": db.VERIFICATION_LINK_USAGE_VERIFY_EMAIL,
}

// HandleVerify godoc
// @Summary      Resend verification-token via email
// @Router       /a/verify_resend [post]
func HandleVerifyResend(store *db.DataStore, v *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, _ := r.Cookie(config.COOKIE_SESSION)

		user, err := store.Auth.GetUserBySessionID(sessionCookie.Value, false)
		if err != nil {
			http.Error(w, "Session invalid", http.StatusUnauthorized)
			return
		}

		var req VerifyResendRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		err = v.Struct(req)
		usageID, valid := MAPPING_RESEND_USAGE[req.Usage]
		if err != nil || !valid {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		tokenID, token, err := store.Auth.RenewVerificationTokenForResend(user.ID, usageID)
		if err != nil {
			if strings.Contains(fmt.Sprintf("%v", err), "limit exceeded") {
				http.Error(w, "Token renewal exceeded", http.StatusTooManyRequests)
			} else {
				http.Error(w, "Failed to renew verification-token", http.StatusInternalServerError)
			}
			return
		}

		if usageID == db.VERIFICATION_LINK_USAGE_VERIFY_EMAIL {
			go account_util.SendEmailVerificationTokenPerEmail(tokenID, token, &user)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(types.MessageResponse{
			Message: "Verified successfully",
		})
	}
}
