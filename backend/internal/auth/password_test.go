package auth

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword_IsBcrypt(t *testing.T) {
	hash, err := HashPassword("supersecret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash == "supersecret" {
		t.Fatal("password was not hashed")
	}
	if !strings.HasPrefix(hash, "$2") {
		t.Fatalf("hash does not look like bcrypt: %q", hash)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("supersecret")); err != nil {
		t.Fatalf("hash does not verify the original password: %v", err)
	}
}

func TestCheckPassword(t *testing.T) {
	hash, err := HashPassword("supersecret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !CheckPassword(hash, "supersecret") {
		t.Error("expected the correct password to verify")
	}
	if CheckPassword(hash, "wrong-password") {
		t.Error("expected a wrong password to be rejected")
	}
	if CheckPassword("not-a-hash", "supersecret") {
		t.Error("expected an invalid hash to be rejected")
	}
}

func TestDummyCheck_DoesNotPanic(t *testing.T) {
	DummyCheck("anything")
}
