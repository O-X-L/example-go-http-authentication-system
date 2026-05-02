package config

import (
	"os"
	"testing"
)

func TestEnsureValidConfig(t *testing.T) {
	originalPepper := os.Getenv(ENV_PEPPER)
	defer os.Setenv(ENV_PEPPER, originalPepper) // Cleanup after test

	tests := []struct {
		name       string
		pepperVal  string
		wantResult bool
	}{
		{"Valid Pepper Length", "this-is-a-very-long-pepper-value", true},
		{"Empty Pepper", "", false},
		{"Short Pepper", "short", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(ENV_PEPPER, tt.pepperVal)
			if got := EnsureValidConfig(); got != tt.wantResult {
				t.Errorf("EnsureValidConfig() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func TestIsDeploymentProduction(t *testing.T) {
	originalDev := os.Getenv(ENV_MODE_DEV)
	defer os.Setenv(ENV_MODE_DEV, originalDev)

	// Test Dev Mode
	os.Setenv(ENV_MODE_DEV, "1")
	if IsDeploymentProduction() {
		t.Error("Expected IsDeploymentProduction to be false when APP_DEV is set")
	}

	// Test Prod Mode (empty APP_DEV)
	os.Setenv(ENV_MODE_DEV, "")
	if !IsDeploymentProduction() {
		t.Error("Expected IsDeploymentProduction to be true when APP_DEV is empty")
	}
}
