package util

import "testing"

func TestGenerateToken(t *testing.T) {
	token1 := GenerateToken(32)
	token2 := GenerateToken32()

	if token1 == "" || token2 == "" {
		t.Error("GenerateToken returned an empty string")
	}

	token1 = GenerateToken(32)
	token2 = GenerateToken(32)

	if token1 == token2 {
		t.Errorf("GenerateToken generated the same token twice: %s", token1)
	}
}
