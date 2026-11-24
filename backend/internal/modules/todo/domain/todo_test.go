package domain_test

import (
	"testing"
	"time"

	"github.com/golang-todo/modular-monolith/internal/modules/todo/domain"
)

func TestTodo_Invariants(t *testing.T) {
	t.Run("Title cannot be blank", func(t *testing.T) {
		_, err := domain.NewTodo("id-1", "tenant-1", "user-1", "   ", "desc", domain.PriorityMedium, nil, "")
		if err == nil {
			t.Fatalf("expected error for empty title, got nil")
		}
	})

	t.Run("Valid creation with defaults", func(t *testing.T) {
		now := time.Now().Add(24 * time.Hour)
		todo, err := domain.NewTodo("id-1", "tenant-1", "user-1", "Ship Features", "Build robust architecture", domain.PriorityHigh, &now, "backend,go")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if todo.Status != domain.StatusPending {
			t.Errorf("expected status %s, got %s", domain.StatusPending, todo.Status)
		}
		if todo.Version != 1 {
			t.Errorf("expected initial version 1, got %d", todo.Version)
		}
	})

	t.Run("State Machine Transitions", func(t *testing.T) {
		todo, _ := domain.NewTodo("id-1", "tenant-1", "user-1", "Task", "desc", domain.PriorityMedium, nil, "")

		// Pending -> InProgress
		if err := todo.TransitionTo(domain.StatusInProgress); err != nil {
			t.Fatalf("expected valid transition to IN_PROGRESS, got: %v", err)
		}
		if todo.Status != domain.StatusInProgress {
			t.Errorf("expected IN_PROGRESS, got %s", todo.Status)
		}

		// InProgress -> Completed
		if err := todo.TransitionTo(domain.StatusCompleted); err != nil {
			t.Fatalf("expected valid transition to COMPLETED, got: %v", err)
		}
		if todo.Status != domain.StatusCompleted {
			t.Errorf("expected COMPLETED, got %s", todo.Status)
		}

		// Completed -> Pending (Reopen)
		if err := todo.TransitionTo(domain.StatusPending); err != nil {
			t.Fatalf("expected reopen to PENDING, got: %v", err)
		}

		// Pending -> Cancelled
		if err := todo.TransitionTo(domain.StatusCancelled); err != nil {
			t.Fatalf("expected transition to CANCELLED, got: %v", err)
		}

		// Cancelled cannot directly go to Completed
		if err := todo.TransitionTo(domain.StatusCompleted); err == nil {
			t.Fatalf("expected invariant error for Cancelled -> Completed, got nil")
		}
	})
}
