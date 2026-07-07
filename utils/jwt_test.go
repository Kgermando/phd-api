package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
)

func loadTestEnv(t *testing.T) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	envPath := filepath.Join(wd, "..", ".env")
	if err := godotenv.Load(envPath); err != nil {
		t.Fatalf("load .env: %v", err)
	}
}

func TestJwtRoundTrip(t *testing.T) {
	loadTestEnv(t)

	token, err := GenerateJwt("test-user-uuid")
	if err != nil {
		t.Fatalf("GenerateJwt: %v", err)
	}

	issuer, err := VerifyJwt(token)
	if err != nil {
		t.Fatalf("VerifyJwt: %v", err)
	}
	if issuer != "test-user-uuid" {
		t.Fatalf("expected test-user-uuid, got %q", issuer)
	}
}
