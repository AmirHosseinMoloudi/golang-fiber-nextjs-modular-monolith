package application

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang-todo/modular-monolith/internal/core/errors"
	"github.com/golang-todo/modular-monolith/internal/modules/identity/domain"
	"github.com/google/uuid"
)

type JWTClaims struct {
	UserID   string      `json:"user_id"`
	TenantID string      `json:"tenant_id"`
	Email    string      `json:"email"`
	FullName string      `json:"full_name"`
	Role     domain.Role `json:"role"`
	jwt.RegisteredClaims
}

type TokenService interface {
	GenerateToken(user *domain.User) (string, error)
	ValidateToken(tokenString string) (*JWTClaims, error)
}

type jwtTokenService struct {
	secretKey []byte
}

func NewJWTTokenService(secret string) TokenService {
	return &jwtTokenService{secretKey: []byte(secret)}
}

func (s *jwtTokenService) GenerateToken(user *domain.User) (string, error) {
	claims := JWTClaims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Email:    user.Email,
		FullName: user.FullName,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			Issuer:    "golang-todo-modular-monolith",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

func (s *jwtTokenService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.NewUnauthorizedError("unexpected signing method")
		}
		return s.secretKey, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.NewUnauthorizedError("invalid or expired authorization token")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, errors.NewUnauthorizedError("malformed token claims")
	}
	return claims, nil
}

type IdentityService struct {
	userRepo     domain.UserRepository
	tokenService TokenService
}

func NewIdentityService(repo domain.UserRepository, tokenService TokenService) *IdentityService {
	return &IdentityService{
		userRepo:     repo,
		tokenService: tokenService,
	}
}

type RegisterRequest struct {
	Email    string      `json:"email"`
	Password string      `json:"password"`
	FullName string      `json:"full_name"`
	Role     domain.Role `json:"role"`
	TenantID string      `json:"tenant_id"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserDTO      `json:"user"`
}

type UserDTO struct {
	ID       string      `json:"id"`
	TenantID string      `json:"tenant_id"`
	Email    string      `json:"email"`
	FullName string      `json:"full_name"`
	Role     domain.Role `json:"role"`
}

func (s *IdentityService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	existing, _ := s.userRepo.FindByEmail(req.Email)
	if existing != nil {
		return nil, errors.NewConflictError("user with this email already exists")
	}

	userID := uuid.New().String()
	user, err := domain.NewUser(userID, req.TenantID, req.Email, req.Password, req.FullName, req.Role)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.Save(user); err != nil {
		return nil, errors.NewInternalError("failed to persist user")
	}

	token, err := s.tokenService.GenerateToken(user)
	if err != nil {
		return nil, errors.NewInternalError("failed to generate access token")
	}

	return &AuthResponse{
		Token: token,
		User: UserDTO{
			ID:       user.ID,
			TenantID: user.TenantID,
			Email:    user.Email,
			FullName: user.FullName,
			Role:     user.Role,
		},
	}, nil
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *IdentityService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil || user == nil {
		return nil, errors.NewUnauthorizedError("invalid credentials provided")
	}

	if !user.VerifyPassword(req.Password) {
		return nil, errors.NewUnauthorizedError("invalid credentials provided")
	}

	token, err := s.tokenService.GenerateToken(user)
	if err != nil {
		return nil, errors.NewInternalError("failed to generate access token")
	}

	return &AuthResponse{
		Token: token,
		User: UserDTO{
			ID:       user.ID,
			TenantID: user.TenantID,
			Email:    user.Email,
			FullName: user.FullName,
			Role:     user.Role,
		},
	}, nil
}

// Policy Decision Point (PDP) / Policy Enforcement Point (PEP)
type PolicyEngine struct{}

func NewPolicyEngine() *PolicyEngine {
	return &PolicyEngine{}
}

func (pe *PolicyEngine) CanAccessAuditLogs(role domain.Role) bool {
	return role == domain.RoleAdmin || role == domain.RoleAuditor
}

func (pe *PolicyEngine) CanManageUser(actorRole domain.Role, targetUserID string, actorUserID string) bool {
	if actorRole == domain.RoleAdmin {
		return true
	}
	return actorUserID == targetUserID
}

func (pe *PolicyEngine) CanModifyTodo(actorRole domain.Role, actorUserID, todoOwnerID string) bool {
	if actorRole == domain.RoleAdmin {
		return true
	}
	return actorUserID == todoOwnerID
}
