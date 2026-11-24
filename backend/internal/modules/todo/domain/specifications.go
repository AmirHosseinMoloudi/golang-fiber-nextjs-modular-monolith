package domain

import (
	"github.com/golang-todo/modular-monolith/internal/core/domain"
)

type StatusSpecification struct {
	Status TodoStatus
}

func (s StatusSpecification) IsSatisfiedBy(candidate *Todo) bool {
	return candidate.Status == s.Status
}

type PrioritySpecification struct {
	Priority TodoPriority
}

func (s PrioritySpecification) IsSatisfiedBy(candidate *Todo) bool {
	return candidate.Priority == s.Priority
}

type OverdueSpecification struct{}

func (s OverdueSpecification) IsSatisfiedBy(candidate *Todo) bool {
	return candidate.IsOverdue()
}

// Composite Specification helpers
func NewUrgentAndOverdueSpec() domain.Specification[*Todo] {
	urgentSpec := PrioritySpecification{Priority: PriorityUrgent}
	overdueSpec := OverdueSpecification{}
	return domain.And[*Todo](urgentSpec, overdueSpec)
}
