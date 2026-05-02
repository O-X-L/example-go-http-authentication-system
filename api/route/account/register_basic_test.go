package account

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/smtp"
	"testing"
	"time"

	"example_api/base/db"
	"example_api/types"
	"example_api/util"

	"github.com/go-playground/validator/v10"
)

func TestRegisterRequestValidation(t *testing.T) {
	v := validator.New()

	tests := []struct {
		name    string
		payload RegisterRequest
		wantErr bool
	}{
		{"Valid Request", RegisterRequest{Email: "test@example.com", Password: "password123"}, false},
		{"Invalid Email", RegisterRequest{Email: "not-an-email", Password: "password123"}, true},
		{"Password Too Short", RegisterRequest{Email: "test@example.com", Password: "short"}, true},
		{"Empty Fields", RegisterRequest{Email: "", Password: ""}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Struct(tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterRequest validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHandleRegister_InvalidJSON(t *testing.T) {
	v := validator.New()
	handler := HandleRegister(nil, v) // dbConn nil for this specific test

	req, _ := http.NewRequest("POST", "/a/register/basic", bytes.NewBuffer([]byte(`{invalid-json}`)))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestHandleRegister_Success(t *testing.T) {
	mockAuth := &db.TestMockAuthStore{
		AddUserBasicFunc: func(email, passwordHash string) error {
			return nil
		},
	}

	// mock email sending
	registrationEmailSent := false
	originalSendMail := util.SendMailFunc
	defer func() { util.SendMailFunc = originalSendMail }() // restore after test
	util.SendMailFunc = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		registrationEmailSent = true
		return nil
	}

	store := &db.DataStore{Auth: mockAuth}
	v := validator.New()
	handler := HandleRegister(store, v)

	body := RegisterRequest{Email: "newuser@example.com", Password: "strongpassword123"}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/register/basic", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var response types.MessageResponse
	json.NewDecoder(rr.Body).Decode(&response)
	if response.Message != "User registered successfully" {
		t.Errorf("unexpected response message: %s", response.Message)
	}

	time.Sleep(time.Millisecond * 100) // wait for async function to be called
	if !registrationEmailSent {
		t.Errorf("expected registration email to be sent")
	}
}

func TestHandleRegister_Conflict(t *testing.T) {
	mockAuth := &db.TestMockAuthStore{
		AddUserBasicFunc: func(email, passwordHash string) error {
			return sql.ErrNoRows // Simulate uniqueness conflict
		},
	}

	store := &db.DataStore{Auth: mockAuth}
	v := validator.New()
	handler := HandleRegister(store, v)

	body := RegisterRequest{Email: "existing@example.com", Password: "strongpassword123"}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/register/basic", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusConflict {
		t.Errorf("expected status %v for existing email, got %v", http.StatusConflict, status)
	}
}
