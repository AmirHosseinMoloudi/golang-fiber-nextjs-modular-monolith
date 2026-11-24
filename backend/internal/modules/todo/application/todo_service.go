package application

import (
	"context"
	"time"

	"github.com/golang-todo/modular-monolith/internal/core/errors"
	"github.com/golang-todo/modular-monolith/internal/core/observability"
	auditApp "github.com/golang-todo/modular-monolith/internal/modules/audit/application"
	auditDomain "github.com/golang-todo/modular-monolith/internal/modules/audit/domain"
	auditInfra "github.com/golang-todo/modular-monolith/internal/modules/audit/infrastructure"
	identityApp "github.com/golang-todo/modular-monolith/internal/modules/identity/application"
	identityDomain "github.com/golang-todo/modular-monolith/internal/modules/identity/domain"
	"github.com/golang-todo/modular-monolith/internal/modules/todo/domain"
	"github.com/google/uuid"
)

type TodoService struct {
	repo         domain.TodoRepository
	policyEngine *identityApp.PolicyEngine
	auditService *auditApp.AuditService
	outbox       *auditInfra.OutboxDispatcher
}

func NewTodoService(
	repo domain.TodoRepository,
	pe *identityApp.PolicyEngine,
	auditSvc *auditApp.AuditService,
	outbox *auditInfra.OutboxDispatcher,
) *TodoService {
	return &TodoService{
		repo:         repo,
		policyEngine: pe,
		auditService: auditSvc,
		outbox:       outbox,
	}
}

type CreateTodoCommand struct {
	TenantID    string              `json:"tenant_id"`
	UserID      string              `json:"user_id"`
	UserEmail   string              `json:"user_email"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Priority    domain.TodoPriority `json:"priority"`
	DueDate     *time.Time          `json:"due_date,omitempty"`
	Tags        string              `json:"tags"`
}

type UpdateTodoCommand struct {
	ID          string              `json:"id"`
	TenantID    string              `json:"tenant_id"`
	UserID      string              `json:"user_id"`
	UserEmail   string              `json:"user_email"`
	UserRole    string              `json:"user_role"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Priority    domain.TodoPriority `json:"priority"`
	DueDate     *time.Time          `json:"due_date,omitempty"`
	Tags        string              `json:"tags"`
	Version     int64               `json:"version"`
}

type ChangeStatusCommand struct {
	ID        string            `json:"id"`
	TenantID  string            `json:"tenant_id"`
	UserID    string            `json:"user_id"`
	UserEmail string            `json:"user_email"`
	UserRole  string            `json:"user_role"`
	Status    domain.TodoStatus `json:"status"`
	Version   int64             `json:"version"`
}

type TodoDTO struct {
	ID          string              `json:"id"`
	TenantID    string              `json:"tenant_id"`
	UserID      string              `json:"user_id"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Status      domain.TodoStatus   `json:"status"`
	Priority    domain.TodoPriority `json:"priority"`
	DueDate     *time.Time          `json:"due_date,omitempty"`
	Tags        string              `json:"tags"`
	IsOverdue   bool                `json:"is_overdue"`
	Version     int64               `json:"version"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

func toDTO(t *domain.Todo) TodoDTO {
	return TodoDTO{
		ID:          t.ID,
		TenantID:    t.TenantID,
		UserID:      t.UserID,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		Priority:    t.Priority,
		DueDate:     t.DueDate,
		Tags:        t.Tags,
		IsOverdue:   t.IsOverdue(),
		Version:     t.Version,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func (s *TodoService) Create(ctx context.Context, cmd CreateTodoCommand) (*TodoDTO, error) {
	ctx, span := observability.Tracer.Start(ctx, "TodoService.Create")
	defer span.End()

	todoID := uuid.New().String()
	todo, err := domain.NewTodo(todoID, cmd.TenantID, cmd.UserID, cmd.Title, cmd.Description, cmd.Priority, cmd.DueDate, cmd.Tags)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(todo); err != nil {
		return nil, errors.NewInternalError("failed to create todo item")
	}

	// Emit events to outbox and audit log
	events := todo.PullEvents()
	for _, e := range events {
		_ = s.outbox.SaveEvent(nil, todo.ID, e.EventType(), e)
	}

	_ = s.auditService.Log(ctx, auditApp.RecordAuditCommand{
		TenantID:   cmd.TenantID,
		ActorID:    cmd.UserID,
		ActorEmail: cmd.UserEmail,
		Action:     auditDomain.ActionCreated,
		EntityType: "TODO",
		EntityID:   todo.ID,
		Data:       toDTO(todo),
	})

	observability.ActiveTodosGauge.WithLabelValues(string(todo.Status), string(todo.Priority)).Inc()

	dto := toDTO(todo)
	return &dto, nil
}

func (s *TodoService) Update(ctx context.Context, cmd UpdateTodoCommand) (*TodoDTO, error) {
	ctx, span := observability.Tracer.Start(ctx, "TodoService.Update")
	defer span.End()

	todo, err := s.repo.FindByID(cmd.ID, cmd.TenantID)
	if err != nil {
		return nil, err
	}

	if !s.policyEngine.CanModifyTodo(identityDomain.Role(cmd.UserRole), cmd.UserID, todo.UserID) {
		return nil, errors.NewForbiddenError("you do not have permission to modify this todo")
	}

	expectedVersion := cmd.Version
	if expectedVersion == 0 {
		expectedVersion = todo.Version
	}

	if err := todo.UpdateDetails(cmd.Title, cmd.Description, cmd.Priority, cmd.DueDate, cmd.Tags); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateWithVersion(todo, expectedVersion); err != nil {
		return nil, err
	}

	events := todo.PullEvents()
	for _, e := range events {
		_ = s.outbox.SaveEvent(nil, todo.ID, e.EventType(), e)
	}

	_ = s.auditService.Log(ctx, auditApp.RecordAuditCommand{
		TenantID:   cmd.TenantID,
		ActorID:    cmd.UserID,
		ActorEmail: cmd.UserEmail,
		Action:     auditDomain.ActionUpdated,
		EntityType: "TODO",
		EntityID:   todo.ID,
		Data:       toDTO(todo),
	})

	dto := toDTO(todo)
	return &dto, nil
}

func (s *TodoService) ChangeStatus(ctx context.Context, cmd ChangeStatusCommand) (*TodoDTO, error) {
	ctx, span := observability.Tracer.Start(ctx, "TodoService.ChangeStatus")
	defer span.End()

	todo, err := s.repo.FindByID(cmd.ID, cmd.TenantID)
	if err != nil {
		return nil, err
	}

	if !s.policyEngine.CanModifyTodo(identityDomain.Role(cmd.UserRole), cmd.UserID, todo.UserID) {
		return nil, errors.NewForbiddenError("you do not have permission to transition status of this todo")
	}

	expectedVersion := cmd.Version
	if expectedVersion == 0 {
		expectedVersion = todo.Version
	}

	prevStatus := todo.Status
	prevPriority := todo.Priority

	if err := todo.TransitionTo(cmd.Status); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateWithVersion(todo, expectedVersion); err != nil {
		return nil, err
	}

	observability.ActiveTodosGauge.WithLabelValues(string(prevStatus), string(prevPriority)).Dec()
	observability.ActiveTodosGauge.WithLabelValues(string(todo.Status), string(todo.Priority)).Inc()

	events := todo.PullEvents()
	for _, e := range events {
		_ = s.outbox.SaveEvent(nil, todo.ID, e.EventType(), e)
	}

	_ = s.auditService.Log(ctx, auditApp.RecordAuditCommand{
		TenantID:   cmd.TenantID,
		ActorID:    cmd.UserID,
		ActorEmail: cmd.UserEmail,
		Action:     auditDomain.ActionStatusChanged,
		EntityType: "TODO",
		EntityID:   todo.ID,
		Data:       map[string]interface{}{"from": prevStatus, "to": todo.Status},
	})

	dto := toDTO(todo)
	return &dto, nil
}

func (s *TodoService) Delete(ctx context.Context, id, tenantID, userID, userRole, userEmail string) error {
	ctx, span := observability.Tracer.Start(ctx, "TodoService.Delete")
	defer span.End()

	todo, err := s.repo.FindByID(id, tenantID)
	if err != nil {
		return err
	}

	if !s.policyEngine.CanModifyTodo(identityDomain.Role(userRole), userID, todo.UserID) {
		return errors.NewForbiddenError("you do not have permission to delete this todo")
	}

	if err := s.repo.Delete(id, tenantID); err != nil {
		return err
	}

	observability.ActiveTodosGauge.WithLabelValues(string(todo.Status), string(todo.Priority)).Dec()

	_ = s.outbox.SaveEvent(nil, id, "TODO_DELETED", map[string]string{"id": id})
	_ = s.auditService.Log(ctx, auditApp.RecordAuditCommand{
		TenantID:   tenantID,
		ActorID:    userID,
		ActorEmail: userEmail,
		Action:     auditDomain.ActionDeleted,
		EntityType: "TODO",
		EntityID:   id,
		Data:       map[string]string{"deleted_id": id},
	})

	return nil
}

func (s *TodoService) GetByID(ctx context.Context, id, tenantID string) (*TodoDTO, error) {
	todo, err := s.repo.FindByID(id, tenantID)
	if err != nil {
		return nil, err
	}
	dto := toDTO(todo)
	return &dto, nil
}

func (s *TodoService) List(
	ctx context.Context,
	tenantID, userID, userRole string,
	status *domain.TodoStatus,
	priority *domain.TodoPriority,
	search string,
	page, pageSize int,
) ([]TodoDTO, int64, error) {
	filterUserID := userID
	if userRole == "ADMIN" || userRole == "AUDITOR" {
		filterUserID = "" // Admins can view all within the tenant
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	todos, total, err := s.repo.List(tenantID, filterUserID, status, priority, search, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]TodoDTO, len(todos))
	for i, t := range todos {
		dtos[i] = toDTO(t)
	}
	return dtos, total, nil
}

func (s *TodoService) GetStats(ctx context.Context, tenantID, userID, userRole string) (map[string]int64, error) {
	filterUserID := userID
	if userRole == "ADMIN" || userRole == "AUDITOR" {
		filterUserID = ""
	}
	return s.repo.GetStats(tenantID, filterUserID)
}
