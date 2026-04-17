package auth

import "testing"

func TestHashPasswordAndCheckPassword(t *testing.T) {
	password := "test-password-123"

	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if hashedPassword == "" {
		t.Fatal("expected hashed password to be non-empty")
	}
	if hashedPassword == password {
		t.Fatal("expected hashed password to differ from original password")
	}

	if err := CheckPassword(hashedPassword, password); err != nil {
		t.Fatalf("expected password check to succeed, got %v", err)
	}
}

func TestCheckPasswordMismatch(t *testing.T) {
	hashedPassword, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if err := CheckPassword(hashedPassword, "wrong-password"); err == nil {
		t.Fatal("expected password check to fail for wrong password")
	}
}
