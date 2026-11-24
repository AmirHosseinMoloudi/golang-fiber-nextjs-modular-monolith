package domain

import (
	"strings"
	"time"

	"github.com/golang-todo/modular-monolith/internal/core/domain"
	"github.com/golang-todo/modular-monolith/internal/core/errors"
)

type TodoStatus string

const (
	StatusPending    TodoStatus = "PENDING"
	StatusInProgress TodoStatus = "IN_PROGRESS"
	StatusCompleted  TodoStatus = "COMPLETED"
	StatusCancelled  TodoStatus = "CANCELLED"
)

func (s TodoStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusInProgress, StatusCompleted, StatusCancelled:
		return true
	default:
		return false
	}
}

type TodoPriority string

const (
	PriorityLow    TodoPriority = "LOW"
	PriorityMedium TodoPriority = "MEDIUM"
	PriorityHigh   TodoPriority = "HIGH"
	PriorityUrgent TodoPriority = "URGENT"
)

func (p TodoPriority) IsValid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent:
		return true
	default:
		return false
	}
}

// Todo Aggregate Root
type Todo struct {
	domain.AggregateRoot
	ID          string       `json:"id"`
	TenantID    string       `json:"tenant_id"`
	UserID      string       `json:"user_id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Status      TodoStatus   `json:"status"`
	Priority    TodoPriority `json:"priority"`
	DueDate     *time.Time   `json:"due_date,omitempty"`
	Tags        string       `json:"tags"`
	Version     int64        `json:"version"` // Optimistic Concurrency Control
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

func NewTodo(id, tenantID, userID, title, description string, priority TodoPriority, dueDate *time.Time, tags string) (*Todo, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, errors.NewValidationError("todo title cannot be blank", map[string]string{"title": "required"})
	}
	if len(title) > 255 {
		return nil, errors.NewValidationError("todo title must not exceed 255 characters", map[string]string{"title": "max_length_exceeded"})
	}

	if !priority.IsValid() {
		priority = PriorityMedium
	}

	now := time.Now().UTC()
	todo := &Todo{
		ID:          id,
		TenantID:    tenantID,
		UserID:      userID,
		Title:       title,
		Description: strings.TrimSpace(description),
		Status:      StatusPending,
		Priority:    priority,
		DueDate:     dueDate,
		Tags:        strings.TrimSpace(tags),
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	todo.RecordEvent(domain.NewBaseDomainEvent("TODO_CREATED", todo.ID))
	return todo, nil
}

// State Machine transitions enforcing domain invariants
func (t *Todo) TransitionTo(targetStatus TodoStatus) error {
	if !targetStatus.IsValid() {
		return errors.NewValidationError("invalid target status", map[string]string{"status": string(targetStatus)})
	}
	if t.Status == targetStatus {
		return nil // No-op
	}

	// Invariant rules: State transitions
	switch t.Status {
	case StatusPending:
		if targetStatus != StatusInProgress && targetStatus != StatusCancelled && targetStatus != StatusCompleted {
			return errors.NewPreconditionFailedError("pending item can only transition to in-progress, completed, or cancelled")
		}
	case StatusInProgress:
		if targetStatus != StatusCompleted && targetStatus != StatusCancelled && targetStatus != StatusPending {
			return errors.NewPreconditionFailedError("in-progress item cannot be transitioned directly to this state")
		}
	case StatusCompleted:
		if targetStatus != StatusInProgress && targetStatus != StatusPending {
			return errors.NewPreconditionFailedError("completed item can only be reopened to pending or in-progress")
		}
	case StatusCancelled:
		if targetStatus != StatusPending {
			return errors.NewPreconditionFailedError("cancelled item can only be reactivated to pending")
		}
	}

	oldStatus := t.Status
	t.Status = targetStatus
	t.UpdatedAt = time.Now().UTC()
	t.Version++

	t.RecordEvent(domain.NewBaseDomainEvent("TODO_STATUS_CHANGED", t.ID))
	_ = oldStatus // Used in event tracking
	return nil
}

func (t *Todo) UpdateDetails(title, description string, priority TodoPriority, dueDate *time.Time, tags string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return errors.NewValidationError("todo title cannot be blank", map[string]string{"title": "required"})
	}
	if !priority.IsValid() {
		priority = PriorityMedium
	}

	t.Title = title
	t.Description = strings.TrimSpace(description)
	t.Priority = priority
	t.DueDate = dueDate
	t.Tags = strings.TrimSpace(tags)
	t.UpdatedAt = time.Now().UTC()
	t.Version++

	t.RecordEvent(domain.NewBaseDomainEvent("TODO_UPDATED", t.ID))
	return nil
}

func (t *Todo) IsOverdue() bool {
	if t.DueDate == nil || t.Status == StatusCompleted || t.Status == StatusCancelled {
		return false
	}
	return time.Now().UTC().After(*t.DueDate)
}

type TodoRepository interface {
	Save(todo *Todo) error
	UpdateWithVersion(todo *Todo, expectedVersion int64) error
	FindByID(id, tenantID string) (*Todo, error)
	Delete(id, tenantID string) error
	List(tenantID, userID string, status *TodoStatus, priority *TodoPriority, search string, limit, offset int) ([]*Todo, int64, error)
	GetStats(tenantID, userID string) (map[string]int64, error)
}
