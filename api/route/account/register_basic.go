package account

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"

	"example_api/base/db"
	"example_api/types"
	"example_api/util"
	account_util "example_api/util/account"
)

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// HandleRegister godoc
// @Summary      Register a new user
// @Router       /register/basic [post]
func HandleRegister(store *db.DataStore, v *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if err := v.Struct(req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		hash, err := util.Argon2HashSecret(req.Password)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if err := store.Auth.AddUserBasic(req.Email, string(hash)); err != nil {
			http.Error(w, "Could not register user. Email might already exist.", http.StatusConflict)
			return
		}

		user, err := store.Auth.GetUserByEmail(req.Email, false)
		if err != nil {
			log.Printf("failed to query user-info for email-verification send: %v", err)

		} else {
			verifyTokenID, verifyToken, err := account_util.AddVerificationToken(store, user.ID, db.VERIFICATION_LINK_USAGE_VERIFY_EMAIL)
			if err != nil {
				log.Printf("failed to create token for email-verification send: %v", err)
			}
			go account_util.SendEmailVerificationTokenPerEmail(verifyTokenID, verifyToken, &user)
		}

		err = performLogin(w, r, store, user.ID, db.AUTH_TYPE_BASIC)
		if err != nil {
			log.Printf("failed to login newly registered user: %v (authType %d)", err, db.AUTH_TYPE_BASIC)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(types.MessageResponse{Message: "User registered successfully"})
	}
}
