package config

import (
	"os"
	"testing"
)

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
