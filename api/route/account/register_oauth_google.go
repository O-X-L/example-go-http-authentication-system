package account

import (
	"context"
	"encoding/json"
	"example_api/base/db"
	"example_api/config"
	"example_api/types"
	"example_api/util"
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"google.golang.org/api/idtoken"
)

type OAuthGoogleRequest struct {
	IDToken string `json:"id_token" validate:"required"`
}

// verifyGoogleIDToken validates the JWT signature and extracts the email.
func verifyGoogleIDToken(token string) (string, error) {
	ctx := context.Background()

	// Validate the token. This checks the signature, expiration, and audience (Client ID).
	payload, err := idtoken.Validate(ctx, token, config.GetGoogleOAuthClientID())
	if err != nil {
		return "", fmt.Errorf("invalid google token: %v", err)
	}

	// Extract the email claim
	emailClaim, ok := payload.Claims["email"]
	if !ok {
		return "", fmt.Errorf("email not found in token claims")
	}

	email, ok := emailClaim.(string)
	if !ok || email == "" {
		return "", fmt.Errorf("invalid email format in token claims")
	}

	// Optional: You can also check if the email is verified by Google
	// emailVerified, _ := payload.Claims["email_verified"].(bool)
	// if !emailVerified { return "", fmt.Errorf("email not verified by google") }

	return email, nil
}

// VerifyGoogleIDTokenFunc allows us to mock the verification in our unit tests
var VerifyGoogleIDTokenFunc = verifyGoogleIDToken

// HandleRegisterOAuthGoogle godoc
// @Summary      Register a new user via Google OAuth
// @Router       /a/register/oauth/google [post]
func HandleRegisterOAuthGoogle(store *db.DataStore, v *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req OAuthGoogleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if err := v.Struct(req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		email, err := VerifyGoogleIDTokenFunc(req.IDToken)
		if err != nil || email == "" {
			http.Error(w, "Invalid Google token", http.StatusUnauthorized)
			return
		}

		if err := store.Auth.AddUserOAuth(email, db.AUTH_TYPE_OAUTH_GOOGLE); err != nil {
			http.Error(w, "Could not register user. Email might already exist.", http.StatusConflict)
			return
		}

		go util.SendRegistrationWelcomeEmail(email)

		user, err := store.Auth.GetUserByEmail(email, false)
		if err != nil {
			http.Error(w, "Failed to perform login", http.StatusInternalServerError)
			return
		}

		err = performLogin(w, r, store, user.ID, db.AUTH_TYPE_OAUTH_GOOGLE)
		if err != nil {
			log.Printf("failed to login newly registered user: %v (authType %d)", err, db.AUTH_TYPE_OAUTH_GOOGLE)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(types.MessageResponse{Message: "User registered successfully using Google-OAuth"})
	}
}
