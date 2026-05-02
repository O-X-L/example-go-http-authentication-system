package base

import (
	"log"
	"net/http"
	"time"

	"example_fe/base/db"
	"example_fe/config"
	"example_fe/util"

	"github.com/go-playground/validator/v10"
)

// GuestMiddleware handles CSRF for unauthenticated routes.
func guestMiddleware(store *db.DataStore, v *validator.Validate, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// ensure they have a guest-session
		_, err := ensureGuestSession(store, w, r)
		if err != nil {
			log.Printf("failed to create guest-session")
		}
		next.ServeHTTP(w, r)
	}
}

func ensureGuestSession(store *db.DataStore, w http.ResponseWriter, r *http.Request) (string, error) {
	if _, err := r.Cookie(config.COOKIE_SESSION_PREAUTH); err == nil {
		return "", nil
	}

	sessionID := util.GenerateToken32()
	clientIP := util.GetClientIP(r)
	csrfToken := util.GenerateToken32()
	expiresAt := time.Now().Add(config.TIMEOUT_SESSION_GUEST)

	if err := store.Auth.AddGuestSession(sessionID, clientIP, csrfToken, expiresAt); err != nil {
		return "", err
	}

	cookieDomain := "localhost"
	if config.IsDeploymentProduction() {
		cookieDomain = ".example.oxl.app"
	}
	http.SetCookie(w, &http.Cookie{
		Name:     config.COOKIE_SESSION_PREAUTH,
		Value:    sessionID,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   config.IsDeploymentProduction(),
		SameSite: http.SameSiteLaxMode,
		Domain:   cookieDomain,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     config.COOKIE_CSRF_PREAUTH,
		Value:    csrfToken,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: false,
		Secure:   config.IsDeploymentProduction(),
		SameSite: http.SameSiteLaxMode,
	})
	return sessionID, nil
}
