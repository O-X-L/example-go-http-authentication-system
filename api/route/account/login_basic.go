package account

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"

	"example_api/base/db"
	"example_api/config"
	"example_api/types"
	"example_api/util"
)

func performLogin(w http.ResponseWriter, r *http.Request, store *db.DataStore, userID int, authType int) error {
	if err := store.Auth.EnforceSessionPerUserLimit(userID); err != nil {
		log.Printf("Failed to enforce user-session limit (authType %d): %v", authType, err)
	}

	session := db.AuthSession{UserID: userID}
	session.New()

	if err := store.Auth.AddSession(session); err != nil {
		return fmt.Errorf("Failed to create session")
	}

	http.SetCookie(w, &http.Cookie{
		Name:     config.COOKIE_SESSION,
		Value:    session.ID,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   config.IsDeploymentProduction(),
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     config.COOKIE_CSRF_POSTAUTH,
		Value:    session.CSRFToken,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: false,
		Secure:   config.IsDeploymentProduction(),
		SameSite: http.SameSiteLaxMode,
	})

	userAgent := r.Header.Get("User-Agent")
	err := store.Auth.RegisterUserLogon(userID, authType, db.LOGON_DEVICE_TYPE_WEB, userAgent)
	if err != nil {
		log.Printf("Failed to log user-logon (authType %d): %v", authType, err)
	}
	return nil
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=254"`
	Password string `json:"password" validate:"required"`
}

// HandleLogin godoc
// @Summary      Login user
// @Router       /a/session/login/basic [post]
func HandleLogin(store *db.DataStore, v *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if err := v.Struct(req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		user, err := store.Auth.GetUserByEmail(req.Email, false)
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		if !user.IsActive {
			http.Error(w, "Account is inactive", http.StatusForbidden)
			return
		}

		correctPassword, err := util.Argon2CompareSecret(req.Password, user.PasswordHash)
		if err != nil || !correctPassword {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		err = performLogin(w, r, store, user.ID, db.AUTH_TYPE_BASIC)
		if err != nil {
			http.Error(w, "Login failed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(types.MessageResponse{
			Message: "Logged in successfully",
		})
	}
}
