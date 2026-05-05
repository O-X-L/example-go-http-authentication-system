package base

import (
	"log"
	"net/http"

	"example_api/base/db"
	"example_api/config"
	"example_api/util"

	"github.com/go-playground/validator/v10"
)

// GuestMiddleware handles CSRF for unauthenticated routes.
func GuestMiddleware(store *db.DataStore, v *validator.Validate, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// check if user-session exists and is valid; else fallback to guest-session (edge cases)
		cookieUserSession, err := r.Cookie(config.COOKIE_SESSION)
		if err == nil {
			validUserSession := true
			if err := v.Var(cookieUserSession.Value, "required,len=43,base64rawurl"); err != nil {
				validUserSession = false
			}

			if validUserSession {
				_, err := store.Auth.GetSession(cookieUserSession.Value)
				if err != nil {
					validUserSession = false
				}
			}

			// if client has valid user-session; jump to 'normal' auth-checks
			if validUserSession {
				authHandler := AuthMiddleware(store, v, next)
				authHandler(w, r)
				return
			}
		}

		// If it's a GET request let them through
		if r.Method == http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		// For mutating requests (POST, etc.), validate the token

		// Validate guest-session
		cookieGuestSession, err := r.Cookie(config.COOKIE_SESSION_PREAUTH)
		if err != nil {
			http.Error(w, "Failed to load guest-session", http.StatusInternalServerError)
			return
		}
		guestSessionID := cookieGuestSession.Value

		clientIP := util.GetClientIP(r)
		dbSession, err := store.Auth.GetGuestSession(guestSessionID, clientIP)
		if err != nil {
			http.Error(w, "Invalid or expired session", http.StatusUnauthorized)
			log.Printf("guest session error: %v (ip: %s)", err, clientIP)
			return
		}

		// Validate the CSRF token format
		reqCSRFToken := r.Header.Get(config.HEADER_CSRF_PREAUTH)
		if err := v.Var(reqCSRFToken, "required,len=43,base64rawurl"); err != nil {
			http.Error(w, "Malformed CSRF token", http.StatusBadRequest)
			return
		}

		// Validate CSRF-token
		csrfTokenMatches := dbSession.CSRFMatches(reqCSRFToken)
		if reqCSRFToken == "" || !csrfTokenMatches {
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// AuthMiddleware validates the session cookie and CSRF token.
func AuthMiddleware(store *db.DataStore, v *validator.Validate, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookieUserSession, err := r.Cookie(config.COOKIE_SESSION)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Validate the Session ID format
		if err := v.Var(cookieUserSession.Value, "required,len=43,base64rawurl"); err != nil {
			http.Error(w, "Malformed session token", http.StatusBadRequest)
			return
		}

		// Validate session
		dbSession, err := store.Auth.GetSession(cookieUserSession.Value)
		if err != nil {
			http.Error(w, "Invalid or expired session", http.StatusUnauthorized)
			log.Printf("user session error: %v", err)
			return
		}

		if dbSession.ShouldRotateToken() {
			newSession := db.AuthSession{}
			newSession.FromExisting(dbSession)
			err = store.Auth.RotateSession(dbSession, newSession)
			if err != nil {
				log.Printf("user session-rotation error: %v", err)
			} else {
				http.SetCookie(w, &http.Cookie{
					Name:     config.COOKIE_SESSION,
					Value:    newSession.ID,
					Path:     "/",
					Expires:  newSession.ExpiresAt,
					HttpOnly: true,
					Secure:   config.IsDeploymentProduction(),
					SameSite: http.SameSiteLaxMode,
				})
				http.SetCookie(w, &http.Cookie{
					Name:     config.COOKIE_CSRF_POSTAUTH,
					Value:    newSession.CSRFToken,
					Path:     "/",
					Expires:  newSession.ExpiresAt,
					HttpOnly: false,
					Secure:   config.IsDeploymentProduction(),
					SameSite: http.SameSiteLaxMode,
				})
			}
		}

		// If it's a GET request let them through
		if r.Method == http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		// Validate the CSRF token format
		reqCSRFToken := r.Header.Get(config.HEADER_CSRF_POSTAUTH)
		if err := v.Var(reqCSRFToken, "required,len=43,base64rawurl"); err != nil {
			http.Error(w, "Malformed CSRF token", http.StatusBadRequest)
			return
		}

		// Check CSRF token from header for mutating requests
		csrfTokenMatches := dbSession.CSRFMatches(reqCSRFToken)
		if reqCSRFToken == "" || !csrfTokenMatches {
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// CORSMiddleware adds CORS response-headers that are required by the frontend.
func CORSMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", config.GetDomainFE())
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token, X-CSRF-Token-Guest")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
