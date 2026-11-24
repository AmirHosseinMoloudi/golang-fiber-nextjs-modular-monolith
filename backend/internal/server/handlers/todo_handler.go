package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-todo/modular-monolith/internal/core/errors"
	todoApp "github.com/golang-todo/modular-monolith/internal/modules/todo/application"
	"github.com/golang-todo/modular-monolith/internal/modules/todo/domain"
)

type TodoHandler struct {
	svc *todoApp.TodoService
}

func NewTodoHandler(svc *todoApp.TodoService) *TodoHandler {
	return &TodoHandler{svc: svc}
}

func (h *TodoHandler) Create(c *fiber.Ctx) error {
	var cmd todoApp.CreateTodoCommand
	if err := c.BodyParser(&cmd); err != nil {
		return errors.NewValidationError("invalid request body format", nil)
	}

	cmd.TenantID = c.Locals("tenant_id").(string)
	cmd.UserID = c.Locals("user_id").(string)
	cmd.UserEmail = c.Locals("user_email").(string)

	todo, err := h.svc.Create(c.Context(), cmd)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(todo)
}

func (h *TodoHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var cmd todoApp.UpdateTodoCommand
	if err := c.BodyParser(&cmd); err != nil {
		return errors.NewValidationError("invalid request body format", nil)
	}

	cmd.ID = id
	cmd.TenantID = c.Locals("tenant_id").(string)
	cmd.UserID = c.Locals("user_id").(string)
	cmd.UserEmail = c.Locals("user_email").(string)
	cmd.UserRole = c.Locals("user_role").(string)

	todo, err := h.svc.Update(c.Context(), cmd)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(todo)
}

func (h *TodoHandler) ChangeStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	var cmd todoApp.ChangeStatusCommand
	if err := c.BodyParser(&cmd); err != nil {
		return errors.NewValidationError("invalid request body format", nil)
	}

	cmd.ID = id
	cmd.TenantID = c.Locals("tenant_id").(string)
	cmd.UserID = c.Locals("user_id").(string)
	cmd.UserEmail = c.Locals("user_email").(string)
	cmd.UserRole = c.Locals("user_role").(string)

	todo, err := h.svc.ChangeStatus(c.Context(), cmd)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(todo)
}

func (h *TodoHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	tenantID := c.Locals("tenant_id").(string)
	userID := c.Locals("user_id").(string)
	userEmail := c.Locals("user_email").(string)
	userRole := c.Locals("user_role").(string)

	if err := h.svc.Delete(c.Context(), id, tenantID, userID, userRole, userEmail); err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "todo deleted successfully", "id": id})
}

func (h *TodoHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	tenantID := c.Locals("tenant_id").(string)

	todo, err := h.svc.GetByID(c.Context(), id, tenantID)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(todo)
}

func (h *TodoHandler) List(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	userID := c.Locals("user_id").(string)
	userRole := c.Locals("user_role").(string)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	search := c.Query("search", "")

	var statusPtr *domain.TodoStatus
	if s := c.Query("status", ""); s != "" {
		st := domain.TodoStatus(s)
		if st.IsValid() {
			statusPtr = &st
		}
	}

	var priorityPtr *domain.TodoPriority
	if p := c.Query("priority", ""); p != "" {
		pr := domain.TodoPriority(p)
		if pr.IsValid() {
			priorityPtr = &pr
		}
	}

	todos, total, err := h.svc.List(c.Context(), tenantID, userID, userRole, statusPtr, priorityPtr, search, page, pageSize)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":      todos,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *TodoHandler) Stats(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	userID := c.Locals("user_id").(string)
	userRole := c.Locals("user_role").(string)

	stats, err := h.svc.GetStats(c.Context(), tenantID, userID, userRole)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"stats": stats})
}
