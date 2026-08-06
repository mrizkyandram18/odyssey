package auth

import (
	"testing"
)

func TestBcryptHasher_HashAndVerify(t *testing.T) {
	h := NewBcryptHasher()
	hashed, err := h.Hash("password123")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if hashed == "" {
		t.Fatal("expected non-empty hash")
	}
	if hashed == "password123" {
		t.Fatal("hash should not equal plaintext")
	}
	if err := h.Verify(hashed, "password123"); err != nil {
		t.Fatalf("Verify valid: %v", err)
	}
}

func TestBcryptHasher_RejectsWrongPassword(t *testing.T) {
	h := NewBcryptHasher()
	hashed, err := h.Hash("password123")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if err := h.Verify(hashed, "wrong"); err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestBcryptHasher_RejectsEmptyHash(t *testing.T) {
	h := NewBcryptHasher()
	if err := h.Verify("", "password123"); err == nil {
		t.Fatal("expected error for empty hash")
	}
}

func TestBcryptHasher_HashIsSalted(t *testing.T) {
	h := NewBcryptHasher()
	h1, err := h.Hash("password123")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	h2, err := h.Hash("password123")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if h1 == h2 {
		t.Fatal("expected different hashes due to salting")
	}
}

func TestBcryptHasher_ImplementsPasswordHasher(t *testing.T) {
	var _ PasswordHasher = (*bcryptHasher)(nil)
}
