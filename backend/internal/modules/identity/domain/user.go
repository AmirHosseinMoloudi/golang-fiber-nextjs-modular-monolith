package domain

import (
	"strings"
	"time"

	"github.com/golang-todo/modular-monolith/internal/core/domain"
	"github.com/golang-todo/modular-monolith/internal/core/errors"
	"golang.org/x/crypto/bcrypt"
)

type Role string

const (
	RoleAdmin   Role = "ADMIN"
	RoleMember  Role = "MEMBER"
	RoleAuditor Role = "AUDITOR"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleMember, RoleAuditor:
		return true
	default:
		return false
	}
}

type User struct {
	domain.AggregateRoot
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	Role         Role      `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func NewUser(id, tenantID, email, rawPassword, fullName string, role Role) (*User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || !strings.Contains(email, "@") {
		return nil, errors.NewValidationError("invalid email address format", nil)
	}
	if len(rawPassword) < 8 {
		return nil, errors.NewValidationError("password must be at least 8 characters long", nil)
	}
	if strings.TrimSpace(fullName) == "" {
		return nil, errors.NewValidationError("full name is required", nil)
	}
	if !role.IsValid() {
		role = RoleMember
	}
	if tenantID == "" {
		tenantID = "default-tenant"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.NewInternalError("failed to securely hash credentials")
	}

	user := &User{
		ID:           id,
		TenantID:     tenantID,
		Email:        email,
		PasswordHash: string(hash),
		FullName:     fullName,
		Role:         role,
		IsActive:     true,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	user.RecordEvent(domain.NewBaseDomainEvent("USER_REGISTERED", user.ID))
	return user, nil
}

func (u *User) VerifyPassword(rawPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(rawPassword))
	return err == nil
}

type UserRepository interface {
	Save(user *User) error
	FindByID(id string) (*User, error)
	FindByEmail(email string) (*User, error)
	List(tenantID string) ([]*User, error)
}
