package account

import (
	"encoding/json"
	"example_api/base/db"
	"example_api/types"
	"net/http"

	"github.com/go-playground/validator/v10"
)

// HandleLoginOAuthGoogle godoc
// @Summary      Login user via Google OAuth
// @Router       /a/session/login/oauth/google [post]
func HandleLoginOAuthGoogle(store *db.DataStore, v *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req OAuthGoogleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if err := v.Struct(req); err != nil {
			http.Error(w, "Invalid request format", http.StatusBadRequest)
			return
		}

		email, err := VerifyGoogleIDTokenFunc(req.IDToken)
		if err != nil || email == "" {
			http.Error(w, "Invalid Google token", http.StatusUnauthorized)
			return
		}

		user, err := store.Auth.GetUserByEmail(email, false)
		if err != nil {
			http.Error(w, "User not found. Please register first.", http.StatusUnauthorized)
			return
		}

		// Ensure the user actually registered via Google
		if user.AuthType != db.AUTH_TYPE_OAUTH_GOOGLE {
			http.Error(w, "Google OAuth not enabled for user.", http.StatusUnauthorized)
			return
		}

		if !user.IsActive {
			http.Error(w, "Account is inactive", http.StatusForbidden)
			return
		}

		err = performLogin(w, r, store, user.ID, db.AUTH_TYPE_OAUTH_GOOGLE)
		if err != nil {
			http.Error(w, "Login failed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(types.MessageResponse{Message: "Logged in successfully via Google"})
	}
}
