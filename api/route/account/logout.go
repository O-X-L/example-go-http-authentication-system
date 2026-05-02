package account

import (
	"encoding/json"
	"net/http"

	"example_api/base/db"
	"example_api/config"
	"example_api/types"
)

// HandleLogout godoc
// @Summary      Logout user
// @Router       /a/session/logout [delete]
func HandleLogout(store *db.DataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, _ := r.Cookie(config.COOKIE_SESSION)

		if err := store.Auth.DeleteSession(sessionCookie.Value); err != nil {
			http.Error(w, "Failed to delete session", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     config.COOKIE_SESSION,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
		})
		http.SetCookie(w, &http.Cookie{
			Name:     config.COOKIE_CSRF_POSTAUTH,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: false,
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(types.MessageResponse{Message: "Logged out successfully"})
	}
}
