package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-todo/modular-monolith/internal/core/errors"
	identityApp "github.com/golang-todo/modular-monolith/internal/modules/identity/application"
)

type AuthHandler struct {
	svc *identityApp.IdentityService
}

func NewAuthHandler(svc *identityApp.IdentityService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req identityApp.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.NewValidationError("invalid request body payload", nil)
	}

	resp, err := h.svc.Register(c.Context(), req)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req identityApp.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.NewValidationError("invalid request body payload", nil)
	}

	resp, err := h.svc.Login(c.Context(), req)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*identityApp.JWTClaims)
	if !ok || claims == nil {
		return errors.NewUnauthorizedError("unauthenticated")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user": identityApp.UserDTO{
			ID:       claims.UserID,
			TenantID: claims.TenantID,
			Email:    claims.Email,
			FullName: claims.FullName,
			Role:     claims.Role,
		},
	})
}
