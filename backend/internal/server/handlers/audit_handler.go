package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	auditApp "github.com/golang-todo/modular-monolith/internal/modules/audit/application"
)

type AuditHandler struct {
	svc *auditApp.AuditService
}

func NewAuditHandler(svc *auditApp.AuditService) *AuditHandler {
	return &AuditHandler{svc: svc}
}

func (h *AuditHandler) List(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	records, total, err := h.svc.ListAuditRecords(c.Context(), tenantID, page, pageSize)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":      records,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *AuditHandler) Verify(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	valid, err := h.svc.VerifyIntegrity(c.Context(), tenantID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"tamper_evident_integrity_valid": valid,
		"tenant_id":                      tenantID,
	})
}
