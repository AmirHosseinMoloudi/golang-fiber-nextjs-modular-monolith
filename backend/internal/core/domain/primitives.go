package domain

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents a notable business event within the bounded context.
type DomainEvent interface {
	EventID() string
	EventType() string
	OccurredAt() time.Time
	AggregateID() string
}

// BaseDomainEvent provides boilerplate for DomainEvent implementations.
type BaseDomainEvent struct {
	ID          string    `json:"event_id"`
	Type        string    `json:"event_type"`
	Timestamp   time.Time `json:"occurred_at"`
	AggID       string    `json:"aggregate_id"`
}

func NewBaseDomainEvent(eventType, aggregateID string) BaseDomainEvent {
	return BaseDomainEvent{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now().UTC(),
		AggID:     aggregateID,
	}
}

func (e BaseDomainEvent) EventID() string      { return e.ID }
func (e BaseDomainEvent) EventType() string    { return e.Type }
func (e BaseDomainEvent) OccurredAt() time.Time { return e.Timestamp }
func (e BaseDomainEvent) AggregateID() string  { return e.AggID }

// AggregateRoot is the base for aggregates, managing uncommitted domain events.
type AggregateRoot struct {
	events []DomainEvent
}

func (a *AggregateRoot) RecordEvent(event DomainEvent) {
	a.events = append(a.events, event)
}

func (a *AggregateRoot) PullEvents() []DomainEvent {
	events := a.events
	a.events = nil
	return events
}

// Specification pattern interface for rich domain querying and rule evaluation.
type Specification[T any] interface {
	IsSatisfiedBy(candidate T) bool
}

type andSpecification[T any] struct {
	left, right Specification[T]
}

func (s andSpecification[T]) IsSatisfiedBy(candidate T) bool {
	return s.left.IsSatisfiedBy(candidate) && s.right.IsSatisfiedBy(candidate)
}

func And[T any](left, right Specification[T]) Specification[T] {
	return andSpecification[T]{left: left, right: right}
}

type orSpecification[T any] struct {
	left, right Specification[T]
}

func (s orSpecification[T]) IsSatisfiedBy(candidate T) bool {
	return s.left.IsSatisfiedBy(candidate) || s.right.IsSatisfiedBy(candidate)
}

func Or[T any](left, right Specification[T]) Specification[T] {
	return orSpecification[T]{left: left, right: right}
}

type notSpecification[T any] struct {
	spec Specification[T]
}

func (s notSpecification[T]) IsSatisfiedBy(candidate T) bool {
	return !s.spec.IsSatisfiedBy(candidate)
}

func Not[T any](spec Specification[T]) Specification[T] {
	return notSpecification[T]{spec: spec}
}

// Result Pattern for functional error handling without panics.
type Result[T any] struct {
	value T
	err   error
}

func Success[T any](value T) Result[T] {
	return Result[T]{value: value, err: nil}
}

func Failure[T any](err error) Result[T] {
	var zero T
	return Result[T]{value: zero, err: err}
}

func (r Result[T]) IsSuccess() bool { return r.err == nil }
func (r Result[T]) IsFailure() bool { return r.err != nil }
func (r Result[T]) Value() T        { return r.value }
func (r Result[T]) Error() error    { return r.err }
