package domain_test

import (
	"testing"

	"github.com/golang-todo/modular-monolith/internal/modules/identity/domain"
)

func TestUser_CreationAndAuth(t *testing.T) {
	user, err := domain.NewUser("uid-1", "tenant-1", "test@example.com", "SecurePass123!", "Alice Smith", domain.RoleMember)
	if err != nil {
		t.Fatalf("unexpected error creating user: %v", err)
	}

	if !user.VerifyPassword("SecurePass123!") {
		t.Errorf("expected valid password verification")
	}

	if user.VerifyPassword("WrongPass") {
		t.Errorf("expected invalid password to fail")
	}

	// Invariant: password must be >= 8 chars
	_, errShort := domain.NewUser("uid-2", "tenant-1", "bob@example.com", "short", "Bob", domain.RoleMember)
	if errShort == nil {
		t.Errorf("expected error for password < 8 characters")
	}
}
