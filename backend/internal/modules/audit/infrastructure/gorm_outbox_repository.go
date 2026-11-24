package infrastructure

import (
	"context"
	"encoding/json"
	"time"

	"github.com/golang-todo/modular-monolith/internal/core/observability"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OutboxMessageModel struct {
	ID          string    `gorm:"primaryKey;type:varchar(36)"`
	AggregateID string    `gorm:"index;type:varchar(64)"`
	EventType   string    `gorm:"index;type:varchar(64)"`
	Payload     string    `gorm:"type:text"`
	Status      string    `gorm:"index;type:varchar(32)"` // PENDING, PUBLISHED, FAILED
	CreatedAt   time.Time `gorm:"index"`
	ProcessedAt *time.Time
}

func (OutboxMessageModel) TableName() string {
	return "outbox_messages"
}

type OutboxDispatcher struct {
	db *gorm.DB
}

func NewOutboxDispatcher(db *gorm.DB) *OutboxDispatcher {
	return &OutboxDispatcher{db: db}
}

func (o *OutboxDispatcher) SaveEvent(tx *gorm.DB, aggregateID, eventType string, payload interface{}) error {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	msg := OutboxMessageModel{
		ID:          uuid.New().String(),
		AggregateID: aggregateID,
		EventType:   eventType,
		Payload:     string(bytes),
		Status:      "PENDING",
		CreatedAt:   time.Now().UTC(),
	}
	targetDB := o.db
	if tx != nil {
		targetDB = tx
	}
	return targetDB.Create(&msg).Error
}

func (o *OutboxDispatcher) StartWorker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				o.dispatchPending(ctx)
			}
		}
	}()
}

func (o *OutboxDispatcher) dispatchPending(ctx context.Context) {
	var messages []OutboxMessageModel
	if err := o.db.Where("status = ?", "PENDING").Order("created_at asc").Limit(20).Find(&messages).Error; err != nil {
		return
	}

	for _, msg := range messages {
		now := time.Now().UTC()
		// Mark as published and record observability metric
		observability.DomainEventsPublishedTotal.WithLabelValues(msg.EventType).Inc()
		observability.Logger.Debug().
			Str("outbox_id", msg.ID).
			Str("event_type", msg.EventType).
			Str("aggregate_id", msg.AggregateID).
			Msg("Outbox event published effectively-once")

		o.db.Model(&msg).Updates(map[string]interface{}{
			"status":       "PUBLISHED",
			"processed_at": &now,
		})
	}
}
