package account

import (
	"encoding/json"
	"net/http"

	"example_api/base/db"
	"example_api/config"
	"example_api/types"
)

// HandleDelete godoc
// @Summary      Delete user profile
// @Router      /a/delete [delete]
func HandleDelete(store *db.DataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, _ := r.Cookie(config.COOKIE_SESSION)

		user, err := store.Auth.GetUserBySessionID(sessionCookie.Value, false)
		if err != nil {
			http.Error(w, "Session invalid", http.StatusUnauthorized)
			return
		}

		// todo: maybe send verification email

		if err := store.Auth.DeleteUser(user.ID); err != nil {
			http.Error(w, "Failed to delete user profile", http.StatusInternalServerError)
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
		json.NewEncoder(w).Encode(types.MessageResponse{Message: "User profile deleted successfully"})
	}
}
