package account

import (
	"encoding/json"
	"example_api/base/db"
	"example_api/config"
	"net/http"
)

type UserStatusResponse struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

// HandleStatus godoc
// @Summary      Validate account-verification-token
// @Router       /a/status [get]
func HandleStatus(store *db.DataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, _ := r.Cookie(config.COOKIE_SESSION)

		user, err := store.Auth.GetUserBySessionID(sessionCookie.Value, false)
		if err != nil {
			http.Error(w, "Session invalid", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(UserStatusResponse{
			Email:         user.Email,
			EmailVerified: user.EmailVerified,
		})
	}
}
