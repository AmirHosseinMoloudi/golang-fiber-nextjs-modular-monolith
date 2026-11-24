package application

import (
	"context"
	"encoding/json"

	"github.com/golang-todo/modular-monolith/internal/core/observability"
	"github.com/golang-todo/modular-monolith/internal/modules/audit/domain"
)

type AuditService struct {
	repo domain.AuditRepository
}

func NewAuditService(repo domain.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

type RecordAuditCommand struct {
	TenantID   string             `json:"tenant_id"`
	ActorID    string             `json:"actor_id"`
	ActorEmail string             `json:"actor_email"`
	Action     domain.AuditAction `json:"action"`
	EntityType string             `json:"entity_type"`
	EntityID   string             `json:"entity_id"`
	Data       interface{}        `json:"data"`
}

func (s *AuditService) Log(ctx context.Context, cmd RecordAuditCommand) error {
	latest, _ := s.repo.GetLatestRecord(cmd.TenantID)
	prevHash := "0000000000000000000000000000000000000000000000000000000000000000"
	if latest != nil {
		prevHash = latest.RecordHash
	}

	payloadBytes, _ := json.Marshal(cmd.Data)
	record := domain.NewAuditRecord(
		cmd.TenantID,
		cmd.ActorID,
		cmd.ActorEmail,
		cmd.Action,
		cmd.EntityType,
		cmd.EntityID,
		string(payloadBytes),
		prevHash,
	)

	observability.Logger.Info().
		Str("audit_id", record.ID).
		Str("action", string(record.Action)).
		Str("entity_type", record.EntityType).
		Str("entity_id", record.EntityID).
		Str("actor", record.ActorEmail).
		Msg("Appending tamper-evident audit record")

	return s.repo.Append(record)
}

func (s *AuditService) ListAuditRecords(ctx context.Context, tenantID string, page, pageSize int) ([]*domain.AuditRecord, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.List(tenantID, pageSize, offset)
}

func (s *AuditService) VerifyIntegrity(ctx context.Context, tenantID string) (bool, error) {
	return s.repo.VerifyChainIntegrity(tenantID)
}
