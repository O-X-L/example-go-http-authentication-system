package util

import (
	"net/smtp"
	"os"
	"testing"
)

func TestSendRegistrationVerificationEmail(t *testing.T) {
	// 1. Setup mock environment variables for the test
	os.Setenv("APP_SMTP_EMAIL", "sender@example.com")
	os.Setenv("APP_SMTP_SERVER", "smtp.example.com:587")
	os.Setenv("APP_SMTP_USER", "smtp_user")
	os.Setenv("APP_SMTP_PASSWORD", "smtp_pass")

	// Cleanup env vars after the test
	defer func() {
		os.Unsetenv("APP_SMTP_EMAIL")
		os.Unsetenv("APP_SMTP_SERVER")
		os.Unsetenv("APP_SMTP_USER")
		os.Unsetenv("APP_SMTP_PASSWORD")
	}()

	var mockCalled bool

	// 2. Mock the actual SendMail function
	originalSendMail := SendMailFunc
	defer func() { SendMailFunc = originalSendMail }() // restore after test

	SendMailFunc = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		mockCalled = true

		if addr != "smtp.example.com:587" {
			t.Errorf("Expected address smtp.example.com:587, got %s", addr)
		}
		if from != "sender@example.com" {
			t.Errorf("Expected from sender@example.com, got %s", from)
		}
		if len(to) != 1 || to[0] != "newuser@example.com" {
			t.Errorf("Expected to contain newuser@example.com, got %v", to)
		}

		return nil
	}

	// 3. Trigger the function
	SendRegistrationVerificationEmail("newuser@example.com", "https://localhost:8080/a/verify?id=aaa&token=bbb")

	// 4. Assert it was successfully called
	if !mockCalled {
		t.Error("Expected SendMailFunc to be called but it was not")
	}
}
